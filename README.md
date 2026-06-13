# subdomain-brute

一个高性能的并发子域名爆破工具，用 Go 实现

## 特性

- **高并发架构**: Worker Pool 模式，默认 50 个并发（可配置）
- **DNS 解析**: 使用 `net.LookupHost()` 进行 DNS 查询
- **假阳性过滤**: 自动过滤 DNS 污染产生的垃圾 IP（如 0.0.0.1, 127.0.0.1 等）
- **500+ 字典词汇**: 内置常见子域名词表，开箱即用
- **多格式导出**: 支持 TXT、JSON、CSV 三种输出格式
- **灵活配置**: 支持自定义字典、并发数、输出格式

## 项目结构

```
subdomain-brute/
├── main.go                 # CLI 入口程序
├── internal/
│   ├── bruteforcer.go      # 核心爆破引擎（goroutines + channels）
│   ├── resolver.go         # DNS 解析模块
│   └── result.go           # 结果导出功能（JSON/CSV/TXT）
├── wordlists/
│   └── subdomains.txt      # 500+ 常用子域名词表
├── go.mod                  # 模块定义
└── README.md               # 本文件
```

## 快速开始

### 编译

```bash
go build -o subdomain-brute main.go
```

### 基础用法

```bash
# 最简单的用法（使用默认字典和 50 个并发）
./subdomain-brute -d example.com

# 自定义并发数
./subdomain-brute -d example.com -t 100

# 使用自定义字典
./subdomain-brute -d example.com -w custom_wordlist.txt

# 导出为 JSON 格式
./subdomain-brute -d example.com -o result.json -f json

# 导出为 CSV 格式
./subdomain-brute -d example.com -o result.csv -f csv
```

## 命令行参数

| 参数 | 说明 | 默认值 | 示例 |
|------|------|--------|------|
| `-d` | 目标域名（**必填**） | - | `example.com` |
| `-w` | 字典文件路径 | `wordlists/subdomains.txt` | `custom.txt` |
| `-t` | 并发 goroutine 数 | `50` | `100` |
| `-o` | 输出文件路径（可选） | - | `result.txt` |
| `-f` | 输出格式（txt/json/csv） | `txt` | `json` |

## 使用示例

### 示例 1：快速扫描

```bash
./subdomain-brute -d github.com
```

输出：
```
============================================================
子域名爆破工具
============================================================
 目标域名: github.com
字典数量: 500
⚡ 并发数: 50
📝 输出格式: txt

进度: 500/500
============================================================
扫描完成！
============================================================
发现 6 个子域名：

[+] api.github.com          -> 140.82.112.6
[+] www.github.com          -> 140.82.114.3
[+] gist.github.com         -> 140.82.113.3
[+] pages.github.com        -> 185.199.109.153
[+] status.github.com       -> 54.174.145.174
[+] docs.github.com         -> 140.82.112.3

⏱扫描耗时: 2.345s
平均耗时: 4.69ms per subdomain
```

### 示例 2：大规模扫描 + JSON 导出

```bash
./subdomain-brute -d example.com -t 200 -o result.json -f json
```

输出的 `result.json`:
```json
[
  {
    "subdomain": "api.example.com",
    "ips": ["192.0.2.1"]
  },
  {
    "subdomain": "www.example.com",
    "ips": ["192.0.2.2", "192.0.2.3"]
  }
]
```

### 示例 3：CSV 格式导出

```bash
./subdomain-brute -d example.com -o result.csv -f csv
```

输出的 `result.csv`:
```
Subdomain,IPs
api.example.com,192.0.2.1
www.example.com,192.0.2.2,192.0.2.3
```

## 安装到系统

```bash
# 构建可执行文件
go build -o subdomain-brute main.go

# 将其移到系统路径
sudo mv subdomain-brute /usr/local/bin/

# 现在可以在任何地方使用
subdomain-brute -d example.com
```

## GitHub 安装

```bash
go install github.com/BabySharkone/subdomain-brute@latest
```

## 学习价值

这个项目是学习 Go 并发编程的绝佳案例：

- **goroutines**: 如何使用 `go` 关键字启动轻量级线程
- **channels**: 任务分发和结果收集的通道通信
- **sync.WaitGroup**: 等待所有 goroutine 完成
- **原子操作**: `atomic.AddInt32()` 用于线程安全的计数
- **接口设计**: 导出功能的策略模式实现
- **错误处理**: Go 的 error 处理模式

## 工作原理

```
1. 读取字典文件
   ↓
2. 启动 N 个 worker goroutine（默认 50）
   ↓
3. 将子域名前缀分发到任务队列
   ↓
4. 每个 worker 执行 DNS 查询
   ↓
5. 过滤假阳性结果
   ↓
6. 收集并导出结果
```

## 已知限制

- 只支持 DNS 解析检测（HTTP 检测在 Phase 2 计划中）
- 不支持代理（在 Phase 2 中计划添加）
- 无法跳过某些子域名前缀的扫描
