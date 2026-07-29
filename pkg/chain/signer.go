// Package chain - see README for details.
//
// Author: zhyyao

package chain

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// SignOffChainSkeyTx 对 requestOffChainSkey 交易本地签名。
func SignOffChainSkeyTx(initResp *InitTxResponse, priKeyHex string) (string, error) {
	privateKey, err := crypto.HexToECDSA(priKeyHex)
	if err != nil {
		return "", fmt.Errorf("解析私钥失败: %w", err)
	}

	methodID := crypto.Keccak256([]byte("requestOffChainSkey(bytes32,bytes32,bytes32)"))[:4]
	data := make([]byte, 4+32*3)
	copy(data, methodID)
	copy(data[4:36], hexToBytes32(initResp.DataOnChainHash))
	copy(data[36:68], hexToBytes32(initResp.RequestOnChainHash))
	copy(data[68:100], hexToBytes32(initResp.ResultOnChainHash))

	to := common.HexToAddress(initResp.ContractAddress)
	tx := types.NewTransaction(
		uint64(initResp.Nonce),
		to,
		big.NewInt(0),
		uint64(initResp.GasLimit),
		big.NewInt(int64(initResp.GasPrice)),
		data,
	)

	signer := types.NewEIP155Signer(big.NewInt(int64(initResp.ChainID)))
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		return "", fmt.Errorf("签名失败: %w", err)
	}

	signedBytes, err := signedTx.MarshalBinary()
	if err != nil {
		return "", fmt.Errorf("序列化交易失败: %w", err)
	}
	return "0x" + hex.EncodeToString(signedBytes), nil
}

// hexToBytes32 converts a hex string to a fixed 32-byte array.
func hexToBytes32(hexStr string) []byte {
	if len(hexStr) >= 2 && hexStr[:2] == "0x" {
		hexStr = hexStr[2:]
	}
	result := make([]byte, 32)
	for i := 0; i < len(hexStr)/2 && i < 32; i++ {
		fmt.Sscanf(hexStr[i*2:i*2+2], "%02x", &result[i])
	}
	return result
}
