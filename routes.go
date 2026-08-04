// Package main - HTTP routes and middleware.
//
// Author: zhyyao
package main

import (
	"net/http"
	"time"

	"dianshu-mcp/config"
	"dianshu-mcp/handler"

	sentrygin "github.com/getsentry/sentry-go/gin"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// setupRoutes configures HTTP routes including MCP endpoint and health check.
func setupRoutes(cfg *config.Config, h *handler.App, mcpServer *mcp.Server) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger())
	if cfg.SentryDSN != "" {
		router.Use(sentrygin.New(sentrygin.Options{
			Repanic: true,
		}))
	}
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	router.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	var mcpHandler http.Handler = mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server { return mcpServer },
		&mcp.StreamableHTTPOptions{JSONResponse: true},
	)
	if cfg.SentryDSN != "" {
		mcpHandler = sentryhttp.New(sentryhttp.Options{
			Repanic:         true,
			WaitForDelivery: false,
			Timeout:         2 * time.Second,
		}).Handle(mcpHandler)
	}
	router.Any("/mcp", gin.WrapH(mcpHandler))
	router.Any("/mcp/*path", gin.WrapH(mcpHandler))

	return router
}

// corsMiddleware adds CORS headers for cross-origin requests.
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
