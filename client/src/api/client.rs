use futures_util::{SinkExt, StreamExt};
use reqwest::header::{HeaderMap, HeaderValue, AUTHORIZATION, CONTENT_TYPE};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::time::Duration;
use thiserror::Error;
use tokio_tungstenite::{connect_async, tungstenite};
use url::Url;
use uuid::Uuid;

use crate::config::Settings;
use crate::crypto;
use crate::models::{
    ConfigFile, DockerInfo, NodeStatus, Site, SiteStatus, SiteType, TraefikInfo,
};

// ---------------------------------------------------------------------------
// Error type
// ---------------------------------------------------------------------------

#[derive(Debug, Error)]
pub enum ApiError {
    #[error("HTTP error: {0}")]
    Http(#[from] reqwest::Error),
    #[error("API error ({status}): {message}")]
    Api { status: u16, message: String },
    #[error("WebSocket error: {0}")]
    WebSocket(String),
    #[error("encryption error: {0}")]
    Encryption(String),
    #[error("deployment failed: {0}")]
    DeploymentFailed(String),
    #[error("URL parse error: {0}")]
    UrlParse(#[from] url::ParseError),
    #[error("JSON error: {0}")]
    Json(#[from] serde_json::Error),
    #[error("{0}")]
    Other(String),
}

// ---------------------------------------------------------------------------
// API types (wire format)
// ---------------------------------------------------------------------------

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DeploymentMessage {
    #[serde(rename = "type")]
    pub msg_type: String,
    pub message: String,
    pub step: String,
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub error: String,
}

pub type DeploymentProgressCallback = Box<dyn Fn(DeploymentMessage) + Send>;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DockerCredentials {
    pub username: String,
    pub password: String,
    #[serde(default, skip_serializing_if = "is_false")]
    pub encrypted: bool,
}

fn is_false(v: &bool) -> bool {
    !v
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Docker {
    pub image: String,
    pub credentials: DockerCredentials,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HealthResponse {
    pub status: NodeStatus,
    pub docker: Option<DockerInfo>,
    pub traefik: Option<TraefikInfo>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ContainerMetrics {
    pub cpu_percent: f64,
    pub memory_usage: i64,
    pub memory_limit: i64,
    pub network_rx_bytes: i64,
    pub network_tx_bytes: i64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DeployResponse {
    pub site_id: Uuid,
    pub container_id: String,
    pub status: String,
    pub message: String,
}

/// Domain mapping as sent to the node API (different from the model's DomainMapping).
#[derive(Debug, Clone, Serialize, Deserialize)]
struct NodeDomainMapping {
    domain: String,
    port: i32,
    #[serde(skip_serializing_if = "is_zero")]
    host_port: i32,
}

fn is_zero(v: &i32) -> bool {
    *v == 0
}

/// Volume as sent to the node API.
#[derive(Debug, Clone, Serialize, Deserialize)]
struct NodeVolume {
    host_path: String,
    container_path: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct SiteStatusResponse {
    site_id: Uuid,
    status: SiteStatus,
    #[serde(default)]
    container_id: String,
    #[serde(default)]
    is_running: bool,
    #[serde(default)]
    message: String,
}

/// The deploy request payload sent to the node.
#[derive(Debug, Serialize)]
struct DeployRequest {
    id: Uuid,
    name: String,
    site_type: SiteType,
    docker: Docker,
    #[serde(skip_serializing_if = "String::is_empty")]
    compose_content: String,
    environment_vars: HashMap<String, String>,
    domain_mappings: Vec<NodeDomainMapping>,
    ssl_enabled: bool,
    #[serde(skip_serializing_if = "String::is_empty")]
    ssl_email: String,
    config_files: Vec<ConfigFile>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    volumes: Vec<NodeVolume>,
    #[serde(skip_serializing_if = "HashMap::is_empty")]
    traefik_labels: HashMap<String, String>,
    #[serde(skip_serializing_if = "std::ops::Not::not")]
    bot_redirect_enabled: bool,
    #[serde(skip_serializing_if = "String::is_empty")]
    bot_redirect_url: String,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    bot_user_agents: Vec<String>,
}

// ---------------------------------------------------------------------------
// NodeClient trait
// ---------------------------------------------------------------------------

#[allow(async_fn_in_trait)]
pub trait NodeClient: Send + Sync {
    async fn deploy_site(
        &self,
        endpoint: &str,
        api_key: &str,
        site: &Site,
        domain_name: &str,
    ) -> Result<(), ApiError>;

    async fn deploy_site_with_encryption(
        &self,
        endpoint: &str,
        api_key: &str,
        site: &Site,
        domain_name: &str,
        settings: Option<&Settings>,
    ) -> Result<(), ApiError>;

    async fn deploy_site_websocket(
        &self,
        endpoint: &str,
        api_key: &str,
        site: &Site,
        domain_name: &str,
        settings: Option<&Settings>,
        progress_callback: DeploymentProgressCallback,
    ) -> Result<(), ApiError>;

    async fn update_site(
        &self,
        endpoint: &str,
        api_key: &str,
        site_id: Uuid,
        docker_username: &str,
        docker_token: &str,
    ) -> Result<(), ApiError>;

    async fn delete_site(
        &self,
        endpoint: &str,
        api_key: &str,
        site_id: Uuid,
        domain: &str,
        site_name: &str,
        site_type: &SiteType,
    ) -> Result<(), ApiError>;

    async fn get_site_status(
        &self,
        endpoint: &str,
        api_key: &str,
        site_id: Uuid,
        site_name: &str,
        site_type: &SiteType,
    ) -> Result<SiteStatus, ApiError>;

    async fn stop_site(
        &self,
        endpoint: &str,
        api_key: &str,
        site_id: Uuid,
        site_name: &str,
        site_type: &SiteType,
    ) -> Result<(), ApiError>;

    async fn restart_site(
        &self,
        endpoint: &str,
        api_key: &str,
        site_id: Uuid,
        pull_latest: bool,
        docker_username: &str,
        docker_token: &str,
    ) -> Result<(), ApiError>;

    async fn health_check(
        &self,
        endpoint: &str,
        api_key: &str,
    ) -> Result<HealthResponse, ApiError>;

    async fn get_docker_info(
        &self,
        endpoint: &str,
        api_key: &str,
    ) -> Result<DockerInfo, ApiError>;

    async fn get_traefik_info(
        &self,
        endpoint: &str,
        api_key: &str,
    ) -> Result<TraefikInfo, ApiError>;

    async fn get_container_logs(
        &self,
        endpoint: &str,
        api_key: &str,
        site_id: Uuid,
        lines: i32,
    ) -> Result<Vec<String>, ApiError>;

    async fn get_container_metrics(
        &self,
        endpoint: &str,
        api_key: &str,
        site_id: Uuid,
    ) -> Result<ContainerMetrics, ApiError>;
}

// ---------------------------------------------------------------------------
// HTTPNodeClient
// ---------------------------------------------------------------------------

pub struct HttpNodeClient {
    client: reqwest::Client,
}

impl HttpNodeClient {
    pub fn new() -> Self {
        let client = reqwest::Client::builder()
            .timeout(Duration::from_secs(30))
            .pool_max_idle_per_host(10)
            .pool_idle_timeout(Duration::from_secs(90))
            .build()
            .expect("failed to build HTTP client");

        Self { client }
    }

    async fn do_request(
        &self,
        method: reqwest::Method,
        url: &str,
        api_key: &str,
        body: Option<&impl Serialize>,
    ) -> Result<reqwest::Response, ApiError> {
        let mut headers = HeaderMap::new();
        if !api_key.is_empty() {
            headers.insert(
                AUTHORIZATION,
                HeaderValue::from_str(&format!("Bearer {}", api_key))
                    .map_err(|e| ApiError::Other(e.to_string()))?,
            );
        }

        let mut builder = self.client.request(method, url).headers(headers);

        if let Some(b) = body {
            builder = builder
                .header(CONTENT_TYPE, "application/json")
                .json(b);
        }

        let resp = builder.send().await?;

        if resp.status().is_client_error() || resp.status().is_server_error() {
            let status = resp.status().as_u16();
            let text = resp.text().await.unwrap_or_default();

            // Try to parse structured error
            #[derive(Deserialize)]
            struct ErrBody {
                #[serde(default)]
                error: String,
                #[serde(default)]
                message: String,
            }

            if let Ok(err_body) = serde_json::from_str::<ErrBody>(&text) {
                let msg = if !err_body.message.is_empty() {
                    err_body.message
                } else if !err_body.error.is_empty() {
                    err_body.error
                } else {
                    text
                };
                return Err(ApiError::Api {
                    status,
                    message: msg,
                });
            }

            return Err(ApiError::Api {
                status,
                message: text,
            });
        }

        Ok(resp)
    }
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

fn resolve_docker_credentials(site: &Site, settings: Option<&Settings>) -> (String, String) {
    if let Some(cred_id) = &site.docker_credential_id {
        if let Some(s) = settings {
            if let Some(cred) = s.get_docker_credential_by_id(*cred_id) {
                return (cred.username.clone(), cred.token.clone());
            }
        }
    }
    (site.docker_username.clone(), site.docker_token.clone())
}

fn encrypt_credentials(
    username: &str,
    token: &str,
    api_key: &str,
) -> Result<DockerCredentials, ApiError> {
    if username.is_empty() && token.is_empty() {
        return Ok(DockerCredentials {
            username: String::new(),
            password: String::new(),
            encrypted: false,
        });
    }

    let enc_user = crypto::encrypt(username, api_key)
        .map_err(|e| ApiError::Encryption(format!("username: {}", e)))?;
    let enc_pass = crypto::encrypt(token, api_key)
        .map_err(|e| ApiError::Encryption(format!("password: {}", e)))?;

    Ok(DockerCredentials {
        username: enc_user,
        password: enc_pass,
        encrypted: true,
    })
}

fn convert_to_node_domain_mappings(site: &Site, domain_name: &str) -> Vec<NodeDomainMapping> {
    let mut base_domain = domain_name.to_string();

    for mapping in &site.domain_mappings {
        if !mapping.subdomain.is_empty()
            && domain_name.starts_with(&format!("{}.", mapping.subdomain))
        {
            base_domain = domain_name
                .strip_prefix(&format!("{}.", mapping.subdomain))
                .unwrap_or(domain_name)
                .to_string();
            break;
        }
    }

    site.domain_mappings
        .iter()
        .map(|m| {
            let full_domain = if m.subdomain.is_empty() {
                base_domain.clone()
            } else {
                format!("{}.{}", m.subdomain, base_domain)
            };
            NodeDomainMapping {
                domain: full_domain,
                port: m.port,
                host_port: m.host_port,
            }
        })
        .collect()
}

fn convert_to_node_volumes(site: &Site) -> Vec<NodeVolume> {
    site.volumes
        .iter()
        .map(|v| NodeVolume {
            host_path: v.host_path.clone(),
            container_path: v.container_path.clone(),
        })
        .collect()
}

fn build_deploy_request(
    site: &Site,
    domain_name: &str,
    creds: DockerCredentials,
) -> DeployRequest {
    DeployRequest {
        id: site.id,
        name: site.name.clone(),
        site_type: site.site_type.clone(),
        docker: Docker {
            image: site.docker_image.clone(),
            credentials: creds,
        },
        compose_content: site.compose_content.clone(),
        environment_vars: site.environment_vars.clone(),
        domain_mappings: convert_to_node_domain_mappings(site, domain_name),
        ssl_enabled: site.ssl_enabled,
        ssl_email: site.ssl_email.clone(),
        config_files: site.config_files.clone(),
        volumes: convert_to_node_volumes(site),
        traefik_labels: site.generate_traefik_labels(domain_name),
        bot_redirect_enabled: site.bot_redirect_enabled,
        bot_redirect_url: site.bot_redirect_url.clone(),
        bot_user_agents: site.bot_user_agents.clone(),
    }
}

fn convert_to_websocket_url(endpoint: &str, path: &str) -> Result<String, ApiError> {
    let mut u = Url::parse(endpoint)?;
    match u.scheme() {
        "http" => u.set_scheme("ws").unwrap(),
        "https" => u.set_scheme("wss").unwrap(),
        _ => u.set_scheme("ws").unwrap(),
    };
    let current_path = u.path().trim_end_matches('/').to_string();
    u.set_path(&format!("{}{}", current_path, path));
    Ok(u.to_string())
}

// ---------------------------------------------------------------------------
// NodeClient implementation
// ---------------------------------------------------------------------------

impl NodeClient for HttpNodeClient {
    async fn deploy_site(
        &self,
        endpoint: &str,
        api_key: &str,
        site: &Site,
        domain_name: &str,
    ) -> Result<(), ApiError> {
        self.deploy_site_with_encryption(endpoint, api_key, site, domain_name, None)
            .await
    }

    async fn deploy_site_with_encryption(
        &self,
        endpoint: &str,
        api_key: &str,
        site: &Site,
        domain_name: &str,
        settings: Option<&Settings>,
    ) -> Result<(), ApiError> {
        let (username, token) = resolve_docker_credentials(site, settings);
        let creds = encrypt_credentials(&username, &token, api_key)?;
        let req = build_deploy_request(site, domain_name, creds);

        let url = format!("{}/api/v1/sites/deploy", endpoint);
        self.do_request(reqwest::Method::POST, &url, api_key, Some(&req))
            .await?;
        Ok(())
    }

    async fn deploy_site_websocket(
        &self,
        endpoint: &str,
        api_key: &str,
        site: &Site,
        domain_name: &str,
        settings: Option<&Settings>,
        progress_callback: DeploymentProgressCallback,
    ) -> Result<(), ApiError> {
        let ws_url = convert_to_websocket_url(endpoint, "/api/v1/sites/deploy/ws")?;

        let (username, token) = resolve_docker_credentials(site, settings);
        let creds = encrypt_credentials(&username, &token, api_key)?;
        let req = build_deploy_request(site, domain_name, creds);

        // Build WebSocket request with auth header
        let ws_request = tungstenite::http::Request::builder()
            .uri(&ws_url)
            .header("Authorization", format!("Bearer {}", api_key))
            .body(())
            .map_err(|e| ApiError::WebSocket(e.to_string()))?;

        let (ws_stream, _) = connect_async(ws_request)
            .await
            .map_err(|e| ApiError::WebSocket(e.to_string()))?;

        let (mut write, mut read) = ws_stream.split();

        // Send deploy request as first message
        let req_json = serde_json::to_string(&req)?;
        write
            .send(tungstenite::Message::Text(req_json))
            .await
            .map_err(|e| ApiError::WebSocket(e.to_string()))?;

        // Listen for progress messages
        while let Some(msg_result) = read.next().await {
            let msg = msg_result.map_err(|e| ApiError::WebSocket(e.to_string()))?;

            match msg {
                tungstenite::Message::Text(text) => {
                    let deploy_msg: DeploymentMessage = serde_json::from_str(&text)?;

                    progress_callback(deploy_msg.clone());

                    match deploy_msg.msg_type.as_str() {
                        "success" => return Ok(()),
                        "error" => {
                            let err_msg = if !deploy_msg.error.is_empty() {
                                deploy_msg.error
                            } else {
                                deploy_msg.message
                            };
                            return Err(ApiError::DeploymentFailed(err_msg));
                        }
                        _ => continue,
                    }
                }
                tungstenite::Message::Close(_) => break,
                _ => continue,
            }
        }

        Err(ApiError::WebSocket(
            "connection closed before deployment completed".into(),
        ))
    }

    async fn update_site(
        &self,
        endpoint: &str,
        api_key: &str,
        site_id: Uuid,
        docker_username: &str,
        docker_token: &str,
    ) -> Result<(), ApiError> {
        let url = format!("{}/api/v1/sites/{}/update", endpoint, site_id);

        let mut body = serde_json::Map::new();
        body.insert(
            "docker_username".into(),
            serde_json::Value::String(docker_username.into()),
        );
        body.insert(
            "docker_token".into(),
            serde_json::Value::String(docker_token.into()),
        );

        if !api_key.is_empty() && (!docker_username.is_empty() || !docker_token.is_empty()) {
            let enc_user = crypto::encrypt(docker_username, api_key)
                .map_err(|e| ApiError::Encryption(e.to_string()))?;
            let enc_token = crypto::encrypt(docker_token, api_key)
                .map_err(|e| ApiError::Encryption(e.to_string()))?;
            body.insert(
                "docker_username".into(),
                serde_json::Value::String(enc_user),
            );
            body.insert(
                "docker_token".into(),
                serde_json::Value::String(enc_token),
            );
            body.insert("credentials_encrypted".into(), serde_json::Value::Bool(true));
        }

        let body_val = serde_json::Value::Object(body);
        self.do_request(reqwest::Method::POST, &url, api_key, Some(&body_val))
            .await?;
        Ok(())
    }

    async fn delete_site(
        &self,
        endpoint: &str,
        api_key: &str,
        site_id: Uuid,
        domain: &str,
        site_name: &str,
        site_type: &SiteType,
    ) -> Result<(), ApiError> {
        let mut url = format!(
            "{}/api/v1/sites/{}?domain={}",
            endpoint, site_id, domain
        );

        if *site_type == SiteType::Compose && !site_name.is_empty() {
            url = format!("{}&type=compose&name={}", url, site_name);
        }

        self.do_request(reqwest::Method::DELETE, &url, api_key, None::<&()>.as_ref())
            .await?;
        Ok(())
    }

    async fn get_site_status(
        &self,
        endpoint: &str,
        api_key: &str,
        site_id: Uuid,
        site_name: &str,
        site_type: &SiteType,
    ) -> Result<SiteStatus, ApiError> {
        let mut url = format!("{}/api/v1/sites/{}/status", endpoint, site_id);

        if *site_type == SiteType::Compose && !site_name.is_empty() {
            url = format!("{}?type=compose&name={}", url, site_name);
        }

        let resp = self
            .do_request(reqwest::Method::GET, &url, api_key, None::<&()>.as_ref())
            .await?;
        let status_resp: SiteStatusResponse = resp.json().await?;
        Ok(status_resp.status)
    }

    async fn stop_site(
        &self,
        endpoint: &str,
        api_key: &str,
        site_id: Uuid,
        site_name: &str,
        site_type: &SiteType,
    ) -> Result<(), ApiError> {
        let mut url = format!("{}/api/v1/sites/{}/stop", endpoint, site_id);

        if *site_type == SiteType::Compose && !site_name.is_empty() {
            url = format!("{}?type=compose&name={}", url, site_name);
        }

        self.do_request(reqwest::Method::POST, &url, api_key, None::<&()>.as_ref())
            .await?;
        Ok(())
    }

    async fn restart_site(
        &self,
        endpoint: &str,
        api_key: &str,
        site_id: Uuid,
        pull_latest: bool,
        docker_username: &str,
        docker_token: &str,
    ) -> Result<(), ApiError> {
        let url = format!("{}/api/v1/sites/{}/restart", endpoint, site_id);

        if !pull_latest {
            self.do_request(reqwest::Method::POST, &url, api_key, None::<&()>.as_ref())
                .await?;
            return Ok(());
        }

        let mut body = serde_json::Map::new();
        body.insert("pull_latest".into(), serde_json::Value::Bool(true));

        if !api_key.is_empty() && (!docker_username.is_empty() || !docker_token.is_empty()) {
            let enc_user = crypto::encrypt(docker_username, api_key)
                .map_err(|e| ApiError::Encryption(e.to_string()))?;
            let enc_token = crypto::encrypt(docker_token, api_key)
                .map_err(|e| ApiError::Encryption(e.to_string()))?;
            body.insert(
                "docker_username".into(),
                serde_json::Value::String(enc_user),
            );
            body.insert(
                "docker_token".into(),
                serde_json::Value::String(enc_token),
            );
            body.insert("credentials_encrypted".into(), serde_json::Value::Bool(true));
        } else if !docker_username.is_empty() || !docker_token.is_empty() {
            body.insert(
                "docker_username".into(),
                serde_json::Value::String(docker_username.into()),
            );
            body.insert(
                "docker_token".into(),
                serde_json::Value::String(docker_token.into()),
            );
        }

        let body_val = serde_json::Value::Object(body);
        self.do_request(reqwest::Method::POST, &url, api_key, Some(&body_val))
            .await?;
        Ok(())
    }

    async fn health_check(
        &self,
        endpoint: &str,
        api_key: &str,
    ) -> Result<HealthResponse, ApiError> {
        let url = format!("{}/health", endpoint);
        let resp = self
            .do_request(reqwest::Method::GET, &url, api_key, None::<&()>.as_ref())
            .await?;
        let health: HealthResponse = resp.json().await?;
        Ok(health)
    }

    async fn get_docker_info(
        &self,
        endpoint: &str,
        api_key: &str,
    ) -> Result<DockerInfo, ApiError> {
        let health = self.health_check(endpoint, api_key).await?;
        health
            .docker
            .ok_or_else(|| ApiError::Other("no Docker info in health response".into()))
    }

    async fn get_traefik_info(
        &self,
        endpoint: &str,
        api_key: &str,
    ) -> Result<TraefikInfo, ApiError> {
        let health = self.health_check(endpoint, api_key).await?;
        health
            .traefik
            .ok_or_else(|| ApiError::Other("no Traefik info in health response".into()))
    }

    async fn get_container_logs(
        &self,
        endpoint: &str,
        api_key: &str,
        site_id: Uuid,
        _lines: i32,
    ) -> Result<Vec<String>, ApiError> {
        let url = format!("{}/api/v1/sites/{}/logs", endpoint, site_id);
        let resp = self
            .do_request(reqwest::Method::GET, &url, api_key, None::<&()>.as_ref())
            .await?;

        #[derive(Deserialize)]
        struct LogsResponse {
            logs: Vec<String>,
        }

        let logs_resp: LogsResponse = resp.json().await?;
        Ok(logs_resp.logs)
    }

    async fn get_container_metrics(
        &self,
        _endpoint: &str,
        _api_key: &str,
        _site_id: Uuid,
    ) -> Result<ContainerMetrics, ApiError> {
        Err(ApiError::Other(
            "metrics endpoint not yet implemented".into(),
        ))
    }
}
