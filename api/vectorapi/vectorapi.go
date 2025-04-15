/*
 * @Author: Ilikara 3435193369@qq.com
 * @Date: 2025-02-02 17:25:29
 * @LastEditors: ilikara 3435193369@qq.com
 * @LastEditTime: 2025-04-15 06:50:22
 * @FilePath: /SynapForest/api/vectorapi/vectorapi.go
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
package vectorapi

import (
	"fmt"
	"net/http"
	"path/filepath"
	"synapforest/database"
	"synapforest/database/dbcommon"
	"synapforest/database/itemdb"
	"synapforest/vector"
	"time"

	"github.com/gin-gonic/gin"
)

// 处理向量化请求
func HandleVectorize(c *gin.Context) {
	// 获取ID
	id := c.Param("id")

	// 查询Item
	items, err := itemdb.GetItemsByIDs(database.DB, []string{id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get item: %v", err)})
		return
	}
	if len(items) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}
	item := items[0]

	// 计算图片向量
	imagePath := filepath.Join(database.DbBaseDir, "raw_files", item.ID, item.Name+"."+item.Ext)
	imageVec, err := vector.VectorizeImage(imagePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to vectorize image: %v", err)})
		return
	}

	imageVec_, _ := vector.SerializeVector(imageVec)

	// 存储向量到数据库
	itemVector := dbcommon.ItemVector{
		ItemID:     item.ID,
		ImageVec:   imageVec_,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}

	result := database.VectorDB.Create(&itemVector)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save vector: %v", result.Error)})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"message":    "Vectorization successful",
		"item_id":    item.ID,
		"image_vec":  imageVec,
		"created_at": itemVector.CreatedAt,
	})
}
