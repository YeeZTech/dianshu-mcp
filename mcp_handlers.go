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

// handleDataSearch 处理数据查询。
func (s *AppServer) handleDataSearch(ctx context.Context, args DataSearchArgs) *MCPToolResult {
	logrus.Infof("MCP: 数据查询 (query=%s)", args.Query)

	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	result, err := s.dianshuService.DataSearch(ctx, args)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "数据查询失败: " + err.Error()}},
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
