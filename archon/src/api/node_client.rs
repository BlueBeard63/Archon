use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use uuid::Uuid;

use crate::models::node::{DockerInfo, NodeStatus, TraefikInfo};
use crate::models::site::{ConfigFile, Site, SiteStatus};
use crate::state::ContainerMetrics;

use super::client::{ApiClient, ApiResult};
use super::error::ApiError;

#[derive(Clone)]
pub struct NodeClient {
    client: ApiClient,
}

// Request/Response types
#[derive(Debug, Serialize)]
struct DeployRequest {
    name: String,
    domain: String,
    docker_image: String,
    environment_vars: HashMap<String, String>,
    port: u16,
    ssl_enabled: bool,
    config_files: Vec<ConfigFileDto>,
    traefik_labels: HashMap<String, String>,
}

#[derive(Debug, Serialize, Deserialize)]
struct ConfigFileDto {
    name: String,
    content: String,
    container_path: String,
}

#[derive(Debug, Deserialize)]
pub struct DeploymentResponse {
    pub site_id: Uuid,
    pub container_id: String,
    pub status: SiteStatus,
}

#[derive(Debug, Deserialize)]
pub struct StatusResponse {
    pub status: SiteStatus,
}

#[derive(Debug, Deserialize)]
pub struct HealthResponse {
    pub status: NodeStatus,
    pub docker: Option<DockerInfo>,
    pub traefik: Option<TraefikInfo>,
}

impl NodeClient {
    pub fn new() -> ApiResult<Self> {
        Ok(Self {
            client: ApiClient::new()?,
        })
    }

    /// Deploy a site to a node
    pub async fn deploy_site(
        &self,
        node_endpoint: &str,
        api_key: &str,
        site: &Site,
        domain_name: &str,
    ) -> ApiResult<DeploymentResponse> {
        let url = format!("{}/api/v1/sites/deploy", node_endpoint);

        let request = DeployRequest {
            name: site.name.clone(),
            domain: domain_name.to_string(),
            docker_image: site.docker_image.clone(),
            environment_vars: site.environment_vars.clone(),
            port: site.port,
            ssl_enabled: site.ssl_enabled,
            config_files: site
                .config_files
                .iter()
                .map(|cf| ConfigFileDto {
                    name: cf.name.clone(),
                    content: cf.content.clone(),
                    container_path: cf.container_path.clone(),
                })
                .collect(),
            traefik_labels: site.generate_traefik_labels(domain_name),
        };

        let response = self
            .client
            .client()
            .post(&url)
            .header("Authorization", format!("Bearer {}", api_key))
            .json(&request)
            .send()
            .await?;

        if !response.status().is_success() {
            return Err(ApiError::ServerError {
                status: response.status().as_u16(),
                message: response.text().await.unwrap_or_else(|_| "Unknown error".to_string()),
            });
        }

        response.json::<DeploymentResponse>().await.map_err(|e| {
            ApiError::InvalidResponse(format!("Failed to parse deployment response: {}", e))
        })
    }

    /// Update site configuration
    pub async fn update_site(
        &self,
        node_endpoint: &str,
        api_key: &str,
        site: &Site,
        domain_name: &str,
    ) -> ApiResult<()> {
        let url = format!("{}/api/v1/sites/{}", node_endpoint, site.id);

        let request = DeployRequest {
            name: site.name.clone(),
            domain: domain_name.to_string(),
            docker_image: site.docker_image.clone(),
            environment_vars: site.environment_vars.clone(),
            port: site.port,
            ssl_enabled: site.ssl_enabled,
            config_files: site
                .config_files
                .iter()
                .map(|cf| ConfigFileDto {
                    name: cf.name.clone(),
                    content: cf.content.clone(),
                    container_path: cf.container_path.clone(),
                })
                .collect(),
            traefik_labels: site.generate_traefik_labels(domain_name),
        };

        let response = self
            .client
            .client()
            .put(&url)
            .header("Authorization", format!("Bearer {}", api_key))
            .json(&request)
            .send()
            .await?;

        if !response.status().is_success() {
            return Err(ApiError::ServerError {
                status: response.status().as_u16(),
                message: response.text().await.unwrap_or_else(|_| "Unknown error".to_string()),
            });
        }

        Ok(())
    }

    /// Get site status
    pub async fn get_site_status(
        &self,
        node_endpoint: &str,
        api_key: &str,
        site_id: Uuid,
    ) -> ApiResult<SiteStatus> {
        let url = format!("{}/api/v1/sites/{}/status", node_endpoint, site_id);

        let response = self
            .client
            .client()
            .get(&url)
            .header("Authorization", format!("Bearer {}", api_key))
            .send()
            .await?;

        if !response.status().is_success() {
            return Err(ApiError::ServerError {
                status: response.status().as_u16(),
                message: response.text().await.unwrap_or_else(|_| "Unknown error".to_string()),
            });
        }

        let status_response = response.json::<StatusResponse>().await.map_err(|e| {
            ApiError::InvalidResponse(format!("Failed to parse status response: {}", e))
        })?;

        Ok(status_response.status)
    }

    /// Delete site from node
    pub async fn delete_site(
        &self,
        node_endpoint: &str,
        api_key: &str,
        site_id: Uuid,
    ) -> ApiResult<()> {
        let url = format!("{}/api/v1/sites/{}", node_endpoint, site_id);

        let response = self
            .client
            .client()
            .delete(&url)
            .header("Authorization", format!("Bearer {}", api_key))
            .send()
            .await?;

        if !response.status().is_success() {
            return Err(ApiError::ServerError {
                status: response.status().as_u16(),
                message: response.text().await.unwrap_or_else(|_| "Unknown error".to_string()),
            });
        }

        Ok(())
    }

    /// Health check
    pub async fn health_check(
        &self,
        node_endpoint: &str,
        api_key: &str,
    ) -> ApiResult<HealthResponse> {
        let url = format!("{}/api/v1/health", node_endpoint);

        let response = self
            .client
            .client()
            .get(&url)
            .header("Authorization", format!("Bearer {}", api_key))
            .send()
            .await?;

        if !response.status().is_success() {
            return Err(ApiError::ServerError {
                status: response.status().as_u16(),
                message: response.text().await.unwrap_or_else(|_| "Unknown error".to_string()),
            });
        }

        response.json::<HealthResponse>().await.map_err(|e| {
            ApiError::InvalidResponse(format!("Failed to parse health response: {}", e))
        })
    }

    /// Get Docker info
    pub async fn get_docker_info(
        &self,
        node_endpoint: &str,
        api_key: &str,
    ) -> ApiResult<DockerInfo> {
        let url = format!("{}/api/v1/docker/info", node_endpoint);

        let response = self
            .client
            .client()
            .get(&url)
            .header("Authorization", format!("Bearer {}", api_key))
            .send()
            .await?;

        if !response.status().is_success() {
            return Err(ApiError::ServerError {
                status: response.status().as_u16(),
                message: response.text().await.unwrap_or_else(|_| "Unknown error".to_string()),
            });
        }

        response.json::<DockerInfo>().await.map_err(|e| {
            ApiError::InvalidResponse(format!("Failed to parse Docker info: {}", e))
        })
    }

    /// Get Traefik info
    pub async fn get_traefik_info(
        &self,
        node_endpoint: &str,
        api_key: &str,
    ) -> ApiResult<TraefikInfo> {
        let url = format!("{}/api/v1/traefik/info", node_endpoint);

        let response = self
            .client
            .client()
            .get(&url)
            .header("Authorization", format!("Bearer {}", api_key))
            .send()
            .await?;

        if !response.status().is_success() {
            return Err(ApiError::ServerError {
                status: response.status().as_u16(),
                message: response.text().await.unwrap_or_else(|_| "Unknown error".to_string()),
            });
        }

        response.json::<TraefikInfo>().await.map_err(|e| {
            ApiError::InvalidResponse(format!("Failed to parse Traefik info: {}", e))
        })
    }

    /// Get container logs
    pub async fn get_container_logs(
        &self,
        node_endpoint: &str,
        api_key: &str,
        site_id: Uuid,
        lines: usize,
    ) -> ApiResult<Vec<String>> {
        let url = format!(
            "{}/api/v1/sites/{}/logs?lines={}",
            node_endpoint, site_id, lines
        );

        let response = self
            .client
            .client()
            .get(&url)
            .header("Authorization", format!("Bearer {}", api_key))
            .send()
            .await?;

        if !response.status().is_success() {
            return Err(ApiError::ServerError {
                status: response.status().as_u16(),
                message: response.text().await.unwrap_or_else(|_| "Unknown error".to_string()),
            });
        }

        #[derive(Deserialize)]
        struct LogsResponse {
            logs: Vec<String>,
        }

        let logs_response = response.json::<LogsResponse>().await.map_err(|e| {
            ApiError::InvalidResponse(format!("Failed to parse logs response: {}", e))
        })?;

        Ok(logs_response.logs)
    }

    /// Get container metrics
    pub async fn get_container_metrics(
        &self,
        node_endpoint: &str,
        api_key: &str,
        site_id: Uuid,
    ) -> ApiResult<ContainerMetrics> {
        let url = format!("{}/api/v1/sites/{}/metrics", node_endpoint, site_id);

        let response = self
            .client
            .client()
            .get(&url)
            .header("Authorization", format!("Bearer {}", api_key))
            .send()
            .await?;

        if !response.status().is_success() {
            return Err(ApiError::ServerError {
                status: response.status().as_u16(),
                message: response.text().await.unwrap_or_else(|_| "Unknown error".to_string()),
            });
        }

        response.json::<ContainerMetrics>().await.map_err(|e| {
            ApiError::InvalidResponse(format!("Failed to parse metrics response: {}", e))
        })
    }

    /// Stop a site
    pub async fn stop_site(
        &self,
        node_endpoint: &str,
        api_key: &str,
        site_id: Uuid,
    ) -> ApiResult<()> {
        let url = format!("{}/api/v1/sites/{}/stop", node_endpoint, site_id);

        let response = self
            .client
            .client()
            .post(&url)
            .header("Authorization", format!("Bearer {}", api_key))
            .send()
            .await?;

        if !response.status().is_success() {
            return Err(ApiError::ServerError {
                status: response.status().as_u16(),
                message: response.text().await.unwrap_or_else(|_| "Unknown error".to_string()),
            });
        }

        Ok(())
    }

    /// Restart a site
    pub async fn restart_site(
        &self,
        node_endpoint: &str,
        api_key: &str,
        site_id: Uuid,
    ) -> ApiResult<()> {
        let url = format!("{}/api/v1/sites/{}/restart", node_endpoint, site_id);

        let response = self
            .client
            .client()
            .post(&url)
            .header("Authorization", format!("Bearer {}", api_key))
            .send()
            .await?;

        if !response.status().is_success() {
            return Err(ApiError::ServerError {
                status: response.status().as_u16(),
                message: response.text().await.unwrap_or_else(|_| "Unknown error".to_string()),
            });
        }

        Ok(())
    }
}

impl Default for NodeClient {
    fn default() -> Self {
        Self::new().expect("Failed to create NodeClient")
    }
}
