// Package main is the entry point for dianshu-mcp.
// It provides an MCP server for the Dianshu (典枢) data platform.
//
// Author: zhyyao
package main

import (
	"dianshu-mcp/config"
	"dianshu-mcp/handler"
	"dianshu-mcp/logger"
	"dianshu-mcp/pkg/matomo"
	"dianshu-mcp/pkg/observability"
	"dianshu-mcp/service"
	"time"

	"github.com/getsentry/sentry-go"
)

// main is the application entry point.
func main() {
	cfg := config.ParseFlags()
	observability.ApplySentryDefaults(cfg)
	matomo.ApplyDefaults(cfg)

	logger.Info("dianshu-mcp starting", "port", cfg.Port, "headless", cfg.Headless)

	if observability.InitSentry(cfg) {
		defer sentry.Flush(2 * time.Second)
	}

	svc := service.New(cfg)
	h := handler.NewApp(svc, svc.Matomo())

	srv := newServer(h)
	if err := srv.start(cfg); err != nil {
		logger.Fatal("server failed to start", "error", err)
	}
}
