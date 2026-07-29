package dianshu

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"dianshu-mcp/crypto"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

const baseURL = "https://data-api.dianshudata.com"

// ---------- 密钥生成 ----------

func newKeypair() (privHex, pubHex string) {
	k, _ := secp256k1.GeneratePrivateKey()
	privHex = fmt.Sprintf("%064x", k.Serialize())
	pub := k.PubKey()
	pubHex = fmt.Sprintf("%064x%064x", pub.X().Bytes(), pub.Y().Bytes())
	return
}

// ---------- SDK 客户端 ----------

type Client struct {
	appCode   string
	uniqueID  string
	localPriv string
	localPub  string
	apiHash   string
	dianPkey  string
	enclave   string
}

func NewClient(appCode, uniqueAPIID string) (*Client, error) {
	c := &Client{appCode: appCode, uniqueID: uniqueAPIID}
	c.localPriv, c.localPub = newKeypair()

	body, _ := json.Marshal(map[string]string{"appCode": appCode, "apiId": uniqueAPIID})
	resp, err := http.Post(baseURL+"/api/privateKey", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var r struct {
		ResultCode int    `json:"resultCode"`
		ResultDesc string `json:"resultDesc"`
		Data       struct {
			APIHash     string `json:"apiHash"`
			DianPkey    string `json:"dianPkey"`
			EnclaveHash string `json:"enclaveHash"`
		} `json:"data"`
	}
	json.Unmarshal(raw, &r)
	if r.ResultCode != 100 || r.Data.APIHash == "" {
		return nil, fmt.Errorf("获取 API 配置失败: %s", string(raw))
	}
	c.apiHash = r.Data.APIHash
	c.dianPkey = r.Data.DianPkey
	c.enclave = r.Data.EnclaveHash
	return c, nil
}

func (c *Client) genShuInfo() string {
	dataHash := c.dianPkey + c.enclave
	_, enc, _, _ := crypto.GenerateForwardSecretKey(c.dianPkey, c.localPriv)

	// 签名的数据是原始字节（dianPkey || enclaveHash），不是 hex 字符串
	dianBytes, _ := hex.DecodeString(c.dianPkey)
	encBytes, _ := hex.DecodeString(c.enclave)
	sigData := append(dianBytes, encBytes...)
	sig, _ := crypto.SignMessage(c.localPriv, sigData)
	m := map[string]string{
		"dataHash":                dataHash,
		"dataShuPublicKey":        c.localPub,
		"encryptedShuPrivateKey":  hex.EncodeToString(enc),
		"shuKeyForwardSignature":  hex.EncodeToString(sig),
		"allowedEnclaveHash":      c.enclave,
	}
	b, _ := json.Marshal(m)
	return string(b)
}

func (c *Client) encryptParams(method string, queryParams, bodyParams map[string]string) string {
	p := map[string]interface{}{
		"api_hash":   c.apiHash,
		"api_method": method,
	}
	if len(queryParams) > 0 {
		p["api_param"] = queryParams
	}
	if len(bodyParams) > 0 {
		p["api_body"] = bodyParams
	}
	plain, _ := json.Marshal(p)
	enc, _ := crypto.GenerateEncryptedInput(c.localPub, plain)
	inner, _ := json.Marshal(map[string]string{
		"encrypted_param": hex.EncodeToString(enc),
		"shu_pkey":        c.localPub,
	})
	return hex.EncodeToString(inner)
}

func (c *Client) Post(bodyParams map[string]string) (string, error) {
	paramsHex := c.encryptParams("POST", nil, bodyParams)
	shuInfo := c.genShuInfo()
	return c.doHTTP("POST", "/api/post/"+c.uniqueID, paramsHex, shuInfo)
}

func (c *Client) Get(queryParams map[string]string) (string, error) {
	paramsHex := c.encryptParams("GET", queryParams, nil)
	shuInfo := c.genShuInfo()
	return c.doHTTP("GET", "/api/get/"+c.uniqueID+"?"+paramsHex, paramsHex, shuInfo)
}

func (c *Client) doHTTP(method, path, paramsHex, shuInfo string) (string, error) {
	url := baseURL + path
	var reqBody io.Reader
	if method == "POST" {
		b, _ := json.Marshal(map[string]string{"params": paramsHex})
		reqBody = bytes.NewReader(b)
	}

	req, _ := http.NewRequest(method, url, reqBody)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("appCode", c.appCode)
	req.Header.Set("credential", c.localPub)
	req.Header.Set("shuInfo", shuInfo)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	var result struct {
		ResultCode int               `json:"resultCode"`
		ResultDesc string            `json:"resultDesc"`
		Data       map[string]string `json:"data"`
	}
	json.Unmarshal(raw, &result)
	if result.ResultCode != 100 && result.ResultCode != 200 {
		return "", fmt.Errorf("API 返回异常: %s", string(raw))
	}
	if result.Data != nil {
		if encResult := result.Data["encrypted_result"]; encResult != "" {
			encBytes, _ := hex.DecodeString(encResult)
			dec, err := crypto.DecryptInput(c.localPriv, encBytes)
			if err != nil {
				return "", fmt.Errorf("解密响应失败: %w", err)
			}
			return string(dec), nil
		}
	}
	return string(raw), nil
}

// ---------- 辅助 API（供 MCP 工具使用）----------

func GetAPIDetail(ctx context.Context, httpClient *http.Client, apiID int, token string, cookies []string) (*APIDetail, error) {
	body, _ := json.Marshal(map[string]interface{}{"apiId": apiID, "deleted": 0})
	req, _ := http.NewRequestWithContext(ctx, "POST", baseURL+"/api/detail", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("token", token)
	if len(cookies) > 0 {
		req.Header.Set("Cookie", strings.Join(cookies, "; "))
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result struct {
		ResultCode int        `json:"resultCode"`
		Data       *APIDetail `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if result.ResultCode != 100 || result.Data == nil {
		return nil, fmt.Errorf("查询失败: resultCode=%d", result.ResultCode)
	}
	return result.Data, nil
}

func GetBuyerAPIList(ctx context.Context, httpClient *http.Client, token string, cookies []string, pageNo, pageSize int) ([]BuyerAPIItem, error) {
	body, _ := json.Marshal(map[string]interface{}{"pageNo": pageNo, "pageSize": pageSize})
	req, _ := http.NewRequestWithContext(ctx, "POST", baseURL+"/api/getBuyerApi", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("token", token)
	if len(cookies) > 0 {
		req.Header.Set("Cookie", strings.Join(cookies, "; "))
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result struct {
		ResultCode int             `json:"resultCode"`
		Data       []BuyerAPIItem  `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if result.ResultCode != 100 {
		return nil, fmt.Errorf("查询失败: resultCode=%d", result.ResultCode)
	}
	return result.Data, nil
}

type BuyerAPIItem struct {
	APIID       int    `json:"apiId"`
	UniqueAPIID string `json:"uniqueApiId"`
	APICode     string `json:"apiCode"`
	APIName     string `json:"apiName"`
	Usage       string `json:"usage"`
}
