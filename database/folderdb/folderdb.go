/*
 * @Author: ilikara 3435193369@qq.com
 * @Date: 2024-12-31 08:55:34
 * @LastEditors: ilikara 3435193369@qq.com
 * @LastEditTime: 2025-01-07 15:52:32
 * @FilePath: /my_eagle/database/folderdb/folderdb.go
 * @Description:
 *
 * Copyright (c) 2024 by ${git_name_email}, All Rights Reserved.
 */
package folderdb

import (
	"fmt"
	"log"
	"my_eagle/database/dbcommon"
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

var defaultName string = "NewFolder"

func CreateFolder(db *gorm.DB, name *string, description string, icon uint32, iconColor uint32, parent_id uuid.UUID, is_expand bool) (*dbcommon.Folder, error) {
	if name == nil || *name == "" {
		name = &defaultName
	}
	if newUUID, err := uuid.NewV4(); err != nil {
		log.Printf("failed to generate UUID %v", err)
		return nil, fmt.Errorf("failed to generate UUID: %v", err)
	} else {
		folder := dbcommon.Folder{
			ID:          newUUID,
			Name:        *name,
			Description: description,
			Icon:        icon,
			IconColor:   iconColor,
			ParentID:    parent_id,
			CreatedAt:   time.Now(),
			ModifiedAt:  time.Now(),
			IsExpand:    is_expand,
		}
		if err := db.Create(&folder).Error; err != nil {
			return nil, err
		}
		return &folder, nil
	}
}

func UpdateFolder(db *gorm.DB, folderID uuid.UUID, name *string, description *string, icon *uint32, iconColor *uint32, parentID *uuid.UUID) (*dbcommon.Folder, error) {
	// 查找指定 ID 的文件夹
	var folder dbcommon.Folder
	if err := db.First(&folder, folderID).Error; err != nil {
		return nil, err
	}

	// 更新文件夹的字段（仅当参数不为 nil 时更新）
	if name != nil {
		folder.Name = *name
	}
	if description != nil {
		folder.Description = *description
	}
	if icon != nil {
		folder.Icon = *icon
	}
	if iconColor != nil {
		folder.IconColor = *iconColor
	}
	if parentID != nil {
		folder.ParentID = *parentID
	}
	folder.ModifiedAt = time.Now() // 更新修改时间

	// 保存更新后的文件夹
	if err := db.Save(&folder).Error; err != nil {
		return nil, err
	}

	return &folder, nil
}

// 查询直接子文件夹的 UUID
func GetChildFolderIDs(db *gorm.DB, folderID uuid.UUID) ([]uuid.UUID, error) {
	var childIDs []uuid.UUID

	// 查询直接子文件夹的 UUID
	err := db.Model(&dbcommon.Folder{}).Where("parent_id = ?", folderID).Pluck("id", &childIDs).Error
	if err != nil {
		return nil, err
	}

	return childIDs, nil
}

// UpdateFolderParents 批量更新文件夹的 ParentID
func UpdateFolderParents(db *gorm.DB, folderIDs []uuid.UUID, newParentID uuid.UUID) error {
	// 检查是否为空
	if len(folderIDs) == 0 {
		return nil // 没有要更新的文件夹
	}

	// 批量更新
	result := db.Model(&dbcommon.Folder{}).
		Where("id IN ?", folderIDs).
		Update("parent_id", newParentID)

	// 检查更新结果
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound // 没有匹配的记录
	}

	return nil
}
