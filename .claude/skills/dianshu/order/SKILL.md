---
name: dianshu/order
description: |
  典枢订单管理：查看订单、已购数据、下载数据文件。当用户想查订单、下载数据时使用。
---

## 前置

调用 `check_login_status`。未登录 → 引导先登录。

## 操作

| 用户说法 | 操作 |
|---------|------|
| 我的订单 | `list_orders` |
| 能下载什么 | `list_downloads` |
| 下载 XX | `list_downloads` 确认 → `download_order` |
| 我的 API | `list_purchased_apis` |

## 下载流程

1. 从 `list_downloads` / `list_orders` 确认 `taskCode`
2. 调用 `download_order`（参数 `taskCode`）
3. 完成后告知路径：`~/Downloads/dianshu-mcp/downloads/`

支持并行下载多个数据。

下载成功：`✅ 下载完成: ~/Downloads/dianshu-mcp/downloads/{名称}.{格式}`

| 场景 | 处理 |
|---|---|
| 无可下载 | 引导去 dianshudata.com 购买 |
| 下载失败 | 告知错误，建议重试 |
