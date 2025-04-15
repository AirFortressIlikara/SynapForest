/*
 * @Author: ilikara 3435193369@qq.com
 * @Date: 2025-04-14 15:02:29
 * @LastEditors: ilikara 3435193369@qq.com
 * @LastEditTime: 2025-04-14 15:59:56
 * @FilePath: /SynapForest/graphql/schema.go
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
package graphql

import (
	"github.com/graphql-go/graphql"
)

var Schema graphql.Schema

func init() {
	var err error
	Schema, err = graphql.NewSchema(graphql.SchemaConfig{
		Query: RootQuery,
		// Mutation: RootMutation,
	})
	if err != nil {
		panic(err)
	}
}
