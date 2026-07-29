package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"testing"
)

// 测试私钥/公钥，与 meta-encryptor test/helper.js 中的 key_pair 一致
const testPrivateKey = "60d61a1d92b26608016dba8cb8e8e96fd44d5dee0a0415a024657e47febcced8"
const testPublicKey = "731234931a081e9beae856318a9bf32ac3698ea8215bf74f517f8377cc6ba8740e28ed87c97d0ee8775bc83505867b0bc34a66adc91f0ea9b44c80533f1a3dca"

func TestHexToBytesToHex(t *testing.T) {
	original := testPrivateKey
	b, err := hexToBytes(original)
	if err != nil {
		t.Fatalf("hexToBytes 失败: %v", err)
	}
	hexStr := toHex(b)
	if hexStr != original {
		t.Fatalf("hex 往返失败: got=%s want=%s", hexStr, original)
	}
}

func TestPublicKeyFromPrivate(t *testing.T) {
	pubKey, err := PublicKeyFromPrivate(testPrivateKey)
	if err != nil {
		t.Fatalf("生成公钥失败: %v", err)
	}
	if pubKey != testPublicKey {
		t.Fatalf("公钥不匹配:\ngot:  %s\nwant: %s", pubKey, testPublicKey)
	}
}

func TestGenerateAESKeyFrom(t *testing.T) {
	key, err := DeriveAESKey(testPublicKey, testPrivateKey)
	if err != nil {
		t.Fatalf("派生 AES 密钥失败: %v", err)
	}
	if len(key) != 16 {
		t.Fatalf("AES 密钥长度错误: got=%d want=16", len(key))
	}
	t.Logf("AES derived key: %s", hex.EncodeToString(key))
}

func TestAESCmac(t *testing.T) {
	result, err := aesCmac(cmacKeyBytes, []byte("hello"))
	if err != nil {
		t.Fatalf("aesCmac 'hello' 失败: %v", err)
	}
	if len(result) != 16 {
		t.Fatalf("aesCmac 结果长度错误: %d", len(result))
	}
	result2, _ := aesCmac(cmacKeyBytes, []byte("hello"))
	if !bytes.Equal(result, result2) {
		t.Fatal("aesCmac 同输入不可复现")
	}
	result3, _ := aesCmac(cmacKeyBytes, []byte("world"))
	if bytes.Equal(result, result3) {
		t.Fatal("aesCmac 不同输入应产生不同结果")
	}
}

func TestEncryptDecryptRoundtrip(t *testing.T) {
	key, _ := DeriveAESKey(testPublicKey, testPrivateKey)
	plaintext := []byte("hello world test data")

	pubBytes, _ := hexToBytes(testPublicKey)
	cipherPkg, err := encryptMessage(testPublicKey, testPrivateKey, plaintext, pubBytes, 0x02)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	cipherLen := len(cipherPkg) - 64 - 16 - 12
	ct := cipherPkg[:cipherLen]
	iv := cipherPkg[cipherLen : cipherLen+12]
	pk := cipherPkg[cipherLen+12 : cipherLen+12+64]
	tag := cipherPkg[cipherLen+12+64:]
	t.Logf("cipherPkg=%d ct=%d iv=%x pk_first8=%x tag_first8=%x", len(cipherPkg), len(ct), iv, pk[:8], tag[:8])

	block, _ := aes.NewCipher(key[:16])
	gcm, _ := cipher.NewGCM(block)
	tad := make([]byte, 64)
	copy(tad, aad)
	tad[24] = 0x02

	sealed := make([]byte, len(ct)+len(tag))
	copy(sealed, ct)
	copy(sealed[len(ct):], tag)

	result, err := gcm.Open(nil, iv, sealed, tad)
	if err != nil {
		t.Fatalf("GCM Open 失败: %v", err)
	}
	if !bytes.Equal(result, plaintext) {
		t.Fatalf("内容不一致")
	}

	decrypted, err := decryptMessage(testPrivateKey, cipherPkg, 0x02)
	if err != nil {
		t.Fatalf("decryptMessage 失败: %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Fatalf("decryptMessage 内容不一致: %x vs %x", decrypted, plaintext)
	}
}

func TestEncryptDecryptRoundtrip2(t *testing.T) {
	otsHex, _ := generatePrivateKey()
	pub, _ := PublicKeyFromPrivate(otsHex)
	pubBytes, _ := hexToBytes(pub)

	remotePubBytes, _ := hexToBytes(testPublicKey)
	plaintext := []byte("test message")

	aesKeyEnc, _ := DeriveAESKey(testPublicKey, otsHex)
	t.Logf("enc key=%x plain=%x pub=%x", aesKeyEnc, plaintext, pubBytes)

	cipherPkg, err := encryptMessage(testPublicKey, otsHex, plaintext, pubBytes, 0x02)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	cipherLen := len(cipherPkg) - 64 - 16 - 12
	extractedPub := cipherPkg[cipherLen+12 : cipherLen+12+64]
	t.Logf("cipherLen=%d extractedPub=%x expected=%x", cipherLen, extractedPub, remotePubBytes)

	aesKeyDec, _ := DeriveAESKey(toHex(extractedPub), testPrivateKey)
	t.Logf("dec key=%x enc key=%x", aesKeyDec, aesKeyEnc)

	if !bytes.Equal(aesKeyEnc, aesKeyDec) {
		t.Fatalf("加密/解密密钥不一致!")
	}

	decrypted, err := decryptMessage(testPrivateKey, cipherPkg, 0x02)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Fatalf("往返失败")
	}
}

func TestForwardSecretKeyRoundtrip(t *testing.T) {
	otsHex, encryptedSkey, _, err := GenerateForwardSecretKey(testPublicKey, testPrivateKey)
	if err != nil {
		t.Fatalf("生成转发密钥失败: %v", err)
	}
	if len(otsHex) != 64 {
		t.Fatalf("OTS 私钥长度异常: %d", len(otsHex))
	}
	decrypted, err := DecryptForwardMessage(testPrivateKey, encryptedSkey)
	if err != nil {
		t.Fatalf("解密转发消息失败: %v", err)
	}
	otsExpectedBytes, _ := hexToBytes(otsHex)
	if !bytes.Equal(decrypted, otsExpectedBytes) {
		t.Fatal("转发密钥解密后与原始 OTS 不一致")
	}
}

func TestSignMessage(t *testing.T) {
	message := []byte("dianshu test message")
	sig, err := SignMessage(testPrivateKey, message)
	if err != nil {
		t.Fatalf("签名失败: %v", err)
	}
	if len(sig) != 65 {
		t.Fatalf("签名长度错误: got=%d want=65", len(sig))
	}
	if sig[64] != 0 && sig[64] != 1 && sig[64] != 27 && sig[64] != 28 {
		t.Fatalf("签名 v 值异常: %d", sig[64])
	}
}

func TestAESCmacKnownVectorEmpty(t *testing.T) {
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i)
	}
	result, err := aesCmac(key, []byte{})
	if err != nil {
		t.Fatalf("aesCmac 失败: %v", err)
	}
	if len(result) != 16 {
		t.Fatalf("aesCmac 结果长度错误: %d", len(result))
	}
	t.Logf("AES-CMAC empty message: %s", hex.EncodeToString(result))
}
