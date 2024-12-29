package main

import (
	"log"
	"os"

	"my_eagle/api/folder"
	"my_eagle/api/item"
	"my_eagle/database"

	"github.com/gin-gonic/gin"
)

func main() {
	_, err := database.Database_init("test")

	if err != nil {
		log.Fatalf("failed init database: %v", err)
	}

	// 创建文件存储和缩略图存储目录
	os.MkdirAll("files", os.ModePerm)
	os.MkdirAll("thumbnails", os.ModePerm)

	// 启动 Gin Web 框架
	r := gin.Default()

	r.POST("/api/folder/create", folder.CreateFolder)
	r.POST("/api/folder/list", folder.ListFolder)
	r.POST("/api/folder/update", folder.UpdateFolder)

	// 预想tag和folder逻辑一致
	r.POST("/api/tag/create", folder.CreateFolder)
	r.POST("/api/tag/list", folder.ListFolder)
	r.POST("/api/tag/update", folder.UpdateFolder)

	r.POST("/api/item/addFromUrls", item.AddFromUrls)
	r.POST("/api/item/addFromPaths", item.AddFromPaths)
	r.POST("/api/item/info", item.Info)
	r.POST("/api/item/moveToTrash", item.MoveToTrash)
	r.POST("/api/item/update", item.Update)
	r.POST("/api/item/list", item.List)

	// 启动服务
	r.Run(":8080")
}
