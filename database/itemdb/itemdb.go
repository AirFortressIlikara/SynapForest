/*
 * @Author: ilikara 3435193369@qq.com
 * @Date: 2024-12-31 08:55:46
 * @LastEditors: ilikara 3435193369@qq.com
 * @LastEditTime: 2025-01-03 10:53:57
 * @FilePath: /my_eagle/database/itemdb/itemdb.go
 * @Description:
 *
 * Copyright (c) 2024 by ${git_name_email}, All Rights Reserved.
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

	// 开始查询
	query := db.Model(&dbcommon.Item{})

	// 查询软删除的记录
	if isDeleted != nil && *isDeleted {
		query = query.Unscoped().Where("deleted_at IS NOT NULL")
	}

	// 根据扩展名 (ext) 过滤
	if len(exts) > 0 {
		query = query.Where("ext IN ?", exts)
	}

	// 根据关键字 (keyword) 模糊查询，假设我们只关心 'Name' 字段
	if keyword != nil && *keyword != "" {
		query = query.Where("name LIKE ?", "%"+*keyword+"%")
	}

	// 根据 tags 过滤，假设 tags 与 item 是多对多关系
	if len(tags) > 0 {
		query = query.Joins("JOIN item_tags ON item_tags.item_id = items.id").
			Where("item_tags.tag_id IN ?", tags)
	}

	// 根据 folders 过滤，假设 folders 与 item 是多对多关系
	if len(folders) > 0 {
		query = query.Joins("JOIN item_folders ON item_folders.item_id = items.id").
			Where("item_folders.folder_id IN ?", folders)
	}

	// 根据排序字段 (orderBy)，如果为空则默认按 ID 排序
	if orderBy != nil && *orderBy != "" {
		query = query.Order(*orderBy)
	} else {
		query = query.Order("items.id ASC") // 默认按 ID 升序
	}

	// 分页支持：使用 OFFSET 和 LIMIT
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

	// 执行查询，查询结果为 item 的列表
	err := query.Find(&items).Error
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
	// 计算原始图像的总像素数
	originalPixels := originalWidth * originalHeight
	if originalPixels <= maxPixels {
		// 如果原始图像的像素小于最大像素数，直接返回原始尺寸
		return uint(originalWidth), uint(originalHeight)
	}

	// 计算缩放比例
	scale := math.Sqrt(float64(maxPixels) / float64(originalPixels))

	// 计算新的宽度和高度，保持原始的长宽比
	newWidth := uint(float64(originalWidth) * scale)
	newHeight := uint(float64(originalHeight) * scale)

	return newWidth, newHeight
}

// 生成缩略图，确保缩略图总像素数不超过 maxPixels
func generateThumbnail(filePath string, thumbPath string, maxPixels int) error {
	// 打开文件
	file, err := os.Open(filePath)

	if err != nil {
		return fmt.Errorf("open file failed: %v", err)
	}

	defer file.Close()

	// 解码图像
	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("decode image failed: %v", err)
	}

	// 获取原始图像的宽度和高度
	originalWidth := img.Bounds().Dx()
	originalHeight := img.Bounds().Dy()

	// 计算缩略图的大小
	newWidth, newHeight := calculateThumbnailSize(originalWidth, originalHeight, maxPixels)

	// 生成缩略图
	thumb := resize.Resize(newWidth, newHeight, img, resize.Lanczos3)

	// 创建缩略图文件
	out, err := os.Create(thumbPath)
	if err != nil {
		return fmt.Errorf("create thumb file failed: %v", err)
	}
	defer out.Close()

	// 保存缩略图为 JPG 格式（也可以保存为 PNG 或其他格式）
	err = webp.Encode(out, thumb, nil)
	if err != nil {
		return err
	}

	return nil
}

func RenameFile(oldPath, Name string) error {
	// 检查源文件是否存在
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", oldPath)
	}

	// 获取文件所在的目录
	dir := filepath.Dir(oldPath)

	// 获取旧文件的扩展名
	ext := filepath.Ext(oldPath)

	if Name == "" {
		return fmt.Errorf("new file name is empty")
	}

	// 构造新的文件路径，附加旧文件的扩展名
	newPath := filepath.Join(dir, Name+ext)

	// 检查目标文件是否已经存在，防止覆盖
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("file with name %s already exists", newPath)
	}

	// 执行文件重命名
	err := os.Rename(oldPath, newPath)
	if err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}

	return nil
}

func AddItem(db *gorm.DB, path string, name *string, url *string, annotation *string, tags []uuid.UUID, folders []uuid.UUID, star *uint8, created_at *time.Time) error {
	// 根据文件路径计算 SHA-256 哈希值
	fileID, err := CalculateFileID(path)
	if err != nil {
		return fmt.Errorf("failed to calculate file ID: %v", err)
	}

	// 检查文件是否已存在
	var existingItem dbcommon.Item
	err = db.Unscoped().First(&existingItem, "id = ?", fileID).Error
	if err == nil {
		updates := map[string]interface{}{
			"modified_at": time.Now(),
		}
		// 如果文件已存在，更新 name、annotation、created_at、url、star 和 ModifiedAt
		if name != nil {
			err = RenameFile(filepath.Join(database.DbBaseDir, existingItem.ID, existingItem.Name+"."+existingItem.Ext), *name)
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

		// 扩展 Tags 和 Folders
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

		// 删除原始路径的文件
		err = os.Remove(path)
		if err != nil {
			log.Printf("Failed to delete original file: %v", err)
		}

		return nil // 已处理完成
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to query existing item: %v", err)
	}

	// 获取文件信息
	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	// 获取文件的扩展名
	ext := filepath.Ext(fileInfo.Name())

	// 判断文件是否为图片
	isImage := ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif"

	// 读取文件大小
	var width, height uint32
	var fileSize uint64 = uint64(fileInfo.Size())

	// 如果是图片，打开文件并解码图像获取宽高
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

	// 移动文件到目标路径
	rawFileDir := filepath.Join(database.DbBaseDir, "raw_files", fileID)
	err = os.MkdirAll(rawFileDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create target directory: %v", err)
	}

	// 移动文件
	var name1 string = fileInfo.Name()[:len(fileInfo.Name())-len(ext)]
	if name != nil && *name != "" {
		name1 = *name
	}
	// 构造新的目标路径
	destPath := filepath.Join(rawFileDir, name1+ext)
	// 移动并重命名文件
	err = os.Rename(path, destPath)
	if err != nil {
		return fmt.Errorf("failed to move and rename file: %v", err)
	}

	// 创建数据库记录
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

	// 生成缩略图（如果是图片）
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

	// 处理 Tags 和 Folders 的多对多关系
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

	//  保存 dbcommon.Item 到数据库
	err = db.Create(&item).Error
	if err != nil {
		return fmt.Errorf("failed to create item in database: %v", err)
	}

	return nil
}

func ItemSoftDelete(db *gorm.DB, itemIDs []string) error {
	result := db.Delete(&dbcommon.Item{}, itemIDs)
	if result.Error != nil {
		return fmt.Errorf("soft delete items failed: %v", result.Error)
	}
	return nil
}

func ItemHardDelete(db *gorm.DB, itemIDs []string) error {
	// 1. 硬删除数据库记录
	result := db.Unscoped().Delete(&dbcommon.Item{}, itemIDs)
	if result.Error != nil {
		return fmt.Errorf("hard delete items failed: %v", result.Error)
	}

	// 2. 删除文件和缩略图
	for _, itemID := range itemIDs {
		// 删除文件夹及其子文件
		itemDir := filepath.Join(database.DbBaseDir, "raw_files", itemID)
		if err := os.RemoveAll(itemDir); err != nil {
			return fmt.Errorf("failed to delete item directory '%s': %v", itemDir, err)
		}

		// 删除缩略图文件
		thumbnailFile := filepath.Join(database.DbBaseDir, "thumbnails", itemID+".webp")
		if err := os.Remove(thumbnailFile); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete thumbnail '%s': %v", thumbnailFile, err)
		}
	}

	return nil
}
