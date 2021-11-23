package middleware

import "encoding/json"

// 返回 json数据
// { code 格式: 0, message : "", data : {}}
func (c *Context) ApiResponse(code int, message string, data interface{}) {
	model := map[string]interface{}{
		"code":    code,
		"message": message,
		"data":    data,
	}
	res, err := json.Marshal(model)
	if ProcessError(err) {
		return
	}
	c.OK(ApplicationJson, res)
	return
}
