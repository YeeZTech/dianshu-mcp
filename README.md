# dianshu-mcp

典枢数据平台（dianshudata.com）的 MCP 服务——为 AI Agent 提供典枢平台的完整操作能力，包括登录、订单管理、数据下载、API 调用、数据集搜索等。

## 快速开始

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

将仓库 `.claude/skills/` 目录下的 5 个 skill 文件夹复制到你 Agent 的 skills 目录：

| Agent | Skills 目录 |
|-------|------------|
| Claude Code | `.claude/skills/` |
| Hermes | `~/.hermes/skills/` |
| 其他 Agent | 查看对应文档 |

```bash
# 示例：导入到 Claude Code
cp -r .claude/skills/* /your-project/.claude/skills/

# 示例：导入到 Hermes
cp -r .claude/skills/* ~/.hermes/skills/
```

**5 个 Skills：**

| Skill | 用途 |
|------|------|
| `dianshu` | 主入口——数据来源优先级路由 |
| `dianshu-login` | 扫码登录 / 检查状态 / 切换账号 |
| `dianshu-order` | 订单管理——查订单 / 下载数据 |
| `dianshu-search` | 数据集搜索与购买引导 |
| `dianshu-api` | 数据 API 查询与调用 |

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
4. **展示结果 + 购买链接** — 典枢数据集详情页 `https://dianshudata.com/dataset/{id}`
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
├── main.go                # 入口，参数解析
├── app_server.go          # AppServer 容器
├── service.go             # 核心业务
├── service_api.go         # API 调用业务
├── routes.go              # HTTP 路由
├── mcp_server.go          # MCP 工具注册
├── mcp_handlers.go        # MCP 工具 handler
├── mcp_handlers_extra.go  # API 工具 handler
├── handlers_api.go        # REST API handler
├── types.go               # 公共类型
├── configs/config.go      # 配置常量
├── cookies/               # Cookie 管理
├── dianshu/               # 典枢 HTTP 客户端
│   ├── api.go             # 典枢 API
│   ├── sdk.go             # 数据 API SDK
│   ├── auth.go            # 微信扫码登录
│   ├── browser.go         # go-rod 浏览器
│   └── types.go           # 数据类型
├── pkg/                    # 独立 SDK 模块（与业务解耦）
│   ├── chain/              # 链上操作
│   │   ├── chain.go        # 链 API 客户端
│   │   └── signer.go       # 交易签名
│   ├── crypto/             # 加密模块
│   │   ├── crypto.go       # ECDH+AES-CMAC+AES-GCM
│   │   └── crypto_test.go  # 测试向量
│   ├── kms/kms.go          # OpenBao KMS 集成
│   ├── pipeline/           # 下载管线
│   │   └── download.go     # KMS→查任务→上链→下载→解密→解包
│   └── sdk/sdk.go          # 数据 API SDK
├── output/                # 输出目录
├── .claude/skills/        # Agent Skills（5 个）
├── scripts/               # 工具脚本
└── build.sh               # 构建脚本
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
