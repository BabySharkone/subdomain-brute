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
		// 一旦遇到重定向，直接返回错误和当前的 resp，不再往下跟
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	return &HTTPDetector{Client:client}
}

// CheckHTTP 检查子域名的 HTTP 状态码
// 返回：(状态码, 错误)
func (h *HTTPDetector) CheckHTTP(domain string) (int,error) {
	urls := []string{
		"https://" + domain,
		"http://" + domain,
	}

	var lastErr error
	for _,url := range urls{
		// 发送 HEAD 请求
		resp,err := h.Client.Head(url)
		if resp != nil {
			resp.Body.Close()
		}
		if err != nil{
			lastErr = err	// 记录错误，不 return，继续下一轮循环测另一个协议
			continue	
		}
		// 走到这里说明 err == nil，成功拿到了状态码，直接返回
		return resp.StatusCode,nil
	}
	// 如果循环走完了都没有 return，说明两个协议都失败了，返回最后的错误
	return 0, lastErr
}