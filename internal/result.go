package internal

// Result 存储单个子域名的爆破结果
type Result struct {
	Subdomain string   // 完整的子域名，例如 api.baidu.com
	IPs       []string // 查询到的 IP 地址列表
}
