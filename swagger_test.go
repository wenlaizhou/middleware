package middleware

import "testing"

func TestSwaggerGenerate(t *testing.T) {

	swaggerData := SwaggerData{
		Title:       "Middleware",
		Version:     "",
		Description: "",
		Host:        "",
	}
	path := SwaggerPath{
		Path:        "",
		Method:      "",
		Description: "",
		Parameters:  nil,
	}
	path.AddParameter(SwaggerParameter{
		Name:        "",
		Description: "",
		Example:     "",
		Default:     "",
		Type:        "",
		Required:    false,
	})
	path.AddParameter(SwaggerParameter{
		Name:        "",
		Description: "",
		Example:     "",
		Default:     "",
		Type:        "",
		Required:    false,
	})
	swaggerData.AddPath(path)
	println(GenerateSwagger(swaggerData))
}
