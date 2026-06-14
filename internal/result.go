package internal

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Result 存储单个子域名的爆破结果
type Result struct {
	Subdomain 	string   	`json:"subdomain"`
	IPs       	[]string 	`json:"ips"`
	HTTPStatus 	int			`json:"http_status,omitempty"`
	IsValid		bool		`json:"is_valid"`
}

// ExportJSON 导出为 JSON 格式
func ExportJSON(results []Result, filepath string) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath, data, 0644)
}

// ExportCSV 导出为 CSV 格式
func ExportCSV(results []Result, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入 header
	writer.Write([]string{"Subdomain", "IPs", "HTTP Status"})

	// 写入数据
	for _, r := range results {
		status := ""
		if r.HTTPStatus > 0 {
			status = fmt.Sprintf("%d", r.HTTPStatus)
		}
		writer.Write([]string{r.Subdomain, strings.Join(r.IPs, ","), status})
	}

	return nil
}

// ExportTXT 导出为纯文本格式
func ExportTXT(results []Result, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, r := range results {
		if r.HTTPStatus > 0 {
			fmt.Fprintf(file, "%s -> %s (HTTP: %d)\n", r.Subdomain, strings.Join(r.IPs, ", "), r.HTTPStatus)
		}else{
			fmt.Fprintf(file, "%s -> %s\n", r.Subdomain, strings.Join(r.IPs, ", "))
		}
	}

	return nil
}

// IsValidHTTPStatus 判断状态码是否表示有效的服务
func IsValidHTTPStatus(status int) bool{
	validCodes := map[int]bool{
		200: true,	// OK
		201: true,	// Created
		302: true,	// 重定向
		301: true,	// 重定向
		401: true,	// 需要认证（但服务存在）
		403: true,	// 禁止访问（但服务存在）
	}
	return validCodes[status]
}