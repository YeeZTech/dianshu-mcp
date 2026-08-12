---
name: dianshu
description: |
  典枢数据平台 MCP 服务 —— AI Agent 对接数据交易市场。当用户提到「数据」「数据集」「下载」「API」「典枢」，
  或需要查已购/搜市场/调用 API 时使用。覆盖登录、订单、下载解密、数据 API、市场搜索、钱包等全部能力。
version: "v1.0.3"
homepage: "https://dianshudata.com"
metadata:
  hermes:
    related_skills: [dianshu/login, dianshu/order, dianshu/search, dianshu/api]
---

# 典枢 MCP 服务使用指南

> 当前 v1.0.3 | 最新：[GitHub Releases](https://github.com/YeeZTech/dianshu-mcp/releases) | [Gitee Releases](https://gitee.com/YeeZTech/dianshu-mcp/releases)

---

## §0 会话启动 / 定时任务

### 0.1 检查 MCP 连接

检查工具列表中是否存在 `check_login_status`：
- **有** → 继续 0.2
- **没有** → 回复：「典枢 MCP 服务未连接，请参考 https://github.com/YeeZTech/dianshu-mcp 部署。」

### 0.2 检查登录状态

调用 `check_login_status`。
- 已登录 → 告知当前账号，继续
- 未登录 → 进入登录流程（见 `dianshu/login`）

### 0.3 检查版本更新

**每次会话启动时执行一次，之后每 1 小时检查一次**：

1. 访问 https://github.com/YeeZTech/dianshu-mcp/releases/latest 获取最新版本号
   （如 GitHub 不可达则尝试 https://gitee.com/YeeZTech/dianshu-mcp/releases）
2. 与当前版本（v1.0.3）对比：
   - 相同 → 跳过
   - 不同 → 回复：
     ```
     ⚠️ 典枢 MCP 有新版本 {最新版本}，当前 v1.0.3
     更新方法：重新下载最新 Release zip，解压替换二进制和 skills 目录
     GitHub: https://github.com/YeeZTech/dianshu-mcp/releases
     Gitee: https://gitee.com/YeeZTech/dianshu-mcp/releases
     ```
3. 每次提醒间隔至少 1 小时，不要频繁打扰用户。

### 0.4 自动更新

当发现新版本时，询问用户是否更新。用户同意后，按照 [README](https://github.com/YeeZTech/dianshu-mcp)（或 [Gitee](https://gitee.com/YeeZTech/dianshu-mcp)）的「部署说明」重新执行一次完整安装流程：

1. 停掉当前 MCP 服务（kill 进程）
2. 重新克隆最新代码并编译（或下载最新 Release zip）
3. 重新导入 Skills
4. 重新配置 MCP
5. 重新启动并验证

> 等同于全新安装，确保所有组件（二进制、Skills、配置）更新到最新版本。

---

## §1 核心操作

### 查订单

| 用户说法 | 工具 | 说明 |
|---------|------|------|
| 我的订单 | `list_orders` | 默认列出全部 |
| 待支付的订单 | `list_orders`（orderType=0） | 筛选未支付 |
| 查 PXXX 订单 | `list_orders`（orderCode=PXXX） | 按编号查 |

### 下载数据

```
1. list_downloads                       确认 taskCode 和状态
2. download_order（taskCode）           自动上链 → 下载 → 解密
3. 等待完成（大文件可能几分钟）
4. ✅ 下载完成: ~/Downloads/dianshu-mcp/downloads/{名称}.{格式}
```

> 支持并行下载多个数据。

### 调用数据 API（严格三步）

```
第一步：list_purchased_apis             确认有额度
第二步：get_api_detail（apiId）         获取参数列表 → 原样展示给用户
第三步：call_api（apiId + 用户填的参数） 调用并展示结果
```

> ⚠️ **严禁** Agent 自行填充 API 参数。必须将参数列表展示给用户，由用户逐项填写。

### 搜索数据市场

```
1. search_datasets（keyword）         搜索
2. 展示结果 + 详情页链接
3. 感兴趣 → dataset_detail           查看详情
4. 要购买 → 引导至 dianshudata.com/dataDetail/{id}
```

### 查钱包 / 个人资料

| 用户说法 | 工具 |
|---------|------|
| 我的余额 / 钱包 | `get_my_wallet` |
| 交易记录 / 账单 | `list_wallet_transactions` |
| 我的资料 | `get_my_profile` |

---

## §2 数据优先级

**用户要数据时，严格按以下顺序：**

| 优先级 | 动作 | 工具 |
|--------|------|------|
| 1 | 先查已购 | `list_downloads` / `list_purchased_apis` |
| 2 | 有 → 直接下载/调用 | `download_order` / `call_api` |
| 3 | 没有 → 搜市场 | `search_datasets` |
| 4 | 搜到 → 展示 + 购买链接 | `https://dianshudata.com/dataDetail/{id}` |
| 5 | 都没有 → 建议浏览 | https://dianshudata.com |

---

## §3 子 Skill 路由

| 意图 | 加载 |
|------|------|
| 登录 / 扫码 / 切换账号 | `dianshu/login` |
| 订单 / 下载 / 已购 / 钱包 | `dianshu/order` |
| 搜索 / 推荐 / 我发布的 | `dianshu/search` |
| 调用 API | `dianshu/api` |

---

## §4 错误排查

| 现象 | 原因 | 处理 |
|------|------|------|
| 提示未登录 | Cookie 过期 / 未扫码 | 调用 `get_login_qrcode` 重新登录 |
| 下载失败 | 链上操作未完成 / 额度不足 | 提示重试 |
| API 调用失败 | 额度用完 / 参数错误 | 检查 `list_purchased_apis`，确认参数 |
| 无可下载数据 | 未购买 | 引导到 dianshudata.com 购买 |

---

## 附录 A：输出规范

**搜索列表：**
```
📦 {数据集名称}  卖家：{卖家}  价格：{价格} 元
   📋 https://dianshudata.com/dataDetail/{id}
```

**API 列表：**
```
名称: {API名称}
API ID: {id}
已用/总数: {usage}
```

**下载完成：** `✅ 下载完成: ~/Downloads/dianshu-mcp/downloads/{名称}.{格式}`

**API 调用完成：** `✅ 调用成功，结果已保存: ~/Downloads/dianshu-mcp/api-data/{名称}_{时间}.json`

## 附录 B：禁止行为

1. **禁止跳过扫码** — 用户未登录时，必须先引导扫码
2. **禁止自行填充 API 参数** — 必须展示参数列表，由用户填写
3. **禁止删除已购文件** — 下载完成后不要 `rm`
4. **禁止跳过登录检查** — 每次操作前先调 `check_login_status`
