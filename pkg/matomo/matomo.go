package matomo

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"dianshu-mcp/config"
)

// 编译时通过 ldflags 注入：
//   -X dianshu-mcp/pkg/matomo.defaultEndpoint=...
//   -X dianshu-mcp/pkg/matomo.defaultSiteID=...
var (
	defaultEndpoint string
	defaultSiteID   string
)

// ApplyDefaults 将编译时注入的默认值写入配置（仅在配置为空时生效）。
func ApplyDefaults(cfg *config.Config) {
	if cfg.MatomoEndpoint == "" && defaultEndpoint != "" {
		cfg.MatomoEndpoint = defaultEndpoint
	}
	if cfg.MatomoSiteID == "" && defaultSiteID != "" {
		cfg.MatomoSiteID = defaultSiteID
	}
}

// Client 上报客户端
type Client struct {
	endpoint string
	siteID   string
	client   *http.Client
}

// New 创建 Matomo 上报客户端
func New(endpoint, siteID string) *Client {
	return &Client{
		endpoint: endpoint,
		siteID:   siteID,
		client:   &http.Client{Timeout: 5 * time.Second},
	}
}

// Event 事件参数
type Event struct {
	Category string
	Action   string
	Name     string
	Value    string
	UserID   string
	PageURL  string
}

// Track 上报事件（异步，失败不报错）
func (c *Client) Track(e Event) {
	go c.trackSync(e)
}

func (c *Client) trackSync(e Event) {
	params := url.Values{
		"idsite": {c.siteID},
		"rec":    {"1"},
		"e_c":    {e.Category},
		"e_a":    {e.Action},
	}
	params.Set("e_n", e.Name)
	if e.Value != "" {
		params.Set("e_v", e.Value)
	}
	if e.UserID != "" {
		params.Set("uid", e.UserID)
	}
	if e.PageURL != "" {
		params.Set("url", e.PageURL)
	}

	req, err := http.NewRequest("POST", c.endpoint, strings.NewReader(params.Encode()))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
	req.Header.Set("User-Agent", "dianshu-mcp/1.0")

	resp, err := c.client.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}

// Trackf 便捷方法
func Trackf(c *Client, category, action, nameFormat string, args ...interface{}) {
	if c == nil {
		return
	}
	c.Track(Event{
		Category: category,
		Action:   action,
		Name:     fmt.Sprintf(nameFormat, args...),
	})
}
