package dianshu

import (
	"strings"
	"testing"
)

func TestBuildDataAPIRequestPayloadForPost(t *testing.T) {
	payload, err := buildDataAPIRequestPayload(dataAPIRequestMethodPost, DataAPIPostRequest{
		BodyParams: []DataAPIParam{
			{Name: "keyword", Value: "香菇"},
			{Name: "page", Value: "1"},
		},
		RequestHeaders: []DataAPIParam{{Name: "X-Test", Value: "demo"}},
	}, "api-hash-value")
	if err != nil {
		t.Fatalf("构建 POST 请求参数失败: %v", err)
	}

	if payload.APIMethod != dataAPIRequestMethodPost {
		t.Fatalf("请求方法错误: got=%s want=%s", payload.APIMethod, dataAPIRequestMethodPost)
	}
	if payload.APIHash != "api-hash-value" {
		t.Fatalf("apiHash 错误: got=%s want=%s", payload.APIHash, "api-hash-value")
	}
	if payload.APIBody["keyword"] != "香菇" {
		t.Fatalf("body keyword 错误: got=%s", payload.APIBody["keyword"])
	}
	if payload.APIBody["page"] != "1" {
		t.Fatalf("body page 错误: got=%s", payload.APIBody["page"])
	}
	if payload.APIHeader["X-Test"] != "demo" {
		t.Fatalf("header 错误: got=%s", payload.APIHeader["X-Test"])
	}
	if len(payload.APIParam) != 0 {
		t.Fatalf("POST 请求不应包含 query 参数: %+v", payload.APIParam)
	}
}

func TestBuildDataAPIRequestPayloadForGet(t *testing.T) {
	payload, err := buildDataAPIRequestPayload(dataAPIRequestMethodGet, DataAPIGetRequest{
		QueryParams: []DataAPIParam{{Name: "startTime", Value: "1779381079"}},
	}, "api-hash-value")
	if err != nil {
		t.Fatalf("构建 GET 请求参数失败: %v", err)
	}
	if payload.APIParam["startTime"] != "1779381079" {
		t.Fatalf("query 参数错误: got=%s", payload.APIParam["startTime"])
	}
	if len(payload.APIBody) != 0 {
		t.Fatalf("GET 请求不应包含 body 参数: %+v", payload.APIBody)
	}
}

func TestBuildDataAPIRequestPayloadRejectsUnknownMethod(t *testing.T) {
	_, err := buildDataAPIRequestPayload("PUT", DataAPIPostRequest{}, "api-hash-value")
	if err == nil {
		t.Fatal("预期未知请求方法报错，但未报错")
	}
}

func TestEncodeHexStringReturnsLowercaseHex(t *testing.T) {
	encoded := encodeHexString([]byte("ABC"))
	if encoded != "414243" {
		t.Fatalf("hex 编码错误: got=%s want=%s", encoded, "414243")
	}
}

func TestBuildGatewayRequestBodyContainsParams(t *testing.T) {
	body := buildGatewayRequestBody("abcd")
	if body.Params != "abcd" {
		t.Fatalf("params 字段错误: got=%s want=%s", body.Params, "abcd")
	}
}

func TestValidateDataAPIRequestMethod(t *testing.T) {
	method, err := normalizeDataAPIRequestMethod("post")
	if err != nil {
		t.Fatalf("标准化方法失败: %v", err)
	}
	if method != dataAPIRequestMethodPost {
		t.Fatalf("方法标准化错误: got=%s want=%s", method, dataAPIRequestMethodPost)
	}

	_, err = normalizeDataAPIRequestMethod(strings.Repeat(" ", 1))
	if err == nil {
		t.Fatal("预期空方法报错，但未报错")
	}
}
