# dianshu-mcp

典枢数据平台（https://dianshudata.com）的 MCP 服务。当前提供微信扫码登录和数据查询能力，支持作为 MCP 工具被 Claude Code 等客户端调用。

## 功能

- **微信扫码登录** — 通过微信开放平台扫码登录典枢账号
- **登录状态管理** — 检查登录状态、清除登录态
- **数据查询** — AI 可直接补全查询参数并路由到具体数据源
- **扣费抽象预留** — 当前保留通用扣费接口，后续可接真实扣费逻辑

## 快速开始

### 前置要求

- Go 1.21+
- Chrome / Chromium 浏览器（用于微信扫码登录）

### 编译

如果本机 `go` 命令正常：

```bash
cd dianshu-mcp
go build -o dianshu-mcp .
```

如果你的 `goenv` shim 已损坏，可临时使用本机缓存的 Go toolchain：

```bash
export PATH="/Users/zhyyao/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.5.darwin-arm64/bin:$PATH"
cd dianshu-mcp
go mod tidy
go build -o dianshu-mcp .
```

### 启动服务

```bash
./dianshu-mcp
```

启动后 MCP 服务运行在 `http://localhost:18061/mcp`（Streamable HTTP）。

> 当前小红书搜索数据源的 `Authorization` 已写死在 `dianshu/provider_xiaohongshu_search.go` 中，仅适用于该数据源。

## MCP 工具

### `check_login_status`

检查当前登录状态。

### `get_login_qrcode`

获取微信登录二维码并等待扫码完成。

### `delete_cookies`

清除保存的登录态，下次操作需要重新登录。

### `data_search`

数据查询工具。AI 可以只传 `query`，也可以主动补全其他字段。

**参数：**

```json
{
  "query": "西瓜",
  "provider": "xiaohongshu",
  "dataset": "search",
  "siteDomain": "xiaohongshu.com",
  "keyword": "西瓜",
  "page": "1",
  "startTime": "1779381079",
  "endTime": "1779427884"
}
```

## REST API

### 登录状态

```bash
GET /api/v1/login/status
```

### 获取登录二维码

```bash
GET /api/v1/login/qrcode
```

### 删除登录态

```bash
DELETE /api/v1/login/cookies
```

### 数据查询

```bash
curl -X POST 'http://localhost:18061/api/v1/data/search' \
  -H 'Content-Type: application/json' \
  -d '{
    "query": "西瓜",
    "provider": "xiaohongshu",
    "dataset": "search",
    "siteDomain": "xiaohongshu.com",
    "keyword": "西瓜",
    "page": "1"
  }'
```

## 当前默认路由规则

当前默认规则：

- `provider` 留空时，默认 `xiaohongshu`
- `dataset` 留空时，默认 `search`
- `siteDomain` 留空时，默认 `xiaohongshu.com`
- `keyword` 留空时，回退为 `query`
- `page` 留空时，默认 `1`

## 代码组织建议

- `dianshu/data_query.go`：统一查询抽象、通用返回值、扣费接口
- `dianshu/provider_*.go`：每个数据源 / 数据集各写一个文件，存放不通用实现
- `service.go`：统一业务编排（登录校验、参数兜底、调用查询、调用扣费）
- `mcp_server.go` / `handlers_api.go`：暴露工具入口，不耦合具体数据源细节
