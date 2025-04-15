/*
 * @Author: Ilikara 3435193369@qq.com
 * @Date: 2025-01-09 19:59:53
 * @LastEditors: ilikara 3435193369@qq.com
 * @LastEditTime: 2025-04-14 16:08:34
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
	"my_eagle/api/graphql"
	"my_eagle/api/itemapi"
	"my_eagle/api/vectorapi"
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

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // 允许所有域名
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	publicRoutes := r.Group("/public")
	{
		publicRoutes.GET("/thumbnails/:id", api.ServeThumbnails)
		publicRoutes.GET("/raw_files/:id", api.ServeRawFile)
		publicRoutes.GET("/previews/:id", api.ServePreviews)

		publicRoutes.POST("/vectorize/:id", vectorapi.HandleVectorize)
	}

	privateRoutes := r.Group("/api")
	privateRoutes.Use(api.AuthMiddleware())
	{
		privateRoutes.POST("/uploadfiles", api.Uploadfiles)

		privateRoutes.POST("/folder/create", folderapi.CreateFolder)
		privateRoutes.POST("/folder/list", folderapi.ListFolder)
		privateRoutes.POST("/folder/update", folderapi.UpdateFolder)
		privateRoutes.POST("/folder/updateParent", folderapi.UpdateFoldersParent)
		privateRoutes.POST("/folder/delete", folderapi.DeleteFolder)

		// 预想tag和folder逻辑一致
		privateRoutes.POST("/tag/create", folderapi.CreateFolder)
		privateRoutes.POST("/tag/list", folderapi.ListFolder)
		privateRoutes.POST("/tag/update", folderapi.UpdateFolder)

		privateRoutes.POST("/item/addFromUrls", itemapi.AddFromUrls)
		privateRoutes.POST("/item/addFromPaths", itemapi.AddFromPaths)
		privateRoutes.POST("/item/info", itemapi.Info)
		privateRoutes.POST("/item/moveToTrash", itemapi.MoveToTrash)
		privateRoutes.POST("/item/update", itemapi.Update)
		privateRoutes.POST("/item/list", itemapi.List)

		privateRoutes.POST("/item/remove-folder", api.RemoveFolderForItems)
		privateRoutes.POST("/item/add-folder", api.AddFolderForItems)

		gqlHandler := graphql.NewHandler()
		r.GET("/graphql", gin.WrapH(gqlHandler))
		r.POST("/graphql", gin.WrapH(gqlHandler))
	}

	r.Run(":41595")
}
