package middleware

import (
	"fmt"
	"github.com/wenlaizhou/yml"
	"strings"
)

type SwaggerPath struct {
	Path        string
	Method      string
	Description string
	Parameters  []SwaggerParameter
}

type SwaggerParameter struct {
	Name        string
	Description string
	Example     string
	Default     string
	Type        string
	Required    bool
}

type SwaggerData struct {
	Title       string
	Version     string
	Description string
	Host        string
	Apis        []SwaggerPath
}

func (thisSelf SwaggerData) AddPath(path SwaggerPath) {
	thisSelf.Apis = append(thisSelf.Apis, path)
}

func (thisSelf SwaggerPath) AddParameter(param SwaggerParameter) {
	thisSelf.Parameters = append(thisSelf.Parameters, param)
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
		var parameters []map[string]interface{}
		if api.Parameters != nil && len(api.Parameters) > 0 {
			for _, p := range api.Parameters {
				parameters = append(parameters, map[string]interface{}{
					"name":        p.Name,
					"description": p.Description,
					"required":    p.Required,
					"type":        p.Type,
					"example":     p.Example,
				})
			}
		}
		paths[api.Path] = map[string]interface{}{
			strings.ToLower(api.Method): map[string]interface{}{
				"summary":    api.Description,
				"parameters": parameters,
			},
		}
	}
	swaggerJson["paths"] = paths
	result, _ := yml.Marshal(swaggerJson)
	return string(result)
}

const swaggerHtml = `
<!-- HTML for static distribution bundle build -->
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <title>Swagger UI</title>
    <link rel="stylesheet" type="text/css" href="/static/swagger-ui.css" />
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
    <script src="/static/swagger-ui-bundle.js" charset="UTF-8"> </script>
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
func RegisterSwagger(path string) {

	RegisterHandler("/static/swagger-ui-bundle.js", func(context Context) {
		context.OK(Js, []byte(SwaggerJs))
	})

	RegisterHandler("/static/swagger-ui.css", func(context Context) {
		context.OK(Css, []byte(SwaggerCss))
	})

	RegisterHandler("/swagger-ui", func(context Context) {
		context.OK(Html, []byte(fmt.Sprintf(swaggerHtml, path)))
	})

}
