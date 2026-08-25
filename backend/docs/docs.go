// Package docs Swagger 文档占位文件。
// 运行 `make swag`（swag init -g cmd/server/main.go -o docs）后会重新生成本文件。
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
  "openapi": "3.0.1",
  "info": {
    "title": "FriendBook AI API",
    "description": "智能好友管理系统后端接口文档",
    "version": "1.0.0"
  },
  "basePath": "/",
  "paths": {}
}`

// SwaggerInfo Swagger 元信息
var SwaggerInfo = &swag.Spec{
	Version:          "1.0.0",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "FriendBook AI API",
	Description:      "智能好友管理系统后端接口文档",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
