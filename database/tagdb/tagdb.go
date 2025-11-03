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
package tagdb

import (
	"fmt"
	"log"
	"synapforest/database/dbcommon"
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

var defaultName string = "NewTag"

func CreateTag(db *gorm.DB, name *string, description string, icon uint32, iconColor uint32, parent_id uuid.UUID, is_expand bool) (*dbcommon.Tag, error) {
	if name == nil || *name == "" {
		name = &defaultName
	}
	if newUUID, err := uuid.NewV4(); err != nil {
		log.Printf("failed to generate UUID %v", err)
		return nil, fmt.Errorf("failed to generate UUID: %v", err)
	} else {
		tag := dbcommon.Tag{
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
		if err := db.Create(&tag).Error; err != nil {
			return nil, err
		}
		return &tag, nil
	}
}

func UpdateTag(db *gorm.DB, tagID uuid.UUID, name *string, description *string, icon *uint32, iconColor *uint32, parentID *uuid.UUID) (*dbcommon.Tag, error) {
	// 查找指定 ID 的标签
	var tag dbcommon.Tag
	if err := db.First(&tag, tagID).Error; err != nil {
		return nil, err
	}

	// 更新标签的字段（仅当参数不为 nil 时更新）
	if name != nil {
		tag.Name = *name
	}
	if description != nil {
		tag.Description = *description
	}
	if icon != nil {
		tag.Icon = *icon
	}
	if iconColor != nil {
		tag.IconColor = *iconColor
	}
	if parentID != nil {
		tag.ParentID = *parentID
	}
	tag.ModifiedAt = time.Now() // 更新修改时间

	// 保存更新后的标签
	if err := db.Save(&tag).Error; err != nil {
		return nil, err
	}

	return &tag, nil
}

// 查询直接子标签的 UUID
func GetChildTagIDs(db *gorm.DB, tagID *uuid.UUID) ([]uuid.UUID, error) {
	var childIDs []uuid.UUID
	var err error

	if tagID == nil {
		// 查询所有标签的 ID
		err = db.Model(&dbcommon.Tag{}).Pluck("id", &childIDs).Error
	} else {
		// 查询直接子标签的 UUID
		err = db.Model(&dbcommon.Tag{}).
			Where("parent_id = ? AND id != ?", tagID, tagID).
			Pluck("id", &childIDs).Error
	}

	if err != nil {
		return nil, err
	}

	return childIDs, nil
}

// UpdateTagParents 批量更新标签的 ParentID
func UpdateTagParents(db *gorm.DB, tagIDs []uuid.UUID, newParentID uuid.UUID) error {
	// 检查是否为空
	if len(tagIDs) == 0 {
		return nil // 没有要更新的标签
	}

	// 批量更新
	result := db.Model(&dbcommon.Tag{}).
		Where("id IN ?", tagIDs).
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

// 递归获取所有子标签的ID
func getChildTagIDs(db *gorm.DB, tagID uuid.UUID) ([]uuid.UUID, error) {
	var tagIDs []uuid.UUID
	var childTags []dbcommon.Tag

	if err := db.Where("parent_id = ?", tagID).Find(&childTags).Error; err != nil {
		return nil, err
	}

	tagIDs = append(tagIDs, tagID)

	for _, child := range childTags {
		childIDs, err := getChildTagIDs(db, child.ID)
		if err != nil {
			return nil, err
		}
		tagIDs = append(tagIDs, childIDs...)
	}

	return tagIDs, nil
}

// 删除标签及其子标签
func DeleteTag(db *gorm.DB, tagID uuid.UUID, hardDelete *bool, deleteAssociatedFiles *bool) error {
	tagIDs, err := getChildTagIDs(db, tagID)
	if err != nil {
		return err
	}

	if deleteAssociatedFiles != nil && *deleteAssociatedFiles {
		// 获取所有关联的 item_id
		var itemIDs []string
		if err := db.Table("item_tags").
			Select("item_id").
			Where("tag_id IN ?", tagIDs).
			Pluck("item_id", &itemIDs).Error; err != nil {
			return err
		}

		// 软删除关联的文件
		if err := db.Delete(&dbcommon.Item{}, itemIDs).Error; err != nil {
			return err
		}
	} else {
		// 删除标签和文件的关联关系
		if err := db.Exec("DELETE FROM item_tags WHERE tag_id IN ?", tagIDs).Error; err != nil {
			return err
		}
	}

	if true { // 由于不知道软删除标签有什么意义，暂时忽略该项
		if err := db.Unscoped().Delete(&dbcommon.Tag{}, "id IN ?", tagIDs).Error; err != nil {
			return err
		}
	} else {
		if err := db.Delete(&dbcommon.Tag{}, "id IN ?", tagIDs).Error; err != nil {
			return err
		}
	}

	return nil
}
