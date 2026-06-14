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
		go worker(domain, jobs, result, total, &wg, httpDetector)
	}

	// 把所有前缀放入 jobs channel
	for _, prefix := range prefixes {
		jobs <- prefix
	}
	close(jobs) // 关闭channel，worker才知道没有更多任务了

	// 等待所有worker完成
	go func() {
		wg.Wait()
		close(result)
	}()

	// 收集结果
	var allResult []Result
	for r := range result {
		allResult = append(allResult, r)
	}

	return allResult
}


var completed int32
// worker 从jobs取任务，查询DNS，把成功的结果放入results
// jobs <-chan string    	只读channel：worker只能从这里读
// results chan<- Result 	只写channel：worker只能往这里写
func worker(domain string, jobs <-chan string, result chan<- Result, total int, wg *sync.WaitGroup, httpDetector *HTTPDetector) {
	defer wg.Done() // 任务完成时调用 Done

	for prefix := range jobs {
		subdomain := prefix + "." + domain

		// 1. 做DNS检测
		ips, err := ResolveDNS(subdomain)
		done := atomic.AddInt32(&completed, 1)
		fmt.Printf("\r进度: %d/%d", done, total)
		if err != nil {
			continue // 查询失败，跳过
		}

		// HTTP 检测
		httpStatus := 0
		isValid := true

		if httpDetector != nil {
			// 2. 做 HTTP 检测
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
