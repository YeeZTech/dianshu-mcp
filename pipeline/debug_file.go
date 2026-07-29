package pipeline

import (
	"fmt"

	"dianshu-mcp/crypto"
)

// DebugFileFormat 分析 sealed 文件格式。
func DebugFileFormat(filePath, sealedPriKeyHex, userPriKeyHex string) {
	fmt.Printf("[debug-file] 分析文件: %s\n", filePath)
	fmt.Printf("[debug-file] sealedPriKey=%s\n", sealedPriKeyHex)
	fmt.Printf("[debug-file] userPriKey=%s\n", userPriKeyHex)

	// 从 sealedPriKey 推公钥
	sealedPub, _ := crypto.PublicKeyFromPrivate(sealedPriKeyHex)
	fmt.Printf("[debug-file] sealedPubKey=%s\n", sealedPub)
	userPub, _ := crypto.PublicKeyFromPrivate(userPriKeyHex)
	fmt.Printf("[debug-file] userPubKey=%s\n", userPub)
}
