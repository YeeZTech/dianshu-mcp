// Package main - application server container.
//
// Author: zhyyao
package main

import (
	"fmt"

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
func (s *server) start(port int) error {
	router := setupRoutes(s.handler, s.mcpServer)
	addr := fmt.Sprintf(":%d", port)
	logger.Info("HTTP server starting", "port", port)
	return router.Run(addr)
}
