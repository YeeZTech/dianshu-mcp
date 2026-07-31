# dianshu-mcp

典枢数据平台（dianshudata.com）的 MCP 服务——为 AI Agent 提供典枢平台的完整操作能力，包括登录、订单管理、数据下载、API 调用、数据集搜索等。

## 快速部署

### 1. 构建

```bash
git clone https://github.com/user/dianshu-mcp.git
cd dianshu-mcp
go build -o dianshu-mcp .
```

### 2. 导入 Skills

将 `skills/dianshu/` 复制到对应 Agent 的 skills 目录：

| Agent | Skills 路径 |
|-------|------------|
| Hermes | `~/.hermes/skills/dianshu/` |
| Claude Code | `.claude/skills/dianshu/` |
| Cursor | `.cursor/skills/dianshu/` |
| Augment Code | `.augment/skills/dianshu/` |
| Windsurf | `.windsurf/skills/dianshu/` |
| CodeBuddy | `.codebuddy/skills/dianshu/` |
| 通用 (Claude 系) | `.claude/skills/dianshu/` |

```bash
# Hermes
cp -r .claude/skills/dianshu ~/.hermes/skills/

# Claude Code / Cursor / 其他 Claude 系
cp -r .claude/skills/dianshu .claude/skills/
```

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

**Claude Code**（`.claude/settings.json` 或 `cla` 全局配置）：

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

## 手动安装

### 前提条件

- 典枢平台账号（https://dianshudata.com 注册）
- 支持的平台：Windows / macOS / Linux

### 安装方式一：下载预编译二进制（推荐）

从 [GitHub Releases](https://github.com/your-username/dianshu-mcp/releases) 下载对应平台的二进制文件：

| 平台 | 文件 |
|------|------|
| Windows (x64) | `dianshu-mcp-windows-amd64.exe` |
| macOS (Apple Silicon) | `dianshu-mcp-darwin-arm64` |
| macOS (Intel) | `dianshu-mcp-darwin-amd64` |
| Linux (x64) | `dianshu-mcp-linux-amd64` |

下载后放到任意目录，添加执行权限（macOS/Linux），直接运行即可。

#### Windows

```powershell
# 下载 dianshu-mcp-windows-amd64.exe 到任意目录
# 双击运行，或命令行启动：
.\dianshu-mcp-windows-amd64.exe -headless=true
```

#### macOS / Linux

```bash
# 下载后添加执行权限
chmod +x dianshu-mcp-*
# 启动
./dianshu-mcp-darwin-arm64 -headless=true
```

### 安装方式二：从源码构建

要求 Go 1.22+。

```bash
# Windows / macOS / Linux 通用
git clone https://github.com/your-username/dianshu-mcp.git
cd dianshu-mcp

# 下载依赖
go mod download

# 构建
go build -o dianshu-mcp .

# Windows
dianshu-mcp.exe -headless=true

# macOS / Linux
./dianshu-mcp -headless=true
```

### 启动参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-headless` | `false` | `true` = 后台模式（不弹浏览器）；`false` = 前台模式 |
| `-port` | `18061` | HTTP 服务端口 |

服务默认监听 `http://localhost:18061`，提供 Streamable HTTP MCP 协议。

下载的数据文件和 API 调用结果保存到系统下载目录：
- 数据文件：`~/Downloads/dianshu-mcp/downloads/`
- API 结果：`~/Downloads/dianshu-mcp/api-data/`

可通过 `-output-dir` 参数自定义根目录。

---

## 配置 AI Agent

### 第一步：配置 MCP 服务连接

在 AI Agent 的 MCP 配置中添加（不同 Agent 配置文件位置不同，以下为常见示例）：

**通用 MCP 配置**（`mcp.json` / `mcp_servers.json` 等）：

```json
{
  "mcpServers": {
    "dianshu-mcp": {
      "url": "http://localhost:18061/mcp",
      "transport": "streamable-http"
    }
  }
}
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

**Hermes**（`~/.hermes/config.yaml`）：

```yaml
mcp_servers:
  dianshu-mcp:
    transport: streamable-http
    url: http://localhost:18061/mcp
```

### 第二步：导入 Skills

复制 `.claude/skills/dianshu/` 一个文件夹即可：

```
.claude/skills/dianshu/
├── SKILL.md           ← 主入口——意图路由 + 数据优先级
├── login/SKILL.md     ← 登录管理
├── order/SKILL.md     ← 订单与下载
├── search/SKILL.md    ← 数据市场搜索
└── api/SKILL.md       ← API 调用
```

| Agent | Skills 目录 |
|-------|------------|
| Claude Code | `.claude/skills/` |
| Hermes | `~/.hermes/skills/` |

```bash
# 复制到 Claude Code
cp -r .claude/skills/dianshu /your-project/.claude/skills/

# 复制到 Hermes
cp -r .claude/skills/dianshu ~/.hermes/skills/
```

导入后加载 `dianshu` 即可，子 skill 自动加载。

---

## 首次使用

1. 启动服务后，在 Agent 中说「登录典枢」
2. Agent 会展示微信二维码，用微信扫码完成登录
3. 登录成功后即可开始使用

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
├── build.sh                 # 构建脚本
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
