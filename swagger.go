package middleware

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

type SwaggerPath struct {
	Path                     string
	Group                    string
	Method                   string
	Description              string
	Parameters               []SwaggerParameter
	ResponseObjectProperties []SwaggerResponseProperty
}

type SwaggerResponseProperty struct {
	Name string
	// string int ...
	Type        string
	Description string
}

type SwaggerParameter struct {
	Name        string
	Description string
	Example     interface{}
	Default     interface{}
	// formData, path, header, body, query
	In       string
	Required bool
}

type SwaggerData struct {
	Title       string
	Version     string
	Description string
	Host        string
	Apis        []SwaggerPath
}

func SwaggerBuildModel(title string, desc string, version string) SwaggerData {
	return SwaggerData{
		Title:       title,
		Version:     version,
		Description: desc,
		Host:        "",
	}
}

func (thisSelf *SwaggerData) AddPath(path SwaggerPath, params []SwaggerParameter, response []SwaggerResponseProperty) SwaggerData {
	pathData := path
	if params != nil && len(params) > 0 {
		for _, param := range params {
			pathData.AddParameter(param)
		}
	}
	if response != nil && len(response) > 0 {
		for _, resp := range response {
			pathData.AddResponseProperty(resp)
		}
	}
	thisSelf.Apis = append(thisSelf.Apis, pathData)
	return *thisSelf
}

func SwaggerBuildPath(path string, group string, method string, description string) SwaggerPath {
	return SwaggerPath{
		Path:        path,
		Group:       group,
		Method:      method,
		Description: description,
	}
}

func (thisSelf *SwaggerPath) AddParameter(param SwaggerParameter) *SwaggerPath {
	thisSelf.Parameters = append(thisSelf.Parameters, param)
	return thisSelf
}

func (thisSelf *SwaggerPath) AddResponseProperty(param SwaggerResponseProperty) *SwaggerPath {
	thisSelf.ResponseObjectProperties = append(thisSelf.ResponseObjectProperties, param)
	return thisSelf
}

/*
swagger: '2.0'
host: 'localhost'
info:
  title: Middleware
  version: 0.0.1
  description: "This is a sample server Petstore server.  You can find out more about     Swagger at [http://swagger.io](http://swagger.io) or on [irc.freenode.net, #swagger](http://swagger.io/irc/).      For this sample, you can use the api key `special-key` to test the authorization     filters."
tags:
  - name: h1
    description: hello h1
paths:
  '/hello':
    post:
      tags:
        - "h1"
      summary: Greet our subject with hello!
      parameters:
        - name: subject
          description: The subject to be greeted.
          required: false
          type: string
          in: body #formData, path, header, body, query
          example: {
              "hello" : "world"
            }
          default: 123
      responses:
        default:
          description: Some description
          schema:
            type: string
*/

func GenerateSwagger(model SwaggerData) string {
	swaggerJson := map[string]interface{}{}
	swaggerJson["swagger"] = "2.0"
	swaggerJson["host"] = model.Host
	swaggerJson["info"] = map[string]string{
		"title":       model.Title,
		"version":     model.Version,
		"description": model.Description,
	}
	paths := map[string]interface{}{}
	for _, api := range model.Apis {
		var parameters []map[string]interface{} = make([]map[string]interface{}, 0)
		if api.Parameters != nil && len(api.Parameters) > 0 {
			for _, p := range api.Parameters {
				parameters = append(parameters, map[string]interface{}{
					"name":        p.Name,
					"description": p.Description,
					"required":    p.Required,
					"default":     p.Default,
					"in":          strings.ToLower(p.In),
					"example":     p.Example,
				})
			}
		}
		apiResponse := map[string]interface{}{
			"type": "string",
		}
		if api.ResponseObjectProperties != nil && len(api.ResponseObjectProperties) > 0 {
			apiResponse["type"] = "object"
			properties := map[string]interface{}{}
			for _, resp := range api.ResponseObjectProperties {
				properties[resp.Name] = map[string]string{
					"type":        resp.Type,
					"description": resp.Description,
				}
			}
			apiResponse["properties"] = properties
		}
		tags := []string{}
		if len(api.Group) > 0 {
			tags = append(tags, api.Group)
		}
		paths[api.Path] = map[string]interface{}{
			strings.ToLower(api.Method): map[string]interface{}{
				"summary":    api.Description,
				"parameters": parameters,
				"produces": []string{
					"application/json",
					"text/plain",
					"application/xml",
				},
				"tags": tags,
				"responses": map[string]interface{}{
					"default": map[string]interface{}{
						"schema": apiResponse,
					},
				},
			},
		}
	}
	swaggerJson["paths"] = paths
	result, _ := json.Marshal(swaggerJson)
	//result, _ := yml.Marshal(swaggerJson)
	return string(result)
}

const swaggerHtml = `
<!-- HTML for static distribution bundle build -->
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <title>Swagger UI</title>
    <link rel="stylesheet" type="text/css" href="/static/swagger-ui" />
    <style>
      html
      {
        box-sizing: border-box;
        overflow: -moz-scrollbars-vertical;
        overflow-y: scroll;
      }

      *,
      *:before,
      *:after
      {
        box-sizing: inherit;
      }

      body
      {
        margin:0;
        background: #fafafa;
      }
    </style>
  </head>

  <body>
    <div id="swagger-ui"></div>
    <script src="/static/swagger-ui-bundle" charset="UTF-8"> </script>
    <script>
    window.onload = function() {
      // Begin Swagger UI call region
      const ui = SwaggerUIBundle({
        url: "%s",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout"
      });
      // End Swagger UI call region
      window.ui = ui;
    };
  </script>
  </body>
</html>
`

// path参数需指定http://host:port
func RegisterSwagger(data SwaggerData) {

	RegisterHandler("/static/swagger-ui-bundle.js", func(context Context) {
		context.OK(Js, []byte(SwaggerJs))
	})

	RegisterHandler("/static/swagger-ui.css", func(context Context) {
		context.OK(Css, []byte(SwaggerCss))
	})

	RegisterHandler("/swagger-ui", func(context Context) {
		context.OK(Html, []byte(fmt.Sprintf(swaggerHtml, "/swagger-ui.json")))
	})

	RegisterHandler("/swagger-ui.json", func(context Context) {
		context.OK(Json, []byte(GenerateSwagger(data)))
	})

}

// 根据Swagger配置文件, 生成swagger文档
func (this *Server) EnableSwaggerWithConf(swaggerConf string) {
	if !Exists(swaggerConf) {
		mLogger.ErrorF("swagger 配置文件不存在在: %v", swaggerConf)
		return
	}
	fileData, err := ioutil.ReadFile(swaggerConf)
	if err != nil {
		mLogger.ErrorF("swagger 配置文件读取错误: %v, %v", swaggerConf, err.Error())
		return
	}

	conf := SwaggerData{}

	err = json.Unmarshal(fileData, &conf)
	if err != nil {
		mLogger.ErrorF("swagger 配置文件读取错误, 错误的结构: %v, %v", string(fileData), err.Error())
		return
	}

	this.RegisterHandler("/static/swagger-ui-bundle.js", func(context Context) {
		context.OK(Js, []byte(SwaggerJs))
	})

	this.RegisterHandler("/static/swagger-ui.css", func(context Context) {
		context.OK(Css, []byte(SwaggerCss))
	})

	this.RegisterHandler("/swagger-ui", func(context Context) {
		context.OK(Html, []byte(fmt.Sprintf(swaggerHtml, "/swagger-ui.json")))
	})

	this.RegisterHandler("/swagger-ui.json", func(context Context) {
		context.OK(Json, []byte(GenerateSwagger(*this.swagger)))
	})
}

// 启动swagger服务
func (this *Server) EnableSwagger(swaggerData SwaggerData) {

	this.RegisterHandler("/static/swagger-ui-bundle", func(context Context) {
		context.OK(Js, []byte(SwaggerJs))
	})

	this.RegisterHandler("/static/swagger-ui", func(context Context) {
		context.OK(Css, []byte(SwaggerCss))
	})

	this.RegisterHandler("/swagger-ui", func(context Context) {
		context.OK(Html, []byte(fmt.Sprintf(swaggerHtml, "/swagger-ui.json")))
	})

	this.RegisterHandler("/swagger-ui.json", func(context Context) {
		context.OK(Json, []byte(GenerateSwagger(swaggerData)))
	})
}

func EnableSwagger(data SwaggerData) {
	globalServer.EnableSwagger(data)
}

func EnableSwaggerWithConf(swaggerConf string) {
	globalServer.EnableSwaggerWithConf(swaggerConf)
}
