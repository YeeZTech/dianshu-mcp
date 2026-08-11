# dianshu-mcp

[English](./README.md) | [简体中文](./README_zh.md)

> GitHub: https://github.com/YeeZTech/dianshu-mcp | Gitee: https://gitee.com/YeeZTech/dianshu-mcp

典枢数据平台（dianshudata.com）的 MCP 服务——为 AI Agent 提供典枢平台的完整操作能力，包括登录、订单管理、数据下载、API 调用、数据集搜索等。

## 部署说明

### 方式一（推荐）：从 Releases 安装

1. 下载最新 Release 的压缩包（根据系统选择）：
   - macOS Apple Silicon（M1/M2/M3）：`macos-arm64.zip`
   - macOS Intel：`macos-amd64.zip`
   - Linux x86_64：`linux-amd64.zip`
   - Linux arm64：`linux-arm64.zip`
   - Windows x86_64：`windows-amd64.zip`
   - Windows arm64：`windows-arm64.zip`
2. 解压后会得到：
   - 可执行文件：`dianshu-mcp`（Windows 为 `dianshu-mcp.exe`）
   - skills 目录：`skills/`

### 方式二：源码编译安装

前提条件：**Go 1.22+**（所有平台）、**Git**

```bash
git clone https://github.com/YeeZTech/dianshu-mcp.git
cd dianshu-mcp
go build -o dianshu-mcp .
```

### 启动服务

```bash
./dianshu-mcp -headless=true
```

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-headless` | `false` | `true` = 后台模式（不弹浏览器）；`false` = 前台模式 |
| `-port` | `18061` | HTTP 服务端口 |
| `-output-dir` | `~/Downloads/dianshu-mcp/` | 自定义输出根目录 |

输出目录：
- 数据文件：`~/Downloads/dianshu-mcp/downloads/`
- API 结果：`~/Downloads/dianshu-mcp/api-data/`

### 导入 Skills & 配置 MCP

将 `.skill/dianshu/` 复制到对应 Agent 的 skills 目录，然后配置 MCP 连接。两步都必须完成。

#### Hermes

**导入 Skills：** `cp -r .skill/dianshu ~/.hermes/skills/`

**配置 MCP（`~/.hermes/config.yaml`）：**
```yaml
mcp_servers:
  dianshu-mcp:
    transport: streamable-http
    url: http://localhost:18061/mcp
```
或通过命令：`hermes config set mcp_servers.dianshu-mcp.transport streamable-http` 和 `hermes config set mcp_servers.dianshu-mcp.url http://localhost:18061/mcp`

#### Claude Code

**导入 Skills：** `cp -r .skill/dianshu .claude/skills/`

**配置 MCP（`.claude/settings.json`）：**
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
> ⚠️ Claude Code 使用 `"type"` 字段，非 `"transport"`。

#### Cursor

**导入 Skills：** `cp -r .skill/dianshu .cursor/skills/`

**配置 MCP（`.cursor/mcp.json`）：**
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

#### Trae / Trae Solo

**导入 Skills：** `cp -r .skill/dianshu .trae/skills/`

**配置 MCP：** 打开设置 → MCP → 手动添加，或编辑 `.trae/mcp.json`：
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

#### WorkBuddy

**导入 Skills：** `cp -r .skill/dianshu .workbuddy/skills/`

**配置 MCP：** 点击侧栏 CodeBuddy Settings → MCP → Add MCP，在 JSON 中添加，或编辑 `.workbuddy/mcp.json`：
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

#### Augment Code

**导入 Skills：** `cp -r .skill/dianshu .augment/skills/`

**配置 MCP（`.augment/mcp.json`）：**
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

#### Windsurf

**导入 Skills：** `cp -r .skill/dianshu .windsurf/skills/`

**配置 MCP（`.windsurf/mcp.json`）：**
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

#### 通义灵码

**导入 Skills：** `cp -r .skill/dianshu .tongyi/skills/`

**配置 MCP：** 右上角头像 → 个人设置 → MCP 服务 → + → 手工添加（SSE 类型，填名称和 `http://localhost:18061/mcp`），或编辑 `.tongyi/mcp.json`：
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

#### VS Code / Cline

**导入 Skills：** `cp -r .skill/dianshu .vscode/skills/`

**配置 MCP（项目根目录 `mcp.json` 或 Cline 扩展设置）：**
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

### 开始使用

Agent 连接成功后，说「登录典枢」即可扫码登录。之后可用自然语言操作：

- 「帮我查已购数据」→ 列出所有订单
- 「下载任务 xxx」→ 自动下载并解密
- 「搜索天气数据集」→ 搜索平台数据
- 「调用小红书 API」→ 调用已购数据 API

---

## 常见问题

### 安装后 Agent 没有加载到 MCP 工具？

Skills 和 MCP 是两个独立配置，漏了任何一个都会失败。逐一确认：

1. Skills 已复制到 Agent 的 skills 目录？
2. MCP 配置文件中已添加 `http://localhost:18061/mcp`？
3. `dianshu-mcp` 服务正在运行？
4. 重启 Agent

仍不行则手动调用 `check_login_status` 验证连接。

---

## MCP 工具清单（19 个）

### 账户与登录

| 工具 | 说明 |
|------|------|
| `check_login_status` | 检查典枢登录状态 |
| `get_login_qrcode` | 获取微信扫码登录二维码（渲染在聊天中） |
| `wait_login` | 等待扫码完成（配合 get_login_qrcode） |
| `open_login_browser` | 打开浏览器登录（支持扫码+账号密码） |
| `set_token` | 手动保存登录 token（浏览器登录备选） |
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
├── mcp.go                   # MCP 工具注册（19 个工具）
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
├── .skill/dianshu/  # Agent Skill（主 + 4 子 skill）
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
