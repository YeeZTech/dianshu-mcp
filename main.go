// Package main is the entry point for dianshu-mcp.
// It provides an MCP server for the Dianshu (典枢) data platform.
//
// Author: zhyyao
package main

import (
	"dianshu-mcp/config"
	"dianshu-mcp/handler"
	"dianshu-mcp/logger"
	"dianshu-mcp/service"
)

// main is the application entry point.
func main() {
	cfg := config.ParseFlags()

	logger.Info("dianshu-mcp starting", "port", cfg.Port, "headless", cfg.Headless)

	svc := service.New(cfg)
	h := handler.NewApp(svc)

	srv := newServer(h)
	if err := srv.start(cfg.Port); err != nil {
		logger.Fatal("server failed to start", "error", err)
	}
}
