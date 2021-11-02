package middleware

import (
	"fmt"
	"github.com/wenlaizhou/yml"
)

type Swagger struct {
	Swagger string `json:"swagger"`
	Info    struct {
		Title       string `json:"title"`
		Version     string `json:"version"`
		Description string `json:"description"`
	} `json:"info"`
	Tags []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"tags"`
	Host  string `json:"host"`
	Paths map[string]struct {
	} `json:"paths"`
}

type SwaggerPath struct {
	Tags        []string           `json:"tags"`
	Path        string             `json:"path"`
	Description string             `json:"description"`
	Method      string             `json:"method"`
	Summary     string             `json:"summary"`
	Parameters  []SwaggerParameter `json:"parameters"`
}

//      parameters:
//      - in: "body"
//        name: "body"
//        description: "Pet object that needs to be added to the store"
//        required: true
//        schema:
//          $ref: "#/definitions/Pet"

type SwaggerParameter struct {
	In          string `json:"in"` //query body path
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"` //array, string, integer
	Required    bool   `json:"required"`
}

func SwaggerGenerate(swagger Swagger) string {
	model := map[string]interface{}{}
	model["swagger"] = "2.0"
	model["info"] = swagger
	res, _ := yml.Marshal(model)
	return string(res)
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
