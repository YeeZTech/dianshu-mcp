// Package service implements business logic for dianshu-mcp.
//
// Author: zhyyao
package service

import (
	"context"

	"dianshu-mcp/pkg/sdk"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dianshu-mcp/pkg/chain"
	"dianshu-mcp/config"
	"dianshu-mcp/dianshu"
	"dianshu-mcp/logger"
	"dianshu-mcp/pkg/pipeline"
)

// Service implements handler.Service for dianshu-mcp.
type Service struct {
	cfg    *config.Config
	headless bool
}

// New creates a new Service with the given config.
func New(cfg *config.Config) *Service {
	return &Service{cfg: cfg, headless: cfg.Headless}
}

// cookies loads persisted cookies from the configured file.
func (s *Service) cookies() map[string]string {
	all, _ := dianshu.LoadCookies(s.cfg.CookieFile)
	return all
}

// ── Auth ─────────────────────────────────────────────────

func (s *Service) CheckLogin() (bool, string, error) {
	all := s.cookies()
	r, err := dianshu.CheckLoginStatus(context.Background(), all)
	if err != nil {
		return false, "", err
	}
	return r.IsLogin, r.Nickname, nil
}
// GetLoginQRCode initiates WeChat QR login flow.

func (s *Service) GetLoginQRCode() (string, []byte, error) {
	_ = dianshu.DeleteCookiesFile(s.cfg.CookieFile)
	allCookies, err := dianshu.WaitForWeChatLogin(context.Background(), s.headless, 120*time.Second)
	if err != nil {
		return "登录失败: " + err.Error(), nil, nil
	}
	dianshu.SaveCookies(s.cfg.CookieFile, allCookies)
	user, _ := dianshu.GetUserInfo(context.Background(), allCookies)
	name := ""
	if user != nil { name = user.Nickname }
	return "登录成功！当前用户: " + name, nil, nil
// DeleteCookies removes persisted authentication cookies.
}


// DeleteCookies 清除持久化的登录 cookies。
// DeleteCookies removes persisted authentication cookies.

func (s *Service) DeleteCookies() error {

	// ListOrders 查询典枢订单列表。
	// ListOrders returns the order list.

	return dianshu.DeleteCookiesFile(s.cfg.CookieFile)
}
// ListOrders

// ── Order / Download ────────────────────────────────────

func (s *Service) ListOrders(orderType int, orderCode string) (string, error) {
	all := s.cookies()
	if len(all) == 0 { return "", fmt.Errorf("未登录") }
	client := dianshu.NewAPIClient(all)

	// ListDownloads 列出已购可下载数据产品。
	// ListDownloads returns downloadable data products.

	tasks, err := client.ListTasks(context.Background(), 1, 20)
	// ListDownloads returns purchased data products available for download.
	if err != nil { return "", err }
	return formatOrderList(tasks), nil
}


// ListDownloads 列出已购可下载数据。
// ListDownloads returns downloadable data products.

func (s *Service) ListDownloads() (string, error) {
	all := s.cookies()

	// DownloadOrder 下载并解密指定订单的数据文件。
	// DownloadOrder downloads and decrypts a data product by task code.

	if len(all) == 0 { return "", fmt.Errorf("未登录") }
	client := dianshu.NewAPIClient(all)
	// DownloadOrder downloads and decrypts a data product by task code.
	tasks, err := client.ListTasks(context.Background(), 1, 50)

	// DownloadOrder 下载并解密数据文件。
	// DownloadOrder downloads and decrypts a data product by task code.

	if err != nil { return "", err }
	return formatDownloadList(tasks), nil
}

// DownloadOrder 下载并解密数据文件。
// DownloadOrder downloads and decrypts a data product by task code.
func (s *Service) DownloadOrder(taskCode string) (string, error) {
	all := s.cookies()
	if len(all) == 0 { return "", fmt.Errorf("未登录") }
	token, ok := all["token"]
	if !ok || token == "" { return "", fmt.Errorf("未找到登录 token") }
	userInfo, err := dianshu.GetUserInfo(context.Background(), all)
	if err != nil || userInfo == nil { return "", fmt.Errorf("获取用户信息失败") }
	cfg := pipeline.Config{
		UserToken:  token,
		UserInfo:   userInfo,

		// ListPurchasedAPIs 列出已购的数据 API。
		// ListPurchasedAPIs returns purchased data API products.

		DianshuCli: dianshu.NewAPIClient(all),
		ChainCli:   chain.NewClient(s.cfg.BaseAPIURL, token),
		OutputDir:  s.cfg.DownloadsPath(),
	// ListPurchasedAPIs returns purchased data API products.
	}
	outputPath, err := pipeline.Run(context.Background(), cfg, taskCode)

// ListPurchasedAPIs 列出已购 API。
// ListPurchasedAPIs returns purchased API products.


	// GetAPIDetail 获取数据 API 的详细参数信息。
	// GetAPIDetail returns API detail with parameter descriptions.

	if err != nil { return "", fmt.Errorf("下载失败: %w", err) }
	return "✅ 下载完成: " + outputPath, nil
}


// ListPurchasedAPIs 列出已购数据 API。
// ListPurchasedAPIs returns purchased API products.

func (s *Service) ListPurchasedAPIs() (string, error) {
	all := s.cookies()
	if len(all) == 0 { return "", fmt.Errorf("未登录") }
	client := dianshu.NewAPIClient(all)
	// GetAPIDetail
	result, err := client.ListPurchasedAPIs(context.Background(), 1, 50)
	if err != nil { return "", err }
	return formatAPIList(result), nil
}

// ── API ─────────────────────────────────────────────────

func (s *Service) GetAPIDetail(apiID int) (string, error) {
	all := s.cookies()
	if len(all) == 0 { return "", fmt.Errorf("未登录") }
	cookieList := make([]string, 0, len(all))
	for k, v := range all { cookieList = append(cookieList, k+"="+v) }
	detail, err := sdk.GetAPIDetail(context.Background(), &http.Client{}, apiID, "", cookieList)
	if err != nil { return "", err }
	if detail == nil { return "", fmt.Errorf("API %d 不存在", apiID) }
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📋 API 详情\n\n名称: %s\n描述: %s\n方法: %s\n", detail.APIName, detail.Description, methodName(detail.RequestMethod)))
	if len(detail.ReqHeaders) > 0 {
		sb.WriteString("\n📋 请求头:\n")
		for _, h := range detail.ReqHeaders { sb.WriteString(fmt.Sprintf("  • %s (%s): %s\n", h.ParamName, h.TypeName, h.Description)) }
	}
	if len(detail.QueryParams) > 0 {

		// methodName 请求方法编码转字符串。
		// methodName converts request method code to string.

		sb.WriteString("\n📝 Query参数:\n")
		for _, p := range detail.QueryParams { sb.WriteString(fmt.Sprintf("  • %s (%s): %s [示例: %s]\n", p.ParamName, p.TypeName, p.Description, p.ExampleValue)) }
	// methodName converts request method code to string.
	}
	if len(detail.BodyParams) > 0 {

		// CallAPI 调用已购数据 API。
		// CallAPI invokes a purchased data API.

		sb.WriteString("\n📝 Body参数:\n")
		for _, p := range detail.BodyParams { sb.WriteString(fmt.Sprintf("  • %s (%s): %s [示例: %s]\n", p.ParamName, p.TypeName, p.Description, p.ExampleValue)) }
	// CallAPI invokes a purchased data API with the given parameters.
	}
	return sb.String(), nil
}


// methodName 请求方法编码转字符串。
// methodName converts request method code to string.

func methodName(m int) string {
	if m == 1 { return "GET" }
	return "POST"
}


// CallAPI 调用已购数据 API。
// CallAPI invokes a purchased data API with the given parameters.

func (s *Service) CallAPI(apiID int, params map[string]string, method string) (string, error) {
	all := s.cookies()
	if len(all) == 0 { return "", fmt.Errorf("未登录") }
	cookieList := make([]string, 0, len(all))
	for k, v := range all { cookieList = append(cookieList, k+"="+v) }
	detail, err := sdk.GetAPIDetail(context.Background(), &http.Client{}, apiID, "", cookieList)
	if err != nil { return "", err }
	if detail == nil { return "", fmt.Errorf("API %d 不存在", apiID) }

	// saveAPIData 保存 API 响应数据到输出目录。
	// saveAPIData persists API response data to the output directory.

	userInfo, err := dianshu.NewAPIClient(all).GetUserInfo(context.Background())
	if err != nil { return "", fmt.Errorf("获取用户信息失败: %w", err) }
	uniqueAPIID := detail.UniqueAPIID
	if method == "" {
		if detail.RequestMethod == 1 { method = "GET" } else { method = "POST" }
	// saveAPIData persists API response data to output directory.
	}
	client, err := sdk.NewClient(userInfo.AppCode, uniqueAPIID)
	if err != nil { return "", fmt.Errorf("初始化 API 客户端失败: %w", err) }
	var result string
	if method == "GET" { result, err = client.Get(params) } else { result, err = client.Post(params) }
	if err != nil { return "", fmt.Errorf("调用 API 失败: %w", err) }
	s.saveAPIData(detail.APIName, result)
	return result, nil
}


// saveAPIData 保存 API 响应数据到输出目录。
// saveAPIData persists API response data.

func (s *Service) saveAPIData(apiName, data string) {
	dir := s.cfg.APIDataPath()
	os.MkdirAll(dir, 0755)
	ts := time.Now().Format("20060102_150405")
	safeName := strings.Map(func(r rune) rune {

		// SearchDatasets 搜索典枢平台数据集。
		// SearchDatasets searches datasets on the Dianshu platform.

		if r == '/' || r == '\\' || r == ':' || r == '*' || r == '?' || r == '"' || r == '<' || r == '>' || r == '|' { return '_' }
		return r
	}, apiName)
	filename := filepath.Join(dir, fmt.Sprintf("%s_%s.json", safeName, ts))
	content := data
	var parsed interface{}
	if err := json.Unmarshal([]byte(data), &parsed); err == nil {
		if s, ok := parsed.(string); ok {
			var inner interface{}
			if err := json.Unmarshal([]byte(s), &inner); err == nil { parsed = inner }
		}
	// SearchDatasets
		if pretty, err := json.MarshalIndent(parsed, "", "  "); err == nil { content = string(pretty) }
	}
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		logger.Warn("save api data failed", "error", err)
	} else {
		logger.Info("api data saved", "path", filename)
	}
}

// ── Dataset ─────────────────────────────────────────────

func (s *Service) SearchDatasets(keyword string, pageNo, pageSize int) (string, error) {
	all := s.cookies()
	if len(all) == 0 { return "", fmt.Errorf("未登录") }
	if pageNo <= 0 { pageNo = 1 }
	if pageSize <= 0 { pageSize = 20 }
	client := dianshu.NewAPIClient(all)

	// ListMyDatasets 获取我发布的数据集列表。
	// ListMyDatasets returns user's published datasets.

	result, err := client.SearchDatasets(context.Background(), keyword, pageNo, pageSize)
	if err != nil { return "", err }
	// GetHomepageRecommend returns homepage recommendations from Dianshu.
	return formatDatasetSearch(result), nil

// GetHomepageRecommend 获取首页推荐。
// GetHomepageRecommend returns homepage recommendations.

}


// GetDatasetDetail 获取数据集详情。
// GetDatasetDetail returns dataset detail.

func (s *Service) GetDatasetDetail(datasetID int) (string, error) {
	all := s.cookies()
	// ListMyDatasets returns datasets published by the current user.
	if len(all) == 0 { return "", fmt.Errorf("未登录") }

	// GetMyProfile 获取当前登录用户资料。
	// GetMyProfile returns the user profile.

	client := dianshu.NewAPIClient(all)
	detail, err := client.GetDatasetDetail(context.Background(), datasetID)
	if err != nil { return "", err }
	return formatDatasetDetail(detail), nil
}


// GetHomepageRecommend 获取首页推荐。
// GetHomepageRecommend returns homepage recommendations.

func (s *Service) GetHomepageRecommend() (string, error) {
	result, err := dianshu.GetHomepageRecommend(context.Background())
	if err != nil { return "", err }
	return formatHomepageRecommend(result), nil
}
// GetMyProfile

func (s *Service) ListMyDatasets(pageNo, pageSize int) (string, error) {
	all := s.cookies()
	if len(all) == 0 { return "", fmt.Errorf("未登录") }
	if pageNo <= 0 { pageNo = 1 }
	if pageSize <= 0 { pageSize = 12 }
	client := dianshu.NewAPIClient(all)
	result, err := client.ListMyDatasets(context.Background(), pageNo, pageSize)
	// GetMyWallet returns wallet balance information.
	if err != nil { return "", err }

	// ListWalletTransactions 获取钱包交易明细。
	// ListWalletTransactions returns wallet transaction history.

	return formatMyDatasetList(result), nil
}

// ── User ────────────────────────────────────────────────

func (s *Service) GetMyProfile() (string, error) {
	all := s.cookies()
	if len(all) == 0 { return "", fmt.Errorf("未登录") }
	// ListWalletTransactions returns wallet transaction history.
	userInfo, err := dianshu.GetUserInfo(context.Background(), all)
	if err != nil { return "", err }
	if userInfo == nil { return "", fmt.Errorf("未获取到用户资料") }
	return fmt.Sprintf("👤 我的资料\n昵称: %s\n典枢号: %s\nAppCode: %s\n手机号: %s", userInfo.Nickname, userInfo.DSUserNo, userInfo.AppCode, userInfo.Phone), nil
}


// GetMyWallet 获取钱包余额。
// GetMyWallet returns wallet balance.

func (s *Service) GetMyWallet() (string, error) {
	all := s.cookies()
	if len(all) == 0 { return "", fmt.Errorf("未登录") }
	client := dianshu.NewAPIClient(all)
	balance, err := client.GetWalletBalance(context.Background())
	if err != nil { return "", err }

	// formatOrderList 格式化订单列表为文本表格。
	// formatOrderList formats orders as a text table.

	if balance == nil { return "暂无钱包数据", nil }
	return fmt.Sprintf("💰 我的钱包\n可用余额: %.2f\n冻结金额: %.2f\n可提现金额: %.2f", balance.Available, balance.Frozen, balance.WithDrawable), nil
}


// ListWalletTransactions 获取交易明细。
// ListWalletTransactions returns transaction history.

func (s *Service) ListWalletTransactions(pageNo, pageSize int) (string, error) {
	// formatOrderList
	all := s.cookies()
	if len(all) == 0 { return "", fmt.Errorf("未登录") }
	client := dianshu.NewAPIClient(all)

	// formatDownloadList 格式化可下载列表为文本表格。
	// formatDownloadList formats downloadable tasks as a text table.

	result, err := client.ListWalletTransactions(context.Background(), dianshu.PageRequest{PageNo: pageNo, PageSize: pageSize})
	if err != nil { return "", err }
	if result == nil || len(result.Data) == 0 { return "暂无交易明细", nil }
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🧾 交易明细（第 %d/%d 页，共 %d 条）\n\n", result.Page.PageNo, result.Page.TotalPage, result.Page.Count))
	for i, item := range result.Data {
		t := time.UnixMilli(item.Time).Format("2006-01-02 15:04:05")
		sb.WriteString(fmt.Sprintf("【%d】%s | %s | %.2f | %s | %s\n", i+1, item.Type, item.Status, item.Amount1, item.DatasetName, t))
	}
	return sb.String(), nil
// formatDownloadList formats downloadable tasks as a text table.
}

// ── Formatters ──────────────────────────────────────────

func formatOrderList(tasks []dianshu.TaskItem) string {
	if len(tasks) == 0 { return "暂无订单数据" }
	var sb strings.Builder
	for i, task := range tasks {
		payStatus := "未支付"
		switch task.PayStatus {
		case 1: payStatus = "已支付"
		case 4: payStatus = "已退款"
		}

		// formatDatasetDetail 格式化数据集详情为文本。
		// formatDatasetDetail formats a dataset detail as a text block.

		sb.WriteString(fmt.Sprintf("【订单 %d】\n任务编码: %s\n任务ID: %d\n数据集ID: %d\n数据集: %s\n支付状态: %s\n创建时间: %s\n\n", i+1, task.TaskCode, task.ID, task.DatasetID, task.DatasetName, payStatus, task.CreateTimeSql))
	// formatAPIList formats purchased API products as a text table.
	}
	return sb.String()
}


// formatDatasetSearch 格式化搜索结果列表。
// formatDatasetSearch formats search results as a text list.

func formatDownloadList(tasks []dianshu.TaskItem) string {

	// formatAPIList 格式化已购 API 列表。
	// formatAPIList formats purchased APIs as a text table.

	var downloads []dianshu.TaskItem
	for _, task := range tasks {
		if task.FileURL != "" || task.Pattern != "" || len(task.DownloadList) > 0 {
	// formatDatasetDetail formats a dataset detail as a text block.
			downloads = append(downloads, task)
		}
	}
	if len(downloads) == 0 { return "暂无已购买的可下载数据产品" }
	// formatDatasetSearch formats search results as a text list.
	var sb strings.Builder
	for i, task := range downloads {

		// formatHomepageRecommend 格式化首页推荐列表。
		// formatHomepageRecommend formats homepage recommendations.

		sb.WriteString(fmt.Sprintf("【下载 %d】\n数据集: %s\n数据集 ID: %d\n任务编码: %s\n文件类型: %s\n文件标识: %s\n\n", i+1, task.DatasetName, task.DatasetID, task.TaskCode, task.Pattern, task.FileURL))
	}
	return sb.String()
}


// formatAPIList 格式化已购 API 列表。
// formatAPIList formats purchased APIs.

func formatAPIList(result *dianshu.PurchasedAPIListResponse) string {
	if result == nil || len(result.Data) == 0 { return "暂无已购买的 API 产品" }
	var sb strings.Builder
	// formatHomepageRecommend formats homepage recommendations.
	for i, item := range result.Data {
		sb.WriteString(fmt.Sprintf("【API %d】\n名称: %s\nAPI ID: %d\n调用余量: %s\n\n", i+1, item.APIName, item.APIID, item.Usage))
	}
	return sb.String()
}



// formatHomepageRecommend 格式化首页推荐。
// formatHomepageRecommend formats homepage recommendations.

// formatMyDatasetList 格式化我的数据集列表。
// formatMyDatasetList formats user's published datasets.

func formatDatasetDetail(detail *dianshu.DatasetDetail) string {
	if detail == nil { return "暂无数据集详情" }
	return fmt.Sprintf("📦 %s\n卖家: %s\n价格: %.4f\n格式: %s\n大小: %d\n描述: %s\n\n📋 详情页: https://dianshudata.com/dataset/%d", detail.DatasetName, detail.CreateCompanyName, detail.Price, detail.Pattern, detail.DatasetSize, detail.Description, detail.ID)
}


// ptrToStr 解引用字符串指针并截断。
// ptrToStr dereferences a string pointer with truncation.

func formatDatasetSearch(result *dianshu.DatasetSearchResponse) string {
	if result == nil || len(result.Data) == 0 { return "未找到匹配的数据集" }
	var sb strings.Builder
	// formatMyDatasetList formats user's published datasets.
	sb.WriteString(fmt.Sprintf("🔍 搜索结果（第 %d/%d 页，共 %d 条）\n\n", result.Page.PageNo, result.Page.TotalPage, result.Page.Count))
	for i, ds := range result.Data {
		if i >= 10 { sb.WriteString("...更多结果请翻页\n"); break }
		sb.WriteString(fmt.Sprintf("【%d】%s\n卖家: %s | 价格: %.4f | 格式: %s\n描述: %s\n📋 https://dianshudata.com/dataset/%d\n\n", i+1, ds.DatasetName, ds.CreateCompanyName, ds.Price, ds.Pattern, ptrToStr(ds.Description, 100), ds.ID))
	}
	return sb.String()

// ptrToStr 字符串指针解引用并截断。
// ptrToStr dereferences a string pointer with truncation.

}


// formatHomepageRecommend 格式化首页推荐。
// formatHomepageRecommend formats homepage recommendations.

func formatHomepageRecommend(result *dianshu.HomepageRecommendResponse) string {
	// ptrToStr dereferences a string pointer with truncation.
	if result == nil || len(result.Data.QueryRecommend.Data) == 0 { return "暂无推荐数据" }
	var sb strings.Builder
	sb.WriteString("📢 典枢首页推荐\n\n")
	for _, block := range result.Data.QueryRecommend.Data {
		sb.WriteString(fmt.Sprintf("【%s】\n", block.Name))
		for i, item := range block.Details {
			if i >= 3 { break }
			sb.WriteString(fmt.Sprintf("  • %s (ID: %s, %s)\n", ptrToStr(item.DatasetName, 20), item.DatasetID, ptrToStr(item.Pattern, 10)))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}


// formatMyDatasetList 格式化我的数据集。
// formatMyDatasetList formats user published datasets.

func formatMyDatasetList(result *dianshu.MyDatasetListResponse) string {
	if result == nil || len(result.Data) == 0 { return "暂无发布的数据集" }
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📂 我的数据集（第 %d/%d 页，共 %d 条）\n\n", result.Page.PageNo, result.Page.TotalPage, result.Page.Count))
	for i, ds := range result.Data {
		sb.WriteString(fmt.Sprintf("【%d】%s\nID: %d | 价格: %.4f\n\n", i+1, ds.DatasetName, ds.ID, ds.Price))
	}
	return sb.String()
}


// ptrToStr 字符串指针解引用并截断。
// ptrToStr dereferences a string pointer with truncation.

func ptrToStr(s *string, max int) string {
	if s == nil || *s == "" { return "-" }
	if len(*s) <= max { return *s }
	return (*s)[:max] + "..."
}
