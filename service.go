package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
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

// DianshuService 典枢业务服务。
type DianshuService struct {
	browserHeadless bool
	chargeService   dianshu.ChargeService
	dataQueryRouter *dianshu.DataQueryRouter
}

// NewDianshuService 创建典枢服务。
func NewDianshuService() *DianshuService {
	return &DianshuService{
		browserHeadless: configs.IsHeadless(),
		chargeService:   &dianshu.NoopChargeService{},
		dataQueryRouter: dianshu.NewDataQueryRouter(),
	}
}

// CheckLoginStatus 检查登录状态。
func (s *DianshuService) CheckLoginStatus(ctx context.Context) (*dianshu.LoginCheckResult, error) {
	allCookies := cookies.GetAllCookies()
	return dianshu.CheckLoginStatus(ctx, allCookies)
}

// GetMyProfile 获取当前登录用户资料。
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

// GetLoginQRCode 获取微信登录二维码（包含等待扫码）。
func (s *DianshuService) GetLoginQRCode(ctx context.Context) (*MCPToolResult, error) {
	_ = cookies.DeleteCookies()

	allCookies, err := dianshu.WaitForWeChatLogin(ctx, s.browserHeadless, 120*time.Second)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: fmt.Sprintf("❌ 登录失败: %v\n\n请重试获取二维码。", err)}},
		}, nil
	}

	savedData := make(map[string]interface{})
	for key, value := range allCookies {
		savedData[key] = value
	}
	if err := cookies.SetCookies(savedData); err != nil {
		logrus.Warnf("保存 cookies 失败: %v", err)
	}

	userInfo, _ := dianshu.GetUserInfo(ctx, allCookies)
	return &MCPToolResult{
		Content: []MCPContent{{Type: "text", Text: formatLoginSuccessText(userInfo)}},
	}, nil
}

// GetWalletBalance 获取我的钱包余额。
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

// ListWalletTransactions 获取钱包交易明细。
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

// QueryOrders 查询订单列表。
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

// ListDownloads 列出可下载数据产品。
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

// ListPurchasedAPIs 列出已购买的 API 产品。
func (s *DianshuService) ListPurchasedAPIs(ctx context.Context) (*MCPToolResult, error) {
	allCookies := cookies.GetAllCookies()
	if len(allCookies) == 0 {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "❌ 未登录，请先使用 get_login_qrcode 扫码登录"}},
			IsError: true,
		}, nil
	}

	client := dianshu.NewAPIClient(allCookies)
	result, err := client.ListPurchasedAPIs(ctx, 1, 50)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: fmt.Sprintf("❌ 查询 API 产品失败: %v", err)}},
			IsError: true,
		}, nil
	}

	return &MCPToolResult{Content: []MCPContent{{Type: "text", Text: formatPurchasedAPIListV2(result)}}}, nil
}

// GetQRCodeOnly 仅获取二维码图片（不等待登录）。
func (s *DianshuService) GetQRCodeOnly(ctx context.Context) (*MCPToolResult, error) {
	imgData, text, err := dianshu.GetLoginQRCodeOnly(ctx, s.browserHeadless)
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

// DeleteCookies 删除 cookies。
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

// SearchDatasets 典枢平台数据集搜索。
func (s *DianshuService) SearchDatasets(ctx context.Context, keyword string, pageNo, pageSize int) (*MCPToolResult, error) {
	allCookies := cookies.GetAllCookies()
	if len(allCookies) == 0 {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "❌ 未登录，请先使用 get_login_qrcode 扫码登录"}},
			IsError: true,
		}, nil
	}

	if pageNo <= 0 {
		pageNo = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	client := dianshu.NewAPIClient(allCookies)
	result, err := client.SearchDatasets(ctx, keyword, pageNo, pageSize)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: fmt.Sprintf("❌ 搜索数据集失败: %v", err)}},
			IsError: true,
		}, nil
	}

	return &MCPToolResult{
		Content: []MCPContent{{Type: "text", Text: formatDatasetSearchResult(result)}},
	}, nil
}

// GetHomepageRecommend 获取典枢首页推荐数据。
func (s *DianshuService) GetHomepageRecommend(ctx context.Context) (*MCPToolResult, error) {
	result, err := dianshu.GetHomepageRecommend(ctx)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: fmt.Sprintf("❌ 获取首页推荐失败: %v", err)}},
			IsError: true,
		}, nil
	}

	return &MCPToolResult{
		Content: []MCPContent{{Type: "text", Text: formatHomepageRecommendResult(result)}},
	}, nil
}

// ListMyDatasets 获取我的数据集列表。
func (s *DianshuService) ListMyDatasets(ctx context.Context, pageNo, pageSize int) (*MCPToolResult, error) {
	allCookies := cookies.GetAllCookies()
	if len(allCookies) == 0 {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "❌ 未登录，请先使用 get_login_qrcode 扫码登录"}},
			IsError: true,
		}, nil
	}
	if pageNo <= 0 {
		pageNo = 1
	}
	if pageSize <= 0 {
		pageSize = 12
	}
	client := dianshu.NewAPIClient(allCookies)
	result, err := client.ListMyDatasets(ctx, pageNo, pageSize)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: fmt.Sprintf("❌ 获取我的数据集列表失败: %v", err)}},
			IsError: true,
		}, nil
	}
	return &MCPToolResult{
		Content: []MCPContent{{Type: "text", Text: formatMyDatasetList(result)}},
	}, nil
}

// DownloadDataset 根据 taskCode 下载已购买的数据文件到 output/downloads/。
func (s *DianshuService) DownloadDataset(ctx context.Context, taskCode string) (*MCPToolResult, error) {
	allCookies := cookies.GetAllCookies()
	if len(allCookies) == 0 {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "❌ 未登录，请先使用 get_login_qrcode 扫码登录"}},
			IsError: true,
		}, nil
	}
	taskCode = strings.TrimSpace(taskCode)
	if taskCode == "" {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "❌ 请提供任务编码 (taskCode)"}},
			IsError: true,
		}, nil
	}

	client := dianshu.NewAPIClient(allCookies)
	tasks, err := client.ListTasks(ctx, 1, 100)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: fmt.Sprintf("❌ 查询订单列表失败: %v", err)}},
			IsError: true,
		}, nil
	}

	var target *dianshu.TaskItem
	for i := range tasks {
		if tasks[i].TaskCode == taskCode {
			target = &tasks[i]
			break
		}
	}
	if target == nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: fmt.Sprintf("❌ 未找到任务编码为 %s 的订单", taskCode)}},
			IsError: true,
		}, nil
	}
	if target.FileURL == "" {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "❌ 该订单没有可下载的文件"}},
			IsError: true,
		}, nil
	}

	dest := filepath.Join("output", "downloads", target.FileURL)
	absPath, _ := filepath.Abs(dest)
	if err := dianshu.DownloadFile(ctx, target.FileURL, dest); err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: fmt.Sprintf("❌ 下载失败: %v", err)}},
			IsError: true,
		}, nil
	}

	resultText := fmt.Sprintf("✅ 下载成功\n\n数据集: %s\n任务编码: %s\n文件: %s\nMEDIA:%s", target.DatasetName, target.TaskCode, absPath, absPath)
	return &MCPToolResult{Content: []MCPContent{{Type: "text", Text: resultText}}}, nil
}

// XhsSearch 执行小红书数据查询，查询结果写入 output/data-search/ 目录。
func (s *DianshuService) XhsSearch(ctx context.Context, args DataSearchArgs) (*MCPToolResult, error) {
	dataQueryRequest := buildDataQueryRequest(args)

	queryResult, err := s.dataQueryRouter.Query(ctx, dataQueryRequest)
	if err != nil {
		return nil, err
	}

	if err = s.chargeService.Charge(ctx, dianshu.ChargeRequest{
		ProviderType: dataQueryRequest.ProviderType,
		DatasetType:  dataQueryRequest.DatasetType,
		Amount:       "",
		Description:  "数据查询扣费预留入口",
	}); err != nil {
		return nil, err
	}

	resultText, err := persistRawResult(dataQueryRequest, queryResult)
	if err != nil {
		return nil, err
	}
	return &MCPToolResult{Content: []MCPContent{{Type: "text", Text: resultText}}}, nil
}

func buildDataQueryRequest(args DataSearchArgs) dianshu.DataQueryRequest {
	keyword := strings.TrimSpace(args.Keyword)
	if keyword == "" {
		keyword = strings.TrimSpace(args.Query)
	}
	page := strings.TrimSpace(args.Page)
	if page == "" {
		page = dianshu.XiaohongshuDefaultPage
	}
	endTime := strings.TrimSpace(args.EndTime)
	if endTime == "" {
		endTime = fmt.Sprintf("%d", time.Now().Unix())
	}
	startTime := strings.TrimSpace(args.StartTime)
	if startTime == "" {
		startTime = fmt.Sprintf("%d", time.Now().Add(-7*24*time.Hour).Unix())
	}

	return dianshu.DataQueryRequest{
		ProviderType: pickString(args.Provider, dianshu.ProviderTypeXiaohongshu),
		DatasetType:  pickString(args.Dataset, dianshu.DatasetTypeSearch),
		SiteDomain:   pickString(args.SiteDomain, dianshu.XiaohongshuSiteDomain),
		Body: map[string]string{
			"startTime": startTime,
			"endTime":   endTime,
			"keyword":   keyword,
			"page":      page,
		},
		RawQuery: strings.TrimSpace(args.Query),
	}
}

func pickString(value string, defaultVal string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return defaultVal
	}
	return value
}

func persistRawResult(queryReq dianshu.DataQueryRequest, result *dianshu.DataQueryResult) (string, error) {
	if len(result.RawJSON) == 0 {
		return "", fmt.Errorf("查询结果为空，无法写入 JSON 文件")
	}

	outputDir := filepath.Join("output", "data-search")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", fmt.Errorf("创建输出目录失败: %w", err)
	}

	fileName := fmt.Sprintf("%s_%s_%s.json", queryReq.ProviderType, queryReq.DatasetType, time.Now().Format("20060102_150405"))
	filePath := filepath.Join(outputDir, fileName)
	if err := os.WriteFile(filePath, result.RawJSON, 0o644); err != nil {
		return "", fmt.Errorf("写入查询结果文件失败: %w", err)
	}

	requestBodyJSON, err := json.MarshalIndent(queryReq.Body, "", "  ")
	if err != nil {
		return "", fmt.Errorf("序列化查询参数失败: %w", err)
	}

	resultText := fmt.Sprintf("✅ 查询成功\n查询内容: %s\n\n数据源: %s/%s\n站点: %s\nDSSeqNo: %s\n结果状态: %d / %s\n查询参数:\n%s\n\n结果文件: %s\n\n%s", queryReq.RawQuery, queryReq.ProviderType, queryReq.DatasetType, queryReq.SiteDomain, result.DSSeqNo, result.ResultCode, result.ResultDesc, string(requestBodyJSON), filePath, string(result.RawJSON))
	return resultText, nil
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

func defaultText(v string) string {
	if v == "" {
		return "-"
	}
	return v
}

func formatDatasetSearchResult(result *dianshu.DatasetSearchResponse) string {
	if result == nil || len(result.Data) == 0 {
		return "未找到匹配的数据集"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔍 搜索结果（第 %d/%d 页，共 %d 条）\n", result.Page.PageNo, result.Page.TotalPage, result.Page.Count))

	maxItems := len(result.Data)
	if maxItems > 10 {
		maxItems = 10
	}

	for i := 0; i < maxItems; i++ {
		item := result.Data[i]
		sb.WriteString(fmt.Sprintf("\n\n【%d】%s\n", i+1, item.DatasetName))
		sb.WriteString(fmt.Sprintf("价格: ¥%.2f\n", item.Price))
		sb.WriteString(fmt.Sprintf("格式: %s\n", defaultText(item.Pattern)))
		sb.WriteString(fmt.Sprintf("大小: %s\n", formatFileSize(float64(item.DatasetSize))))
		sb.WriteString(fmt.Sprintf("卖家: %s\n", defaultText(item.CreateCompanyName)))
		sb.WriteString(fmt.Sprintf("标签: %s\n", defaultText(item.Tag)))
		sb.WriteString(fmt.Sprintf("销量: %s\n", defaultText(item.SalesVolume)))
		sb.WriteString(fmt.Sprintf("数据集编码: %s\n", item.DatasetCode))
	}

	if len(result.Data) > maxItems {
		sb.WriteString(fmt.Sprintf("\n\n...... 其余 %d 条未展开", len(result.Data)-maxItems))
	}
	return sb.String()
}

func formatFileSize(size float64) string {
	const unit = 1024.0
	if size < unit {
		return fmt.Sprintf("%.0f B", size)
	}
	size /= unit
	if size < unit {
		return fmt.Sprintf("%.1f KB", size)
	}
	size /= unit
	return fmt.Sprintf("%.1f MB", size)
}

func formatHomepageRecommendResult(resp *dianshu.HomepageRecommendResponse) string {
	if resp == nil || len(resp.Data.QueryRecommend.Data) == 0 {
		return "暂无首页推荐数据"
	}

	var sb strings.Builder
	sb.WriteString("🏠 典枢首页推荐数据\n")

	blocks := resp.Data.QueryRecommend.Data
	for _, block := range blocks {
		if len(block.Details) == 0 {
			continue
		}
		sb.WriteString(fmt.Sprintf("\n📌 %s\n", block.Name))

		maxItems := len(block.Details)
		if maxItems > 5 {
			maxItems = 5
		}
		for i := 0; i < maxItems; i++ {
			d := block.Details[i]
			name := d.Name
			if d.DatasetName != nil && *d.DatasetName != "" {
				name = *d.DatasetName
			}
			sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, name))
			if d.Description != "" {
				desc := d.Description
				if len(desc) > 80 {
					desc = desc[:80] + "..."
				}
				sb.WriteString(fmt.Sprintf("     %s\n", desc))
			}
			if d.Price != nil && *d.Price != "" {
				sb.WriteString(fmt.Sprintf("     价格: ¥%s\n", *d.Price))
			}
			if d.Pattern != nil && *d.Pattern != "" {
				sb.WriteString(fmt.Sprintf("     格式: %s\n", *d.Pattern))
			}
		}
		if len(block.Details) > maxItems {
			sb.WriteString(fmt.Sprintf("  ... 其余 %d 条未展开\n", len(block.Details)-maxItems))
		}
	}
	return sb.String()
}

func formatMyDatasetList(result *dianshu.MyDatasetListResponse) string {
	if result == nil || len(result.Data) == 0 {
		return "你还没有发布过数据集"
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📦 我的数据集（第 %d/%d 页，共 %d 个）\n", result.Page.PageNo, result.Page.TotalPage, result.Page.Count))
	for i, item := range result.Data {
		statusText := "未知"
		switch item.Status {
		case 1:
			statusText = "审核中"
		case 2:
			statusText = "已发布"
		case 3:
			statusText = "已下架"
		}
		t := time.UnixMilli(item.CreateTime).Format("2006-01-02 15:04")
		sb.WriteString(fmt.Sprintf("\n【%d】%s\n", i+1, item.DatasetName))
		sb.WriteString(fmt.Sprintf("价格: ¥%.2f | 状态: %s | %s\n", item.Price, statusText, t))
		sb.WriteString(fmt.Sprintf("编码: %s\n", item.DatasetCode))
	}
	return sb.String()
}

func formatPurchasedAPIListV2(result *dianshu.PurchasedAPIListResponse) string {
	if result == nil || len(result.Data) == 0 {
		return "暂无已购买的 API 产品"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔌 已购买 API 产品（第 %d/%d 页，共 %d 个）\n", result.Page.PageNo, result.Page.TotalPage, result.Page.Count))
	for i, item := range result.Data {
		sb.WriteString(fmt.Sprintf("\n【%d】%s\n", i+1, item.APIName))
		sb.WriteString(fmt.Sprintf("API 编码: %s\n", item.APICode))
		sb.WriteString(fmt.Sprintf("调用次数: %s\n", item.Usage))
		sb.WriteString(fmt.Sprintf("购买时间: %s\n", item.CreateTime))
	}
	return sb.String()
}
