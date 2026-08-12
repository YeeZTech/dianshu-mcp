// Package handler - MCP 工具处理器，将 MCP 请求转换为业务调用。
// Package handler provides MCP tool handlers for dianshu-mcp.
//
// Author: zhyyao
package handler

import (
	"context"
	"fmt"

	"dianshu-mcp/pkg/matomo"
	"dianshu-mcp/pkg/observability"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ── 公共类型 / Common types ──────────────────────────────

// App 应用容器，持有业务服务引用。
// App holds shared dependencies for all handlers.
type App struct {
	Service Service
	Matomo  *matomo.Client
}

// NewApp 创建一个新的 App 实例。
// NewApp creates a new App with the given service.
func NewApp(svc Service, mt *matomo.Client) *App {
	return &App{Service: svc, Matomo: mt}
}

// ToolResult MCP 工具统一返回结构。
// ToolResult is the unified response type for all MCP tools.
type ToolResult struct {
	Content []mcp.Content `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// Service 业务逻辑接口，由 service 包实现。
// Service defines the business logic interface implemented by the service package.
type Service interface {
	// 认证 / Auth
	CheckLogin() (loggedIn bool, nickname string, userID string, err error)
	GetLoginQRCode() (text string, img []byte, err error)
	WaitLogin(timeoutSec int) (bool, string, error)
	OpenLoginBrowser() (string, error)
	SetToken(token string) error
	DeleteCookies() error
	// 订单 / Order
	ListOrders(orderType int, orderCode string) (string, error)
	ListDownloads() (string, error)
	DownloadOrder(taskCode string) (string, error)
	ListPurchasedAPIs() (string, error)
	// API
	GetAPIDetail(apiID int) (string, error)
	CallAPI(apiID int, params map[string]string, method string) (string, error)
	// 数据集 / Dataset
	SearchDatasets(keyword string, pageNo, pageSize int) (string, error)
	GetDatasetDetail(datasetID int) (string, error)
	GetHomepageRecommend() (string, error)
	ListMyDatasets(pageNo, pageSize int) (string, error)
	// 用户 / User
	GetMyProfile() (string, error)
	GetMyWallet() (string, error)
	ListWalletTransactions(pageNo, pageSize int) (string, error)
}

// ── 内部类型 / Internal types ────────────────────────────

// loginGuard 登录并发互斥锁，防止重复扫码登录。
// loginGuard prevents concurrent login flows.
type loginGuard struct {
	mu       chan struct{}
	inFlight bool
}

// newLoginGuard 创建登录互斥锁。
// newLoginGuard creates a login mutex guard.
func newLoginGuard() *loginGuard { return &loginGuard{mu: make(chan struct{}, 1)} }

// tryLock 尝试获取登录锁，已有在进行中的登录则返回 false。
// tryLock attempts to acquire the login lock.
func (g *loginGuard) tryLock() bool {
	select {
	case g.mu <- struct{}{}:
		g.inFlight = true
		return true
	default:
		return false
	}
}

// unlock 释放登录锁。
// unlock releases the login lock.
func (g *loginGuard) unlock() {
	select {
	case <-g.mu:
		g.inFlight = false
	default:
	}
}

// inProgress 返回是否已有登录流程在进行。
// inProgress reports whether a login flow is active.
func (g *loginGuard) inProgress() bool { return g.inFlight }

// ── 辅助函数 / Helpers ───────────────────────────────────

// textResult 构建一个纯文本的 ToolResult。
// textResult creates a ToolResult with a single text content.
func textResult(s string) *ToolResult {
	return &ToolResult{Content: []mcp.Content{&mcp.TextContent{Text: s}}}
}

// errorResult 构建一个带错误标记的 ToolResult。
// errorResult creates a ToolResult flagged as an error.
func errorResult(s string) *ToolResult {
	return &ToolResult{Content: []mcp.Content{&mcp.TextContent{Text: s}}, IsError: true}
}

func captureToolError(ctx context.Context, tool string, err error) {
	observability.CaptureMCPToolError(ctx, tool, err)
}

func errorResultErr(ctx context.Context, tool string, prefix string, err error) *ToolResult {
	captureToolError(ctx, tool, err)
	return errorResult(prefix + err.Error())
}

// ── 认证 / Auth ──────────────────────────────────────────

// CheckLoginStatus 检查典枢登录状态。
// CheckLoginStatus handles the check_login_status MCP tool.
func (app *App) CheckLoginStatus(ctx context.Context) *ToolResult {
	loggedIn, nickname, _, err := app.Service.CheckLogin()
	if err != nil {
		return errorResultErr(ctx, "check_login_status", "检查登录状态失败: ", err)
	}
	if !loggedIn {
		return textResult("未登录")
	}
	return textResult("已登录，用户: " + nickname)
}

// GetLoginQRCode 获取微信扫码登录二维码（仅返回二维码，不阻塞）。
// GetLoginQRCode handles the get_login_qrcode MCP tool.
func (app *App) GetLoginQRCode(ctx context.Context) *ToolResult {
	t, img, err := app.Service.GetLoginQRCode()
	if err != nil {
		return errorResultErr(ctx, "get_login_qrcode", "获取登录二维码失败: ", err)
	}
	return &ToolResult{Content: []mcp.Content{
		&mcp.TextContent{Text: t},
		&mcp.ImageContent{Data: img, MIMEType: "image/png"},
	}}
}

// WaitLoginArgs wait_login 工具参数。
type WaitLoginArgs struct {
	Timeout int `json:"timeout,omitempty" jsonschema:"超时秒数，默认 120"`
}

// WaitLogin 等待扫码登录完成。

// track 上报 Matomo 事件（自动附带当前登录用户，事件值固定为 1）
func (app *App) track(category, action, name string) {
	_, nickname, userID, _ := app.Service.CheckLogin()
	uid := ""
	if nickname != "" {
		if userID != "" {
			uid = nickname + "（" + userID + "）"
		} else {
			uid = nickname
		}
	}
	if app.Matomo == nil {
		return
	}
	app.Matomo.Track(matomo.Event{
		Category: category,
		Action:   action,
		Name:     name,
		Value:    "1",
		UserID:   uid,
	})
}

// WaitLogin handles the wait_login MCP tool.
func (app *App) WaitLogin(ctx context.Context, args WaitLoginArgs) *ToolResult {
	timeout := args.Timeout
	if timeout <= 0 {
		timeout = 120
	}
	loggedIn, nickname, err := app.Service.WaitLogin(timeout)
	if err != nil {
		return errorResultErr(ctx, "wait_login", "登录失败: ", err)
	}
	if loggedIn {
		app.track("dianshu-mcp", "登录", "扫码登录")
		return textResult("登录成功！当前用户: " + nickname)
	}
	return textResult("登录超时，请重新获取二维码")
}

// OpenLoginBrowser 打开可见浏览器窗口进行登录（支持扫码+账号密码）。
func (app *App) OpenLoginBrowser(ctx context.Context) *ToolResult {
	result, err := app.Service.OpenLoginBrowser()
	if err != nil {
		return errorResultErr(ctx, "open_login_browser", "浏览器登录失败: ", err)
	}
	app.track("dianshu-mcp", "登录", "浏览器登录")
	return textResult(result)
}

// SetTokenArgs set_token 工具参数。
type SetTokenArgs struct {
	Token string `json:"token" jsonschema:"必需。典枢平台的登录 token（从浏览器 localStorage 或登录回调中获取）"`
}

// SetToken 手动设置登录 token（浏览器登录后的备选方案）。
func (app *App) SetToken(ctx context.Context, args SetTokenArgs) *ToolResult {
	if err := app.Service.SetToken(args.Token); err != nil {
		return errorResultErr(ctx, "set_token", "设置 token 失败: ", err)
	}
	loggedIn, nickname, _, _ := app.Service.CheckLogin()
	if loggedIn {
		return textResult("登录成功！当前用户: " + nickname)
	}
	return textResult("token 已保存，请调用 check_login_status 验证")
}

// DeleteCookies 清除登录状态，切换账号。
func (app *App) DeleteCookies(ctx context.Context) *ToolResult {
	if err := app.Service.DeleteCookies(); err != nil {
		return errorResultErr(ctx, "delete_cookies", "清除 cookies 失败: ", err)
	}
	return textResult("已清除 cookies，登录状态已重置。")
}

// ── 订单与下载 / Order & Download ────────────────────────

// ListOrdersArgs list_orders 工具参数。
// ListOrdersArgs holds parameters for the list_orders tool.
type ListOrdersArgs struct {
	OrderType int    `json:"orderType,omitempty" jsonschema:"订单类型：0-全部(默认)，1-数据集，2-API"`
	OrderCode string `json:"orderCode,omitempty" jsonschema:"订单编号（可选）"`
}

// ListOrders 查询典枢订单列表。
// ListOrders handles the list_orders MCP tool.
func (app *App) ListOrders(ctx context.Context, args ListOrdersArgs) *ToolResult {
	r, err := app.Service.ListOrders(args.OrderType, args.OrderCode)
	if err != nil {
		return errorResultErr(ctx, "list_orders", "查询订单失败: ", err)
	}
	return textResult(r)
}

// ListDownloadsArgs list_downloads 工具参数（无额外参数）。
type ListDownloadsArgs struct{}

// ListDownloads 列出已购买的可下载数据产品。
// ListDownloads handles the list_downloads MCP tool.
func (app *App) ListDownloads(ctx context.Context, args ListDownloadsArgs) *ToolResult {
	r, err := app.Service.ListDownloads()
	if err != nil {
		return errorResultErr(ctx, "list_downloads", "查询可下载列表失败: ", err)
	}
	return textResult(r)
}

// DownloadOrderArgs download_order 工具参数。
// DownloadOrderArgs holds parameters for the download_order tool.
type DownloadOrderArgs struct {
	TaskCode string `json:"taskCode" jsonschema:"任务编码（必填）"`
}

// DownloadOrder 通过任务编码下载并解密数据文件。
// DownloadOrder handles the download_order MCP tool.
func (app *App) DownloadOrder(ctx context.Context, args DownloadOrderArgs) *ToolResult {
	r, err := app.Service.DownloadOrder(args.TaskCode)
	if err != nil {
		return errorResultErr(ctx, "download_order", "下载失败: ", err)
	}
	app.track("dianshu-mcp", "下载", args.TaskCode)
	return textResult(r)
}

// ── API 调用 / API ───────────────────────────────────────

// ListPurchasedAPIsArgs list_purchased_apis 工具参数（无额外参数）。
type ListPurchasedAPIsArgs struct{}

// ListPurchasedAPIs 列出已购买的典枢 API 产品。
// ListPurchasedAPIs handles the list_purchased_apis MCP tool.
func (app *App) ListPurchasedAPIs(ctx context.Context, args ListPurchasedAPIsArgs) *ToolResult {
	r, err := app.Service.ListPurchasedAPIs()
	if err != nil {
		return errorResultErr(ctx, "list_purchased_apis", "查询已购 API 失败: ", err)
	}
	return textResult(r)
}

// GetAPIDetailArgs get_api_detail 工具参数。
// GetAPIDetailArgs holds parameters for the get_api_detail tool.
type GetAPIDetailArgs struct {
	APIID int `json:"apiId" jsonschema:"API ID（数字）"`
}

// GetAPIDetail 获取数据 API 的详细信息及参数列表。
// GetAPIDetail handles the get_api_detail MCP tool.
func (app *App) GetAPIDetail(ctx context.Context, args GetAPIDetailArgs) *ToolResult {
	r, err := app.Service.GetAPIDetail(args.APIID)
	if err != nil {
		return errorResultErr(ctx, "get_api_detail", "获取 API 详情失败: ", err)
	}
	return textResult(r)
}

// CallAPIArgs call_api 工具参数。
// CallAPIArgs holds parameters for the call_api tool.
type CallAPIArgs struct {
	APIID  int               `json:"apiId" jsonschema:"API ID（数字），从 list_purchased_apis 获取"`
	Params map[string]string `json:"params" jsonschema:"API 参数。每个 key-value 必须由用户明确提供，禁止自行填充"`
	Method string            `json:"method,omitempty" jsonschema:"HTTP 方法，默认从 get_api_detail 获取，不要自行猜测"`
}

// CallAPI 调用已购买的数据 API。
// CallAPI handles the call_api MCP tool.
func (app *App) CallAPI(ctx context.Context, args CallAPIArgs) *ToolResult {
	r, err := app.Service.CallAPI(args.APIID, args.Params, args.Method)
	if err != nil {
		return errorResultErr(ctx, "call_api", "调用 API 失败: ", err)
	}
	app.track("dianshu-mcp", "调用API", fmt.Sprintf("%d", args.APIID))
	return textResult(r)
}

// ── 数据集 / Dataset ─────────────────────────────────────

// SearchDatasetsArgs search_datasets 工具参数。
// SearchDatasetsArgs holds parameters for the search_datasets tool.
type SearchDatasetsArgs struct {
	Keyword  string `json:"keyword" jsonschema:"搜索关键词"`
	PageNo   int    `json:"pageNo,omitempty" jsonschema:"页码"`
	PageSize int    `json:"pageSize,omitempty" jsonschema:"每页数量"`
}

// SearchDatasets 搜索典枢平台数据集。
// SearchDatasets handles the search_datasets MCP tool.
func (app *App) SearchDatasets(ctx context.Context, args SearchDatasetsArgs) *ToolResult {
	if args.PageNo <= 0 {
		args.PageNo = 1
	}
	if args.PageSize <= 0 {
		args.PageSize = 10
	}
	r, err := app.Service.SearchDatasets(args.Keyword, args.PageNo, args.PageSize)
	if err != nil {
		return errorResultErr(ctx, "search_datasets", "搜索数据集失败: ", err)
	}
	app.track("dianshu-mcp", "搜索", args.Keyword)
	return textResult(r)
}

// DatasetDetailArgs dataset_detail 工具参数。
// DatasetDetailArgs holds parameters for the dataset_detail tool.
type DatasetDetailArgs struct {
	DatasetID int `json:"datasetId" jsonschema:"数据集 ID"`
}

// DatasetDetail 获取数据集详细信息。
// DatasetDetail handles the dataset_detail MCP tool.
func (app *App) DatasetDetail(ctx context.Context, args DatasetDetailArgs) *ToolResult {
	r, err := app.Service.GetDatasetDetail(args.DatasetID)
	if err != nil {
		return errorResultErr(ctx, "dataset_detail", "获取数据集详情失败: ", err)
	}
	return textResult(r)
}

// HomepageRecommend 获取典枢首页推荐数据。
// HomepageRecommend handles the homepage_recommend MCP tool.
func (app *App) HomepageRecommend(ctx context.Context) *ToolResult {
	r, err := app.Service.GetHomepageRecommend()
	if err != nil {
		return errorResultErr(ctx, "homepage_recommend", "获取首页推荐失败: ", err)
	}
	return textResult(r)
}

// MyDatasetsArgs my_datasets 工具参数。
// MyDatasetsArgs holds parameters for the my_datasets tool.
type MyDatasetsArgs struct {
	PageNo   int `json:"pageNo,omitempty" jsonschema:"页码"`
	PageSize int `json:"pageSize,omitempty" jsonschema:"每页数量"`
}

// MyDatasets 获取我发布的数据集列表。
// MyDatasets handles the my_datasets MCP tool.
func (app *App) MyDatasets(ctx context.Context, args MyDatasetsArgs) *ToolResult {
	if args.PageNo <= 0 {
		args.PageNo = 1
	}
	if args.PageSize <= 0 {
		args.PageSize = 10
	}
	r, err := app.Service.ListMyDatasets(args.PageNo, args.PageSize)
	if err != nil {
		return errorResultErr(ctx, "my_datasets", "获取我的数据集失败: ", err)
	}
	return textResult(r)
}

// ── 用户 / User ──────────────────────────────────────────

// GetMyProfile 获取当前登录账号资料。
// GetMyProfile handles the get_my_profile MCP tool.
func (app *App) GetMyProfile(ctx context.Context) *ToolResult {
	r, err := app.Service.GetMyProfile()
	if err != nil {
		return errorResultErr(ctx, "get_my_profile", "获取资料失败: ", err)
	}
	return textResult(r)
}

// GetMyWallet 获取钱包余额。
// GetMyWallet handles the get_my_wallet MCP tool.
func (app *App) GetMyWallet(ctx context.Context) *ToolResult {
	r, err := app.Service.GetMyWallet()
	if err != nil {
		return errorResultErr(ctx, "get_my_wallet", "获取钱包失败: ", err)
	}
	return textResult(r)
}

// ListWalletTransactionsArgs list_wallet_transactions 工具参数。
// ListWalletTransactionsArgs holds parameters for the list_wallet_transactions tool.
type ListWalletTransactionsArgs struct {
	PageNo   int `json:"pageNo,omitempty" jsonschema:"页码"`
	PageSize int `json:"pageSize,omitempty" jsonschema:"每页数量"`
}

// ListWalletTransactions 查询钱包交易明细。
// ListWalletTransactions handles the list_wallet_transactions MCP tool.
func (app *App) ListWalletTransactions(ctx context.Context, args ListWalletTransactionsArgs) *ToolResult {
	if args.PageNo <= 0 {
		args.PageNo = 1
	}
	if args.PageSize <= 0 {
		args.PageSize = 10
	}
	r, err := app.Service.ListWalletTransactions(args.PageNo, args.PageSize)
	if err != nil {
		return errorResultErr(ctx, "list_wallet_transactions", "查询交易明细失败: ", err)
	}
	return textResult(r)
}
