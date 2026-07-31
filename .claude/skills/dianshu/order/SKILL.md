---
name: dianshu/order
description: |
  典枢订单与账户管理：订单、下载、已购、钱包余额、个人资料。当用户查订单、下载数据、查余额时使用。
---

## 前置

调用 `check_login_status`。未登录 → 引导先登录。

## 操作

| 用户说法 | 操作 |
|---------|------|
| 我的订单 | `list_orders` |
| 能下载什么 / 已购数据 | `list_downloads` |
| 下载 XX | `list_downloads` 确认 → `download_order` |
| 我的 API | `list_purchased_apis` |
| 钱包 / 余额 | `get_my_wallet` |
| 账单 / 交易记录 | `list_wallet_transactions` |
| 我的资料 / 个人信息 | `get_my_profile` |

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
