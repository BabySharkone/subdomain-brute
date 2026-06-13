package internal

import (
	"net"
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
	var vakidIPS []string
	for _, ip := range ips {
		if !junkIPs[ip] {
			vakidIPS = append(vakidIPS, ip)
		}
	}
	if len(vakidIPS) == 0 {
		return nil, &net.DNSError{Err: "no valid IP", Name: domain, IsNotFound: true}
	}
	return vakidIPS, nil
}
