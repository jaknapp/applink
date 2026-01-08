# applink

A CLI tool for authenticating with SaaS applications and connecting them to AI tools like Cursor via MCP.

## Features

- **OAuth Authentication**: Authenticate with Slack, Notion, Linear via OAuth
- **API Key Support**: Store API keys for services like Honeycomb
- **Secure Storage**: Credentials stored in your system keychain (macOS Keychain, GNOME Keyring, Windows Credential Manager)
- **MCP Configuration**: Automatically configure Cursor's MCP servers

## Installation

**Homebrew:**
```bash
brew install jaknapp/tap/applink
```

**Install script:**
```bash
curl -sSL https://raw.githubusercontent.com/jaknapp/applink/main/install.sh | sh
```

**Go install:**
```bash
go install github.com/jaknapp/applink/cmd/applink@latest
```

**From source:**
```bash
git clone https://github.com/jaknapp/applink.git
cd applink
go build -o applink ./cmd/applink
```

## Quick Start

### 1. Set up OAuth credentials

```bash
applink setup slack
```

This will:
- Show you instructions for creating a Slack OAuth app
- Prompt for your Client ID and Client Secret
- Store them securely in your system keychain

### 2. Authenticate

```bash
applink login slack
```

This opens your browser to complete the OAuth flow.

### 3. Configure Cursor

```bash
applink mcp install
```

This configures Cursor's MCP servers for your authenticated services.

### 4. Restart Cursor

Restart Cursor to activate the MCP servers.

## Usage

### Setup & Authentication

```bash
# Set up OAuth credentials (one-time per service)
applink setup slack
applink setup notion
applink setup linear

# Authenticate (opens browser)
applink login slack
applink login notion
applink login linear

# API key services (prompts for key)
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

## Environment Variables

For CI/CD or systems without a keychain, use environment variables:

```bash
# OAuth credentials
export APPLINK_SLACK_CLIENT_ID="your-client-id"
export APPLINK_SLACK_CLIENT_SECRET="your-client-secret"

# Then authenticate
applink login slack
```

Environment variables take priority over keychain credentials.

## Supported Services

| Service   | Auth Type | MCP Server |
|-----------|-----------|------------|
| Slack     | OAuth     | ✓          |
| Notion    | OAuth     | ✓          |
| Linear    | OAuth     | ✓          |
| Honeycomb | API Key   | ✗          |

## Security

- **Keychain storage**: OAuth credentials and tokens are stored in your system keychain, not in plaintext files
- **User tokens**: Slack uses user tokens (not bot tokens) so you only see what you have access to
- **Local OAuth**: OAuth callbacks use localhost - no public URLs required

## Troubleshooting

### Keychain not available (Linux)

On headless Linux systems or containers without a keychain daemon, use environment variables:

```bash
export APPLINK_SLACK_CLIENT_ID="..."
export APPLINK_SLACK_CLIENT_SECRET="..."
applink login slack
```

### OAuth callback fails

Ensure port 8888 is available and not blocked by a firewall.

## Releasing

See [RELEASING.md](RELEASING.md) for how to publish new versions.

## License

MIT
