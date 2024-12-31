/*
 * @Author: ilikara 3435193369@qq.com
 * @Date: 2024-12-31 08:55:34
 * @LastEditors: ilikara 3435193369@qq.com
 * @LastEditTime: 2024-12-31 09:12:45
 * @FilePath: /my_eagle/database/folderdb/folder.go
 * @Description:
 *
 * Copyright (c) 2024 by ${git_name_email}, All Rights Reserved.
 */
package folderdb

import (
	"fmt"
	"log"
	"my_eagle/database/common"
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

var defaultName string = "NewFolder"

func CreateFolder(db *gorm.DB, name *string, description string, icon uint32, iconColor uint32, parent_id uuid.UUID, is_expand bool) (*common.Folder, error) {
	if name == nil || *name == "" {
		name = &defaultName
	}
	if newUUID, err := uuid.NewV4(); err != nil {
		log.Printf("failed to generate UUID %v", err)
		return nil, fmt.Errorf("failed to generate UUID: %v", err)
	} else {
		folder := common.Folder{
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

func UpdateFolder(db *gorm.DB, folderID uuid.UUID, name string, description string, icon uint32, iconColor uint32, parent_id uuid.UUID) (*common.Folder, error) {
	// 查找指定 ID 的文件夹
	var folder common.Folder
	if err := db.First(&folder, folderID).Error; err != nil {
		return nil, err
	}

	// 更新文件夹的字段
	folder.Name = name
	folder.Description = description
	folder.Icon = icon
	folder.IconColor = iconColor
	folder.ParentID = parent_id
	folder.ModifiedAt = time.Now() // 更新修改时间

	// 保存更新后的文件夹
	if err := db.Save(&folder).Error; err != nil {
		return nil, err
	}

	return &folder, nil
}
