/*
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
package vector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

// 序列化向量为 JSON
func SerializeVector(vec []float32) ([]byte, error) {
	return json.Marshal(vec)
}

// 反序列化 JSON 为向量
func DeserializeVector(data []byte) ([]float32, error) {
	var vec []float32
	err := json.Unmarshal(data, &vec)
	return vec, err
}

// 计算图片向量
func VectorizeImage(imagePath string) ([]float32, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "image.jpg")
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %v", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file to form: %v", err)
	}
	writer.Close()

	url := "http://localhost:8000/vectorize_image"
	resp, err := http.Post(url, writer.FormDataContentType(), body)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var result map[string][]float32
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return result["vector"], nil
}

// 计算文本向量
func VectorizeText(text string) ([]float32, error) {
	formData := strings.NewReader("text=" + text)

	url := "http://localhost:8000/vectorize_text"
	resp, err := http.Post(url, "application/x-www-form-urlencoded", formData)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var result map[string][]float32
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return result["vector"], nil
}
