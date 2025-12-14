# Archon - Web Server Management TUI

A powerful Terminal User Interface (TUI) application for managing web servers, Docker deployments, domains, and DNS records using Rust and Ratatui.

## Features

✅ **Implemented:**
- 🚀 **Site Management** - View and navigate sites with status indicators
- 🌐 **Domain Management** - Manage domains with DNS provider integration
- 🖥️  **Node Management** - Monitor and manage server nodes
- 📊 **Dashboard** - Real-time overview of your infrastructure
- 🔐 **SSL Support** - Automatic Let's Encrypt via Traefik
- 📝 **Config Files** - Upload .toml and other config files to containers
- ⚡ **Async Operations** - Non-blocking deployment and monitoring
- 💾 **Auto-Save** - Automatic configuration persistence
- 🎨 **Beautiful UI** - Clean, intuitive terminal interface

🚧 **In Development:**
- Create/Edit/Delete screens for Sites, Domains, and Nodes
- Container logs viewer with real-time updates
- Resource metrics graphs (CPU, memory, network)
- DNS record editor with provider sync

## Architecture

Built using **The Elm Architecture (TEA)** pattern:
- **Model**: `AppState` - Single source of truth
- **Update**: `Action` enum - All state changes via actions
- **View**: `render()` - Pure rendering based on state

### Technology Stack

- **UI Framework**: Ratatui 0.29.0
- **Async Runtime**: Tokio 1.42
- **HTTP Client**: Reqwest 0.12 (with rustls)
- **Serialization**: Serde + TOML
- **Error Handling**: Anyhow + Thiserror + Color-eyre

## Installation

```bash
# Clone the repository
cd /home/bluebeard/Source/Archon/archon

# Build the application
cargo build --release

# Run the application
cargo run --release
```

## Usage

### Keybindings

#### Global
- `q` / `Ctrl+C` - Quit application
- `?` - Show help screen
- `Esc` - Go back to previous screen

#### Navigation
- `1` / `d` - Dashboard
- `2` / `s` - Sites list
- `3` - Domains list
- `4` / `n` - Nodes list

#### List Navigation (Sites, Domains, Nodes)
- `↑` / `k` - Move up
- `↓` / `j` - Move down
- `c` - Create new item
- `e` - Edit selected item
- `d` - Delete selected item
- `Enter` - View details

#### Site Management
- `r` - Redeploy site
- `s` - Stop site
- `l` - Refresh logs

#### Node Management
- `h` - Run health check
- `r` - Refresh stats

#### DNS Management
- `a` - Add DNS record
- `s` - Sync with provider (if not manual)
- `e` - Edit DNS records

## Configuration

Configuration is stored at `~/.config/archon/config.toml` (or `./archon.toml` if config dir not available).

### Example Configuration

```toml
version = "0.1.0"

[settings]
auto_save = true
health_check_interval_seconds = 300
default_dns_ttl = 300
theme = "default"

[[sites]]
id = "550e8400-e29b-41d4-a716-446655440000"
name = "my-website"
domain_id = "..."
node_id = "..."
docker_image = "nginx:latest"
port = 8080
ssl_enabled = true
status = "Running"
created_at = "2025-01-01T00:00:00Z"
updated_at = "2025-01-01T00:00:00Z"

[sites.environment_vars]
NODE_ENV = "production"

[[sites.config_files]]
name = "app.toml"
container_path = "/app/config/app.toml"
content = """
[server]
port = 8080
host = "0.0.0.0"
"""

[[domains]]
id = "..."
name = "example.com"
traefik_enabled = true
created_at = "2025-01-01T00:00:00Z"

[domains.dns_provider]
Cloudflare = { api_token = "your-token", zone_id = "your-zone-id" }

[[nodes]]
id = "..."
name = "node-1"
api_endpoint = "https://node1.example.com:8080"
api_key = "your-api-key"
ip_address = "192.168.1.10"
status = "Online"
```

## Node API Specification

Archon communicates with nodes via REST API. Each node should implement:

### Endpoints

#### Deploy Site
```
POST /api/v1/sites/deploy
Authorization: Bearer {api_key}

{
  "name": "my-site",
  "domain": "example.com",
  "docker_image": "nginx:latest",
  "environment_vars": { "KEY": "value" },
  "port": 8080,
  "ssl_enabled": true,
  "config_files": [
    {
      "name": "config.toml",
      "content": "[server]\nport = 8080",
      "container_path": "/app/config.toml"
    }
  ],
  "traefik_labels": {
    "traefik.enable": "true",
    "traefik.http.routers.site-{id}.rule": "Host(`example.com`)"
  }
}
```

#### Health Check
```
GET /api/v1/health
Authorization: Bearer {api_key}

Response:
{
  "status": "Online",
  "docker": {
    "version": "24.0.0",
    "containers_running": 5,
    "images_count": 10
  },
  "traefik": {
    "version": "2.10.0",
    "routers_count": 3,
    "services_count": 3
  }
}
```

#### Get Container Logs
```
GET /api/v1/sites/{site_id}/logs?lines=100
Authorization: Bearer {api_key}

Response:
{
  "logs": [
    "2025-01-01 00:00:00 Starting server...",
    "2025-01-01 00:00:01 Server started on port 8080"
  ]
}
```

#### Get Container Metrics
```
GET /api/v1/sites/{site_id}/metrics
Authorization: Bearer {api_key}

Response:
{
  "cpu_usage_percent": 25.5,
  "memory_usage_mb": 512,
  "memory_limit_mb": 2048,
  "network_rx_bytes": 1024000,
  "network_tx_bytes": 2048000
}
```

## DNS Providers

### Cloudflare
Requires:
- API Token with Zone:Read and DNS:Edit permissions
- Zone ID for your domain

### Route53 (Planned)
Will require:
- AWS Access Key
- AWS Secret Key
- Hosted Zone ID

### Manual DNS
No API integration. The TUI tracks DNS records locally, but you must configure them manually with your provider.

⚠️ **Warning**: When using Manual DNS, you'll see warnings reminding you to configure records manually.

## Traefik Integration

Sites are deployed with Docker labels for automatic Traefik configuration:

```
traefik.enable=true
traefik.http.routers.site-{id}.rule=Host(`example.com`)
traefik.http.routers.site-{id}.entrypoints=websecure
traefik.http.routers.site-{id}.tls=true
traefik.http.routers.site-{id}.tls.certresolver=letsencrypt
traefik.http.services.site-{id}.loadbalancer.server.port=8080
```

## Project Structure

```
archon/
├── src/
│   ├── main.rs              # Entry point
│   ├── app.rs               # Main app logic (TEA Update)
│   ├── api/                 # Node REST API client
│   ├── config/              # Configuration loading/saving
│   ├── dns/                 # DNS provider integrations
│   ├── events/              # Event handling and Actions
│   ├── models/              # Data models
│   ├── state/               # Application state
│   └── ui/                  # All UI components
│       ├── components/      # Reusable UI components
│       ├── screens/         # Screen implementations
│       ├── render.rs        # Main render function (TEA View)
│       ├── router.rs        # Screen enum
│       └── theme.rs         # Color theme
├── Cargo.toml
└── README.md
```

## Development

### Build
```bash
cargo build
```

### Run
```bash
cargo run
```

### Test
```bash
cargo test
```

### Format
```bash
cargo fmt
```

### Lint
```bash
cargo clippy
```

## Contributing

This is a personal project, but suggestions and improvements are welcome!

## License

MIT License - See LICENSE file for details

## Roadmap

- [ ] Complete create/edit/delete screens
- [ ] Real-time log streaming
- [ ] Resource usage graphs
- [ ] Multi-node deployment
- [ ] Site templates
- [ ] Backup and restore
- [ ] Docker Compose support
- [ ] Kubernetes integration
- [ ] Plugin system

## Acknowledgments

- Built with [Ratatui](https://ratatui.rs/) - Rust library for cooking up terminal user interfaces
- Inspired by [k9s](https://k9scli.io/) and [lazydocker](https://github.com/jesseduffield/lazydocker)
- Uses [Traefik](https://traefik.io/) for reverse proxy and automatic SSL
