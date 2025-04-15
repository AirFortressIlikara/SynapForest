/*
 * @Author: ilikara 3435193369@qq.com
 * @Date: 2025-04-14 15:02:37
 * @LastEditors: ilikara 3435193369@qq.com
 * @LastEditTime: 2025-04-15 02:48:23
 * @FilePath: /SynapForest/graphql/queries.go
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
package graphql

import (
	"errors"
	"fmt"
	"my_eagle/database"
	"my_eagle/database/dbcommon"

	"github.com/gofrs/uuid"
	"github.com/graphql-go/graphql"
	"gorm.io/gorm"
)

var itemType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Item",
	Fields: graphql.Fields{
		"id":             &graphql.Field{Type: graphql.String},
		"created_at":     &graphql.Field{Type: graphql.DateTime},
		"imported_at":    &graphql.Field{Type: graphql.DateTime},
		"modified_at":    &graphql.Field{Type: graphql.DateTime},
		"deleted_at":     &graphql.Field{Type: graphql.DateTime},
		"name":           &graphql.Field{Type: graphql.String},
		"ext":            &graphql.Field{Type: graphql.String},
		"width":          &graphql.Field{Type: graphql.Int},
		"height":         &graphql.Field{Type: graphql.Int},
		"size":           &graphql.Field{Type: graphql.Int},
		"url":            &graphql.Field{Type: graphql.String},
		"annotation":     &graphql.Field{Type: graphql.String},
		"star":           &graphql.Field{Type: graphql.Int},
		"have_thumbnail": &graphql.Field{Type: graphql.Boolean},
		"have_preview":   &graphql.Field{Type: graphql.Boolean},
		// 注意：tags 和 folders 关系字段后面添加
	},
})

var tagType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Tag",
	Fields: graphql.Fields{
		"id":          &graphql.Field{Type: graphql.String},
		"created_at":  &graphql.Field{Type: graphql.DateTime},
		"modified_at": &graphql.Field{Type: graphql.DateTime},
		"deleted_at":  &graphql.Field{Type: graphql.DateTime},
		"name":        &graphql.Field{Type: graphql.String},
		"description": &graphql.Field{Type: graphql.String},
		"icon":        &graphql.Field{Type: graphql.Int},
		"icon_color":  &graphql.Field{Type: graphql.Int},
		"is_expand":   &graphql.Field{Type: graphql.Boolean},
		// 注意：parent 和 children 关系字段后面添加
	},
})

var folderType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Folder",
	Fields: graphql.Fields{
		"id":          &graphql.Field{Type: graphql.String},
		"created_at":  &graphql.Field{Type: graphql.DateTime},
		"modified_at": &graphql.Field{Type: graphql.DateTime},
		"deleted_at":  &graphql.Field{Type: graphql.DateTime},
		"name":        &graphql.Field{Type: graphql.String},
		"description": &graphql.Field{Type: graphql.String},
		"icon":        &graphql.Field{Type: graphql.Int},
		"icon_color":  &graphql.Field{Type: graphql.Int},
		"is_expand":   &graphql.Field{Type: graphql.Boolean},
		// 注意：parent, children 和 items 关系字段后面添加
	},
})

// 添加关系字段
func init() {
	// 添加 Item 的关系字段
	itemType.AddFieldConfig("tags", &graphql.Field{
		Type: graphql.NewList(tagType),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			// 这里实现从数据库加载 tags 的逻辑
			item, ok := p.Source.(dbcommon.Item)
			if !ok {
				return nil, fmt.Errorf("expected Item type, got %T", p.Source)
			}

			return item.Tags, nil
		},
	})

	itemType.AddFieldConfig("folders", &graphql.Field{
		Type: graphql.NewList(folderType),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			// 从父对象获取文件夹ID
			item, ok := p.Source.(dbcommon.Item)
			if !ok {
				return nil, fmt.Errorf("expected Item type, got %T", p.Source)
			}

			return item.Folders, nil
		},
	})

	// 添加 Folder 的关系字段
	folderType.AddFieldConfig("parent", &graphql.Field{
		Type: folderType,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			// 从父对象获取文件夹ID
			folder, ok := p.Source.(dbcommon.Folder)
			if !ok {
				return nil, fmt.Errorf("expected Folder type, got %T", p.Source)
			}

			// 如果 Parent 是 nil，尝试从数据库加载
			if folder.Parent == nil {
				var parent dbcommon.Folder
				err := database.DB.
					Where("id = ?", folder.ParentID).
					Find(&parent).Error
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return nil, nil
					}
					return nil, err
				}
				return parent, nil
			}

			return folder.Parent, nil
		},
	})

	folderType.AddFieldConfig("children", &graphql.Field{
		Type: graphql.NewList(folderType),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			// 从父对象获取文件夹ID
			folder, ok := p.Source.(dbcommon.Folder)
			if !ok {
				return nil, fmt.Errorf("expected Folder type, got %T", p.Source)
			}
			var childFolders []dbcommon.Folder
			var err error
			if folder.ID == uuid.Nil {
				// 查询根目录的子目录时，确保parent_id是nil且不是根目录本身
				err = database.DB.
					Where("parent_id = ? AND id != ?", uuid.Nil, uuid.Nil).
					Preload("Parent").
					Find(&childFolders).Error
			} else {
				// 普通目录的查询
				err = database.DB.
					Where("parent_id = ?", folder.ID).
					Preload("Parent").
					Find(&childFolders).Error
			}

			if err != nil {
				return nil, fmt.Errorf("failed to query folders: %w", err)
			}

			return childFolders, nil
		},
	})

	folderType.AddFieldConfig("items", &graphql.Field{
		Type: graphql.NewList(itemType),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			// 从父对象获取文件夹ID
			folder, ok := p.Source.(dbcommon.Folder)
			if !ok {
				return nil, fmt.Errorf("expected Folder type, got %T", p.Source)
			}
			var items []dbcommon.Item
			err := database.DB.
				Joins("JOIN item_folders ON item_folders.item_id = items.id").
				Where("item_folders.folder_id = ?", folder.ID). // 使用folder.ID
				Preload("Folders").Preload("Tags").             // 如果需要预加载关联
				Find(&items).Error

			if err != nil {
				return nil, fmt.Errorf("failed to query items: %w", err)
			}

			return items, nil
		},
	})

	// 添加 Tag 的关系字段
	tagType.AddFieldConfig("parent", &graphql.Field{
		Type: folderType,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			// 这里实现从数据库加载 parent 的逻辑
			return nil, nil
		},
	})

	tagType.AddFieldConfig("children", &graphql.Field{
		Type: graphql.NewList(folderType),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			// 这里实现从数据库加载 children 的逻辑
			return nil, nil
		},
	})
}

var RootQuery = graphql.NewObject(graphql.ObjectConfig{
	Name: "Query",
	Fields: graphql.Fields{
		"folder": &graphql.Field{
			Type: folderType,
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"includeItems": &graphql.ArgumentConfig{
					Type:         graphql.Boolean,
					DefaultValue: true,
				},
				"includeFolders": &graphql.ArgumentConfig{
					Type:         graphql.Boolean,
					DefaultValue: true,
				},
				"itemFields": &graphql.ArgumentConfig{
					Type: graphql.NewList(graphql.String),
				},
				"folderFields": &graphql.ArgumentConfig{
					Type: graphql.NewList(graphql.String),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				id, _ := p.Args["id"].(string)
				includeItems, _ := p.Args["includeItems"].(bool)
				includeFolders, _ := p.Args["includeFolders"].(bool)
				itemFields, _ := p.Args["itemFields"].([]interface{})
				folderFields, _ := p.Args["folderFields"].([]interface{})

				var folder dbcommon.Folder
				if err := database.DB.First(&folder, "id = ?", id).Error; err != nil {
					return nil, err
				}

				if includeItems {
					// 根据 itemFields 选择性加载字段
					query := database.DB.Model(&dbcommon.Item{})
					if len(itemFields) > 0 {
						var selectedFields []string
						for _, field := range itemFields {
							if f, ok := field.(string); ok {
								selectedFields = append(selectedFields, f)
							}
						}
						query = query.Select(selectedFields)
					}
					if err := query.Where("id = ?", id).Find(&folder.Items).Error; err != nil {
						return nil, err
					}
				}

				if includeFolders {
					// 根据 folderFields 选择性加载字段
					query := database.DB.Model(&dbcommon.Folder{})
					if len(folderFields) > 0 {
						var selectedFields []string
						for _, field := range folderFields {
							if f, ok := field.(string); ok {
								selectedFields = append(selectedFields, f)
							}
						}
						query = query.Select(selectedFields)
					}
					if err := query.Where("parent_id = ?", id).Find(&folder.Children).Error; err != nil {
						return nil, err
					}
				}

				return folder, nil
			},
		},
	},
})
