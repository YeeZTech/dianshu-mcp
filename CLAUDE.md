# Project Guidelines

## 本地开发规范

- 每次修改完后，需要 `go fmt ./...` 格式化 Go 源码文件
- 测试过程中产生的 build 中间文件（如编译产物），如果没有必要则删除
- 所有的 feature 变更，都需要使用分支进行开发
- 在我未同意之前，不能推送到远程
- 变更流程：1.本地 review → 2.远程 PR review
- 不要过度设计，保持代码的简洁和易读
- 使用中文注释，简洁明了，专业名词可以用英文

## 项目结构

```
dianshu-mcp/
├── main.go                # 入口，参数解析
├── app_server.go          # AppServer 容器
├── service.go             # 核心业务
├── service_api.go         # API 调用业务
├── routes.go              # HTTP 路由 + MCP
├── mcp_server.go          # MCP 工具注册（16 个工具）
├── mcp_handlers.go        # MCP 工具 handler（通用）
├── mcp_handlers_extra.go  # MCP 工具 handler（API 工具）
├── handlers_api.go        # REST API handler
├── types.go               # 公共类型
├── configs/config.go      # 配置常量
├── cookies/               # Cookie 管理
├── dianshu/               # 典枢 HTTP 客户端
│   ├── api.go             # 典枢 API 客户端
│   ├── sdk.go             # 数据 API SDK
│   ├── auth.go            # 微信扫码登录
│   ├── browser.go         # go-rod 浏览器
│   └── types.go           # 数据类型
├── pipeline/download.go   # 下载管线（KMS→查任务→上链→下载→解密→解包）
├── chain/                 # 链上操作（chain.go + signer.go）
├── crypto/                # 加解密（ECDH+AES-CMAC+AES-GCM）
├── kms/kms.go             # KMS 集成
├── output/                # 下载输出目录
└── .claude/skills/        # Agent Skills
```

## MCP 工具清单（16 个）

| 工具 | 功能 |
|------|------|
| `check_login_status` | 检查登录状态 |
| `get_login_qrcode` | 微信扫码登录（返回 PNG 二维码） |
| `delete_cookies` | 清除 cookies，切换账号 |
| `list_orders` | 查询订单列表 |
| `list_downloads` | 列出已购买的可下载数据 |
| `download_order` | 下载并解密数据文件 |
| `list_purchased_apis` | 列出已购买的 API |
| `get_api_detail` | 获取 API 参数列表 |
| `call_api` | 调用数据 API（自动加解密） |
| `search_datasets` | 搜索典枢数据集 |
| `dataset_detail` | 获取数据集详情 |
| `homepage_recommend` | 首页推荐数据集 |
| `my_datasets` | 我发布的数据集 |
| `get_my_profile` | 获取账号资料 |
| `get_my_wallet` | 获取钱包余额 |
| `list_wallet_transactions` | 钱包交易明细 |

## 技术栈

- 语言：Go 1.22+
- Web 框架：Gin
- MCP SDK：`github.com/modelcontextprotocol/go-sdk`
- 浏览器自动化：`github.com/go-rod/rod`
- MCP 传输：Streamable HTTP（`:18061`）
