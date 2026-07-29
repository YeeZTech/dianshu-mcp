package kms

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const baseURL = "https://kms.dianshudata.com"

// AuthResponse JWT 登录响应
type AuthResponse struct {
	Auth *Auth `json:"auth"`
}

// Auth 认证信息
type Auth struct {
	ClientToken string `json:"client_token"`
}

// SecretResponse 获取 secret 响应
type SecretResponse struct {
	Data *SecretData `json:"data"`
}

// SecretData secret 数据层
type SecretData struct {
	Data     map[string]string `json:"data"`
	Metadata any               `json:"metadata"`
}

// Login 使用典枢 JWT 登录 KMS，返回 client_token。
func Login(jwt string) (string, error) {
	body := strings.NewReader(fmt.Sprintf(`{"role":"dianshu-user","jwt":"%s"}`, jwt))
	resp, err := http.Post(baseURL+"/v1/auth/jwt/login", "application/json", body)
	if err != nil {
		return "", fmt.Errorf("KMS 登录请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("KMS 登录返回 %d: %s", resp.StatusCode, string(respBytes))
	}

	var authResp AuthResponse
	if err := json.Unmarshal(respBytes, &authResp); err != nil {
		return "", fmt.Errorf("解析 KMS 登录响应失败: %w", err)
	}
	if authResp.Auth == nil || authResp.Auth.ClientToken == "" {
		return "", fmt.Errorf("KMS 登录响应缺少 client_token")
	}
	return authResp.Auth.ClientToken, nil
}

// GetAuthPassword 获取用户在 KMS 中存储的授权密码。
// userNo: 典枢用户编号
// clientToken: KMS 登录后获得的 token
func GetAuthPassword(userNo, clientToken string) (string, error) {
	url := fmt.Sprintf("%s/v1/secret/data/dianshu/users/%s/auth-password", baseURL, userNo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("创建 KMS 请求失败: %w", err)
	}
	req.Header.Set("X-Vault-Token", clientToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("KMS 获取授权密码请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("KMS 获取授权密码返回 %d: %s", resp.StatusCode, string(respBytes))
	}

	var secretResp SecretResponse
	if err := json.Unmarshal(respBytes, &secretResp); err != nil {
		return "", fmt.Errorf("解析 KMS secret 响应失败: %w", err)
	}
	if secretResp.Data == nil || secretResp.Data.Data == nil {
		return "", fmt.Errorf("KMS secret 响应缺少 data.data 字段")
	}
	password, ok := secretResp.Data.Data["password"]
	if !ok || password == "" {
		return "", fmt.Errorf("KMS secret 响应缺少 password 字段")
	}
	return password, nil
}
