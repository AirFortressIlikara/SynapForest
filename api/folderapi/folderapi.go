/*
 * @Author: ilikara 3435193369@qq.com
 * @Date: 2024-12-29 12:43:00
 * @LastEditors: ilikara 3435193369@qq.com
 * @LastEditTime: 2025-01-07 16:23:52
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
	var req struct {
		FolderID    string  `json:"folderId" binding:"required"` // 使用 string 类型接收 UUID
		FolderName  *string `json:"folderName"`
		Description *string `json:"description"`
		Icon        *uint32 `json:"icon"`
		IconColor   *uint32 `json:"iconColor"`
		Parent      *string `json:"parent"` // 使用 string 类型接收 UUID

		Token string `json:"token" binding:"required"`
	}

	// 绑定 JSON 数据到结构体
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request data",
		})
		return
	}

	// 验证 Token
	if req.Token != database.Token {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Invalid token",
		})
		return
	}

	// 将 FolderID 从 string 转换为 uuid.UUID
	folderID, err := uuid.FromString(req.FolderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid folder ID",
		})
		return
	}

	// 将 Parent 从 string 转换为 uuid.UUID（如果 Parent 不为 nil）
	var parentID *uuid.UUID
	if req.Parent != nil {
		parent, err := uuid.FromString(*req.Parent)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Invalid parent folder ID",
			})
			return
		}
		parentID = &parent
	}

	// 调用 UpdateFolder 函数更新文件夹
	folder, err := folderdb.UpdateFolder(database.DB, folderID, req.FolderName, req.Description, req.Icon, req.IconColor, parentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to update folder",
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

	// 返回 JSON 响应
	c.JSON(http.StatusOK, resp)
}
