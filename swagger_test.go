package middleware

import "testing"

func TestSwaggerGenerate(t *testing.T) {
	println(SwaggerGenerate(Swagger{
		Title:       "",
		Host:        "",
		Version:     "",
		Description: "",
	}))
}
