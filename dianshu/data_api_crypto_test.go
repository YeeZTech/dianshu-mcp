package dianshu

import (
	"encoding/hex"
	"testing"
)

func TestBuildAdditionalAuthenticatedData(t *testing.T) {
	aad := buildAdditionalAuthenticatedData(dataAPIRequestEncryptPrefix)
	if len(aad) != 64 {
		t.Fatalf("AAD 长度错误: got=%d want=64", len(aad))
	}
	if aad[24] != dataAPIRequestEncryptPrefix {
		t.Fatalf("AAD prefix 错误: got=%x want=%x", aad[24], dataAPIRequestEncryptPrefix)
	}

	prefixText := string(aad[:len(dataAPIAdditionalAuthenticatedDataSeed)])
	if prefixText != dataAPIAdditionalAuthenticatedDataSeed {
		t.Fatalf("AAD 种子错误: got=%q want=%q", prefixText, dataAPIAdditionalAuthenticatedDataSeed)
	}
}

func TestBuildEncryptedPayloadLayout(t *testing.T) {
	cipherText := []byte{0x01, 0x02, 0x03}
	iv := []byte{0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b}
	tempPublicKey := make([]byte, 64)
	for i := range tempPublicKey {
		tempPublicKey[i] = byte(i)
	}
	tag := []byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a}

	payload := buildEncryptedPayload(cipherText, iv, tempPublicKey, tag)
	wantLength := len(cipherText) + len(iv) + len(tempPublicKey) + len(tag)
	if len(payload) != wantLength {
		t.Fatalf("密文长度错误: got=%d want=%d", len(payload), wantLength)
	}

	cipherPart, ivPart, publicKeyPart, tagPart, err := splitEncryptedPayload(payload)
	if err != nil {
		t.Fatalf("拆解密文失败: %v", err)
	}
	assertBytesEqual(t, "cipherText", cipherPart, cipherText)
	assertBytesEqual(t, "iv", ivPart, iv)
	assertBytesEqual(t, "tempPublicKey", publicKeyPart, tempPublicKey)
	assertBytesEqual(t, "tag", tagPart, tag)
}

func TestSplitEncryptedPayloadRejectsShortCipherText(t *testing.T) {
	_, _, _, _, err := splitEncryptedPayload([]byte{0x01, 0x02, 0x03})
	if err == nil {
		t.Fatal("预期短密文报错，但未报错")
	}
}

func TestNormalizePrivateKeyBytes(t *testing.T) {
	inputHex := "01"
	inputBytes, err := hex.DecodeString(inputHex)
	if err != nil {
		t.Fatalf("解码输入失败: %v", err)
	}

	normalized, err := normalizePrivateKeyBytes(inputBytes)
	if err != nil {
		t.Fatalf("标准化私钥失败: %v", err)
	}
	if len(normalized) != 32 {
		t.Fatalf("私钥长度错误: got=%d want=32", len(normalized))
	}
	if normalized[31] != 0x01 {
		t.Fatalf("私钥末尾字节错误: got=%x want=%x", normalized[31], byte(0x01))
	}
}

func TestConvertRawPublicKeyToUncompressed(t *testing.T) {
	rawPublicKey := make([]byte, 64)
	for i := range rawPublicKey {
		rawPublicKey[i] = byte(i + 1)
	}

	uncompressed, err := convertRawPublicKeyToUncompressed(rawPublicKey)
	if err != nil {
		t.Fatalf("转换公钥失败: %v", err)
	}
	if len(uncompressed) != 65 {
		t.Fatalf("未压缩公钥长度错误: got=%d want=65", len(uncompressed))
	}
	if uncompressed[0] != 0x04 {
		t.Fatalf("未压缩公钥前缀错误: got=%x want=%x", uncompressed[0], byte(0x04))
	}
	assertBytesEqual(t, "未压缩公钥主体", uncompressed[1:], rawPublicKey)
}

func assertBytesEqual(t *testing.T, field string, got, want []byte) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s 长度错误: got=%d want=%d", field, len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("%s 第 %d 字节错误: got=%x want=%x", field, i, got[i], want[i])
		}
	}
}
