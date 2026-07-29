package chain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	initTxPath    = "/chain/transaction/initDataTransaction"
	sendTxPath    = "/chain/transaction/sendDataTransaction"
	checkTaskPath = "/task/check/task"
)

// Client 链操作客户端，通过典枢后端 API 完成链上操作。
type Client struct {
	baseURL string
	token   string
	client  *http.Client
}

// NewClient 创建链客户端
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		client:  &http.Client{},
	}
}

// InitTxResponse initDataTransaction 响应
type InitTxResponse struct {
	FromAddress        string `json:"fromAddress"`
	Nonce              int    `json:"nonce"`
	GasPrice           int    `json:"gasPrice"`
	GasLimit           int    `json:"gasLimit"`
	ContractAddress    string `json:"contractAddress"`
	ChainID            int    `json:"chainId"`
	FunctionName       string `json:"functionName"`
	DataOnChainHash    string `json:"dataOnChainHash"`
	RequestOnChainHash string `json:"requestOnChainHash"`
	ResultOnChainHash  string `json:"resultOnChainHash"`
}

// InitOffChainSkey 初始化 requestOffChainSkey 交易。
func (c *Client) InitOffChainSkey(ctx context.Context, orderCode string) (*InitTxResponse, error) {
	body := map[string]interface{}{
		"functionName": "requestOffChainSkey",
		"orderCode":    orderCode,
	}
	resp, err := c.doPost(ctx, initTxPath, body)
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		ResultCode int             `json:"resultCode"`
		Data       *InitTxResponse `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("解析 initTx 响应失败: %w", err)
	}
	if wrapper.ResultCode != 100 || wrapper.Data == nil {
		return nil, fmt.Errorf("initTx 失败: %s", string(resp))
	}
	return wrapper.Data, nil
}

// SendTxRequest sendDataTransaction 请求体
type SendTxRequest struct {
	UUID                string `json:"UUID"`
	SignedTransactionTx string `json:"signedTransactionTx"`
}

// SendTxResponse sendDataTransaction 响应
type SendTxResponse struct {
	TransactionHash string `json:"transactionHash"`
	UUID            string `json:"UUID"`
}

// SendTransaction 发送已签名的交易。
func (c *Client) SendTransaction(ctx context.Context, req SendTxRequest) (*SendTxResponse, error) {
	body := map[string]interface{}{
		"UUID":                req.UUID,
		"signedTransactionTx": req.SignedTransactionTx,
		"functionName":        "requestOffChainSkey",
	}
	resp, err := c.doPost(ctx, sendTxPath, body)
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		ResultCode int             `json:"resultCode"`
		Data       *SendTxResponse `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return nil, fmt.Errorf("解析 sendTx 响应失败: %w", err)
	}
	if wrapper.ResultCode != 100 || wrapper.Data == nil {
		return nil, fmt.Errorf("发送交易失败: %s", string(resp))
	}
	return wrapper.Data, nil
}

// CheckTask 检查任务链上状态（0/1 待处理，2 成功，3 失败）。
func (c *Client) CheckTask(ctx context.Context, orderCode string) (int, error) {
	body := map[string]string{"orderCode": orderCode}
	resp, err := c.doPost(ctx, checkTaskPath, body)
	if err != nil {
		return 0, err
	}
	var wrapper struct {
		Data struct {
			PublishStatus int `json:"publishStatus"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		return 0, fmt.Errorf("解析 checkTask 响应失败: %w", err)
	}
	return wrapper.Data.PublishStatus, nil
}

func (c *Client) doPost(ctx context.Context, path string, body interface{}) ([]byte, error) {
	jsonBody, _ := json.Marshal(body)
	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("token", c.token)
	req.Header.Set("Referer", "app://.")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(data))
	}
	return data, nil
}
