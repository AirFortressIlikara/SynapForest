/*
 * @Author: Ilikara 3435193369@qq.com
 * @Date: 2025-01-09 19:59:53
 * @LastEditors: ilikara 3435193369@qq.com
 * @LastEditTime: 2025-06-06 08:13:40
 * @FilePath: /SynapForest/api/tagapi/tagapi.go
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
package tagapi

import (
	"errors"
	"fmt"
	"net/http"
	"synapforest/database"
	"synapforest/database/dbcommon"
	"synapforest/database/tagdb"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type Tag struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Items       []string  `json:"items"`

	Parent  uuid.UUID   `json:"parent"`
	SubTags []uuid.UUID `json:"subTags"`

	ModifiedAt time.Time `json:"modifiedAt"`
	Tags       []string  `json:"tags"`
	IsExpand   bool      `json:"isExpand"`
}

// 返回的结构体
type TagResponse struct {
	Status string `json:"status"`
	Data   []Tag  `json:"data"`
}

func CreateTag(c *gin.Context) {
	var req struct {
		TagName *string   `json:"tagName"`
		Parent  uuid.UUID `json:"parent"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request data",
		})
		return
	}

	tag, err := tagdb.CreateTag(database.DB, req.TagName, "", 0, 0, req.Parent, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Tag Create Failed",
		})
		return
	}

	resp := TagResponse{
		Status: "success",
	}
	subtags, _ := tagdb.GetChildTagIDs(database.DB, &tag.ID)
	items, _ := database.GetItemIDsByTag(database.DB, tag.ID)
	data := Tag{
		ID:          tag.ID,
		Name:        tag.Name,
		Description: tag.Description,
		Items:       items,
		Parent:      tag.ParentID,
		SubTags:     subtags,
		ModifiedAt:  tag.ModifiedAt,
		Tags:        nil,
		IsExpand:    false,
	}
	resp.Data = append(resp.Data, data)

	c.JSON(http.StatusOK, resp)
}

func ListTag(c *gin.Context) {
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

	resp := TagResponse{
		Status: "success",
	}

	var parent *uuid.UUID = nil
	if req.Parent != nil {
		parsedUUID, _ := uuid.FromString(*req.Parent)
		parent = &parsedUUID
	}

	tagIDs, err := tagdb.GetChildTagIDs(database.DB, parent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed"})
		return
	}

	for _, tagid := range tagIDs {
		var tag dbcommon.Tag
		if err := database.DB.First(&tag, tagid).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "failed"})
			return
		} else {
			fmt.Printf("Found Tag: %v\n", tag)
		}
		subtags, _ := tagdb.GetChildTagIDs(database.DB, &tag.ID)
		items, _ := database.GetItemIDsByTag(database.DB, tag.ID)
		data := Tag{
			ID:          tag.ID,
			Name:        tag.Name,
			Description: tag.Description,
			Items:       items,
			Parent:      tag.ParentID,
			SubTags:     subtags,
			ModifiedAt:  tag.ModifiedAt,
			Tags:        nil,
			IsExpand:    tag.IsExpand,
		}
		resp.Data = append(resp.Data, data)
	}

	c.JSON(http.StatusOK, resp)
}

func UpdateTag(c *gin.Context) {
	var req struct {
		TagID       string  `json:"tagId" binding:"required"` // 使用 string 类型接收 UUID
		TagName     *string `json:"tagName"`
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

	tagID, err := uuid.FromString(req.TagID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid tag ID",
		})
		return
	}

	var parentID *uuid.UUID
	if req.Parent != nil {
		parent, err := uuid.FromString(*req.Parent)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Invalid parent tag ID",
			})
			return
		}
		parentID = &parent
	}

	tag, err := tagdb.UpdateTag(database.DB, tagID, req.TagName, req.Description, req.Icon, req.IconColor, parentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to update tag",
		})
		return
	}

	resp := TagResponse{
		Status: "success",
	}
	subtags, _ := tagdb.GetChildTagIDs(database.DB, &tag.ID)
	items, _ := database.GetItemIDsByTag(database.DB, tag.ID)
	data := Tag{
		ID:          tag.ID,
		Name:        tag.Name,
		Description: tag.Description,
		Items:       items,
		Parent:      tag.ParentID,
		SubTags:     subtags,
		ModifiedAt:  tag.ModifiedAt,
		Tags:        nil,
		IsExpand:    false,
	}
	resp.Data = append(resp.Data, data)

	c.JSON(http.StatusOK, resp)
}

func DeleteTag(c *gin.Context) {
	var req struct {
		TagID       string `json:"tagId" binding:"required"` // 使用 string 类型接收 UUID
		HardDelete  *bool  `json:"hardDelete"`               // 由于不知道软删除标签有什么意义，暂时忽略该项
		DeleteItems *bool  `json:"deleteItems"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request data",
		})
		return
	}

	tagID, err := uuid.FromString(req.TagID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid tag ID",
		})
		return
	}

	if err := tagdb.DeleteTag(database.DB, tagID, req.HardDelete, req.DeleteItems); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Tag delete failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Tags deleted successfully",
	})
}

func UpdateTagsParent(c *gin.Context) {
	var req struct {
		TagIDs    []string `json:"tagIds" binding:"required"` // 要更新的标签ID数组
		NewParent *string  `json:"newParent"`                 // 新的父标签ID
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request data",
		})
		return
	}

	// 将 TagIDs 转换为 uuid.UUID 数组
	var tagIDs []uuid.UUID
	for _, tagIDStr := range req.TagIDs {
		tagID, err := uuid.FromString(tagIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": fmt.Sprintf("Invalid tag ID: %s", tagIDStr),
			})
			return
		}
		tagIDs = append(tagIDs, tagID)
	}

	// 处理新的父标签ID
	var newParentID uuid.UUID
	if req.NewParent != nil {
		parent, err := uuid.FromString(*req.NewParent)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Invalid new parent tag ID",
			})
			return
		}
		newParentID = parent
	} else {
		newParentID = uuid.Nil
	}

	// 调用 tagdb.UpdateTagParents 进行批量更新
	err := tagdb.UpdateTagParents(database.DB, tagIDs, newParentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": "No matching tags found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to update tags",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Tags updated successfully",
	})
}
