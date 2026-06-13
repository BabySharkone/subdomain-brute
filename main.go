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
	domain := flag.String("d", "", "目标域名，例如 baidu,com")
	wordlistPath := flag.String("w", "wordlists/subdomains.txt", "字典文件路径")
	workers := flag.Int("t", 20, "并发数")
	outputPath := flag.String("o", "", "输出文件路径（可选）")
	flag.Parse()

	// 必填参数检查
	if *domain == "" {
		fmt.Println("用法：subdomain-brute -d example.com [-w wordlist.txt] [-t 20]")
		os.Exit(1)
	}

	file, err := os.Open(*wordlistPath)
	if err != nil {
		fmt.Println("打开字典文件失败：", err)
		return
	}

	defer file.Close()

	var prefixes []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		prefix := strings.TrimSpace(scanner.Text())
		if prefix != "" {
			prefixes = append(prefixes, prefix)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("读取字典文件失败：", err)
		return
	}

	fmt.Printf("目标域名: %s | 字典: %d 条 | 并发数: %d\n\n", *domain, len(prefixes), *workers)

	start := time.Now()
	results := internal.Bruteforce(*domain, prefixes, *workers)
	elapsed := time.Since(start)

	if *outputPath != "" {
		f, err := os.Create(*outputPath)
		if err != nil {
			fmt.Println("创建输出文件失败:", err)
		} else {
			defer f.Close()
			for _, r := range results {
				fmt.Fprintf(f, "%s -> %v\n", r.Subdomain, r.IPs)
			}
			fmt.Println()
			fmt.Printf("结果已保存到 %s\n", *outputPath)
		}
	}
	for _, r := range results {
		fmt.Printf("[+] %s -> %v\n", r.Subdomain, r.IPs)
	}
	fmt.Printf("\n共发现 %d 个子域名，耗时 %v\n", len(results), elapsed)
}
