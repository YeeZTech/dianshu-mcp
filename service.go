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

// DataSearch 执行数据查询，查询结果写入 output/data-search/ 目录。
func (s *DianshuService) DataSearch(ctx context.Context, args DataSearchArgs) (*MCPToolResult, error) {
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

	resultText := fmt.Sprintf("✅ 查询成功\n查询内容: %s\n\n数据源: %s/%s\n站点: %s\nDSSeqNo: %s\n结果状态: %d / %s\n查询参数:\n%s\n\n结果文件: %s\nMEDIA:%s", queryReq.RawQuery, queryReq.ProviderType, queryReq.DatasetType, queryReq.SiteDomain, result.DSSeqNo, result.ResultCode, result.ResultDesc, string(requestBodyJSON), filePath, filePath)
	return resultText, nil
}
