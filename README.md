# dianshu-mcp

典枢数据平台（https://dianshudata.com）的 MCP 服务。提供微信扫码登录、订单查询等功能，支持作为 MCP 工具被 Claude Code 等客户端调用。

## 功能

- **微信扫码登录** — 通过微信开放平台扫码登录典枢账号
- **登录状态管理** — 检查登录状态、清除登录态
- **订单/任务查询** — 查看已购买的数据订单列表
- **数据下载** — 查看已购买数据产品的下载链接和校验信息
- **API 产品** — 查看已购买 API 产品的调用参数和代码示例

## 快速开始

### 前置要求

- Go 1.21+
- Chrome / Chromium 浏览器（用于微信扫码登录）

### 编译

```bash
cd dianshu-mcp
go build -o dianshu-mcp .
```

### 启动服务

```bash
# 默认无头模式（不显示浏览器窗口）
./dianshu-mcp

# 有头模式（显示浏览器窗口，方便扫码）
./dianshu-mcp -headless=false

# 指定端口
./dianshu-mcp -port :18061

# 后台运行（关掉终端也不影响）
nohup ./dianshu-mcp > /tmp/dianshu-mcp.log 2>&1 &
```

启动后 MCP 服务运行在 `http://localhost:18061/mcp`（Streamable HTTP）。

### 停止服务

```bash
pkill -f dianshu-mcp
```

### 查看日志

```bash
cat /tmp/dianshu-mcp.log        # 查看日志
tail -f /tmp/dianshu-mcp.log    # 实时查看日志
```

### 重新编译并重启

```bash
cd dianshu-mcp
pkill -f dianshu-mcp
GOPROXY=https://goproxy.cn,direct go build -o dianshu-mcp .
nohup ./dianshu-mcp -headless=false > /tmp/dianshu-mcp.log 2>&1 &
```

### 配置 MCP 连接

```bash
claude mcp add dianshu --transport http http://localhost:18061/mcp
```

或写入 `~/.claude/settings.json`：

```json
{
  "mcpServers": {
    "dianshu": {
      "url": "http://localhost:18061/mcp"
    }
  }
}
```

配置后需重启 Claude Code 会话。

## MCP 工具

### `check_login_status`

检查当前登录状态，返回是否已登录及用户名。

**参数：** 无

**示例：**
```
✅ 已登录
用户名: TANG
```

### `get_login_qrcode`

获取微信登录二维码并等待扫码完成。

**参数：** 无

**流程：**
1. 自动打开浏览器窗口显示微信二维码
2. 用户使用手机微信扫码
3. 扫码后自动提取 token 并保存
4. 返回登录成功信息

### `delete_cookies`

清除保存的登录态，下次操作需要重新登录。

**参数：** 无

**示例：**
```
Cookies 已成功删除，登录状态已重置。
```

### `list_orders`

查询已购买的数据订单/任务列表。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `orderType` | number | 否 | 订单类型：0-全部(默认) |
| `orderCode` | string | 否 | 按订单编号查询 |

**示例：**
```
📋 任务/订单列表（共 3 条）

─── 1 ───
  数据名称: 捉妖物语2
  任务编号: P17757964217107820
  价格: ¥2.10
  卖家: 游戏频道
  购买时间: 2026-04-10 12:47
  状态: 已完成
```

### `list_downloads`

列出已购买的可下载数据产品及其下载信息。

**参数：** 无

**示例：**
```
📥 可下载数据产品（共 2 条）

─── 1 ───
  数据名称: 捉妖物语2
  任务编号: P17757964217107820
  文件格式: zip
  下载地址: https://cdn.dianshudata.com/xxx.sealed
  校验地址: https://download.dianshudata.com/...checksum
  客户端下载: https://d.dianshudata.com
  购买时间: 2026-04-10 12:47

💡 文件为 .sealed 格式，需使用典枢客户端解密
```

### `list_purchased_apis`

列出已购买的 API 产品及调用详情。

**参数：** 无

**示例：**
```
🔌 已购买的 API 产品（共 1 条）

  API 名称: 金融风控场景信息调取API
  API 类型: sync
  请求方式: GET
  参数:
    - appCode (string, 必填): 用户标识
    - apiCode (string, 必填): 产品标识
  Java 调用示例:
    DSAPIClient client = new DSAPIClient(...)

💡 可通过 DSAPIClient SDK 或直接 HTTP 调用 API
```

## Skill 命令

在 Claude Code 中可使用以下 skill：

| 命令 | 用途 |
|------|------|
| `/dianshu` | 典枢平台通用入口 |
| `/dianshu-login` | 微信扫码登录管理 |
| `/dianshu-order` | 查询订单信息 |
| `/dianshu-download` | 查看/下载已购买的数据产品 |
| `/dianshu-api` | 查看已购买的 API 产品信息 |

## 项目结构

```
dianshu-mcp/
├── main.go              # 入口
├── app_server.go        # 应用服务
├── routes.go            # HTTP 路由 & MCP Handler
├── mcp_server.go        # MCP 工具注册
├── mcp_handlers.go      # MCP 工具处理
├── handlers_api.go      # REST API
├── service.go           # 业务逻辑
├── types.go             # 类型定义
├── cookies/             # Cookie 持久化管理
├── dianshu/             # 典枢平台交互层
│   ├── auth.go          # 微信扫码登录
│   ├── api.go           # HTTP API 客户端
│   ├── browser.go       # 浏览器自动化
│   └── types.go         # 数据结构
└── configs/             # 配置
```

## 认证说明

典枢平台的 API 认证方式为 **请求头 `token`**（非 Cookie 或 Authorization Bearer 头）。

登录流程：
1. 调用 `get_login_qrcode` 打开微信开放平台二维码
2. 用户使用手机微信扫码
3. 微信认证后重定向到典枢 SSO
4. SSO 回调后设置 JWT token
5. 服务自动提取 token 并保存

所有后续 API 调用在请求头中添加 `token: <JWT>` 进行认证。

## 技术栈

- **语言：** Go
- **Web 框架：** Gin
- **MCP SDK：** `github.com/modelcontextprotocol/go-sdk` (v1.6.1+)
- **浏览器自动化：** `github.com/go-rod/rod` (v0.116+)
- **MCP 传输：** Streamable HTTP (JSON mode)

## 开发

```bash
# 编译
go build -o dianshu-mcp .

# 格式化
go fmt ./...

# 依赖管理
GOPROXY=https://goproxy.cn,direct go mod tidy
```
