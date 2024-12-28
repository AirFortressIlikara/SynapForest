package database

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Item struct {
	ID            uint      `json:"id" gorm:"primaryKey"` // 主键，文件的Hash
	Name          string    `json:"name"`                 // 名称
	Ext           string    `json:"ext"`                  // 扩展名
	Width         uint      `json:"width"`                // 宽度
	Height        uint      `json:"height"`               // 高度
	Url           string    `json:"url"`                  // 文件来源URL
	IsDeleted     bool      `json:"is_deleted"`           // 是否删除
	Annotation    string    `json:"annotation"`           // 注释
	Tags          []uint32  `json:"tags"`                 // Tags
	Folders       []uint32  `json:"folders"`              // 文件夹ID列表
	Palettes      []uint32  `json:"palettes"`             // 色票（这是什么？）
	Size          int64     `json:"size"`                 // 文件大小
	Star          uint      `json:"star"`                 // 星级评分
	CreatedAt     time.Time `json:"created_at"`           // 创建时间
	ImportedAt    time.Time `json:"imported_at"`          // 导入时间
	ModifiedAt    time.Time `json:"modified_at"`          // 修改时间
	NoThumbnail   bool      `json:"no_thumbnail"`         // 是否有缩略图
	NoPreview     bool      `json:"no_preview"`           // 是否有预览图
	FilePath      string    `json:"file_path"`            // 文件路径
	FileUrl       string    `json:"file_url"`             // 文件URL
	ThumbnailPath string    `json:"thumbnail_path"`       // 缩略图路径
	ThumbnailUrl  string    `json:"thumbnail_url"`        // 缩略图URL
}

type Folder struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Icon        uint      `json:"icon"`
	IconColor   uint      `json:"icon_color"`
	CreatedAt   time.Time `json:"created_at"`
	ModifiedAt  time.Time `json:"modified_at"`
	Parent      uint32    `json:"parent"`
	Children    []uint32  `json:"children" gorm:"type:json"`
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
