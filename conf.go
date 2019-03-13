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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
)

const ConfDir = "CONF_DIR"

type Config map[string]interface{}

// 读取json类型配置文件
func LoadConfig(confPath string) Config {
	res := make(Config)
	if !Exists(confPath) {
		return nil
	}
	data, err := ioutil.ReadFile(confPath)
	if ProcessError(err) {
		return nil
	}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil
	}
	res[ConfDir] = filepath.Dir(confPath)
	return res
}

// 获取配置中的value值
func ConfValue(conf Config, key string) (interface{}, error) {
	if len(key) <= 0 {
		return -1, errors.New("key 不为空")
	}
	v, hasData := conf[key]
	if !hasData {
		return nil, errors.New("没有匹配的value")
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

// 获取配置中的字符串类型值
func ConfString(conf Config, key string) (string, error) {
	v, err := ConfValue(conf, key)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", v), nil
}

// 获取配置中的value值
//
// 没有该值, 返回nil
func ConfValueUnsafe(conf Config, key string) interface{} {
	v, _ := ConfValue(conf, key)
	return v
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

// 获取配置中的字符串类型值
//
// 错误返回""
func ConfStringUnsafe(conf Config, key string) string {
	v, err := ConfValue(conf, key)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}
