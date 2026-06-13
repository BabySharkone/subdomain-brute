package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"subdomain-brute/internal"
	"time"
)

func main() {
	// 定义命令行参数
	domain := flag.String("d", "", "目标域名，例如 example.com")
	wordlistPath := flag.String("w", "wordlists/subdomains.txt", "字典文件路径")
	workers := flag.Int("t", 50, "并发数（默认 50）")
	outputPath := flag.String("o", "", "输出文件路径（可选）")
	outputFormat := flag.String("f", "txt", "输出格式：txt|json|csv（默认 txt）")
	flag.Parse()

	// 必填参数检查
	if *domain == "" {
		fmt.Println("用法：subdomain-brute -d example.com [-w wordlist.txt] [-t 50] [-o output.txt] [-f txt|json|csv]")
		fmt.Println("\n选项说明：")
		fmt.Println("  -d string  目标域名（必填）")
		fmt.Println("  -w string  字典文件路径（默认：wordlists/subdomains.txt）")
		fmt.Println("  -t int     并发数（默认：50）")
		fmt.Println("  -o string  输出文件路径（可选）")
		fmt.Println("  -f string  输出格式：txt|json|csv（默认：txt）")
		os.Exit(1)
	}

	// 读取字典文件
	file, err := os.Open(*wordlistPath)
	if err != nil {
		fmt.Printf("打开字典文件失败：%v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	var prefixes []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		prefix := strings.TrimSpace(scanner.Text())
		if prefix != "" && !strings.HasPrefix(prefix, "#") {
			prefixes = append(prefixes, prefix)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("读取字典文件失败：%v\n", err)
		os.Exit(1)
	}

	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("子域名爆破工具\n")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("目标域名: %s\n", *domain)
	fmt.Printf("字典数量: %d\n", len(prefixes))
	fmt.Printf("⚡ 并发数: %d\n", *workers)
	fmt.Printf("输出格式: %s\n\n", *outputFormat)

	// 验证输出格式
	validFormats := map[string]bool{"txt": true, "json": true, "csv": true}
	if !validFormats[*outputFormat] {
		fmt.Printf("无效的输出格式：%s，支持：txt|json|csv\n", *outputFormat)
		os.Exit(1)
	}

	start := time.Now()
	results := internal.Bruteforce(*domain, prefixes, *workers)
	elapsed := time.Since(start)

	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("扫描完成！\n")
	fmt.Println(strings.Repeat("=", 60))

	// 导出到文件（如果指定了）
	if *outputPath != "" {
		err := exportResults(results, *outputPath, *outputFormat)
		if err != nil {
			fmt.Printf("导出结果失败：%v\n", err)
		} else {
			fmt.Printf("结果已保存到 %s\n\n", *outputPath)
		}
	}

	// 显示结果摘要
	if len(results) == 0 {
		fmt.Println("未发现任何有效子域名")
	} else {
		fmt.Printf("发现 %d 个子域名：\n\n", len(results))
		for _, r := range results {
			fmt.Printf("[+] %-30s -> %s\n", r.Subdomain, strings.Join(r.IPs, ", "))
		}
	}

	fmt.Println()
	fmt.Printf("扫描耗时: %v\n", elapsed)
	fmt.Printf("平均耗时: %v per subdomain\n", elapsed/time.Duration(len(prefixes)))
}

func exportResults(results []internal.Result, outputPath, format string) error {
	switch format {
	case "json":
		return internal.ExportJSON(results, outputPath)
	case "csv":
		return internal.ExportCSV(results, outputPath)
	case "txt":
		fallthrough
	default:
		return internal.ExportTXT(results, outputPath)
	}
}

