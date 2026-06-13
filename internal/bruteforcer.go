package internal

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// Bruteforce 并发爆破子域名
// domain: 目标域名
// prefixes: 字典前缀列表
// workers: 并发数
func Bruteforce(domain string, prefixes []string, workers int) []Result {
	jobs := make(chan string, len(prefixes))   // 任务通道，存储待查询的子域名前缀
	result := make(chan Result, len(prefixes)) // 结果通道，存储查询成功的子域名结果

	var wg sync.WaitGroup

	// 启动 workers 个 goroutine
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker(domain, jobs, result, 0, &wg)
	}

	// b把所有前缀放入 jobs channel
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

// worker 从jobs取任务，查询DNS，把成功的结果放入results
// jobs <-chan string    	只读channel：worker只能从这里读
// results chan<- Result 	只写channel：worker只能往这里写
var completed int32

func worker(domain string, jobs <-chan string, result chan<- Result, total int, wg *sync.WaitGroup) {
	defer wg.Done() // 任务完成时调用 Done

	for prefix := range jobs {
		subdomain := prefix + "." + domain

		ips, err := ResolveDNS(subdomain)
		done := atomic.AddInt32(&completed, 1)
		fmt.Printf("\r进度: %d/%d", done, total)
		if err != nil {
			continue // 查询失败，跳过
		}
		result <- Result{
			Subdomain: subdomain,
			IPs:       ips,
		}
	}
}
