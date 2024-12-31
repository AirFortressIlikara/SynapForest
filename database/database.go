/*
 * @Author: ilikara 3435193369@qq.com
 * @Date: 2024-12-29 12:43:00
 * @LastEditors: ilikara 3435193369@qq.com
 * @LastEditTime: 2024-12-31 09:15:29
 * @FilePath: /my_eagle/database/database.go
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
package database

import (
	"log"
	"my_eagle/database/common"
	"os"
	"path/filepath"

	"github.com/gofrs/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var Token string = "TEST123123"
var DB *gorm.DB
var DbBaseDir string

func Database_init(library_dir string) (*gorm.DB, error) {
	var err error

	DbBaseDir = library_dir

	// 创建数据库存储路径
	if err := os.MkdirAll(DbBaseDir, os.ModePerm); err != nil {
		log.Fatalf("failed to create data directory: %v", err)
	}

	// 创建文件存储和缩略图存储目录
	os.MkdirAll(filepath.Join(DbBaseDir, "raw_files"), os.ModePerm)
	os.MkdirAll(filepath.Join(DbBaseDir, "thumbnails"), os.ModePerm)

	// 打开 SQLite 数据库
	db, err := gorm.Open(sqlite.Open(filepath.Join(DbBaseDir, "files.db")), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// 自动迁移数据库
	if err := db.AutoMigrate(&common.Item{}, &common.Folder{}, &common.Tag{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	return db, err
}

// 查询直接子文件夹的 UUID
func GetChildFolderIDs(db *gorm.DB, folderID uuid.UUID) ([]uuid.UUID, error) {
	var childIDs []uuid.UUID

	// 查询直接子文件夹的 UUID
	err := db.Model(&common.Folder{}).Where("parent_id = ?", folderID).Pluck("id", &childIDs).Error
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
	result := db.Model(&common.Folder{}).
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

// GetItemIDsByFolder 查询指定文件夹下所有图片的 ID
func GetItemIDsByFolder(db *gorm.DB, folderID uuid.UUID) ([]string, error) {
	var itemIDs []string

	// 查询中间表
	err := db.Table("item_folders").
		Where("folder_id = ?", folderID).
		Pluck("item_id", &itemIDs).Error

	if err != nil {
		return nil, err
	}

	return itemIDs, nil
}

// GetFoldersByItemID 查询指定图片 ID 所属的所有文件夹
func GetFoldersByItemID(db *gorm.DB, itemID string) ([]uuid.UUID, error) {
	var folderIDs []uuid.UUID

	// 查询中间表
	err := db.Table("item_folders").
		Where("item_id = ?", itemID).
		Pluck("folder_id", &folderIDs).Error

	if err != nil {
		return nil, err
	}

	return folderIDs, nil
}

// UpdateFoldersForItem 批量更新指定图片的所属文件夹 ID 列表
func UpdateFoldersForItem(db *gorm.DB, itemID string, newFolderIDs []uuid.UUID) error {
	if len(newFolderIDs) == 0 {
		return nil // 如果没有新的文件夹 ID，直接返回
	}

	// 开启事务
	return db.Transaction(func(tx *gorm.DB) error {
		// 删除图片当前的所有文件夹关联
		if err := tx.Table("item_folders").
			Where("item_id = ?", itemID).
			Delete(nil).Error; err != nil {
			return err
		}

		// 创建新的文件夹关联
		var records []map[string]interface{}
		for _, folderID := range newFolderIDs {
			records = append(records, map[string]interface{}{
				"item_id":   itemID,
				"folder_id": folderID,
			})
		}
		if len(records) > 0 {
			if err := tx.Table("item_folders").Create(records).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
