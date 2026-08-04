# dianshu-mcp

[English](./README.md) | [简体中文](./README_zh.md)

MCP server for the [Dianshu Data Platform](https://dianshudata.com) — giving AI Agents full access to dataset trading, including login, order management, encrypted downloads, API calls, and marketplace search.

## For AI Agents

Users typically provide this GitHub repository link to an AI agent, and the agent installs the MCP server and skills by following one of the flows below.

Two installation options:
- **Option A (Recommended)**: Download the latest Release zip for your OS and run it
- **Option B**: Build from source

### Option A (Recommended): Install from Releases

1. Download the latest Release asset (pick one):
   - macOS Apple Silicon (M1/M2/M3): `macos-arm64.zip`
   - macOS Intel: `macos-amd64.zip`
   - Linux x86_64: `linux-amd64.zip`
   - Linux arm64: `linux-arm64.zip`
   - Windows x86_64: `windows-amd64.zip`
   - Windows arm64: `windows-arm64.zip`
2. Unzip to get:
   - Binary: `dianshu-mcp` (Windows: `dianshu-mcp.exe`)
   - Skills: `skills/`
3. Install skills: copy `skills/dianshu/` to your agent's skills directory (see “Import Skills”)
4. Launch the MCP server (see “Launch”)
5. Configure your agent to connect to MCP (see “Configure MCP”)

### Option B: Build from source

#### Prerequisites
- **Go 1.22+** (all platforms)
- **Git**

```bash
git clone https://github.com/YeeZTech/dianshu-mcp.git
cd dianshu-mcp
go build -o dianshu-mcp .
```

### Import Skills

Copy `.claude/skills/dianshu/` to your AI agent's skills directory:

| Agent | macOS / Linux | Windows |
|-------|--------------|---------|
| Hermes | `~/.hermes/skills/` | `%USERPROFILE%\\.hermes\\skills\\` |
| Claude Code | `.claude/skills/` | `.claude\\skills\\` |
| Cursor | `.cursor/skills/` | `.cursor\\skills\\` |

| Platform | Command |
|----------|---------|
| macOS / Linux | `cp -r .claude/skills/dianshu ~/.hermes/skills/` |
| Windows (PowerShell) | `Copy-Item -Recurse .claude/skills/dianshu $env:USERPROFILE\\.hermes\\skills\\` |
| Windows (CMD) | `xcopy /E /I .claude\\skills\\dianshu %USERPROFILE%\\.hermes\\skills\\` |

Sub-skills are auto-loaded when loading `dianshu` — no separate import needed.

### Configure MCP

Service listens on `http://localhost:18061/mcp` via Streamable HTTP.

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

After connecting your agent, say "Log in to Dianshu" to scan the QR code with WeChat. Then use natural language:

- "Show my purchased data" → list all orders
- "Download task XXX" → auto-download and decrypt
- "Search weather datasets" → search the marketplace
- "Call the Xiaohongshu API" → invoke a purchased data API

---

## Configuration

| Flag | Default | Description |
|------|---------|-------------|
| `-headless` | `false` | `true` = background mode (no browser popup) |
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
4. **Show results + purchase link** → `https://dianshudata.com/dataDetail/{id}` (API products use `/dataAPIDetail/{id}`)
5. **No results** → suggest visiting https://dianshudata.com

---

## MCP Tools (16)

### Account & Login

| Tool | Description |
|------|-------------|
| `check_login_status` | Check Dianshu login status |
| `get_login_qrcode` | Get WeChat QR code for login (PNG) |
| `delete_cookies` | Clear session, switch account |

### Orders & Downloads

| Tool | Description |
|------|-------------|
| `list_orders` | List orders, filter by type / code |
| `list_downloads` | List purchased downloadable datasets |
| `download_order` | Download and decrypt with task code |
| `list_purchased_apis` | List purchased data APIs |
| `get_api_detail` | Get API parameter schema |
| `call_api` | Call a purchased API (auto encrypt/decrypt) |

### Dataset Search

| Tool | Description |
|------|-------------|
| `search_datasets` | Search Dianshu marketplace by keyword |
| `dataset_detail` | Get dataset details |
| `homepage_recommend` | Get homepage recommendations |
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
├── mcp.go                   # MCP tool registration (16 tools)
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
├── .claude/skills/dianshu/  # Agent skills (1 main + 4 sub)
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
