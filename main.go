/*
 * @Author: ilikara 3435193369@qq.com
 * @Date: 2024-12-29 12:43:00
 * @LastEditors: ilikara 3435193369@qq.com
 * @LastEditTime: 2024-12-31 16:21:30
 * @FilePath: /my_eagle/main.go
 * @Description:
 *
 * Copyright (c) 2024 by ${git_name_email}, All Rights Reserved.
 */
package main

import (
	"log"
	"path/filepath"

	"my_eagle/api/folderapi"
	"my_eagle/api/itemapi"
	"my_eagle/database"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// ServeImage 根据 id 返回图片
func ServeImage(c *gin.Context) {
	// 从 URL 路由获取 id
	id := c.Param("id")

	// 构建图片路径
	imagePath := filepath.Join(database.DbBaseDir, "thumbnails", id+".webp")

	// 检查文件是否存在（可以进一步完善此部分）
	c.File(imagePath)
}

func main() {
	_, err := database.Database_init("test")

	if err != nil {
		log.Fatalf("failed init database: %v", err)
	}

	// // TEST
	// itemdb.AddItem(database.DB, "test/OvO/1.png", nil, nil, nil, nil, nil, nil, nil)
	// itemdb.AddItem(database.DB, "test/OvO/1.jpg", nil, nil, nil, nil, nil, nil, nil)
	// itemdb.AddItem(database.DB, "test/OvO/2.png", nil, nil, nil, nil, nil, nil, nil)
	// itemdb.AddItem(database.DB, "test/OvO/2.jpg", nil, nil, nil, nil, nil, nil, nil)
	// itemdb.AddItem(database.DB, "test/OvO/3.png", nil, nil, nil, nil, nil, nil, nil)
	// itemdb.AddItem(database.DB, "test/OvO/3.jpg", nil, nil, nil, nil, nil, nil, nil)
	// itemdb.ItemHardDelete(database.DB, []string{"f81723419c242656ef53b4eeb471bf97909193bd58e407f969f9b4e6748f26de"})

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

	// 设置路由，使用 /image/:id 路由，:id 表示动态参数
	r.GET("/image/:id", ServeImage)

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
