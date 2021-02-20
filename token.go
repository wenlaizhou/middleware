package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
)

// 获取权限云token, 可缓存
func GetToken(service string, appId string, appSecret string, cache *Cache) (string, error) {
	if cache != nil {
		res := cache.GetData(fmt.Sprintf("%v%v", appId, appSecret))
		if res != nil {
			return fmt.Sprintf("%v", res), nil
		}

		token, _ := RequestToken(service, appId, appSecret)
		if IsEmpty(token) {
			return "", errors.New("token获取错误")
		}
		cache.InsertData(fmt.Sprintf("%v%v", appId, appSecret), token)
		return token, nil
	}
	return RequestToken(service, appId, appSecret)
}

// 获取权限云token, 直接发起请求
func RequestToken(service string, appId string, appSecret string) (string, error) {
	status, _, body, err := PostJsonFull(3, service, nil, map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     appId,
		"client_secret": appSecret,
	})

	if err != nil {
		mLogger.Error(err.Error())
		return "", err
	}
	if status != 200 {
		mLogger.ErrorF("请求token错误: %v %v %v 返回状态:%v", service, appId, appSecret, status)
	}
	res := make(map[string]interface{})
	json.Unmarshal(body, &res)
	val, hasVal := res["access_token"]
	if hasVal {
		return fmt.Sprintf("%v", val), nil
	}
	return "", errors.New(fmt.Sprintf("token请求返回错误: %v", string(body)))
}
