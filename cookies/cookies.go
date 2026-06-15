package cookies

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Cookies 持久化的 cookies 文件管理
type Cookies struct {
	mu       sync.RWMutex
	filePath string
	Data     map[string]interface{} `json:"data"`
}

var defaultCookies *Cookies

// InitCookies 初始化 cookies 管理器
func InitCookies(filePath string) error {
	defaultCookies = &Cookies{
		filePath: filePath,
		Data:     make(map[string]interface{}),
	}

	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 尝试加载已有 cookies
	return defaultCookies.load()
}

// GetCookies 获取指定 key 的 cookie 值
func GetCookies() map[string]interface{} {
	if defaultCookies == nil {
		return nil
	}
	defaultCookies.mu.RLock()
	defer defaultCookies.mu.RUnlock()
	return defaultCookies.Data
}

// SetCookies 设置 cookies
func SetCookies(data map[string]interface{}) error {
	if defaultCookies == nil {
		return nil
	}
	defaultCookies.mu.Lock()
	defer defaultCookies.mu.Unlock()
	defaultCookies.Data = data
	return defaultCookies.save()
}

// DeleteCookies 删除 cookies
func DeleteCookies() error {
	if defaultCookies == nil {
		return nil
	}
	defaultCookies.mu.Lock()
	defer defaultCookies.mu.Unlock()
	defaultCookies.Data = make(map[string]interface{})
	if err := defaultCookies.save(); err != nil {
		return err
	}
	// 删除文件
	if err := os.Remove(defaultCookies.filePath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// GetCookiesFilePath 获取 cookies 文件路径
func GetCookiesFilePath() string {
	if defaultCookies == nil {
		return ""
	}
	return defaultCookies.filePath
}

// GetToken 获取认证 token
func GetToken() string {
	if defaultCookies == nil {
		return ""
	}
	defaultCookies.mu.RLock()
	defer defaultCookies.mu.RUnlock()
	if token, ok := defaultCookies.Data["token"]; ok {
		if t, ok := token.(string); ok {
			return t
		}
	}
	return ""
}

// GetCookieString 获取所有 cookies 的 HTTP Header 格式字符串
// 格式: "name1=value1; name2=value2"
func GetCookieString() string {
	if defaultCookies == nil {
		return ""
	}
	defaultCookies.mu.RLock()
	defer defaultCookies.mu.RUnlock()

	var parts []string
	for name, value := range defaultCookies.Data {
		if str, ok := value.(string); ok {
			parts = append(parts, name+"="+str)
		}
	}
	return strings.Join(parts, "; ")
}

func (c *Cookies) load() error {
	data, err := os.ReadFile(c.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(data, &c.Data)
}

func (c *Cookies) save() error {
	data, err := json.MarshalIndent(c.Data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.filePath, data, 0644)
}

// GetAllCookies 获取所有 cookies 为 map[string]string
func GetAllCookies() map[string]string {
	if defaultCookies == nil {
		return nil
	}
	defaultCookies.mu.RLock()
	defer defaultCookies.mu.RUnlock()

	result := make(map[string]string)
	for name, value := range defaultCookies.Data {
		if str, ok := value.(string); ok {
			result[name] = str
		}
	}
	return result
}
