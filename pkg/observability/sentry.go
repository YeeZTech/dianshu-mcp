package observability

import (
	"context"

	"dianshu-mcp/config"
	"dianshu-mcp/logger"

	"github.com/getsentry/sentry-go"
)

var defaultSentryDSN string
var defaultSentryEnvironment string
var defaultSentryRelease string

func ApplySentryDefaults(cfg *config.Config) {
	if cfg.SentryDSN == "" {
		cfg.SentryDSN = defaultSentryDSN
	}
	if cfg.SentryEnvironment == "" || (cfg.SentryEnvironment == "local" && defaultSentryEnvironment != "") {
		cfg.SentryEnvironment = defaultSentryEnvironment
	}
	if cfg.SentryRelease == "" {
		cfg.SentryRelease = defaultSentryRelease
	}
}

func InitSentry(cfg *config.Config) bool {
	if cfg.SentryDSN == "" {
		return false
	}
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:         cfg.SentryDSN,
		Environment: cfg.SentryEnvironment,
		Release:     cfg.SentryRelease,
	}); err != nil {
		logger.Error("sentry initialization failed", "error", err)
		return false
	}
	logger.Info("sentry initialized", "environment", cfg.SentryEnvironment, "release", cfg.SentryRelease)
	return true
}

func CaptureException(ctx context.Context, err error, level sentry.Level, tags map[string]string) {
	if err == nil {
		return
	}
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub()
	}
	if hub == nil || hub.Client() == nil {
		return
	}
	hub.WithScope(func(scope *sentry.Scope) {
		scope.SetLevel(level)
		for k, v := range tags {
			scope.SetTag(k, v)
		}
		hub.CaptureException(err)
	})
}

func CaptureMCPToolError(ctx context.Context, tool string, err error) {
	tags := map[string]string{"component": "mcp"}
	if tool != "" {
		tags["mcp.tool"] = tool
	}
	CaptureException(ctx, err, sentry.LevelWarning, tags)
}
