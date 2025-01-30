/*
 * @Author: Ilikara 3435193369@qq.com
 * @Date: 2025-01-09 19:59:53
 * @LastEditors: Ilikara 3435193369@qq.com
 * @LastEditTime: 2025-01-30 19:22:49
 * @FilePath: /my_eagle/api/folderapi/folderapi.go
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
