package dianshu

import "testing"

func TestNormalizeProviderAndDatasetType(t *testing.T) {
	if got := normalizeProviderType(" Xiaohongshu "); got != ProviderTypeXiaohongshu {
		t.Fatalf("providerType 标准化错误: got=%s", got)
	}
	if got := normalizeDatasetType(" Search "); got != DatasetTypeSearch {
		t.Fatalf("datasetType 标准化错误: got=%s", got)
	}
}
