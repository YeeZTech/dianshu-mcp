// Package main - MCP server initialization and tool registration.
//
// Author: zhyyao

package main

import (
	"context"
	"runtime/debug"

	"dianshu-mcp/handler"
	"dianshu-mcp/logger"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ── MCP Server ─────────────────────────────────────────

func initMCPServer(h *handler.App) *mcp.Server {
	srv := mcp.NewServer(
		&mcp.Implementation{Name: "dianshu-mcp", Version: "1.0.0"},
		nil,
	)
	registerTools(srv, h)
	logger.Info("MCP server initialized", "tools", 16)
	return srv
}

// boolPtr returns a pointer to the given bool value.
func boolPtr(b bool) *bool { return &b }

// ── Tool Registration ──────────────────────────────────

func registerTools(srv *mcp.Server, h *handler.App) {
	// Auth
	add0(srv, "check_login_status", "检查典枢平台的登录状态", true, false, h.CheckLoginStatus)
	add0(srv, "get_login_qrcode", "获取微信登录二维码（返回 Base64 图片），用户扫码后登录典枢平台", true, false, h.GetLoginQRCode)
	add0(srv, "delete_cookies", "删除 cookies 文件，重置登录状态", false, true, h.DeleteCookies)

	// Order / Download
	add1(srv, "list_orders", "查询典枢平台的订单列表，支持按类型和订单编号筛选", true, h.ListOrders)
	add1(srv, "list_downloads", "列出已购买的典枢数据产品及下载信息", true, h.ListDownloads)
	add1(srv, "download_order", "通过任务编码下载并解密数据文件", false, h.DownloadOrder)

	// API
	add1(srv, "list_purchased_apis", "列出已购买的典枢 API 产品及调用信息", true, h.ListPurchasedAPIs)
	add1(srv, "get_api_detail", "获取 API 的参数列表。拿到参数后务必先展示给用户，让用户填写具体参数值", true, h.GetAPIDetail)
	add1(srv, "call_api", "调用数据 API。params 的值必须由用户明确提供", false, h.CallAPI)

	// Dataset
	add1(srv, "search_datasets", "搜索典枢平台上的数据集，可按关键词查找", true, h.SearchDatasets)
	add1(srv, "dataset_detail", "获取典枢数据集的详细信息，需提供数据集 ID", true, h.DatasetDetail)
	add0(srv, "homepage_recommend", "获取典枢平台首页推荐数据", true, false, h.HomepageRecommend)
	add1(srv, "my_datasets", "获取我在典枢平台上发布的数据集列表", true, h.MyDatasets)

	// User
	add0(srv, "get_my_profile", "获取当前登录账号的资料信息", true, false, h.GetMyProfile)
	add0(srv, "get_my_wallet", "获取当前登录账号的钱包余额信息", true, false, h.GetMyWallet)
	add1(srv, "list_wallet_transactions", "查询钱包交易明细，支持分页", true, h.ListWalletTransactions)
}

// ── Helpers ────────────────────────────────────────────

// add0 registers a tool with no arguments.
func add0(srv *mcp.Server, name, desc string, readOnly, destructive bool, fn func(context.Context) *handler.ToolResult) {
	ann := &mcp.ToolAnnotations{Title: name, ReadOnlyHint: readOnly}
	if destructive { ann.DestructiveHint = boolPtr(true) }
	mcp.AddTool(srv,
		&mcp.Tool{Name: name, Description: desc, Annotations: ann},
		withRecover(name, func(ctx context.Context, req *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
			r := fn(ctx)
			return toCallResult(r), nil, nil
		}),
	)
}

// add1
// add1 registers a tool with one typed argument.
func add1[T any](srv *mcp.Server, name, desc string, readOnly bool, fn func(context.Context, T) *handler.ToolResult) {
	ann := &mcp.ToolAnnotations{Title: name, ReadOnlyHint: readOnly}
	mcp.AddTool(srv,
		&mcp.Tool{Name: name, Description: desc, Annotations: ann},
		withRecover(name, func(ctx context.Context, req *mcp.CallToolRequest, args T) (*mcp.CallToolResult, any, error) {
			r := fn(ctx, args)
			return toCallResult(r), nil, nil
		}),
	)
// withRecover wraps a tool handler with panic recovery.
}


// withRecover 包装工具 handler 增加 panic 恢复。
// withRecover wraps a tool handler with panic recovery.

func withRecover[T any](name string, fn func(context.Context, *mcp.CallToolRequest, T) (*mcp.CallToolResult, any, error)) func(context.Context, *mcp.CallToolRequest, T) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, args T) (*mcp.CallToolResult, any, error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("tool panic", "tool", name, "panic", r)
				logger.Error("stack trace", "trace", string(debug.Stack()))
			}
		}()
		return fn(ctx, req, args)
	// toCallResult converts a ToolResult to an MCP CallToolResult.
	}
}


// toCallResult 转换 ToolResult 为 MCP CallToolResult。
// toCallResult converts a ToolResult to an MCP CallToolResult.

func toCallResult(r *handler.ToolResult) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: r.Content, IsError: r.IsError}
}
