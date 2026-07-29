package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"dianshu-mcp/cookies"
	"dianshu-mcp/dianshu"
)

// handleGetAPIDetail 获取 API 详情（含参数列表）。
type GetAPIDetailArgs struct {
	APIID int `json:"apiId" jsonschema:"API ID（数字），从 list_purchased_apis 获取"`
}

func (s *AppServer) handleGetAPIDetail(ctx context.Context, args GetAPIDetailArgs) *MCPToolResult {
	allCookies := cookies.GetAllCookies()
	if len(allCookies) == 0 {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "未登录"}},
			IsError: true,
		}
	}
	cookieList := make([]string, 0, len(allCookies))
	for k, v := range allCookies {
		cookieList = append(cookieList, k+"="+v)
	}

	detail, err := dianshu.GetAPIDetail(ctx, &http.Client{}, args.APIID, "", cookieList)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "查询失败: " + err.Error()}},
			IsError: true,
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("API 详情: %s\n\n", detail.APIName))
	if detail.Description != "" {
		sb.WriteString(fmt.Sprintf("说明: %s\n\n", detail.Description))
	}
	if detail.RequestMethod == 1 {
		sb.WriteString("请求方法: GET\n\n")
	} else {
		sb.WriteString("请求方法: POST\n\n")
	}
	if len(detail.QueryParams) > 0 {
		sb.WriteString("Query 参数:\n")
		for _, p := range detail.QueryParams {
			req := ""
			if p.Required == 1 {
				req = " [必填]"
			}
			sb.WriteString(fmt.Sprintf("  %s (%s)%s: %s\n", p.ParamName, p.TypeName, req, p.Description))
		}
		sb.WriteString("\n")
	}
	if len(detail.BodyParams) > 0 {
		sb.WriteString("Body 参数:\n")
		for _, p := range detail.BodyParams {
			req := ""
			if p.Required == 1 {
				req = " [必填]"
			}
			sb.WriteString(fmt.Sprintf("  %s (%s)%s: %s\n", p.ParamName, p.TypeName, req, p.Description))
		}
		sb.WriteString("\n")
	}
	if len(detail.ReqHeaders) > 0 {
		sb.WriteString("请求头:\n")
		for _, h := range detail.ReqHeaders {
			sb.WriteString(fmt.Sprintf("  %s (%s): %s\n", h.ParamName, h.TypeName, h.Description))
		}
	}
	return &MCPToolResult{
		Content: []MCPContent{{Type: "text", Text: sb.String()}},
	}
}

// handleCallAPI 调用数据 API。
type CallAPIArgs struct {
	APIID  int               `json:"apiId" jsonschema:"API ID（数字），从 list_purchased_apis 获取"`
	Params map[string]string `json:"params" jsonschema:"API 参数 key-value"`
	Method string            `json:"method,omitempty" jsonschema:"GET 或 POST，不传自动判断"`
}

func (s *AppServer) handleCallAPI(ctx context.Context, args CallAPIArgs) *MCPToolResult {
	result, err := s.dianshuService.CallAPI(ctx, args.APIID, args.Params, args.Method)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "调用 API 失败: " + err.Error()}},
			IsError: true,
		}
	}
	return result
}
