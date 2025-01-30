/*
 * @Author: Ilikara 3435193369@qq.com
 * @Date: 2025-01-10 15:53:51
 * @LastEditors: Ilikara 3435193369@qq.com
 * @LastEditTime: 2025-01-30 19:21:14
 * @FilePath: /my_eagle/database/itemdb/itemdb.go
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
package itemdb

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"time"

	"my_eagle/database"
	"my_eagle/database/dbcommon"

	"github.com/chai2010/webp"
	"github.com/gofrs/uuid"
	"github.com/nfnt/resize"
	"gorm.io/gorm"
)

// 查找符合条件的 items
func ItemList(db *gorm.DB, isDeleted *bool, orderBy *string, page *int, pageSize *int, exts []string, keyword *string, tags []uuid.UUID, folders []uuid.UUID) ([]dbcommon.Item, error) {
	var items []dbcommon.Item

	query := db.Model(&dbcommon.Item{})

	if isDeleted != nil && *isDeleted {
		query = query.Unscoped().Where("deleted_at IS NOT NULL")
	}

	if len(exts) > 0 {
		query = query.Where("ext IN ?", exts)
	}

	if keyword != nil && *keyword != "" {
		query = query.Where("name LIKE ?", "%"+*keyword+"%")
	}

	if len(tags) > 0 {
		query = query.Joins("JOIN item_tags ON item_tags.item_id = items.id").
			Where("item_tags.tag_id IN ?", tags)
	}

	if len(folders) > 0 {
		query = query.Joins("JOIN item_folders ON item_folders.item_id = items.id").
			Where("item_folders.folder_id IN ?", folders)
	}

	if orderBy != nil && *orderBy != "" {
		query = query.Order(*orderBy)
	} else {
		query = query.Order("items.id ASC")
	}

	var page1 int
	var pageSize1 int
	if page != nil {
		page1 = *page
	} else {
		page1 = 0
	}
	if page1 < 0 {
		page1 = 0
	}
	if pageSize != nil {
		pageSize1 = *pageSize
	} else {
		pageSize1 = 1000
	}
	if pageSize1 > 1000 {
		pageSize1 = 1000
	}
	if pageSize1 < 1 {
		pageSize1 = 1
	}
	query = query.Offset(page1 * pageSize1).Limit(pageSize1)

	err := query.Preload("Folders").Preload("Tags").Find(&items).Error
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	return items, nil
}

// 计算文件内容的 SHA256 哈希值
func CalculateSHA256(reader io.Reader) (string, error) {
	hash := sha256.New()
	_, err := io.Copy(hash, reader)
	if err != nil {
		return "", err
	}
	hashBytes := hash.Sum(nil)
	return hex.EncodeToString(hashBytes), nil
}

// 根据文件路径计算 SHA256 哈希值
func CalculateFileID(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	return CalculateSHA256(file)
}

// 计算缩略图的宽度和高度，确保总像素数不超过 maxPixels，并保持原始图像的长宽比
func calculateThumbnailSize(originalWidth, originalHeight, maxPixels int) (uint, uint) {
	originalPixels := originalWidth * originalHeight
	if originalPixels <= maxPixels {
		return uint(originalWidth), uint(originalHeight)
	}

	scale := math.Sqrt(float64(maxPixels) / float64(originalPixels))

	newWidth := uint(float64(originalWidth) * scale)
	newHeight := uint(float64(originalHeight) * scale)

	return newWidth, newHeight
}

// 生成缩略图，确保缩略图总像素数不超过 maxPixels
func generateThumbnail(filePath string, thumbPath string, maxPixels int) error {
	file, err := os.Open(filePath)

	if err != nil {
		return fmt.Errorf("open file failed: %v", err)
	}

	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("decode image failed: %v", err)
	}

	originalWidth := img.Bounds().Dx()
	originalHeight := img.Bounds().Dy()

	newWidth, newHeight := calculateThumbnailSize(originalWidth, originalHeight, maxPixels)

	thumb := resize.Resize(newWidth, newHeight, img, resize.Lanczos3)

	out, err := os.Create(thumbPath)
	if err != nil {
		return fmt.Errorf("create thumb file failed: %v", err)
	}
	defer out.Close()

	err = webp.Encode(out, thumb, nil)
	if err != nil {
		return err
	}

	return nil
}

func RenameFile(oldPath string, Name string, Ext *string) error {
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", oldPath)
	}

	dir := filepath.Dir(oldPath)

	ext := filepath.Ext(oldPath)
	if Ext != nil {
		ext = *Ext
		if ext != "" {
			ext = "." + ext
		}
	}

	if Name == "" {
		return fmt.Errorf("new file name is empty")
	}

	newPath := filepath.Join(dir, Name+ext)

	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("file with name %s already exists", newPath)
	}

	err := os.Rename(oldPath, newPath)
	if err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}

	return nil
}

func AddItem(db *gorm.DB, path string, name *string, url *string, annotation *string, tags []uuid.UUID, folders []uuid.UUID, star *uint8, created_at *time.Time) error {
	fileID, err := CalculateFileID(path)
	if err != nil {
		return fmt.Errorf("failed to calculate file ID: %v", err)
	}

	var existingItem dbcommon.Item
	err = db.Unscoped().First(&existingItem, "id = ?", fileID).Error
	if err == nil {
		updates := map[string]interface{}{
			"modified_at": time.Now(),
		}
		if name != nil {
			err = RenameFile(filepath.Join(database.DbBaseDir, existingItem.ID, existingItem.Name+"."+existingItem.Ext), *name, nil)
			if err != nil {
				return fmt.Errorf("db_add_item rename exist file name failed %v", err)
			}
			updates["name"] = *name
		}
		if created_at != nil {
			updates["created_at"] = *created_at
		}
		if annotation != nil {
			updates["annotation"] = *annotation
		}
		if url != nil {
			updates["url"] = *url
		}
		if star != nil {
			updates["star"] = *star
		}
		err = db.Model(&existingItem).Updates(updates).Error
		if err != nil {
			return fmt.Errorf("failed to update existing item: %v", err)
		}

		for _, tagID := range tags {
			tag := dbcommon.Tag{ID: tagID}
			err = db.Model(&existingItem).Association("Tags").Append(&tag)
			if err != nil {
				return fmt.Errorf("failed to append tag: %v", err)
			}
		}

		for _, folderID := range folders {
			folder := dbcommon.Folder{ID: folderID}
			err = db.Model(&existingItem).Association("Folders").Append(&folder)
			if err != nil {
				return fmt.Errorf("failed to append folder: %v", err)
			}
		}

		err = os.Remove(path)
		if err != nil {
			log.Printf("Failed to delete original file: %v", err)
		}

		return nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to query existing item: %v", err)
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	ext := filepath.Ext(fileInfo.Name())

	isImage := ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif"

	var width, height uint32
	var fileSize uint64 = uint64(fileInfo.Size())

	if isImage {
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file: %v", err)
		}
		defer file.Close()

		img, _, err := image.Decode(file)
		if err == nil {
			width = uint32(img.Bounds().Dx())
			height = uint32(img.Bounds().Dy())
		} else {
			log.Printf("Failed to decode image: %v", err)
		}
	}

	rawFileDir := filepath.Join(database.DbBaseDir, "raw_files", fileID)
	err = os.MkdirAll(rawFileDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create target directory: %v", err)
	}

	var name1 string = fileInfo.Name()[:len(fileInfo.Name())-len(ext)]
	if name != nil && *name != "" {
		name1 = *name
	}
	destPath := filepath.Join(rawFileDir, name1+ext)
	err = os.Rename(path, destPath)
	if err != nil {
		return fmt.Errorf("failed to move and rename file: %v", err)
	}

	item := dbcommon.Item{
		ID:            fileID,
		CreatedAt:     time.Now(),
		ImportedAt:    time.Now(),
		ModifiedAt:    time.Now(),
		Name:          name1,
		Ext:           ext[1:],
		Width:         width,
		Height:        height,
		Size:          fileSize,
		Tags:          []dbcommon.Tag{},    // 待处理
		Folders:       []dbcommon.Folder{}, // 待处理
		HaveThumbnail: false,
		HavePreview:   false,
	}
	if created_at != nil {
		item.CreatedAt = *created_at
	}
	if star != nil {
		item.Star = *star
	}
	if url != nil {
		item.Url = *url
	}
	if annotation != nil {
		item.Annotation = *annotation
	}

	if isImage {
		thumbPath := filepath.Join(database.DbBaseDir, "thumbnails", fileID+".webp")
		err = generateThumbnail(destPath, thumbPath, 256*256)
		if err != nil {
			log.Printf("Failed to generate thumbnail: %v", err)
		} else {
			item.HaveThumbnail = true
		}

		previewPath := filepath.Join(database.DbBaseDir, "previews", fileID+".webp")
		err = generateThumbnail(destPath, previewPath, 768*768)
		if err != nil {
			log.Printf("Failed to generate preview: %v", err)
		} else {
			item.HavePreview = true
		}
	}

	for _, tagID := range tags {
		tag := dbcommon.Tag{ID: tagID}
		err = db.Model(&item).Association("Tags").Append(&tag)
		if err != nil {
			return fmt.Errorf("failed to append tag: %v", err)
		}
	}

	for _, folderID := range folders {
		folder := dbcommon.Folder{ID: folderID}
		err = db.Model(&item).Association("Folders").Append(&folder)
		if err != nil {
			return fmt.Errorf("failed to append folder: %v", err)
		}
	}

	err = db.Create(&item).Error
	if err != nil {
		return fmt.Errorf("failed to create item in database: %v", err)
	}

	return nil
}

func UpdateItem(db *gorm.DB, fileID string, name *string, ext *string, url *string, annotation *string, tags []uuid.UUID, folders []uuid.UUID, star *uint8, created_at *time.Time) error {
	var existingItem dbcommon.Item
	err := db.Unscoped().First(&existingItem, "id = ?", fileID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("item with ID %s not found", fileID)
		}
		return fmt.Errorf("failed to query existing item: %v", err)
	}

	updates := map[string]interface{}{
		"modified_at": time.Now(),
	}

	if name != nil || ext != nil {
		newName := existingItem.Name
		newExt := existingItem.Ext

		if name != nil {
			newName = *name
		}
		if ext != nil {
			newExt = *ext
		}
		if newName != existingItem.Name || newExt != existingItem.Ext {
			err = RenameFile(filepath.Join(database.DbBaseDir, "raw_files", existingItem.ID, existingItem.Name+"."+existingItem.Ext), newName, ext)
			fmt.Println(newName + "." + newExt)
			fmt.Println(filepath.Join(database.DbBaseDir, "raw_files", existingItem.ID, existingItem.Name+"."+existingItem.Ext))
			if err != nil {
				return fmt.Errorf("failed to rename file: %v", err)
			}
		}
		updates["name"] = newName
		updates["ext"] = newExt
	}

	if created_at != nil {
		updates["created_at"] = *created_at
	}

	if annotation != nil {
		updates["annotation"] = *annotation
	}

	if url != nil {
		updates["url"] = *url
	}

	if star != nil {
		updates["star"] = *star
	}

	err = db.Model(&existingItem).Updates(updates).Error
	if err != nil {
		return fmt.Errorf("failed to update existing item: %v", err)
	}

	if tags != nil {
		err = db.Model(&existingItem).Association("Tags").Clear()
		if err != nil {
			return fmt.Errorf("failed to clear tags: %v", err)
		}

		for _, tagID := range tags {
			tag := dbcommon.Tag{ID: tagID}
			err = db.Model(&existingItem).Association("Tags").Append(&tag)
			if err != nil {
				return fmt.Errorf("failed to append tag: %v", err)
			}
		}
	}

	if folders != nil {
		err = db.Model(&existingItem).Association("Folders").Clear()
		if err != nil {
			return fmt.Errorf("failed to clear folders: %v", err)
		}

		for _, folderID := range folders {
			folder := dbcommon.Folder{ID: folderID}
			err = db.Model(&existingItem).Association("Folders").Append(&folder)
			if err != nil {
				return fmt.Errorf("failed to append folder: %v", err)
			}
		}
	}

	return nil
}

func ItemSoftDelete(db *gorm.DB, itemIDs []string) error {
	err := db.Transaction(func(tx *gorm.DB) error {
		result := tx.Delete(&dbcommon.Item{}, itemIDs)
		if result.Error != nil {
			return fmt.Errorf("soft delete items failed: %v", result.Error)
		}

		return nil
	})

	return err
}

func ItemHardDelete(db *gorm.DB, itemIDs []string) error {
	err := db.Unscoped().Transaction(func(tx *gorm.DB) error {
		result := tx.Delete(&dbcommon.Item{}, itemIDs)
		if result.Error != nil {
			return fmt.Errorf("hard delete items failed: %v", result.Error)
		}

		return nil
	})

	if err != nil {
		return err
	}

	for _, itemID := range itemIDs {
		itemDir := filepath.Join(database.DbBaseDir, "raw_files", itemID)
		if err := os.RemoveAll(itemDir); err != nil {
			return fmt.Errorf("failed to delete item directory '%s': %v", itemDir, err)
		}

		thumbnailFile := filepath.Join(database.DbBaseDir, "thumbnails", itemID+".webp")
		if err := os.Remove(thumbnailFile); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete thumbnail '%s': %v", thumbnailFile, err)
		}

		previewFile := filepath.Join(database.DbBaseDir, "previews", itemID+".webp")
		if err := os.Remove(previewFile); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete thumbnail '%s': %v", previewFile, err)
		}
	}

	return nil
}
