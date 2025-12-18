# Archon Node Server

The Archon Node Server is a standalone Go binary that runs on remote servers to manage Docker deployments with reverse proxy configuration.

## Features

- **Docker Integration**: Deploy and manage Docker containers
- **Multiple Reverse Proxies**: Support for Nginx, Apache, and Traefik
- **SSL Management**:
  - Manual mode: Upload your own cert/key files
  - Let's Encrypt: Automatic certificate generation with certbot
  - Traefik Auto: Let Traefik handle SSL automatically
- **REST API**: Full API for remote management
- **Health Monitoring**: Docker and proxy status endpoints

## Installation

### Build from source

```bash
cd archon-node
go build -o archon-node
```

### Install

```bash
sudo cp archon-node /usr/local/bin/
sudo chmod +x /usr/local/bin/archon-node
```

## Configuration

The node server uses a TOML configuration file. By default, it looks for `/etc/archon/node-config.toml`.

### Example Configuration - Nginx with Let's Encrypt

```toml
[server]
host = "0.0.0.0"
port = 8080
api_key = "your-secure-api-key-here"
data_dir = "/var/lib/archon"

[proxy]
type = "nginx"
config_dir = "/etc/nginx/sites-enabled"
reload_command = "nginx -s reload"

[docker]
host = "unix:///var/run/docker.sock"
network = "archon-net"

[ssl]
mode = "letsencrypt"
cert_dir = "/etc/archon/ssl"
email = "admin@example.com"

[letsencrypt]
enabled = true
email = "admin@example.com"
staging_mode = false
```

### Example Configuration - Apache with Manual SSL

```toml
[server]
host = "0.0.0.0"
port = 8080
api_key = "your-secure-api-key-here"
data_dir = "/var/lib/archon"

[proxy]
type = "apache"
config_dir = "/etc/apache2/sites-enabled"
reload_command = "apache2ctl graceful"

[docker]
host = "unix:///var/run/docker.sock"
network = "archon-net"

[ssl]
mode = "manual"
cert_dir = "/etc/archon/ssl"
email = ""
```

### Example Configuration - Traefik with Auto SSL

```toml
[server]
host = "0.0.0.0"
port = 8080
api_key = "your-secure-api-key-here"
data_dir = "/var/lib/archon"

[proxy]
type = "traefik"
config_dir = ""
reload_command = ""

[docker]
host = "unix:///var/run/docker.sock"
network = "archon-net"

[ssl]
mode = "traefik-auto"
cert_dir = ""
email = "admin@example.com"
```

## Usage

### Start the server

```bash
archon-node --config /etc/archon/node-config.toml
```

### Run as a systemd service

Create `/etc/systemd/system/archon-node.service`:

```ini
[Unit]
Description=Archon Node Server
After=network.target docker.service

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/archon-node --config /etc/archon/node-config.toml
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable archon-node
sudo systemctl start archon-node
```

## API Endpoints

All protected endpoints require an `Authorization: Bearer <api-key>` header.

### Health Check (Public)

```
GET /health
```

Returns node health status including Docker and proxy information.

### Deploy Site

```
POST /api/v1/sites/deploy
Authorization: Bearer <api-key>
Content-Type: application/json

{
  "id": "uuid",
  "name": "myapp",
  "domain": "example.com",
  "docker_image": "nginx:latest",
  "port": 8080,
  "ssl_enabled": true,
  "ssl_cert": "base64-encoded-cert",  // For manual SSL mode
  "ssl_key": "base64-encoded-key",    // For manual SSL mode
  "environment_vars": {
    "KEY": "value"
  },
  "config_files": [
    {
      "name": "app.conf",
      "content": "...",
      "container_path": "/etc/app/app.conf"
    }
  ]
}
```

### Get Site Status

```
GET /api/v1/sites/{siteID}/status
Authorization: Bearer <api-key>
```

### Stop Site

```
POST /api/v1/sites/{siteID}/stop
Authorization: Bearer <api-key>
```

### Restart Site

```
POST /api/v1/sites/{siteID}/restart
Authorization: Bearer <api-key>
```

### Delete Site

```
DELETE /api/v1/sites/{siteID}?domain=example.com
Authorization: Bearer <api-key>
```

### Get Container Logs

```
GET /api/v1/sites/{siteID}/logs
Authorization: Bearer <api-key>
```

## SSL Modes

### Manual Mode

Upload your own SSL certificates. Certificates should be base64-encoded and included in the deploy request.

```bash
# Encode certificate
base64 -w 0 cert.pem > cert.b64

# Encode key
base64 -w 0 key.pem > key.b64
```

### Let's Encrypt Mode

Automatically obtains certificates using certbot. Requires:
- Certbot installed (`apt install certbot` or `yum install certbot`)
- Port 80 available for HTTP-01 challenge
- Valid email address in configuration

### Traefik Auto Mode

Let Traefik handle SSL certificates automatically. Requires Traefik to be properly configured with Let's Encrypt.

## Requirements

### All Modes
- Docker installed and running
- Go 1.22+ (for building)

### Nginx Mode
- Nginx installed
- User must have permission to write to config directory
- User must have permission to reload Nginx

### Apache Mode
- Apache2 installed
- mod_proxy and mod_ssl enabled
- User must have permission to write to config directory
- User must have permission to reload Apache

#### Apache with Let's Encrypt

Apache can be used with Let's Encrypt for automatic SSL certificate management. This uses Certbot's webroot validation method to avoid port 80 conflicts.

**Setup Instructions:**

1. **Enable required Apache modules:**
   ```bash
   sudo a2enmod rewrite
   sudo a2enmod proxy
   sudo a2enmod proxy_http
   sudo a2enmod headers
   sudo a2enmod ssl
   ```

2. **Create webroot directory for Let's Encrypt challenges:**
   ```bash
   sudo mkdir -p /var/www/letsencrypt
   sudo chown www-data:www-data /var/www/letsencrypt
   sudo chmod 755 /var/www/letsencrypt
   ```

3. **Use the Apache + Let's Encrypt configuration file:**
   ```bash
   cp config-apache-letsencrypt-example.toml /etc/archon/node-config.toml
   sudo chown root:root /etc/archon/node-config.toml
   sudo chmod 600 /etc/archon/node-config.toml
   ```

4. **Update configuration values:**
   - Change `api_key` to a strong random value
   - Change email addresses to your actual email

5. **How it works:**
   - When deploying with SSL, Archon creates a simple HTTP vhost that serves the webroot
   - Apache is reloaded to enable this temporary configuration
   - Certbot validates domain ownership by placing a challenge file in the webroot
   - Once validated, the SSL certificate is obtained from Let's Encrypt
   - The Apache configuration is updated with the full HTTPS setup
   - Apache is reloaded with the final configuration

**Troubleshooting Apache + Let's Encrypt:**

If you see "Unable to find a virtual host listening on port 80":
- Verify Apache is running: `sudo systemctl status apache2`
- Check that port 80 is accessible: `curl -I http://your-domain.com/`
- Verify webroot directory: `ls -la /var/www/letsencrypt`
- Check logs: `sudo tail -f /var/log/apache2/error.log` and `sudo tail -f /var/log/letsencrypt/letsencrypt.log`
- Test manually: See `config-apache-letsencrypt-example.toml` for manual Certbot commands

### Traefik Mode
- Traefik running as a Docker container
- Traefik configured with Docker provider
- Docker socket mounted to Traefik container

### Let's Encrypt Mode
- Certbot installed
- Port 80 accessible from internet
- Valid domain pointing to server

## Security Notes

1. **API Key**: Generate a strong, random API key and keep it secure
2. **SSL Certificates**: Store private keys securely with restricted permissions
3. **Docker Socket**: Access to Docker socket gives root-equivalent permissions
4. **Firewall**: Ensure only necessary ports are exposed

## Troubleshooting

### Certificate errors with Let's Encrypt

```bash
# Check certbot logs
sudo tail -f /var/log/letsencrypt/letsencrypt.log

# Test with staging mode first
[letsencrypt]
staging_mode = true
```

### Nginx/Apache configuration errors

```bash
# Test nginx config
sudo nginx -t

# Test apache config
sudo apache2ctl configtest
```

### Docker connection errors

```bash
# Check Docker is running
sudo systemctl status docker

# Test Docker connection
docker ps
```

## License

MIT
