package internal

import(
	"net"
	"net/http"
	"time"
)

type HTTPDetector struct {
	Client *http.Client
}

// NewHTTPDetector 创建 HTTP 检测器
func NewHTTPDetector(timeout time.Duration) *HTTPDetector {
	client := &http.Client {
		Timeout:timeout,
		Transport: &http.Transport{
			Dial:(&net.Dialer{
				Timeout:timeout,
			}).Dial,
			ResponseHeaderTimeout:timeout,
		},
	}
	return &HTTPDetector{Client:client}
}

// CheckHTTP 检查子域名的 HTTP 状态码
// 返回：(状态码, 错误)
func (h *HTTPDetector) CheckHTTP(domain string) (int,error) {
	urls := []string{
		"http://" + domain,
		"https://" + domain,
	}
	for _,url := range urls{
		// 发送 HEAD 请求
		resp,err := h.Client.Head(url)
		if err == nil{
			resp.Body.Close()
			// 返回状态码
			return resp.StatusCode,nil
		}
	}
	return 0,&net.DNSError{Err:"no valid HTTP response"}
}