package configs

import "os"

var (
	headlessMode bool
	browserBin   string
)

const (
	// DefaultPort 默认服务端口
	DefaultPort = ":18061"
	// CookieFileName cookie 文件名
	CookieFileName = "cookies.json"
)

// InitHeadless 初始化无头模式
func InitHeadless(headless bool) {
	headlessMode = headless
}

// IsHeadless 是否无头模式
func IsHeadless() bool {
	return headlessMode
}

// SetBinPath 设置浏览器二进制路径
func SetBinPath(binPath string) {
	browserBin = binPath
	if binPath != "" {
		os.Setenv("ROD_BROWSER_BIN", binPath)
	}
}

// GetBinPath 获取浏览器二进制路径
func GetBinPath() string {
	return browserBin
}
