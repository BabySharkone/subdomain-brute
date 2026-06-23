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
	httpCheck := flag.Bool("http",false,"启用HTTP检测（可选）")
	httpTimeout := flag.Int("http-timeout",5,"HTTP 请求超时（秒）")
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
		fmt.Println("  -http	  启用HTTP检测（可选，更慢但更准确）")
		fmt.Println("  -http-time int HTTP 超时秒数（默认：5）")
		os.Exit(1)
	}

	// 将用户输入的域名清洗成最纯净的“标准域名”格式
	d := strings.TrimSpace(*domain)				 // 去掉字符串两端看不见的空格、制表符或换行
	d = strings.ToLower(d)						 // 域名不区分大小写，统一转小写
	d = strings.TrimPrefix(d,"https://")		 // 去掉协议头
	d = strings.TrimPrefix(d,"http://")			 // 去掉协议头
	// 1. 先切掉可能带有的路径（根据 / 分割）
	if idx := strings.Index(d,"/"); idx != -1 {
		d = d[:idx]
	}
	// 2. 切掉可能带有的端口号（根据 : 分割）
	if idx := strings.Index(d, ":"); idx != -1 {
    	d = d[:idx]
	}
	*domain = d

	// 防止误把选项当字典路径
    if strings.HasPrefix(*wordlistPath, "-") {
        fmt.Fprintf(os.Stderr, "错误：-w 参数格式错误，不能以 '-' 开头。\n")
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
	// 防止超大行导致读取中断
    const maxCapacity = 1024 * 1024
    buf := make([]byte, maxCapacity)
    scanner.Buffer(buf, maxCapacity)
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
	fmt.Printf("并发数: %d\n", *workers)
	fmt.Printf("输出格式: %s\n\n", *outputFormat)

	// 验证输出格式
	validFormats := map[string]bool{"txt": true, "json": true, "csv": true}
	if !validFormats[*outputFormat] {
		fmt.Printf("无效的输出格式：%s，支持：txt|json|csv\n", *outputFormat)
		os.Exit(1)
	}

	start := time.Now()
	var results []internal.Result
	if *httpCheck {
		// 启用HTTP检测
		results = internal.BruteforceWithHTTP(*domain,prefixes,*workers,true,time.Duration(*httpTimeout)*time.Second)
	}else{
		// 不启用HTTP
		results = internal.Bruteforce(*domain, prefixes, *workers)
	}
	elapsed := time.Since(start)

	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("扫描完成！\n")

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
			if r.HTTPStatus > 0 {
				fmt.Printf("[+] %-30s -> %s (HTTP: %d)\n", r.Subdomain, strings.Join(r.IPs, ", "), r.HTTPStatus)
			}else {
				fmt.Printf("[+] %-30s -> %s\n", r.Subdomain, strings.Join(r.IPs, ", "))
			}
		}
	}

	fmt.Println()
	fmt.Printf("扫描耗时: %v\n", elapsed)
	// 爆破数量为 0 时的除零崩溃防御
	if len(prefixes) > 0 {
    	fmt.Printf("平均耗时: %v per subdomain\n", elapsed/time.Duration(len(prefixes)))
	} else {
    	fmt.Println("平均耗时: N/A (有效字典数量为 0)")
	}
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

