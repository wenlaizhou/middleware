/**
 *                                         ,s555SB@@&amp;
 *                                      :9H####@@@@@Xi
 *                                     1@@@@@@@@@@@@@@8
 *                                   ,8@@@@@@@@@B@@@@@@8
 *                                  :B@@@@X3hi8Bs;B@@@@@Ah,
 *             ,8i                  r@@@B:     1S ,M@@@@@@#8;
 *            1AB35.i:               X@@8 .   SGhr ,A@@@@@@@@S
 *            1@h31MX8                18Hhh3i .i3r ,A@@@@@@@@@5
 *            ;@&amp;i,58r5                 rGSS:     :B@@@@@@@@@@A
 *             1#i  . 9i                 hX.  .: .5@@@@@@@@@@@1
 *              sG1,  ,G53s.              9#Xi;hS5 3B@@@@@@@B1
 *               .h8h.,A@@@MXSs,           #@H1:    3ssSSX@1
 *               s ,@@@@@@@@@@@@Xhi,       r#@@X1s9M8    .GA981
 *               ,. rS8H#@@@@@@@@@@#HG51;.  .h31i;9@r    .8@@@@BS;i;
 *                .19AXXXAB@@@@@@@@@@@@@@#MHXG893hrX#XGGXM@@@@@@@@@@MS
 *                s@@MM@@@hsX#@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@&amp;,
 *              :GB@#3G@@Brs ,1GM@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@B,
 *            .hM@@@#@@#MX 51  r;iSGAM@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@8
 *          :3B@@@@@@@@@@@&amp;9@h :Gs   .;sSXH@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@:
 *      s&amp;HA#@@@@@@@@@@@@@@M89A;.8S.       ,r3@@@@@@@@@@@@@@@@@@@@@@@@@@@r
 *   ,13B@@@@@@@@@@@@@@@@@@@5 5B3 ;.         ;@@@@@@@@@@@@@@@@@@@@@@@@@@@i
 *  5#@@#&amp;@@@@@@@@@@@@@@@@@@9  .39:          ;@@@@@@@@@@@@@@@@@@@@@@@@@@@;
 *  9@@@X:MM@@@@@@@@@@@@@@@#;    ;31.         H@@@@@@@@@@@@@@@@@@@@@@@@@@:
 *   SH#@B9.rM@@@@@@@@@@@@@B       :.         3@@@@@@@@@@@@@@@@@@@@@@@@@@5
 *     ,:.   9@@@@@@@@@@@#HB5                 .M@@@@@@@@@@@@@@@@@@@@@@@@@B
 *           ,ssirhSM@&amp;1;i19911i,.             s@@@@@@@@@@@@@@@@@@@@@@@@@@S
 *              ,,,rHAri1h1rh&amp;@#353Sh:          8@@@@@@@@@@@@@@@@@@@@@@@@@#:
 *            .A3hH@#5S553&amp;@@#h   i:i9S          #@@@@@@@@@@@@@@@@@@@@@@@@@A.
 *
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
			value = strings.TrimSpace(value)
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
						res[subKey] = res[subValue]
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
