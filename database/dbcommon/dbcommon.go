/*
 * @Author: ilikara 3435193369@qq.com
 * @Date: 2024-12-31 09:03:50
 * @LastEditors: ilikara 3435193369@qq.com
 * @LastEditTime: 2024-12-31 10:36:29
 * @FilePath: /my_eagle/database/common/common.go
 * @Description:
 *
 * Copyright (c) 2024 by ${git_name_email}, All Rights Reserved.
 */
package dbcommon

import (
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

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

	HaveThumbnail bool `json:"have_thumbnail"` // 是否有缩略图
	HavePreview   bool `json:"have_preview"`   // 是否有预览图
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
