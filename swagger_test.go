package middleware

import "testing"

func TestSwaggerGenerate(t *testing.T) {
	println(SwaggerGenerate(Swagger{
		Swagger: "2.0",
		Info: struct {
			Title       string `json:"title"`
			Version     string `json:"version"`
			Description string `json:"description"`
		}{
			Title:       "Middleware",
			Version:     "2.0.0",
			Description: "hello world",
		},
		Host: "http://localhost",
	}))
}
