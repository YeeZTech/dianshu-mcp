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
			Content: []MCPContent{{Type: "text", Text: fmt.Sprintf("删除 cookies 失败: %v", err)}},
			IsError: true,
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

// ListDownloads 列出可下载的数据产品
func (s *DianshuService) ListDownloads(ctx context.Context) (*MCPToolResult, error) {
	allCookies := cookies.GetAllCookies()
	if len(allCookies) == 0 {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "❌ 未登录，请先使用 /dianshu-login 扫码登录"}},
			IsError: true,
		}, nil
	}

	client := dianshu.NewAPIClient(allCookies)
	tasks, err := client.ListTasks(ctx, 1, 50)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: fmt.Sprintf("❌ 查询任务列表失败: %v", err)}},
			IsError: true,
		}, nil
	}

	// 筛选可下载的数据产品（有 fileUrl 或 downloadList 的）
	var downloads []dianshu.TaskItem
	for _, t := range tasks {
		if t.FileURL != "" || t.Pattern != "" {
			downloads = append(downloads, t)
		}
	}

	if len(downloads) == 0 {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "暂无已购买的可下载数据产品"}},
		}, nil
	}

	return &MCPToolResult{
		Content: []MCPContent{{Type: "text", Text: formatDownloadList(downloads)}},
	}, nil
}

// ListPurchasedAPIs 列出已购买的 API 产品
func (s *DianshuService) ListPurchasedAPIs(ctx context.Context) (*MCPToolResult, error) {
	allCookies := cookies.GetAllCookies()
	if len(allCookies) == 0 {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "❌ 未登录，请先使用 /dianshu-login 扫码登录"}},
			IsError: true,
		}, nil
	}

	client := dianshu.NewAPIClient(allCookies)
	tasks, err := client.ListTasks(ctx, 1, 50)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: fmt.Sprintf("❌ 查询任务列表失败: %v", err)}},
			IsError: true,
		}, nil
	}

	// 筛选 API 类型的产品（APIType == 1，或者没有 fileUrl 但有 datasetId 的）
	var apiTasks []dianshu.TaskItem
	for _, t := range tasks {
		if t.APIType == 1 || (t.FileURL == "" && t.DatasetID > 0) {
			apiTasks = append(apiTasks, t)
		}
	}

	if len(apiTasks) == 0 {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "暂无已购买的 API 产品"}},
		}, nil
	}

	// 获取每个 API 产品的详情
	var details []string
	for _, t := range apiTasks {
		detail, err := client.GetAPIDetail(ctx, t.DatasetID)
		if err != nil {
			details = append(details, formatAPISummary(t, nil, err.Error()))
			continue
		}
		details = append(details, formatAPISummary(t, detail, ""))
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("🔌 已购买的 API 产品（共 %d 条）\n\n", len(details)))
	for _, d := range details {
		b.WriteString(d)
		b.WriteString("\n")
	}
	b.WriteString("💡 可通过 DSAPIClient SDK 或直接 HTTP 调用 API\n")

	return &MCPToolResult{
		Content: []MCPContent{{Type: "text", Text: b.String()}},
	}, nil
}

// formatDownloadList 格式化可下载数据产品列表
func formatDownloadList(items []dianshu.TaskItem) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("📥 可下载数据产品（共 %d 条）\n\n", len(items)))

	for i, item := range items {
		b.WriteString(fmt.Sprintf("─── %d ───\n", i+1))
		b.WriteString(fmt.Sprintf("  数据名称: %s\n", item.DatasetName))
		b.WriteString(fmt.Sprintf("  任务编号: %s\n", item.TaskCode))

		if item.Pattern != "" {
			b.WriteString(fmt.Sprintf("  文件格式: %s\n", item.Pattern))
		}

		// 显示下载链接
		for key, urls := range item.DownloadList {
			for _, u := range urls {
				b.WriteString(fmt.Sprintf("  下载地址[%s]: %s\n", key, u.URL))
			}
		}

		if item.ChecksumUrl != "" {
			b.WriteString(fmt.Sprintf("  校验地址: %s\n", item.ChecksumUrl))
		}
		if item.ClientDownloadUrl != "" {
			b.WriteString(fmt.Sprintf("  客户端下载: %s\n", item.ClientDownloadUrl))
		}

		createTime := "-"
		if item.CreateTime > 0 {
			t := time.UnixMilli(item.CreateTime)
			createTime = t.Format("2006-01-02 15:04")
		}
		b.WriteString(fmt.Sprintf("  购买时间: %s\n", createTime))
		b.WriteString("\n")
	}

	b.WriteString("💡 文件为 .sealed（加密封包格式），需使用典枢客户端解密\n")
	return b.String()
}

// formatAPISummary 格式化单个 API 产品信息
func formatAPISummary(task dianshu.TaskItem, detail *dianshu.APIDetail, errMsg string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("  API 名称: %s\n", task.DatasetName))

	if errMsg != "" {
		b.WriteString(fmt.Sprintf("  ⚠️ 获取详情失败: %s\n", errMsg))
		return b.String()
	}

	if detail == nil {
		return b.String()
	}

	b.WriteString(fmt.Sprintf("  API 类型: %s\n", detail.ApiType))
	method := "GET"
	if detail.RequestMethod == 1 {
		method = "POST"
	}
	b.WriteString(fmt.Sprintf("  请求方式: %s\n", method))

	if detail.ApiEndpoint != "" {
		b.WriteString(fmt.Sprintf("  调用地址: %s\n", detail.ApiEndpoint))
	}

	// 参数列表
	if len(detail.BodyParams) > 0 {
		b.WriteString("  参数:\n")
		for _, p := range detail.BodyParams {
			req := "可选"
			if p.Required {
				req = "必填"
			}
			example := ""
			if p.Example != "" {
				example = fmt.Sprintf(" (示例: %s)", p.Example)
			}
			b.WriteString(fmt.Sprintf("    - %s (%s, %s)%s: %s\n", p.Name, p.Type, req, example, p.Description))
		}
	}

	// Java 代码示例（截取前 200 字符）
	if detail.JavaRequestExample != "" {
		code := detail.JavaRequestExample
		if len(code) > 200 {
			code = code[:200] + "..."
		}
		b.WriteString(fmt.Sprintf("  Java 调用示例:\n    %s\n", code))
	}

	// 多语言代码示例
	for _, ex := range detail.ExampleCodeList {
		code := ex.Code
		if len(code) > 150 {
			code = code[:150] + "..."
		}
		b.WriteString(fmt.Sprintf("  [%s] 代码示例:\n    %s\n", ex.Language, code))
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
