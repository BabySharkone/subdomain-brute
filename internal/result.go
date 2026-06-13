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
	Subdomain string   `json:"subdomain"`
	IPs       []string `json:"ips"`
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
	writer.Write([]string{"Subdomain", "IPs"})

	// 写入数据
	for _, r := range results {
		writer.Write([]string{r.Subdomain, strings.Join(r.IPs, ",")})
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
		fmt.Fprintf(file, "%s -> %s\n", r.Subdomain, strings.Join(r.IPs, ", "))
	}

	return nil
}
