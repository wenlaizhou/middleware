package middleware

import "fmt"

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

	example := `
{
    "swagger": "2.0",
    "info": {
        "description": "This is a sample server Petstore server.  You can find out more about Swagger at [http://swagger.io](http://swagger.io) or on [irc.freenode.net, #swagger](http://swagger.io/irc/).  For this sample, you can use the api key "special-key" to test the authorization filters.",
        "version": "1.0.5",
        "title": "Swagger Petstore",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {
            "email": "apiteam@swagger.io"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        }
    },
    "host": "petstore.swagger.io",
    "basePath": "/v2",
    "tags": [
        {
            "name": "pet",
            "description": "Everything about your Pets",
            "externalDocs": {
                "description": "Find out more",
                "url": "http://swagger.io"
            }
        },
        {
            "name": "store",
            "description": "Access to Petstore orders"
        },
        {
            "name": "user",
            "description": "Operations about user",
            "externalDocs": {
                "description": "Find out more about our store",
                "url": "http://swagger.io"
            }
        }
    ],
    "schemes": [
        "https",
        "http"
    ],
    "paths": {
        "/pet/{petId}/uploadImage": {
            "post": {
                "tags": [
                    "pet"
                ],
                "summary": "uploads an image",
                "description": "",
                "operationId": "uploadFile",
                "consumes": [
                    "multipart/form-data"
                ],
                "produces": [
                    "application/json"
                ],
                "parameters": [
                    {
                        "name": "petId",
                        "in": "path",
                        "description": "ID of pet to update",
                        "required": true,
                        "type": "integer",
                        "format": "int64"
                    },
                    {
                        "name": "additionalMetadata",
                        "in": "formData",
                        "description": "Additional data to pass to server",
                        "required": false,
                        "type": "string"
                    },
                    {
                        "name": "file",
                        "in": "formData",
                        "description": "file to upload",
                        "required": false,
                        "type": "file"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "successful operation",
                        "schema": {
                            "$ref": "#/definitions/ApiResponse"
                        }
                    }
                },
                "security": [
                    {
                        "petstore_auth": [
                            "write:pets",
                            "read:pets"
                        ]
                    }
                ]
            }
        },
        "/pet/findByStatus": {
            "get": {
                "tags": [
                    "pet"
                ],
                "summary": "Finds Pets by status",
                "description": "Multiple status values can be provided with comma separated strings",
                "operationId": "findPetsByStatus",
                "produces": [
                    "application/json",
                    "application/xml"
                ],
                "parameters": [
                    {
                        "name": "status",
                        "in": "query",
                        "description": "Status values that need to be considered for filter",
                        "required": true,
                        "type": "array",
                        "items": {
                            "type": "string",
                            "enum": [
                                "available",
                                "pending",
                                "sold"
                            ],
                            "default": "available"
                        },
                        "collectionFormat": "multi"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "successful operation",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/Pet"
                            }
                        }
                    },
                    "400": {
                        "description": "Invalid status value"
                    }
                },
                "security": [
                    {
                        "petstore_auth": [
                            "write:pets",
                            "read:pets"
                        ]
                    }
                ]
            }
        },
        "/user/{username}": {
            "get": {
                "tags": [
                    "user"
                ],
                "summary": "Get user by user name",
                "description": "",
                "operationId": "getUserByName",
                "produces": [
                    "application/json",
                    "application/xml"
                ],
                "parameters": [
                    {
                        "name": "username",
                        "in": "path",
                        "description": "The name that needs to be fetched. Use user1 for testing. ",
                        "required": true,
                        "type": "string"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "successful operation",
                        "schema": {
                            "$ref": "#/definitions/User"
                        }
                    },
                    "400": {
                        "description": "Invalid username supplied"
                    },
                    "404": {
                        "description": "User not found"
                    }
                }
            },
            "put": {
                "tags": [
                    "user"
                ],
                "summary": "Updated user",
                "description": "This can only be done by the logged in user.",
                "operationId": "updateUser",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json",
                    "application/xml"
                ],
                "parameters": [
                    {
                        "name": "username",
                        "in": "path",
                        "description": "name that need to be updated",
                        "required": true,
                        "type": "string"
                    },
                    {
                        "in": "body",
                        "name": "body",
                        "description": "Updated user object",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/User"
                        }
                    }
                ],
                "responses": {
                    "400": {
                        "description": "Invalid user supplied"
                    },
                    "404": {
                        "description": "User not found"
                    }
                }
            },
            "delete": {
                "tags": [
                    "user"
                ],
                "summary": "Delete user",
                "description": "This can only be done by the logged in user.",
                "operationId": "deleteUser",
                "produces": [
                    "application/json",
                    "application/xml"
                ],
                "parameters": [
                    {
                        "name": "username",
                        "in": "path",
                        "description": "The name that needs to be deleted",
                        "required": true,
                        "type": "string"
                    }
                ],
                "responses": {
                    "400": {
                        "description": "Invalid username supplied"
                    },
                    "404": {
                        "description": "User not found"
                    }
                }
            }
        },
        "/user/login": {
            "get": {
                "tags": [
                    "user"
                ],
                "summary": "Logs user into the system",
                "description": "",
                "operationId": "loginUser",
                "produces": [
                    "application/json",
                    "application/xml"
                ],
                "parameters": [
                    {
                        "name": "username",
                        "in": "query",
                        "description": "The user name for login",
                        "required": true,
                        "type": "string"
                    },
                    {
                        "name": "password",
                        "in": "query",
                        "description": "The password for login in clear text",
                        "required": true,
                        "type": "string"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "successful operation",
                        "headers": {
                            "X-Expires-After": {
                                "type": "string",
                                "format": "date-time",
                                "description": "date in UTC when token expires"
                            },
                            "X-Rate-Limit": {
                                "type": "integer",
                                "format": "int32",
                                "description": "calls per hour allowed by the user"
                            }
                        },
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Invalid username/password supplied"
                    }
                }
            }
        },
        "/user/logout": {
            "get": {
                "tags": [
                    "user"
                ],
                "summary": "Logs out current logged in user session",
                "description": "",
                "operationId": "logoutUser",
                "produces": [
                    "application/json",
                    "application/xml"
                ],
                "parameters": [],
                "responses": {
                    "default": {
                        "description": "successful operation"
                    }
                }
            }
        },
        "/user/createWithArray": {
            "post": {
                "tags": [
                    "user"
                ],
                "summary": "Creates list of users with given input array",
                "description": "",
                "operationId": "createUsersWithArrayInput",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json",
                    "application/xml"
                ],
                "parameters": [
                    {
                        "in": "body",
                        "name": "body",
                        "description": "List of user object",
                        "required": true,
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/User"
                            }
                        }
                    }
                ],
                "responses": {
                    "default": {
                        "description": "successful operation"
                    }
                }
            }
        },
        "/user": {
            "post": {
                "tags": [
                    "user"
                ],
                "summary": "Create user",
                "description": "This can only be done by the logged in user.",
                "operationId": "createUser",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json",
                    "application/xml"
                ],
                "parameters": [
                    {
                        "in": "body",
                        "name": "body",
                        "description": "Created user object",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/User"
                        }
                    }
                ],
                "responses": {
                    "default": {
                        "description": "successful operation"
                    }
                }
            }
        }
    },
    "securityDefinitions": {
        "api_key": {
            "type": "apiKey",
            "name": "api_key",
            "in": "header"
        },
        "petstore_auth": {
            "type": "oauth2",
            "authorizationUrl": "https://petstore.swagger.io/oauth/authorize",
            "flow": "implicit",
            "scopes": {
                "read:pets": "read your pets",
                "write:pets": "modify pets in your account"
            }
        }
    },
    "definitions": {
        "ApiResponse": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "integer",
                    "format": "int32"
                },
                "type": {
                    "type": "string"
                },
                "message": {
                    "type": "string"
                }
            }
        },
        "Category": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "integer",
                    "format": "int64"
                },
                "name": {
                    "type": "string"
                }
            },
            "xml": {
                "name": "Category"
            }
        },
        "Tag": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "integer",
                    "format": "int64"
                },
                "name": {
                    "type": "string"
                }
            },
            "xml": {
                "name": "Tag"
            }
        },
        "User": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "integer",
                    "format": "int64"
                },
                "username": {
                    "type": "string"
                },
                "firstName": {
                    "type": "string"
                },
                "lastName": {
                    "type": "string"
                },
                "email": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                },
                "phone": {
                    "type": "string"
                },
                "userStatus": {
                    "type": "integer",
                    "format": "int32",
                    "description": "User Status"
                }
            },
            "xml": {
                "name": "User"
            }
        }
    },
    "externalDocs": {
        "description": "Find out more about Swagger",
        "url": "http://swagger.io"
    }
}
`

	RegisterHandler(path, func(context Context) {
		context.OK(ApplicationJson, []byte(example))
	})

}
