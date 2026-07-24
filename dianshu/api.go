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
	baseAPIURL        = "https://api.dianshudata.com"
	dataAPIDetailURL  = "https://data-api.dianshudata.com"
	userAgent         = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"
	apiRequestTimeout = 30 * time.Second
)

// GetDefaultCookiePath 获取默认 cookies 文件路径
func GetDefaultCookiePath() string {
	return "cookies.json"
}

// APIClient API 客户端
type APIClient struct {
	httpClient *http.Client
	cookies    map[string]string
}

// NewAPIClient 创建 API 客户端
func NewAPIClient(cookies map[string]string) *APIClient {
	return &APIClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		cookies:    cookies,
	}
}

// GetUserInfo 获取用户信息（使用 cookies map）
func GetUserInfo(ctx context.Context, cookies map[string]string) (*UserInfo, error) {
	client := NewAPIClient(cookies)
	return client.GetUserInfo(ctx)
}

func (c *APIClient) GetUserInfo(ctx context.Context) (*UserInfo, error) {
	resp, err := c.doRequest(ctx, "POST", "/login/getUserInfo", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		ResultCode int       `json:"resultCode"`
		ResultDesc string    `json:"resultDesc"`
		Data       *UserInfo `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析用户信息失败: %w", err)
	}
	if result.ResultCode != 100 {
		return nil, fmt.Errorf("获取用户信息失败: %s", result.ResultDesc)
	}
	return result.Data, nil
}

// GetWalletBalance 获取我的钱包余额
func (c *APIClient) GetWalletBalance(ctx context.Context) (*WalletBalance, error) {
	resp, err := c.doRequest(ctx, "POST", "/system/wallet/balance", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		ResultCode int            `json:"resultCode"`
		ResultDesc string         `json:"resultDesc"`
		Data       *WalletBalance `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析钱包余额失败: %w", err)
	}
	if result.ResultCode != 100 {
		return nil, fmt.Errorf("获取钱包余额失败: %s", result.ResultDesc)
	}
	return result.Data, nil
}

// ListWalletTransactions 获取钱包交易明细
func (c *APIClient) ListWalletTransactions(ctx context.Context, page PageRequest) (*WalletTransactionListResponse, error) {
	resp, err := c.doRequest(ctx, "POST", "/system/wallet/order_list", page)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result WalletTransactionListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析钱包交易明细失败: %w", err)
	}
	if result.ResultCode != 100 {
		return nil, fmt.Errorf("获取钱包交易明细失败: %s", result.ResultDesc)
	}
	return &result, nil
}

// QueryOrders 查询订单统计
func (c *APIClient) QueryOrders(ctx context.Context, orderType int, orderCode string) (*OrderQueryData, error) {
	reqBody := map[string]interface{}{"orderType": orderType, "orderCode": orderCode}
	resp, err := c.doRequest(ctx, "POST", "/unused/order/query", reqBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	logrus.Debugf("查询订单响应: %s", string(body))

	var result struct {
		ResultCode int             `json:"resultCode"`
		ResultDesc string          `json:"resultDesc"`
		Data       *OrderQueryData `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析订单数据失败: %w", err)
	}
	if result.ResultCode != 100 {
		return nil, fmt.Errorf("查询订单失败: %s", result.ResultDesc)
	}
	return result.Data, nil
}

// ListTasks 获取任务/订单列表
func (c *APIClient) ListTasks(ctx context.Context, pageNo, pageSize int) ([]TaskItem, error) {
	resp, err := c.doRequest(ctx, "POST", "/system/task/taskList", PageRequest{PageNo: pageNo, PageSize: pageSize})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result TaskListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析任务列表失败: %w", err)
	}
	if result.ResultCode != 100 {
		return nil, fmt.Errorf("查询任务列表失败: %s", result.ResultDesc)
	}
	return result.Data, nil
}

// GetTaskDetail 获取单个任务详情
func (c *APIClient) GetTaskDetail(ctx context.Context, id int) (*TaskItem, error) {
	resp, err := c.doRequest(ctx, "POST", "/system/task/taskDetail", map[string]interface{}{"id": id})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		ResultCode int       `json:"resultCode"`
		ResultDesc string    `json:"resultDesc"`
		Data       *TaskItem `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析任务详情失败: %w", err)
	}
	if result.ResultCode != 100 {
		return nil, fmt.Errorf("查询任务详情失败: %s", result.ResultDesc)
	}
	return result.Data, nil
}

const dataAPIGateway = "https://data-api.dianshudata.com"

// GetAPIDetail 获取 API 产品详情（使用 data-api 子网关）
func (c *APIClient) GetAPIDetail(ctx context.Context, apiID int) (*APIDetail, error) {
	resp, err := c.doRequestToGateway(ctx, "POST", dataAPIGateway, "/api/detail", map[string]interface{}{"apiId": apiID, "deleted": 0})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result APIDetailResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析 API 详情失败: %w", err)
	}
	if result.ResultCode != 100 {
		return nil, fmt.Errorf("查询 API 详情失败: %s", result.ResultDesc)
	}
	return result.Data, nil
}

// GetHomepageRecommend 获取典枢首页推荐数据
func GetHomepageRecommend(ctx context.Context) (*HomepageRecommendResponse, error) {
	reqBody := map[string]interface{}{
		"query": `query QueryRecommend($typeList: [Int!]!, $limit: Int) { queryRecommend(dto: { typeList: $typeList, limit: $limit }) { resultCode resultDesc data { id type name details { id name description imageBgUrl imageFrontUrl hrefUrl datasetId datasetName price pattern securityLevel } } } }`,
		"variables": map[string]interface{}{
			"typeList": []int{1, 2, 3, 4},
			"limit":    9,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %w", err)
	}

	url := baseAPIURL + "/graphql"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Origin", "https://dianshudata.com")
	req.Header.Set("Referer", "https://dianshudata.com/")
	req.Header.Set("Accept", "application/json, text/plain, */*")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result HomepageRecommendResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析首页推荐响应失败: %w", err)
	}
	return &result, nil
}

// SearchDatasets 典枢平台数据集搜索
func (c *APIClient) SearchDatasets(ctx context.Context, keyword string, pageNo, pageSize int) (*DatasetSearchResponse, error) {
	reqBody := map[string]interface{}{
		"name":     keyword,
		"order":    "",
		"orderBy":  "",
		"pageNo":   pageNo,
		"pageSize": pageSize,
	}
	resp, err := c.doRequest(ctx, "POST", "/dataset/datasetListRight", reqBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result DatasetSearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析数据集搜索响应失败: %w", err)
	}
	if result.ResultCode != 100 {
		return nil, fmt.Errorf("搜索数据集失败: %s", result.ResultDesc)
	}
	return &result, nil
}

func (c *APIClient) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	url := baseAPIURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Origin", "https://dianshudata.com")
	req.Header.Set("Referer", "https://dianshudata.com/")
	req.Header.Set("Accept", "application/json, text/plain, */*")

	if len(c.cookies) > 0 {
		if token, ok := c.cookies["token"]; ok && token != "" {
			req.Header.Set("token", token)
		}
		var cookieParts []string
		for name, value := range c.cookies {
			if name != "token" {
				cookieParts = append(cookieParts, name+"="+value)
			}
		}
		if len(cookieParts) > 0 {
			req.Header.Set("Cookie", strings.Join(cookieParts, "; "))
		}
	}

	logrus.Debugf("请求 %s %s", method, url)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}
	return resp, nil
}

// doRequestToGateway 向指定网关发送请求（支持多网关）
func (c *APIClient) doRequestToGateway(ctx context.Context, method, gateway, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	url := gateway + path
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Origin", "https://dianshudata.com")
	req.Header.Set("Referer", "https://dianshudata.com/")
	req.Header.Set("Accept", "application/json, text/plain, */*")

	if len(c.cookies) > 0 {
		if token, ok := c.cookies["token"]; ok && token != "" {
			req.Header.Set("token", token)
		}
		var cookieParts []string
		for name, value := range c.cookies {
			if name != "token" {
				cookieParts = append(cookieParts, name+"="+value)
			}
		}
		if len(cookieParts) > 0 {
			req.Header.Set("Cookie", strings.Join(cookieParts, "; "))
		}
	}

	logrus.Debugf("请求 %s %s", method, url)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}
	return resp, nil
}
