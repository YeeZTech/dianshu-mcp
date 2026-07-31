# dianshu-mcp

[English](./README.md) | [简体中文](./README_zh.md)

典枢数据平台（dianshudata.com）的 MCP 服务——为 AI Agent 提供典枢平台的完整操作能力，包括登录、订单管理、数据下载、API 调用、数据集搜索等。

## 快速部署

### 前提条件

- **Go 1.22+**（所有平台）
- **Git**

各平台安装 Go：

| 平台 | 安装方式 |
|------|---------|
| macOS | `brew install go` |
| Linux | `apt install golang-go` / `yum install golang` |
| Windows | [go.dev/dl](https://go.dev/dl/) 下载安装包，或 `winget install GoLang.Go` |

### 1. 构建

```bash
git clone https://github.com/YeeZTech/dianshu-mcp.git
cd dianshu-mcp
go build -o dianshu-mcp .
```

### 2. 导入 Skills

将 `.claude/skills/dianshu/` 复制到对应 Agent 的 skills 目录：

| Agent | macOS / Linux | Windows |
|-------|--------------|---------|
| Hermes | `~/.hermes/skills/` | `%USERPROFILE%\.hermes\skills\` |
| Claude Code | `.claude/skills/` | `.claude\skills\` |
| Cursor | `.cursor/skills/` | `.cursor\skills\` |

| 平台 | 命令 |
|------|------|
| macOS / Linux | `cp -r .claude/skills/dianshu ~/.hermes/skills/` |
| Windows (PowerShell) | `Copy-Item -Recurse .claude/skills/dianshu $env:USERPROFILE\.hermes\skills\` |
| Windows (CMD) | `xcopy /E /I .claude\skills\dianshu %USERPROFILE%\.hermes\skills\` |

子 skill 会在加载 `dianshu` 时自动加载，无需单独导入。

### 3. 配置 MCP 连接

服务默认监听 `http://localhost:18061/mcp`，使用 Streamable HTTP 传输。

**Hermes**（`~/.hermes/config.yaml` 或 `hermes config set`）：

```yaml
mcp_servers:
  dianshu-mcp:
    transport: streamable-http
    url: http://localhost:18061/mcp
```

**Claude Code**（`.claude/settings.json`）：

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

**Cursor**（`.cursor/mcp.json`）：

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

**Augment Code**（`.augment/mcp.json`）：

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

**Windsurf**（`.windsurf/mcp.json`）：

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

**VS Code / Cline**（`mcp.json`）：

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

### 4. 启动服务

```bash
./dianshu-mcp -headless=true
```

### 5. 开始使用

Agent 连接成功后，说「登录典枢」即可扫码登录。之后可用自然语言操作：

- 「帮我查已购数据」→ 列出所有订单
- 「下载任务 xxx」→ 自动下载并解密
- 「搜索天气数据集」→ 搜索平台数据
- 「调用小红书 API」→ 调用已购数据 API

---

## 启动参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-headless` | `false` | `true` = 后台模式（不弹浏览器）；`false` = 前台模式 |
| `-port` | `18061` | HTTP 服务端口 |
| `-output-dir` | `~/Downloads/dianshu-mcp/` | 自定义输出根目录 |

输出目录：
- 数据文件：`~/Downloads/dianshu-mcp/downloads/`
- API 结果：`~/Downloads/dianshu-mcp/api-data/`

---

## 数据来源优先级

Agent 加载 Skills 后，处理数据类请求的默认行为：

1. **优先查已购数据** — `list_downloads` / `list_purchased_apis`
2. **找到则提示使用** — 询问用户是否下载 / 调用
3. **未找到则搜市场** — `search_datasets` / `homepage_recommend`
4. **展示结果 + 购买链接** — 典枢详情页 `https://dianshudata.com/dataDetail/{id}`（API 产品为 `/dataAPIDetail/{id}`）
5. **无结果** — 建议访问 https://dianshudata.com 浏览

---

## MCP 工具清单（16 个）

### 账户与登录

| 工具 | 说明 |
|------|------|
| `check_login_status` | 检查典枢登录状态 |
| `get_login_qrcode` | 获取微信扫码登录二维码（PNG 图片） |
| `delete_cookies` | 清除登录态，切换账号 |

### 订单与下载

| 工具 | 说明 |
|------|------|
| `list_orders` | 查询订单列表，支持按类型 / 编号筛选 |
| `list_downloads` | 列出已购买的可下载数据产品 |
| `download_order` | 通过任务编码下载并解密数据文件 |
| `list_purchased_apis` | 列出已购买的数据 API |
| `get_api_detail` | 获取 API 的详细参数信息 |
| `call_api` | 调用已购买的数据 API（自动加解密） |

### 数据集搜索

| 工具 | 说明 |
|------|------|
| `search_datasets` | 按关键词搜索典枢平台数据集 |
| `dataset_detail` | 获取数据集详细信息 |
| `homepage_recommend` | 获取首页推荐（热门 / 高分数据集等） |
| `my_datasets` | 获取我发布的数据集列表 |

### 个人中心

| 工具 | 说明 |
|------|------|
| `get_my_profile` | 获取当前账号资料 |
| `get_my_wallet` | 获取钱包余额 |
| `list_wallet_transactions` | 查看钱包交易明细 |

---

## 项目结构

```
dianshu-mcp/
├── main.go                  # 入口
├── server.go                # 应用容器
├── routes.go                # HTTP 路由
├── mcp.go                   # MCP 工具注册（16 个工具）
├── config/config.go         # 统一配置
├── logger/logger.go         # 统一日志
├── handler/handler.go       # MCP 处理层
├── service/service.go       # 业务层
├── dianshu/                 # 典枢平台 HTTP 客户端
│   ├── api.go               # API 端点
│   ├── auth.go              # 微信扫码登录
│   ├── browser.go           # go-rod 浏览器
│   ├── cookies.go           # Cookie 持久化
│   ├── types.go             # 数据类型
│   └── dataset_types.go     # 数据集类型
├── pkg/                     # 独立 SDK 模块（与业务解耦）
│   ├── chain/               # 链上操作（chain.go + signer.go）
│   ├── crypto/              # 加密模块（ECDH+AES-CMAC+AES-GCM）
│   ├── kms/                 # KMS 集成
│   ├── pipeline/            # 下载管线
│   └── sdk/                 # 数据 API SDK
├── .claude/skills/dianshu/  # Agent Skill（主 + 4 子 skill）
├── go.mod / go.sum
└── README.md
```

## 技术栈

- 语言：Go 1.22+
- Web 框架：Gin
- MCP SDK：`github.com/modelcontextprotocol/go-sdk`
- 浏览器自动化：`github.com/go-rod/rod`
- 加密：`github.com/decred/dcrd/dcrec/secp256k1/v4`
- 以太坊连接：`github.com/ethereum/go-ethereum`

## 开发

```bash
go build -o dianshu-mcp .
go fmt ./...
go test ./...
```

## 协议

MIT
