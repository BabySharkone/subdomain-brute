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

	var httpDetector *HTTPDetector
	if checkHTTP {
		httpDetector = NewHTTPDetector(httpTimeout)
	}

	// 启动 workers 个 goroutine
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker(domain, jobs, result, &wg, httpDetector)
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
func worker(domain string, jobs <-chan string, result chan<- Result, wg *sync.WaitGroup, httpDetector *HTTPDetector) {
	defer wg.Done() // 任务完成时调用 Done

	for prefix := range jobs {
		subdomain := prefix + "." + domain

		// 1. 做DNS检测
		ips, err := ResolveDNS(subdomain)
		atomic.AddInt32(&completed, 1)	// 直接自增
		if err != nil {
			continue // 查询失败，跳过
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
