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
├── app_server.go        # AppServer 容器
├── service.go           # 核心业务（登录/查询/持久化）
├── routes.go            # HTTP 路由 + MCP
├── mcp_server.go        # MCP 工具注册
├── mcp_handlers.go      # MCP 工具 handler
├── handlers_api.go      # REST API handler
├── types.go             # 公共类型
├── configs/config.go    # 配置常量
├── cookies/cookies.go   # Cookie 管理
├── dianshu/
│   ├── api.go           # 典枢 HTTP 客户端（登录）
│   ├── auth.go          # 微信扫码登录
│   ├── browser.go       # go-rod 浏览器辅助
│   ├── types.go         # 典枢数据结构（UserInfo）
│   ├── data_query.go    # 数据查询抽象（router/provider/charge）
│   └── provider_xiaohongshu_search.go  # 小红书搜索 provider
├── service_test.go      # 核心业务测试
└── dianshu/consult_api_test.go  # provider 测试
```

## MCP 工具清单

| 工具 | 功能 | 参数 |
|------|------|------|
| `check_login_status` | 检查登录状态 | 无 |
| `get_login_qrcode` | 微信扫码登录 | 无 |
| `delete_cookies` | 清除登录态 | 无 |
| `data_search` | 数据查询，结果写入 output/data-search/ | query(必), provider, dataset, siteDomain, keyword, page, startTime, endTime |

## 技术栈

- 语言：Go
- Web 框架：Gin
- MCP SDK：`github.com/modelcontextprotocol/go-sdk`
- 浏览器自动化：`github.com/go-rod/rod`
- MCP 传输方式：Streamable HTTP（`:18061`）
