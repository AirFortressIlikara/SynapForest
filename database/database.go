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
package database

import (
	"log"
	"os"
	"path/filepath"
	"synapforest/database/dbcommon"
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var Token string = "TEST123123"
var DB *gorm.DB
var VectorDB *gorm.DB
var DbBaseDir string

func CreateRootFolder(db *gorm.DB) error {
	rootFolder := dbcommon.Folder{
		ID:         uuid.Nil, // Root 文件夹 ID 固定为 uuid.Nil
		Name:       "Root",
		IsExpand:   true,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}

	// 使用 FirstOrCreate 避免重复创建
	result := db.Where(dbcommon.Folder{ID: uuid.Nil}).FirstOrCreate(&rootFolder)
	return result.Error
}

func Database_init(library_dir string) (*gorm.DB, error) {
	var err error

	DbBaseDir = library_dir

	if err := os.MkdirAll(DbBaseDir, os.ModePerm); err != nil {
		log.Fatalf("failed to create data directory: %v", err)
	}

	os.MkdirAll(filepath.Join(DbBaseDir, "raw_files"), os.ModePerm)
	os.MkdirAll(filepath.Join(DbBaseDir, "thumbnails"), os.ModePerm)
	os.MkdirAll(filepath.Join(DbBaseDir, "previews"), os.ModePerm)

	DB, err = gorm.Open(sqlite.Open(filepath.Join(DbBaseDir, "files.db")), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	VectorDB, err = gorm.Open(sqlite.Open(filepath.Join(DbBaseDir, "vectors.db")), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// 启用SQLite的WAL
	DB.Exec("PRAGMA journal_mode=WAL;")

	DB = DB.Debug()

	VectorDB.Exec("PRAGMA journal_mode=WAL;")

	VectorDB = VectorDB.Debug()

	if err := DB.AutoMigrate(&dbcommon.Item{}, &dbcommon.Folder{}, &dbcommon.Tag{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	if err := VectorDB.AutoMigrate(&dbcommon.ItemVector{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	CreateRootFolder(DB)
	return DB, err
}

// GetItemIDsByFolder 查询指定文件夹下所有图片的 ID
func GetItemIDsByFolder(db *gorm.DB, folderID uuid.UUID) ([]string, error) {
	var itemIDs []string

	err := db.Table("item_folders").
		Where("folder_id = ?", folderID).
		Pluck("item_id", &itemIDs).Error

	if err != nil {
		return nil, err
	}

	return itemIDs, nil
}

// GetItemIDsByTag 查询指定标签下所有图片的 ID
func GetItemIDsByTag(db *gorm.DB, tagID uuid.UUID) ([]string, error) {
	var itemIDs []string

	err := db.Table("item_tags").
		Where("folder_id = ?", tagID).
		Pluck("item_id", &itemIDs).Error

	if err != nil {
		return nil, err
	}

	return itemIDs, nil
}

// GetFoldersByItemID 查询指定图片 ID 所属的所有文件夹
func GetFoldersByItemID(db *gorm.DB, itemID string) ([]uuid.UUID, error) {
	var folderIDs []uuid.UUID

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
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("item_folders").
			Where("folder_id = ? AND item_id IN (?)", folderID, itemIDs).
			Delete(nil).Error; err != nil {
			return err
		}

		return nil
	})
}

// AddFolderForItems 批量添加指定图片与多个文件夹的关联
func AddFolderForItems(db *gorm.DB, itemIDs []string, folderIDs []uuid.UUID) error {
	if len(itemIDs) == 0 || len(folderIDs) == 0 {
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		var records []map[string]interface{}
		for _, itemID := range itemIDs {
			for _, folderID := range folderIDs {
				// 检查记录是否已经存在
				var count int64
				if err := tx.Table("item_folders").
					Where("item_id = ? AND folder_id = ?", itemID, folderID).
					Count(&count).Error; err != nil {
					return err
				}

				// 如果记录不存在，则插入
				if count == 0 {
					records = append(records, map[string]interface{}{
						"item_id":   itemID,
						"folder_id": folderID,
					})
				}
			}
		}
		if len(records) > 0 {
			if err := tx.Table("item_folders").Create(records).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
