package main

import (
	"encoding/json"
	"strings"
	"testing"

	"dianshu-mcp/dianshu"
)

func TestBuildDataQueryRequestDefaults(t *testing.T) {
	request := buildDataQueryRequest(DataSearchArgs{Query: "西瓜"})
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

func TestBuildDataQueryRequestCustomKeyword(t *testing.T) {
	request := buildDataQueryRequest(DataSearchArgs{Query: "西瓜", Keyword: "大西瓜"})
	if request.Body["keyword"] != "大西瓜" {
		t.Fatalf("自定义 keyword 映射错误: got=%s", request.Body["keyword"])
	}
}

func TestPersistRawResult(t *testing.T) {
	t.Run("empty payload", func(t *testing.T) {
		_, err := persistRawResult(dianshu.DataQueryRequest{RawQuery: "西瓜"}, &dianshu.DataQueryResult{})
		if err == nil {
			t.Fatal("预期空响应报错，但未报错")
		}
	})

	t.Run("valid payload", func(t *testing.T) {
		const rawJSON = `{"resultCode":200,"resultDesc":"成功","data":{"hello":"world"},"DSSeqNo":"DS1"}`
		text, err := persistRawResult(dianshu.DataQueryRequest{
			ProviderType: dianshu.ProviderTypeXiaohongshu,
			DatasetType:  dianshu.DatasetTypeSearch,
			SiteDomain:   dianshu.XiaohongshuSiteDomain,
			Body:         map[string]string{"keyword": "西瓜", "page": "1"},
			RawQuery:     "西瓜",
		}, &dianshu.DataQueryResult{
			ResultCode: 200, ResultDesc: "成功", DSSeqNo: "DS1",
			RawJSON: json.RawMessage(rawJSON),
		})
		if err != nil {
			t.Fatalf("写入原始结果失败: %v", err)
		}
		if !strings.Contains(text, "结果文件:") {
			t.Fatalf("返回文案未包含结果文件路径: %s", text)
		}
		if !strings.Contains(text, rawJSON) {
			t.Fatalf("返回文案未包含原始 JSON 内容: %s", text)
		}
	})
}
