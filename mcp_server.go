package main

import (
	"context"
	"encoding/base64"
	"runtime/debug"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
)

// InitMCPServer 初始化 MCP Server
func InitMCPServer(appServer *AppServer) *mcp.Server {
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "dianshu-mcp",
			Version: "1.0.0",
		},
		nil,
	)

	registerTools(server, appServer)

	logrus.Info("MCP Server initialized with official SDK")
	return server
}

func boolPtr(b bool) *bool { return &b }

// ListOrdersArgs 查询订单参数
type ListOrdersArgs struct {
	OrderType int    `json:"orderType" jsonschema:"订单类型：0-全部(默认)，1-数据集，2-API"`
	OrderCode string `json:"orderCode,omitempty" jsonschema:"订单编号（可选），留空查询所有订单"`
}

func withPanicRecovery[T any](
	toolName string,
	handler func(context.Context, *mcp.CallToolRequest, T) (*mcp.CallToolResult, any, error),
) func(context.Context, *mcp.CallToolRequest, T) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, args T) (result *mcp.CallToolResult, resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				logrus.WithFields(logrus.Fields{
					"tool":  toolName,
					"panic": r,
				}).Error("Tool handler panicked")
				logrus.Errorf("Stack trace:\n%s", debug.Stack())
				result = &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "工具 " + toolName + " 执行时发生内部错误，请查看服务端日志。",
						},
					},
					IsError: true,
				}
				resp = nil
				err = nil
			}
		}()
		return handler(ctx, req, args)
	}
}

// registerTools 注册所有 MCP 工具
func registerTools(server *mcp.Server, appServer *AppServer) {
	// 工具 1: 检查登录状态
	mcp.AddTool(server,
		&mcp.Tool{
			Name:        "check_login_status",
			Description: "检查典枢平台的登录状态",
			Annotations: &mcp.ToolAnnotations{
				Title:        "Check Login Status",
				ReadOnlyHint: true,
			},
		},
		withPanicRecovery("check_login_status", func(ctx context.Context, req *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
			result := appServer.handleCheckLoginStatus(ctx)
			return convertToMCPResult(result), nil, nil
		}),
	)

	// 工具 2: 获取登录二维码
	mcp.AddTool(server,
		&mcp.Tool{
			Name:        "get_login_qrcode",
			Description: "获取微信登录二维码（返回 Base64 图片），用户扫码后登录典枢平台",
			Annotations: &mcp.ToolAnnotations{
				Title:        "Get Login QR Code",
				ReadOnlyHint: true,
			},
		},
		withPanicRecovery("get_login_qrcode", func(ctx context.Context, req *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
			result := appServer.handleGetLoginQRCode(ctx)
			return convertToMCPResult(result), nil, nil
		}),
	)

	// 工具 3: 删除 cookies
	mcp.AddTool(server,
		&mcp.Tool{
			Name:        "delete_cookies",
			Description: "删除 cookies 文件，重置登录状态。删除后需要重新登录。",
			Annotations: &mcp.ToolAnnotations{
				Title:           "Delete Cookies",
				DestructiveHint: boolPtr(true),
			},
		},
		withPanicRecovery("delete_cookies", func(ctx context.Context, req *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
			result := appServer.handleDeleteCookies(ctx)
			return convertToMCPResult(result), nil, nil
		}),
	)

	// 工具 4: 查询订单
	mcp.AddTool(server,
		&mcp.Tool{
			Name:        "list_orders",
			Description: "查询典枢平台的订单列表，支持按类型和订单编号筛选",
			Annotations: &mcp.ToolAnnotations{
				Title:        "List Orders",
				ReadOnlyHint: true,
			},
		},
		withPanicRecovery("list_orders", func(ctx context.Context, req *mcp.CallToolRequest, args ListOrdersArgs) (*mcp.CallToolResult, any, error) {
			result := appServer.handleListOrders(ctx, args)
			return convertToMCPResult(result), nil, nil
		}),
	)

	// 工具 5: 列出可下载的数据产品
	mcp.AddTool(server,
		&mcp.Tool{
			Name:        "list_downloads",
			Description: "列出已购买的典枢数据产品及下载信息",
			Annotations: &mcp.ToolAnnotations{
				Title:        "List Downloads",
				ReadOnlyHint: true,
			},
		},
		withPanicRecovery("list_downloads", func(ctx context.Context, req *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
			result := appServer.handleListDownloads(ctx)
			return convertToMCPResult(result), nil, nil
		}),
	)

	// 工具 6: 列出已购买的 API 产品
	mcp.AddTool(server,
		&mcp.Tool{
			Name:        "list_purchased_apis",
			Description: "列出已购买的典枢 API 产品及调用信息",
			Annotations: &mcp.ToolAnnotations{
				Title:        "List Purchased APIs",
				ReadOnlyHint: true,
			},
		},
		withPanicRecovery("list_purchased_apis", func(ctx context.Context, req *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
			result := appServer.handleListPurchasedAPIs(ctx)
			return convertToMCPResult(result), nil, nil
		}),
	)

	logrus.Infof("Registered %d MCP tools", 6)
}

// convertToMCPResult 将自定义 MCPToolResult 转换为官方 SDK 格式
func convertToMCPResult(result *MCPToolResult) *mcp.CallToolResult {
	var contents []mcp.Content
	for _, c := range result.Content {
		switch c.Type {
		case "text":
			contents = append(contents, &mcp.TextContent{Text: c.Text})
		case "image":
			imageData, err := base64.StdEncoding.DecodeString(c.Data)
			if err != nil {
				logrus.WithError(err).Error("Failed to decode base64 image data")
				contents = append(contents, &mcp.TextContent{
					Text: "图片数据解码失败: " + err.Error(),
				})
			} else {
				contents = append(contents, &mcp.ImageContent{
					Data:     imageData,
					MIMEType: c.MimeType,
				})
			}
		}
	}
	return &mcp.CallToolResult{
		Content: contents,
		IsError: result.IsError,
	}
}
