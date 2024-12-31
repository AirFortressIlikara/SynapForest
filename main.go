/*
 * @Author: ilikara 3435193369@qq.com
 * @Date: 2024-12-29 12:43:00
 * @LastEditors: ilikara 3435193369@qq.com
 * @LastEditTime: 2024-12-31 09:15:07
 * @FilePath: /my_eagle/main.go
 * @Description:
 *
 * Copyright (c) 2024 by ${git_name_email}, All Rights Reserved.
 */
package main

import (
	"log"

	"my_eagle/api/folderapi"
	"my_eagle/api/itemapi"
	"my_eagle/database"

	"github.com/gin-gonic/gin"
)

func main() {
	_, err := database.Database_init("test")

	if err != nil {
		log.Fatalf("failed init database: %v", err)
	}

	// 启动 Gin Web 框架
	r := gin.Default()

	r.POST("/api/folder/create", folderapi.CreateFolder)
	r.POST("/api/folder/list", folderapi.ListFolder)
	r.POST("/api/folder/update", folderapi.UpdateFolder)

	// 预想tag和folder逻辑一致
	r.POST("/api/tag/create", folderapi.CreateFolder)
	r.POST("/api/tag/list", folderapi.ListFolder)
	r.POST("/api/tag/update", folderapi.UpdateFolder)

	r.POST("/api/item/addFromUrls", itemapi.AddFromUrls)
	r.POST("/api/item/addFromPaths", itemapi.AddFromPaths)
	r.POST("/api/item/info", itemapi.Info)
	r.POST("/api/item/moveToTrash", itemapi.MoveToTrash)
	r.POST("/api/item/update", itemapi.Update)
	r.POST("/api/item/list", itemapi.List)

	// 启动服务
	r.Run(":8080")
}
