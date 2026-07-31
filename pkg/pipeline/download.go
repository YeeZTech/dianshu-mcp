// Package pipeline - 下载管线：KMS→查任务→上链→下载→解密→解包。
// Package pipeline downloads and decrypts sealed data files.
//
// Author: zhyyao
package pipeline

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dianshu-mcp/dianshu"
	"dianshu-mcp/pkg/chain"
	"dianshu-mcp/pkg/crypto"
	"dianshu-mcp/pkg/kms"

	"github.com/ethereum/go-ethereum/accounts/keystore"
)

// Config 下载管线配置。
type Config struct {
	UserToken  string             // 典枢 JWT token
	UserInfo   *dianshu.UserInfo  // 用户信息（含 keystore）
	DianshuCli *dianshu.APIClient // 典枢 API 客户端
	ChainCli   *chain.Client      // 链 API 客户端
	OutputDir  string             // 输出目录
}

// Run 执行完整下载管线，返回输出文件路径。
// Run executes the full download pipeline.
func Run(ctx context.Context, cfg Config, taskCode string) (string, error) {
	priKeyHex, err := decryptKeystore(ctx, cfg.UserToken, cfg.UserInfo)
	if err != nil {
		return "", fmt.Errorf("解密 keystore: %w", err)
	}

	task, err := fetchTask(ctx, cfg.DianshuCli, taskCode)
	if err != nil {
		return "", fmt.Errorf("查询任务: %w", err)
	}
	if task.FileURL == "" {
		return "", fmt.Errorf("任务 %s 无可下载文件", taskCode)
	}

	encryptedKey := task.PrivateKey
	if encryptedKey == "" {
		encryptedKey, err = ensurePrivateKey(ctx, cfg, priKeyHex, taskCode, task.ID)
		if err != nil {
			return "", fmt.Errorf("获取封装私钥: %w", err)
		}
	}

	sealedKey, err := decryptSealedKey(encryptedKey, priKeyHex)
	if err != nil {
		return "", fmt.Errorf("解密封装私钥: %w", err)
	}

	return downloadAndUnseal(ctx, &task, sealedKey, priKeyHex, cfg.OutputDir)
}

// fetchTask 查询任务信息并获取买方专属私钥。
// fetchTask retrieves task info and buyer-specific private key.
func fetchTask(ctx context.Context, cli *dianshu.APIClient, taskCode string) (dianshu.TaskItem, error) {
	task, err := cli.GetTaskByCode(ctx, taskCode)
	if err != nil {
		return dianshu.TaskItem{}, fmt.Errorf("查询任务: %w", err)
	}
	if task == nil {
		return dianshu.TaskItem{}, fmt.Errorf("任务 %s 不存在", taskCode)
	}

	pkResult, pkErr := cli.GetTaskPrivateKey(ctx, task.ID)
	if pkResult != nil {
		task.PublishStatus = pkResult.PublishStatus
		if pkResult.PrivateKey != "" {
			task.PrivateKey = pkResult.PrivateKey
		}
		_ = pkErr
	}
	return *task, nil
}

// ensurePrivateKey 上链并轮询直到获取封装私钥。
// ensurePrivateKey submits chain tx and polls for the sealed private key.
func ensurePrivateKey(ctx context.Context, cfg Config, priKeyHex, taskCode string, taskID int) (string, error) {
	publishStatus, err := cfg.ChainCli.CheckTask(ctx, taskCode)
	if err != nil {
		return "", fmt.Errorf("检查链上状态失败: %w", err)
	}

	if publishStatus == 0 {
		initResp, err := cfg.ChainCli.InitOffChainSkey(ctx, taskCode)
		if err != nil {
			return "", fmt.Errorf("初始化链上交易失败: %w", err)
		}
		signedTx, err := chain.SignOffChainSkeyTx(initResp, priKeyHex)
		if err != nil {
			return "", fmt.Errorf("签名交易失败: %w", err)
		}
		_, err = cfg.ChainCli.SendTransaction(ctx, chain.SendTxRequest{
			UUID:                taskCode,
			SignedTransactionTx: signedTx,
		})
		if err != nil {
			return "", fmt.Errorf("发送交易失败: %w", err)
		}
	}

	deadline := time.Now().Add(10 * time.Minute)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return "", fmt.Errorf("等待封装私钥超时")
			}
			pkResult, err := cfg.DianshuCli.GetTaskPrivateKey(ctx, taskID)
			if err != nil {
				continue
			}
			if pkResult != nil && pkResult.PrivateKey != "" {
				return pkResult.PrivateKey, nil
			}
			if pkResult != nil && pkResult.PublishStatus == 3 {
				return "", fmt.Errorf("链上交易失败(publishStatus=3)")
			}
		}
	}
}

// decryptKeystore 通过 KMS 解密用户 keystore 获取私钥。
// decryptKeystore decrypts the user keystore via KMS.
func decryptKeystore(ctx context.Context, userToken string, userInfo *dianshu.UserInfo) (string, error) {
	clientToken, err := kms.Login(userToken)
	if err != nil {
		return "", fmt.Errorf("KMS 登录失败: %w", err)
	}
	password, err := kms.GetAuthPassword(userInfo.UserID, clientToken)
	if err != nil {
		return "", fmt.Errorf("获取授权密码失败: %w", err)
	}
	if userInfo.PrivateKey == "" {
		return "", fmt.Errorf("用户 keystore 为空")
	}
	key, err := keystore.DecryptKey([]byte(userInfo.PrivateKey), password)
	if err != nil {
		return "", fmt.Errorf("解密 keystore 失败: %w", err)
	}
	return fmt.Sprintf("%x", key.PrivateKey.D.Bytes()), nil
}

// decryptSealedKey 用用户私钥解密封装文件密钥。
// decryptSealedKey decrypts the sealed file encryption key.
func decryptSealedKey(encryptedKey, priKeyHex string) (string, error) {
	encBytes, err := hexToBytes(encryptedKey)
	if err != nil {
		return "", fmt.Errorf("封装私钥 hex 解码失败: %w", err)
	}
	if len(encBytes) < 12+64+16 {
		return "", fmt.Errorf("封装私钥长度异常: %d bytes", len(encBytes))
	}

	decrypted, err := crypto.DecryptForwardMessage(priKeyHex, encBytes)
	if err != nil {
		decrypted, err = crypto.DecryptInput(priKeyHex, encBytes)
		if err != nil {
			return "", fmt.Errorf("解密封装私钥失败: %w", err)
		}
	}
	return fmt.Sprintf("%x", decrypted), nil
}

// downloadAndUnseal 下载密封文件并解密。
// downloadAndUnseal downloads and decrypts a sealed file.
func downloadAndUnseal(ctx context.Context, task *dianshu.TaskItem, sealedPriKeyHex, userPriKeyHex, outputDir string) (string, error) {
	sealedPath, err := downloadSealedFile(ctx, task, outputDir)
	if err != nil {
		return "", fmt.Errorf("下载密封文件: %w", err)
	}
	defer os.Remove(sealedPath)

	outputPath, err := unsealFile(sealedPath, outputDir, sealedPriKeyHex, userPriKeyHex, task)
	if err != nil {
		return "", fmt.Errorf("解密密封文件: %w", err)
	}
	return outputPath, nil
}

// downloadSealedFile 从 CDN 下载密封文件。
// downloadSealedFile downloads a sealed file from CDN.
func downloadSealedFile(ctx context.Context, task *dianshu.TaskItem, outputDir string) (string, error) {
	url := "https://d.dianshudata.com/" + task.FileURL

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("创建下载请求: %w", err)
	}
	req.Header.Set("User-Agent", "dianshu-mcp/1.0")
	req.Header.Set("Origin", "https://dianshudata.com")
	req.Header.Set("Referer", "https://dianshudata.com/")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("下载失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return "", fmt.Errorf("下载返回异常状态: %d", resp.StatusCode)
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", fmt.Errorf("创建输出目录失败: %w", err)
	}

	destPath := filepath.Join(outputDir, task.FileURL)
	f, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}
	return destPath, nil
}

// unsealFile 解密密封文件。
// 格式：[itemSize:8 LE][block_header:32][cipher_package:(itemSize-32)]
// 其中 cipher_package = [encrypted][iv:12][pubKey:64][tag:16]
// 解密后丢弃每块前 32 字节 block header。
// unsealFile decrypts the sealed file by processing all item blocks.
func unsealFile(sealedPath, outputDir, sealedPriKeyHex, userPriKeyHex string, task *dianshu.TaskItem) (string, error) {
	data, err := os.ReadFile(sealedPath)
	if err != nil {
		return "", fmt.Errorf("读取密封文件失败: %w", err)
	}

	const blockHeader = 32
	var allDecrypted []byte
	offset := 0

	for offset+8 <= len(data) {
		itemSize := int(data[offset]) | int(data[offset+1])<<8 | int(data[offset+2])<<16 | int(data[offset+3])<<24 |
			int(data[offset+4])<<32 | int(data[offset+5])<<40 | int(data[offset+6])<<48 | int(data[offset+7])<<56
		offset += 8

		if itemSize <= blockHeader || offset+itemSize > len(data) {
			break
		}

		cipherPkg := data[offset : offset+itemSize]
		decrypted, err := crypto.DecryptInput(sealedPriKeyHex, cipherPkg)
		if err != nil {
			return "", fmt.Errorf("解密块失败: %w", err)
		}
		// 丢弃每块前 48 字节 block header
		if len(decrypted) > blockHeader {
			decrypted = decrypted[blockHeader:]
		}
		allDecrypted = append(allDecrypted, decrypted...)
		offset += itemSize
	}

	if len(allDecrypted) == 0 {
		return "", fmt.Errorf("密封文件中没有可解密的块")
	}

	outName := sanitizeFileName(task.DatasetName)
	if ext := task.Pattern; ext != "" {
		outName += "." + ext
	}
	outPath := filepath.Join(outputDir, outName)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", fmt.Errorf("创建输出目录: %w", err)
	}
	if err := os.WriteFile(outPath, allDecrypted, 0o644); err != nil {
		return "", fmt.Errorf("写入文件: %w", err)
	}
	return outPath, nil
}

// sanitizeFileName 替换文件名中的非法字符。
// sanitizeFileName replaces forbidden characters in filenames.
func sanitizeFileName(name string) string {
	replacer := strings.NewReplacer(
		"/", "_", "\\", "_", ":", "_", "*", "_",
		"?", "_", "\"", "_", "<", "_", ">", "_", "|", "_",
	)
	return replacer.Replace(name)
}

// hexToBytes 十六进制转字节数组。
// hexToBytes converts hex string to bytes.
func hexToBytes(s string) ([]byte, error) {
	if len(s) >= 2 && s[:2] == "0x" {
		s = s[2:]
	}
	result := make([]byte, len(s)/2)
	for i := 0; i < len(result); i++ {
		_, err := fmt.Sscanf(s[i*2:i*2+2], "%02x", &result[i])
		if err != nil {
			return nil, fmt.Errorf("hex 解码失败: %w", err)
		}
	}
	return result, nil
}
