package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Item struct {
	ID         uint           `json:"id" gorm:"primaryKey"` // 主键，文件的Hash
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

	Palettes []uint32 `json:"palettes"` // 色票（这是什么？）
	Star     uint8    `json:"star"`     // 星级评分

	NoThumbnail   bool   `json:"no_thumbnail"`   // 是否有缩略图
	NoPreview     bool   `json:"no_preview"`     // 是否有预览图
	FilePath      string `json:"file_path"`      // 文件路径
	FileUrl       string `json:"file_url"`       // 文件URL
	ThumbnailPath string `json:"thumbnail_path"` // 缩略图路径
	ThumbnailUrl  string `json:"thumbnail_url"`  // 缩略图URL
}

type Folder struct {
	ID         uuid.UUID      `json:"id" gorm:"primaryKey"`
	CreatedAt  time.Time      `json:"created_at"`  // 创建时间
	ModifiedAt time.Time      `json:"modified_at"` // 修改时间
	DeletedAt  gorm.DeletedAt `json:"deleted_at"`  // 删除时间

	ParentID uuid.UUID `json:"parent"` // 父文件夹
	Parent   *Folder   `gorm:"foreignKey:ParentID"`
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

	ParentID uuid.UUID `json:"parent"` // 父Tag
	Parent   *Folder   `gorm:"foreignKey:ParentID"`
	Children []Folder  `gorm:"foreignKey:ParentID"` // 子Tag

	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        uint32 `json:"icon"`
	IconColor   uint32 `json:"icon_color"`

	Items []Item `gorm:"many2many:item_tags;"`

	IsExpand bool `json:"is_expand"`
}

var db *gorm.DB

func Database_init(library_dir string) error {
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
	if err := db.AutoMigrate(&Item{}, &Folder{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	return err
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

func UpdateFolder(db *gorm.DB, folderID uuid.UUID, name string, description string, icon uint32, iconColor uint32, parent_id uuid.UUID, children []uuid.UUID) (*Folder, error) {
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
	db.Model(&folder).Association("Children").Replace()
	folder.ModifiedAt = time.Now() // 更新修改时间

	// 保存更新后的文件夹
	if err := db.Save(&folder).Error; err != nil {
		return nil, err
	}

	return &folder, nil
}
