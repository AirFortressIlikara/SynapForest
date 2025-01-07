/*
 * @Author: ilikara 3435193369@qq.com
 * @Date: 2024-12-29 12:43:00
 * @LastEditors: ilikara 3435193369@qq.com
 * @LastEditTime: 2025-01-07 14:35:46
 * @FilePath: /my_eagle/api/folderapi/folder.go
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
package folderapi

import (
	"fmt"
	"my_eagle/database"
	"my_eagle/database/dbcommon"
	"my_eagle/database/folderdb"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

type Folder struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Items       []string  `json:"items"`

	Parent     uuid.UUID   `json:"parent"`
	SubFolders []uuid.UUID `json:"sub_folders"`

	ModifiedAt time.Time `json:"modified_at"`
	Tags       []string  `json:"tags"`
	IsExpand   bool      `json:"isExpand"`
}

// 返回的结构体
type FolderResponse struct {
	Status string   `json:"status"`
	Data   []Folder `json:"data"`
}

func CreateFolder(c *gin.Context) {
	var req struct {
		FolderName *string   `json:"folderName"`
		Parent     uuid.UUID `json:"parent"`

		Token string `json:"token" binding:"required"`
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

	folder, err := folderdb.CreateFolder(database.DB, req.FolderName, "", 0, 0, req.Parent, false)
	if err != nil {
		// 返回JSON响应
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Folder Create Failed",
		})
		return
	}

	// 构建响应数据
	resp := FolderResponse{
		Status: "success",
	}
	subfolders, _ := folderdb.GetChildFolderIDs(database.DB, folder.ID)
	items, _ := database.GetItemIDsByFolder(database.DB, folder.ID)
	data := Folder{
		ID:          folder.ID,
		Name:        folder.Name,
		Description: folder.Description,
		Items:       items,
		Parent:      folder.ParentID,
		SubFolders:  subfolders,
		ModifiedAt:  folder.ModifiedAt,
		Tags:        nil,
		IsExpand:    false,
	}
	resp.Data = append(resp.Data, data)

	// 返回JSON响应
	c.JSON(http.StatusOK, resp)
}

func ListFolder(c *gin.Context) {
	var req struct {
		Parent *string `json:"parent"`

		Token string `json:"token" binding:"required"`
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

	// 构建响应数据
	resp := FolderResponse{
		Status: "success",
	}

	var parent uuid.UUID
	if req.Parent != nil {
		parent, _ = uuid.FromString(*req.Parent)
	}

	folderIDs, err := folderdb.GetChildFolderIDs(database.DB, parent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed"})
		return
	}

	for _, folderid := range folderIDs {
		var folder dbcommon.Folder
		if err := database.DB.First(&folder, folderid).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "failed"})
			return
		} else {
			fmt.Printf("Found Folder: %v\n", folder)
		}
		subfolders, _ := folderdb.GetChildFolderIDs(database.DB, folder.ID)
		items, _ := database.GetItemIDsByFolder(database.DB, folder.ID)
		data := Folder{
			ID:          folder.ID,
			Name:        folder.Name,
			Description: folder.Description,
			Items:       items,
			Parent:      folder.ParentID,
			SubFolders:  subfolders,
			ModifiedAt:  folder.ModifiedAt,
			Tags:        nil,
			IsExpand:    false,
		}
		resp.Data = append(resp.Data, data)
	}

	// 返回JSON响应
	c.JSON(http.StatusOK, resp)
}

func UpdateFolder(c *gin.Context) {

}
