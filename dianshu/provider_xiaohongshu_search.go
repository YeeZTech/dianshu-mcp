package dianshu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	xiaohongshuSearchAPIBaseURL       = "..."
	xiaohongshuSearchPath             = "/api/istarshine/consult/search"
	xiaohongshuSearchDomainHeaderName = "X-Site-Domain"
	xiaohongshuSearchAuthHeaderName   = "Authorization"
	xiaohongshuSearchAuthorization    = "..."
	ProviderTypeXiaohongshu           = "xiaohongshu"
	DatasetTypeSearch                 = "search"
	XiaohongshuSiteDomain             = "xiaohongshu.com"
	XiaohongshuDefaultPage            = "1"
)

// XiaohongshuSearchProvider 表示小红书搜索数据源实现。
type XiaohongshuSearchProvider struct {
	httpClient *http.Client
	baseURL    string
}

// NewXiaohongshuSearchProvider 创建小红书搜索数据源实现。
func NewXiaohongshuSearchProvider() *XiaohongshuSearchProvider {
	return &XiaohongshuSearchProvider{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    xiaohongshuSearchAPIBaseURL,
	}
}

// Query 调用小红书搜索查询接口。
func (p *XiaohongshuSearchProvider) Query(ctx context.Context, request DataQueryRequest) (*DataQueryResult, error) {
	siteDomain := strings.TrimSpace(request.SiteDomain)
	if siteDomain == "" {
		return nil, fmt.Errorf("siteDomain 不能为空")
	}
	accessToken := strings.TrimSpace(request.AccessToken)
	if accessToken == "" {
		accessToken = xiaohongshuSearchAuthorization
	}
	if len(request.Body) == 0 {
		return nil, fmt.Errorf("查询请求体不能为空")
	}

	requestBody, err := json.Marshal(request.Body)
	if err != nil {
		return nil, fmt.Errorf("序列化查询请求失败: %w", err)
	}

	requestURL := p.baseURL + xiaohongshuSearchPath
	headers := map[string]string{
		"Content-Type":                    "application/json",
		xiaohongshuSearchDomainHeaderName: siteDomain,
		xiaohongshuSearchAuthHeaderName:   accessToken,
	}
	debugHeadersJSON, _ := json.MarshalIndent(headers, "", "  ")
	debugBodyJSON, _ := json.MarshalIndent(request.Body, "", "  ")
	logrus.Infof("[XiaohongshuSearchProvider-Query] request_url=%s", requestURL)
	logrus.Infof("[XiaohongshuSearchProvider-Query] request_headers=%s", string(debugHeadersJSON))
	logrus.Infof("[XiaohongshuSearchProvider-Query] request_body=%s", string(debugBodyJSON))

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("创建查询请求失败: %w", err)
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set(xiaohongshuSearchDomainHeaderName, siteDomain)
	httpRequest.Header.Set(xiaohongshuSearchAuthHeaderName, accessToken)

	httpResponse, err := p.httpClient.Do(httpRequest)
	if err != nil {
		return nil, fmt.Errorf("调用查询接口失败: %w", err)
	}
	defer httpResponse.Body.Close()

	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("读取查询响应失败: %w", err)
	}
	logrus.Infof("[XiaohongshuSearchProvider-Query] response_status=%d", httpResponse.StatusCode)
	logrus.Infof("[XiaohongshuSearchProvider-Query] response_body=%s", string(responseBody))
	if httpResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("查询接口返回异常状态码 %d: %s", httpResponse.StatusCode, string(responseBody))
	}

	var response XiaohongshuSearchResponse
	if err = json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("解析小红书搜索响应失败: %w", err)
	}

	return &DataQueryResult{
		ResultCode: response.ResultCode,
		ResultDesc: response.ResultDesc,
		Data:       response.Data,
		DSSeqNo:    response.DSSeqNo,
		RawJSON:    append(json.RawMessage(nil), responseBody...),
	}, nil
}

func normalizeProviderType(providerType string) string {
	return strings.TrimSpace(strings.ToLower(providerType))
}

func normalizeDatasetType(datasetType string) string {
	return strings.TrimSpace(strings.ToLower(datasetType))
}
