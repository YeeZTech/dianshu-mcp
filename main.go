package main

import (
	"flag"
	"os"

	"dianshu-mcp/configs"
	"dianshu-mcp/cookies"
	"dianshu-mcp/dianshu"

	"github.com/sirupsen/logrus"
)

func main() {
	var (
		headless bool
		binPath  string
		port     string
	)
	flag.BoolVar(&headless, "headless", true, "是否无头模式（默认启用，不显示浏览器窗口）")
	flag.StringVar(&binPath, "bin", "", "浏览器二进制文件路径")
	flag.StringVar(&port, "port", configs.DefaultPort, "服务端口")
	flag.Parse()

	if len(binPath) == 0 {
		binPath = os.Getenv("ROD_BROWSER_BIN")
	}

	configs.InitHeadless(headless)
	configs.SetBinPath(binPath)

	// 初始化 cookies 持久化
	cookiePath := dianshu.GetDefaultCookiePath()
	if err := cookies.InitCookies(cookiePath); err != nil {
		logrus.Warnf("初始化 cookies 失败: %v", err)
	}

	// 初始化服务
	dianshuService := NewDianshuService()

	// 创建并启动应用服务器
	appServer := NewAppServer(dianshuService)
	if err := appServer.Start(port); err != nil {
		logrus.Fatalf("服务启动失败: %v", err)
	}
}
