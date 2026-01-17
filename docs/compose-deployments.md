# Docker Compose Deployments

Archon supports deploying multi-service applications using Docker Compose files. This document explains how to use compose deployments and their differences from single-container deployments.

## Overview

Compose deployments allow you to deploy applications that consist of multiple services defined in a `docker-compose.yml` file. Archon handles:

- Writing compose content to a temporary file
- Running `docker compose up -d` with a project name
- Managing the lifecycle (start, stop, delete)
- Automatic port detection from compose YAML

## Creating a Compose Site

### Via TUI

1. Navigate to the site creation form
2. Select **"Compose"** as the site type
3. Enter a unique site name
4. Paste your `docker-compose.yml` content
5. The TUI will auto-detect exposed ports and suggest them
6. Add domain mappings with the appropriate target ports
7. Submit the form

### Required Fields

| Field | Description |
|-------|-------------|
| Name | Unique site identifier (used as project name prefix) |
| Type | Must be "Compose" |
| Compose Content | Full `docker-compose.yml` content |
| Domains | At least one domain with target port |

## Port Auto-Detection

When you enter compose content, Archon automatically parses the YAML and detects exposed ports. This helps you configure domain routing correctly.

### Supported Port Formats

**Short form:**
```yaml
ports:
  - "3000"           # Container port only
  - "8080:80"        # Host:Container
  - "127.0.0.1:80:80" # IP:Host:Container
  - "6060:6060/udp"  # With protocol
  - "8000-8005:8000-8005"  # Port range
```

**Long form:**
```yaml
ports:
  - target: 80
    published: 8080
    protocol: tcp
```

### Detection Behavior

- First detected port is used as the default for domain mappings
- All detected ports are shown in the help text
- If no ports are detected, you must manually specify the target port

## Project Naming

Compose deployments use the project name format: `archon-{site-name}`

For example, a site named `myapp` creates containers with the prefix `archon-myapp`.

You can verify this with:
```bash
docker compose -p archon-myapp ps
```

## Lifecycle Management

### Deploy
Creates or updates the compose deployment:
1. Writes compose content to temp file
2. Runs `docker compose down` (if redeploying)
3. Runs `docker compose up -d`
4. Cleans up temp files

### Stop
Stops all services without removing them:
```bash
docker compose -p archon-{name} stop
```

### Delete
Removes all services, volumes, and orphans:
```bash
docker compose -p archon-{name} down --volumes --remove-orphans
```

## Example

### Compose File
```yaml
version: "3.8"

services:
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
    depends_on:
      - app

  app:
    image: myapp:latest
    ports:
      - "3000:3000"
    environment:
      - DATABASE_URL=postgres://db:5432/myapp

  db:
    image: postgres:14
    environment:
      - POSTGRES_PASSWORD=secret
    volumes:
      - pgdata:/var/lib/postgresql/data

volumes:
  pgdata:
```

### Domain Configuration
- Domain: `myapp.example.com`
- Target Port: `80` (routes to nginx service)

The reverse proxy will route traffic from `myapp.example.com` to the nginx container's port 80, which then proxies to the app service.

## Limitations

1. **Build context**: Compose files with `build:` directives require the build context to exist on the node. Consider using pre-built images instead.

2. **Networks**: Archon doesn't manage external networks. Services can communicate within the compose project network.

3. **Health checks**: Archon checks if services are running but doesn't validate health check status.

4. **Secrets**: Docker secrets are supported but must be managed separately.

5. **Single node**: Compose deployments run on a single node. For multi-node deployments, consider Docker Swarm or Kubernetes.

## Comparison: Container vs Compose

| Feature | Container | Compose |
|---------|-----------|---------|
| Services | Single | Multiple |
| Configuration | Image + environment | Full compose YAML |
| Port detection | N/A | Automatic |
| Lifecycle | Individual container | Compose project |
| Use case | Simple apps | Multi-service apps |

## Troubleshooting

### Services not starting
Check compose logs:
```bash
docker compose -p archon-{name} logs
```

### Port detection not working
- Ensure `services:` key exists in YAML
- Check port format matches supported formats
- Verify YAML is valid

### Domain not routing
- Confirm target port matches an exposed port
- Verify the service is actually listening on that port
- Check reverse proxy configuration
