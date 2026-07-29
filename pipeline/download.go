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

// Config 下载管线配置。
type Config struct {
	UserToken  string             // 典枢 JWT token
	UserInfo   *dianshu.UserInfo  // 用户信息（含 keystore）
	DianshuCli *dianshu.APIClient // 典枢 API 客户端
	ChainCli   *chain.Client      // 链 API 客户端（privateKey 为空时才需要）
	OutputDir  string             // 输出目录
}

// Run 执行下载管线。
func Run(ctx context.Context, cfg Config, taskCode string) error {
	// 1. KMS 解密 keystore → 用户私钥
	priKeyHex, err := decryptKeystore(ctx, cfg.UserToken, cfg.UserInfo)
	if err != nil {
		return fmt.Errorf("解密 keystore: %w", err)
	}

	// 2. 查询 trade 信息
	task, err := fetchTask(ctx, cfg.DianshuCli, taskCode)
	if err != nil {
		return fmt.Errorf("查询任务: %w", err)
	}
	if task.FileURL == "" {
		return fmt.Errorf("任务 %s 无可下载文件", taskCode)
	}

	// 3. 获取封装私钥（已有直接用，空则上链轮询）
	encryptedKey := task.PrivateKey
	if encryptedKey == "" {
		encryptedKey, err = ensurePrivateKey(ctx, cfg, priKeyHex, taskCode)
		if err != nil {
			return fmt.Errorf("获取封装私钥: %w", err)
		}
	}

	// 4. 解密封装私钥 → sealed 文件密钥
	sealedKey, err := decryptSealedKey(encryptedKey, priKeyHex)
	if err != nil {
		return fmt.Errorf("解密封装私钥: %w", err)
	}

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

	// tradeList 的 privateKey 是平台密钥加密的，需要用 /system/task/privateKey 获取买方专属密钥
	pkResult, pkErr := cli.GetTaskPrivateKey(ctx, task.ID)
	if pkErr == nil && pkResult != nil && pkResult.PrivateKey != "" {
		task.PrivateKey = pkResult.PrivateKey
		task.PublishStatus = pkResult.PublishStatus
	}
	return task, nil
}

// ensurePrivateKey 上链并轮询直到拿到封装私钥。
func ensurePrivateKey(ctx context.Context, cfg Config, priKeyHex, taskCode string) (string, error) {
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

// decryptKeystore 通过 KMS 授权密码解密用户 keystore。
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

// decryptSealedKey 用用户私钥解密封装私钥，得到 sealed 文件加密密钥。
func decryptSealedKey(encryptedKey, priKeyHex string) (string, error) {
	encBytes, err := hexToBytes(encryptedKey)
	if err != nil {
		return "", fmt.Errorf("封装私钥 hex 解码失败: %w", err)
	}
	if len(encBytes) < 12+64+16 {
		return "", fmt.Errorf("封装私钥长度异常: %d bytes", len(encBytes))
	}

	// 优先 DecryptForwardMessage (0x01)，失败则 DecryptInput (0x02)
	decrypted, err := crypto.DecryptForwardMessage(priKeyHex, encBytes)
	if err != nil {
		decrypted, err = crypto.DecryptInput(priKeyHex, encBytes)
		if err != nil {
			return "", fmt.Errorf("解密封装私钥失败: %w", err)
		}
	}
	return fmt.Sprintf("%x", decrypted), nil
}

func downloadAndUnseal(ctx context.Context, task *dianshu.TaskItem, sealedPriKeyHex, userPriKeyHex, outputDir string) error {
	sealedPath, err := downloadSealedFile(ctx, task, outputDir)
	if err != nil {
		return fmt.Errorf("下载密封文件: %w", err)
	}
	defer os.Remove(sealedPath)

	sourceDir := filepath.Join(outputDir, "source-data")
	outputPath, err := unsealFile(sealedPath, sourceDir, sealedPriKeyHex, userPriKeyHex, task)
	if err != nil {
		return fmt.Errorf("解密密封文件: %w", err)
	}
	fmt.Printf("解密完成: %s\n", outputPath)
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

func unsealFile(sealedPath, outputDir, sealedPriKeyHex, userPriKeyHex string, task *dianshu.TaskItem) (string, error) {
	data, err := os.ReadFile(sealedPath)
	if err != nil {
		return "", fmt.Errorf("读取密封文件失败: %w", err)
	}

	// 格式：[itemSize:8 LE][encryptMessage:itemSize][尾部]
	itemSize := int(data[0]) | int(data[1])<<8 | int(data[2])<<16 | int(data[3])<<24 |
		int(data[4])<<32 | int(data[5])<<40 | int(data[6])<<48 | int(data[7])<<56

	if itemSize <= 0 || 8+itemSize > len(data) {
		return "", fmt.Errorf("密封文件 itemSize 异常: %d", itemSize)
	}

	// 解密
	decrypted, err := crypto.DecryptInput(sealedPriKeyHex, data[8:8+itemSize])
	if err != nil {
		return "", fmt.Errorf("解密失败: %w", err)
	}

	// 解包 NT package
	plainData, err := unpackNtPackage(decrypted)
	if err != nil {
		return "", fmt.Errorf("解包失败: %w", err)
	}

	// 输出
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
	return outPath, nil
}

func sanitizeFileName(name string) string {
	replacer := strings.NewReplacer(
		"/", "_", "\\", "_", ":", "_", "*", "_",
		"?", "_", "\"", "_", "<", "_", ">", "_", "|", "_",
	)
	return replacer.Replace(name)
}

// unpackNtPackage 解包 NT package：ntpackage2batch → fromNtInput。
func unpackNtPackage(pkg []byte) ([]byte, error) {
	if len(pkg) < 12 {
		return nil, fmt.Errorf("NT 包太短: %d bytes", len(pkg))
	}

	offset := 4 // 跳过 package_id
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
