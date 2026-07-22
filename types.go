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

// OrderQueryRequest 订单查询请求。
type OrderQueryRequest struct {
	OrderType int    `json:"orderType"`
	OrderCode string `json:"orderCode"`
}

// NameValueParam 表示 MCP 传入的键值参数。
type NameValueParam struct {
	Name  string `json:"name" jsonschema:"参数名，必填"`
	Value string `json:"value" jsonschema:"参数值，必填"`
}

// CallPurchasedAPIArgs 表示调用已购买 API 的请求参数。
type CallPurchasedAPIArgs struct {
	APICode      string           `json:"apiCode" jsonschema:"API 标识，必填"`
	Method       string           `json:"method,omitempty" jsonschema:"请求方式：GET 或 POST，默认 POST"`
	QueryParams  []NameValueParam `json:"queryParams,omitempty" jsonschema:"GET 请求 query 参数列表"`
	BodyParams   []NameValueParam `json:"bodyParams,omitempty" jsonschema:"POST 请求 body 参数列表"`
	HeaderParams []NameValueParam `json:"headerParams,omitempty" jsonschema:"卖家 API 透传 header 参数列表"`
}
