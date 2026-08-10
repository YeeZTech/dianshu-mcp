# dianshu-mcp

[English](./README.md) | [简体中文](./README_zh.md)

MCP server for the [Dianshu Data Platform](https://dianshudata.com) — giving AI Agents full access to dataset trading, including login, order management, encrypted downloads, API calls, and marketplace search.

## Deployment

Two options:
- **Option A (Recommended)**: Download the latest Release zip for your OS and run it directly
- **Option B**: Clone the source and build it yourself

### Option A (Recommended): Install from Releases

1. Download the latest Release archive (choose your system):
   - macOS Apple Silicon (M1/M2/M3): `macos-arm64.zip`
   - macOS Intel: `macos-amd64.zip`
   - Linux x86_64: `linux-amd64.zip`
   - Linux arm64: `linux-arm64.zip`
   - Windows x86_64: `windows-amd64.zip`
   - Windows arm64: `windows-arm64.zip`
2. After extraction you'll get:
   - Binary: `dianshu-mcp` (Windows: `dianshu-mcp.exe`)
   - Skills directory: `skills/`
3. Install skills: copy `skills/dianshu/` to your agent's skills directory (see "Import Skills" below)
4. Launch the MCP server (see "Launch" below)
5. Configure your agent to connect to MCP (see "Configure MCP" below)

### Option B: Build from Source

#### Prerequisites
- **Go 1.22+** (all platforms)
- **Git**

```bash
git clone https://github.com/YeeZTech/dianshu-mcp.git
cd dianshu-mcp
go build -o dianshu-mcp .
```

### Import Skills

Copy `.skill/dianshu/` to your agent's skills directory:

| Agent | macOS / Linux | Windows |
|-------|--------------|---------|
| Hermes | `~/.hermes/skills/` | `%USERPROFILE%\\.hermes\\skills\\` |
| Claude Code | `.claude/skills/` | `.claude\\skills\\` |
| Cursor | `.cursor/skills/` | `.cursor\\skills\\` |

| Platform | Command |
|----------|---------|
| macOS / Linux | `cp -r .skill/dianshu ~/.hermes/skills/` |
| Windows (PowerShell) | `Copy-Item -Recurse .skill/dianshu $env:USERPROFILE\\.hermes\\skills\\` |
| Windows (CMD) | `xcopy /E /I .skill\\dianshu %USERPROFILE%\\.hermes\\skills\\` |

Sub-skills are auto-loaded when loading `dianshu` — no separate import needed.

### Configure MCP

Service listens on `http://localhost:18061/mcp` via Streamable HTTP.

> **Universal config for all agents**: Add an MCP server, choose "Streamable HTTP" transport, and enter `http://localhost:18061/mcp` as the URL.

Agent-specific examples:

**Hermes** (`~/.hermes/config.yaml` or `hermes config set`):

```yaml
mcp_servers:
  dianshu-mcp:
    transport: streamable-http
    url: http://localhost:18061/mcp
```

**Claude Code** (`.claude/settings.json`):

```json
{
  "mcpServers": {
    "dianshu-mcp": {
      "type": "streamable-http",
      "url": "http://localhost:18061/mcp"
    }
  }
}
```

**Cursor** (`.cursor/mcp.json`):

```json
{
  "mcpServers": {
    "dianshu-mcp": {
      "transport": "streamable-http",
      "url": "http://localhost:18061/mcp"
    }
  }
}
```

**Augment Code** (`.augment/mcp.json`):

```json
{
  "mcpServers": {
    "dianshu-mcp": {
      "transport": "streamable-http",
      "url": "http://localhost:18061/mcp"
    }
  }
}
```

**Windsurf** (`.windsurf/mcp.json`):

```json
{
  "mcpServers": {
    "dianshu-mcp": {
      "transport": "streamable-http",
      "url": "http://localhost:18061/mcp"
    }
  }
}
```

**VS Code / Cline** (`mcp.json`):

```json
{
  "mcpServers": {
    "dianshu-mcp": {
      "transport": "streamable-http",
      "url": "http://localhost:18061/mcp"
    }
  }
}
```

### Launch

```bash
./dianshu-mcp -headless=true
```

### Get Started

After connecting your agent, say "Log in to Dianshu" to scan the QR code. Then use natural language:

- "Show my purchased data" → list all orders
- "Download task XXX" → auto-download and decrypt
- "Search weather datasets" → search the marketplace
- "Call the Xiaohongshu API" → invoke a purchased data API

---

## Troubleshooting

### MCP tools not appearing after installation?

Skills and MCP are two independent configurations — missing either will cause failure. Verify each step:

1. **Skills copied?** — Confirm `.skill/dianshu/` was copied to your agent's skills directory
2. **MCP configured?** — Confirm the MCP config has `http://localhost:18061/mcp` added
3. **Server running?** — Confirm `dianshu-mcp` is running
4. **Restart agent** — Some agents require a restart to load new MCP connections

If still not working, manually call the `check_login_status` tool to verify connectivity.

### Don't know which agent you're using?

All major AI coding agents support MCP. Find an "MCP Server" or "Integrations" settings page, add a Streamable HTTP server at `http://localhost:18061/mcp`, and you're done.

---

## Configuration

| Flag | Default | Description |
|------|---------|-------------|
| `-headless` | `false` | `true` = background mode (no browser popup); `false` = foreground mode |
| `-port` | `18061` | HTTP listen port |
| `-output-dir` | `~/Downloads/dianshu-mcp/` | Override output directory |

Output paths:
- Downloaded files → `~/Downloads/dianshu-mcp/downloads/`
- API results → `~/Downloads/dianshu-mcp/api-data/`

---

## Data Priority

When skills are loaded, the agent follows this priority for data requests:

1. **Check purchased first** — `list_downloads` / `list_purchased_apis`
2. **Found** → ask user whether to download / call
3. **Not found** → search marketplace → `search_datasets` / `homepage_recommend`
4. **Show results + purchase link** → `https://dianshudata.com/dataDetail/{id}` (API products: `/dataAPIDetail/{id}`)
5. **No results** → suggest visiting https://dianshudata.com

---

## MCP Tools (19)

### Account & Login

| Tool | Description |
|------|-------------|
| `check_login_status` | Check Dianshu login status |
| `get_login_qrcode` | Get WeChat QR login image (displays in chat) |
| `wait_login` | Wait for QR scan completion (use after get_login_qrcode) |
| `open_login_browser` | Open browser for login (QR + password) |
| `set_token` | Manually save login token (browser login fallback) |
| `delete_cookies` | Clear session, switch account |

### Orders & Downloads

| Tool | Description |
|------|-------------|
| `list_orders` | List orders, filter by type / code |
| `list_downloads` | List purchased downloadable data products |
| `download_order` | Download and decrypt with task code |
| `list_purchased_apis` | List purchased data APIs |
| `get_api_detail` | Get API parameter details |
| `call_api` | Call a purchased API (auto encrypt/decrypt) |

### Dataset Search

| Tool | Description |
|------|-------------|
| `search_datasets` | Search Dianshu marketplace by keyword |
| `dataset_detail` | Get dataset details |
| `homepage_recommend` | Get homepage recommendations (popular / high-rated) |
| `my_datasets` | Get my published datasets |

### Profile & Wallet

| Tool | Description |
|------|-------------|
| `get_my_profile` | Get account profile |
| `get_my_wallet` | Get wallet balance |
| `list_wallet_transactions` | View transaction history |

---

## Project Structure

```
dianshu-mcp/
├── main.go                  # Entry point
├── server.go                # Application container
├── routes.go                # HTTP routes
├── mcp.go                   # MCP tool registration (19 tools)
├── config/config.go         # Unified configuration
├── logger/logger.go         # Unified logging
├── handler/handler.go       # MCP handler layer
├── service/service.go       # Business layer
├── dianshu/                 # Dianshu HTTP client
│   ├── api.go               # API endpoints
│   ├── auth.go              # WeChat QR login
│   ├── browser.go           # go-rod browser automation
│   ├── cookies.go           # Cookie persistence
│   ├── types.go             # Data types
│   └── dataset_types.go     # Dataset types
├── pkg/                     # Standalone SDK modules
│   ├── chain/               # On-chain ops (chain.go + signer.go)
│   ├── crypto/              # Crypto (ECDH + AES-CMAC + AES-GCM)
│   ├── kms/                 # KMS integration
│   ├── pipeline/            # Download pipeline
│   └── sdk/                 # Data API SDK
├── .skill/dianshu/  # Agent skills (1 main + 4 sub)
├── go.mod / go.sum
└── README.md
```

## Tech Stack

- Language: Go 1.22+
- Web framework: Gin
- MCP SDK: `github.com/modelcontextprotocol/go-sdk`
- Browser automation: `github.com/go-rod/rod`
- Cryptography: `github.com/decred/dcrd/dcrec/secp256k1/v4`
- Ethereum: `github.com/ethereum/go-ethereum`

## Development

```bash
go build -o dianshu-mcp .
go fmt ./...
go test ./...
```

## License

MIT
