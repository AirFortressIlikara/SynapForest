/*
 * @Author: ilikara 3435193369@qq.com
 * @Date: 2024-12-29 12:43:00
 * @LastEditors: ilikara 3435193369@qq.com
 * @LastEditTime: 2024-12-29 15:00:49
 * @FilePath: /my_eagle/api/folder/folder.go
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
package folder

import (
	"my_eagle/database"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

// 请求的结构体
type CreateFolderRequest struct {
	FolderName string    `json:"folderName"`
	Parent     uuid.UUID `json:"parent"`
	Token      string    `json:"token"`
}

// 返回的结构体
type FolderResponse struct {
	Status string `json:"status"`
	Data   struct {
		ID             uuid.UUID         `json:"id"`
		Name           string            `json:"name"`
		Images         []string          `json:"images"`
		Folders        []uuid.UUID       `json:"folders"`
		ModifiedAt     time.Time         `json:"modified_at"`
		ImagesMappings map[string]string `json:"imagesMappings"`
		Tags           []string          `json:"tags"`
		Parent         uuid.UUID         `json:"parent"`
		Children       []uuid.UUID       `json:"children"`
		IsExpand       bool              `json:"isExpand"`
	} `json:"data"`
}

func CreateFolder(c *gin.Context) {
	var req CreateFolderRequest

	// 绑定JSON数据到结构体
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request data",
		})
		return
	}

	folder, err := database.CreateFolder(database.DB, req.FolderName, "", 0, 0, req.Parent, false)
	if err != nil {
		// 返回JSON响应
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Folder Create Failed",
		})
		return
	}

	// 构建响应数据
	resp := FolderResponse{
		Status: "success",
	}
	resp.Data.ID = folder.ID
	resp.Data.Name = req.FolderName
	resp.Data.ModifiedAt = folder.ModifiedAt
	resp.Data.Parent = folder.ParentID
	resp.Data.Children, _ = database.GetChildFolderIDs(database.DB, folder.ID)
	resp.Data.IsExpand = true

	// 返回JSON响应
	c.JSON(http.StatusOK, resp)
}

func ListFolder(c *gin.Context) {

}

func UpdateFolder(c *gin.Context) {

}
