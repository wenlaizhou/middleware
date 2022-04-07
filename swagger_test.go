package middleware

import "testing"

func TestSwaggerGenerate(t *testing.T) {

	swaggerData := SwaggerData{
		Title:       "Middleware",
		Version:     "1.0.0",
		Description: "First Api",
		Host:        "ai.mycyclone.com",
	}
	path := SwaggerPath{
		Path:        "/hello",
		Method:      "get",
		Description: "index",
	}
	path.AddParameter(SwaggerParameter{
		Name:        "id",
		Description: "id for hello",
		Example:     "1",
		Default:     "0",
		Required:    false,
	})
	swaggerData.AddPath(path)
	println(GenerateSwagger(swaggerData))
}
