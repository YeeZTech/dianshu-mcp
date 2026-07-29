package sealed

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// 与 meta-encryptor src/common/limits.js 对齐的常量
const (
	headerSize              = 64
	magicNumHex             = "1fe2ef7f3ed18847"
	currentBlockFileVersion = 2
)

var magicNum, _ = hexToBytes(magicNumHex)

// Header 密封文件尾部 64 字节头
type Header struct {
	MagicNumber [8]byte
	Version     uint64
	BlockNumber uint64
	ItemNumber  uint64
	DataHash    [32]byte
}

// ReadHeader 从文件开头读取密封文件头
func ReadHeader(data []byte) (*Header, error) {
	if len(data) < headerSize {
		return nil, fmt.Errorf("数据长度不足，无法读取文件头")
	}

	headerBytes := data[0:headerSize]
	return ParseHeader(headerBytes)
}

// ParseHeader 解析密封文件头
func ParseHeader(headerBytes []byte) (*Header, error) {
	if len(headerBytes) != headerSize {
		return nil, fmt.Errorf("文件头长度错误: %d", headerSize)
	}

	h := &Header{}
	copy(h.MagicNumber[:], headerBytes[0:8])
	h.Version = binary.LittleEndian.Uint64(headerBytes[8:16])
	h.BlockNumber = binary.LittleEndian.Uint64(headerBytes[16:24])
	h.ItemNumber = binary.LittleEndian.Uint64(headerBytes[24:32])
	copy(h.DataHash[:], headerBytes[32:64])

	if !bytes.Equal(h.MagicNumber[:], magicNum) {
		return nil, fmt.Errorf("无效的魔数: %x", h.MagicNumber)
	}
	if h.Version != currentBlockFileVersion {
		return nil, fmt.Errorf("不支持的版本: %d", h.Version)
	}
	return h, nil
}

// TryReadItem 尝试从数据中读取一个加密项。
// TryReadItem 尝试从数据中读取一个加密项。
// 格式：8 字节长度前缀（little-endian uint64）+ 数据
func TryReadItem(data []byte) (cipher []byte, remaining []byte, err error) {
	if len(data) < 8 {
		return nil, nil, nil
	}

	itemSize := binary.LittleEndian.Uint64(data[0:8])
	if itemSize > uint64(len(data)-8) || itemSize > (1<<30) {
		return nil, nil, fmt.Errorf("item 长度异常: itemSize=%d (0x%x), dataLen=%d, first8=%x", itemSize, itemSize, len(data), data[:8])
	}
	totalItemBytes := 8 + int(itemSize)
	if len(data) < totalItemBytes {
		return nil, nil, nil
	}

	cipher = data[8:totalItemBytes]
	remaining = data[totalItemBytes:]
	return cipher, remaining, nil
}

// UnpackDecrypted 解包解密后的数据（ntpackage2batch + fromNtInput）
// 格式：[4 字节 packageID] + [8 字节 itemCount] + 若干 [8字节长度 + 数据]
// 每一项去除前 12 字节（NtObject 头）
func UnpackDecrypted(decrypted []byte) ([][]byte, error) {
	if len(decrypted) < 12 {
		return nil, fmt.Errorf("解密数据长度不足: %d", len(decrypted))
	}

	offset := 4 // 跳过 packageID
	count := binary.LittleEndian.Uint64(decrypted[offset:])
	offset += 8

	outputs := make([][]byte, 0, count)
	for i := uint64(0); i < count; i++ {
		if offset+8 > len(decrypted) {
			break
		}
		length := binary.LittleEndian.Uint64(decrypted[offset:])
		offset += 8
		if int(length) > 0 && offset+int(length) <= len(decrypted) {
			item := decrypted[offset : offset+int(length)]
			if len(item) > 12 {
				outputs = append(outputs, item[12:]) // fromNtInput: strip 12 bytes
			}
			offset += int(length)
		}
	}
	return outputs, nil
}

// Unseal 解密单个密封项，返回解包后的明文块。
// decrypt: 解密函数，输入密文，输出解密后的明文
func Unseal(decrypt func(cipher []byte) ([]byte, error), data []byte) ([][]byte, error) {
	// 跳过文件头（文件开头 64 字节）
	payload := data[headerSize:]

	var allOutputs [][]byte

	for {
		cipher, remaining, err := TryReadItem(payload)
		if err != nil {
			return nil, err
		}
		if cipher == nil {
			break
		}

		decrypted, err := decrypt(cipher)
		if err != nil {
			return nil, fmt.Errorf("解密项失败: %w", err)
		}

		outputs, err := UnpackDecrypted(decrypted)
		if err != nil {
			return nil, fmt.Errorf("解包失败: %w", err)
		}
		allOutputs = append(allOutputs, outputs...)

		payload = remaining
	}
	return allOutputs, nil
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
