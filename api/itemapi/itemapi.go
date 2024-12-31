/*
 * @Author: ilikara 3435193369@qq.com
 * @Date: 2024-12-29 12:43:00
 * @LastEditors: ilikara 3435193369@qq.com
 * @LastEditTime: 2024-12-31 10:53:14
 * @FilePath: /my_eagle/api/itemapi/item.go
 * @Description:
 *
 * Copyright (c) 2024 by ${git_name_email}, All Rights Reserved.
 */
package itemapi

import (
	"fmt"
	"io"
	"my_eagle/database"
	"my_eagle/database/itemdb"
	"net/http"
	"os"
	"path"
	"strings"
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

	TagIds    []uuid.UUID `json:"tag_id"`    // Tags ID列表
	FolderIds []uuid.UUID `json:"folder_id"` // 文件夹ID列表

	// Palettes []uint32 `json:"palettes"` // 色票（这是什么？）
	Star uint8 `json:"star"` // 星级评分

	HaveThumbnail bool `json:"have_thumbnail"` // 是否有缩略图
	HavePreview   bool `json:"have_preview"`   // 是否有预览图
}

type ItemResponse struct {
	Status string `json:"status"`
	Data   []Item `json:"data"`
}

func AddFromUrls(c *gin.Context) {
	// 解析请求数据
	var req struct {
		Items []struct {
			URL              string            `json:"url" binding:"required"` // 图片链接
			Name             *string           `json:"name"`                   // 图片名称
			Website          *string           `json:"website"`                // 来源网址
			Annotation       *string           `json:"annotation"`             // 注释
			Tags             []uuid.UUID       `json:"tags"`                   // 标签
			ModificationTime *time.Time        `json:"modificationTime"`       // 修改时间
			Headers          map[string]string `json:"headers"`                // 自定义 HTTP headers
		} `json:"items" binding:"required"` // 图片信息列表
		FolderIDs []uuid.UUID `json:"folderId"`                 // 可选，文件夹 ID
		Token     string      `json:"token" binding:"required"` // API Token
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request data",
		})
		return
	}

	if req.Token != database.Token {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Invalid token",
		})
		return
	}

	// 循环处理每个图片
	for _, item := range req.Items {
		filePath, err := saveFileFromURL(item.URL, item.Headers)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to download file: %v", err)})
			return
		}
		defer os.Remove(filePath) // 确保请求结束后删除临时文件

		// 将文件路径传递给 AddItem 函数
		star := uint8(0)
		err = itemdb.AddItem(database.DB, filePath, item.Name, item.Website, item.Annotation, item.Tags, req.FolderIDs, &star, item.ModificationTime)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to add item: %v", err)})
			return
		}
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// saveFileFromURL 下载文件并保存，返回文件路径
func saveFileFromURL(url string, headers map[string]string) (string, error) {
	// 发起 HTTP 请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	// 添加请求头
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	// 尝试从响应头的 Content-Disposition 中提取文件名
	fileName := getFileNameFromContentDisposition(resp.Header.Get("Content-Disposition"))
	if fileName == "" {
		// 如果响应头没有文件名，则从 URL 中推断文件名
		fileName = getFileNameFromURL(url)
	}

	// 如果文件名仍然为空，使用文件内容的 SHA256 哈希值
	if fileName == "" {
		fileName, err = itemdb.CalculateSHA256(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to calc hash: %v", err)
		}
	}

	// 保存文件
	filePath := path.Join("/tmp", fileName) // 可以自定义路径
	outFile, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %v", err)
	}

	return filePath, nil
}

// 从 URL 提取文件名（如果可以）
func getFileNameFromURL(url string) string {
	// 这里假设 URL 末尾可能包含文件名
	segments := strings.Split(url, "/")
	if len(segments) > 0 {
		return segments[len(segments)-1]
	}
	return ""
}

// 从 Content-Disposition 头提取文件名
func getFileNameFromContentDisposition(contentDisposition string) string {
	// 解析 Content-Disposition 格式: attachment; filename="example.jpg"
	parts := strings.Split(contentDisposition, "filename=")
	if len(parts) > 1 {
		return strings.Trim(parts[1], "\"")
	}
	return ""
}

func AddFromPaths(c *gin.Context) {

}

func Info(c *gin.Context) {

}

func MoveToTrash(c *gin.Context) {
	var req struct {
		ItemIDs []string `json:"item_ids"`
		Token   string   `json:"token" binding:"required"`
	}

	// 绑定JSON数据到结构体
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request data",
		})
		return
	}

	if req.Token != database.Token {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Invalid token",
		})
		return
	}
	err := itemdb.ItemSoftDelete(database.DB, req.ItemIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed"})
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func Update(c *gin.Context) {

}

func List(c *gin.Context) {
	var req struct {
		Limit     *int        `json:"limit"`
		Offset    *int        `json:"offset"`
		OrderBy   *string     `json:"order_by"`
		Exts      []string    `json:"exts"`
		Keyword   *string     `json:"keyword"`
		TagIDs    []uuid.UUID `json:"tags"`
		FolderIDs []uuid.UUID `json:"folder_ids"`
		IsDeleted *bool       `json:"is_deleted"`
		Token     string      `json:"token" binding:"required"`
	}

	if req.Token != database.Token {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Invalid token",
		})
		return
	}
	// 绑定JSON数据到结构体
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request data",
		})
		return
	}

	items, err := itemdb.ItemList(database.DB, req.IsDeleted, req.OrderBy, req.Offset, req.Limit, req.Exts, req.Keyword, req.TagIDs, req.FolderIDs)
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

			HaveThumbnail: item.HaveThumbnail,
			HavePreview:   item.HavePreview,
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
