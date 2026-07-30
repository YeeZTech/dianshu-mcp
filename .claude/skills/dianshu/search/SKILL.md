---
name: dianshu/search
description: |
  搜索典枢数据市场：关键词搜索、首页推荐、数据集详情、购买引导。当用户找数据、搜索时使用。
---

## 前置

调用 `check_login_status`。未登录可搜索，但提示登录后可购买。

## 关键词搜索

调用 `search_datasets`（`keyword`），展示结果：
```
📦 {名称}  卖家：{卖家}  价格：{价格} 元
   📋 https://dianshudata.com/dataDetail/{id}
```

## 首页推荐

无明确关键词时调用 `homepage_recommend`。

## 数据集详情

感兴趣时调用 `dataset_detail`（`datasetId`）。

## 购买引导

- 已登录 → 引导购买链接 + 提示购买后可帮下载
- 未登录 → 引导先登录

| 场景 | 处理 |
|---|---|
| 无结果 | 换关键词或访问 dianshudata.com |
