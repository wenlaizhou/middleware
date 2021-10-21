package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

type tokenCache struct {
	Token   string `json:"token"`
	Expires int64  `json:"expires"`
}

// 获取权限云token, 可缓存
func GetToken(service string, appId string, appSecret string) (string, error) {
	mLogger.InfoLn(fmt.Sprintf("获取token请求: %v, %v, %v", service, appId, appSecret))
	cacheFilename := fmt.Sprintf("token_%v_%v", appId, appSecret)
	tokenFile, err := ioutil.ReadFile(cacheFilename)
	if err != nil {
		return cacheToken(service, appId, appSecret)
	}
	cacheData := tokenCache{}
	err = json.Unmarshal(tokenFile, &cacheData)
	if err != nil {
		return cacheToken(service, appId, appSecret)
	}
	if cacheData.Expires < time.Now().Unix() {
		// 	cache已过期
		return cacheToken(service, appId, appSecret)
	}
	return cacheData.Token, nil
	// return RequestToken(service, appId, appSecret)
}

func cacheToken(service string, appId string, appSecret string) (string, error) {
	cacheFilename := fmt.Sprintf("token_%v_%v", appId, appSecret)
	os.Remove(cacheFilename)
	token, err := RequestToken(service, appId, appSecret)
	if err != nil {
		return "", err
	}
	tokenData, _ := json.Marshal(
		tokenCache{
			Token:   token,
			Expires: time.Now().Unix() + 600,
		},
	)
	ioutil.WriteFile(cacheFilename, tokenData, os.ModePerm)
	return token, nil
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
