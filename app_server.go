package main

import (
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
)

// AppServer 应用服务
type AppServer struct {
	dianshuService *DianshuService
	mcpServer      *mcp.Server

	loginMu         sync.Mutex
	loginInProgress bool
}

// NewAppServer 创建应用服务
func NewAppServer(dianshuService *DianshuService) *AppServer {
	app := &AppServer{
		dianshuService: dianshuService,
	}

	// 初始化 MCP Server
	app.mcpServer = InitMCPServer(app)

	return app
}

// Start 启动 HTTP 服务
func (s *AppServer) Start(port string) error {
	router := setupRoutes(s)
	logrus.Infof("启动 HTTP 服务器: %s", port)
	return router.Run(port)
}
