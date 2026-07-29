package pipeline

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"

	"dianshu-mcp/crypto"
)

var aadDebug = []byte("tech.yeez.key.manager")

// DebugDecrypt 直接用 crypto 包的内部函数手动推导 AES key 并解密。
func DebugDecrypt(encryptedKeyHex, priKeyHex string) {
	encBytes, _ := hex.DecodeString(encryptedKeyHex)
	cLen := len(encBytes) - 92
	ct := encBytes[:cLen]
	iv := encBytes[cLen : cLen+12]
	pk := encBytes[cLen+12 : cLen+12+64]
	tag := encBytes[cLen+12+64:]

	fmt.Printf("[debug] ===== 解密调试 =====\n")
	fmt.Printf("[debug] encrypted: %s\n", encryptedKeyHex)
	fmt.Printf("[debug] encLen=%d cLen=%d\n", len(encBytes), cLen)
	fmt.Printf("[debug] iv=%x\n", iv)
	fmt.Printf("[debug] pk=%x\n", pk)
	fmt.Printf("[debug] tag=%x\n", tag)

	// 用户公钥
	userPub, _ := crypto.PublicKeyFromPrivate(priKeyHex)
	fmt.Printf("[debug] userPri=%s\n", priKeyHex)
	fmt.Printf("[debug] userPub=%s\n", userPub)

	// 生成 ForwardSecretKey，验证 crypto 包自身没问题
	ots, fwdEnc, _, _ := crypto.GenerateForwardSecretKey(userPub, priKeyHex)
	fwdDec, fwdErr := crypto.DecryptForwardMessage(priKeyHex, fwdEnc)
	fmt.Printf("[debug] self-fwd: otsLen=%d encLen=%d decErr=%v\n", len(ots), len(fwdEnc), fwdErr)
	if fwdErr == nil {
		fmt.Printf("[debug] self-fwd: decLen=%d match=%v\n", len(fwdDec), bytes.Equal([]byte(ots), fwdDec))
	}

	// 用 ts 里的方式：DecryptInput (prefix=0x02)
	// 这是 decryptPrivateKey 在 JS 里用的
	dataEnc, _ := crypto.GenerateEncryptedInput(userPub, []byte("test-plaintext"))
	dataDec, dataErr := crypto.DecryptInput(priKeyHex, dataEnc)
	fmt.Printf("[debug] self-data: encLen=%d decErr=%v decPlain=%s\n", len(dataEnc), dataErr, string(dataDec))

	// 验证：用 crypto 包的 DeriveAESKey 看实际派生密钥
	otsPub := hex.EncodeToString(pk)
	derivedKey, deriveErr := crypto.DeriveAESKey(otsPub, priKeyHex)
	fmt.Printf("[debug] deriveKey=%x err=%v\n", derivedKey, deriveErr)

	// 试试反向：用 userPub 作为 pkey
	derivedKey2, _ := crypto.DeriveAESKey(userPub, priKeyHex)
	fmt.Printf("[debug] deriveKey(self)=%x\n", derivedKey2)
	// DecryptForwardMessage = prefix 0x01
	// DecryptInput = prefix 0x02 = decryptMessage in JS = decryptPrivateKey
	fmt.Printf("[debug] === 真实数据解密 ===\n")
	r1, e1 := crypto.DecryptForwardMessage(priKeyHex, encBytes) // prefix 0x01
	fmt.Printf("[debug] forward(0x01): err=%v\n", e1)
	r2, e2 := crypto.DecryptInput(priKeyHex, encBytes) // prefix 0x02 = decryptPrivateKey
	fmt.Printf("[debug] input(0x02):   err=%v\n", e2)

	if e1 == nil {
		fmt.Printf("[debug] forward success: dec=%x\n", r1)
	}
	if e2 == nil {
		fmt.Printf("[debug] input success: dec=%x\n", r2)
	}

	// 再试试错误信息 - 用错误 prefix 应该会有不同的错误信息
	fmt.Printf("[debug] === GCM 手动验证 ===\n")
	block, _ := aes.NewCipher(make([]byte, 16)) // wrong key
	gcm, _ := cipher.NewGCM(block)
	tad01 := make([]byte, 64)
	copy(tad01, aadDebug)
	tad01[24] = 0x01
	sealed := make([]byte, cLen+16)
	copy(sealed, ct)
	copy(sealed[cLen:], tag)
	_, wrongKeyErr := gcm.Open(nil, iv, sealed, tad01)
	fmt.Printf("[debug] wrong-key GCM: err=%v\n", wrongKeyErr)
}
