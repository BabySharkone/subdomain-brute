# subdomain-brute

一个高性能的并发子域名爆破工具，用 Go 实现。支持 DNS 解析、HTTP/HTTPS 检测、泛解析识别，快速准确地发现目标域名的所有有效子域名。

## 特性

- **高并发架构**: Worker Pool 模式，默认 50 个并发（可配置）
- **智能 DNS 解析**: 集成 `miekg/dns` 库，支持多 DNS 服务器查询和故障转移
- **泛解析检测**: 自动识别和过滤泛解析域名产生的伪子域名
- **HTTP/HTTPS 检测**: 可选的 HTTP 状态码检测，提高准确性
- **假阳性过滤**: 自动过滤 DNS 污染产生的垃圾 IP（如 0.0.0.1, 127.0.0.1 等）
- **内置词表**: 500+ 常见子域名词汇，开箱即用，支持自定义字典
- **多格式导出**: 支持 TXT、JSON、CSV 三种输出格式，便于数据分析
- **进度展示**: 实时显示扫描进度和发现的子域名数量
- **灵活配置**: 支持自定义并发数、超时、字典等参数

## 项目结构

```
subdomain-brute/
├── main.go                      # CLI 入口程序，参数解析和程序流程控制
├── internal/
│   ├── bruteforcer.go           # 核心爆破引擎
│   ├── resolver.go              # DNS 解析模块
│   ├── http_detector.go         # HTTP/HTTPS 状态码检测模块
│   └── result.go                # 结果导出功能（JSON/CSV/TXT 格式）
├── wordlists/
│   └── subdomains.txt           # 500+ 常用子域名词表（可自定义）
├── go.mod                       # Go 模块定义
├── go.sum                       # 依赖校验文件
└── README.md                    # 项目文档
```

## 快速开始

### 安装

```bash
# 方式一：使用 go install（推荐）
go install github.com/BabySharkone/subdomain-brute@latest

# 方式二：从源码编译
git clone https://github.com/BabySharkone/subdomain-brute.git
cd subdomain-brute
go build -o subdomain-brute main.go
```

### 前置要求

- Go 1.18 或更高版本
- 网络连接（用于 DNS 查询和 HTTP 检测）

### 基础用法

```bash
# 最简单：DNS 查询（快速但可能有误报）
./subdomain-brute -d example.com

# 启用 HTTP 检测（更准确，推荐）
./subdomain-brute -d example.com -http

# 自定义并发数
./subdomain-brute -d example.com -t 100

# 使用自定义字典
./subdomain-brute -d example.com -w custom_wordlist.txt

# 导出为 JSON 格式（包含 HTTP 状态码）
./subdomain-brute -d example.com -http -o result.json -f json

# 导出为 CSV 格式
./subdomain-brute -d example.com -http -o result.csv -f csv

# 自定义 HTTP 超时（默认 5 秒）
./subdomain-brute -d example.com -http -http-timeout 10
```

## 命令行参数

| 参数 | 说明 | 默认值 | 示例 |
|------|------|--------|------|
| `-d` | 目标域名（**必填**） | - | `example.com` |
| `-w` | 字典文件路径 | `wordlists/subdomains.txt` | `custom.txt` |
| `-t` | 并发 goroutine 数 | `50` | `100` |
| `-o` | 输出文件路径（可选） | - | `result.txt` |
| `-f` | 输出格式（txt/json/csv） | `txt` | `json` |
| `-http` | 启用 HTTP 检测 | `false` | `-http` |
| `-http-timeout` | HTTP 请求超时（秒） | `5` | `-http-timeout 10` |

### 参数说明

- **-d 目标域名**：支持输入格式：`example.com`、`https://example.com`、`http://example.com:8080/path`，工具会自动规范化
- **-w 字典路径**：每行一个子域名前缀，例如 `api`、`www`、`admin` 等
- **-t 并发数**：增加并发数可提高扫描速度，但会占用更多系统资源（建议 50-200）
- **-http 启用 HTTP 检测**：会对每个发现的子域名进行 HTTP/HTTPS 请求，确保该子域名确实可达，准确率更高但速度会慢 3-5 倍
- **-http-timeout**：HTTP 请求超时时间，网络不稳定时建议增加

## 使用示例

### 场景 1：快速 DNS 扫描
```bash
./subdomain-brute -d github.com -t 50
```
优点：速度快（秒级）
注意：可能被 DNS 污染影响，会产生一些误报

### 场景 2：精准 HTTP 检测（推荐）
```bash
./subdomain-brute -d github.com -t 50 -http
```
优点：只返回真实存在的子域名，准确率接近 100%
耗时：根据并发数和子域名数量，通常 2-10 分钟

### 场景 3：导出为 JSON 格式
```bash
./subdomain-brute -d github.com -t 50 -http -o result.json -f json
```

**输出示例 result.json**：
```json
[
  {
    "subdomain": "api.github.com",
    "ips": ["140.82.112.6"],
    "http_status": 200,
    "is_valid": true
  },
  {
    "subdomain": "www.github.com",
    "ips": ["140.82.114.3"],
    "http_status": 200,
    "is_valid": true
  },
  {
    "subdomain": "gist.github.com",
    "ips": ["140.82.113.4"],
    "http_status": 200,
    "is_valid": true
  }
]
```

### 场景 4：导出为 CSV 格式（用于 Excel 分析）
```bash
./subdomain-brute -d github.com -t 100 -http -o result.csv -f csv
```

## 核心技术与学习价值

这个项目是学习 Go 并发编程和网络安全工具开发的绝佳案例：

### Go 并发编程
- **goroutines**: 轻量级线程的高效管理和启动
- **channels**: 任务分发和结果收集的通道通信模式
- **sync.WaitGroup**: 优雅地等待所有 goroutine 完成
- **原子操作**: `atomic.AddInt32()` 实现线程安全的计数器
- **并发控制**: Worker Pool 模式限制并发数，避免资源耗尽

### 网络编程
- **DNS 查询**: 使用 `miekg/dns` 库进行底层 DNS 操作
- **HTTP 客户端**: 实现带超时和重试的 HTTP 请求
- **网络故障处理**: 超时、连接失败等常见场景的处理

### 软件设计
- **接口设计**: 导出功能的策略模式实现（TXT/JSON/CSV）
- **错误处理**: Go 的 error 处理模式和最佳实践
- **代码组织**: 清晰的模块划分（DNS、HTTP、结果导出）

## 工作原理

### 整体流程
```
┌─ 读取字典文件 (500+ 子域名前缀)
│
├─ 初始化 N 个 worker goroutine (默认 50)
│
├─ 将子域名列表分发到任务队列 (tasks channel)
│
├─ 每个 worker 执行 DNS 查询
│  └─ 失败则重试或跳过
│
├─ 自动检测并过滤泛解析域名
│  └─ 识别通配符解析产生的伪子域名
│
├─ [可选] 对成功解析的子域名进行 HTTP/HTTPS 检测
│  └─ 验证子域名确实可达（HTTP 200 等响应）
│
├─ 去重处理（Map 数据结构）
│  └─ 相同的子域名和 IP 组合只保留一份
│
└─ 按格式导出结果 (TXT/JSON/CSV)
```

### DNS 泛解析检测机制
工具会在扫描开始时检测目标域名是否存在泛解析：
- 向一个明确不存在的前缀（如 `test-nonexistent-subdomain`）进行 DNS 查询
- 如果该查询返回 IP 地址，说明目标域名配置了通配符解析（`*.example.com`）
- 后续所有解析结果都会与泛解析 IP 进行比对，自动过滤伪子域名

##  已知限制与注意事项

### 当前限制
- **代理支持**：暂不支持 HTTP/SOCKS 代理（计划在后续版本支持）
- **选择性扫描**：暂无黑名单或白名单功能
- **DNS 服务器**：使用系统默认 DNS 服务器，暂无自定义 DNS 服务器选项

### 使用注意
- **速率限制**：某些 DNS 服务器可能对高频查询进行限制，建议在生产环境中适度调整并发数
- **IP 黑名单**：目前过滤的假阳性 IP 列表有限，可能在某些特殊网络环境下产生误报
- **HTTP 检测耗时**：启用 HTTP 检测会显著增加扫描耗时（通常 3-5 倍），但准确率提升明显
- **字典质量**：结果质量取决于字典词表的质量，建议使用高质量的、针对目标的字典
