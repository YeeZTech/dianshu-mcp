// Package main - application server container.
//
// Author: zhyyao
package main

import (
	"fmt"

	"dianshu-mcp/config"
	"dianshu-mcp/handler"
	"dianshu-mcp/logger"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// server holds the application runtime dependencies.
type server struct {
	handler   *handler.App
	mcpServer *mcp.Server
}

// newServer creates a new server instance with MCP tools registered.
func newServer(h *handler.App) *server {
	mcpSrv := initMCPServer(h)
	return &server{handler: h, mcpServer: mcpSrv}
}

// start begins the HTTP server on the given port.
func (s *server) start(cfg *config.Config) error {
	router := setupRoutes(cfg, s.handler, s.mcpServer)
	addr := fmt.Sprintf(":%d", cfg.Port)
	logger.Info("HTTP server starting", "port", cfg.Port)
	return router.Run(addr)
}
