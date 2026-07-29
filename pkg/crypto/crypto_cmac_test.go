// Author: zhyyao
package crypto

import (
	"encoding/hex"
	"testing"
)

// RFC 4493 Section 4 — AES-CMAC 测试向量
func TestAESCmacRFC4493(t *testing.T) {
	// 128-bit key
	key, _ := hex.DecodeString("2b7e151628aed2a6abf7158809cf4f3c")

	cases := []struct {
		name     string
		message  string
		expected string
	}{
		{"empty", "", "bb1d6929e95937287fa37d129b756746"},
		{"16 bytes", "6bc1bee22e409f96e93d7e117393172a", "070a16b46b4d4144f79bdd9dd04a287c"},
		{"40 bytes", "6bc1bee22e409f96e93d7e117393172aae2d8a571e03ac9c9eb76fac45af8e5130c81c46a35ce411", "dfa66747de9ae63030ca32611497c827"},
		{"64 bytes", "6bc1bee22e409f96e93d7e117393172aae2d8a571e03ac9c9eb76fac45af8e5130c81c46a35ce411e5fbc1191a0a52eff69f2445df4f9b17ad2b417be66c3710", "51f0bebf7e3b9d92fc49741779363cfe"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg, _ := hex.DecodeString(tc.message)
			result, err := aesCmac(key, msg)
			if err != nil {
				t.Fatalf("aesCmac 失败: %v", err)
			}
			got := hex.EncodeToString(result)
			if got != tc.expected {
				t.Fatalf("AES-CMAC 不匹配:\ngot:  %s\nwant: %s", got, tc.expected)
			}
		})
	}
}
