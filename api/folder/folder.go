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

	folder, err := database.CreateFolder(database.db, req.FolderName, nil, nil, nil, req.Parent)
	if err != nil {
		// 返回JSON响应
		c.JSON(http.StatusBadRequest)
	}
	// 构建响应数据
	resp := FolderResponse{
		Status: "success",
	}
	resp.Data.ID = folder.ID
	resp.Data.Name = req.FolderName
	resp.Data.Images = []string{}
	resp.Data.Folders = []uuid.UUID{}
	resp.Data.ModifiedAt = folder.ModifiedAt
	resp.Data.ImagesMappings = make(map[string]string)
	resp.Data.Tags = []string{}
	resp.Data.Parent = folder.Parent
	resp.Data.Children = folder.Children
	resp.Data.IsExpand = true

	// 返回JSON响应
	c.JSON(http.StatusOK, resp)
}

func ListFolder(c *gin.Context) {

}

func UpdateFolder(c *gin.Context) {

}
