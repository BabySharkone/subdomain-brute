package internal

import (
	"github.com/miekg/dns"
	"context"
	"strings"
	"time"
	"crypto/rand"
	"math/big"
)

var junkIPs = map[string]bool{
	"0.0.0.1":   true,
	"127.0.0.1": true,
	"0.0.0.0":   true,
}

// 固定的公共 DNS 服务器（可以绕过系统本地 DNS 缓存）
const DNSServer = "223.5.5.5:53"

// ResolveDNS 负责查询一个域名的 IP（使用 miekg/dns 高性能异步解析
// 如果查询成功，返回 IP 列表；如果域名不存在，返回错误
func ResolveDNS(domain string) ([]string, error) {
	// 确保域名以点号结束（DNS 规范要求）
	if !strings.HasSuffix(domain, ".") {
		domain = domain + "."
	}

	c := dns.Client{
		Net:     "udp",
		Timeout: 2 * time.Second, // 设置超时，防止死等
	}

	m := dns.Msg{}
	// 设置查询类型为 A 记录（IPv4 地址）
	m.SetQuestion(domain, dns.TypeA)

	// 直接向指定的 DNS 服务器发送 UDP 包
	r, _, err := c.Exchange(&m, DNSServer)
	if err != nil {
		return nil, err
	}
	
	var validIPs []string
	// 解析返回的结果
	for _, ans := range r.Answer {
		if aRecord, ok := ans.(*dns.A); ok {
			ip := aRecord.A.String()
			if !junkIPs[ip] {
				validIPs = append(validIPs, ip)
			}
		}
	}

	if len(validIPs) == 0 {
		return nil, context.DeadlineExceeded // 模拟一个未找到或超时的错误，供外层 continue
	}

	return validIPs, nil
}

// DetectWildcard 检测目标域名是否存在泛解析
// 返回：(是否开启泛解析, 泛解析指向的黑名单IP集合)
func DetectWildcard(domain string) (bool, map[string]bool) {
    wildcardIPs := make(map[string]bool)
    isWildcard := false
    
    // 连续用 3 个不同的随机子域名去测
    for i := 0; i < 3; i++ {
        randStr := generateRandomString(12)
        testDomain := randStr + "." + domain

        ips, err := ResolveDNS(testDomain)
        if err == nil && len(ips) > 0 {
            isWildcard = true // 只要有任意一个随机域名通了，说明就有泛解析
            for _, ip := range ips {
                wildcardIPs[ip] = true
            }
        }
    }

    return isWildcard, wildcardIPs
}

// 辅助函数：生成随机字符串
func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		ret[i] = letters[num.Int64()]
	}
	return string(ret)
}