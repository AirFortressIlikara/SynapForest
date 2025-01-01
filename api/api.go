/*
 * @Author: ilikara 3435193369@qq.com
 * @Date: 2025-01-01 15:52:53
 * @LastEditors: ilikara 3435193369@qq.com
 * @LastEditTime: 2025-01-01 15:59:37
 * @FilePath: /my_eagle/api/api.go
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
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

	files := form.File["files"] // 获取表单中的多个文件
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
	// 从 URL 路由获取 id
	id := c.Param("id")

	var Item dbcommon.Item
	err := database.DB.First(&Item, "id = ?", id).Error
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
		})
	}
	if Item.HaveThumbnail {
		// 构建图片路径
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
	// 从 URL 路由获取 id
	id := c.Param("id")

	var Item dbcommon.Item
	err := database.DB.First(&Item, "id = ?", id).Error
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
		})
	}
	if Item.HavePreview {
		// 构建图片路径
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
	err := database.DB.First(&Item, "id = ?", id).Error
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
