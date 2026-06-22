# Project Guidelines

## 本地开发规范

- 每次修改完后，需要 `go fmt ./...` 格式化 Go 源码文件
- 测试过程中产生的 build 中间文件（如编译产物），如果没有必要则删除
- 所有的 feature 变更，都需要使用分支进行开发
- 在我未同意之前，不能推送到远程
- 变更流程：1.本地 review → 2.远程 PR review
- 不要过度设计，保持代码的简洁和易读
- 使用中文注释，一定要简洁明了，专业名词可以用英文

## 项目结构

```
dianshu-mcp/
├── main.go              # 入口，参数解析
├── app_server.go        # 应用服务容器
├── routes.go            # HTTP 路由 + MCP Streamable HTTP Handler
├── mcp_server.go        # MCP 工具注册（6个工具）
├── mcp_handlers.go      # MCP 工具处理函数
├── handlers_api.go      # REST API 处理函数
├── service.go           # 业务逻辑层
├── types.go             # 公共类型定义
├── cookies/cookies.go   # Cookie 持久化管理
├── dianshu/
│   ├── auth.go          # 微信扫码登录逻辑
│   ├── api.go           # 典枢平台 HTTP API 客户端
│   ├── browser.go       # go-rod 浏览器自动化辅助
│   └── types.go         # 典枢平台数据结构
└── configs/config.go    # 配置常量
```

## MCP 工具清单

| 工具 | 功能 | 参数 |
|------|------|------|
| `check_login_status` | 检查登录状态 | 无 |
| `get_login_qrcode` | 微信扫码登录（含等待扫码） | 无 |
| `delete_cookies` | 清除登录态 | 无 |
| `list_orders` | 查询订单/任务列表 | orderType(可选), orderCode(可选) |
| `list_downloads` | 列出可下载数据产品及下载信息 | 无 |
| `list_purchased_apis` | 列出已购买的 API 产品及调用信息 | 无 |

## 认证机制

典枢平台 API 使用 **请求头 `token`** 传递 JWT 认证（非 Cookie / Authorization Bearer），详见 `dianshu/api.go` 中的 `doRequest` 方法。

## Skill 关联

- `/dianshu` — 通用入口
- `/dianshu-login` — 微信扫码登录管理
- `/dianshu-order` — 订单查询
- `/dianshu-download` — 数据产品下载
- `/dianshu-api` — API 产品信息

## 技术栈

- 语言：Go
- Web 框架：Gin
- MCP SDK：`github.com/modelcontextprotocol/go-sdk`
- 浏览器自动化：`github.com/go-rod/rod`
- MCP 传输方式：Streamable HTTP（`:18061`）

## PR Review 重点

- 所有对典枢平台 API 的调用需确认认证方式是否正确（`token` 请求头）
- 浏览器自动化操作需注意页面结构变化，如果失败应有降级方案
- 新增 MCP 工具需按现有 `mcp_server.go` 中的模式注册
