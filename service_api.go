package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dianshu-mcp/cookies"
	"dianshu-mcp/dianshu"

	"github.com/sirupsen/logrus"
)

func (s *DianshuService) CallAPI(ctx context.Context, apiID int, params map[string]string, method string) (*MCPToolResult, error) {
	allCookies := cookies.GetAllCookies()
	if len(allCookies) == 0 {
		return nil, fmt.Errorf("未登录，请先登录典枢")
	}

	cookieList := make([]string, 0, len(allCookies))
	for k, v := range allCookies {
		cookieList = append(cookieList, k+"="+v)
	}

	detail, err := dianshu.GetAPIDetail(ctx, &http.Client{}, apiID, "", cookieList)
	if err != nil {
		return nil, fmt.Errorf("查询 API 详情失败: %w", err)
	}
	if detail == nil {
		return nil, fmt.Errorf("API %d 不存在或无权限", apiID)
	}

	userInfo, err := dianshu.NewAPIClient(allCookies).GetUserInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	uniqueAPIID := detail.UniqueAPIID
	apiMethod := method
	if apiMethod == "" {
		if detail.RequestMethod == 1 {
			apiMethod = "GET"
		} else {
			apiMethod = "POST"
		}
	}

	client, err := dianshu.NewClient(userInfo.AppCode, uniqueAPIID)
	if err != nil {
		return nil, fmt.Errorf("初始化 API 客户端失败: %w", err)
	}

	var result string
	if apiMethod == "GET" {
		result, err = client.Get(params)
	} else {
		result, err = client.Post(params)
	}
	if err != nil {
		return nil, fmt.Errorf("调用 API 失败: %w", err)
	}

	// 保存结果到 output/api-data/
	saveAPIData(detail.APIName, result)

	return &MCPToolResult{
		Content: []MCPContent{{Type: "text", Text: result}},
	}, nil
}

func saveAPIData(apiName, data string) {
	dir := filepath.Join("output", "api-data")
	os.MkdirAll(dir, 0755)

	ts := time.Now().Format("20060102_150405")
	safeName := strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r == ':' || r == '*' || r == '?' || r == '"' || r == '<' || r == '>' || r == '|' {
			return '_'
		}
		return r
	}, apiName)

	filename := filepath.Join(dir, fmt.Sprintf("%s_%s.json", safeName, ts))

	// 尝试解析为 JSON 并 pretty-print
	content := data
	var parsed interface{}
	if err := json.Unmarshal([]byte(data), &parsed); err == nil {
		// 如果是嵌套的 JSON 字符串，再解一层
		if s, ok := parsed.(string); ok {
			var inner interface{}
			if err := json.Unmarshal([]byte(s), &inner); err == nil {
				parsed = inner
			}
		}
		if pretty, err := json.MarshalIndent(parsed, "", "  "); err == nil {
			content = string(pretty)
		}
	}

	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		logrus.Warnf("保存 API 数据失败: %v", err)
	} else {
		logrus.Infof("API 数据已保存: %s", filename)
	}
}
