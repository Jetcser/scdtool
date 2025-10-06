package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/axgle/mahonia"
	marker "github.com/nopdan/pinyin-marker"
	"github.com/saintfish/chardet"
)

var (
	chineseRegex = regexp.MustCompile(`\p{Han}+`)
	splitRegex   = regexp.MustCompile(`[,\s，]+`)
	m            = marker.New()
)

// 加载 ~/.config/scdtool/override.txt
// 支持两种格式：
// 覆写: 词组=拼音
// 追加: 单字+=拼音
func loadOverrides() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("无法获取用户主目录:", err)
		return
	}
	configDir := filepath.Join(home, ".config", "scdtool")
	overrideFile := filepath.Join(configDir, "override.txt")

	// 确保目录存在
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			fmt.Println("无法创建配置目录:", err)
			return
		}
	}

	// 如果文件不存在，写入示例
	if _, err := os.Stat(overrideFile); os.IsNotExist(err) {
		example := "会计=kuai ji\n似+=shi\n"
		if err := os.WriteFile(overrideFile, []byte(example), 0644); err != nil {
			fmt.Println("无法创建 override.txt:", err)
			return
		}
		fmt.Println("已创建示例覆盖文件:", overrideFile)
	}

	// 读取并应用规则
	f, err := os.Open(overrideFile)
	if err != nil {
		fmt.Println("无法打开 override.txt:", err)
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// 追加规则：单字+=拼音
		if strings.Contains(line, "+=") {
			parts := strings.SplitN(line, "+=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				runes := []rune(key)
				if len(runes) == 1 {
					m.Append(runes[0], val) // ✅ rune
				} else {
					fmt.Println("警告: Append 仅支持单字，跳过:", key)
				}
			}
			continue
		}
		// 覆写规则：词组=拼音
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				m.Overwrite(key, val)
			}
		}
	}
}

// 检测文件编码
func detectEncoding(filePath string, sampleSize int) string {
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("无法打开文件: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	buf := make([]byte, sampleSize)
	n, _ := f.Read(buf)

	detector := chardet.NewTextDetector()
	result, err := detector.DetectBest(buf[:n])
	if err != nil {
		fmt.Printf("检测编码失败: %v\n", err)
		os.Exit(1)
	}

	encoding := strings.ToUpper(result.Charset)
	fmt.Printf("检测到输入文件编码：%s \n", encoding)

	if encoding == "GB2312" || encoding == "GBK" {
		fmt.Println("使用 GB18030 解码以兼容 GB2312 / GBK")
		return "GB18030"
	}
	return encoding
}

// 提取中文
func extractChinese(text string) string {
	return strings.Join(chineseRegex.FindAllString(text, -1), "")
}

// 转换为拼音（每个音节前加单引号，汉字原样保留）
func convertToPinyinLine(text string) string {
	chinese := extractChinese(text)
	if len([]rune(chinese)) <= 1 {
		return ""
	}

	pyList := m.Mark(chinese) // []string
	if len(pyList) == 0 {
		return ""
	}

	// 每个拼音音节前加单引号
	for i, p := range pyList {
		pyList[i] = "'" + p
	}

	return strings.Join(pyList, "") + " " + chinese
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

	resultLines := make([]string, 0)
	uniqueLines := make(map[string]struct{}) // 去重用

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(decoder.ConvertString(scanner.Text()))
		if line == "" {
			continue
		}
		groups := splitRegex.Split(line, -1)
		for _, group := range groups {
			group = strings.TrimSpace(group)
			if group == "" {
				continue
			}
			converted := convertToPinyinLine(group)
			if converted != "" {
				// 去重
				if _, exists := uniqueLines[converted]; !exists {
					uniqueLines[converted] = struct{}{}
					resultLines = append(resultLines, converted)
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
		os.Exit(1)
	}

	out, err := os.Create(outputPath)
	if err != nil {
		fmt.Printf("写入文件失败: %v\n", err)
		os.Exit(1)
	}
	defer out.Close()

	writer := bufio.NewWriter(out)
	for _, line := range resultLines {
		_, _ = writer.WriteString(line + "\n") // 直接写 UTF-8
	}
	writer.Flush()

	fmt.Printf("转换完成，输出文件：%s\n", outputPath)
	fmt.Printf("处理词条数量：%d\n", len(resultLines))
}

func main() {
	loadOverrides() // 加载覆盖拼音规则

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
