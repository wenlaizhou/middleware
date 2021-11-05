package middleware

import "encoding/json"

// 返回 json数据
// { code 格式: 0, message : "", data : {}}
func (this *Context) ApiResponse(code int, message string, data interface{}) error {
	model := map[string]interface{}{
		"code":    code,
		"message": message,
		"data":    data,
	}
	res, err := json.Marshal(model)
	if ProcessError(err) {
		return err
	}
	err = this.OK(ApplicationJson, res)
	return err
}
