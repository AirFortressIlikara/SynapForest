<!--
 * @Author: Ilikara 3435193369@qq.com
 * @Date: 2025-01-20 13:10:40
 * @LastEditors: ilikara 3435193369@qq.com
 * @LastEditTime: 2025-04-15 06:51:06
 * @FilePath: /SynapForest/api/doc/api.md
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
-->
# API文档

## 文件

### 获取缩略图

**URL**: `/thumbnails/:id`

**Method**: `GET`

**Response**:

- 成功: 返回图片文件（webp 格式）。

- 失败:
  ```json
  {
    "status": "error"
  }
  ```