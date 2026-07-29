package main

import (
	"context"
	"fmt"
	"time"

	"dianshu-mcp/dianshu"

	"github.com/sirupsen/logrus"
)

// MCP 工具处理函数

// handleCheckLoginStatus 处理检查登录状态。
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
		text := fmt.Sprintf("✅ 已登录\n我的昵称: %s", status.Nickname)
		if status.UserInfo != nil {
			text = formatMyProfileText(status.UserInfo)
		}
		return &MCPToolResult{Content: []MCPContent{{Type: "text", Text: text}}}
	}

	return &MCPToolResult{Content: []MCPContent{{Type: "text", Text: "❌ 未登录\n请使用 get_login_qrcode 获取微信二维码并扫码登录。"}}}
}

// handleGetMyProfile 处理获取我的资料。
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

// handleGetWalletBalance 处理获取我的钱包。
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

// handleListWalletTransactions 处理获取交易明细。
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

// handleGetLoginQRCode 处理获取登录并等待扫码。
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

// handleDeleteCookies 处理删除 cookies。
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

// handleListOrders 处理查询订单列表。
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

// handleListDownloads 处理列出可下载数据产品。
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

// handleListPurchasedAPIs 处理列出已购买的 API 产品。
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

func formatLoginSuccessText(userInfo *dianshu.UserInfo) string {
	nickname := "用户"
	dsUserNo := "-"
	if userInfo != nil {
		if userInfo.Nickname != "" {
			nickname = userInfo.Nickname
		}
		if userInfo.DSUserNo != "" {
			dsUserNo = userInfo.DSUserNo
		}
	}

	return fmt.Sprintf("✅ 登录成功！\n我的昵称: %s\n典枢号: %s\n\n你现在可以使用数据查询功能了。", nickname, dsUserNo)
}

func formatMyProfileText(userInfo *dianshu.UserInfo) string {
	nickname := "-"
	dsUserNo := "-"
	description := "-"
	appCode := "-"

	if userInfo != nil {
		if userInfo.Nickname != "" {
			nickname = userInfo.Nickname
		}
		if userInfo.DSUserNo != "" {
			dsUserNo = userInfo.DSUserNo
		}
		if userInfo.Description != "" {
			description = userInfo.Description
		}
		if userInfo.AppCode != "" {
			appCode = userInfo.AppCode
		}
	}

	return fmt.Sprintf("👤 我的资料\n我的昵称: %s\n典枢号: %s\n卖家介绍: %s\n我的AppCode: %s", nickname, dsUserNo, description, appCode)
}

// handleDatasetSearch 处理典枢数据集搜索。
type DatasetSearchArgs struct {
	Keyword  string `json:"keyword" jsonschema:"搜索关键词，必填"`
	PageNo   int    `json:"pageNo,omitempty" jsonschema:"页码，可选，默认 1"`
	PageSize int    `json:"pageSize,omitempty" jsonschema:"每页条数，可选，默认 20"`
}

func (s *AppServer) handleDatasetSearch(ctx context.Context, args DatasetSearchArgs) *MCPToolResult {
	logrus.Infof("MCP: 典枢数据集搜索 (keyword=%s)", args.Keyword)

	result, err := s.dianshuService.SearchDatasets(ctx, args.Keyword, args.PageNo, args.PageSize)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "数据集搜索失败: " + err.Error()}},
			IsError: true,
		}
	}
	return result
}

// handleHomepageRecommend 获取典枢首页推荐数据。
func (s *AppServer) handleHomepageRecommend(ctx context.Context) *MCPToolResult {
	logrus.Info("MCP: 获取典枢首页推荐数据")

	result, err := s.dianshuService.GetHomepageRecommend(ctx)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "获取首页推荐失败: " + err.Error()}},
			IsError: true,
		}
	}
	return result
}

// handleListMyDatasets 获取我的数据集列表。
func (s *AppServer) handleListMyDatasets(ctx context.Context, args DatasetSearchArgs) *MCPToolResult {
	logrus.Infof("MCP: 我的数据集列表 (pageNo=%d)", args.PageNo)

	result, err := s.dianshuService.ListMyDatasets(ctx, args.PageNo, args.PageSize)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "获取我的数据集列表失败: " + err.Error()}},
			IsError: true,
		}
	}
	return result
}

// handleDatasetDetail 获取数据集详情。
type DatasetDetailArgs struct {
	DatasetID int `json:"datasetId" jsonschema:"数据集 ID，必填"`
}

func (s *AppServer) handleDatasetDetail(ctx context.Context, args DatasetDetailArgs) *MCPToolResult {
	logrus.Infof("MCP: 数据集详情 (datasetId=%d)", args.DatasetID)

	result, err := s.dianshuService.GetDatasetDetail(ctx, args.DatasetID)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "获取数据集详情失败: " + err.Error()}},
			IsError: true,
		}
	}
	return result
}

// handleDownloadOrder 处理下载订单。
type DownloadOrderArgs struct {
	TaskCode string `json:"taskCode" jsonschema:"任务编码，必填，如 P17848795595058039"`
}

func (s *AppServer) handleDownloadOrder(ctx context.Context, args DownloadOrderArgs) *MCPToolResult {
	logrus.Infof("MCP: 下载订单 (taskCode=%s)", args.TaskCode)

	result, err := s.dianshuService.DownloadOrder(ctx, args.TaskCode)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "下载订单失败: " + err.Error()}},
			IsError: true,
		}
	}
	return result
}
