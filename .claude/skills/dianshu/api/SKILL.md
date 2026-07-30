---
name: dianshu/api
description: |
  管理已购数据 API：查看列表、获取参数、调用 API。当用户调用 API 时使用。
---

## 前置

调用 `check_login_status`。未登录 → 引导登录。

## 查看已购 API

调用 `list_purchased_apis`。

## 调用流程（两步，顺序不可颠倒）

### 第一步：获取参数

调用 `get_api_detail`（`apiId`），将返回的参数列表**原样展示给用户**，询问用户填写每个参数的值。

**⚠️ 禁止行为：**
- 禁止在未展示参数列表的情况下直接调用 call_api
- 禁止看到参数后自行推测合理的值
- 禁止使用"默认值""空字符串"等填充参数
- 禁止参考"示例值"直接填入

### 第二步：用户确认后调用

**用户逐项明确提供参数值后**，调用 `call_api`：

```
call_api {
  apiId: <数字>,
  params: { <用户提供的 key: value> }
}
```

### 第三步：展示结果

展示调用结果，告知保存路径 `~/Downloads/dianshu-mcp/api-data/`。

## 示例

用户："用微博热搜查热搜"

1. `list_purchased_apis` → apiId=10736
2. `get_api_detail {"apiId": 10736}` → 展示参数列表
3. 询问："该 API 有以下参数：page(页码)、pageSize(每页数量)。请告诉我你需要什么值？"
4. 用户回复 "page: 1" → `call_api {"apiId": 10736, "params": {"page": "1"}}`

| 场景 | 处理 |
|---|---|
| 无已购 API | 引导购买 |
| 用户未提供参数 | 再次询问，不要自行填充 |
| 调用失败 | 检查参数，告知错误 |
