/*
 * @Author: ilikara 3435193369@qq.com
 * @Date: 2025-04-14 15:05:51
 * @LastEditors: ilikara 3435193369@qq.com
 * @LastEditTime: 2025-04-14 15:06:27
 * @FilePath: /SynapForest/api/graphql/handler.go
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
package graphql

import (
	"my_eagle/graphql"

	"github.com/graphql-go/handler"
)

func NewHandler() *handler.Handler {
	return handler.New(&handler.Config{
		Schema:   &graphql.Schema,
		Pretty:   true,
		GraphiQL: true,
	})
}
