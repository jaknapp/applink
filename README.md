# applink

A CLI tool for authenticating with SaaS applications and connecting them to AI tools like Cursor via MCP.

## Features

- **OAuth Authentication**: Authenticate with Slack, Notion, Linear via OAuth
- **API Key Support**: Store API keys for services like Honeycomb
- **Secure Storage**: Credentials stored in your system keychain
- **MCP Configuration**: Automatically configure Cursor's MCP servers

## Installation

```bash
go install github.com/jaknapp/applink/cmd/applink@latest
```

Or build from source:

```bash
git clone https://github.com/jaknapp/applink.git
cd applink
go build -o applink ./cmd/applink
```

## Quick Start

1. **Configure OAuth credentials** (create `~/.config/applink/config.yaml`):

```yaml
services:
  slack:
    client_id: "your-slack-client-id"
    client_secret: "your-slack-client-secret"
  notion:
    client_id: "your-notion-client-id"
    client_secret: "your-notion-client-secret"
  linear:
    client_id: "your-linear-client-id"
    client_secret: "your-linear-client-secret"

settings:
  callback_port: 8888
```

2. **Authenticate with services**:

```bash
applink login slack
applink login notion
applink login linear
```

3. **Configure Cursor's MCP servers**:

```bash
applink mcp install
```

4. **Restart Cursor** to activate MCP servers.

## Usage

### Authentication

```bash
# OAuth-based services (opens browser)
applink login slack
applink login notion
applink login linear

# API key based services (prompts for key)
applink login honeycomb
```

### Status & Token Management

```bash
# View authentication status
applink status

# Print a token for scripts
applink token slack

# Remove credentials
applink logout slack
```

### MCP Configuration

```bash
# Auto-configure Cursor's MCP servers
applink mcp install

# Manage individually
applink mcp add slack
applink mcp remove notion
applink mcp list
```

### API Requests

```bash
# Send authenticated requests
applink request slack GET /api/conversations.list
applink request notion POST /v1/search --data '{"query": "meeting notes"}'
applink request linear POST /graphql --data '{"query": "{ viewer { id } }"}'
```

## Supported Services

| Service   | Auth Type | MCP Server |
|-----------|-----------|------------|
| Slack     | OAuth     | ✓          |
| Notion    | OAuth     | ✓          |
| Linear    | OAuth     | ✓          |
| Honeycomb | API Key   | ✗          |

## Security

- Tokens are stored in your system keychain (macOS Keychain, GNOME Keyring, Windows Credential Manager)
- Slack uses user tokens (not bot tokens) so you only see what you have access to
- OAuth flows use localhost callbacks - no public URLs required

## License

MIT
