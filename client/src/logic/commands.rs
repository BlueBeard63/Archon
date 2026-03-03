use std::collections::HashMap;
use std::net::IpAddr;

use floem::reactive::{SignalGet, SignalUpdate};
use uuid::Uuid;

use crate::api::{HttpNodeClient, NodeClient};
use crate::config::{DockerCredential, FileConfigLoader};
use crate::logic::forms::{validate_required, validate_url};
use crate::models::{
    DnsProvider, DnsProviderType, Domain, DomainMapping, Node, NodeStatus, ProxyType, Site,
    SiteStatus, SiteType, Volume,
};

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

/// Update an existing node in state and save.
pub fn update_node(state: &AppState, updated: Node) {
    let node_name = updated.name.clone();

    state.nodes.update(|nodes| {
        if let Some(n) = nodes.iter_mut().find(|n| n.id == updated.id) {
            *n = updated.clone();
        }
    });

    let loader = FileConfigLoader::new();
    if let Err(e) = loader.save_node(&updated) {
        state.notifications.push(Notification::error(format!(
            "Failed to save node: {}",
            e
        )));
    } else {
        state.notifications.push(Notification::success(format!(
            "Node {} updated",
            node_name
        )));
    }

    state.navigation.navigate_back();
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

pub fn update_domain(state: &AppState, updated: Domain) {
    let domain_name = updated.name.clone();

    state.domains.update(|domains| {
        if let Some(d) = domains.iter_mut().find(|d| d.id == updated.id) {
            *d = updated.clone();
        }
    });

    if let Err(e) = state.save() {
        state.notifications.push(Notification::error(format!(
            "Failed to save: {}",
            e
        )));
    } else {
        state.notifications.push(Notification::success(format!(
            "Domain {} updated",
            domain_name
        )));
    }

    state.navigation.navigate_back();
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
// Submit commands (read FormState, validate, build model, call CRUD)
// ---------------------------------------------------------------------------

pub fn submit_create_site(state: &AppState) {
    let f = &state.form;
    let name = f.site_name.get_untracked();
    let site_type_str = f.site_type.get_untracked();
    let node_id_str = f.site_node_id.get_untracked();
    let docker_image = f.site_docker_image.get_untracked();
    let docker_cred_id_str = f.site_docker_credential_id.get_untracked();
    let ssl_email = f.site_ssl_email.get_untracked();
    let compose_content = f.site_compose_content.get_untracked();

    if let Err(e) = validate_required(&name, "Name") {
        state.notifications.push(Notification::error(e));
        return;
    }

    let node_id = match Uuid::parse_str(&node_id_str) {
        Ok(id) => id,
        Err(_) => {
            state.notifications.push(Notification::error("Please select a node"));
            return;
        }
    };

    let site_type = if site_type_str == "compose" {
        SiteType::Compose
    } else {
        SiteType::Container
    };

    if site_type == SiteType::Container {
        if let Err(e) = validate_required(&docker_image, "Docker Image") {
            state.notifications.push(Notification::error(e));
            return;
        }
    }

    // Build domain mappings
    let mappings_data = f.domain_mapping_pairs.get_untracked();
    let mut domain_mappings = Vec::new();
    let mut first_domain_id = Uuid::nil();
    let mut first_port = 80;

    for pair in &mappings_data {
        if pair.domain_id.is_empty() {
            continue;
        }
        let domain_id = match Uuid::parse_str(&pair.domain_id) {
            Ok(id) => id,
            Err(_) => continue,
        };
        let port = pair.port.parse::<i32>().unwrap_or(80);
        if first_domain_id.is_nil() {
            first_domain_id = domain_id;
            first_port = port;
        }
        domain_mappings.push(DomainMapping {
            domain_id,
            subdomain: pair.subdomain.clone(),
            port,
            host_port: 0,
        });
    }

    if domain_mappings.is_empty() {
        state
            .notifications
            .push(Notification::error("At least one domain mapping is required"));
        return;
    }

    // Build volumes
    let vol_data = f.volume_pairs.get_untracked();
    let volumes: Vec<Volume> = vol_data
        .iter()
        .filter(|v| !v.host_path.is_empty() && !v.container_path.is_empty())
        .map(|v| Volume {
            host_path: v.host_path.clone(),
            container_path: v.container_path.clone(),
        })
        .collect();

    let docker_credential_id = if docker_cred_id_str.is_empty() {
        None
    } else {
        Uuid::parse_str(&docker_cred_id_str).ok()
    };

    let now = chrono::Utc::now();
    let site = Site {
        id: Uuid::new_v4(),
        name,
        site_type,
        domain_id: first_domain_id,
        node_id,
        docker_image,
        docker_credential_id,
        docker_username: String::new(),
        docker_token: String::new(),
        compose_content,
        environment_vars: HashMap::new(),
        port: first_port,
        domain_mappings,
        ssl_enabled: !ssl_email.is_empty(),
        ssl_email,
        config_files: Vec::new(),
        volumes,
        bot_redirect_enabled: false,
        bot_redirect_url: String::new(),
        bot_user_agents: Vec::new(),
        status: SiteStatus::Inactive,
        created_at: now,
        updated_at: now,
    };

    f.reset();
    create_site(state, site);
}

pub fn submit_update_site(state: &AppState) {
    let f = &state.form;
    let site_id = match state.selected_site_id.get_untracked() {
        Some(id) => id,
        None => {
            state.notifications.push(Notification::error("No site selected"));
            return;
        }
    };

    let existing = match state.find_site(site_id) {
        Some(s) => s,
        None => {
            state.notifications.push(Notification::error("Site not found"));
            return;
        }
    };

    let name = f.site_name.get_untracked();
    let node_id_str = f.site_node_id.get_untracked();
    let docker_image = f.site_docker_image.get_untracked();
    let docker_cred_id_str = f.site_docker_credential_id.get_untracked();
    let ssl_email = f.site_ssl_email.get_untracked();
    let compose_content = f.site_compose_content.get_untracked();

    if let Err(e) = validate_required(&name, "Name") {
        state.notifications.push(Notification::error(e));
        return;
    }

    let node_id = match Uuid::parse_str(&node_id_str) {
        Ok(id) => id,
        Err(_) => {
            state.notifications.push(Notification::error("Please select a node"));
            return;
        }
    };

    // Build domain mappings
    let mappings_data = f.domain_mapping_pairs.get_untracked();
    let mut domain_mappings = Vec::new();
    let mut first_domain_id = existing.domain_id;
    let mut first_port = existing.port;

    for (i, pair) in mappings_data.iter().enumerate() {
        if pair.domain_id.is_empty() {
            continue;
        }
        let domain_id = match Uuid::parse_str(&pair.domain_id) {
            Ok(id) => id,
            Err(_) => continue,
        };
        let port = pair.port.parse::<i32>().unwrap_or(80);
        if i == 0 {
            first_domain_id = domain_id;
            first_port = port;
        }
        domain_mappings.push(DomainMapping {
            domain_id,
            subdomain: pair.subdomain.clone(),
            port,
            host_port: 0,
        });
    }

    // Build volumes
    let vol_data = f.volume_pairs.get_untracked();
    let volumes: Vec<Volume> = vol_data
        .iter()
        .filter(|v| !v.host_path.is_empty() && !v.container_path.is_empty())
        .map(|v| Volume {
            host_path: v.host_path.clone(),
            container_path: v.container_path.clone(),
        })
        .collect();

    let docker_credential_id = if docker_cred_id_str.is_empty() {
        None
    } else {
        Uuid::parse_str(&docker_cred_id_str).ok()
    };

    let updated = Site {
        id: site_id,
        name,
        site_type: existing.site_type,
        domain_id: first_domain_id,
        node_id,
        docker_image,
        docker_credential_id,
        docker_username: existing.docker_username,
        docker_token: existing.docker_token,
        compose_content,
        environment_vars: existing.environment_vars,
        port: first_port,
        domain_mappings,
        ssl_enabled: !ssl_email.is_empty(),
        ssl_email,
        config_files: existing.config_files,
        volumes,
        bot_redirect_enabled: existing.bot_redirect_enabled,
        bot_redirect_url: existing.bot_redirect_url,
        bot_user_agents: existing.bot_user_agents,
        status: existing.status,
        created_at: existing.created_at,
        updated_at: chrono::Utc::now(),
    };

    f.reset();
    update_site(state, updated);
}

pub fn submit_create_domain(state: &AppState) {
    let f = &state.form;
    let name = f.domain_name.get_untracked();
    let provider_str = f.domain_provider.get_untracked();
    let zone_id = f.domain_zone_id.get_untracked();
    let access_key = f.domain_access_key.get_untracked();
    let secret_key = f.domain_secret_key.get_untracked();

    if let Err(e) = validate_required(&name, "Domain Name") {
        state.notifications.push(Notification::error(e));
        return;
    }

    let (provider_type, api_token, zone, ak, sk) = match provider_str.as_str() {
        "cloudflare" => {
            if zone_id.is_empty() {
                state
                    .notifications
                    .push(Notification::error("Zone ID is required for Cloudflare"));
                return;
            }
            let token = state.settings.get_untracked().cloudflare_api_token.clone();
            (
                DnsProviderType::Cloudflare,
                Some(token),
                Some(zone_id),
                None,
                None,
            )
        }
        "route53" => {
            if zone_id.is_empty() {
                state
                    .notifications
                    .push(Notification::error("Zone ID is required for Route53"));
                return;
            }
            (
                DnsProviderType::Route53,
                None,
                Some(zone_id),
                if access_key.is_empty() {
                    None
                } else {
                    Some(access_key)
                },
                if secret_key.is_empty() {
                    None
                } else {
                    Some(secret_key)
                },
            )
        }
        _ => (DnsProviderType::Manual, None, None, None, None),
    };

    let provider = DnsProvider {
        provider_type,
        api_token,
        zone_id: zone,
        access_key: ak,
        secret_key: sk,
        hosted_zone_id: None,
    };

    let domain = Domain::new(name, provider);

    f.reset();
    create_domain(state, domain);
}

pub fn submit_update_domain(state: &AppState) {
    let f = &state.form;
    let domain_id = match state.selected_domain_id.get_untracked() {
        Some(id) => id,
        None => {
            state.notifications.push(Notification::error("No domain selected"));
            return;
        }
    };

    let existing = match state.find_domain(domain_id) {
        Some(d) => d,
        None => {
            state.notifications.push(Notification::error("Domain not found"));
            return;
        }
    };

    let name = f.domain_name.get_untracked();
    let provider_str = f.domain_provider.get_untracked();
    let zone_id = f.domain_zone_id.get_untracked();
    let access_key = f.domain_access_key.get_untracked();
    let secret_key = f.domain_secret_key.get_untracked();

    if let Err(e) = validate_required(&name, "Domain Name") {
        state.notifications.push(Notification::error(e));
        return;
    }

    let (provider_type, api_token, zone, ak, sk) = match provider_str.as_str() {
        "cloudflare" => {
            let token = state.settings.get_untracked().cloudflare_api_token.clone();
            (
                DnsProviderType::Cloudflare,
                Some(token),
                Some(zone_id),
                None,
                None,
            )
        }
        "route53" => (
            DnsProviderType::Route53,
            None,
            Some(zone_id),
            if access_key.is_empty() {
                None
            } else {
                Some(access_key)
            },
            if secret_key.is_empty() {
                None
            } else {
                Some(secret_key)
            },
        ),
        _ => (DnsProviderType::Manual, None, None, None, None),
    };

    let provider = DnsProvider {
        provider_type,
        api_token,
        zone_id: zone,
        access_key: ak,
        secret_key: sk,
        hosted_zone_id: existing.dns_provider.hosted_zone_id.clone(),
    };

    let updated = Domain {
        id: domain_id,
        name,
        dns_provider: provider,
        dns_records: existing.dns_records,
        traefik_enabled: existing.traefik_enabled,
        created_at: existing.created_at,
    };

    f.reset();
    update_domain(state, updated);
}

pub fn submit_create_node(state: &AppState) {
    let f = &state.form;
    let name = f.node_name.get_untracked();
    let api_endpoint = f.node_api_endpoint.get_untracked();
    let proxy_type_str = f.node_proxy_type.get_untracked();
    let api_key = f.node_api_key.get_untracked();
    let ip_str = f.node_ip_address.get_untracked();

    if let Err(e) = validate_required(&name, "Name") {
        state.notifications.push(Notification::error(e));
        return;
    }
    if let Err(e) = validate_url(&api_endpoint, "API Endpoint") {
        state.notifications.push(Notification::error(e));
        return;
    }

    let ip_address: IpAddr = if ip_str.is_empty() {
        "0.0.0.0".parse().unwrap()
    } else {
        match ip_str.parse() {
            Ok(ip) => ip,
            Err(_) => {
                state
                    .notifications
                    .push(Notification::error("Invalid IP address"));
                return;
            }
        }
    };

    let proxy_type = match proxy_type_str.as_str() {
        "nginx" => ProxyType::Nginx,
        "apache" => ProxyType::Apache,
        _ => ProxyType::Traefik,
    };

    let node = Node::new(name, api_endpoint, api_key, ip_address, proxy_type);

    f.reset();
    create_node(state, node);
}

pub fn submit_update_node(state: &AppState) {
    let f = &state.form;
    let node_id = match state.selected_node_id.get_untracked() {
        Some(id) => id,
        None => {
            state.notifications.push(Notification::error("No node selected"));
            return;
        }
    };

    let existing = match state.find_node(node_id) {
        Some(n) => n,
        None => {
            state.notifications.push(Notification::error("Node not found"));
            return;
        }
    };

    let name = f.node_name.get_untracked();
    let api_endpoint = f.node_api_endpoint.get_untracked();
    let proxy_type_str = f.node_proxy_type.get_untracked();
    let api_key = f.node_api_key.get_untracked();
    let ip_str = f.node_ip_address.get_untracked();

    if let Err(e) = validate_required(&name, "Name") {
        state.notifications.push(Notification::error(e));
        return;
    }
    if let Err(e) = validate_url(&api_endpoint, "API Endpoint") {
        state.notifications.push(Notification::error(e));
        return;
    }

    let ip_address: IpAddr = if ip_str.is_empty() {
        existing.ip_address
    } else {
        match ip_str.parse() {
            Ok(ip) => ip,
            Err(_) => {
                state
                    .notifications
                    .push(Notification::error("Invalid IP address"));
                return;
            }
        }
    };

    let proxy_type = match proxy_type_str.as_str() {
        "nginx" => ProxyType::Nginx,
        "apache" => ProxyType::Apache,
        _ => ProxyType::Traefik,
    };

    let updated = Node {
        id: node_id,
        name,
        api_endpoint,
        api_key,
        ip_address,
        proxy_type,
        status: existing.status,
        docker_info: existing.docker_info,
        traefik_info: existing.traefik_info,
        last_health_check: existing.last_health_check,
    };

    f.reset();
    update_node(state, updated);
}

pub fn submit_delete(state: &AppState) {
    let target_id = match state.form.deletion_target_id.get_untracked() {
        Some(id) => id,
        None => return,
    };
    let target_type = state.form.deletion_target_type.get_untracked();

    state.form.reset_deletion();

    match target_type.as_str() {
        "site" => {
            delete_site(state, target_id);
            state.navigation.navigate_to(Screen::SitesList);
        }
        "domain" => {
            delete_domain(state, target_id);
            state.navigation.navigate_to(Screen::DomainsList);
        }
        "node" => {
            delete_node(state, target_id);
            state.navigation.navigate_to(Screen::NodesList);
        }
        _ => {}
    }
}

pub fn submit_save_settings(state: &AppState) {
    let f = &state.form;
    let cf_token = f.settings_cloudflare_token.get_untracked();
    let r53_access = f.settings_route53_access_key.get_untracked();
    let r53_secret = f.settings_route53_secret_key.get_untracked();

    state.settings.update(|s| {
        s.cloudflare_api_token = cf_token;
        s.route53_access_key = r53_access;
        s.route53_secret_key = r53_secret;
    });

    if let Err(e) = state.save() {
        state.notifications.push(Notification::error(format!(
            "Failed to save settings: {}",
            e
        )));
    } else {
        state
            .notifications
            .push(Notification::success("Settings saved"));
    }
}

pub fn submit_create_docker_credential(state: &AppState) {
    let f = &state.form;
    let name = f.docker_cred_name.get_untracked();
    let registry = f.docker_cred_registry.get_untracked();
    let username = f.docker_cred_username.get_untracked();
    let token = f.docker_cred_token.get_untracked();

    if let Err(e) = validate_required(&name, "Name") {
        state.notifications.push(Notification::error(e));
        return;
    }
    if let Err(e) = validate_required(&username, "Username") {
        state.notifications.push(Notification::error(e));
        return;
    }

    let registry = if registry.is_empty() {
        "docker.io".to_string()
    } else {
        registry
    };

    let cred = DockerCredential {
        id: Uuid::new_v4(),
        name: name.clone(),
        registry,
        username,
        token,
    };

    state.settings.update(|s| {
        let _ = s.add_docker_credential(cred);
    });

    if let Err(e) = state.save() {
        state.notifications.push(Notification::error(format!(
            "Failed to save: {}",
            e
        )));
    } else {
        state.notifications.push(Notification::success(format!(
            "Credential {} created",
            name
        )));
    }

    f.reset();
    state.navigation.navigate_to(Screen::Settings);
}

pub fn submit_update_docker_credential(state: &AppState) {
    let f = &state.form;
    let cred_id = match state.selected_docker_credential_id.get_untracked() {
        Some(id) => id,
        None => {
            state.notifications.push(Notification::error("No credential selected"));
            return;
        }
    };

    let name = f.docker_cred_name.get_untracked();
    let registry = f.docker_cred_registry.get_untracked();
    let username = f.docker_cred_username.get_untracked();
    let token = f.docker_cred_token.get_untracked();

    if let Err(e) = validate_required(&name, "Name") {
        state.notifications.push(Notification::error(e));
        return;
    }

    let registry = if registry.is_empty() {
        "docker.io".to_string()
    } else {
        registry
    };

    let updated = DockerCredential {
        id: cred_id,
        name: name.clone(),
        registry,
        username,
        token,
    };

    state.settings.update(|s| {
        let _ = s.update_docker_credential(cred_id, updated);
    });

    if let Err(e) = state.save() {
        state.notifications.push(Notification::error(format!(
            "Failed to save: {}",
            e
        )));
    } else {
        state.notifications.push(Notification::success(format!(
            "Credential {} updated",
            name
        )));
    }

    f.reset();
    state.navigation.navigate_to(Screen::Settings);
}

pub fn submit_delete_docker_credential(state: &AppState, cred_id: Uuid) {
    state.settings.update(|s| {
        let _ = s.delete_docker_credential(cred_id);
    });

    if let Err(e) = state.save() {
        state.notifications.push(Notification::error(format!(
            "Failed to save: {}",
            e
        )));
    } else {
        state
            .notifications
            .push(Notification::success("Credential deleted"));
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
