/*
 * @Author: ilikara 3435193369@qq.com
 * @Date: 2024-12-29 12:43:00
 * @LastEditors: ilikara 3435193369@qq.com
 * @LastEditTime: 2024-12-30 06:42:37
 * @FilePath: /my_eagle/api/item/item.go
 * @Description:
 *
 * Copyright (c) 2024 by ${git_name_email}, All Rights Reserved.
 */
package item

import (
	"fmt"
	"my_eagle/database"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type Item struct {
	ID         string         `json:"id" gorm:"primaryKey"` // 主键，文件的Hash
	CreatedAt  time.Time      `json:"created_at"`           // 创建时间
	ImportedAt time.Time      `json:"imported_at"`          // 导入时间
	ModifiedAt time.Time      `json:"modified_at"`          // 修改时间
	DeletedAt  gorm.DeletedAt `json:"deleted_at"`           // 删除时间

	Name string `json:"name"` // 名称
	Ext  string `json:"ext"`  // 扩展名

	Width  uint32 `json:"width"`  // 宽度
	Height uint32 `json:"height"` // 高度
	Size   uint64 `json:"size"`   // 文件大小

	Url        string `json:"url"`        // 文件来源URL
	Annotation string `json:"annotation"` // 注释

	TagIds    []uuid.UUID `json:"tag_id"`    // Tags
	FolderIds []uuid.UUID `json:"folder_id"` // 文件夹ID列表

	// Palettes []uint32 `json:"palettes"` // 色票（这是什么？）
	Star uint8 `json:"star"` // 星级评分

	NoThumbnail   bool   `json:"no_thumbnail"`   // 是否有缩略图
	NoPreview     bool   `json:"no_preview"`     // 是否有预览图
	FilePath      string `json:"file_path"`      // 文件路径
	FileUrl       string `json:"file_url"`       // 文件URL
	ThumbnailPath string `json:"thumbnail_path"` // 缩略图路径
	ThumbnailUrl  string `json:"thumbnail_url"`  // 缩略图URL
}
type ItemResponse struct {
	Status string `json:"status"`
	Data   []Item `json:"data"`
}

func AddFromUrls(c *gin.Context) {

}

func AddFromPaths(c *gin.Context) {

}

func Info(c *gin.Context) {

}

func MoveToTrash(c *gin.Context) {

}

func Update(c *gin.Context) {

}

type ItemListRequest struct {
	OrderBy   string      `json:"order_by"`
	Exts      []string    `json:"exts"`
	Keyword   string      `json:"keyword"`
	TagIDs    []uuid.UUID `json:"tags"`
	FolderIDs []uuid.UUID `json:"folder_ids"`
}

func List(c *gin.Context) {
	var req ItemListRequest

	// 绑定JSON数据到结构体
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request data",
		})
		return
	}

	items, err := database.ItemList(database.DB, req.OrderBy, req.Exts, req.Keyword, req.TagIDs, req.FolderIDs)
	if err != nil {
		// 返回JSON响应
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Item Query Failed",
		})
		return
	}

	result := database.DB.Preload("Tags").Preload("Folders").Find(&items)

	if result.Error != nil {
		// 错误处理
		fmt.Println("Error loading items:", result.Error)
	}

	// 构造返回值
	resp := ItemResponse{
		Status: "success",
	}

	// 遍历每个 Item，转换为 ItemResponse.Data 中的元素
	for _, item := range items {
		// 构建 ItemResponse.Data
		dataItem := Item{
			ID:         item.ID,
			CreatedAt:  item.CreatedAt,
			ImportedAt: item.ImportedAt,
			ModifiedAt: item.ModifiedAt,
			DeletedAt:  item.DeletedAt,

			Name: item.Name,
			Ext:  item.Ext,

			Width:  item.Width,
			Height: item.Height,
			Size:   item.Size,

			Url:        item.Url,
			Annotation: item.Annotation,

			Star: item.Star,

			NoThumbnail: item.NoThumbnail,
			NoPreview:   item.NoPreview,

			FilePath:      item.FilePath,
			FileUrl:       item.FileUrl,
			ThumbnailPath: item.ThumbnailPath,
			ThumbnailUrl:  item.ThumbnailUrl,
		}

		// 提取 Tags 和 Folders 的 ID
		for _, tag := range item.Tags {
			dataItem.TagIds = append(dataItem.TagIds, tag.ID)
		}
		for _, folder := range item.Folders {
			dataItem.FolderIds = append(dataItem.FolderIds, folder.ID)
		}
		resp.Data = append(resp.Data, dataItem)
	}
	c.JSON(http.StatusOK, resp)
}
