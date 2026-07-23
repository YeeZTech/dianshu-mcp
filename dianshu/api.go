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
	baseAPIURL = "https://api.dianshudata.com"
	userAgent  = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"
)

// GetDefaultCookiePath 获取默认 cookies 文件路径
func GetDefaultCookiePath() string {
	return "cookies.json"
}

// APIClient 典枢平台 API 客户端
type APIClient struct {
	httpClient *http.Client
	cookies    map[string]string
}

// NewAPIClient 创建典枢平台 API 客户端
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
