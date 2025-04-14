/*
 * @Author: Ilikara 3435193369@qq.com
 * @Date: 2025-01-09 19:59:53
 * @LastEditors: ilikara 3435193369@qq.com
 * @LastEditTime: 2025-02-21 07:46:27
 * @FilePath: /my_eagle/api/api.go
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
package api

import (
	"fmt"
	"my_eagle/database"
	"my_eagle/database/dbcommon"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

var UploadDir = "uploads"

func ApiInit(uploadDir string) error {
	UploadDir = filepath.Join(database.DbBaseDir, uploadDir)
	if _, err := os.Stat(UploadDir); os.IsNotExist(err) {
		if err := os.Mkdir(UploadDir, os.ModePerm); err != nil {
			return fmt.Errorf("Failed to create upload directory: %v", err)
		}
	}
	return nil
}

func Uploadfiles(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get files"})
		return
	}

	files := form.File["files"]
	var savedFiles []string

	for _, file := range files {
		dst := filepath.Join(UploadDir, filepath.Base(file.Filename))
		if err := c.SaveUploadedFile(file, dst); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save file %s", file.Filename)})
			return
		}
		savedFiles = append(savedFiles, file.Filename)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Files uploaded successfully", "files": savedFiles})
}

// ServeImage 根据 id 返回图片
func ServeThumbnails(c *gin.Context) {
	id := c.Param("id")

	var Item dbcommon.Item
	err := database.DB.Unscoped().First(&Item, "id = ?", id).Error
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
		})
	}
	if Item.HaveThumbnail {
		imagePath := filepath.Join(database.DbBaseDir, "thumbnails", id+".webp")
		c.File(imagePath)
	} else {
		// 可能修改为返回通用占位符？
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
		})
	}
}

func ServePreviews(c *gin.Context) {
	id := c.Param("id")

	var Item dbcommon.Item
	err := database.DB.Unscoped().First(&Item, "id = ?", id).Error
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
		})
	}
	if Item.HavePreview {
		imagePath := filepath.Join(database.DbBaseDir, "previews", id+".webp")
		c.File(imagePath)
	} else {
		// 可能修改为返回通用占位符？
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
		})
	}
}

func ServeRawFile(c *gin.Context) {
	id := c.Param("id")

	var Item dbcommon.Item
	err := database.DB.Unscoped().First(&Item, "id = ?", id).Error
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
		})
	}
	// 构建图片路径
	imagePath := filepath.Join(database.DbBaseDir, "raw_files", id, Item.Name+"."+Item.Ext)

	// 检查文件是否存在（可以进一步完善此部分）
	c.File(imagePath)
}

func RemoveFolderForItems(c *gin.Context) {
	var req struct {
		ItemIDs  []string `json:"itemIds" binding:"required"`  // 图片 ID 列表
		FolderID string   `json:"folderId" binding:"required"` // 文件夹 ID
	}

	// 绑定 JSON 数据到结构体
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request data",
		})
		return
	}

	// 解析文件夹 ID
	folderID, err := uuid.FromString(req.FolderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid folder ID",
		})
		return
	}

	// 调用批量删除函数
	if err := database.RemoveFoldersForItems(database.DB, req.ItemIDs, folderID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to remove folder associations",
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Folder associations removed successfully",
	})
}

func AddFolderForItems(c *gin.Context) {
	var req struct {
		ItemIDs   []string `json:"itemIds" binding:"required"`   // 图片 ID 列表
		FolderIDs []string `json:"folderIds" binding:"required"` // 文件夹 ID
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request data",
		})
		return
	}

	var folderIDs []uuid.UUID
	for _, folderStr := range req.FolderIDs {
		folderUUID, err := uuid.FromString(folderStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": fmt.Sprintf("Invalid tag UUID: %v", folderStr),
			})
			return
		}
		folderIDs = append(folderIDs, folderUUID)
	}

	if err := database.AddFolderForItems(database.DB, req.ItemIDs, folderIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to add folder associations",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Folder associations added successfully",
	})
}
