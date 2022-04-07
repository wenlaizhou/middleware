package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"
)

// RequestDefaultTimeout http请求默认超时时间30s, 可在自己代码中进行全局修改
var RequestDefaultTimeout = 30

// PostWithTimeout post 带超时时间
// return : statusCode, header, body, error
func PostWithTimeout(timeoutSecond int, url string, contentType string, data []byte) (int, map[string][]string, []byte, error) {
	return PostFull(timeoutSecond, url, nil, contentType, data)
}

// PostJsonWithTimeout post json 带超时时间
// return : statusCode, header, body, error
func PostJsonWithTimeout(timeoutSecond int, url string, data interface{}) (int, map[string][]string, []byte, error) {
	return PostJsonFull(timeoutSecond, url, nil, data)
}

// GetWithTimeout get 带超时时间
// return : statusCode, header, body, error
func GetWithTimeout(timeoutSecond int, url string) (int, map[string][]string, []byte, error) {
	return GetFull(timeoutSecond, url, nil)
}

// Post
// return : statusCode, header, body, error
func Post(url string, contentType string, data []byte) (int, map[string][]string, []byte, error) {
	return PostFull(RequestDefaultTimeout, url, nil, contentType, data)
}

// PostJson : post json, data will parse to json string
// return : statusCode, header, body, error
func PostJson(url string, data interface{}) (int, map[string][]string, []byte, error) {
	return PostJsonFull(RequestDefaultTimeout, url, nil, data)
}

// Get : get请求
// return : statusCode, header, body, error
func Get(url string) (int, map[string][]string, []byte, error) {
	return GetFull(RequestDefaultTimeout, url, nil)
}

// PostFull : post data to url
//
// return : statusCode, header, body, error
func PostFull(timeoutSecond int, url string, headers map[string]string, contentType string, data []byte) (int, map[string][]string, []byte, error) {
	return DoRequest(timeoutSecond, POST, url, headers, contentType, data)
}

// PostJsonFull : post json data to url, contentType设置为: application/json utf8
//
// data : interface, 自动将其转成json格式
//
// return : statusCode, header, body, error
func PostJsonFull(timeoutSecond int, url string, headers map[string]string, data interface{}) (int, map[string][]string, []byte, error) {
	if data != nil {
		if reflect.TypeOf(data).Name() == reflect.TypeOf("").Name() {
			return DoRequest(timeoutSecond, POST, url, headers, ApplicationJson, []byte(fmt.Sprintf("%s", data)))
		}
		dataJson, _ := json.Marshal(data)
		return DoRequest(timeoutSecond, POST, url, headers, ApplicationJson, dataJson)
	}
	return DoRequest(timeoutSecond, POST, url, headers, ApplicationJson, nil)
}

// GetFull : get data from url
//
// return : statusCode, header, body, error
func GetFull(timeoutSecond int, url string, headers map[string]string) (int, map[string][]string, []byte, error) {
	return DoRequest(timeoutSecond, GET, url, headers, "", nil)
}

// DoRequest : post data to url
//
// return : statusCode, header, body, error
func DoRequest(timeoutSecond int, method string, url string,
	headers map[string]string, contentType string,
	body []byte) (int, map[string][]string, []byte, error) {

	bodyReader := bytes.NewReader(body)

	req, err := http.NewRequest(method, url, bodyReader)
	if ProcessError(err) {
		return -1, nil, nil, err
	}

	client := &http.Client{}
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}
	if len(contentType) > 0 {
		req.Header.Set(ContentType, contentType)
	}

	if timeoutSecond > 0 { // 设置超时时间
		client.Timeout = time.Second * time.Duration(timeoutSecond)
	}

	resp, err := client.Do(req)
	if ProcessError(err) {
		return -1, nil, nil, err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	return resp.StatusCode, resp.Header, respBody, nil
}

// DoRequestWithBaseAuth : post data to url
//
// return : statusCode, header, body, error
func DoRequestWithBaseAuth(timeoutSecond int, method string, url string,
	headers map[string]string, body []byte, username string, password string) (int, map[string][]string, []byte, error) {

	bodyReader := bytes.NewReader(body)

	req, err := http.NewRequest(method, url, bodyReader)
	if ProcessError(err) {
		return -1, nil, nil, err
	}

	req.SetBasicAuth(username, password)

	client := &http.Client{}
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	if timeoutSecond > 0 { // 设置超时时间
		client.Timeout = time.Second * time.Duration(timeoutSecond)
	}

	resp, err := client.Do(req)
	if ProcessError(err) {
		return -1, nil, nil, err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	return resp.StatusCode, resp.Header, respBody, nil
}
