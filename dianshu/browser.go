// Package dianshu - see README for details.
//
// Author: zhyyao
package dianshu

import (
	"context"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/sirupsen/logrus"
)

// NewBrowser 创建一个新的浏览器实例
func NewBrowser(ctx context.Context, headless bool) (*rod.Browser, error) {
	l := launcher.New().Headless(headless).NoSandbox(true)

	// 使用自定义浏览器路径
	if binPath, ok := launcher.LookPath(); ok && binPath != "" {
		l = l.Bin(binPath)
	}

	controlURL, err := l.Launch()
	if err != nil {
		return nil, fmt.Errorf("启动浏览器失败: %w", err)
	}

	browser := rod.New().ControlURL(controlURL)
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("连接浏览器失败: %w", err)
	}

	return browser, nil
}

// NewPage 创建新页面
func NewPage(browser *rod.Browser) (*rod.Page, error) {
	page, err := browser.Page(proto.TargetCreateTarget{})
	if err != nil {
		return nil, fmt.Errorf("创建页面失败: %w", err)
	}
	return page, nil
}

// PollForToken 轮询等待 token 登录成功
func PollForToken(ctx context.Context, page *rod.Page, timeout time.Duration) (string, error) {
	startTime := time.Now()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-ticker.C:
			if time.Since(startTime) > timeout {
				return "", fmt.Errorf("登录超时（%v）", timeout)
			}

			// 检查 URL
			urlObj, err := page.Evaluate(&rod.EvalOptions{JS: "() => window.location.href"})
			if err != nil {
				continue
			}
			urlStr := urlObj.Value.String()

			// 如果还在登录页面或 Casdoor 页面，继续等待
			if contains(urlStr, "/login") || contains(urlStr, "sso.dianshudata.com") {
				continue
			}

			// 如果回到了 dianshudata.com，尝试获取 token
			if contains(urlStr, "dianshudata.com") {
				time.Sleep(2 * time.Second) // 等待页面加载

				// 尝试从 cookie 获取 token
				cookies, err := page.Cookies(nil)
				if err != nil {
					continue
				}

				for _, c := range cookies {
					if c.Name == "token" && c.Value != "" {
						logrus.Infof("检测到 token，登录成功")
						return c.Value, nil
					}
				}

				// 也可能 token 在 localStorage
				tokenObj, err := page.Evaluate(&rod.EvalOptions{JS: "() => localStorage.getItem('token')"})
				if err == nil && tokenObj != nil && tokenObj.Value.String() != "" {
					token := tokenObj.Value.String()
					logrus.Infof("从 localStorage 获取到 token")
					return token, nil
				}
			}
		}
	}
}

// FindElementByText 根据文本查找元素
func FindElementByText(page *rod.Page, text string) (*rod.Element, error) {
	return page.ElementX(fmt.Sprintf("//*[contains(text(), '%s')]", text))
}

// contains checks if a byte slice contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}
// containsStr checks if a string contains a substring.

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
