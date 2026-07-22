package main

import (
	"context"
	"encoding/base64"
	"runtime/debug"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
)

// InitMCPServer 初始化 MCP Server。
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

func boolPtr(value bool) *bool { return &value }

func withPanicRecovery[T any](
	toolName string,
	handler func(context.Context, *mcp.CallToolRequest, T) (*mcp.CallToolResult, any, error),
) func(context.Context, *mcp.CallToolRequest, T) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, args T) (result *mcp.CallToolResult, resp any, err error) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logrus.WithFields(logrus.Fields{"tool": toolName, "panic": recovered}).Error("Tool handler panicked")
				logrus.Errorf("Stack trace:\n%s", debug.Stack())
				result = &mcp.CallToolResult{
					Content: []mcp.Content{&mcp.TextContent{Text: "工具 " + toolName + " 执行时发生内部错误，请查看服务端日志。"}},
					IsError: true,
				}
				resp = nil
				err = nil
			}
		}()
		return handler(ctx, req, args)
	}
}

// registerTools 注册所有 MCP 工具。
func registerTools(server *mcp.Server, appServer *AppServer) {
	mcp.AddTool(server,
		&mcp.Tool{Name: "check_login_status", Description: "检查典枢平台的登录状态", Annotations: &mcp.ToolAnnotations{Title: "Check Login Status", ReadOnlyHint: true}},
		withPanicRecovery("check_login_status", func(ctx context.Context, req *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
			result := appServer.handleCheckLoginStatus(ctx)
			return convertToMCPResult(result), nil, nil
		}),
	)

	mcp.AddTool(server,
		&mcp.Tool{Name: "get_login_qrcode", Description: "获取微信登录二维码（返回 Base64 图片），用户扫码后登录典枢平台", Annotations: &mcp.ToolAnnotations{Title: "Get Login QR Code", ReadOnlyHint: true}},
		withPanicRecovery("get_login_qrcode", func(ctx context.Context, req *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
			result := appServer.handleGetLoginQRCode(ctx)
			return convertToMCPResult(result), nil, nil
		}),
	)

	mcp.AddTool(server,
		&mcp.Tool{Name: "delete_cookies", Description: "删除 cookies 文件，重置登录状态。删除后需要重新登录。", Annotations: &mcp.ToolAnnotations{Title: "Delete Cookies", DestructiveHint: boolPtr(true)}},
		withPanicRecovery("delete_cookies", func(ctx context.Context, req *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
			result := appServer.handleDeleteCookies(ctx)
			return convertToMCPResult(result), nil, nil
		}),
	)

	mcp.AddTool(server,
		&mcp.Tool{Name: "data_search", Description: "数据查询入口，例如：查询一下小红书里关于鸡腿的内容", Annotations: &mcp.ToolAnnotations{Title: "Data Search", ReadOnlyHint: true}},
		withPanicRecovery("data_search", func(ctx context.Context, req *mcp.CallToolRequest, args DataSearchArgs) (*mcp.CallToolResult, any, error) {
			result := appServer.handleDataSearch(ctx, args)
			return convertToMCPResult(result), nil, nil
		}),
	)

	logrus.Infof("Registered %d MCP tools", 4)
}

// convertToMCPResult 将自定义 MCPToolResult 转换为官方 SDK 格式。
func convertToMCPResult(result *MCPToolResult) *mcp.CallToolResult {
	var contents []mcp.Content
	for _, content := range result.Content {
		switch content.Type {
		case "text":
			contents = append(contents, &mcp.TextContent{Text: content.Text})
		case "image":
			imageData, err := base64.StdEncoding.DecodeString(content.Data)
			if err != nil {
				logrus.WithError(err).Error("Failed to decode base64 image data")
				contents = append(contents, &mcp.TextContent{Text: "图片数据解码失败: " + err.Error()})
			} else {
				contents = append(contents, &mcp.ImageContent{Data: imageData, MIMEType: content.MimeType})
			}
		}
	}
	return &mcp.CallToolResult{Content: contents, IsError: result.IsError}
}
