package dianshu

import (
	"context"
	"encoding/json"
	"fmt"
)

// DataQueryResult 表示统一的数据查询返回结果。
type DataQueryResult struct {
	ResultCode int             `json:"resultCode"`
	ResultDesc string          `json:"resultDesc"`
	Data       interface{}     `json:"data"`
	DSSeqNo    string          `json:"DSSeqNo"`
	RawJSON    json.RawMessage `json:"-"`
}

// XiaohongshuSearchResponse 表示小红书搜索数据源的原始返回结构。
type XiaohongshuSearchResponse struct {
	ResultCode int                    `json:"resultCode"`
	ResultDesc string                 `json:"resultDesc"`
	Data       *XiaohongshuSearchData `json:"data"`
	DSSeqNo    string                 `json:"DSSeqNo"`
}

// XiaohongshuSearchData 表示小红书搜索数据区块。
type XiaohongshuSearchData struct {
	Took   int                      `json:"took"`
	Result []map[string]interface{} `json:"result"`
}

// DataQueryRequest 表示统一的数据查询请求。
type DataQueryRequest struct {
	ProviderType string
	DatasetType  string
	SiteDomain   string
	AccessToken  string
	Body         map[string]string
	RawQuery     string
}

// DataQueryProvider 定义统一的数据查询接口。
type DataQueryProvider interface {
	Query(ctx context.Context, request DataQueryRequest) (*DataQueryResult, error)
}

// ChargeRequest 表示扣费请求。
type ChargeRequest struct {
	ProviderType string
	DatasetType  string
	Amount       string
	Description  string
}

// ChargeService 定义通用扣费接口，当前仅预留抽象。
type ChargeService interface {
	Charge(ctx context.Context, request ChargeRequest) error
}

// NoopChargeService 表示未启用扣费逻辑的占位实现。
type NoopChargeService struct{}

// Charge 当前不执行扣费，仅保留统一入口。
func (s *NoopChargeService) Charge(ctx context.Context, request ChargeRequest) error {
	return nil
}

// DataQueryRouter 按 providerType + datasetType 路由到具体数据源实现。
type DataQueryRouter struct {
	xiaohongshuSearchProvider DataQueryProvider
}

// NewDataQueryRouter 创建统一的数据查询路由器。
func NewDataQueryRouter() *DataQueryRouter {
	return &DataQueryRouter{
		xiaohongshuSearchProvider: NewXiaohongshuSearchProvider(),
	}
}

// Query 根据数据源类型路由查询。
func (r *DataQueryRouter) Query(ctx context.Context, request DataQueryRequest) (*DataQueryResult, error) {
	providerType := normalizeProviderType(request.ProviderType)
	datasetType := normalizeDatasetType(request.DatasetType)
	if providerType == "" {
		return nil, fmt.Errorf("providerType 不能为空")
	}
	if datasetType == "" {
		return nil, fmt.Errorf("datasetType 不能为空")
	}

	switch {
	case providerType == ProviderTypeXiaohongshu && datasetType == DatasetTypeSearch:
		return r.xiaohongshuSearchProvider.Query(ctx, request)
	default:
		return nil, fmt.Errorf("暂不支持的数据源类型: providerType=%s datasetType=%s", providerType, datasetType)
	}
}
