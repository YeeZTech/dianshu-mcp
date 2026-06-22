package main

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// MCP 工具处理函数

// handleCheckLoginStatus 处理检查登录状态
func (s *AppServer) handleCheckLoginStatus(ctx context.Context) *MCPToolResult {
	logrus.Info("MCP: 检查登录状态")

	status, err := s.dianshuService.CheckLoginStatus(ctx)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "检查登录状态失败: " + err.Error(),
			}},
			IsError: true,
		}
	}

	if status.IsLogin {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: fmt.Sprintf("✅ 已登录\n用户名: %s\n\n你可以使用其他功能了。", status.Nickname),
			}},
		}
	}

	return &MCPToolResult{
		Content: []MCPContent{{
			Type: "text",
			Text: "❌ 未登录\n请使用 get_login_qrcode 获取微信二维码并扫码登录。",
		}},
	}
}

// handleGetLoginQRCode 处理获取登录并等待扫码
// 1. 打开浏览器展示微信二维码（浏览器窗口可见，方便用户扫码）
// 2. 用户扫码后自动检测登录成功
// 3. 保存 token 并返回结果
func (s *AppServer) handleGetLoginQRCode(ctx context.Context) *MCPToolResult {
	logrus.Info("MCP: 获取登录二维码并等待扫码")

	result, err := s.dianshuService.GetLoginQRCode(ctx)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "登录失败: " + err.Error(),
			}},
			IsError: true,
		}
	}

	return result
}

// handleDeleteCookies 处理删除 cookies
func (s *AppServer) handleDeleteCookies(ctx context.Context) *MCPToolResult {
	logrus.Info("MCP: 删除 cookies")
	result, err := s.dianshuService.DeleteCookies(ctx)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "删除 cookies 失败: " + err.Error(),
			}},
			IsError: true,
		}
	}
	return result
}

// handleListOrders 处理查询订单列表
func (s *AppServer) handleListOrders(ctx context.Context, args ListOrdersArgs) *MCPToolResult {
	logrus.Infof("MCP: 查询订单 (orderType=%d, orderCode=%s)", args.OrderType, args.OrderCode)

	// 设置 30 秒超时
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	result, err := s.dianshuService.QueryOrders(ctx, args.OrderType, args.OrderCode)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "查询订单失败: " + err.Error(),
			}},
			IsError: true,
		}
	}
	return result
}

// handleListDownloads 处理列出可下载数据产品
func (s *AppServer) handleListDownloads(ctx context.Context) *MCPToolResult {
	logrus.Info("MCP: 列出可下载数据产品")

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	result, err := s.dianshuService.ListDownloads(ctx)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "查询下载列表失败: " + err.Error(),
			}},
			IsError: true,
		}
	}
	return result
}

// handleListPurchasedAPIs 处理列出已购买的 API 产品
func (s *AppServer) handleListPurchasedAPIs(ctx context.Context) *MCPToolResult {
	logrus.Info("MCP: 列出已购买的 API 产品")

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	result, err := s.dianshuService.ListPurchasedAPIs(ctx)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "查询 API 产品失败: " + err.Error(),
			}},
			IsError: true,
		}
	}
	return result
}
