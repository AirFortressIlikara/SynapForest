/*
 * @Author: Ilikara 3435193369@qq.com
 * @Date: 2025-01-09 19:59:53
 * @LastEditors: Ilikara 3435193369@qq.com
 * @LastEditTime: 2025-01-27 20:43:39
 * @FilePath: /my_eagle/main.go
 * @Description:
 *
 * Copyright (c) 2025 AirFortressIlikara
 * SynapForest is licensed under Mulan PubL v2.
 * You can use this software according to the terms and conditions of the Mulan PubL v2.
 * You may obtain a copy of Mulan PubL v2 at:
 *          http://license.coscl.org.cn/MulanPubL-2.0
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PubL v2 for more details.
 */
package main

import (
	"log"

	"my_eagle/api"
	"my_eagle/api/folderapi"
	"my_eagle/api/itemapi"
	"my_eagle/database"

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

	r.Use(api.AuthMiddleware())

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

	// 注册 API
	r.POST("/api/item/remove-folder", api.RemoveFolderForItems)
	r.POST("/api/item/add-folder", api.AddFolderForItems)

	// 启动服务
	r.Run(":41595")
}
