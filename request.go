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

var defaultTimeout = 30

// return : statusCode, header, body, error
func PostWithTimeout(timeoutSecond int, url string, contentType string, data []byte) (int, map[string][]string, []byte, error) {
	return PostFull(timeoutSecond, url, nil, contentType, data)
}

// return : statusCode, header, body, error
func PostJsonWithTimeout(timeoutSecond int, url string, data interface{}) (int, map[string][]string, []byte, error) {
	return PostJsonFull(timeoutSecond, url, nil, data)
}

// return : statusCode, header, body, error
func GetWithTimeout(timeoutSecond int, url string) (int, map[string][]string, []byte, error) {
	return GetFull(timeoutSecond, url, nil)
}

// return : statusCode, header, body, error
func Post(url string, contentType string, data []byte) (int, map[string][]string, []byte, error) {
	return PostFull(defaultTimeout, url, nil, contentType, data)
}

// return : statusCode, header, body, error
func PostJson(url string, data interface{}) (int, map[string][]string, []byte, error) {
	return PostJsonFull(defaultTimeout, url, nil, data)
}

// return : statusCode, header, body, error
func Get(url string) (int, map[string][]string, []byte, error) {
	return GetFull(defaultTimeout, url, nil)
}

// Post: post data to url
//
// return : statusCode, header, body, error
func PostFull(timeoutSecond int, url string, headers map[string]string, contentType string, data []byte) (int, map[string][]string, []byte, error) {
	return DoRequest(timeoutSecond, POST, url, headers, contentType, data)
}

// PostJson: post json data to url, contentType设置为: application/json utf8
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

// Get: get data from url
//
// return : statusCode, header, body, error
func GetFull(timeoutSecond int, url string, headers map[string]string) (int, map[string][]string, []byte, error) {
	return DoRequest(timeoutSecond, GET, url, headers, "", nil)
}

// DoRequest: post data to url
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
