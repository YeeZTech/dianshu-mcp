package main

import (
	"context"
	"fmt"
	"time"

	"dianshu-mcp/dianshu"

	"github.com/sirupsen/logrus"
)

// WalletTransactionArgs 钱包交易明细参数
type WalletTransactionArgs struct {
	PageNo   int `json:"pageNo" jsonschema:"页码，默认 1"`
	PageSize int `json:"pageSize" jsonschema:"每页条数，默认 10"`
}

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

	// 登录状态检查
	if status.IsLogin {
		text := fmt.Sprintf("✅ 已登录\n我的昵称: %s", status.Nickname)
		if status.UserInfo != nil {
			text = formatMyProfileText(status.UserInfo)
		}
		return &MCPToolResult{Content: []MCPContent{{Type: "text", Text: text}}}
	}

	return &MCPToolResult{Content: []MCPContent{{Type: "text", Text: "❌ 未登录\n请使用 get_login_qrcode 获取微信二维码并扫码登录。"}}}
}

// handleGetMyProfile 处理获取我的资料
func (s *AppServer) handleGetMyProfile(ctx context.Context) *MCPToolResult {
	logrus.Info("MCP: 获取我的资料")

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	userInfo, err := s.dianshuService.GetMyProfile(ctx)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "获取我的资料失败: " + err.Error()}},
			IsError: true,
		}
	}

	return &MCPToolResult{Content: []MCPContent{{Type: "text", Text: formatMyProfileText(userInfo)}}}
}

// handleGetWalletBalance 处理获取我的钱包
func (s *AppServer) handleGetWalletBalance(ctx context.Context) *MCPToolResult {
	logrus.Info("MCP: 获取我的钱包")

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	result, err := s.dianshuService.GetWalletBalance(ctx)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "获取我的钱包失败: " + err.Error()}},
			IsError: true,
		}
	}

	return result
}

// handleListWalletTransactions 处理获取交易明细
func (s *AppServer) handleListWalletTransactions(ctx context.Context, args WalletTransactionArgs) *MCPToolResult {
	logrus.Infof("MCP: 获取交易明细 (pageNo=%d, pageSize=%d)", args.PageNo, args.PageSize)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	result, err := s.dianshuService.ListWalletTransactions(ctx, dianshu.PageRequest{PageNo: args.PageNo, PageSize: args.PageSize})
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "获取交易明细失败: " + err.Error()}},
			IsError: true,
		}
	}

	return result
}

// handleGetLoginQRCode 处理获取登录并等待扫码
func (s *AppServer) handleGetLoginQRCode(ctx context.Context) *MCPToolResult {
	logrus.Info("MCP: 获取登录二维码并等待扫码")

	s.loginMu.Lock()
	if s.loginInProgress {
		s.loginMu.Unlock()
		return &MCPToolResult{Content: []MCPContent{{Type: "text", Text: "已有登录流程正在进行，请直接在已打开的浏览器窗口中完成扫码。"}}}
	}
	s.loginInProgress = true
	s.loginMu.Unlock()

	defer func() {
		s.loginMu.Lock()
		s.loginInProgress = false
		s.loginMu.Unlock()
	}()

	result, err := s.dianshuService.GetLoginQRCode(ctx)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "登录失败: " + err.Error()}},
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
			Content: []MCPContent{{Type: "text", Text: "删除 cookies 失败: " + err.Error()}},
			IsError: true,
		}
	}

	return result
}

// handleListOrders 处理查询订单列表
func (s *AppServer) handleListOrders(ctx context.Context, args ListOrdersArgs) *MCPToolResult {
	logrus.Infof("MCP: 查询订单 (orderType=%d, orderCode=%s)", args.OrderType, args.OrderCode)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	result, err := s.dianshuService.QueryOrders(ctx, args.OrderType, args.OrderCode)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "查询订单失败: " + err.Error()}},
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
			Content: []MCPContent{{Type: "text", Text: "查询下载列表失败: " + err.Error()}},
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
			Content: []MCPContent{{Type: "text", Text: "查询 API 产品失败: " + err.Error()}},
			IsError: true,
		}
	}

	return result
}
