package main

import (
	"context"
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

// DataSearch 执行数据查询，并预留通用扣费入口。
func (s *DianshuService) DataSearch(ctx context.Context, request DataSearchArgs) (*MCPToolResult, error) {
	dataQueryRequest, err := s.buildDataQueryRequest(request)
	if err != nil {
		return nil, err
	}

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

	resultText, err := s.persistRawResult(dataQueryRequest, queryResult)
	if err != nil {
		return nil, err
	}
	return &MCPToolResult{Content: []MCPContent{{Type: "text", Text: resultText}}}, nil
}

func (s *DianshuService) buildDataQueryRequest(args DataSearchArgs) (dianshu.DataQueryRequest, error) {
	trimmedQuery := strings.TrimSpace(args.Query)
	if trimmedQuery == "" {
		return dianshu.DataQueryRequest{}, fmt.Errorf("查询内容不能为空")
	}

	providerType := strings.TrimSpace(args.Provider)
	if providerType == "" {
		providerType = dianshu.ProviderTypeXiaohongshu
	}
	datasetType := strings.TrimSpace(args.Dataset)
	if datasetType == "" {
		datasetType = dianshu.DatasetTypeSearch
	}
	siteDomain := strings.TrimSpace(args.SiteDomain)
	if siteDomain == "" {
		siteDomain = dianshu.XiaohongshuSiteDomain
	}

	keyword := strings.TrimSpace(args.Keyword)
	if keyword == "" {
		keyword = trimmedQuery
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
		ProviderType: providerType,
		DatasetType:  datasetType,
		SiteDomain:   siteDomain,
		Body: map[string]string{
			"startTime": startTime,
			"endTime":   endTime,
			"keyword":   keyword,
			"page":      page,
		},
		RawQuery: trimmedQuery,
	}, nil
}

func (s *DianshuService) persistRawResult(request dianshu.DataQueryRequest, result *dianshu.DataQueryResult) (string, error) {
	if len(result.RawJSON) == 0 {
		return "", fmt.Errorf("查询结果为空，无法写入 JSON 文件")
	}

	outputDir := filepath.Join("output", "data-search")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", fmt.Errorf("创建输出目录失败: %w", err)
	}

	fileName := fmt.Sprintf("%s_%s_%s.json", request.ProviderType, request.DatasetType, time.Now().Format("20060102_150405"))
	filePath := filepath.Join(outputDir, fileName)
	if err := os.WriteFile(filePath, result.RawJSON, 0o644); err != nil {
		return "", fmt.Errorf("写入查询结果文件失败: %w", err)
	}

	requestBodyJSON, err := json.MarshalIndent(request.Body, "", "  ")
	if err != nil {
		return "", fmt.Errorf("序列化查询参数失败: %w", err)
	}

	resultText := fmt.Sprintf("✅ 查询成功\n查询内容: %s\n\n数据源: %s/%s\n站点: %s\nDSSeqNo: %s\n结果状态: %d / %s\n查询参数:\n%s\n\n结果文件: %s\nMEDIA:%s", request.RawQuery, request.ProviderType, request.DatasetType, request.SiteDomain, result.DSSeqNo, result.ResultCode, result.ResultDesc, string(requestBodyJSON), filePath, filePath)
	return resultText, nil
}
