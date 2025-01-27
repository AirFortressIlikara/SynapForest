/*
 * @Author: Ilikara 3435193369@qq.com
 * @Date: 2025-01-27 16:41:59
 * @LastEditors: Ilikara 3435193369@qq.com
 * @LastEditTime: 2025-01-27 20:43:18
 * @FilePath: /my_eagle/api/auth.go
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
package api

import (
	"my_eagle/database"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")

		// if token == "" {
		// 	var json struct {
		// 		Token string `json:"token"`
		// 	}
		// 	if err := c.ShouldBindJSON(&json); err == nil {
		// 		token = json.Token
		// 	}
		// }

		if token != database.Token {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		c.Next()
	}
}
