package dianshu

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/aead/cmac"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	dataAPIAESKeySizeBytes                       = 16
	dataAPIGCMIVLengthBytes                      = 12
	dataAPIGCMTagLengthBytes                     = 16
	dataAPIEphemeralPublicKeyLengthBytes         = 64
	dataAPIUncompressedPublicKeyLengthBytes      = 65
	dataAPIPrivateKeyLengthBytes                 = 32
	dataAPIAdditionalAuthenticatedDataSize       = 64
	dataAPIRequestEncryptPrefix             byte = 0x02
	dataAPIAuditEncryptPrefix               byte = 0x01
	dataAPIAdditionalAuthenticatedDataSeed       = "tech.yeez.key.manager"
	dataAPICmacSeedHex                           = "7965657a2e746563682e7374626f7800"
)

var dataAPIAdditionalAuthenticatedDataSeedBytes = []byte(dataAPIAdditionalAuthenticatedDataSeed)

// DataAPIKeyPair 表示一次 API 调用使用的临时密钥对。
type DataAPIKeyPair struct {
	PublicKeyHex  string
	PrivateKeyHex string
}

// GenerateDataAPIKeyPair 生成与 Java SDK 一致格式的 secp256k1 临时密钥对。
func GenerateDataAPIKeyPair() (*DataAPIKeyPair, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("生成 secp256k1 私钥失败: %w", err)
	}

	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyBytes, err = normalizePrivateKeyBytes(privateKeyBytes)
	if err != nil {
		return nil, err
	}

	publicKeyBytes := crypto.FromECDSAPub(&privateKey.PublicKey)
	if len(publicKeyBytes) != dataAPIUncompressedPublicKeyLengthBytes {
		return nil, fmt.Errorf("生成的未压缩公钥长度错误: %d", len(publicKeyBytes))
	}

	return &DataAPIKeyPair{
		PublicKeyHex:  hex.EncodeToString(publicKeyBytes[1:]),
		PrivateKeyHex: hex.EncodeToString(privateKeyBytes),
	}, nil
}

// EncryptDataAPIMessage 使用 data-api 协议对明文内容加密。
func EncryptDataAPIMessage(publicKeyHex string, plainText []byte, prefix byte) (string, error) {
	publicKeyBytes, err := decodeHexString(publicKeyHex)
	if err != nil {
		return "", fmt.Errorf("解析公钥失败: %w", err)
	}

	cipherText, err := encryptDataAPIMessageBytes(publicKeyBytes, plainText, prefix)
	if err != nil {
		return "", err
	}
	return encodeHexString(cipherText), nil
}

// DecryptDataAPIMessage 使用 data-api 协议解密密文内容。
func DecryptDataAPIMessage(privateKeyHex, cipherHex string, prefix byte) ([]byte, error) {
	privateKeyBytes, err := decodeHexString(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("解析私钥失败: %w", err)
	}
	cipherBytes, err := decodeHexString(cipherHex)
	if err != nil {
		return nil, fmt.Errorf("解析密文失败: %w", err)
	}
	return decryptDataAPIMessageBytes(cipherBytes, privateKeyBytes, prefix)
}

// BuildDataAPIAuditInfo 构造 data-api 请求头中的 shuInfo 字段。
func BuildDataAPIAuditInfo(privateKeyHex, publicKeyHex, dianPublicKeyHex, enclaveHash string) (string, error) {
	privateKeyBytes, err := decodeHexString(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("解析临时私钥失败: %w", err)
	}
	if _, err = decodeHexString(publicKeyHex); err != nil {
		return "", fmt.Errorf("解析临时公钥失败: %w", err)
	}
	if _, err = decodeHexString(dianPublicKeyHex); err != nil {
		return "", fmt.Errorf("解析典公钥失败: %w", err)
	}
	if _, err = decodeHexString(enclaveHash); err != nil {
		return "", fmt.Errorf("解析 enclaveHash 失败: %w", err)
	}

	dataHash := dianPublicKeyHex + enclaveHash
	encryptedPrivateKeyHex, err := EncryptDataAPIMessage(dianPublicKeyHex, privateKeyBytes, dataAPIAuditEncryptPrefix)
	if err != nil {
		return "", fmt.Errorf("加密临时私钥失败: %w", err)
	}

	signatureHex, err := signDataAPIAuditMessage(privateKeyBytes, dataHash)
	if err != nil {
		return "", fmt.Errorf("生成 shuInfo 签名失败: %w", err)
	}

	auditInfo := dataAPIAuditInfo{
		DataHash:               dataHash,
		DataShuPublicKey:       publicKeyHex,
		AllowedEnclaveHash:     enclaveHash,
		EncryptedShuPrivateKey: encryptedPrivateKeyHex,
		ShuKeyForwardSignature: signatureHex,
	}
	return marshalJSON(auditInfo)
}

func encryptDataAPIMessageBytes(publicKeyBytes, plainText []byte, prefix byte) ([]byte, error) {
	if len(plainText) == 0 {
		return nil, errors.New("明文内容不能为空")
	}

	uncompressedPublicKey, err := convertRawPublicKeyToUncompressed(publicKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("转换目标公钥失败: %w", err)
	}

	ephemeralKeyPair, err := GenerateDataAPIKeyPair()
	if err != nil {
		return nil, err
	}
	ephemeralPrivateKeyBytes, err := decodeHexString(ephemeralKeyPair.PrivateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("解析临时私钥失败: %w", err)
	}
	ephemeralPublicKeyBytes, err := decodeHexString(ephemeralKeyPair.PublicKeyHex)
	if err != nil {
		return nil, fmt.Errorf("解析临时公钥失败: %w", err)
	}

	derivedKey, err := deriveDataAPISharedKey(uncompressedPublicKey, ephemeralPrivateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("派生加密密钥失败: %w", err)
	}

	aad := buildAdditionalAuthenticatedData(prefix)
	cipherText, tag, iv, err := encryptAESGCM(plainText, derivedKey, aad)
	if err != nil {
		return nil, fmt.Errorf("AES-GCM 加密失败: %w", err)
	}
	return buildEncryptedPayload(cipherText, iv, ephemeralPublicKeyBytes, tag), nil
}

func decryptDataAPIMessageBytes(cipherText, privateKeyBytes []byte, prefix byte) ([]byte, error) {
	privateKeyBytes, err := normalizePrivateKeyBytes(privateKeyBytes)
	if err != nil {
		return nil, err
	}

	cipherBody, iv, ephemeralPublicKey, tag, err := splitEncryptedPayload(cipherText)
	if err != nil {
		return nil, err
	}

	uncompressedEphemeralPublicKey, err := convertRawPublicKeyToUncompressed(ephemeralPublicKey)
	if err != nil {
		return nil, fmt.Errorf("转换临时公钥失败: %w", err)
	}

	derivedKey, err := deriveDataAPISharedKey(uncompressedEphemeralPublicKey, privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("派生解密密钥失败: %w", err)
	}

	aad := buildAdditionalAuthenticatedData(prefix)
	return decryptAESGCM(cipherBody, tag, iv, derivedKey, aad)
}

func deriveDataAPISharedKey(uncompressedPublicKey, privateKeyBytes []byte) ([]byte, error) {
	privateKeyECDSA, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("构造 ECDSA 私钥失败: %w", err)
	}
	publicKeyECDSA, err := crypto.UnmarshalPubkey(uncompressedPublicKey)
	if err != nil {
		return nil, fmt.Errorf("构造 ECDSA 公钥失败: %w", err)
	}

	sharedX, sharedY := publicKeyECDSA.Curve.ScalarMult(publicKeyECDSA.X, publicKeyECDSA.Y, privateKeyECDSA.D.Bytes())
	if sharedX == nil || sharedY == nil {
		return nil, errors.New("ECDH 共享密钥计算失败")
	}

	sharedSecret := normalizeCoordinateBytes(sharedX)
	firstStageKey, err := generateAESCMAC(decodeMust(dataAPICmacSeedHex), sharedSecret)
	if err != nil {
		return nil, fmt.Errorf("第一阶段 CMAC 派生失败: %w", err)
	}

	deviationBuffer := make([]byte, len(dataAPIAdditionalAuthenticatedDataSeedBytes)+4)
	deviationBuffer[0] = 0x01
	copy(deviationBuffer[1:], dataAPIAdditionalAuthenticatedDataSeedBytes)
	deviationBuffer[len(dataAPIAdditionalAuthenticatedDataSeedBytes)+1] = 0x00
	deviationBuffer[len(dataAPIAdditionalAuthenticatedDataSeedBytes)+2] = 0x80
	deviationBuffer[len(dataAPIAdditionalAuthenticatedDataSeedBytes)+3] = 0x00

	derivedKey, err := generateAESCMAC(firstStageKey, deviationBuffer)
	if err != nil {
		return nil, fmt.Errorf("第二阶段 CMAC 派生失败: %w", err)
	}
	return derivedKey, nil
}

func signDataAPIAuditMessage(privateKeyBytes []byte, messageHex string) (string, error) {
	messageBytes, err := decodeHexString(messageHex)
	if err != nil {
		return "", fmt.Errorf("解析签名消息失败: %w", err)
	}

	messageHash := crypto.Keccak256(messageBytes)
	messagePrefix := fmt.Appendf(nil, "\x19Ethereum Signed Message:\n%d", len(messageHash))
	prefixedHash := crypto.Keccak256(messagePrefix, messageHash)
	signature, err := crypto.Sign(prefixedHash, mustToECDSA(privateKeyBytes))
	if err != nil {
		return "", fmt.Errorf("生成签名失败: %w", err)
	}

	if len(signature) != crypto.SignatureLength {
		return "", fmt.Errorf("签名长度错误: %d", len(signature))
	}
	signature[crypto.RecoveryIDOffset] += 27
	return encodeHexString(signature), nil
}

func encryptAESGCM(plainText, key, aad []byte) (cipherText, tag, iv []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, nil, err
	}

	iv = make([]byte, dataAPIGCMIVLengthBytes)
	if _, err = rand.Read(iv); err != nil {
		return nil, nil, nil, err
	}

	sealed := gcm.Seal(nil, iv, plainText, aad)
	tagStart := len(sealed) - dataAPIGCMTagLengthBytes
	cipherText = append([]byte(nil), sealed[:tagStart]...)
	tag = append([]byte(nil), sealed[tagStart:]...)
	return cipherText, tag, iv, nil
}

func decryptAESGCM(cipherText, tag, iv, key, aad []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	sealed := make([]byte, 0, len(cipherText)+len(tag))
	sealed = append(sealed, cipherText...)
	sealed = append(sealed, tag...)
	return gcm.Open(nil, iv, sealed, aad)
}

func buildAdditionalAuthenticatedData(prefix byte) []byte {
	aad := make([]byte, dataAPIAdditionalAuthenticatedDataSize)
	copy(aad, dataAPIAdditionalAuthenticatedDataSeedBytes)
	aad[24] = prefix
	return aad
}

func buildEncryptedPayload(cipherText, iv, ephemeralPublicKey, tag []byte) []byte {
	payload := make([]byte, 0, len(cipherText)+len(iv)+len(ephemeralPublicKey)+len(tag))
	payload = append(payload, cipherText...)
	payload = append(payload, iv...)
	payload = append(payload, ephemeralPublicKey...)
	payload = append(payload, tag...)
	return payload
}

func splitEncryptedPayload(payload []byte) (cipherText, iv, ephemeralPublicKey, tag []byte, err error) {
	minimumLength := dataAPIGCMIVLengthBytes + dataAPIEphemeralPublicKeyLengthBytes + dataAPIGCMTagLengthBytes + 1
	if len(payload) < minimumLength {
		return nil, nil, nil, nil, fmt.Errorf("密文长度不足: %d", len(payload))
	}

	cipherTextLength := len(payload) - dataAPIGCMIVLengthBytes - dataAPIEphemeralPublicKeyLengthBytes - dataAPIGCMTagLengthBytes
	offset := 0
	cipherText = append([]byte(nil), payload[offset:offset+cipherTextLength]...)
	offset += cipherTextLength
	iv = append([]byte(nil), payload[offset:offset+dataAPIGCMIVLengthBytes]...)
	offset += dataAPIGCMIVLengthBytes
	ephemeralPublicKey = append([]byte(nil), payload[offset:offset+dataAPIEphemeralPublicKeyLengthBytes]...)
	offset += dataAPIEphemeralPublicKeyLengthBytes
	tag = append([]byte(nil), payload[offset:offset+dataAPIGCMTagLengthBytes]...)
	return cipherText, iv, ephemeralPublicKey, tag, nil
}

func normalizePrivateKeyBytes(privateKeyBytes []byte) ([]byte, error) {
	if len(privateKeyBytes) == 0 {
		return nil, errors.New("私钥不能为空")
	}
	if len(privateKeyBytes) > dataAPIPrivateKeyLengthBytes {
		if len(privateKeyBytes) == dataAPIPrivateKeyLengthBytes+1 && privateKeyBytes[0] == 0x00 {
			privateKeyBytes = privateKeyBytes[1:]
		} else {
			return nil, fmt.Errorf("私钥长度非法: %d", len(privateKeyBytes))
		}
	}
	if len(privateKeyBytes) == dataAPIPrivateKeyLengthBytes {
		return append([]byte(nil), privateKeyBytes...), nil
	}
	normalized := make([]byte, dataAPIPrivateKeyLengthBytes)
	copy(normalized[dataAPIPrivateKeyLengthBytes-len(privateKeyBytes):], privateKeyBytes)
	return normalized, nil
}

func convertRawPublicKeyToUncompressed(rawPublicKey []byte) ([]byte, error) {
	if len(rawPublicKey) != dataAPIEphemeralPublicKeyLengthBytes {
		return nil, fmt.Errorf("裸公钥长度非法: %d", len(rawPublicKey))
	}
	uncompressed := make([]byte, dataAPIUncompressedPublicKeyLengthBytes)
	uncompressed[0] = 0x04
	copy(uncompressed[1:], rawPublicKey)
	return uncompressed, nil
}

func normalizeCoordinateBytes(value *big.Int) []byte {
	bytes := value.Bytes()
	normalized := make([]byte, dataAPIPrivateKeyLengthBytes)
	copy(normalized[dataAPIPrivateKeyLengthBytes-len(bytes):], bytes)
	return normalized
}

func decodeHexString(value string) ([]byte, error) {
	trimmedValue := strings.TrimPrefix(strings.TrimSpace(value), "0x")
	if trimmedValue == "" {
		return nil, errors.New("hex 字符串不能为空")
	}
	if len(trimmedValue)%2 != 0 {
		trimmedValue = "0" + trimmedValue
	}
	decoded, err := hex.DecodeString(trimmedValue)
	if err != nil {
		return nil, err
	}
	return decoded, nil
}

func generateAESCMAC(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cmac.Sum(data, block, dataAPIAESKeySizeBytes)
}

func decodeMust(value string) []byte {
	decoded, err := hex.DecodeString(value)
	if err != nil {
		panic(err)
	}
	return decoded
}

func mustToECDSA(privateKeyBytes []byte) *ecdsa.PrivateKey {
	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		panic(err)
	}
	return privateKey
}
