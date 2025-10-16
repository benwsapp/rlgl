# Red Light/Green Light

![rlgl](https://github.com/benwsapp/rlgl/blob/main/img/screenshot.png)

A lightweight status dashboard for developers to showcase their current work-in-progress. Display what you're focusing on, your task queue, and availability status in real-time. Perfect for solo developers, remote teams, or anyone who wants a simple way to broadcast "what I'm working on right now."

**rlgl** runs in two modes:
- **Server Mode**: Hosts the dashboard web interface and receives status updates via WebSocket
- **Client Mode**: Reads your local YAML config and pushes updates to the server

[![Go](https://img.shields.io/badge/go-1.25-00ADD8.svg?logo=go)](https://tip.golang.org/doc/go1.25)
[![Go Report Card](https://goreportcard.com/badge/github.com/benwsapp/rlgl)](https://goreportcard.com/report/github.com/benwsapp/rlgl)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE.md)
[![Test Status](https://github.com/benwsapp/rlgl/actions/workflows/ci.yml/badge.svg)](https://github.com/benwsapp/rlgl/actions/workflows/ci.yml)
[![Coverage](https://codecov.io/gh/benwsapp/rlgl/graph/badge.svg?token=XDXHMQEK2L)](https://codecov.io/gh/benwsapp/rlgl)

## Features

- **Client/Server Architecture**: Run a central server and push updates from multiple clients
- **WebSocket Communication**: Clients push config updates to the server via WebSocket
- **Real-time Updates**: Server-Sent Events (SSE) stream config changes instantly to web viewers
- **Simple Configuration**: Single YAML file to manage your work status
- **Focus Indicator**: Show what you're currently working on
- **Task Queue**: Display your upcoming tasks
- **In-Memory Storage**: Server stores client configs in memory (no database required)

## Prerequisites

- Go 1.25.1 or later
- Docker (optional, for containerized deployment)

## Configuration

Create a configuration file at `config/site.yaml`:

```yaml
name: "Ben's WIP Status"
description: "Current work and availability"
user: "bensapp"
contributor:
  active: true
  focus: "Implementing CSRF protection for rlgl"
  queue:
    - "Add Docker multi-arch support"
    - "Write comprehensive tests"
    - "Update documentation"
```

The config file structure:
- `name`: Your status page title
- `description`: Brief description of the page
- `user`: Your username or identifier
- `contributor.active`: Status indicator (true = green light/available, false = red light/busy)
- `contributor.focus`: What you're currently working on
- `contributor.queue`: Your upcoming tasks/backlog

Update the YAML file anytime to change your status - the web page will update automatically via SSE!

## Building from Source

### Go Build

```bash
# Clone the repository
$ git clone https://github.com/benwsapp/rlgl.git
$ cd rlgl

# Download dependencies
$ go mod download

# Build the binary
$ go build -o rlgl .

# Run with config file
$ ./rlgl run --config config/site.yaml --addr :8080
```

### Docker Build

#### Single Architecture

```bash
# Build for current platform
$ docker build -t rlgl:latest .
```

#### Multi-Architecture (using buildx)

```bash
# Create a new builder (first time only)
$ docker buildx create --name multiarch --use

# Build for multiple architectures
$ docker buildx build \
    --platform linux/amd64,linux/arm64 \
    -t your-registry/rlgl:latest \
    --push .
```

## Running

### Server Mode

Run the server to host the dashboard and receive updates from clients:

```bash
# Run server with defaults (listens on :8080)
$ ./rlgl serve

# Specify custom address
$ ./rlgl serve --addr :3000

# With trusted origins for CSRF (comma-separated)
$ ./rlgl serve --trusted-origins https://example.com,https://app.example.com

# Using environment variables
$ export RLGL_SERVER_ADDR=":3000"
$ export RLGL_TRUSTED_ORIGINS="https://example.com,https://app.example.com"
$ ./rlgl serve
```

### Client Mode

Run the client to push your local config to the server:

```bash
# Push config continuously (every 30s by default)
$ ./rlgl client --client-id my-laptop --config config/site.yaml --server ws://localhost:8080/ws

# Push config once and exit
$ ./rlgl client --client-id my-laptop --config config/site.yaml --server ws://localhost:8080/ws --once

# Custom push interval
$ ./rlgl client --client-id my-laptop --config config/site.yaml --server ws://localhost:8080/ws --interval 1m

# Using environment variables
$ export RLGL_REMOTE_HOST="ws://localhost:8080/ws"
$ export RLGL_CLIENT_ID="my-laptop"
$ export RLGL_CLIENT_INTERVAL="1m"
$ ./rlgl client --config config/site.yaml
```

### Environment Variables

**Server:**
- `RLGL_SERVER_ADDR` - Server address (default: `:8080`)
- `RLGL_TRUSTED_ORIGINS` - Comma-separated list of trusted origins for CSRF protection

**Client:**
- `RLGL_REMOTE_HOST` - WebSocket server URL (default: `ws://localhost:8080/ws`)
- `RLGL_CLIENT_ID` - Unique client identifier (required)
- `RLGL_CLIENT_INTERVAL` - Interval between config pushes (default: `30s`)
- `RLGL_CLIENT_ONCE` - Push config once and exit (default: `false`)

### Docker

#### Run Server

```bash
$ docker run -p 8080:8080 \
    rlgl:latest serve --addr :8080
```

#### Run Client

```bash
$ docker run \
    -v $(pwd)/config/site.yaml:/config/site.yaml:ro \
    rlgl:latest client --client-id docker-client --config /config/site.yaml --server ws://host.docker.internal:8080/ws
```

#### With Environment Variables

```bash
# Server
$ docker run -p 8080:8080 \
    -e RLGL_TRUSTED_ORIGINS="https://example.com,https://app.example.com" \
    rlgl:latest serve

# Client
$ docker run \
    -v $(pwd)/config:/config:ro \
    -e RLGL_REMOTE_HOST="ws://host.docker.internal:8080/ws" \
    -e RLGL_CLIENT_ID="docker-client" \
    rlgl:latest client --config /config/site.yaml
```

## Endpoints

**Web Interface:**
- `GET /` - Main page (renders template with first available client config)
- `GET /config` - JSON endpoint returning first available client config
- `GET /events` - Server-Sent Events stream for real-time config updates

**WebSocket API:**
- `WS /ws` - WebSocket endpoint for client connections (push config, ping/pong)
- `GET /status` - JSON endpoint returning all client configs (keyed by client ID)

## Development

### Linting

```bash
# Run all linters
$ golangci-lint run

# Check cognitive complexity
$ gocognit -d -over 10 .

# Lint Dockerfile
$ hadolint Dockerfile
```

### Testing

```bash
# Run tests
$ go test ./...

# Run tests with coverage
$ go test -cover ./...
```

## Security

This project implements defense-in-depth CSRF protection:
- Go 1.25's `http.CrossOriginProtection` middleware
- Security headers (X-Content-Type-Options, X-Frame-Options, CSP, etc.)
- SameSite cookies (Lax/Strict modes)
- Runs as non-root user (uid 65532) in Docker
- Read-only filesystem in container

## License

MIT License - see [LICENSE](LICENSE) file for details.

Copyright © 2025 Ben Sapp
