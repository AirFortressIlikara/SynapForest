/*
 * @Author: ilikara 3435193369@qq.com
 * @Date: 2025-04-14 15:02:37
 * @LastEditors: ilikara 3435193369@qq.com
 * @LastEditTime: 2025-04-17 12:28:09
 * @FilePath: /SynapForest/graphql/queries.go
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
package graphql

import (
	"errors"
	"fmt"
	"strings"
	"synapforest/database"
	"synapforest/database/dbcommon"

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
			Type: graphql.NewList(folderType),
			Args: graphql.FieldConfigArgument{
				"ids": &graphql.ArgumentConfig{
					Type: graphql.NewList(graphql.String),
				},
				"itemIds": &graphql.ArgumentConfig{
					Type: graphql.NewList(graphql.String),
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
				ids, _ := p.Args["ids"].([]interface{})
				itemIds, _ := p.Args["itemIds"].([]interface{})
				includeItems, _ := p.Args["includeItems"].(bool)
				includeFolders, _ := p.Args["includeFolders"].(bool)
				itemFields, _ := p.Args["itemFields"].([]interface{})
				folderFields, _ := p.Args["folderFields"].([]interface{})

				// 处理 itemIds 参数，转换为字符串切片
				var itemIDStrings []string
				for _, id := range itemIds {
					if str, ok := id.(string); ok {
						itemIDStrings = append(itemIDStrings, str)
					}
				}

				var folders []dbcommon.Folder
				var err error

				// 根据参数决定查询方式
				if len(ids) > 0 {
					var folderList []dbcommon.Folder
					if err = database.DB.Where("id IN (?)", ids).Find(&folderList).Error; err != nil {
						return nil, err
					}
					folders = append(folders, folderList...)
				} else if len(itemIDStrings) > 0 {
					// 通过itemIds查询关联的文件夹
					if err = database.DB.
						Select("folders.*").
						Joins("JOIN item_folders ON item_folders.folder_id = folders.id").
						Where("item_folders.item_id IN (?)", itemIDStrings).
						Group("folders.id").
						Having("COUNT(DISTINCT item_folders.item_id) = ?", len(itemIDStrings)).
						Find(&folders).Error; err != nil {
						return nil, err
					}
				} else {
					return nil, fmt.Errorf("必须提供id或itemIds参数")
				}

				// 处理每个文件夹的关联数据
				for i := range folders {
					folder := &folders[i]

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
						// 根据实际关联关系调整查询条件
						if err := query.Where("id = ?", folder.ID).Find(&folder.Items).Error; err != nil {
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
						if err := query.Where("parent_id = ?", folder.ID).Find(&folder.Children).Error; err != nil {
							return nil, err
						}
					}
				}

				return folders, nil
			},
		},
		"items": &graphql.Field{
			Type: graphql.NewList(itemType),
			Args: graphql.FieldConfigArgument{
				"folderIds": &graphql.ArgumentConfig{
					Type:        graphql.NewList(graphql.String),
					Description: "List of folder IDs to filter items",
				},
				"tagIds": &graphql.ArgumentConfig{
					Type:        graphql.NewList(graphql.String),
					Description: "List of tag IDs to filter items",
				},
				"folderLogic": &graphql.ArgumentConfig{
					Type:         graphql.String,
					DefaultValue: "OR",
					Description:  "Logic to apply for folder filtering (AND/OR)",
				},
				"tagLogic": &graphql.ArgumentConfig{
					Type:         graphql.String,
					DefaultValue: "OR",
					Description:  "Logic to apply for tag filtering (AND/OR)",
				},
				"combinedLogic": &graphql.ArgumentConfig{
					Type:         graphql.String,
					DefaultValue: "AND",
					Description:  "Logic to combine folder and tag filters (AND/OR)",
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				folderIds, _ := p.Args["folderIds"].([]interface{})
				tagIds, _ := p.Args["tagIds"].([]interface{})
				folderLogic, _ := p.Args["folderLogic"].(string)
				folderLogic = strings.ToUpper(folderLogic)
				tagLogic, _ := p.Args["tagLogic"].(string)
				tagLogic = strings.ToUpper(tagLogic)
				combinedLogic, _ := p.Args["combinedLogic"].(string)

				// 转换ID为字符串切片
				var folderIDStrs []string
				for _, id := range folderIds {
					if str, ok := id.(string); ok {
						folderIDStrs = append(folderIDStrs, str)
					}
				}

				var tagIDStrs []string
				for _, id := range tagIds {
					if str, ok := id.(string); ok {
						tagIDStrs = append(tagIDStrs, str)
					}
				}

				query := database.DB.Model(&dbcommon.Item{}).Distinct("items.*")

				var folderQuery, tagQuery *gorm.DB
				// 构建文件夹条件
				if len(folderIDStrs) > 0 {
					folderQuery = database.DB.
						Select("item_id").
						Table("item_folders").
						Where("folder_id IN (?)", folderIDStrs)

					switch {
					case folderLogic == "EQUAL":
						// 精确匹配：必须且只能包含这些文件夹
						// 先找出包含所有指定文件夹的item
						folderQuery = folderQuery.
							Group("item_folders.item_id").
							Having("COUNT(DISTINCT item_folders.folder_id) = ?", len(folderIDStrs))

						// 然后排除那些还包含其他文件夹的item
						query = query.
							Where("NOT EXISTS (SELECT 1 FROM item_folders WHERE item_folders.item_id = items.id AND item_folders.folder_id NOT IN (?))", folderIDStrs)
					case folderLogic == "AND":
						// AND 逻辑：必须包含所有指定的文件夹（但可以包含其他文件夹）
						folderQuery = folderQuery.
							Group("item_id").
							Having("COUNT(DISTINCT folder_id) = ?", len(folderIDStrs))
					case folderLogic == "OR":
						// OR 逻辑：只需包含任意一个指定的文件夹
						// 不需要额外处理
					}
				}

				// 构建标签条件
				if len(tagIDStrs) > 0 {
					tagQuery = database.DB.
						Select("item_id").
						Table("item_tags").
						Where("tag_id IN (?)", tagIDStrs)

					switch {
					case tagLogic == "EQUAL":
						// 精确匹配：必须且只能包含这些标签
						// 先找出包含所有指定标签的item
						tagQuery = tagQuery.
							Group("item_tags.item_id").
							Having("COUNT(DISTINCT item_tags.tag_id) = ?", len(tagIDStrs))

						// 然后排除那些还包含其他标签的item
						query = query.
							Where("NOT EXISTS (SELECT 1 FROM item_tags WHERE item_tags.item_id = items.id AND item_tags.tag_id NOT IN (?))", tagIDStrs)
					case tagLogic == "AND":
						// AND 逻辑：必须包含所有指定的标签（但可以包含其他标签）
						tagQuery = tagQuery.
							Group("item_id").
							Having("COUNT(DISTINCT tag_id) = ?", len(tagIDStrs))
					case tagLogic == "OR":
						// OR 逻辑：只需包含任意一个指定的标签
						// 不需要额外处理
					}
				}

				// 组合文件夹和标签条件（统一 JOIN 处理）
				if len(folderIDStrs) > 0 && len(tagIDStrs) > 0 {
					if combinedLogic == "OR" {
						// OR 逻辑：使用 UNION 组合 folderQuery 和 tagQuery
						combinedQuery := database.DB.
							Table("(?) UNION (?) AS combined_items",
								folderQuery.Select("item_id"),
								tagQuery.Select("item_id")).
							Select("DISTINCT item_id")
						query = query.Joins("JOIN (?) AS combined_items ON combined_items.item_id = items.id", combinedQuery)
					} else {
						// AND 逻辑（默认）：分别 JOIN folderQuery 和 tagQuery
						query = query.
							Joins("JOIN (?) AS folder_items ON folder_items.item_id = items.id", folderQuery).
							Joins("JOIN (?) AS tag_items ON tag_items.item_id = items.id", tagQuery)
					}
				} else if len(folderIDStrs) > 0 {
					// 只有文件夹条件
					query = query.Joins("JOIN (?) AS folder_items ON folder_items.item_id = items.id", folderQuery)
				} else if len(tagIDStrs) > 0 {
					// 只有标签条件
					query = query.Joins("JOIN (?) AS tag_items ON tag_items.item_id = items.id", tagQuery)
				}

				// 预加载关联数据
				query = query.Preload("Folders").Preload("Tags")

				var items []dbcommon.Item
				if err := query.Find(&items).Error; err != nil {
					return nil, fmt.Errorf("failed to query items: %w", err)
				}

				return items, nil
			},
		},
	},
})
