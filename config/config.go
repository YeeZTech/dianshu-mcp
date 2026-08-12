// Package config provides centralized configuration for dianshu-mcp.
// All configurable values are defined here with sensible defaults,
// and can be overridden via command-line flags or environment variables.
//
// Author: zhyyao
package config

import (
	"flag"
	"os"
	"path/filepath"
)

// Config holds all configuration for the application.
type Config struct {
	// Server
	Port     int
	Headless bool

	// Observability
	SentryDSN         string
	SentryEnvironment string
	SentryRelease     string
	MatomoEndpoint    string
	MatomoSiteID      string

	// Dianshu API
	BaseAPIURL     string
	DataAPIGateway string
	DownloadCDN    string
	WebURL         string

	// Output directories
	OutputDir    string
	DownloadsDir string
	APIDataDir   string

	// Cookie
	CookieFile string

	// Chain
	ChainTaskCheckInterval int // seconds
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	downloadsBase := filepath.Join(home, "Downloads", "dianshu-mcp")
	return &Config{
		Port:                   18061,
		Headless:               false,
		SentryDSN:              "",
		SentryEnvironment:      "local",
		SentryRelease:          "",
		MatomoEndpoint:         "",
		MatomoSiteID:           "",
		BaseAPIURL:             "https://api.dianshudata.com",
		DataAPIGateway:         "https://data-api.dianshudata.com",
		DownloadCDN:            "https://d.dianshudata.com",
		WebURL:                 "https://dianshudata.com",
		OutputDir:              downloadsBase,
		DownloadsDir:           "downloads",
		APIDataDir:             "api-data",
		CookieFile:             "cookies.json",
		ChainTaskCheckInterval: 2,
	}
}

// ParseFlags parses command-line flags into the config.
// Returns the config with flag values applied over defaults.
func ParseFlags() *Config {
	cfg := DefaultConfig()

	flag.IntVar(&cfg.Port, "port", cfg.Port, "HTTP server port")
	flag.BoolVar(&cfg.Headless, "headless", cfg.Headless, "Run browser in headless mode")
	flag.StringVar(&cfg.SentryDSN, "sentry-dsn", cfg.SentryDSN, "Sentry DSN")
	flag.StringVar(&cfg.MatomoEndpoint, "matomo-endpoint", cfg.MatomoEndpoint, "Matomo endpoint URL")
	flag.StringVar(&cfg.MatomoSiteID, "matomo-site-id", cfg.MatomoSiteID, "Matomo site ID")
	flag.StringVar(&cfg.SentryEnvironment, "sentry-env", cfg.SentryEnvironment, "Sentry environment")
	flag.StringVar(&cfg.SentryRelease, "sentry-release", cfg.SentryRelease, "Sentry release")
	flag.StringVar(&cfg.BaseAPIURL, "api-url", cfg.BaseAPIURL, "Dianshu API base URL")
	flag.StringVar(&cfg.CookieFile, "cookie-file", cfg.CookieFile, "Cookie file path")
	flag.StringVar(&cfg.OutputDir, "output-dir", cfg.OutputDir, "Output root directory")
	flag.Parse()

	// Apply environment variable overrides
	if v := os.Getenv("DS_PORT"); v != "" {
		cfg.Port = atoiOrDefault(v, cfg.Port)
	}
	if v := os.Getenv("DS_API_URL"); v != "" {
		cfg.BaseAPIURL = v
	}
	if v := os.Getenv("DS_COOKIE_FILE"); v != "" {
		cfg.CookieFile = v
	}
	if v := os.Getenv("DS_SENTRY_DSN"); v != "" {
		cfg.SentryDSN = v
	}
	if v := os.Getenv("DS_SENTRY_ENV"); v != "" {
		cfg.SentryEnvironment = v
	}
	if v := os.Getenv("DS_SENTRY_RELEASE"); v != "" {
		cfg.SentryRelease = v
	}
	if v := os.Getenv("DS_MATOMO_ENDPOINT"); v != "" {
		cfg.MatomoEndpoint = v
	}
	if v := os.Getenv("DS_MATOMO_SITE_ID"); v != "" {
		cfg.MatomoSiteID = v
	}

	return cfg
}

// DownloadsPath returns the full path to the downloads directory.
func (c *Config) DownloadsPath() string {
	return filepath.Join(c.OutputDir, c.DownloadsDir)
}

// APIDataPath returns the full path to the API data directory.
func (c *Config) APIDataPath() string {
	return filepath.Join(c.OutputDir, c.APIDataDir)
}

// DatasetURL returns the web URL for a dataset detail page.
func (c *Config) DatasetURL(id int) string {
	return c.WebURL + "/dataset/" + itoa(id)
}

// atoiOrDefault parses a string to int, returning def on failure.
func atoiOrDefault(s string, def int) int {
	n := 0
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return def
		}
		n = n*10 + int(ch-'0')
	}
	return n
}

// itoa converts an int to its decimal string representation.

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
