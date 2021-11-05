package middleware

import "encoding/json"

func GenerateSwaggerJson() {

}

// 返回 json数据
// { code 格式: 0, message : "", data : {}}
func (this *Context) ApiResponse(code int, message string, data interface{}) error {
	model := make(map[string]interface{})
	model["code"] = code
	model["message"] = message
	if len(this.restProcessors) > 0 {
		for _, p := range this.restProcessors {
			data = p(data)
		}
	}
	model["data"] = data
	res, err := json.Marshal(model)
	if ProcessError(err) {
		return err
	}
	err = this.OK(ApplicationJson, res)
	return err
}
