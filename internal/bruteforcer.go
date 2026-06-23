package internal

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Bruteforce 并发爆破子域名 （DNS 检测）
// domain: 目标域名
// prefixes: 字典前缀列表
// workers: 并发数
func Bruteforce(domain string, prefixes []string, workers int) []Result {
	return BruteforceWithHTTP(domain, prefixes, workers, false, 0)
}

// BruteforceWithHTTP 并发爆破子域名（支持 HTTP 检测）
func BruteforceWithHTTP(domain string, prefixes []string, workers int, checkHTTP bool, httpTimeout time.Duration) []Result {
	jobs := make(chan string, len(prefixes))   // 任务通道，存储待查询的子域名前缀
	result := make(chan Result, len(prefixes)) // 结果通道，存储查询成功的子域名结果

	var wg sync.WaitGroup
	total := len(prefixes)

	// 在扫描前，先对主域名进行一次泛解析检测
    fmt.Println("[*] 正在检测目标域名是否存在泛解析...")
    isWildcard, wildcardIPs := DetectWildcard(domain)
    if isWildcard {
        fmt.Printf("[!] 警告：检测到目标存在泛解析！将自动过滤以下黑名单 IP: %v\n", wildcardIPs)
    } else {
        fmt.Println("[+] 目标域名没有泛解析，环境干净。")
    }

	var httpDetector *HTTPDetector
	if checkHTTP {
		httpDetector = NewHTTPDetector(httpTimeout)
	}

	// 启动 workers 个 goroutine
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker(domain, jobs, result, &wg, httpDetector, isWildcard, wildcardIPs)
	}

	// 把所有前缀放入 jobs channel
	for _, prefix := range prefixes {
		jobs <- prefix
	}
	close(jobs) // 关闭channel，worker才知道没有更多任务了

	// 重置全局进度计数器
	atomic.StoreInt32(&completed, 0)

	stopProgress := make(chan struct{})
	go progressPrinter(&completed, total, stopProgress)

	// 等待所有worker完成
	go func() {
		wg.Wait()
		close(result)
		close(stopProgress)
	}()

	// 收集结果并去重
	var allResult []Result
	seen := make(map[string]bool)	// 引入哈希表，用于存储已经发现的子域名

	for r := range result {
		if _, exists := seen[r.Subdomain]; !exists {
			seen[r.Subdomain] = true
			allResult = append(allResult, r)
		}
	}

	return allResult
}


var completed int32
// worker 从jobs取任务，查询DNS，把成功的结果放入results
// jobs <-chan string    	只读channel：worker只能从这里读
// results chan<- Result 	只写channel：worker只能往这里写
func worker(domain string, jobs <-chan string, result chan<- Result, wg *sync.WaitGroup, httpDetector *HTTPDetector, isWildcard bool, wildcardIPs map[string]bool) {
	defer wg.Done() // 任务完成时调用 Done

	for prefix := range jobs {
		subdomain := prefix + "." + domain

		// 1. 做DNS检测
		ips, err := ResolveDNS(subdomain)
		atomic.AddInt32(&completed, 1)	// 直接自增
		if err != nil {
			continue // 查询失败，跳过
		}
		// 如果开启了泛解析，检查返回的 IP 是否全在黑名单里
        if isWildcard {
            allJunk := true
            for _, ip := range ips {
                if !wildcardIPs[ip] {
                    allJunk = false // 只要有一个 IP 不在黑名单里，说明这个子域名可能是真实的
                    break
                }
            }
            if allJunk {
                continue // 如果所有 IP 都是泛解析的黑名单 IP，直接跳过该子域名
            }
        }

		// 2. 做 HTTP 检测
		httpStatus := 0
		isValid := true

		if httpDetector != nil {
			status, err := httpDetector.CheckHTTP(subdomain)
			if err == nil {
				httpStatus = status
				isValid = IsValidHTTPStatus(status)
			}else {
				isValid = false
			}

			// 如果启用了 HTTP 检测但检测失败，则跳过
			if !isValid {
				continue
			}
		}

		result <- Result{
			Subdomain: 	subdomain,
			IPs:       	ips,
			HTTPStatus: httpStatus,
			IsValid:	isValid,
		}
	}
}

// 新增一个专门负责打印进度的函数
func progressPrinter(completed *int32, total int, stop chan struct{}) {
    ticker := time.NewTicker(200 * time.Millisecond) // 每200ms刷新一次，不废CPU
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            done := atomic.LoadInt32(completed)
            fmt.Printf("\r[*] 扫描进度: %d/%d (%.2f%%)", done, total, float64(done)/float64(total)*100)
        case <-stop:
            // 扫描结束，打印最后一次 100% 并退出
            fmt.Printf("\r[*] 扫描进度: %d/%d (100.00%%)\n", total, total)
            return
        }
    }
}
