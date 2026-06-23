package internal

import (
	"net"
	"crypto/rand"
	"fmt"
	"math/big"
)

var junkIPs = map[string]bool{
	"0.0.0.1":   true,
	"127.0.0.1": true,
	"0.0.0.0":   true,
}

// ResolveDNS 负责查询一个域名的 IP
// 如果查询成功，返回 IP 列表；如果域名不存在，返回错误
func ResolveDNS(domain string) ([]string, error) {
	ips, err := net.LookupHost(domain)
	if err != nil {
		return nil, err
	}
	var validIPs []string
	for _, ip := range ips {
		if !junkIPs[ip] {
			validIPs = append(validIPs, ip)
		}
	}
	if len(validIPs) == 0 {
		return nil, &net.DNSError{Err: "no valid IP", Name: domain, IsNotFound: true}
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

        ips, err := net.LookupHost(testDomain)
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