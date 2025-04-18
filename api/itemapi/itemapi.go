/*
 * @Author: Ilikara 3435193369@qq.com
 * @Date: 2025-01-10 15:53:51
 * @LastEditors: ilikara 3435193369@qq.com
 * @LastEditTime: 2025-04-15 06:50:01
 * @FilePath: /SynapForest/api/itemapi/itemapi.go
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
package itemapi

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"synapforest/api"
	"synapforest/database"
	"synapforest/database/itemdb"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type Item struct {
	ID         string         `json:"id" gorm:"primaryKey"` // 主键，文件的Hash
	CreatedAt  time.Time      `json:"createdAt"`            // 创建时间
	ImportedAt time.Time      `json:"importedAt"`           // 导入时间
	ModifiedAt time.Time      `json:"modifiedAt"`           // 修改时间
	DeletedAt  gorm.DeletedAt `json:"deletedAt"`            // 删除时间

	Name string `json:"name"` // 名称
	Ext  string `json:"ext"`  // 扩展名

	Width  uint32 `json:"width"`  // 宽度
	Height uint32 `json:"height"` // 高度
	Size   uint64 `json:"size"`   // 文件大小

	Url        string `json:"url"`        // 文件来源URL
	Annotation string `json:"annotation"` // 注释

	TagIds    []uuid.UUID `json:"tagIds"`    // Tags ID列表
	FolderIds []uuid.UUID `json:"folderIds"` // 文件夹ID列表

	// Palettes []uint32 `json:"palettes"` // 色票（这是什么？）
	Star uint8 `json:"star"` // 星级评分

	HaveThumbnail bool `json:"haveThumbnail"` // 是否有缩略图
	HavePreview   bool `json:"havePreview"`   // 是否有预览图
}

type ItemResponse struct {
	Status string `json:"status"`
	Data   []Item `json:"data"`
}

func AddFromUrls(c *gin.Context) {
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
		FolderIDs []uuid.UUID `json:"folderIds"` // 可选，文件夹 ID
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request data",
		})
		return
	}

	for _, item := range req.Items {
		filePath, err := saveFileFromURL(item.URL, item.Headers)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to download file: %v", err)})
			return
		}
		defer os.Remove(filePath)

		star := uint8(0)
		err = itemdb.AddItem(database.DB, filePath, item.Name, item.Website, item.Annotation, item.Tags, req.FolderIDs, &star, item.ModificationTime)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to add item: %v", err)})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// saveFileFromURL 下载文件并保存，返回文件路径
func saveFileFromURL(url string, headers map[string]string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	fileName := getFileNameFromContentDisposition(resp.Header.Get("Content-Disposition"))
	if fileName == "" {
		fileName = getFileNameFromURL(url)
	}

	if fileName == "" {
		fileName, err = itemdb.CalculateSHA256(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to calc hash: %v", err)
		}
	}

	filePath := path.Join("/tmp", fileName)
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
	parts := strings.Split(contentDisposition, "filename=")
	if len(parts) > 1 {
		return strings.Trim(parts[1], "\"")
	}
	return ""
}

func AddFromPaths(c *gin.Context) {
	var req struct {
		FileNames []string `json:"fileNames" binding:"required"`
		FolderIDs []string `json:"folderIds"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request data",
		})
		return
	}

	var err error
	var folderUUIDs []uuid.UUID = nil
	if req.FolderIDs != nil {
		folderUUIDs, err = parseUUIDs(req.FolderIDs)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Invalid FolderIDs",
			})
			return
		}
	}

	for _, filename := range req.FileNames {
		err := itemdb.AddItem(database.DB, filepath.Join(api.UploadDir, filename), nil, nil, nil, nil, folderUUIDs, nil, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "failed"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func Info(c *gin.Context) {

}

func MoveToTrash(c *gin.Context) {
	var req struct {
		ItemIDs    []string `json:"itemIds"`
		HardDelete *bool    `json:"hardDelete"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request data",
		})
		return
	}

	var err error
	if req.HardDelete != nil && *req.HardDelete {
		err = itemdb.ItemHardDelete(database.DB, req.ItemIDs)
	} else {
		err = itemdb.ItemSoftDelete(database.DB, req.ItemIDs)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed"})
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// UpdateImage 更新图片属性的 API 函数
func Update(c *gin.Context) {
	var req struct {
		ID         string     `json:"id" binding:"required"` // 图片 ID
		Name       *string    `json:"name"`                  // 图片名称
		Ext        *string    `json:"ext"`                   // 图片名称
		URL        *string    `json:"url"`                   // 图片链接
		Annotation *string    `json:"annotation"`            // 注释
		Tags       []string   `json:"tags"`                  // 标签
		Folders    []string   `json:"folders"`               // 文件夹
		Star       *uint8     `json:"star"`                  // 星级评分
		CreatedAt  *time.Time `json:"createdAt"`             // 创建时间
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request data",
		})
		return
	}

	var tagUUIDs []uuid.UUID
	if req.Tags != nil {
		for _, tagStr := range req.Tags {
			tagUUID, err := uuid.FromString(tagStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  "error",
					"message": fmt.Sprintf("Invalid tag UUID: %v", tagStr),
				})
				return
			}
			tagUUIDs = append(tagUUIDs, tagUUID)
		}
	} else {
		tagUUIDs = nil
	}

	var folderUUIDs []uuid.UUID
	if req.Folders != nil { // 只有当 Folders 不为 nil 时才进行转换
		for _, folderStr := range req.Folders {
			folderUUID, err := uuid.FromString(folderStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  "error",
					"message": fmt.Sprintf("Invalid folder UUID: %v", folderStr),
				})
				return
			}
			folderUUIDs = append(folderUUIDs, folderUUID)
		}
	} else {
		folderUUIDs = nil
	}

	err := itemdb.UpdateItem(database.DB, req.ID, req.Name, req.Ext, req.URL, req.Annotation, tagUUIDs, folderUUIDs, req.Star, req.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": fmt.Sprintf("Failed to update item: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Item updated successfully",
	})
}

// parseUUIDs 将字符串切片转换为 uuid.UUID 切片
func parseUUIDs(ids []string) ([]uuid.UUID, error) {
	var uuids []uuid.UUID
	for _, id := range ids {
		parsedUUID, err := uuid.FromString(id)
		if err != nil {
			return nil, err
		}
		uuids = append(uuids, parsedUUID)
	}
	return uuids, nil
}

func List(c *gin.Context) {
	var req struct {
		Limit     *int     `json:"limit"`
		Offset    *int     `json:"offset"`
		OrderBy   *string  `json:"orderBy"`
		Exts      []string `json:"exts"`
		Keyword   *string  `json:"keyword"`
		TagIDs    []string `json:"tagIds"`
		FolderIDs []string `json:"folderIds"`
		IsDeleted *bool    `json:"isDeleted"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request data",
		})
		return
	}

	tagUUIDs, err := parseUUIDs(req.TagIDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid TagIDs",
		})
		return
	}

	folderUUIDs, err := parseUUIDs(req.FolderIDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid FolderIDs",
		})
		return
	}

	items, err := itemdb.ItemList(database.DB, req.IsDeleted, req.OrderBy, req.Offset, req.Limit, req.Exts, req.Keyword, tagUUIDs, folderUUIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Item Query Failed",
		})
		return
	}

	resp := ItemResponse{
		Status: "success",
	}

	for _, item := range items {
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
