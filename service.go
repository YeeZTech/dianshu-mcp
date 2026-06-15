package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dianshu-mcp/configs"
	"dianshu-mcp/cookies"
	"dianshu-mcp/dianshu"

	"github.com/sirupsen/logrus"
)

// DianshuService 典枢业务服务
type DianshuService struct {
	browserHeadless bool
}

// NewDianshuService 创建典枢服务
func NewDianshuService() *DianshuService {
	return &DianshuService{
		browserHeadless: configs.IsHeadless(),
	}
}

// CheckLoginStatus 检查登录状态
func (s *DianshuService) CheckLoginStatus(ctx context.Context) (*dianshu.LoginCheckResult, error) {
	allCookies := cookies.GetAllCookies()
	return dianshu.CheckLoginStatus(ctx, allCookies)
}

// GetLoginQRCode 获取微信登录二维码（包含等待扫码）
// 返回二维码图片 + 等待用户扫码，扫码成功后自动保存 token
func (s *DianshuService) GetLoginQRCode(ctx context.Context) (*MCPToolResult, error) {
	// 先删除旧 token，确保重新登录
	_ = cookies.DeleteCookies()

	// 获取二维码并等待登录，超时 120 秒
	allCookies, err := dianshu.WaitForWeChatLogin(ctx, s.browserHeadless, 120*time.Second)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{
				{
					Type: "text",
					Text: fmt.Sprintf("❌ 登录失败: %v\n\n请重试获取二维码。", err),
				},
			},
		}, nil
	}

	// 保存所有 cookies
	savedData := make(map[string]interface{})
	for k, v := range allCookies {
		savedData[k] = v
	}
	if err := cookies.SetCookies(savedData); err != nil {
		logrus.Warnf("保存 cookies 失败: %v", err)
	}

	// 获取用户信息
	userInfo, err := dianshu.GetUserInfo(ctx, allCookies)
	nickname := "用户"
	if err == nil && userInfo != nil {
		nickname = userInfo.Nickname
	}

	return &MCPToolResult{
		Content: []MCPContent{
			{
				Type: "text",
				Text: fmt.Sprintf("✅ 登录成功！\n欢迎: %s\n\n你现在可以使用订单查询等功能了。", nickname),
			},
		},
	}, nil
}

// GetQRCodeOnly 仅获取二维码图片（不等待登录）
func (s *DianshuService) GetQRCodeOnly(ctx context.Context) (*MCPToolResult, error) {
	imgData, text, err := dianshu.GetLoginQRCode(ctx, s.browserHeadless)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{
				{
					Type: "text",
					Text: fmt.Sprintf("无法获取二维码: %v\n\n请手动打开以下链接登录：\n%s", err, dianshu.WeChatQRLoginURL),
				},
			},
		}, nil
	}

	content := []MCPContent{
		{Type: "text", Text: text},
	}

	if len(imgData) > 0 {
		content = append(content, MCPContent{
			Type:     "image",
			Data:     base64.StdEncoding.EncodeToString(imgData),
			MimeType: "image/png",
		})
	}

	return &MCPToolResult{Content: content}, nil
}

// DeleteCookies 删除 cookies
func (s *DianshuService) DeleteCookies(ctx context.Context) (*MCPToolResult, error) {
	filePath := cookies.GetCookiesFilePath()
	if err := cookies.DeleteCookies(); err != nil {
		return &MCPToolResult{
			Content:  []MCPContent{{Type: "text", Text: fmt.Sprintf("删除 cookies 失败: %v", err)}},
			IsError:  true,
		}, nil
	}

	absPath, _ := filepath.Abs(filePath)
	dir := filepath.Dir(absPath)
	return &MCPToolResult{
		Content: []MCPContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Cookies 已成功删除，登录状态已重置。\n\n文件目录: %s\n\n下次操作时，需要重新登录。", dir),
			},
		},
	}, nil
}

// QueryOrders 查询订单列表
func (s *DianshuService) QueryOrders(ctx context.Context, orderType int, orderCode string) (*MCPToolResult, error) {
	allCookies := cookies.GetAllCookies()
	if len(allCookies) == 0 {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "❌ 未登录，请先使用 /dianshu-login 扫码登录"}},
			IsError: true,
		}, nil
	}

	client := dianshu.NewAPIClient(allCookies)
	tasks, err := client.ListTasks(ctx, 1, 20)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: fmt.Sprintf("❌ 查询订单失败: %v", err)}},
			IsError: true,
		}, nil
	}

	return &MCPToolResult{
		Content: []MCPContent{{Type: "text", Text: formatTaskList(tasks)}},
	}, nil
}

// formatTaskList 格式化任务/订单列表
func formatTaskList(tasks []dianshu.TaskItem) string {
	if len(tasks) == 0 {
		return "暂无订单数据"
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("📋 任务/订单列表（共 %d 条）\n\n", len(tasks)))

	for i, task := range tasks {
		b.WriteString(fmt.Sprintf("─── %d ───\n", i+1))
		b.WriteString(fmt.Sprintf("  数据名称: %s\n", task.DatasetName))
		b.WriteString(fmt.Sprintf("  任务编号: %s\n", task.TaskCode))
		b.WriteString(fmt.Sprintf("  价格: ¥%.2f\n", task.Price))
		b.WriteString(fmt.Sprintf("  卖家: %s\n", task.DatasetUserName))

		createTime := "-"
		if task.CreateTime > 0 {
			t := time.UnixMilli(task.CreateTime)
			createTime = t.Format("2006-01-02 15:04")
		}
		b.WriteString(fmt.Sprintf("  购买时间: %s\n", createTime))

		status := "未知"
		switch task.TaskStatus {
		case 0:
			status = "待处理"
		case 1:
			status = "处理中"
		case 2, 3:
			status = "已完成"
		case 4:
			status = "已取消"
		case 5:
			status = "退款中"
		}
		b.WriteString(fmt.Sprintf("  状态: %s\n", status))
		b.WriteString("\n")
	}

	return b.String()
}

// formatOrderList 格式化订单统计（保留备用）
func formatOrderList(data *dianshu.OrderQueryData) string {
	if data == nil {
		return "暂无订单数据"
	}
	if len(data.OrderList) == 0 {
		return fmt.Sprintf("暂无订单（共 %d 条）", data.Total)
	}

	var b strings.Builder
	totalPages := (data.Total + data.PageSize - 1) / data.PageSize
	b.WriteString(fmt.Sprintf("📋 订单列表（共 %d 条，当前第 %d/%d 页）\n\n", data.Total, data.PageNo, totalPages))

	for i, order := range data.OrderList {
		b.WriteString(fmt.Sprintf("─── 订单 %d ───\n", i+1))
		b.WriteString(fmt.Sprintf("  订单编号: %s\n", order.OrderCode))
		b.WriteString(fmt.Sprintf("  订单名称: %s\n", order.OrderName))
		b.WriteString(fmt.Sprintf("  订单金额: %s\n", order.OrderAmount))
		b.WriteString(fmt.Sprintf("  订单价格: %s\n", order.OrderPrice))
		b.WriteString(fmt.Sprintf("  订单状态: %s\n", order.OrderStatus))
		b.WriteString(fmt.Sprintf("  支付状态: %s\n", order.OrderPayStatus))
		b.WriteString(fmt.Sprintf("  支付方式: %s\n", order.OrderPayWay))
		b.WriteString(fmt.Sprintf("  创建时间: %s\n", order.OrderCreateTime))
		b.WriteString("\n")
	}

	return b.String()
}

// SaveQRCodeImage 保存二维码图片到临时文件
func SaveQRCodeImage(imgData []byte) (string, error) {
	tmpDir := os.TempDir()
	filePath := filepath.Join(tmpDir, "dianshu-qrcode.png")
	if err := os.WriteFile(filePath, imgData, 0644); err != nil {
		return "", err
	}
	return filePath, nil
}
