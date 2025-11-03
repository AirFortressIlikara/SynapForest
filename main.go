/*
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

	"synapforest/api"
	"synapforest/api/folderapi"
	"synapforest/api/graphql"
	"synapforest/api/itemapi"
	"synapforest/api/tagapi"
	"synapforest/api/vectorapi"
	"synapforest/database"

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

		privateRoutes.POST("/tag/create", tagapi.CreateTag)
		privateRoutes.POST("/tag/list", tagapi.ListTag)
		privateRoutes.POST("/tag/update", tagapi.UpdateTag)
		privateRoutes.POST("/tag/updateParent", tagapi.UpdateTagsParent)
		privateRoutes.POST("/tag/delete", tagapi.DeleteTag)

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
