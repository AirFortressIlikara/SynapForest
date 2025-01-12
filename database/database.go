/*
 * @Author: ilikara 3435193369@qq.com
 * @Date: 2024-12-29 12:43:00
 * @LastEditors: ilikara 3435193369@qq.com
 * @LastEditTime: 2025-01-07 13:39:14
 * @FilePath: /my_eagle/database/database.go
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
package database

import (
	"log"
	"my_eagle/database/dbcommon"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var Token string = "TEST123123"
var DB *gorm.DB
var DbBaseDir string

func CreateRootFolder(db *gorm.DB) error {
	rootFolder := dbcommon.Folder{
		ID:         uuid.Nil,
		Name:       "Root",
		IsExpand:   true,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}
	return db.Create(&rootFolder).Error
}

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
	os.MkdirAll(filepath.Join(DbBaseDir, "previews"), os.ModePerm)

	// 打开 SQLite 数据库
	DB, err = gorm.Open(sqlite.Open(filepath.Join(DbBaseDir, "files.db")), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	DB = DB.Debug()
	// 自动迁移数据库
	if err := DB.AutoMigrate(&dbcommon.Item{}, &dbcommon.Folder{}, &dbcommon.Tag{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	CreateRootFolder(DB)
	return DB, err
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

// RemoveFoldersForItems 批量删除指定图片与某个文件夹的关联
func RemoveFoldersForItems(db *gorm.DB, itemIDs []string, folderID uuid.UUID) error {
	if len(itemIDs) == 0 {
		return nil // 如果没有图片 ID，直接返回
	}

	// 开启事务
	return db.Transaction(func(tx *gorm.DB) error {
		// 删除指定图片与某个文件夹的关联
		if err := tx.Table("item_folders").
			Where("folder_id = ? AND item_id IN (?)", folderID, itemIDs).
			Delete(nil).Error; err != nil {
			return err
		}

		return nil
	})
}

// AddFolderForItems 批量添加指定图片与某个文件夹的关联
func AddFolderForItems(db *gorm.DB, itemIDs []string, folderID uuid.UUID) error {
	if len(itemIDs) == 0 {
		return nil // 如果没有图片 ID，直接返回
	}

	// 开启事务
	return db.Transaction(func(tx *gorm.DB) error {
		// 创建新的文件夹关联
		var records []map[string]interface{}
		for _, itemID := range itemIDs {
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
