---
name: dianshu-order
description: |
  典枢订单管理：查询订单列表、查看已购数据、下载数据文件、调用已购 API。
  当用户想查看典枢订单、已买数据、下载数据、查看下载列表、下载文件时使用。
---

## 执行流程

### 1. 登录检查

调用 `check_login_status` 检查登录状态。

- 未登录 → 提示用户先扫码登录（`dianshu-login`）
- 已登录 → 进入步骤 2

### 2. 识别意图

| 用户说法 | 操作 |
|---------|------|
| 我的订单 / 买了什么 | 调用 `list_orders` |
| 能下载什么 / 我有什么数据 | 调用 `list_downloads` 展示可下载列表 |
| 下载 XX / 帮我下载 XX | 先用 `list_downloads` 确认，再 `download_order` |
| 我的 API / 能调什么 API | 调用 `list_purchased_apis` |

### 3. 下载数据

用户指定要下载某个数据时：

1. 从 `list_downloads` 或 `list_orders` 中确认任务编码（`taskCode`）
2. 调用 `download_order`，参数 `taskCode`（订单里的"任务编码"字段）
3. 下载完成后告知用户文件路径（如 `output/downloads/数据集名称.zip`）

**支持并行下载**：用户要求下载多个数据时，同时调用多个 `download_order`。

### 4. 调用已购 API

转由 `dianshu-api` skill 处理。

### 5. 展示结果

下载成功提示：
```
✅ 下载完成: output/downloads/{数据集名称}.{格式}
```

## 失败处理

| 场景 | 处理 |
|---|---|
| 未登录 | 引导用户先使用 `dianshu-login` 登录 |
| 无可下载数据 | 引导用户去典枢平台购买：https://dianshudata.com |
| 下载失败 | 告知用户错误信息，建议重试 |
| 超时 | 提示可能数据较大或网络较慢，耐心等待 |
