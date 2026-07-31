// Package pipeline - 密封文件解密测试。
// 密封文件格式：[itemSize:8 LE][block_header:32][cipher_package:(itemSize-32)]
// 其中 cipher_package = [encrypted][iv:12][pubKey:64][tag:16]
//
// 运行：
//
//	SEALED_FILE=/path/to/file.sealed SEALED_KEY=<hex_key> \
//	  go test -run TestSealedFileFormat -v ./pkg/pipeline/
//
// Author: zhyyao
package pipeline

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"testing"

	"dianshu-mcp/pkg/crypto"
)

func TestSealedFileFormat(t *testing.T) {
	sealedPath := os.Getenv("SEALED_FILE")
	sealedKey := os.Getenv("SEALED_KEY")
	if sealedPath == "" || sealedKey == "" {
		t.Skip("设置 SEALED_FILE 和 SEALED_KEY 环境变量")
	}

	data, err := os.ReadFile(sealedPath)
	if err != nil {
		t.Fatalf("读取密封文件: %v", err)
	}

	const blockHeader = 32
	var allDecrypted []byte
	offset := 0
	blockCount := 0

	for offset+8 <= len(data) {
		itemSize := int(data[offset]) | int(data[offset+1])<<8 | int(data[offset+2])<<16 | int(data[offset+3])<<24 |
			int(data[offset+4])<<32 | int(data[offset+5])<<40 | int(data[offset+6])<<48 | int(data[offset+7])<<56
		offset += 8

		if itemSize <= blockHeader || offset+itemSize > len(data) {
			break
		}

		cipherPkg := data[offset : offset+itemSize]
		decrypted, err := crypto.DecryptInput(sealedKey, cipherPkg)
		if err != nil {
			t.Fatalf("块 %d (offset=%d, size=%d) 解密失败: %v", blockCount, offset, itemSize, err)
		}

		// 丢弃 32 字节 block header
		decrypted = decrypted[blockHeader:]
		allDecrypted = append(allDecrypted, decrypted...)
		blockCount++
		offset += itemSize
	}

	hash := fmt.Sprintf("%x", sha256.Sum256(allDecrypted))
	t.Logf("解密 %d 个块 → %d 字节, sha256=%s", blockCount, len(allDecrypted), hash)

	// 尝试作为 ZIP 打开
	if zr, err := zip.NewReader(bytes.NewReader(allDecrypted), int64(len(allDecrypted))); err == nil {
		t.Logf("有效 ZIP: %d 个文件", len(zr.File))
	}

	// 写入输出文件
	outPath := sealedPath + ".out"
	os.WriteFile(outPath, allDecrypted, 0644)
	t.Logf("输出: %s", outPath)
}
