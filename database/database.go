/*
 * @Author: ilikara 3435193369@qq.com
 * @Date: 2024-12-29 12:43:00
 * @LastEditors: ilikara 3435193369@qq.com
 * @LastEditTime: 2024-12-30 09:14:33
 * @FilePath: /my_eagle/database/database.go
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
package database

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"log"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/chai2010/webp"
	"github.com/gofrs/uuid"
	"github.com/nfnt/resize"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Item struct {
	ID         string         `json:"id" gorm:"primaryKey"` // 主键，文件的Hash
	CreatedAt  time.Time      `json:"created_at"`           // 创建时间
	ImportedAt time.Time      `json:"imported_at"`          // 导入时间
	ModifiedAt time.Time      `json:"modified_at"`          // 修改时间
	DeletedAt  gorm.DeletedAt `json:"deleted_at"`           // 删除时间

	Name string `json:"name"` // 名称
	Ext  string `json:"ext"`  // 扩展名

	Width  uint32 `json:"width"`  // 宽度
	Height uint32 `json:"height"` // 高度
	Size   uint64 `json:"size"`   // 文件大小

	Url        string `json:"url"`        // 文件来源URL
	Annotation string `json:"annotation"` // 注释

	Tags    []Tag    `gorm:"many2many:item_tags;"`    // Tags
	Folders []Folder `gorm:"many2many:item_folders;"` // 文件夹ID列表

	// Palettes []uint32 `json:"palettes"` // 色票（这是什么？）
	Star uint8 `json:"star"` // 星级评分

	NoThumbnail bool `json:"no_thumbnail"` // 是否有缩略图
	NoPreview   bool `json:"no_preview"`   // 是否有预览图
}

type Folder struct {
	ID         uuid.UUID      `json:"id" gorm:"primaryKey"`
	CreatedAt  time.Time      `json:"created_at"`  // 创建时间
	ModifiedAt time.Time      `json:"modified_at"` // 修改时间
	DeletedAt  gorm.DeletedAt `json:"deleted_at"`  // 删除时间

	ParentID uuid.UUID `json:"parent"`              // 父文件夹 ID
	Parent   *Folder   `gorm:"foreignKey:ParentID"` // 父文件夹
	Children []Folder  `gorm:"foreignKey:ParentID"` // 子文件夹

	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        uint32 `json:"icon"`
	IconColor   uint32 `json:"icon_color"`

	Items []Item `gorm:"many2many:item_folders;"`

	IsExpand bool `json:"is_expand"`
}

type Tag struct {
	ID         uuid.UUID      `json:"id" gorm:"primaryKey"`
	CreatedAt  time.Time      `json:"created_at"`  // 创建时间
	ModifiedAt time.Time      `json:"modified_at"` // 修改时间
	DeletedAt  gorm.DeletedAt `json:"deleted_at"`  // 删除时间

	ParentID uuid.UUID `json:"parent"`              // 父Tag ID
	Parent   *Folder   `gorm:"foreignKey:ParentID"` // 父Tag
	Children []Folder  `gorm:"foreignKey:ParentID"` // 子Tag

	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        uint32 `json:"icon"`
	IconColor   uint32 `json:"icon_color"`

	Items []Item `gorm:"many2many:item_tags;"`

	IsExpand bool `json:"is_expand"`
}

var DB *gorm.DB

func Database_init(library_dir string) (*gorm.DB, error) {
	var err error

	// 创建数据库存储路径
	if err := os.MkdirAll(library_dir, os.ModePerm); err != nil {
		log.Fatalf("failed to create data directory: %v", err)
	}

	// 打开 SQLite 数据库
	db, err := gorm.Open(sqlite.Open(filepath.Join(library_dir, "files.db")), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// 自动迁移数据库
	if err := db.AutoMigrate(&Item{}, &Folder{}, &Tag{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	return db, err
}

func CreateFolder(db *gorm.DB, name string, description string, icon uint32, iconColor uint32, parent_id uuid.UUID, is_expand bool) (*Folder, error) {
	if newUUID, err := uuid.NewV4(); err != nil {
		log.Printf("failed to generate UUID %v", err)
		return nil, fmt.Errorf("failed to generate UUID: %v", err)
	} else {
		folder := Folder{
			ID:          newUUID,
			Name:        name,
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

func UpdateFolder(db *gorm.DB, folderID uuid.UUID, name string, description string, icon uint32, iconColor uint32, parent_id uuid.UUID) (*Folder, error) {
	// 查找指定 ID 的文件夹
	var folder Folder
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

// 查询直接子文件夹的 UUID
func GetChildFolderIDs(db *gorm.DB, folderID uuid.UUID) ([]uuid.UUID, error) {
	var childIDs []uuid.UUID

	// 查询直接子文件夹的 UUID
	err := db.Model(&Folder{}).Where("parent_id = ?", folderID).Pluck("id", &childIDs).Error
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
	result := db.Model(&Folder{}).
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

// 查找符合条件的 items
func ItemList(db *gorm.DB, orderBy string, page int, pageSize int, exts []string, keyword string, tags []uuid.UUID, folders []uuid.UUID) ([]Item, error) {
	var items []Item

	// 校验 page 和 pageSize 参数
	if page <= 0 {
		page = 1 // 默认设置为第一页
	}
	if pageSize <= 0 {
		pageSize = 20 // 默认每页 20 条记录
	}
	if pageSize > 1000 {
		pageSize = 1000 // 设置最大每页 1000 条记录
	}

	// 开始查询
	query := db.Model(&Item{})

	// 根据扩展名 (ext) 过滤
	if len(exts) > 0 {
		query = query.Where("ext IN ?", exts)
	}

	// 根据关键字 (keyword) 模糊查询，假设我们只关心 'Name' 字段
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
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
	if orderBy != "" {
		query = query.Order(orderBy)
	} else {
		query = query.Order("items.id ASC") // 默认按 ID 升序
	}

	// 分页支持：使用 OFFSET 和 LIMIT
	if page > 0 && pageSize > 0 {
		query = query.Offset((page - 1) * pageSize).Limit(pageSize)
	}

	// 执行查询，查询结果为 item 的列表
	err := query.Find(&items).Error
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	return items, nil
}

// 用来根据文件计算 SHA-256 ID
func calculateFileID(filePath string) (string, error) {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	hash := sha256.New()
	hash.Write(fileData)
	return hex.EncodeToString(hash.Sum(nil)), nil
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
		return fmt.Errorf("Open File Failed: %v", err)
	}

	defer file.Close()

	// 解码图像
	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("Decode image failed: %v", err)
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
		return fmt.Errorf("Create thumb file failed: %v", err)
	}
	defer out.Close()

	// 保存缩略图为 JPG 格式（也可以保存为 PNG 或其他格式）
	err = webp.Encode(out, thumb, nil)
	if err != nil {
		return err
	}

	return nil
}

func AddItem(db *gorm.DB, path string, url string, annotation string, tags []uuid.UUID, folders []uuid.UUID, star uint8) error {
	// 1. 根据文件路径计算 SHA-256 哈希值
	fileID, err := calculateFileID(path)
	if err != nil {
		return fmt.Errorf("failed to calculate file ID: %v", err)
	}

	// 2. 获取文件信息
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

	// 3. 如果是图片，打开文件并解码图像获取宽高
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

	// 4. 移动文件到目标路径
	rawFileDir := "./raw_file/" + fileID
	err = os.MkdirAll(rawFileDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create target directory: %v", err)
	}

	// 移动文件
	destPath := filepath.Join(rawFileDir, fileInfo.Name())
	err = os.Rename(path, destPath)
	if err != nil {
		return fmt.Errorf("failed to move file: %v", err)
	}

	// 5. 生成缩略图（如果是图片）
	if isImage {
		thumbPath := "./thumbnails/" + fileID + ".webp"
		err = generateThumbnail(destPath, thumbPath, 256*256)
		if err != nil {
			log.Printf("Failed to generate thumbnail: %v", err)
		}
	}

	// 6. 创建数据库记录
	item := Item{
		ID:          fileID,
		CreatedAt:   time.Now(),
		ImportedAt:  time.Now(),
		ModifiedAt:  time.Now(),
		Name:        fileInfo.Name(),
		Ext:         ext,
		Width:       width,
		Height:      height,
		Size:        fileSize,
		Url:         url,
		Annotation:  annotation,
		Tags:        []Tag{},    // 待处理
		Folders:     []Folder{}, // 待处理
		Star:        star,
		NoThumbnail: false,
		NoPreview:   false,
	}

	// 7. 处理 Tags 和 Folders 的多对多关系
	for _, tagID := range tags {
		tag := Tag{ID: tagID}
		db.Model(&item).Association("Tags").Append(&tag)
	}

	for _, folderID := range folders {
		folder := Folder{ID: folderID}
		db.Model(&item).Association("Folders").Append(&folder)
	}

	// 8. 保存 Item 到数据库
	err = db.Create(&item).Error
	if err != nil {
		return fmt.Errorf("failed to create item in database: %v", err)
	}

	return nil
}
