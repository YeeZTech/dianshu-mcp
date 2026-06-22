package main

// MCPToolResult MCP 工具返回结果
type MCPToolResult struct {
	Content []MCPContent `json:"content"`
	IsError bool         `json:"isError,omitempty"`
}

// MCPContent MCP 内容块
type MCPContent struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Data     string `json:"data,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
}

// SuccessResponse API 成功响应
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

// ErrorResponse API 错误响应
type ErrorResponse struct {
	Error   string      `json:"error"`
	Code    string      `json:"code"`
	Details interface{} `json:"details"`
}

// OrderQueryRequest 订单查询请求
type OrderQueryRequest struct {
	OrderType int    `json:"orderType"`
	OrderCode string `json:"orderCode"`
}
