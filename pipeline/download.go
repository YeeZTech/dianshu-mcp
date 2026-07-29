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

	"dianshu-mcp/chain"
	"dianshu-mcp/crypto"
	"dianshu-mcp/dianshu"
	"dianshu-mcp/kms"

	"github.com/ethereum/go-ethereum/accounts/keystore"
)

// Config 下载管线配置
type Config struct {
	UserToken  string             // 典枢 JWT token
	UserInfo   *dianshu.UserInfo  // 用户信息（含 keystore）
	DianshuCli *dianshu.APIClient // 典枢 API 客户端
	ChainCli   *chain.Client      // 链 API 客户端（privateKey 为空时才需要）
	OutputDir  string             // 输出目录
}

// Run 执行下载管线。
// 流程：解密 keystore → 查 trade → 有 privateKey 直接解密下载，
// 没有则上链→轮询直到有 privateKey→解密下载。
func Run(ctx context.Context, cfg Config, taskCode string) error {
	// 1. 解密 keystore 获取用户私钥
	priKeyHex, err := decryptKeystore(ctx, cfg.UserToken, cfg.UserInfo)
	if err != nil {
		return fmt.Errorf("解密 keystore: %w", err)
	}
	fmt.Printf("[1] 用户私钥已获取\n")

	// 2. 查询 trade，获取任务信息
	task, err := fetchTask(ctx, cfg.DianshuCli, taskCode)
	if err != nil {
		return fmt.Errorf("查询任务: %w", err)
	}
	if task.FileURL == "" {
		return fmt.Errorf("任务 %s 无可下载文件", taskCode)
	}
	fmt.Printf("[2] 任务: %s, 文件: %s\n", task.TaskCode, task.FileURL)

	// 3. 获取封装私钥（已有直接用，没有就上链轮询）
	encryptedKey := task.PrivateKey
	if encryptedKey == "" {
		fmt.Printf("[3] privateKey 为空，需要上链\n")
		encryptedKey, err = ensurePrivateKey(ctx, cfg, priKeyHex, taskCode)
		if err != nil {
			return fmt.Errorf("获取封装私钥: %w", err)
		}
	} else {
		fmt.Printf("[3] privateKey 已存在\n")
	}

	// 4. 解密封装私钥得到 sealed 文件密钥
	sealedKey, err := decryptSealedKey(encryptedKey, priKeyHex)
	if err != nil {
		return fmt.Errorf("解密封装私钥: %w", err)
	}
	fmt.Printf("[4] 密封文件密钥已解密\n")

	// 5. 下载并解密
	return downloadAndUnseal(ctx, &task, sealedKey, priKeyHex, cfg.OutputDir)
}

// fetchTask 查询任务信息，通过 /system/task/privateKey 获取买方可解的封装私钥。
func fetchTask(ctx context.Context, cli *dianshu.APIClient, taskCode string) (dianshu.TaskItem, error) {
	tasks, err := cli.GetTradeList(ctx, taskCode)
	if err != nil {
		return dianshu.TaskItem{}, err
	}
	task := tasks[0]

	// tradeList 的 privateKey 是平台密钥加密的，买方无法解密
	// 需要通过 /system/task/privateKey 获取买方专属的封装私钥
	if task.PrivateKey == "" && task.ID > 0 {
		pkResult, pkErr := cli.GetTaskPrivateKey(ctx, task.ID)
		if pkErr == nil && pkResult != nil {
			task.PrivateKey = pkResult.PrivateKey
			task.PublishStatus = pkResult.PublishStatus
		}
	} else if task.PrivateKey != "" {
		// 如果 tradeList 里已有 privateKey，尝试用 GetTaskPrivateKey 覆盖
		pkResult, pkErr := cli.GetTaskPrivateKey(ctx, task.ID)
		if pkErr == nil && pkResult != nil && pkResult.PrivateKey != "" {
			fmt.Printf("[fetchTask] 使用 /system/task/privateKey 返回的买方专属密钥\n")
			task.PrivateKey = pkResult.PrivateKey
			task.PublishStatus = pkResult.PublishStatus
		}
	}
	return task, nil
}

// ensurePrivateKey 上链并轮询直到拿到 privateKey。
func ensurePrivateKey(ctx context.Context, cfg Config, priKeyHex, taskCode string) (string, error) {
	// 先检查链上状态，publishStatus==0 才发交易
	publishStatus, err := cfg.ChainCli.CheckTask(ctx, taskCode)
	if err != nil {
		return "", fmt.Errorf("检查链上状态失败: %w", err)
	}

	if publishStatus == 0 {
		// 初始化交易
		initResp, err := cfg.ChainCli.InitOffChainSkey(ctx, taskCode)
		if err != nil {
			return "", fmt.Errorf("初始化链上交易失败: %w", err)
		}

		// 本地签名
		signedTx, err := chain.SignOffChainSkeyTx(initResp, priKeyHex)
		if err != nil {
			return "", fmt.Errorf("签名交易失败: %w", err)
		}

		// 发送交易
		sendResp, err := cfg.ChainCli.SendTransaction(ctx, chain.SendTxRequest{
			UUID:                taskCode,
			SignedTransactionTx: signedTx,
		})
		if err != nil {
			return "", fmt.Errorf("发送交易失败: %w", err)
		}
		fmt.Printf("  链上交易已发送: %s\n", sendResp.TransactionHash)
	}

	// 轮询等待 privateKey 下发
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

			publishStatus, _ := cfg.ChainCli.CheckTask(ctx, taskCode)
			switch publishStatus {
			case 2:
				// 上链成功，查 trade 拿到 privateKey
				task, err := fetchTask(ctx, cfg.DianshuCli, taskCode)
				if err == nil && task.PrivateKey != "" {
					return task.PrivateKey, nil
				}
			case 3:
				return "", fmt.Errorf("链上交易失败(publishStatus=3)")
			}
		}
	}
}

// decryptKeystore 通过 KMS 获取授权密码，解密 keystore 得到用户私钥。
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

// decryptSealedKey 用用户私钥解密 trade 返回的封装私钥。
// decryptSealedKey 用用户私钥解密 trade 返回的封装私钥。
func decryptSealedKey(encryptedKey, priKeyHex string) (string, error) {
	DebugDecrypt(encryptedKey, priKeyHex)
	encBytes, err := hexToBytes(encryptedKey)
	if err != nil {
		return "", fmt.Errorf("[decryptSealedKey] hex 解码失败: %w", err)
	}
	fmt.Printf("[decryptSealedKey] 步骤4-解密封装私钥\n")
	fmt.Printf("[decryptSealedKey] encLen=%d\n", len(encBytes))

	if len(encBytes) < 12+64+16 {
		return "", fmt.Errorf("[decryptSealedKey] 长度异常: %d bytes", len(encBytes))
	}

	cipherLen := len(encBytes) - 12 - 64 - 16
	embeddedPubKey := encBytes[cipherLen+12 : cipherLen+12+64]
	userPubKey, _ := crypto.PublicKeyFromPrivate(priKeyHex)
	fmt.Printf("[decryptSealedKey] cipherLen=%d\n", cipherLen)
	fmt.Printf("[decryptSealedKey] embeddedPubKey=%x\n", embeddedPubKey)
	fmt.Printf("[decryptSealedKey] userPubKey=%s\n", userPubKey)
	fmt.Printf("[decryptSealedKey] 尝试 DecryptForwardMessage(prefix=0x01)...\n")

	decrypted, err := crypto.DecryptForwardMessage(priKeyHex, encBytes)
	if err != nil {
		fmt.Printf("[decryptSealedKey] DecryptForwardMessage 失败: %v, 尝试 DecryptInput(prefix=0x02)...\n", err)
		decrypted, err = crypto.DecryptInput(priKeyHex, encBytes)
		if err != nil {
			return "", fmt.Errorf("[decryptSealedKey] 两种prefix均失败: %w", err)
		}
		fmt.Printf("[decryptSealedKey] DecryptInput 成功!\n")
	} else {
		fmt.Printf("[decryptSealedKey] DecryptForwardMessage 成功!\n")
	}
	return fmt.Sprintf("%x", decrypted), nil
}

func downloadAndUnseal(ctx context.Context, task *dianshu.TaskItem, sealedPriKeyHex, userPriKeyHex, outputDir string) error {
	sealedPath, err := downloadSealedFile(ctx, task, outputDir)
	if err != nil {
		return fmt.Errorf("下载密封文件: %w", err)
	}
	defer os.Remove(sealedPath) // 解密完删除 .sealed 文件
	fmt.Printf("[5] 密封文件已下载: %s\n", sealedPath)

	sourceDir := filepath.Join(outputDir, "source-data")
	outputPath, err := unsealFile(sealedPath, sourceDir, sealedPriKeyHex, userPriKeyHex, task)
	if err != nil {
		return fmt.Errorf("解密密封文件: %w", err)
	}
	fmt.Printf("[5] 解密完成: %s\n", outputPath)
	return nil
}

func downloadSealedFile(ctx context.Context, task *dianshu.TaskItem, outputDir string) (string, error) {
	url := "https://d.dianshudata.com/" + task.FileURL

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("创建下载请求: %w", err)
	}
	req.Header.Set("User-Agent", "dianshu-mcp/1.0")
	req.Header.Set("Origin", "https://dianshudata.com")
	req.Header.Set("Referer", "https://dianshudata.com/")

	fmt.Printf("[下载] URL: %s\n", url)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("下载失败: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("[下载] status=%d content-type=%s\n", resp.StatusCode, resp.Header.Get("Content-Type"))

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

	n, err := io.Copy(f, resp.Body)
	if err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}
	fmt.Printf("[下载] 已写入 %d bytes\n", n)

	// 验证 sealed 文件魔数
	header := make([]byte, 8)
	if f2, err := os.Open(destPath); err == nil {
		f2.Read(header)
		f2.Close()
		fmt.Printf("[下载] 文件前8字节: %x (期望: 1fe2ef7f3ed18847)\n", header)
	}

	return destPath, nil
}
func unsealFile(sealedPath, outputDir, sealedPriKeyHex, userPriKeyHex string, task *dianshu.TaskItem) (string, error) {
	data, err := os.ReadFile(sealedPath)
	if err != nil {
		return "", fmt.Errorf("读取密封文件失败: %w", err)
	}

	fmt.Printf("[unsealFile] 文件大小: %d bytes, 前8字节: %x\n", len(data), data[:8])

	// 格式：[itemSize:8][encryptMessage:11882][block_info:32][header:64]
	// itemSize = first 8 bytes as uint64 LE
	itemSize := int(data[0]) | int(data[1])<<8 | int(data[2])<<16 | int(data[3])<<24 | int(data[4])<<32 | int(data[5])<<40 | int(data[6])<<48 | int(data[7])<<56
	fmt.Printf("[unsealFile] itemSize=%d, total=%d\n", itemSize, len(data))
	
	encData := data[8 : 8+itemSize]
	fmt.Printf("[unsealFile] encData len=%d\n", len(encData))

	decrypted, err := crypto.DecryptInput(sealedPriKeyHex, encData)
	if err != nil {
		return "", fmt.Errorf("解密失败: %w", err)
	}
	fmt.Printf("[unsealFile] 解密原始: %d bytes\n", len(decrypted))

	// 解包 NT package → 逐项 fromNtInput 去头
	plainData, err := unpackNtPackage(decrypted)
	if err != nil {
		return "", fmt.Errorf("解包失败: %w", err)
	}
	fmt.Printf("[unsealFile] 解包后: %d bytes\n", len(plainData))

	// 输出文件名用数据集名称
	outName := sanitizeFileName(task.DatasetName)
	if ext := task.Pattern; ext != "" {
		outName += "." + ext
	}
	outPath := filepath.Join(outputDir, outName)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", fmt.Errorf("创建输出目录: %w", err)
	}
	if err := os.WriteFile(outPath, plainData, 0o644); err != nil {
		return "", fmt.Errorf("写入文件: %w", err)
	}
	fmt.Printf("[unsealFile] 解密成功: %s (%d bytes)\n", outPath, len(plainData))
	return outPath, nil
}

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

// sanitizeFileName 清除文件名中的非法字符
func sanitizeFileName(name string) string {
	replacer := strings.NewReplacer("/", "_", "\\", "_", ":", "_", "*", "_", "?", "_", "\"", "_", "<", "_", ">", "_", "|", "_")
	return replacer.Replace(name)
}

// unpackNtPackage 解包 NT package：ntpackage2batch → fromNtInput
func unpackNtPackage(pkg []byte) ([]byte, error) {
	if len(pkg) < 12 {
		return nil, fmt.Errorf("NT 包太短: %d bytes", len(pkg))
	}
	offset := 4 // 跳过 package_id (4 bytes)
	count := int(pkg[offset]) | int(pkg[offset+1])<<8 | int(pkg[offset+2])<<16 | int(pkg[offset+3])<<24 |
		int(pkg[offset+4])<<32 | int(pkg[offset+5])<<40 | int(pkg[offset+6])<<48 | int(pkg[offset+7])<<56
	offset += 8

	var out []byte
	for i := 0; i < count; i++ {
		if offset+8 > len(pkg) {
			break
		}
		itemLen := int(pkg[offset]) | int(pkg[offset+1])<<8 | int(pkg[offset+2])<<16 | int(pkg[offset+3])<<24 |
			int(pkg[offset+4])<<32 | int(pkg[offset+5])<<40 | int(pkg[offset+6])<<48 | int(pkg[offset+7])<<56
		offset += 8
		if itemLen > 0 && offset+itemLen <= len(pkg) {
			item := pkg[offset : offset+itemLen]
			if len(item) > 12 {
				out = append(out, item[12:]...) // fromNtInput: strip 12 bytes
			}
			offset += itemLen
		}
	}
	return out, nil
}
