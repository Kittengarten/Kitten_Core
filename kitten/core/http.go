package core

import (
	"io"
	"net/http"
	"strconv"
	"time"
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

// GET 获取 HTTP GET 响应体
func GET(url string) (io.ReadCloser, error) {
	res, err := doRequest(http.MethodGet, url, ``, nil)
	if nil != err {
		return nil, err
	}
	return res.Body, nil
}

// GETData 获取 HTTP GET 数据
func GETData(url string) ([]byte, error) {
	res, err := doRequest(http.MethodGet, url, ``, nil)
	if nil != err {
		return nil, err
	}
	return io.ReadAll(res.Body)
}

// POST 获取 HTTP POST 响应体
func POST(url, contentType string, body io.Reader) (io.ReadCloser, error) {
	res, err := doRequest(http.MethodPost, url, contentType, body)
	if nil != err {
		return nil, err
	}
	return res.Body, nil
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
func doRequest(method, url, contentType string, body io.Reader) (res *http.Response, err error) {
	req, err := http.NewRequest(method, url, body)
	if nil != err {
		return
	}
	req.Header.Set(UA, UserAgent)
	req.Header.Set(CT, contentType)
	client := &http.Client{Timeout: TimeOutSeconds * time.Second}
	res, err = client.Do(req)
	if nil != err {
		return
	}
	return res, checkError(res, url)
}

// 判断 HTTP 错误
func checkError(res *http.Response, url string) error {
	if http.StatusOK <= res.StatusCode && http.StatusMultipleChoices > res.StatusCode {
		return nil
	}
	defer res.Body.Close()
	return &HTTPErr{
		URL:        url,
		Method:     res.Request.Method,
		StatusCode: res.StatusCode,
	}
}
