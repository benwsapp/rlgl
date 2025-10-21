# Red Light/Green Light

A lightweight dashboard for developers to display their current work-in-progress.

[![Go](https://img.shields.io/badge/go-1.25-00ADD8.svg?logo=go)](https://tip.golang.org/doc/go1.25)
[![Go Report Card](https://goreportcard.com/badge/github.com/benwsapp/rlgl)](https://goreportcard.com/report/github.com/benwsapp/rlgl)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE.md)
[![Test Status](https://github.com/benwsapp/rlgl/actions/workflows/ci.yml/badge.svg)](https://github.com/benwsapp/rlgl/actions/workflows/ci.yml)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/35f11d9745244b6ea9b4f7656bd0ed74)](https://app.codacy.com/gh/benwsapp/rlgl/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade)
[![Codacy Badge](https://app.codacy.com/project/badge/Coverage/35f11d9745244b6ea9b4f7656bd0ed74)](https://app.codacy.com/gh/benwsapp/rlgl/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_coverage)
[![golangci-lint](https://img.shields.io/badge/linted%20by-golangci--lint-brightgreen)](https://golangci-lint.run/)
[![GitHub tag](https://img.shields.io/github/v/tag/benwsapp/rlgl?color=blue)](https://github.com/benwsapp/rlgl/releases/latest)


![rlgl](https://github.com/benwsapp/rlgl/blob/main/img/screenshot.png)

**rlgl** runs in two modes:
- **Server Mode**: Hosts the dashboard web interface and receives status updates via WebSocket
- **Client Mode**: Reads your local YAML config and pushes updates to the server

## Features

- **Client/Server Architecture**: Server hosts web UI and Slack integration while client manages state
- **WebSocket Authentication**: Secure WebSocket connections with token-based authentication
- **WebSocket Communication**: Clients push config updates to the server via WebSocket
- **Real-time Updates**: Server-Sent Events (SSE) stream config changes instantly to web viewers
- **Simple Configuration**: Single YAML file to manage your work status
- **Focus Indicator**: Show what you're currently working on
- **Task Queue**: Display your upcoming tasks
- **In-Memory Storage**: Server stores client configs in memory (no persistent storage required)

## Getting Started

### Quick Install (Recommended)

```bash
$ curl -sSL https://raw.githubusercontent.com/benwsapp/rlgl/main/install.sh | bash
```

This will install the latest version to `~/.local/bin` and add it to your PATH.

**Install Options:**
```bash
# Install specific version
$ curl -sSL https://raw.githubusercontent.com/benwsapp/rlgl/main/install.sh | bash -s -- --version v1.0.0

# Install to custom directory
$ curl -sSL https://raw.githubusercontent.com/benwsapp/rlgl/main/install.sh | bash -s -- --dir /usr/local/bin

# Skip adding to PATH
$ curl -sSL https://raw.githubusercontent.com/benwsapp/rlgl/main/install.sh | bash -s -- --no-path
```

### Docker

```bash
$ docker run -p 8080:8080 benwsapp/rlgl:latest serve
```

See [Docker documentation](#docker) below for more details.

## Configuration

Create a configuration file at `config/rlgl.yaml`:

```yaml
name: "Ben's WIP Status"
description: "Current work and availability"
user: "Ben Sapp"
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

Update the YAML file anytime to change your status - the client will push updates to the server automatically!

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
# Authentication token will be auto-generated and displayed
$ ./rlgl serve

# Specify custom address
$ ./rlgl serve --addr :3000

# With pre-configured authentication token
$ ./rlgl serve --token rlgl_your_secret_token_here

# With trusted origins for CSRF (comma-separated)
$ ./rlgl serve --trusted-origins https://example.com,https://app.example.com

# Using environment variables
$ export RLGL_SERVER_ADDR=":3000"
$ export RLGL_TOKEN="rlgl_your_secret_token_here"
$ export RLGL_TRUSTED_ORIGINS="https://example.com,https://app.example.com"
$ ./rlgl serve
```

**Authentication:** The server requires a token for WebSocket connections. If you don't provide one via `--token` or `RLGL_TOKEN`, the server will generate a secure random token and display it on startup. **Save this token** - you'll need it for client connections!

### Client Mode

Run the client to push your local config to the server:

```bash
# Push config continuously (every 30s by default)
# Replace TOKEN with the value from server startup
$ ./rlgl client \
    --client-id my-laptop \
    --config config/rlgl.yaml \
    --server ws://localhost:8080/ws \
    --token rlgl_your_token_here

# Push config once and exit
$ ./rlgl client \
    --client-id my-laptop \
    --config config/rlgl.yaml \
    --server ws://localhost:8080/ws \
    --token rlgl_your_token_here \
    --once

# Custom push interval
$ ./rlgl client \
    --client-id my-laptop \
    --config config/rlgl.yaml \
    --server ws://localhost:8080/ws \
    --token rlgl_your_token_here \
    --interval 1m

# Using environment variables (defaults to rlgl.yaml in current directory)
$ export RLGL_REMOTE_HOST="ws://localhost:8080/ws"
$ export RLGL_CLIENT_ID="my-laptop"
$ export RLGL_TOKEN="rlgl_your_token_here"
$ export RLGL_CLIENT_INTERVAL="1m"
$ ./rlgl client
```

### Environment Variables

**Server:**

| Variable | Description | Default |
|----------|-------------|---------|
| `RLGL_SERVER_ADDR` | Server address | `:8080` |
| `RLGL_TOKEN` | WebSocket authentication token | Auto-generated if not provided |
| `RLGL_TRUSTED_ORIGINS` | Comma-separated list of trusted origins for CSRF protection | None |

**Client:**

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `RLGL_REMOTE_HOST` | WebSocket server URL | `ws://localhost:8080/ws` | No |
| `RLGL_CLIENT_ID` | Unique client identifier | None | Yes |
| `RLGL_TOKEN` | WebSocket authentication token | None | Yes |
| `RLGL_CLIENT_INTERVAL` | Interval between config pushes | `30s` | No |
| `RLGL_CLIENT_ONCE` | Push config once and exit | `false` | No |

### Docker

#### Run Server

```bash
# Server will generate and display authentication token on startup
$ docker run -p 8080:8080 \
    rlgl:latest serve --addr :8080

# Or use pre-configured token
$ docker run -p 8080:8080 \
    -e RLGL_TOKEN="rlgl_your_token_here" \
    rlgl:latest serve --addr :8080
```

#### Run Client

```bash
$ docker run \
    -v $(pwd)/config/rlgl.yaml:/config/rlgl.yaml:ro \
    rlgl:latest client \
        --client-id docker-client \
        --config /config/rlgl.yaml \
        --server ws://host.docker.internal:8080/ws \
        --token rlgl_your_token_here
```

#### With Environment Variables

```bash
# Server
$ docker run -p 8080:8080 \
    -e RLGL_TOKEN="rlgl_your_token_here" \
    -e RLGL_TRUSTED_ORIGINS="https://example.com,https://app.example.com" \
    rlgl:latest serve

# Client
$ docker run \
    -v $(pwd)/config:/config:ro \
    -e RLGL_REMOTE_HOST="ws://host.docker.internal:8080/ws" \
    -e RLGL_CLIENT_ID="docker-client" \
    -e RLGL_TOKEN="rlgl_your_token_here" \
    rlgl:latest client --config /config/rlgl.yaml
```

## Endpoints

**Web Interface:**
- `GET /` - Main page (renders template with first available client config)
- `GET /config` - JSON endpoint returning first available client config
- `GET /events` - Server-Sent Events stream for real-time config updates

**WebSocket API:**
- `WS /ws` - WebSocket endpoint for client connections (requires authentication via `Authorization: Bearer <token>` header)
  - Supports push config and ping/pong messages
  - Backward compatible: also accepts token via `?token=<token>` query parameter
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

## License

MIT License - see [LICENSE](LICENSE) file for details.

Copyright © 2025 Ben Sapp
