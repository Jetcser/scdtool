package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
	
	"github.com/axgle/mahonia"
	"github.com/mozillazg/go-pinyin"
	"github.com/saintfish/chardet"
)

var (
	// 匹配所有中文
	chineseRegex = regexp.MustCompile(`[\p{Han}]+`)
	// 存储覆盖规则
	overrideDict = map[string]string{}
)

// 加载覆盖拼音规则
func loadOverrides() {
	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".config", "scdtool")
	overrideFile := filepath.Join(configDir, "override.txt")

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		_ = os.MkdirAll(configDir, 0755)
	}

	if _, err := os.Stat(overrideFile); os.IsNotExist(err) {
		example := "会计=kuai ji\n柏临河=bo lin he\n"
		_ = os.WriteFile(overrideFile, []byte(example), 0644)
	}

	f, err := os.Open(overrideFile)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			overrideDict[key] = val
		}
	}
}

// 检测文件编码 (优化版)
func detectEncoding(filePath string, sampleSize int) string {
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("无法打开文件: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	// 读取样本数据，但这次我们不限制只读 sampleSize
	// 因为 utf8.ValidString 需要完整的数据才能最准确
	// 我们限制一个合理的上限，例如 512KB
	const readLimit = 512 * 1024
	r := io.LimitReader(f, readLimit)
	buf, err := io.ReadAll(r)
	if err != nil {
		fmt.Printf("读取文件样本失败: %v\n", err)
		os.Exit(1)
	}

	// 1. 优先检测BOM (Byte Order Mark)
	// 这是最准确的编码判断依据
	if bytes.HasPrefix(buf, []byte{0xEF, 0xBB, 0xBF}) {
		fmt.Println("检测到UTF-8 BOM")
		return "UTF-8"
	}
	// 可以根据需要添加UTF-16等其他BOM的检测
	// if bytes.HasPrefix(buf, []byte{0xFF, 0xFE}) { return "UTF-16LE" }
	// if bytes.HasPrefix(buf, []byte{0xFE, 0xFF}) { return "UTF-16BE" }
	
	// 2. 快速验证是否为合法的UTF-8
	// 对于纯ASCII或标准的UTF-8文件，这个检查非常快且100%准确
	if utf8.Valid(buf) {
		fmt.Println("检测为合法的无BOM UTF-8")
		return "UTF-8"
	}

	// 3. 如果不是UTF-8，再使用 chardet 进行统计学猜测
	fmt.Println("非UTF-8编码，尝试使用 chardet 进行检测...")
	detector := chardet.NewTextDetector()
	// 使用部分样本进行检测，避免样本过大影响性能
	detectSample := buf
	if len(detectSample) > sampleSize {
		detectSample = detectSample[:sampleSize]
	}
	result, err := detector.DetectBest(detectSample)
	if err != nil {
		fmt.Printf("检测编码失败: %v\n", err)
		os.Exit(1) // or return a default encoding
	}

	encoding := strings.ToUpper(result.Charset)
	fmt.Printf("Chardet检测到输入文件编码：%s\n", encoding)

	// 针对中文环境的常见编码进行修正
	switch encoding {
	case "GB2312", "GBK", "HZ-GB-2312": // GBK系列统一使用GB18030解码
		return "GB18030"
	case "EUC-KR", "SHIFT_JIS": // 对于韩文和日文的误判，强制认为是GB18030
		fmt.Println("检测结果为日韩编码，在中文环境下，强制使用 GB18030 解码")
		return "GB18030"
	case "BIG5": // 如果需要支持繁体中文
		return "BIG5"
	default:
		// 对于其他未知的编码，可以返回一个默认值或 chardet 的结果
		return encoding
	}
}

// 提取中文词条，长度 >= 2
func extractChinesePhrases(text string) []string {
	matches := chineseRegex.FindAllString(text, -1)
	var results []string
	for _, m := range matches {
		if len([]rune(m)) >= 2 {
			results = append(results, m)
		}
	}
	return results
}

// 转换为拼音
func convertToPinyin(phrase string) string {
	// 1. 准备和排序覆盖规则的 key，按长度倒序，确保优先匹配长词
	keys := make([]string, 0, len(overrideDict))
	for k := range overrideDict {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return len([]rune(keys[i])) > len([]rune(keys[j]))
	})

	// 准备 go-pinyin 的参数
	a := pinyin.NewArgs()
	a.Style = pinyin.Normal
	a.Heteronym = false

	// --- 全新的扫描匹配逻辑 ---
	var resultPinyins []string
	phraseRunes := []rune(phrase)
	// i 是当前在 phrase 中的扫描位置（按 rune 索引）
	for i := 0; i < len(phraseRunes); {
		// 检查从当前位置 i 开始，是否能匹配任何一个覆盖规则
		matched := false
		for _, key := range keys {
			keyRunes := []rune(key)
			// 如果剩余词条的长度足够匹配当前 key
			if i+len(keyRunes) <= len(phraseRunes) {
				// 提取子串进行比较
				subPhrase := string(phraseRunes[i : i+len(keyRunes)])
				if subPhrase == key {
					// 匹配成功！
					pys := strings.Fields(overrideDict[key])
					resultPinyins = append(resultPinyins, pys...)
					// 将扫描位置向后移动匹配到的 key 的长度
					i += len(keyRunes)
					matched = true
					// 已经找到了最长的匹配（因为 keys 是排序过的），跳出内层循环
					break
				}
			}
		}

		// 如果在当前位置 i，所有覆盖规则都未匹配成功
		if !matched {
			// 只处理当前单个字符
			char := string(phraseRunes[i])
			pys := pinyin.Pinyin(char, a)
			if len(pys) > 0 && len(pys[0]) > 0 {
				resultPinyins = append(resultPinyins, pys[0][0])
			}
			// 将扫描位置向后移动 1
			i++
		}
	}

	// 格式化最终输出
	for i, p := range resultPinyins {
		resultPinyins[i] = "'" + p
	}
	return strings.Join(resultPinyins, "") + " " + phrase
}

// 处理文件
func processFile(inputPath string) {
	base := strings.TrimSuffix(inputPath, filepath.Ext(inputPath))
	outputPath := base + "_sg.txt"

	encoding := detectEncoding(inputPath, 10000)
	decoder := mahonia.NewDecoder(encoding)
	if decoder == nil {
		fmt.Printf("不支持的编码: %s\n", encoding)
		os.Exit(1)
	}

	file, err := os.Open(inputPath)
	if err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	resultLines := []string{}
	uniqueLines := map[string]struct{}{}

	scanner := bufio.NewScanner(file)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(decoder.ConvertString(scanner.Text()))
		if line == "" {
			continue
		}
		phrases := extractChinesePhrases(line)
		for _, p := range phrases {
			out := convertToPinyin(p)
			// fmt.Printf("第 %d 行中文词条: %s -> %s\n", lineNo, p, out) // 调试信息
			if _, exists := uniqueLines[out]; !exists {
				uniqueLines[out] = struct{}{}
				resultLines = append(resultLines, out)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
		os.Exit(1)
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		fmt.Printf("写入文件失败: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)
	for _, line := range resultLines {
		writer.WriteString(line + "\n")
	}
	writer.Flush()

	fmt.Printf("转换完成，输出文件：%s\n", outputPath)
	fmt.Printf("处理词条数量：%d\n", len(resultLines))
}

func main() {
	loadOverrides()

	if len(os.Args) != 2 {
		fmt.Println("用法：txtmaker 输入文件.txt")
		os.Exit(1)
	}
	inputFile := os.Args[1]
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		fmt.Printf("错误：文件不存在：%s\n", inputFile)
		os.Exit(1)
	}
	processFile(inputFile)
}
