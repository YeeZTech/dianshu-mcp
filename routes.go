package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// setupRoutes 设置路由配置
func setupRoutes(appServer *AppServer) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// 中间件
	router.Use(errorHandlingMiddleware())
	router.Use(corsMiddleware())

	// 健康检查
	router.GET("/health", healthHandler)

	// MCP 端点 - 使用官方 SDK 的 Streamable HTTP Handler
	mcpHandler := mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server {
			return appServer.mcpServer
		},
		&mcp.StreamableHTTPOptions{
			JSONResponse: true,
		},
	)
	router.Any("/mcp", gin.WrapH(mcpHandler))
	router.Any("/mcp/*path", gin.WrapH(mcpHandler))

	// REST API 路由组（可选）
	api := router.Group("/api/v1")
	{
		api.GET("/login/status", appServer.checkLoginStatusHandler)
		api.GET("/login/qrcode", appServer.getLoginQRCodeHandler)
		api.DELETE("/login/cookies", appServer.deleteCookiesHandler)
		api.POST("/orders/query", appServer.listOrdersHandler)
		api.POST("/data/search", appServer.xhsSearchHandler)
		api.POST("/dataset/search", appServer.datasetSearchHandler)
	}

	return router
}

// healthHandler 健康检查
func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "dianshu-mcp",
		"version":   "1.0.0",
		"timestamp": "now",
	})
}

// errorHandlingMiddleware 错误处理中间件
func errorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": c.Errors.Last().Error(),
			})
		}
	}
}

// corsMiddleware CORS 中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
