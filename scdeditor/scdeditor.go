package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"strings"
)

// 偏移常量定义
const (
	magicBytesStart = 0x004
	magicBytesEnd   = 0x006
	idStart         = 0x01C
	idEnd           = 0x05B
	nameStart       = 0x130
	nameEnd         = 0x337
	categoryStart   = 0x338
	categoryEnd     = 0x53F
	remarkStart     = 0x540
	remarkEnd       = 0xD3F
)

// 写 magic bytes
func writeMagicBytes(filePath string, checkedOfficial bool) {
	if !checkedOfficial {
		return
	}
	f, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		fmt.Printf("无法打开文件: %s, 错误: %v\n", filePath, err)
		return
	}
	defer f.Close()

	_, err = f.Seek(magicBytesStart, 0)
	if err != nil {
		fmt.Printf("写入 magic bytes 失败：%v\n", err)
		return
	}

	_, err = f.Write([]byte("DCS"))
	if err != nil {
		fmt.Printf("写入 magic bytes 失败：%v\n", err)
	}
}

// 写字符串
func writeString(filePath string, offsetStart, offsetEnd int, input string, fieldName string) {
	size := offsetEnd - offsetStart
	if size <= 0 {
		fmt.Printf("跳过写入 %s：无效偏移区间（0x%X - 0x%X）\n", fieldName, offsetStart, offsetEnd)
		return
	}

	// 转换为 UTF-16LE 编码
	utf16Data := []byte{}
	for _, r := range input {
		buf := make([]byte, 2)
		binary.LittleEndian.PutUint16(buf, uint16(r))
		utf16Data = append(utf16Data, buf...)
	}

	if len(utf16Data) > size {
		fmt.Printf("跳过写入 %s：内容超出限制（%d > %d 字节）\n", fieldName, len(utf16Data), size)
		return
	}

	f, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		fmt.Printf("写入 %s 失败：%v\n", fieldName, err)
		return
	}
	defer f.Close()

	// 清空该区域
	_, _ = f.Seek(int64(offsetStart), 0)
	_, _ = f.Write(make([]byte, size))

	// 再写入实际数据
	_, _ = f.Seek(int64(offsetStart), 0)
	_, err = f.Write(utf16Data)
	if err != nil {
		fmt.Printf("写入 %s 失败：%v\n", fieldName, err)
	}
}

func isValidString(s string) bool {
	return strings.TrimSpace(s) != ""
}

func main() {
	// 参数定义
	filePath := flag.String("file", "", "词库文件路径")
	id := flag.String("id", "", "词库ID")
	name := flag.String("name", "", "词库名称")
	category := flag.String("category", "", "词库类别")
	remark := flag.String("remark", "", "词库备注")
	official := flag.Bool("official", false, "标记为官方词库")

	// 手动处理短参数映射
	for i, arg := range os.Args {
		switch arg {
		case "-f":
			os.Args[i] = "--file"
		case "-i":
			os.Args[i] = "--id"
		case "-n":
			os.Args[i] = "--name"
		case "-c":
			os.Args[i] = "--category"
		case "-r":
			os.Args[i] = "--remark"
		case "-o":
			os.Args[i] = "--official"
		}
	}

	// 自定义帮助输出
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "用法: %s [选项]\n\n", os.Args[0])
		fmt.Println("选项:")
		fmt.Println("  -f, --file <路径>        词库文件路径 (必填)")
		fmt.Println("  -i, --id <字符串>        词库ID")
		fmt.Println("  -n, --name <字符串>      词库名称")
		fmt.Println("  -c, --category <字符串>  词库类别")
		fmt.Println("  -r, --remark <字符串>    词库备注")
		fmt.Println("  -o, --official           标记为官方词库")
		fmt.Println("  -h, --help               显示帮助信息")
	}

	flag.Parse()

	// 如果没有任何参数，显示帮助
	if len(os.Args) == 1 {
		flag.Usage()
		os.Exit(0)
	}

	if *filePath == "" {
		flag.Usage()
		os.Exit(1)
	}

	if _, err := os.Stat(*filePath); os.IsNotExist(err) {
		fmt.Printf("文件不存在：%s\n", *filePath)
		os.Exit(1)
	}

	// 写入数据
	writeMagicBytes(*filePath, *official)

	if isValidString(*id) {
		writeString(*filePath, idStart, idEnd, strings.TrimSpace(*id), "id")
	}

	if isValidString(*name) {
		writeString(*filePath, nameStart, nameEnd, strings.TrimSpace(*name), "name")
	}

	if isValidString(*category) {
		writeString(*filePath, categoryStart, categoryEnd, strings.TrimSpace(*category), "category")
	}

	if isValidString(*remark) {
		writeString(*filePath, remarkStart, remarkEnd, strings.TrimSpace(*remark), "remark")
	}
}