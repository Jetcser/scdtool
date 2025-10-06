package main

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"bytes"
)

func checkFileHeader(fpath string) (string, string, error) {
	f, err := os.Open(fpath)
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	headerBytes := make([]byte, 9)
	if _, err := f.Read(headerBytes); err != nil {
		return "", "", err
	}
	headerHex := hex.EncodeToString(headerBytes)

	switch headerHex {
	case "40150000d26d530101":
		return "用户自定义词库", headerHex, nil
	case "401500004443530101":
		return "官方词库", headerHex, nil
	default:
		return "", "", fmt.Errorf("%s 似乎不是搜狗细胞词库，请检查。", fpath)
	}
}

func decodeUTF16LE(buf []byte) string {
	decoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()
	reader := transform.NewReader(bytes.NewReader(buf), decoder)
	decoded, err := io.ReadAll(reader)
	if err != nil {
		return ""
	}
	// 去掉 \x00 和换行符
	result := bytes.ReplaceAll(decoded, []byte{0x00}, []byte{})
	result = bytes.ReplaceAll(result, []byte("\n"), []byte(" "))
	return string(result)
}

func extractUTF16LEString(fpath string, start, end int64) string {
	f, err := os.Open(fpath)
	if err != nil {
		return ""
	}
	defer f.Close()

	buf := make([]byte, end-start+1)
	if _, err := f.ReadAt(buf, start); err != nil && err != io.EOF {
		return ""
	}

	return decodeUTF16LE(buf)
}

func extractExampleString(fpath string, start, end int64) string {
	f, err := os.Open(fpath)
	if err != nil {
		return ""
	}
	defer f.Close()

	buf := make([]byte, end-start+1)
	if _, err := f.ReadAt(buf, start); err != nil && err != io.EOF {
		return ""
	}

	// 替换 0D 00 20 00 -> 20 00
	for i := 0; i+3 < len(buf); i += 2 {
		if buf[i] == 0x0d && buf[i+1] == 0x00 && buf[i+2] == 0x20 && buf[i+3] == 0x00 {
			buf = append(buf[:i], buf[i+2:]...)
		}
	}

	return decodeUTF16LE(buf)
}

func extractEntryCount(fpath string) int {
	f, err := os.Open(fpath)
	if err != nil {
		return 0
	}
	defer f.Close()

	buf := make([]byte, 4)
	if _, err := f.ReadAt(buf, 0x124); err != nil {
		return 0
	}
	return int(binary.LittleEndian.Uint32(buf))
}

func extractTimestamp(fpath string) (string, uint32) {
	f, err := os.Open(fpath)
	if err != nil {
		return "", 0
	}
	defer f.Close()

	buf := make([]byte, 4)
	if _, err := f.ReadAt(buf, 0x11C); err != nil {
		return "", 0
	}
	ts := binary.LittleEndian.Uint32(buf)
	t := time.Unix(int64(ts), 0)
	return fmt.Sprintf("%s （时间戳：%d）", t.Format("2006-01-02 15:04:05"), ts), ts
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("用法: %s <搜狗细胞词库路径>\n", os.Args[0])
		os.Exit(1)
	}
	filePath := os.Args[1]

	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		fmt.Printf("错误: 文件 '%s' 不存在。\n", filePath)
		os.Exit(1)
	}

	source, header, err := checkFileHeader(filePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	dictInfo := map[string]interface{}{
		"id":       extractUTF16LEString(filePath, 0x1C, 0x11B),
		"name":     extractUTF16LEString(filePath, 0x130, 0x337),
		"category": extractUTF16LEString(filePath, 0x338, 0x53F),
		"remark":   extractUTF16LEString(filePath, 0x540, 0xD3F),
		"example":  extractExampleString(filePath, 0xD40, 0x153F),
		"count":    extractEntryCount(filePath),
		"source":   source,
		"header":   header,
	}

	dictInfo["timestamp"], dictInfo["timestamp_raw"] = extractTimestamp(filePath)

	out, _ := json.MarshalIndent(dictInfo, "", "  ")
	fmt.Println(string(out))
}