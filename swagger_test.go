package middleware

import "testing"

func TestSwaggerGenerate(t *testing.T) {

	swaggerData := SwaggerBuildModel("Middleware",
		"1.0.0",
		"First Api",
	)
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
	swaggerData.AddPath(SwaggerBuildPath("a", "a", "get", "aaa").AddParameter(
		SwaggerParameter{
			Name:        "aaa",
			Description: "bbb",
			Example: `{
  "hello" : "world"
}`,
			In:       "body",
			Required: true,
		}))
	println(GenerateSwagger(swaggerData))
}
