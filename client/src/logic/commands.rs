use floem::reactive::{SignalGet, SignalUpdate};
use uuid::Uuid;

use crate::api::{HttpNodeClient, NodeClient};
use crate::config::FileConfigLoader;
use crate::models::{Node, NodeStatus, Site, SiteStatus};

use super::app_state::AppState;
use super::navigation::Screen;
use super::notifications::Notification;

// ---------------------------------------------------------------------------
// Site commands
// ---------------------------------------------------------------------------

/// Deploy a site via WebSocket with progress tracking.
pub fn deploy_site(state: &AppState, site_id: Uuid) {
    let site = match state.find_site(site_id) {
        Some(s) => s,
        None => {
            state
                .notifications
                .push(Notification::error("Site not found"));
            return;
        }
    };

    let node = match state.node_for_site(&site) {
        Some(n) => n,
        None => {
            state
                .notifications
                .push(Notification::error("Node not found for site"));
            return;
        }
    };

    let domain_name = state.domain_name_for_site(&site);

    // Update site status to deploying
    state.sites.update(|sites| {
        if let Some(s) = sites.iter_mut().find(|s| s.id == site_id) {
            s.status = SiteStatus::Deploying;
        }
    });

    state
        .notifications
        .push(Notification::info(format!("Deploying {}...", site.name)));

    let op_id = state
        .notifications
        .add_operation("deploy_site", &site.name);

    let settings = state.settings.get_untracked();
    let sites_signal = state.sites;
    let notifications = state.notifications.clone_signals();

    // Spawn async deployment
    tokio::spawn(async move {
        let client = HttpNodeClient::new();
        let result = client
            .deploy_site_with_encryption(
                &node.api_endpoint,
                &node.api_key,
                &site,
                &domain_name,
                Some(&settings),
            )
            .await;

        match result {
            Ok(()) => {
                sites_signal.update(|sites| {
                    if let Some(s) = sites.iter_mut().find(|s| s.id == site_id) {
                        s.status = SiteStatus::Running;
                    }
                });
                push_notification(&notifications, Notification::success(format!(
                    "{} deployed successfully",
                    site.name
                )));
                complete_operation(&notifications, op_id);
            }
            Err(e) => {
                sites_signal.update(|sites| {
                    if let Some(s) = sites.iter_mut().find(|s| s.id == site_id) {
                        s.status = SiteStatus::Failed;
                    }
                });
                push_notification(&notifications, Notification::error(format!(
                    "Deploy failed: {}",
                    e
                )));
                fail_operation(&notifications, op_id);
            }
        }
    });
}

/// Stop a running site.
pub fn stop_site(state: &AppState, site_id: Uuid) {
    let site = match state.find_site(site_id) {
        Some(s) => s,
        None => return,
    };
    let node = match state.node_for_site(&site) {
        Some(n) => n,
        None => return,
    };

    let sites_signal = state.sites;
    let notifications = state.notifications.clone_signals();
    let site_name = site.name.clone();

    tokio::spawn(async move {
        let client = HttpNodeClient::new();
        let result = client
            .stop_site(
                &node.api_endpoint,
                &node.api_key,
                site_id,
                &site.name,
                &site.site_type,
            )
            .await;

        match result {
            Ok(()) => {
                sites_signal.update(|sites| {
                    if let Some(s) = sites.iter_mut().find(|s| s.id == site_id) {
                        s.status = SiteStatus::Stopped;
                    }
                });
                push_notification(
                    &notifications,
                    Notification::success(format!("{} stopped", site_name)),
                );
            }
            Err(e) => {
                push_notification(
                    &notifications,
                    Notification::error(format!("Stop failed: {}", e)),
                );
            }
        }
    });
}

/// Restart a site, optionally pulling the latest image.
pub fn restart_site(state: &AppState, site_id: Uuid, pull_latest: bool) {
    let site = match state.find_site(site_id) {
        Some(s) => s,
        None => return,
    };
    let node = match state.node_for_site(&site) {
        Some(n) => n,
        None => return,
    };

    let (username, token) = resolve_credentials(state, &site);
    let sites_signal = state.sites;
    let notifications = state.notifications.clone_signals();
    let site_name = site.name.clone();

    tokio::spawn(async move {
        let client = HttpNodeClient::new();
        let result = client
            .restart_site(
                &node.api_endpoint,
                &node.api_key,
                site_id,
                pull_latest,
                &username,
                &token,
            )
            .await;

        match result {
            Ok(()) => {
                sites_signal.update(|sites| {
                    if let Some(s) = sites.iter_mut().find(|s| s.id == site_id) {
                        s.status = SiteStatus::Running;
                    }
                });
                push_notification(
                    &notifications,
                    Notification::success(format!("{} restarted", site_name)),
                );
            }
            Err(e) => {
                push_notification(
                    &notifications,
                    Notification::error(format!("Restart failed: {}", e)),
                );
            }
        }
    });
}

/// Delete a site from the node and archive locally.
pub fn delete_site(state: &AppState, site_id: Uuid) {
    let site = match state.find_site(site_id) {
        Some(s) => s,
        None => return,
    };
    let node = match state.node_for_site(&site) {
        Some(n) => n,
        None => return,
    };
    let domain_name = state.domain_name_for_site(&site);

    let sites_signal = state.sites;
    let notifications = state.notifications.clone_signals();
    let site_name = site.name.clone();
    let domain_name_clone = domain_name.clone();

    tokio::spawn(async move {
        let client = HttpNodeClient::new();
        let result = client
            .delete_site(
                &node.api_endpoint,
                &node.api_key,
                site_id,
                &domain_name,
                &site.name,
                &site.site_type,
            )
            .await;

        match result {
            Ok(()) => {
                // Archive locally
                let loader = FileConfigLoader::new();
                let _ = loader.archive_site(&site_name, &domain_name_clone, &site);

                // Remove from state
                sites_signal.update(|sites| sites.retain(|s| s.id != site_id));

                push_notification(
                    &notifications,
                    Notification::success(format!("{} deleted", site_name)),
                );
            }
            Err(e) => {
                push_notification(
                    &notifications,
                    Notification::error(format!("Delete failed: {}", e)),
                );
            }
        }
    });
}

/// Add a new site to state and save.
pub fn create_site(state: &AppState, site: Site) {
    let domain_name = state.domain_name_for_site(&site);
    let site_name = site.name.clone();

    state.sites.update(|sites| sites.push(site.clone()));

    let loader = FileConfigLoader::new();
    if let Err(e) = loader.save_site(&site, &domain_name) {
        state.notifications.push(Notification::error(format!(
            "Failed to save site: {}",
            e
        )));
    } else {
        state.notifications.push(Notification::success(format!(
            "Site {} created",
            site_name
        )));
    }

    state.navigation.navigate_to(Screen::SitesList);
}

/// Update an existing site in state and save.
pub fn update_site(state: &AppState, updated: Site) {
    let domain_name = state.domain_name_for_site(&updated);
    let site_name = updated.name.clone();

    state.sites.update(|sites| {
        if let Some(s) = sites.iter_mut().find(|s| s.id == updated.id) {
            *s = updated.clone();
        }
    });

    let loader = FileConfigLoader::new();
    if let Err(e) = loader.save_site(&updated, &domain_name) {
        state.notifications.push(Notification::error(format!(
            "Failed to save site: {}",
            e
        )));
    } else {
        state.notifications.push(Notification::success(format!(
            "Site {} updated",
            site_name
        )));
    }

    state.navigation.navigate_back();
}

// ---------------------------------------------------------------------------
// Node commands
// ---------------------------------------------------------------------------

/// Run a health check on a single node.
pub fn health_check_node(state: &AppState, node_id: Uuid) {
    let node = match state.find_node(node_id) {
        Some(n) => n,
        None => return,
    };

    let nodes_signal = state.nodes;
    let notifications = state.notifications.clone_signals();

    tokio::spawn(async move {
        let client = HttpNodeClient::new();
        let result = client
            .health_check(&node.api_endpoint, &node.api_key)
            .await;

        nodes_signal.update(|nodes| {
            if let Some(n) = nodes.iter_mut().find(|n| n.id == node_id) {
                match result {
                    Ok(health) => {
                        n.status = health.status;
                        n.docker_info = health.docker;
                        n.traefik_info = health.traefik;
                        n.last_health_check = chrono::Utc::now();
                    }
                    Err(_) => {
                        n.status = NodeStatus::Offline;
                        n.last_health_check = chrono::Utc::now();
                    }
                }
            }
        });
    });
}

/// Run health checks on all nodes.
pub fn health_check_all(state: &AppState) {
    let nodes = state.nodes.get_untracked();
    state.force_refresh_in_progress.set(true);
    state.force_refresh_total.set(nodes.len());
    state.force_refresh_completed.set(0);

    for node in nodes {
        health_check_node(state, node.id);
    }
}

/// Add a new node to state and save.
pub fn create_node(state: &AppState, node: Node) {
    let node_name = node.name.clone();
    state.nodes.update(|nodes| nodes.push(node.clone()));

    let loader = FileConfigLoader::new();
    if let Err(e) = loader.save_node(&node) {
        state.notifications.push(Notification::error(format!(
            "Failed to save node: {}",
            e
        )));
    } else {
        state.notifications.push(Notification::success(format!(
            "Node {} created",
            node_name
        )));
    }

    state.navigation.navigate_to(Screen::NodesList);
}

/// Delete a node from state and disk.
pub fn delete_node(state: &AppState, node_id: Uuid) {
    let node = match state.find_node(node_id) {
        Some(n) => n,
        None => return,
    };

    let loader = FileConfigLoader::new();
    let _ = loader.delete_node(&node.name);

    state.nodes.update(|nodes| nodes.retain(|n| n.id != node_id));

    state.notifications.push(Notification::success(format!(
        "Node {} deleted",
        node.name
    )));
}

// ---------------------------------------------------------------------------
// Domain commands
// ---------------------------------------------------------------------------

pub fn create_domain(state: &AppState, domain: crate::models::Domain) {
    let domain_name = domain.name.clone();
    state.domains.update(|domains| domains.push(domain));

    if let Err(e) = state.save() {
        state.notifications.push(Notification::error(format!(
            "Failed to save: {}",
            e
        )));
    } else {
        state.notifications.push(Notification::success(format!(
            "Domain {} created",
            domain_name
        )));
    }

    state.navigation.navigate_to(Screen::DomainsList);
}

pub fn delete_domain(state: &AppState, domain_id: Uuid) {
    let domain = match state.find_domain(domain_id) {
        Some(d) => d,
        None => return,
    };

    state
        .domains
        .update(|domains| domains.retain(|d| d.id != domain_id));

    if let Err(e) = state.save() {
        state.notifications.push(Notification::error(format!(
            "Failed to save: {}",
            e
        )));
    } else {
        state.notifications.push(Notification::success(format!(
            "Domain {} deleted",
            domain.name
        )));
    }
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

fn resolve_credentials(state: &AppState, site: &Site) -> (String, String) {
    if let Some(cred_id) = &site.docker_credential_id {
        let settings = state.settings.get_untracked();
        if let Some(cred) = settings.get_docker_credential_by_id(*cred_id) {
            return (cred.username.clone(), cred.token.clone());
        }
    }
    (site.docker_username.clone(), site.docker_token.clone())
}

/// Helper signals bundle for use in spawned tasks.
/// Since RwSignal is Copy, we can send them to async tasks.
struct NotificationSignals {
    notifications: floem::reactive::RwSignal<Vec<Notification>>,
    pending_operations: floem::reactive::RwSignal<Vec<super::notifications::AsyncOperation>>,
}

impl super::notifications::NotificationState {
    pub(crate) fn clone_signals(&self) -> NotificationSignals {
        NotificationSignals {
            notifications: self.notifications,
            pending_operations: self.pending_operations,
        }
    }
}

fn push_notification(signals: &NotificationSignals, notification: Notification) {
    signals
        .notifications
        .update(|n| n.push(notification));
}

fn complete_operation(signals: &NotificationSignals, id: Uuid) {
    signals.pending_operations.update(|ops| {
        if let Some(op) = ops.iter_mut().find(|o| o.id == id) {
            op.status = super::notifications::AsyncOpStatus::Completed;
        }
    });
}

fn fail_operation(signals: &NotificationSignals, id: Uuid) {
    signals.pending_operations.update(|ops| {
        if let Some(op) = ops.iter_mut().find(|o| o.id == id) {
            op.status = super::notifications::AsyncOpStatus::Failed;
        }
    });
}
