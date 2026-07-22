package main

// MCPToolResult MCP 工具返回结果。
type MCPToolResult struct {
	Content []MCPContent `json:"content"`
	IsError bool         `json:"isError,omitempty"`
}

// MCPContent MCP 内容块。
type MCPContent struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Data     string `json:"data,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
}

// SuccessResponse API 成功响应。
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

// ErrorResponse API 错误响应。
type ErrorResponse struct {
	Error   string      `json:"error"`
	Code    string      `json:"code"`
	Details interface{} `json:"details"`
}

// DataSearchArgs 表示数据查询请求。
type DataSearchArgs struct {
	Query      string `json:"query" jsonschema:"原始查询内容，AI 可直接传用户问题"`
	Provider   string `json:"provider,omitempty" jsonschema:"数据源名称，可选，AI 可自动填写；默认 xiaohongshu"`
	Dataset    string `json:"dataset,omitempty" jsonschema:"数据集类型，可选，AI 可自动填写；默认 search"`
	SiteDomain string `json:"siteDomain,omitempty" jsonschema:"站点域名，可选，AI 可自动填写；默认 xiaohongshu.com"`
	Page       string `json:"page,omitempty" jsonschema:"页码，可选，默认 1"`
	Keyword    string `json:"keyword,omitempty" jsonschema:"搜索关键词，可选；若 AI 已提取则直接传"`
	StartTime  string `json:"startTime,omitempty" jsonschema:"开始时间，可选，Unix 秒时间戳；未传时默认回退到最近 7 天"`
	EndTime    string `json:"endTime,omitempty" jsonschema:"结束时间，可选，Unix 秒时间戳；未传时默认使用当前时间"`
}
