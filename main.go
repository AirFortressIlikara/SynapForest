/*
 * @Author: ilikara 3435193369@qq.com
 * @Date: 2024-12-29 12:43:00
 * @LastEditors: ilikara 3435193369@qq.com
 * @LastEditTime: 2025-01-01 16:00:11
 * @FilePath: /my_eagle/main.go
 * @Description:
 *
 * Copyright (c) 2024 by ${git_name_email}, All Rights Reserved.
 */
package main

import (
	"log"

	"my_eagle/api"
	"my_eagle/api/folderapi"
	"my_eagle/api/itemapi"
	"my_eagle/database"
	"my_eagle/database/itemdb"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	_, err := database.Database_init("test")
	if err != nil {
		log.Fatalf("failed init database: %v", err)
	}

	err = api.ApiInit("uploads")
	if err != nil {
		log.Fatalf("failed init Api: %v", err)
	}

	// Add TEST Case
	itemdb.AddItem(database.DB, "test/OvO/2c05709ef84692f0d3c3c16fb7814068.gif", nil, nil, nil, nil, nil, nil, nil)
	itemdb.AddItem(database.DB, "test/OvO/8d8b7ef81781a9db389de8302cf0c2db.png", nil, nil, nil, nil, nil, nil, nil)
	itemdb.AddItem(database.DB, "test/OvO/32c64c0d90bf96f8051f88635caf92c8.png", nil, nil, nil, nil, nil, nil, nil)
	itemdb.AddItem(database.DB, "test/OvO/830ef7585f5b7b3743fb876b1486b806.jpg", nil, nil, nil, nil, nil, nil, nil)
	itemdb.AddItem(database.DB, "test/OvO/6597b9e03261ee039cc980a9be00a155.gif", nil, nil, nil, nil, nil, nil, nil)
	itemdb.AddItem(database.DB, "test/OvO/168056bc504ef92dcc9df184aad6fa1d.gif", nil, nil, nil, nil, nil, nil, nil)
	itemdb.AddItem(database.DB, "test/OvO/a35cb6e60df1f0057f4d1ab441ea2e67.jpg", nil, nil, nil, nil, nil, nil, nil)
	itemdb.AddItem(database.DB, "test/OvO/acb12c4029e0b9bf99419bdd47376dbb.gif", nil, nil, nil, nil, nil, nil, nil)
	itemdb.AddItem(database.DB, "test/OvO/c3dead9e9eddb0504a2707cf882bb4f4.jpg", nil, nil, nil, nil, nil, nil, nil)
	itemdb.AddItem(database.DB, "test/OvO/c6e088baee28d16cf77e52c7aa1a0cf5.jpg", nil, nil, nil, nil, nil, nil, nil)
	itemdb.AddItem(database.DB, "test/OvO/cb6509f98138e1c22eb086ca9f9ecf94.png", nil, nil, nil, nil, nil, nil, nil)
	itemdb.AddItem(database.DB, "test/OvO/f9472f46f618345019e16960f357f8cd.png", nil, nil, nil, nil, nil, nil, nil)

	// 启动 Gin Web 框架
	r := gin.Default()

	// 配置 CORS 中间件
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // 允许所有域名
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 设置路由，:id 表示动态参数
	r.GET("/thumbnails/:id", api.ServeThumbnails)
	r.GET("/raw_files/:id", api.ServeRawFile)
	r.GET("/previews/:id", api.ServePreviews)

	r.POST("/api/uploadfiles", api.Uploadfiles)

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
	r.Run(":41595")
}
