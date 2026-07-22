package main

import (
	"encoding/json"
	"strings"
	"testing"

	"dianshu-mcp/dianshu"
)

func TestBuildDataQueryRequestUsesDefaults(t *testing.T) {
	service := NewDianshuService()
	request, err := service.buildDataQueryRequest(DataSearchArgs{Query: "西瓜"})
	if err != nil {
		t.Fatalf("构建查询请求失败: %v", err)
	}
	if request.ProviderType != dianshu.ProviderTypeXiaohongshu {
		t.Fatalf("providerType 默认值错误: got=%s", request.ProviderType)
	}
	if request.DatasetType != dianshu.DatasetTypeSearch {
		t.Fatalf("datasetType 默认值错误: got=%s", request.DatasetType)
	}
	if request.SiteDomain != dianshu.XiaohongshuSiteDomain {
		t.Fatalf("siteDomain 默认值错误: got=%s", request.SiteDomain)
	}
	if request.Body["keyword"] != "西瓜" {
		t.Fatalf("keyword 映射错误: got=%s", request.Body["keyword"])
	}
	if request.Body["page"] != dianshu.XiaohongshuDefaultPage {
		t.Fatalf("page 默认值错误: got=%s", request.Body["page"])
	}
	if strings.TrimSpace(request.Body["startTime"]) == "" {
		t.Fatal("startTime 未自动填充")
	}
	if strings.TrimSpace(request.Body["endTime"]) == "" {
		t.Fatal("endTime 未自动填充")
	}
}

func TestPersistRawResultRejectsEmptyPayload(t *testing.T) {
	service := NewDianshuService()
	_, err := service.persistRawResult(dianshu.DataQueryRequest{RawQuery: "西瓜"}, &dianshu.DataQueryResult{})
	if err == nil {
		t.Fatal("预期空响应报错，但未报错")
	}
}

func TestPersistRawResultKeepsOriginalJSON(t *testing.T) {
	service := NewDianshuService()
	raw := json.RawMessage(`{"resultCode":200,"resultDesc":"成功","data":{"hello":"world"},"DSSeqNo":"DS1"}`)
	text, err := service.persistRawResult(dianshu.DataQueryRequest{
		ProviderType: dianshu.ProviderTypeXiaohongshu,
		DatasetType:  dianshu.DatasetTypeSearch,
		SiteDomain:   dianshu.XiaohongshuSiteDomain,
		Body:         map[string]string{"keyword": "西瓜", "page": "1"},
		RawQuery:     "西瓜",
	}, &dianshu.DataQueryResult{ResultCode: 200, ResultDesc: "成功", DSSeqNo: "DS1", RawJSON: raw})
	if err != nil {
		t.Fatalf("写入原始结果失败: %v", err)
	}
	if !strings.Contains(text, "结果文件:") {
		t.Fatalf("返回文案未包含结果文件路径: %s", text)
	}
}
