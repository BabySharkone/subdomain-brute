# subdomain-brute

一个简单的并发子域名爆破工具，用 Go 实现

## 特性
- Worker Pool 并发架构
- 自动过滤 DNS 污染产生的假阳性结果
- 支持自定义字典和并发数

## 项目结构
.
├── main.go                 # 入口程序
├── internal/
│   ├── bruteforcer.go      # 并发爆破子域名
│   ├── resolver.go         # DNS 解析模块
│   └── result.go           # 存储单个子域名的爆破结果
├── wordlists/              # 字典目录
│   └── subdomains.txt
└── go.mod

## 安装
go install github.com/yourusername/subdomain-brute@latest

## 使用
go run main.go -d example.com -w wordlist.txt -t 20 -o result.txt
or
go build -o subdomain-brute main.go
./subdomain-brute -d example.com -w wordlist.txt -t 20 -o result.txt