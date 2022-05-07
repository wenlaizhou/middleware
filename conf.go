/*
	config module
*/
package middleware

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
)

const ConfDir = "CONF_DIR"

type Config map[string]string

// 读取properties类型配置文件
func LoadConfig(confPath string) Config {
	res := make(Config)
	if !Exists(confPath) {
		return nil
	}
	data, err := ioutil.ReadFile(confPath)
	if ProcessError(err) {
		return nil
	}
	res[ConfDir] = filepath.Dir(confPath)
	confStr := string(data)
	lines := strings.Split(strings.TrimSpace(confStr), "\n")
	if len(lines) <= 0 {
		return res
	}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.ReplaceAll(line, "\\n", "\n")
		if len(line) <= 0 || strings.HasPrefix(line, "#") {
			continue
		}
		kvs := strings.Split(line, "=")
		if len(kvs) <= 1 {
			continue
		}
		key := strings.TrimSpace(kvs[0])
		value := strings.TrimSpace(kvs[1])
		if len(kvs) > 2 {
			value = strings.Join(kvs[1:], "=")
		}
		if len(key) <= 0 {
			continue
		}
		if key == "include" {
			if len(value) > 0 {
				if !strings.HasSuffix(value, ".properties") {
					value = fmt.Sprintf("%v.properties", value)
				}
				subConf := LoadConfig(value)
				if len(subConf) > 0 {
					for subKey, subValue := range subConf {
						res[subKey] = subValue
					}
				}
			}
			continue
		}
		res[key] = value
	}
	return res
}

// 获取配置值, 不存在该值, 则返回 ""
func (c Config) Unsafe(key string) string {
	v, err := c.Value(key)
	if err != nil {
		return ""
	}
	return v
}

// 获取配置中的value值
func (c Config) Value(key string) (string, error) {
	if len(key) <= 0 {
		return "", errors.New("key 不为空")
	}
	v, hasData := c[key]
	if !hasData {
		return "", errors.New("没有匹配的value")
	}
	return strings.TrimSpace(v), nil
}

// 获取配置中的数值类型值
func (c Config) Int(key string) (int, error) {
	v, err := c.Value(key)
	if err != nil {
		return -1, err
	}
	vStr := fmt.Sprintf("%v", v)
	return strconv.Atoi(vStr)
}

// 获取配置中的数值类型值
//
// 错误返回-1
func (c Config) IntUnsafe(key string) int {
	v, err := c.Value(key)
	if err != nil {
		return -1
	}
	res, _ := strconv.Atoi(fmt.Sprintf("%v", v))
	return res
}

// 获取配置中的bool
//
// 错误, 类型错误, 没有该值, 返回false
// "1", "t", "true" 定义为true
func (c Config) Bool(key string) bool {
	v, err := c.Value(key)
	if err != nil {
		return false
	}
	switch strings.ToLower(fmt.Sprintf("%v", v)) {
	case "1", "t", "true":
		return true
	}
	return false
}

// 获取配置文件内容并返回json
func (c Config) Print() string {
	if len(c) <= 0 {
		return ""
	}
	res := ""
	for k, v := range c {
		if len(res) > 0 {
			res = fmt.Sprintf("%s\n%s = %s", res, k, v)
		} else {
			res = fmt.Sprintf("%s = %s", k, v)
		}
	}
	return res
}

func (c Config) IntUnsafeDefault(key string, defaultVal int) int {
	v, err := c.Value(key)
	if err != nil {
		return -1
	}
	res, err := strconv.Atoi(fmt.Sprintf("%v", v))
	if err != nil {
		return defaultVal
	}
	return res
}

// 获取配置值, 不存在该值, 则返回 ""
func (c Config) UnsafeDefault(key string, defaultVal string) string {
	v, err := c.Value(key)
	if err != nil {
		return defaultVal
	}
	return v
}

// 获取配置值, 不存在该值, 则返回 ""
func ConfUnsafe(conf Config, key string) string {
	v, err := ConfValue(conf, key)
	if err != nil {
		return ""
	}
	return v
}

// 获取配置中的value值
func ConfValue(conf Config, key string) (string, error) {
	if len(key) <= 0 {
		return "", errors.New("key 不为空")
	}
	v, hasData := conf[key]
	if !hasData {
		return "", errors.New("没有匹配的value")
	}
	return v, nil
}

// 获取配置中的数值类型值
func ConfInt(conf Config, key string) (int, error) {
	v, err := ConfValue(conf, key)
	if err != nil {
		return -1, err
	}
	vStr := fmt.Sprintf("%v", v)
	return strconv.Atoi(vStr)
}

// 获取配置中的数值类型值
//
// 错误返回-1
func ConfIntUnsafe(conf Config, key string) int {
	v, err := ConfValue(conf, key)
	if err != nil {
		return -1
	}
	res, _ := strconv.Atoi(fmt.Sprintf("%v", v))
	return res
}

// 获取配置中的bool
//
// 错误, 类型错误, 没有该值, 返回false
func ConfBool(conf Config, key string) bool {
	v, err := ConfValue(conf, key)
	if err != nil {
		return false
	}
	switch fmt.Sprintf("%v", v) {
	case "1", "t", "T", "true", "TRUE", "True":
		return true
	case "0", "f", "F", "false", "FALSE", "False":
		return false
	}
	return false
}

// 获取配置文件内容并返回json
func ConfPrint(conf Config) string {
	if len(conf) <= 0 {
		return ""
	}
	res := ""
	for k, v := range conf {
		if len(res) > 0 {
			res = fmt.Sprintf("%s\n%s = %s", res, k, v)
		} else {
			res = fmt.Sprintf("%s = %s", k, v)
		}
	}
	return res
}

func ConfIntUnsafeDefault(conf Config, key string, defaultVal int) int {
	v, err := ConfValue(conf, key)
	if err != nil {
		return -1
	}
	res, err := strconv.Atoi(fmt.Sprintf("%v", v))
	if err != nil {
		return defaultVal
	}
	return res
}

// 获取配置值, 不存在该值, 则返回 ""
func ConfUnsafeDefault(conf Config, key string, defaultVal string) string {
	v, err := ConfValue(conf, key)
	if err != nil {
		return defaultVal
	}
	return v
}
