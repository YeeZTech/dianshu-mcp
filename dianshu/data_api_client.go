package dianshu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	dataAPIBaseURL           = "https://data-api.dianshudata.com"
	dataAPIPrivateKeyPath    = "/api/privateKey"
	dataAPIPostGatewayPath   = "/api/post/"
	dataAPIGetGatewayPath    = "/api/get/"
	dataAPIRequestMethodGet  = "GET"
	dataAPIRequestMethodPost = "POST"
)

// DataAPIContext 表示一次 data-api 调用的上下文。
type DataAPIContext struct {
	AppCode    string
	BaseURL    string
	KeyPair    *DataAPIKeyPair
	HTTPClient *http.Client
}

// DataAPIConfig 表示 /api/privateKey 返回的加密配置。
type DataAPIConfig struct {
	EnclaveHash   string `json:"enclaveHash"`
	DianPublicKey string `json:"dianPkey"`
	APIHash       string `json:"apiHash"`
	ResultAPIHash string `json:"resultApiHash"`
}

// DataAPIParam 表示请求参数键值对。
type DataAPIParam struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// DataAPIPostRequest 表示卖家 API 的 POST 请求参数。
type DataAPIPostRequest struct {
	BodyParams     []DataAPIParam
	RequestHeaders []DataAPIParam
}

// DataAPIGetRequest 表示卖家 API 的 GET 请求参数。
type DataAPIGetRequest struct {
	QueryParams    []DataAPIParam
	RequestHeaders []DataAPIParam
}

// DataAPIEncryptedPayload 表示加密前的卖家 API 请求体。
type DataAPIEncryptedPayload struct {
	APIHash   string            `json:"api_hash"`
	APIMethod string            `json:"api_method"`
	APIHeader map[string]string `json:"api_header,omitempty"`
	APIBody   map[string]string `json:"api_body,omitempty"`
	APIParam  map[string]string `json:"api_param,omitempty"`
}

type dataAPIPrivateKeyRequest struct {
	AppCode string `json:"appCode"`
	APIID   string `json:"apiId"`
}

type dataAPIGatewayRequest struct {
	Params string `json:"params"`
}

type dataAPIEncryptEnvelope struct {
	EncryptedParam string `json:"encrypted_param"`
	ShuPublicKey   string `json:"shu_pkey"`
}

type dataAPIAuditInfo struct {
	DataHash               string `json:"dataHash"`
	DataShuPublicKey       string `json:"dataShuPublicKey"`
	AllowedEnclaveHash     string `json:"allowedEnclaveHash"`
	EncryptedShuPrivateKey string `json:"encryptedShuPrivateKey"`
	ShuKeyForwardSignature string `json:"shuKeyForwardSignature"`
}

type dataAPIResult[T any] struct {
	ResultCode int    `json:"resultCode"`
	ResultDesc string `json:"resultDesc"`
	Data       T      `json:"data"`
}

// NewDataAPIContext 创建 data-api 调用上下文。
func NewDataAPIContext(appCode string) (*DataAPIContext, error) {
	appCode = strings.TrimSpace(appCode)
	if appCode == "" {
		return nil, fmt.Errorf("appCode 不能为空")
	}
	keyPair, err := GenerateDataAPIKeyPair()
	if err != nil {
		return nil, err
	}
	return &DataAPIContext{
		AppCode:    appCode,
		BaseURL:    dataAPIBaseURL,
		KeyPair:    keyPair,
		HTTPClient: &http.Client{Timeout: apiRequestTimeout},
	}, nil
}

// DataAPIGatewayClient 封装 data-api 网关请求。
type DataAPIGatewayClient struct {
	context *DataAPIContext
}

// NewDataAPIGatewayClient 创建 data-api 网关客户端。
func NewDataAPIGatewayClient(dataAPIContext *DataAPIContext) (*DataAPIGatewayClient, error) {
	if dataAPIContext == nil {
		return nil, fmt.Errorf("dataAPIContext 不能为空")
	}
	if dataAPIContext.KeyPair == nil {
		return nil, fmt.Errorf("dataAPIContext.KeyPair 不能为空")
	}
	if dataAPIContext.HTTPClient == nil {
		dataAPIContext.HTTPClient = &http.Client{Timeout: apiRequestTimeout}
	}
	if strings.TrimSpace(dataAPIContext.BaseURL) == "" {
		dataAPIContext.BaseURL = dataAPIBaseURL
	}
	return &DataAPIGatewayClient{context: dataAPIContext}, nil
}

// FetchConfig 查询指定 API 的加密配置。
func (c *DataAPIGatewayClient) FetchConfig(ctx context.Context, apiCode string) (*DataAPIConfig, error) {
	apiCode = strings.TrimSpace(apiCode)
	if apiCode == "" {
		return nil, fmt.Errorf("apiCode 不能为空")
	}

	requestBody := dataAPIPrivateKeyRequest{AppCode: c.context.AppCode, APIID: apiCode}
	responseBody, err := c.postJSON(ctx, dataAPIPrivateKeyPath, nil, requestBody)
	if err != nil {
		return nil, err
	}

	var result dataAPIResult[*DataAPIConfig]
	if err = json.Unmarshal(responseBody, &result); err != nil {
		return nil, fmt.Errorf("解析 data-api 配置失败: %w", err)
	}
	if result.ResultCode != 100 || result.Data == nil {
		return nil, fmt.Errorf("获取 data-api 配置失败: %s", result.ResultDesc)
	}
	return result.Data, nil
}

// CallPost 调用已购买 API 的 POST 接口。
func (c *DataAPIGatewayClient) CallPost(ctx context.Context, apiCode string, request DataAPIPostRequest) (string, error) {
	return c.call(ctx, apiCode, dataAPIRequestMethodPost, request)
}

// CallGet 调用已购买 API 的 GET 接口。
func (c *DataAPIGatewayClient) CallGet(ctx context.Context, apiCode string, request DataAPIGetRequest) (string, error) {
	return c.call(ctx, apiCode, dataAPIRequestMethodGet, request)
}

func (c *DataAPIGatewayClient) call(ctx context.Context, apiCode, requestMethod string, request any) (string, error) {
	requestMethod, err := normalizeDataAPIRequestMethod(requestMethod)
	if err != nil {
		return "", err
	}
	apiCode = strings.TrimSpace(apiCode)
	if apiCode == "" {
		return "", fmt.Errorf("apiCode 不能为空")
	}

	config, err := c.FetchConfig(ctx, apiCode)
	if err != nil {
		return "", err
	}

	encryptedPayload, err := buildDataAPIRequestPayload(requestMethod, request, config.APIHash)
	if err != nil {
		return "", err
	}
	payloadJSON, err := marshalJSON(encryptedPayload)
	if err != nil {
		return "", fmt.Errorf("序列化加密前请求体失败: %w", err)
	}

	encryptedParamHex, err := EncryptDataAPIMessage(c.context.KeyPair.PublicKeyHex, []byte(payloadJSON), dataAPIRequestEncryptPrefix)
	if err != nil {
		return "", fmt.Errorf("加密卖家 API 请求参数失败: %w", err)
	}

	envelopeJSON, err := marshalJSON(dataAPIEncryptEnvelope{
		EncryptedParam: encryptedParamHex,
		ShuPublicKey:   c.context.KeyPair.PublicKeyHex,
	})
	if err != nil {
		return "", fmt.Errorf("序列化加密信封失败: %w", err)
	}

	requestHeaders, err := c.buildGatewayHeaders(config)
	if err != nil {
		return "", err
	}

	gatewayPath := dataAPIPostGatewayPath + apiCode
	if requestMethod == dataAPIRequestMethodGet {
		gatewayPath = dataAPIGetGatewayPath + apiCode
	}

	responseBody, err := c.postJSON(ctx, gatewayPath, requestHeaders, buildGatewayRequestBody(encodeHexString([]byte(envelopeJSON))))
	if err != nil {
		return "", err
	}
	return c.parseGatewayResponse(responseBody)
}

func (c *DataAPIGatewayClient) buildGatewayHeaders(config *DataAPIConfig) (map[string]string, error) {
	shuInfo, err := BuildDataAPIAuditInfo(
		c.context.KeyPair.PrivateKeyHex,
		c.context.KeyPair.PublicKeyHex,
		config.DianPublicKey,
		config.EnclaveHash,
	)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"Content-Type": "application/json",
		"appCode":      c.context.AppCode,
		"credential":   c.context.KeyPair.PublicKeyHex,
		"shuInfo":      shuInfo,
	}, nil
}

func (c *DataAPIGatewayClient) parseGatewayResponse(responseBody []byte) (string, error) {
	var result dataAPIResult[json.RawMessage]
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return "", fmt.Errorf("解析 data-api 网关响应失败: %w", err)
	}

	var encryptedResultEnvelope map[string]string
	if len(result.Data) > 0 {
		_ = json.Unmarshal(result.Data, &encryptedResultEnvelope)
	}

	encryptedResultHex, hasEncryptedResult := encryptedResultEnvelope["encrypted_result"]
	if hasEncryptedResult && encryptedResultHex != "" {
		plainTextBytes, err := DecryptDataAPIMessage(c.context.KeyPair.PrivateKeyHex, encryptedResultHex, dataAPIRequestEncryptPrefix)
		if err != nil {
			return "", fmt.Errorf("解密 data-api 响应失败: %w", err)
		}
		return string(plainTextBytes), nil
	}

	if result.ResultCode == 100 {
		return string(result.Data), nil
	}
	return "", fmt.Errorf("调用已购买 API 失败: %s", result.ResultDesc)
}

func buildDataAPIRequestPayload(requestMethod string, request any, apiHash string) (*DataAPIEncryptedPayload, error) {
	requestMethod, err := normalizeDataAPIRequestMethod(requestMethod)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(apiHash) == "" {
		return nil, fmt.Errorf("apiHash 不能为空")
	}

	payload := &DataAPIEncryptedPayload{
		APIHash:   apiHash,
		APIMethod: requestMethod,
		APIHeader: map[string]string{},
		APIBody:   map[string]string{},
		APIParam:  map[string]string{},
	}

	switch requestMethod {
	case dataAPIRequestMethodPost:
		postRequest, ok := request.(DataAPIPostRequest)
		if !ok {
			return nil, fmt.Errorf("POST 请求类型错误")
		}
		payload.APIHeader = convertDataAPIParamsToMap(postRequest.RequestHeaders)
		payload.APIBody = convertDataAPIParamsToMap(postRequest.BodyParams)
	case dataAPIRequestMethodGet:
		getRequest, ok := request.(DataAPIGetRequest)
		if !ok {
			return nil, fmt.Errorf("GET 请求类型错误")
		}
		payload.APIHeader = convertDataAPIParamsToMap(getRequest.RequestHeaders)
		payload.APIParam = convertDataAPIParamsToMap(getRequest.QueryParams)
	default:
		return nil, fmt.Errorf("不支持的请求方法: %s", requestMethod)
	}

	return payload, nil
}

func normalizeDataAPIRequestMethod(requestMethod string) (string, error) {
	trimmedMethod := strings.ToUpper(strings.TrimSpace(requestMethod))
	if trimmedMethod == "" {
		return "", fmt.Errorf("请求方法不能为空")
	}
	if trimmedMethod != dataAPIRequestMethodPost && trimmedMethod != dataAPIRequestMethodGet {
		return "", fmt.Errorf("不支持的请求方法: %s", trimmedMethod)
	}
	return trimmedMethod, nil
}

func convertDataAPIParamsToMap(params []DataAPIParam) map[string]string {
	result := make(map[string]string, len(params))
	for _, item := range params {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			continue
		}
		result[name] = item.Value
	}
	return result
}

func buildGatewayRequestBody(paramsHex string) dataAPIGatewayRequest {
	return dataAPIGatewayRequest{Params: paramsHex}
}

func encodeHexString(value []byte) string {
	return fmt.Sprintf("%x", value)
}

func marshalJSON(value any) (string, error) {
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func (c *DataAPIGatewayClient) postJSON(ctx context.Context, path string, headers map[string]string, body any) ([]byte, error) {
	requestJSON, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.context.BaseURL+path, bytes.NewReader(requestJSON))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		request.Header.Set(key, value)
	}

	response, err := c.context.HTTPClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("请求 data-api 网关失败: %w", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("读取网关响应失败: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("data-api 网关返回异常状态码 %d: %s", response.StatusCode, string(responseBody))
	}
	return responseBody, nil
}
