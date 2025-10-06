package main

import (
	"bufio"
	_ "embed"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/net/html/charset"
)

//go:embed pinyin.bin
var pinyin []byte

var syllMap = list2map(list)

func list2map(list []string) map[string]int {
	m := make(map[string]int)
	for i := 0; i < len(list); i++ {
		m[list[i]] = i
	}
	return m
}

func Make(name string) {
	// 按行读取 name 文件
	in, err := os.Open(name)
	if err != nil {
		panic(err)
	}
	defer in.Close()

	fileName := strings.TrimSuffix(filepath.Base(name), filepath.Ext(name))
	out, err := os.Create(strings.TrimSuffix(name, filepath.Ext(name)) + ".scel")
	if err != nil {
		panic(err)
	}
	defer out.Close()
	wr := bufio.NewWriter(out)

	// utf-16le
	utf16, _ := charset.Lookup("utf-16le")
	enc := utf16.NewEncoder()

	header := make([]byte, 0x120)
	copy(header, []byte{0x40, 0x15, 0x00, 0x00})
	// 文件头的值代表的词库类型：0xD2 0x6D为自定义词库；0x44 0x43为官方词库
	// 搜狗输入法Linux NG版仅支持导入官方词库，为了兼容，修改0xD2 0x6D为0x44 0x43
	copy(header[4:], []byte{0x44, 0x43, 0x53, 0x01})
	copy(header[8:], []byte{1, 0, 0, 0})
	copy(header[12:], make([]byte, 16))

	// 生成随机数
	rand_num := rand.Uint32()
	str := fmt.Sprintf("L%d", uint16(rand_num))
	b, _ := enc.Bytes([]byte(str))
	copy(header[0x1C:], b)

	// 时间戳
	now := uint32(time.Now().Unix())
	binary.LittleEndian.PutUint32(header[0x11C:], now)

	wr.Write(header)

	// 填充 0 直到 0x1540
	wr.Write(make([]byte, 0x1540-0x120))

	// 写拼音表
	wr.Write(pinyin)

	// 读取文件内容
	cSize, cCount, wSize, wCount := 0, 0, 0, 0

	//!TODO 暂未使用 合并相同拼音的词条
	code_map := make(map[string]struct{})

	// 示例词
	examples := make([]string, 0)

	buf := bufio.NewScanner(ConvertReader(in))
	for buf.Scan() {
		line := buf.Text()
		items := strings.Split(line, " ")
		if len(items) < 2 {
			continue
		}

		//!TODO
		wr.Write([]byte{0x01, 0x00})

		code, word := items[0], items[1]
		code = strings.TrimPrefix(code, "'")
		sylls := strings.Split(code, "'")

		// 拼音占用字节数
		b = make([]byte, 2)
		binary.LittleEndian.PutUint16(b, uint16(len(sylls)*2))
		wr.Write(b)
		for _, s := range sylls {
			idx := syllMap[s]
			b = make([]byte, 2)
			binary.LittleEndian.PutUint16(b, uint16(idx))
			wr.Write(b)
		}

		cSize += len(sylls)*2 + 2
		cCount++

		//!TODO
		// if _, ok := code_map[code]; !ok {
		// 	cSize += len(sylls)*2 + 2
		// 	cCount++
		// 	code_map[code] = struct{}{}
		// }

		// 添加示例词
		if len(examples) < 6 {
			examples = append(examples, word)
		}
		b, _ := enc.Bytes([]byte(word))
		b2 := make([]byte, 2)
		binary.LittleEndian.PutUint16(b2, uint16(len(b)))
		wr.Write(b2)
		wr.Write(b)
		wr.Write([]byte{0x0A, 0x00, 0x2D, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

		wSize += len(b) + 2
		wCount++
	}
	wr.Flush()

	// 回写一些关键信息
	out.Seek(0x1540, 0)

	// 校验和
	chksum := CheckSumStream(out)

	for i, v := range chksum {
		b := make([]byte, 4)
		binary.LittleEndian.PutUint32(b, v)
		out.WriteAt(b, int64(i*4+0xC))
	}

	b = make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(cCount))
	out.WriteAt(b, 0x120)
	binary.LittleEndian.PutUint32(b, uint32(wCount))
	out.WriteAt(b, 0x124)
	binary.LittleEndian.PutUint32(b, uint32(cSize))
	out.WriteAt(b, 0x128)
	binary.LittleEndian.PutUint32(b, uint32(wSize))
	out.WriteAt(b, 0x12C)

	s, _ := enc.String(fileName)
	out.WriteAt([]byte(s), 0x130)
	s, _ = enc.String("本地")
	out.WriteAt([]byte(s), 0x338)
	s, _ = enc.String("由 scel-maker 生成的细胞词库")
	out.WriteAt([]byte(s), 0x540)
	s, _ = enc.String(strings.Join(examples, "   "))
	out.WriteAt([]byte(s), 0xD40)

	_ = code_map
	
	// 打印输出词库的绝对路径
	outPath := strings.TrimSuffix(name, filepath.Ext(name)) + ".scel"
	absPath, _ := filepath.Abs(outPath)
	fmt.Println("生成细胞词库：", absPath)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: scel-maker <input>")
		os.Exit(1)
	}
	Make(os.Args[1])
}
