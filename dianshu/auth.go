// Package dianshu - see README for details.
//
// Author: zhyyao

package dianshu

import (
	"context"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/sirupsen/logrus"
)

// 微信开放平台 QR 码登录 URL（典枢配置）
const WeChatQRLoginURL = "https://open.weixin.qq.com/connect/qrconnect?appid=wxb99c7869527dc265&redirect_uri=https://sso.dianshudata.com/callback&scope=snsapi_login&response_type=code&state=P2NsaWVudF9pZD1mY2RmZWI2NTMxYjEzMTUxODUxZiZyZXNwb25zZV90eXBlPWNvZGUmcmVkaXJlY3RfdXJpPWh0dHBzJTNBJTJGJTJGYWNjb3VudC5kaWFuc2h1ZGF0YS5jb20lMkZjYWxsYmFjayUyRiZzY29wZT1yZWFkJnN0YXRlPWh0dHBzJTNBJTJGJTJGZGlhbnNodWRhdGEuY29tJTJGY2FsbGJhY2smYXBwbGljYXRpb249ZGlhbnNodSZwcm92aWRlcj1wcm92aWRlcl9sb2dpbl93ZWNoYXQmbWV0aG9kPXNpZ251cA==#wechat_redirect"

// LoginCheckResult 登录检查结果
type LoginCheckResult struct {
	IsLogin  bool      `json:"isLogin"`
	Nickname string    `json:"nickname,omitempty"`
	UserID   string    `json:"userId,omitempty"`
	UserInfo *UserInfo `json:"userInfo,omitempty"`
	Error    string    `json:"error,omitempty"`
}

// GetLoginQRCode 获取微信登录二维码
func GetLoginQRCode(ctx context.Context, headless bool) ([]byte, string, error) {
	browser, err := NewBrowser(ctx, headless)
	if err != nil {
		return nil, "", fmt.Errorf("启动浏览器失败: %w", err)
	}
	defer browser.Close()

	page, err := NewPage(browser)
	if err != nil {
		return nil, "", fmt.Errorf("创建页面失败: %w", err)
	}
	defer page.Close()

	logrus.Info("正在打开微信登录二维码页面...")
	if err := page.Navigate(WeChatQRLoginURL); err != nil {
		return nil, "", fmt.Errorf("打开微信二维码页失败: %w", err)
	}
	time.Sleep(5 * time.Second)

	screenshot, err := page.Screenshot(true, nil)
	if err != nil {
		return nil, "", fmt.Errorf("截图失败: %w", err)
	}

	text := "请使用微信扫描二维码登录典枢平台\n二维码有效期约 5 分钟"
	return screenshot, text, nil
}

// WaitForWeChatLogin 等待微信扫码登录完成
// 返回所有 cookies（name→value map）
func WaitForWeChatLogin(ctx context.Context, headless bool, timeout time.Duration) (map[string]string, error) {
	browser, err := NewBrowser(ctx, headless)
	if err != nil {
		return nil, fmt.Errorf("启动浏览器失败: %w", err)
	}
	defer browser.Close()

	page, err := NewPage(browser)
	if err != nil {
		return nil, fmt.Errorf("创建页面失败: %w", err)
	}
	defer page.Close()

	logrus.Info("正在打开微信登录二维码页面...")
	if err := page.Navigate(WeChatQRLoginURL); err != nil {
		return nil, fmt.Errorf("打开微信登录页失败: %w", err)
	}
	time.Sleep(5 * time.Second)

	logrus.Info("等待用户扫码登录（请用微信扫描二维码）...")

	startTime := time.Now()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			if time.Since(startTime) > timeout {
				return nil, fmt.Errorf("登录超时（%v），请重新获取二维码", timeout)
			}

			urlObj, err := page.Evaluate(&rod.EvalOptions{JS: "() => window.location.href"})
			if err != nil {
				continue
			}
			urlStr := urlObj.Value.String()

			if contains(urlStr, "open.weixin.qq.com") || contains(urlStr, "sso.dianshudata.com/callback") {
				continue
			}

			if contains(urlStr, "dianshudata.com") {
				time.Sleep(3 * time.Second)

				cookieMap := make(map[string]string)

				// 先从 localStorage 拿 token
				tokenObj, err := page.Evaluate(&rod.EvalOptions{JS: "() => localStorage.getItem('token')"})
				if err == nil && tokenObj != nil && tokenObj.Value.String() != "" {
					cookieMap["token"] = tokenObj.Value.String()
					logrus.Info("从 localStorage 获取到 token")
				}

				// 再从 page cookies 拿所有 cookies
				cookies, err := page.Cookies(nil)
				if err == nil {
					for _, c := range cookies {
						cookieMap[c.Name] = c.Value
					}
					logrus.Infof("从 page cookies 获取到 %d 个 cookies", len(cookies))
				}

				// 只要有 token 就算登录成功
				if token, ok := cookieMap["token"]; ok && token != "" {
					return cookieMap, nil
				}

				continue
			}
		}
	}
}

// CheckLoginStatus 检查登录状态
func CheckLoginStatus(ctx context.Context, cookies map[string]string) (*LoginCheckResult, error) {
	if len(cookies) == 0 {
		return &LoginCheckResult{IsLogin: false}, nil
	}

	userInfo, err := GetUserInfo(ctx, cookies)
	if err != nil {
		return &LoginCheckResult{
			IsLogin: false,
			Error:   err.Error(),
		}, nil
	}

	return &LoginCheckResult{
		IsLogin:  true,
		Nickname: userInfo.Nickname,
		UserID:   userInfo.UserID,
		UserInfo: userInfo,
	}, nil
}

// GetLoginQRCodeOnly 仅获取二维码截图，不等待扫码。
func GetLoginQRCodeOnly(ctx context.Context, headless bool) ([]byte, string, error) {
	browser, err := NewBrowser(ctx, headless)
	if err != nil {
		return nil, "", fmt.Errorf("启动浏览器失败: %w", err)
	}
	defer browser.Close()

	page, err := NewPage(browser)
	if err != nil {
		return nil, "", fmt.Errorf("创建页面失败: %w", err)
	}
	defer page.Close()

	if err := page.Navigate(WeChatQRLoginURL); err != nil {
		return nil, "", fmt.Errorf("打开微信二维码页失败: %w", err)
	}
	time.Sleep(5 * time.Second)

	screenshot, err := page.Screenshot(true, nil)
	if err != nil {
		return nil, "", fmt.Errorf("截图失败: %w", err)
	}

	return screenshot, "请使用微信扫描二维码登录典枢平台\n二维码有效期约 5 分钟", nil
}
