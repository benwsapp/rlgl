# Red Light/Green Light

![rlgl](https://github.com/benwsapp/rlgl/blob/main/img/screenshot.png)

A lightweight status dashboard for developers to showcase their current work-in-progress. Display what you're focusing on, your task queue, and availability status in real-time. Perfect for solo developers, remote teams, or anyone who wants a simple way to broadcast "what I'm working on right now."

[![Go](https://img.shields.io/badge/go-1.25-00ADD8.svg?logo=go)](https://tip.golang.org/doc/go1.25)
[![Go Report Card](https://goreportcard.com/badge/github.com/benwsapp/rlgl)](https://goreportcard.com/report/github.com/benwsapp/rlgl)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE.md)
[![Test Status](https://github.com/benwsapp/rlgl/actions/workflows/ci.yml/badge.svg)](https://github.com/benwsapp/rlgl/actions/workflows/ci.yml)
[![Coverage](https://codecov.io/gh/benwsapp/rlgl/graph/badge.svg?token=XDXHMQEK2L)](https://codecov.io/gh/benwsapp/rlgl)

## Features

- **Real-time Updates**: Server-Sent Events (SSE) stream config changes instantly
- **Simple Configuration**: Single YAML file to manage your work status
- **Focus Indicator**: Show what you're currently working on
- **Task Queue**: Display your upcoming tasks

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

### Local Binary

```bash
# Run with default config path
$ ./rlgl run

# Specify custom config and address
$ ./rlgl run --config /path/to/site.yaml --addr :3000

# With trusted origins for CSRF (comma-separated)
$ ./rlgl run --trusted-origins https://example.com,https://app.example.com

# Using environment variables
$ export RLGL_ADDR=":3000"
$ export RLGL_CONFIG="config/site.yaml"
$ export RLGL_TRUSTED_ORIGINS="https://example.com,https://app.example.com"
$ ./rlgl run
```

### Environment Variables

All flags can be set via environment variables:
- `RLGL_ADDR` - Server address (default: `:8080`)
- `RLGL_CONFIG` - Path to config file (default: `site.yaml` in current dir or `config/`)
- `RLGL_TRUSTED_ORIGINS` - Comma-separated list of trusted origins for CSRF protection

### Docker

#### Basic Run

```bash
$ docker run -p 8080:8080 \
    -v $(pwd)/config/site.yaml:/config/site.yaml:ro \
    rlgl:latest run --config /config/site.yaml
```

#### With Environment Variables

```bash
$ docker run -p 8080:8080 \
    -v $(pwd)/config:/config:ro \
    -e RLGL_CONFIG=/config/site.yaml \
    -e RLGL_TRUSTED_ORIGINS="https://example.com,https://app.example.com" \
    rlgl:latest run
```

#### Mount Entire Config Directory

```bash
$ docker run -p 8080:8080 \
    -v $(pwd)/config:/config:ro \
    rlgl:latest run --config /config/site.yaml
```

## Endpoints

- `GET /` - Main page (renders template with config)
- `GET /config` - JSON endpoint returning current config
- `GET /events` - Server-Sent Events stream for real-time config updates

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
