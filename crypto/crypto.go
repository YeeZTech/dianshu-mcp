package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/sha3"
)

// 与前端 meta-encryptor 对齐的常量
var (
	aad              = []byte("tech.yeez.key.manager")
	cmacKeyBytes, _  = hex.DecodeString("7965657a2e746563682e7374626f7800")
	ethHashPrefix    = append([]byte{0x19}, []byte("Ethereum Signed Message:\n32")...)
	derivationBuffer = buildDerivationBuffer()
)

func buildDerivationBuffer() []byte {
	buf := make([]byte, len(aad)+4)
	buf[0] = 0x01
	copy(buf[1:], aad)
	buf[len(aad)+1] = 0
	buf[len(aad)+2] = 0x80
	buf[len(aad)+3] = 0x00
	return buf
}

// ---------- hex 编解码 ----------

func hexToBytes(s string) ([]byte, error) {
	if len(s) >= 2 && s[:2] == "0x" {
		s = s[2:]
	}
	return hex.DecodeString(s)
}

func toHex(b []byte) string {
	return hex.EncodeToString(b)
}

// ---------- AES-CMAC ----------

// aesCmac 计算 AES-CMAC，用于密钥派生。
func aesCmac(key, message []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aesCmac NewCipher: %w", err)
	}

	const blockSize = 16
	const constRb = 0x87

	zeros := make([]byte, blockSize)
	L := make([]byte, blockSize)
	block.Encrypt(L, zeros)

	K1 := leftShift(L)
	if L[0]&0x80 != 0 {
		K1[blockSize-1] ^= constRb
	}
	K2 := leftShift(K1)
	if K1[0]&0x80 != 0 {
		K2[blockSize-1] ^= constRb
	}

	n := (len(message) + blockSize - 1) / blockSize
	flagComplete := len(message) > 0 && len(message)%blockSize == 0

	var lastBlock []byte
	if n == 0 {
		flagComplete = false
		lastBlock = xor16(K2, append([]byte{0x80}, make([]byte, 15)...))
	} else {
		startLast := (n - 1) * blockSize
		lb := make([]byte, blockSize)
		copy(lb, message[startLast:])
		if flagComplete {
			lastBlock = xor16(lb, K1)
		} else {
			pad := make([]byte, blockSize)
			copy(pad, lb)
			pad[len(message)-startLast] = 0x80
			lastBlock = xor16(pad, K2)
		}
	}

	X := make([]byte, blockSize)
	for i := 0; i < n-1; i++ {
		block.Encrypt(X, xor16(X, message[i*blockSize:(i+1)*blockSize]))
	}
	block.Encrypt(X, xor16(X, lastBlock))
	return X, nil
}

func leftShift(buf []byte) []byte {
	out := make([]byte, len(buf))
	carry := byte(0)
	for i := len(buf) - 1; i >= 0; i-- {
		out[i] = ((buf[i] << 1) & 0xFF) | carry
		if buf[i]&0x80 != 0 {
			carry = 1
		} else {
			carry = 0
		}
	}
	return out
}

func xor16(a, b []byte) []byte {
	out := make([]byte, 16)
	for i := 0; i < 16; i++ {
		out[i] = a[i] ^ b[i]
	}
	return out
}

// ---------- secp256k1 ECDH ----------

// PublicKeyFromPrivate 从私钥（32 字节 hex）生成公钥（64 字节 hex，uncompressed 去除前缀 04）。
func PublicKeyFromPrivate(skeyHex string) (string, error) {
	skeyBytes, err := hexToBytes(skeyHex)
	if err != nil {
		return "", fmt.Errorf("私钥 hex 解码失败: %w", err)
	}
	privKey := secp256k1.PrivKeyFromBytes(skeyBytes)
	pubKey := privKey.PubKey()
	return toHex(pubKey.SerializeUncompressed()[1:]), nil
}

// DeriveAESKey 通过 ECDH + AES-CMAC 派生 AES-128 密钥。
func DeriveAESKey(pkeyHex, skeyHex string) ([]byte, error) {
	pkeyBytes, err := hexToBytes(pkeyHex)
	if err != nil {
		return nil, fmt.Errorf("公钥 hex 解码失败: %w", err)
	}
	if len(pkeyBytes) == 64 {
		pkeyBytes = append([]byte{0x04}, pkeyBytes...)
	}

	skeyBytes, err := hexToBytes(skeyHex)
	if err != nil {
		return nil, fmt.Errorf("私钥 hex 解码失败: %w", err)
	}

	sharedKey := genECDHKey(skeyBytes, pkeyBytes)

	keyDeriveKey, err := aesCmac(cmacKeyBytes, sharedKey)
	if err != nil {
		return nil, fmt.Errorf("AES-CMAC 派生失败: %w", err)
	}
	derivedKey, err := aesCmac(keyDeriveKey, derivationBuffer)
	if err != nil {
		return nil, fmt.Errorf("AES-CMAC 派生失败: %w", err)
	}
	return derivedKey, nil
}

func genECDHKey(skey, pkey []byte) []byte {
	// 使用 go-ethereum S256 曲线，与 JS secp256k1 的 C 实现一致
	return genECDHKeyGeth(skey, pkey)
}

func genECDHKeyGeth(skey, pkey []byte) []byte {
	curve := ethcrypto.S256()

	priv := new(big.Int).SetBytes(skey)
	pubX, pubY := new(big.Int).SetBytes(pkey[1:33]), new(big.Int).SetBytes(pkey[33:65])

	sharedX, sharedY := curve.ScalarMult(pubX, pubY, priv.Bytes())

	compressed := make([]byte, 33)
	if sharedY.Bit(0) == 1 {
		compressed[0] = 0x03
	} else {
		compressed[0] = 0x02
	}
	sharedX.FillBytes(compressed[1:33])

	h := sha256.Sum256(compressed)
	return h[:]
}

// ---------- EIP-191 签名 ----------

// SignMessage EIP-191 签名：keccak256("\x19Ethereum Signed Message:\n32" + keccak256(message))
func SignMessage(skeyHex string, message []byte) ([]byte, error) {
	skeyBytes, err := hexToBytes(skeyHex)
	if err != nil {
		return nil, fmt.Errorf("私钥解码失败: %w", err)
	}
	privKey := secp256k1.PrivKeyFromBytes(skeyBytes)

	rawHash := keccak256(message)
	msg := append(ethHashPrefix, rawHash...)
	msgHash := keccak256(msg)

	sigCompact := ecdsa.SignCompact(privKey, msgHash, false)
	sig := make([]byte, 65)
	copy(sig, sigCompact[1:65])
	sig[64] = sigCompact[0]
	return sig, nil
}

func keccak256(data []byte) []byte {
	h := sha3.NewLegacyKeccak256()
	h.Write(data)
	return h.Sum(nil)
}

// ---------- AES-128-GCM 加密/解密 ----------

const aesKeySize = 16
const gcmTagSize = 16
const gcmIVSize = 12

// encryptMessage 使用 AES-128-GCM 加密消息，返回完整密文包。
func encryptMessage(pkeyHex, skeyHex string, plaintext, generatedPubKey []byte, prefix byte) ([]byte, error) {
	aesKey, err := DeriveAESKey(pkeyHex, skeyHex)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(aesKey[:aesKeySize])
	if err != nil {
		return nil, err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	iv := make([]byte, gcmIVSize)
	if _, err := rand.Read(iv); err != nil {
		return nil, fmt.Errorf("生成 IV 失败: %w", err)
	}

	tad := make([]byte, 64)
	copy(tad, aad)
	tad[24] = prefix

	sealed := aesGCM.Seal(nil, iv, plaintext, tad)
	encrypted := make([]byte, len(sealed)-gcmTagSize)
	tag := make([]byte, gcmTagSize)
	copy(encrypted, sealed[:len(sealed)-gcmTagSize])
	copy(tag, sealed[len(sealed)-gcmTagSize:])

	result := make([]byte, 0, len(encrypted)+gcmIVSize+64+gcmTagSize)
	result = append(result, encrypted...)
	result = append(result, iv...)
	result = append(result, generatedPubKey...)
	result = append(result, tag...)
	return result, nil
}

// decryptMessage 解密 AES-128-GCM 密文包。
func decryptMessage(skeyHex string, cipherPackage []byte, prefix byte) ([]byte, error) {
	if len(cipherPackage) < 64+gcmTagSize+gcmIVSize {
		return nil, fmt.Errorf("密文包长度不足")
	}
	cipherLen := len(cipherPackage) - 64 - gcmTagSize - gcmIVSize
	encrypted := cipherPackage[:cipherLen]
	iv := cipherPackage[cipherLen : cipherLen+gcmIVSize]
	pubKey := cipherPackage[cipherLen+gcmIVSize : cipherLen+gcmIVSize+64]
	tag := cipherPackage[cipherLen+gcmIVSize+64:]

	aesKey, err := DeriveAESKey(toHex(pubKey), skeyHex)
	if err != nil {
		return nil, err
	}

	tad := make([]byte, 64)
	copy(tad, aad)
	tad[24] = prefix

	block, err := aes.NewCipher(aesKey[:aesKeySize])
	if err != nil {
		return nil, err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	sealed := make([]byte, len(encrypted)+len(tag))
	copy(sealed, encrypted)
	copy(sealed[len(encrypted):], tag)
	return aesGCM.Open(nil, iv, sealed, tad)
}

// ---------- 高级封装 ----------

func generatePrivateKey() (string, error) {
	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return "", fmt.Errorf("生成私钥失败: %w", err)
	}
	return toHex(privKey.Serialize()), nil
}

// GenerateForwardSecretKey 生成临时加密私钥，加密后发送给对方。
func GenerateForwardSecretKey(remotePkeyHex, skeyHex string) (otsHex string, encryptedSkey []byte, forwardSig []byte, err error) {
	otsHex, err = generatePrivateKey()
	if err != nil {
		return "", nil, nil, err
	}
	otsBytes, _ := hexToBytes(otsHex)

	otsPubKey, _ := PublicKeyFromPrivate(otsHex)
	pubKeyBytes, _ := hexToBytes(otsPubKey)

	encryptedSkey, err = encryptMessage(remotePkeyHex, otsHex, otsBytes, pubKeyBytes, 0x01)
	if err != nil {
		return "", nil, nil, err
	}
	return otsHex, encryptedSkey, nil, nil
}

// DecryptForwardMessage 解密转发消息，提取出临时加密私钥。
func DecryptForwardMessage(skeyHex string, cipherPackage []byte) ([]byte, error) {
	return decryptMessage(skeyHex, cipherPackage, 0x01)
}

// GenerateEncryptedInput 用接收方公钥加密输入数据。
func GenerateEncryptedInput(localPkeyHex string, input []byte) ([]byte, error) {
	otsHex, err := generatePrivateKey()
	if err != nil {
		return nil, err
	}
	otsPubKey, _ := PublicKeyFromPrivate(otsHex)
	pubKeyBytes, _ := hexToBytes(otsPubKey)
	return encryptMessage(localPkeyHex, otsHex, input, pubKeyBytes, 0x02)
}

// DecryptInput 解密输入数据。
func DecryptInput(skeyHex string, cipherPackage []byte) ([]byte, error) {
	return decryptMessage(skeyHex, cipherPackage, 0x02)
}
