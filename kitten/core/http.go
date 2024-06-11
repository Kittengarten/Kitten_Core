package core

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

// HTTPErr 是一个表示 HTTP 错误的结构体
type HTTPErr struct {
	URL        string // 请求的 URL
	Method     string // 请求的方法
	StatusCode int    // HTTP 状态码
}

const (
	CT        = `Content-Type`
	UA        = `User-Agent`
	UserAgent = `Mozilla/5.0 (Windows NT 10.0; Win64; x64)` +
		` AppleWebKit/537.36 (KHTML, like Gecko)` +
		` Chrome/124.0.0.0 Safari/537.36 Edg/124.0.0.0`
	TimeOutSeconds = 5
)

// Error 实现 error
func (e *HTTPErr) Error() string {
	return `链接：` + e.URL + `
方法：` + e.Method + `
HTTP 错误：` + strconv.Itoa(e.StatusCode)
}

// GET 获取 HTTP GET 响应体（无需关闭）
func GET(url string) (io.Reader, error) {
	res, err := doRequest(http.MethodGet, url, ``, nil)
	if nil != err {
		return nil, err
	}
	return charset.NewReader(res.Body, res.Header.Get(CT))
}

// GETData 获取 HTTP GET 数据
func GETData(url string) ([]byte, error) {
	res, err := doRequest(http.MethodGet, url, ``, nil)
	if nil != err {
		return nil, err
	}
	return io.ReadAll(res.Body)
}

// POST 获取 HTTP POST 响应体（无需关闭）
func POST(url, contentType string, body io.Reader) (io.Reader, error) {
	res, err := doRequest(http.MethodPost, url, contentType, body)
	if nil != err {
		return nil, err
	}
	return charset.NewReader(res.Body, res.Header.Get(CT))
}

// POSTData 获取 HTTP POST 数据
func POSTData(url, contentType string, body io.Reader) ([]byte, error) {
	res, err := doRequest(http.MethodPost, url, contentType, body)
	if nil != err {
		return nil, err
	}
	return io.ReadAll(res.Body)
}

// 执行 HTTP 请求
func doRequest(method, url, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if nil != err {
		return nil, err
	}
	req.Header.Set(UA, UserAgent)
	req.Header.Set(CT, contentType)
	res, err := (&http.Client{Timeout: TimeOutSeconds * time.Second}).Do(req)
	if nil != err {
		return nil, err
	}
	return res, checkError(res, url)
}

// 判断 HTTP 错误
func checkError(res *http.Response, url string) error {
	if http.StatusOK <= res.StatusCode && http.StatusMultipleChoices > res.StatusCode {
		// 不能处理 3xx 重定向状态码
		return nil
	}
	defer res.Body.Close()
	return &HTTPErr{
		URL:        url,
		Method:     res.Request.Method,
		StatusCode: res.StatusCode,
	}
}

// InnerText 在 *html.Node 中使用 XPath 获取文本
func InnerText(top *html.Node, expr string) string {
	node := htmlquery.FindOne(top, expr)
	if nil == node {
		return ``
	}
	return htmlquery.InnerText(node)
}

// 获取网页 Node
func FetchNode(url string) (*html.Node, error) {
	// 获取响应体
	body, err := GET(url)
	if nil != err {
		return nil, err
	}
	// 解析网页
	return html.Parse(body)
}
