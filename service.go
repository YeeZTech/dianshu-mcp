package main

import (
	"context"
	"encoding/base64"
	"fmt"
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

// GetMyProfile 获取当前登录用户资料
func (s *DianshuService) GetMyProfile(ctx context.Context) (*dianshu.UserInfo, error) {
	allCookies := cookies.GetAllCookies()
	if len(allCookies) == 0 {
		return nil, fmt.Errorf("未登录，请先使用 get_login_qrcode 扫码登录")
	}

	userInfo, err := dianshu.GetUserInfo(ctx, allCookies)
	if err != nil {
		return nil, err
	}
	if userInfo == nil {
		return nil, fmt.Errorf("未获取到用户资料")
	}
	return userInfo, nil
}

// GetWalletBalance 获取我的钱包
func (s *DianshuService) GetWalletBalance(ctx context.Context) (*MCPToolResult, error) {
	allCookies := cookies.GetAllCookies()
	if len(allCookies) == 0 {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "❌ 未登录，请先使用 get_login_qrcode 扫码登录"}},
			IsError: true,
		}, nil
	}

	client := dianshu.NewAPIClient(allCookies)
	balance, err := client.GetWalletBalance(ctx)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: fmt.Sprintf("❌ 获取钱包余额失败: %v", err)}},
			IsError: true,
		}, nil
	}

	return &MCPToolResult{
		Content: []MCPContent{{Type: "text", Text: formatWalletBalanceText(balance)}},
	}, nil
}

// ListWalletTransactions 获取交易明细
func (s *DianshuService) ListWalletTransactions(ctx context.Context, page dianshu.PageRequest) (*MCPToolResult, error) {
	allCookies := cookies.GetAllCookies()
	if len(allCookies) == 0 {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "❌ 未登录，请先使用 get_login_qrcode 扫码登录"}},
			IsError: true,
		}, nil
	}
	if page.PageNo <= 0 {
		page.PageNo = 1
	}
	if page.PageSize <= 0 {
		page.PageSize = 10
	}

	client := dianshu.NewAPIClient(allCookies)
	result, err := client.ListWalletTransactions(ctx, page)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: fmt.Sprintf("❌ 获取交易明细失败: %v", err)}},
			IsError: true,
		}, nil
	}

	return &MCPToolResult{
		Content: []MCPContent{{Type: "text", Text: formatWalletTransactionsText(result)}},
	}, nil
}

// GetLoginQRCode 获取微信登录二维码（包含等待扫码）
func (s *DianshuService) GetLoginQRCode(ctx context.Context) (*MCPToolResult, error) {
	_ = cookies.DeleteCookies()

	allCookies, err := dianshu.WaitForWeChatLogin(ctx, s.browserHeadless, 120*time.Second)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: fmt.Sprintf("❌ 登录失败: %v\n\n请重试获取二维码。", err)}},
		}, nil
	}

	savedData := make(map[string]interface{})
	for k, v := range allCookies {
		savedData[k] = v
	}
	if err := cookies.SetCookies(savedData); err != nil {
		logrus.Warnf("保存 cookies 失败: %v", err)
	}

	userInfo, _ := dianshu.GetUserInfo(ctx, allCookies)
	return &MCPToolResult{
		Content: []MCPContent{{Type: "text", Text: formatLoginSuccessText(userInfo)}},
	}, nil
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

	return fmt.Sprintf("✅ 登录成功！\n我的昵称: %s\n典枢号: %s\n\n你现在可以使用订单查询等功能了。", nickname, dsUserNo)
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

func formatWalletBalanceText(balance *dianshu.WalletBalance) string {
	if balance == nil {
		return "暂无钱包数据"
	}

	profit := "-"
	if balance.Profit != nil {
		profit = fmt.Sprintf("%.2f", *balance.Profit)
	}

	return fmt.Sprintf("💰 我的钱包\n可用余额: %.2f\n冻结金额: %.2f\n累计收益: %s\n可提现金额: %.2f", balance.Available, balance.Frozen, profit, balance.WithDrawable)
}

func formatWalletTransactionsText(result *dianshu.WalletTransactionListResponse) string {
	if result == nil || len(result.Data) == 0 {
		return "暂无交易明细"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🧾 交易明细（第 %d/%d 页，共 %d 条）\n\n", result.Page.PageNo, result.Page.TotalPage, result.Page.Count))
	for i, item := range result.Data {
		timeText := time.UnixMilli(item.Time).Format("2006-01-02 15:04:05")
		sb.WriteString(fmt.Sprintf(
			"【明细 %d】\n类型: %s\n状态: %s\n金额: %.2f\n到账金额: %.2f\n服务费: %.2f\n数据集: %s\n订单号: %s\n创建人: %s\n时间: %s\n\n",
			i+1,
			defaultText(item.Type),
			defaultText(item.Status),
			item.Amount1,
			item.Amount2,
			item.ServiceCharge,
			defaultText(item.DatasetName),
			defaultText(item.Code),
			defaultText(item.CreateUser),
			timeText,
		))
	}
	return sb.String()
}

// GetQRCodeOnly 仅获取二维码图片（不等待登录）
func (s *DianshuService) GetQRCodeOnly(ctx context.Context) (*MCPToolResult, error) {
	imgData, text, err := dianshu.GetLoginQRCode(ctx, s.browserHeadless)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: fmt.Sprintf("无法获取二维码: %v\n\n请手动打开以下链接登录：\n%s", err, dianshu.WeChatQRLoginURL)}},
		}, nil
	}

	content := []MCPContent{{Type: "text", Text: text}}
	if len(imgData) > 0 {
		content = append(content, MCPContent{Type: "image", Data: base64.StdEncoding.EncodeToString(imgData), MimeType: "image/png"})
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
		Content: []MCPContent{{Type: "text", Text: fmt.Sprintf("Cookies 已成功删除，登录状态已重置。\n\n文件目录: %s\n\n下次操作时，需要重新登录。", dir)}},
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

	return &MCPToolResult{Content: []MCPContent{{Type: "text", Text: formatTaskList(tasks)}}}, nil
}

// ListDownloads 列出可下载数据产品
func (s *DianshuService) ListDownloads(ctx context.Context) (*MCPToolResult, error) {
	allCookies := cookies.GetAllCookies()
	if len(allCookies) == 0 {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "❌ 未登录，请先使用 get_login_qrcode 扫码登录"}},
			IsError: true,
		}, nil
	}

	client := dianshu.NewAPIClient(allCookies)
	tasks, err := client.ListTasks(ctx, 1, 50)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: fmt.Sprintf("❌ 查询下载列表失败: %v", err)}},
			IsError: true,
		}, nil
	}

	return &MCPToolResult{Content: []MCPContent{{Type: "text", Text: formatDownloadList(tasks)}}}, nil
}

// ListPurchasedAPIs 列出已购买的 API 产品
func (s *DianshuService) ListPurchasedAPIs(ctx context.Context) (*MCPToolResult, error) {
	allCookies := cookies.GetAllCookies()
	if len(allCookies) == 0 {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "❌ 未登录，请先使用 get_login_qrcode 扫码登录"}},
			IsError: true,
		}, nil
	}

	client := dianshu.NewAPIClient(allCookies)
	tasks, err := client.ListTasks(ctx, 1, 50)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: fmt.Sprintf("❌ 查询 API 产品失败: %v", err)}},
			IsError: true,
		}, nil
	}

	var apiTasks []dianshu.TaskItem
	for _, task := range tasks {
		if task.APIType == 1 || (task.FileURL == "" && task.DatasetID > 0) {
			apiTasks = append(apiTasks, task)
		}
	}

	return &MCPToolResult{Content: []MCPContent{{Type: "text", Text: formatPurchasedAPIList(apiTasks)}}}, nil
}

// CallPurchasedAPI 调用已购买的 API 产品。
func (s *DianshuService) CallPurchasedAPI(ctx context.Context, args CallPurchasedAPIArgs) (*MCPToolResult, error) {
	profile, err := s.GetMyProfile(ctx)
	if err != nil {
		return nil, err
	}
	if profile == nil || strings.TrimSpace(profile.AppCode) == "" {
		return nil, fmt.Errorf("当前登录账号缺少 appCode，无法调用已购买 API")
	}

	dataAPIContext, err := dianshu.NewDataAPIContext(profile.AppCode)
	if err != nil {
		return nil, err
	}
	gatewayClient, err := dianshu.NewDataAPIGatewayClient(dataAPIContext)
	if err != nil {
		return nil, err
	}

	requestMethod := strings.TrimSpace(args.Method)
	if requestMethod == "" {
		requestMethod = "POST"
	}

	bodyParams := convertToDataAPIParams(args.BodyParams)
	queryParams := convertToDataAPIParams(args.QueryParams)
	headerParams := convertToDataAPIParams(args.HeaderParams)

	var responseText string
	switch strings.ToUpper(requestMethod) {
	case "POST":
		responseText, err = gatewayClient.CallPost(ctx, args.APICode, dianshu.DataAPIPostRequest{
			BodyParams:     bodyParams,
			RequestHeaders: headerParams,
		})
	case "GET":
		responseText, err = gatewayClient.CallGet(ctx, args.APICode, dianshu.DataAPIGetRequest{
			QueryParams:    queryParams,
			RequestHeaders: headerParams,
		})
	default:
		return nil, fmt.Errorf("不支持的请求方式: %s", requestMethod)
	}
	if err != nil {
		return nil, err
	}

	resultText := fmt.Sprintf("✅ 已调用购买的 API\nAPI 标识: %s\n请求方式: %s\n\n返回结果:\n%s", args.APICode, strings.ToUpper(requestMethod), responseText)
	return &MCPToolResult{Content: []MCPContent{{Type: "text", Text: resultText}}}, nil
}

func formatTaskList(tasks []dianshu.TaskItem) string {
	if len(tasks) == 0 {
		return "暂无订单数据"
	}

	var sb strings.Builder
	for i, task := range tasks {
		statusText := "未知"
		switch task.Status {
		case 0:
			statusText = "待支付"
		case 1:
			statusText = "进行中"
		case 2:
			statusText = "已完成"
		case 3:
			statusText = "已取消"
		}

		payStatusText := "未知"
		switch task.PayStatus {
		case 0:
			payStatusText = "未支付"
		case 1:
			payStatusText = "已支付"
		case 2:
			payStatusText = "退款中"
		case 3:
			payStatusText = "已退款"
		}

		sb.WriteString(fmt.Sprintf("【订单 %d】\n任务编码: %s\n数据集: %s\n卖家: %s\n金额: %.4f\n状态: %s\n支付状态: %s\n创建时间: %s\n\n", i+1, task.TaskCode, task.DatasetName, task.DatasetUserName, task.Price, statusText, payStatusText, task.CreateTimeSql))
	}
	return sb.String()
}

func formatDownloadList(tasks []dianshu.TaskItem) string {
	var downloads []dianshu.TaskItem
	for _, task := range tasks {
		if task.FileURL != "" || task.Pattern != "" || len(task.DownloadList) > 0 {
			downloads = append(downloads, task)
		}
	}
	if len(downloads) == 0 {
		return "暂无已购买的可下载数据产品"
	}

	var sb strings.Builder
	for i, task := range downloads {
		sb.WriteString(fmt.Sprintf("【下载 %d】\n数据集: %s\n任务编码: %s\n文件类型: %s\n", i+1, task.DatasetName, task.TaskCode, defaultText(task.Pattern)))
		if task.FileURL != "" {
			sb.WriteString(fmt.Sprintf("文件标识: %s\n", task.FileURL))
		}
		if task.ClientDownloadUrl != "" {
			sb.WriteString(fmt.Sprintf("客户端下载地址: %s\n", task.ClientDownloadUrl))
		}
		if task.ChecksumUrl != "" {
			sb.WriteString(fmt.Sprintf("校验文件地址: %s\n", task.ChecksumUrl))
		}
		for group, urls := range task.DownloadList {
			for _, item := range urls {
				sb.WriteString(fmt.Sprintf("%s下载地址: %s\n", group, item.URL))
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func formatPurchasedAPIList(tasks []dianshu.TaskItem) string {
	if len(tasks) == 0 {
		return "暂无已购买的 API 产品"
	}

	var sb strings.Builder
	for i, task := range tasks {
		sb.WriteString(fmt.Sprintf("【API %d】\n名称: %s\n任务编码: %s\n卖家: %s\n\n", i+1, task.DatasetName, task.TaskCode, task.DatasetUserName))
	}
	return sb.String()
}

func defaultText(v string) string {
	if v == "" {
		return "-"
	}
	return v
}
