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
package folderapi

import (
	"errors"
	"fmt"
	"net/http"
	"synapforest/database"
	"synapforest/database/dbcommon"
	"synapforest/database/folderdb"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type Folder struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Items       []string  `json:"items"`

	Parent     uuid.UUID   `json:"parent"`
	SubFolders []uuid.UUID `json:"subFolders"`

	ModifiedAt time.Time `json:"modifiedAt"`
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
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request data",
		})
		return
	}

	folder, err := folderdb.CreateFolder(database.DB, req.FolderName, "", 0, 0, req.Parent, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Folder Create Failed",
		})
		return
	}

	resp := FolderResponse{
		Status: "success",
	}
	subfolders, _ := folderdb.GetChildFolderIDs(database.DB, &folder.ID)
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

	c.JSON(http.StatusOK, resp)
}

func ListFolder(c *gin.Context) {
	var req struct {
		Parent *string `json:"parent"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request data",
		})
		return
	}

	resp := FolderResponse{
		Status: "success",
	}

	var parent *uuid.UUID = nil
	if req.Parent != nil {
		parsedUUID, _ := uuid.FromString(*req.Parent)
		parent = &parsedUUID
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
		subfolders, _ := folderdb.GetChildFolderIDs(database.DB, &folder.ID)
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
			IsExpand:    folder.IsExpand,
		}
		resp.Data = append(resp.Data, data)
	}

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
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request data",
		})
		return
	}

	folderID, err := uuid.FromString(req.FolderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid folder ID",
		})
		return
	}

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

	folder, err := folderdb.UpdateFolder(database.DB, folderID, req.FolderName, req.Description, req.Icon, req.IconColor, parentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to update folder",
		})
		return
	}

	resp := FolderResponse{
		Status: "success",
	}
	subfolders, _ := folderdb.GetChildFolderIDs(database.DB, &folder.ID)
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

	c.JSON(http.StatusOK, resp)
}

func DeleteFolder(c *gin.Context) {
	var req struct {
		FolderID    string `json:"folderId" binding:"required"` // 使用 string 类型接收 UUID
		HardDelete  *bool  `json:"hardDelete"`                  // 由于不知道软删除文件夹有什么意义，暂时忽略该项
		DeleteItems *bool  `json:"deleteItems"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request data",
		})
		return
	}

	folderID, err := uuid.FromString(req.FolderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid folder ID",
		})
		return
	}

	if err := folderdb.DeleteFolder(database.DB, folderID, req.HardDelete, req.DeleteItems); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Folder delete failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Folders deleted successfully",
	})
}

func UpdateFoldersParent(c *gin.Context) {
	var req struct {
		FolderIDs []string `json:"folderIds" binding:"required"` // 要更新的文件夹ID数组
		NewParent *string  `json:"newParent"`                    // 新的父文件夹ID
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request data",
		})
		return
	}

	// 将 FolderIDs 转换为 uuid.UUID 数组
	var folderIDs []uuid.UUID
	for _, folderIDStr := range req.FolderIDs {
		folderID, err := uuid.FromString(folderIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": fmt.Sprintf("Invalid folder ID: %s", folderIDStr),
			})
			return
		}
		folderIDs = append(folderIDs, folderID)
	}

	// 处理新的父文件夹ID
	var newParentID uuid.UUID
	if req.NewParent != nil {
		parent, err := uuid.FromString(*req.NewParent)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Invalid new parent folder ID",
			})
			return
		}
		newParentID = parent
	} else {
		newParentID = uuid.Nil
	}

	// 调用 folderdb.UpdateFolderParents 进行批量更新
	err := folderdb.UpdateFolderParents(database.DB, folderIDs, newParentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": "No matching folders found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to update folders",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Folders updated successfully",
	})
}
