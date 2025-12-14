use anyhow::Result;
use crossterm::event::{self, Event};
use ratatui::DefaultTerminal;
use std::path::PathBuf;
use tokio::sync::mpsc;
use uuid::Uuid;

use crate::api::NodeClient;
use crate::dns::create_provider;
use crate::events::{handle_event, Action};
use crate::models::site::SiteStatus;
use crate::state::{
    AppState, AsyncOperation, AsyncOperationResult, AsyncStatus, NotificationLevel,
    OperationType,
};
use crate::ui::{render, Screen};

pub struct App {
    pub state: AppState,
    pub node_client: NodeClient,
    pub action_tx: mpsc::UnboundedSender<Action>,
    pub action_rx: mpsc::UnboundedReceiver<Action>,
}

impl App {
    pub fn new(config_path: PathBuf) -> Result<Self> {
        let (action_tx, action_rx) = mpsc::unbounded_channel();

        Ok(Self {
            state: AppState::new(config_path)?,
            node_client: NodeClient::new()?,
            action_tx,
            action_rx,
        })
    }

    /// Main event loop following The Elm Architecture pattern
    pub async fn run(&mut self, mut terminal: DefaultTerminal) -> Result<()> {
        loop {
            // Render (TEA View)
            terminal.draw(|frame| render(frame, &self.state))?;

            // Handle events - multiplex between user input and async results
            tokio::select! {
                // User input events
                event_result = tokio::task::spawn_blocking(|| event::read()) => {
                    if let Ok(Ok(event)) = event_result {
                        let action = handle_event(event, &self.state.current_screen);
                        if !matches!(action, Action::None) {
                            self.update(action)?;
                        }
                    }
                }

                // Async operation results
                Some(action) = self.action_rx.recv() => {
                    self.update(action)?;
                }
            }

            // Check if we should quit
            if self.state.should_quit {
                break;
            }
        }

        Ok(())
    }

    /// Update function (TEA Update) - processes Actions and updates state
    pub fn update(&mut self, action: Action) -> Result<()> {
        match action {
            // Navigation
            Action::NavigateTo(screen) => {
                self.state.navigate_to(screen);
            }

            Action::NavigateBack => {
                self.state.navigate_back();
            }

            // Site management
            Action::CreateSite(site) => {
                let site_id = site.id;
                self.state.add_site(site);
                self.state.add_notification(
                    "Site created successfully".to_string(),
                    NotificationLevel::Success,
                );

                // Auto-deploy on creation
                self.spawn_deploy_site(site_id);

                if self.state.auto_save {
                    self.state.save()?;
                }
            }

            Action::UpdateSite(site_id, updated_site) => {
                if let Some(site) = self.state.get_site_mut(site_id) {
                    *site = updated_site;
                    self.state.add_notification(
                        "Site updated successfully".to_string(),
                        NotificationLevel::Success,
                    );

                    // Auto-redeploy on update
                    self.spawn_deploy_site(site_id);

                    if self.state.auto_save {
                        self.state.save()?;
                    }
                }
            }

            Action::DeleteSite(site_id) => {
                // Spawn async delete operation
                self.spawn_delete_site(site_id);
            }

            Action::DeploySite(site_id) => {
                self.spawn_deploy_site(site_id);
            }

            Action::StopSite(site_id) => {
                self.spawn_stop_site(site_id);
            }

            Action::RestartSite(site_id) => {
                self.spawn_restart_site(site_id);
            }

            // Domain management
            Action::CreateDomain(domain) => {
                self.state.add_domain(domain);
                self.state.add_notification(
                    "Domain created successfully".to_string(),
                    NotificationLevel::Success,
                );

                if self.state.auto_save {
                    self.state.save()?;
                }
            }

            Action::DeleteDomain(domain_id) => {
                self.state.remove_domain(domain_id);
                self.state.add_notification(
                    "Domain deleted successfully".to_string(),
                    NotificationLevel::Success,
                );

                if self.state.auto_save {
                    self.state.save()?;
                }
            }

            Action::AddDnsRecord(domain_id, record) => {
                if let Some(domain) = self.state.get_domain_mut(domain_id) {
                    domain.dns_records.push(record);
                    self.state.add_notification(
                        "DNS record added".to_string(),
                        NotificationLevel::Success,
                    );

                    if self.state.auto_save {
                        self.state.save()?;
                    }
                }
            }

            Action::UpdateDnsRecord(domain_id, index, record) => {
                if let Some(domain) = self.state.get_domain_mut(domain_id) {
                    if index < domain.dns_records.len() {
                        domain.dns_records[index] = record;
                        self.state.add_notification(
                            "DNS record updated".to_string(),
                            NotificationLevel::Success,
                        );

                        if self.state.auto_save {
                            self.state.save()?;
                        }
                    }
                }
            }

            Action::DeleteDnsRecord(domain_id, index) => {
                if let Some(domain) = self.state.get_domain_mut(domain_id) {
                    if index < domain.dns_records.len() {
                        domain.dns_records.remove(index);
                        self.state.add_notification(
                            "DNS record deleted".to_string(),
                            NotificationLevel::Success,
                        );

                        if self.state.auto_save {
                            self.state.save()?;
                        }
                    }
                }
            }

            Action::SyncDnsRecords(domain_id) => {
                self.spawn_sync_dns(domain_id);
            }

            // Node management
            Action::AddNode(node) => {
                self.state.add_node(node);
                self.state.add_notification(
                    "Node added successfully".to_string(),
                    NotificationLevel::Success,
                );

                if self.state.auto_save {
                    self.state.save()?;
                }
            }

            Action::UpdateNode(node_id, updated_node) => {
                if let Some(node) = self.state.get_node_mut(node_id) {
                    *node = updated_node;
                    self.state.add_notification(
                        "Node updated successfully".to_string(),
                        NotificationLevel::Success,
                    );

                    if self.state.auto_save {
                        self.state.save()?;
                    }
                }
            }

            Action::RemoveNode(node_id) => {
                self.state.remove_node(node_id);
                self.state.add_notification(
                    "Node removed successfully".to_string(),
                    NotificationLevel::Success,
                );

                if self.state.auto_save {
                    self.state.save()?;
                }
            }

            Action::CheckNodeHealth(node_id) => {
                self.spawn_node_health_check(node_id);
            }

            Action::FetchNodeStats(node_id) => {
                self.spawn_fetch_node_stats(node_id);
            }

            // Site monitoring
            Action::FetchLogs(site_id) => {
                self.spawn_fetch_logs(site_id);
            }

            Action::FetchMetrics(site_id) => {
                self.spawn_fetch_metrics(site_id);
            }

            // Async operation results
            Action::AsyncOperationCompleted(op_id, result) => {
                if let Some(op) = self.state.pending_operations.iter_mut().find(|o| o.id == op_id) {
                    match result {
                        Ok(async_result) => {
                            op.complete();
                            self.handle_async_result(async_result);
                        }
                        Err(error) => {
                            op.fail(error.clone());
                            self.state.add_notification(
                                format!("Operation failed: {}", error),
                                NotificationLevel::Error,
                            );
                        }
                    }
                }
            }

            // Notifications
            Action::ShowNotification(notification) => {
                self.state.notifications.push_back(notification);
            }

            Action::DismissNotification => {
                self.state.dismiss_notification();
            }

            // System
            Action::SaveConfig => {
                self.state.save()?;
                self.state.add_notification(
                    "Configuration saved".to_string(),
                    NotificationLevel::Success,
                );
            }

            Action::LoadConfig => {
                // Reload state from disk
                let new_state = AppState::new(self.state.config_path.clone())?;
                self.state = new_state;
                self.state.add_notification(
                    "Configuration reloaded".to_string(),
                    NotificationLevel::Info,
                );
            }

            Action::Quit => {
                if self.state.auto_save {
                    self.state.save()?;
                }
                self.state.should_quit = true;
            }

            // Selection navigation
            Action::SelectNext => {
                match self.state.current_screen {
                    Screen::SitesList => {
                        if !self.state.sites.is_empty() {
                            self.state.selection_state.sites_list_index =
                                (self.state.selection_state.sites_list_index + 1) % self.state.sites.len();
                        }
                    }
                    Screen::DomainsList => {
                        if !self.state.domains.is_empty() {
                            self.state.selection_state.domains_list_index =
                                (self.state.selection_state.domains_list_index + 1) % self.state.domains.len();
                        }
                    }
                    Screen::NodesList => {
                        if !self.state.nodes.is_empty() {
                            self.state.selection_state.nodes_list_index =
                                (self.state.selection_state.nodes_list_index + 1) % self.state.nodes.len();
                        }
                    }
                    _ => {}
                }
            }

            Action::SelectPrevious => {
                match self.state.current_screen {
                    Screen::SitesList => {
                        if !self.state.sites.is_empty() {
                            let len = self.state.sites.len();
                            self.state.selection_state.sites_list_index =
                                (self.state.selection_state.sites_list_index + len - 1) % len;
                        }
                    }
                    Screen::DomainsList => {
                        if !self.state.domains.is_empty() {
                            let len = self.state.domains.len();
                            self.state.selection_state.domains_list_index =
                                (self.state.selection_state.domains_list_index + len - 1) % len;
                        }
                    }
                    Screen::NodesList => {
                        if !self.state.nodes.is_empty() {
                            let len = self.state.nodes.len();
                            self.state.selection_state.nodes_list_index =
                                (self.state.selection_state.nodes_list_index + len - 1) % len;
                        }
                    }
                    _ => {}
                }
            }

            // No-op and unhandled
            Action::None | _ => {}
        }

        Ok(())
    }

    fn handle_async_result(&mut self, result: AsyncOperationResult) {
        match result {
            AsyncOperationResult::SiteDeployed(site_id) => {
                if let Some(site) = self.state.get_site_mut(site_id) {
                    site.status = SiteStatus::Running;
                }
                self.state.add_notification(
                    "Site deployed successfully".to_string(),
                    NotificationLevel::Success,
                );
            }

            AsyncOperationResult::SiteUpdated(site_id) => {
                if let Some(site) = self.state.get_site_mut(site_id) {
                    site.status = SiteStatus::Running;
                }
                self.state.add_notification(
                    "Site updated successfully".to_string(),
                    NotificationLevel::Success,
                );
            }

            AsyncOperationResult::SiteDeleted(site_id) => {
                self.state.remove_site(site_id);
                self.state.add_notification(
                    "Site deleted successfully".to_string(),
                    NotificationLevel::Success,
                );
            }

            AsyncOperationResult::DnsSynced(domain_id, records) => {
                if let Some(domain) = self.state.get_domain_mut(domain_id) {
                    domain.dns_records = records;
                }
                self.state.add_notification(
                    "DNS records synced successfully".to_string(),
                    NotificationLevel::Success,
                );
            }

            AsyncOperationResult::NodeHealth(node_id, status, docker_info, traefik_info) => {
                if let Some(node) = self.state.get_node_mut(node_id) {
                    node.update_health(status, docker_info, traefik_info);
                }
            }

            AsyncOperationResult::NodeStats(node_id, docker_info, traefik_info) => {
                if let Some(node) = self.state.get_node_mut(node_id) {
                    node.docker_info = Some(docker_info);
                    node.traefik_info = Some(traefik_info);
                }
            }

            _ => {}
        }
    }

    // Async operation spawners
    fn spawn_deploy_site(&self, site_id: Uuid) {
        let site = match self.state.get_site(site_id).cloned() {
            Some(s) => s,
            None => return,
        };

        let domain = match self.state.get_domain_for_site(site_id) {
            Some(d) => d.clone(),
            None => return,
        };

        let node = match self.state.get_node_for_site(site_id) {
            Some(n) => n.clone(),
            None => return,
        };

        let client = self.node_client.clone();
        let tx = self.action_tx.clone();
        let op_id = Uuid::new_v4();

        // Add to pending operations
        let mut op = AsyncOperation::new(OperationType::DeploySite(site_id));
        op.id = op_id;

        tokio::spawn(async move {
            let result = client
                .deploy_site(&node.api_endpoint, &node.api_key, &site, &domain.name)
                .await;

            let action = match result {
                Ok(_) => Action::AsyncOperationCompleted(
                    op_id,
                    Ok(AsyncOperationResult::SiteDeployed(site_id)),
                ),
                Err(e) => Action::AsyncOperationCompleted(op_id, Err(e.to_string())),
            };

            let _ = tx.send(action);
        });
    }

    fn spawn_delete_site(&self, site_id: Uuid) {
        let node = match self.state.get_node_for_site(site_id) {
            Some(n) => n.clone(),
            None => return,
        };

        let client = self.node_client.clone();
        let tx = self.action_tx.clone();
        let op_id = Uuid::new_v4();

        tokio::spawn(async move {
            let result = client.delete_site(&node.api_endpoint, &node.api_key, site_id).await;

            let action = match result {
                Ok(_) => Action::AsyncOperationCompleted(
                    op_id,
                    Ok(AsyncOperationResult::SiteDeleted(site_id)),
                ),
                Err(e) => Action::AsyncOperationCompleted(op_id, Err(e.to_string())),
            };

            let _ = tx.send(action);
        });
    }

    fn spawn_stop_site(&self, site_id: Uuid) {
        let node = match self.state.get_node_for_site(site_id) {
            Some(n) => n.clone(),
            None => return,
        };

        let client = self.node_client.clone();
        let tx = self.action_tx.clone();
        let op_id = Uuid::new_v4();

        tokio::spawn(async move {
            let result = client.stop_site(&node.api_endpoint, &node.api_key, site_id).await;

            let action = match result {
                Ok(_) => {
                    Action::ShowNotification(crate::state::Notification::success("Site stopped"))
                }
                Err(e) => Action::AsyncOperationCompleted(op_id, Err(e.to_string())),
            };

            let _ = tx.send(action);
        });
    }

    fn spawn_restart_site(&self, site_id: Uuid) {
        let node = match self.state.get_node_for_site(site_id) {
            Some(n) => n.clone(),
            None => return,
        };

        let client = self.node_client.clone();
        let tx = self.action_tx.clone();
        let op_id = Uuid::new_v4();

        tokio::spawn(async move {
            let result = client.restart_site(&node.api_endpoint, &node.api_key, site_id).await;

            let action = match result {
                Ok(_) => {
                    Action::ShowNotification(crate::state::Notification::success("Site restarted"))
                }
                Err(e) => Action::AsyncOperationCompleted(op_id, Err(e.to_string())),
            };

            let _ = tx.send(action);
        });
    }

    fn spawn_sync_dns(&self, domain_id: Uuid) {
        let domain = match self.state.get_domain(domain_id).cloned() {
            Some(d) => d,
            None => return,
        };

        if domain.is_manual_dns() {
            return; // Cannot sync manual DNS
        }

        let tx = self.action_tx.clone();
        let op_id = Uuid::new_v4();

        tokio::spawn(async move {
            let result = async {
                let provider = create_provider(&domain.dns_provider)?;
                provider.list_records(&domain.name).await
            }
            .await;

            let action = match result {
                Ok(records) => Action::AsyncOperationCompleted(
                    op_id,
                    Ok(AsyncOperationResult::DnsSynced(domain_id, records)),
                ),
                Err(e) => Action::AsyncOperationCompleted(op_id, Err(e.to_string())),
            };

            let _ = tx.send(action);
        });
    }

    fn spawn_node_health_check(&self, node_id: Uuid) {
        let node = match self.state.get_node(node_id).cloned() {
            Some(n) => n,
            None => return,
        };

        let client = self.node_client.clone();
        let tx = self.action_tx.clone();
        let op_id = Uuid::new_v4();

        tokio::spawn(async move {
            let result = client.health_check(&node.api_endpoint, &node.api_key).await;

            let action = match result {
                Ok(health) => Action::AsyncOperationCompleted(
                    op_id,
                    Ok(AsyncOperationResult::NodeHealth(
                        node_id,
                        health.status,
                        health.docker,
                        health.traefik,
                    )),
                ),
                Err(e) => Action::AsyncOperationCompleted(op_id, Err(e.to_string())),
            };

            let _ = tx.send(action);
        });
    }

    fn spawn_fetch_node_stats(&self, node_id: Uuid) {
        let node = match self.state.get_node(node_id).cloned() {
            Some(n) => n,
            None => return,
        };

        let client = self.node_client.clone();
        let tx = self.action_tx.clone();
        let op_id = Uuid::new_v4();

        tokio::spawn(async move {
            let docker_result = client.get_docker_info(&node.api_endpoint, &node.api_key).await;
            let traefik_result = client.get_traefik_info(&node.api_endpoint, &node.api_key).await;

            let action = match (docker_result, traefik_result) {
                (Ok(docker), Ok(traefik)) => Action::AsyncOperationCompleted(
                    op_id,
                    Ok(AsyncOperationResult::NodeStats(node_id, docker, traefik)),
                ),
                _ => Action::AsyncOperationCompleted(
                    op_id,
                    Err("Failed to fetch node stats".to_string()),
                ),
            };

            let _ = tx.send(action);
        });
    }

    fn spawn_fetch_logs(&self, site_id: Uuid) {
        let node = match self.state.get_node_for_site(site_id) {
            Some(n) => n.clone(),
            None => return,
        };

        let client = self.node_client.clone();
        let tx = self.action_tx.clone();
        let op_id = Uuid::new_v4();

        tokio::spawn(async move {
            let result = client
                .get_container_logs(&node.api_endpoint, &node.api_key, site_id, 100)
                .await;

            let action = match result {
                Ok(logs) => {
                    Action::AsyncOperationCompleted(op_id, Ok(AsyncOperationResult::Logs(site_id, logs)))
                }
                Err(e) => Action::AsyncOperationCompleted(op_id, Err(e.to_string())),
            };

            let _ = tx.send(action);
        });
    }

    fn spawn_fetch_metrics(&self, site_id: Uuid) {
        let node = match self.state.get_node_for_site(site_id) {
            Some(n) => n.clone(),
            None => return,
        };

        let client = self.node_client.clone();
        let tx = self.action_tx.clone();
        let op_id = Uuid::new_v4();

        tokio::spawn(async move {
            let result = client
                .get_container_metrics(&node.api_endpoint, &node.api_key, site_id)
                .await;

            let action = match result {
                Ok(metrics) => Action::AsyncOperationCompleted(
                    op_id,
                    Ok(AsyncOperationResult::Metrics(site_id, metrics)),
                ),
                Err(e) => Action::AsyncOperationCompleted(op_id, Err(e.to_string())),
            };

            let _ = tx.send(action);
        });
    }
}
