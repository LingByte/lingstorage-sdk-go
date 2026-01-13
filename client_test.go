package lingstorage

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	config := &Config{
		BaseURL:   "https://example.com",
		APIKey:    "test-key",
		APISecret: "test-secret",
	}

	client := NewClient(config)
	assert.NotNil(t, client)
	assert.Equal(t, config.BaseURL, client.config.BaseURL)
	assert.Equal(t, config.APIKey, client.config.APIKey)
	assert.Equal(t, config.APISecret, client.config.APISecret)
	assert.Equal(t, 30*time.Second, client.config.Timeout)
	assert.Equal(t, 3, client.config.RetryCount)
	assert.Equal(t, "LingStorage-SDK/1.0.0", client.config.UserAgent)
}

func TestUploadFile(t *testing.T) {
	// 创建测试文件
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Hello, World!"
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	require.NoError(t, err)

	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/public/upload", r.URL.Path)
		assert.Equal(t, "test-key", r.Header.Get("X-API-Key"))
		assert.Equal(t, "test-secret", r.Header.Get("X-API-Secret"))
		assert.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")

		// 解析 multipart form
		err := r.ParseMultipartForm(32 << 20)
		require.NoError(t, err)

		// 验证文件
		file, header, err := r.FormFile("file")
		require.NoError(t, err)
		defer file.Close()

		assert.Equal(t, "test.txt", header.Filename)

		content, err := io.ReadAll(file)
		require.NoError(t, err)
		assert.Equal(t, testContent, string(content))

		// 验证其他字段
		assert.Equal(t, "test-bucket", r.FormValue("bucket"))
		assert.Equal(t, "test/key.txt", r.FormValue("key"))

		// 返回 LingStorage 格式的成功响应
		response := map[string]interface{}{
			"code": 200,
			"msg":  "File uploaded successfully",
			"data": map[string]interface{}{
				"key":          "test/key.txt",
				"bucket":       "test-bucket",
				"filename":     "test.txt",
				"size":         int64(len(testContent)),
				"originalSize": int64(len(testContent)),
				"compressed":   false,
				"watermarked":  false,
				"url":          "https://example.com/uploads/test/key.txt",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// 创建客户端
	client := NewClient(&Config{
		BaseURL:   server.URL,
		APIKey:    "test-key",
		APISecret: "test-secret",
	})

	// 测试上传
	result, err := client.UploadFile(&UploadRequest{
		FilePath: testFile,
		Bucket:   "test-bucket",
		Key:      "test/key.txt",
	})

	require.NoError(t, err)
	assert.Equal(t, "test/key.txt", result.Key)
	assert.Equal(t, "test-bucket", result.Bucket)
	assert.Equal(t, "test.txt", result.Filename)
	assert.Equal(t, int64(len(testContent)), result.Size)
	assert.Equal(t, "https://example.com/uploads/test/key.txt", result.URL)
}

func TestUploadBytes(t *testing.T) {
	testContent := []byte("Hello from bytes!")

	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证认证头
		assert.Equal(t, "test-key", r.Header.Get("X-API-Key"))
		assert.Equal(t, "test-secret", r.Header.Get("X-API-Secret"))

		// 解析 multipart form
		err := r.ParseMultipartForm(32 << 20)
		require.NoError(t, err)

		// 验证文件
		file, header, err := r.FormFile("file")
		require.NoError(t, err)
		defer file.Close()

		assert.Equal(t, "data.txt", header.Filename)

		content, err := io.ReadAll(file)
		require.NoError(t, err)
		assert.Equal(t, testContent, content)

		// 返回 LingStorage 格式的成功响应
		response := map[string]interface{}{
			"code": 200,
			"msg":  "File uploaded successfully",
			"data": map[string]interface{}{
				"key":          "bytes/data.txt",
				"bucket":       "default",
				"filename":     "data.txt",
				"size":         int64(len(testContent)),
				"originalSize": int64(len(testContent)),
				"compressed":   false,
				"watermarked":  false,
				"url":          "https://example.com/uploads/bytes/data.txt",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// 创建客户端
	client := NewClient(&Config{
		BaseURL:   server.URL,
		APIKey:    "test-key",
		APISecret: "test-secret",
	})

	// 测试上传
	result, err := client.UploadBytes(&UploadBytesRequest{
		Data:     testContent,
		Filename: "data.txt",
		Bucket:   "default",
		Key:      "bytes/data.txt",
	})

	require.NoError(t, err)
	assert.Equal(t, "bytes/data.txt", result.Key)
	assert.Equal(t, "default", result.Bucket)
	assert.Equal(t, "data.txt", result.Filename)
	assert.Equal(t, int64(len(testContent)), result.Size)
}

func TestUploadWithImageProcessing(t *testing.T) {
	// 创建测试文件
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.jpg")
	testContent := "fake image content"
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	require.NoError(t, err)

	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证认证头
		assert.Equal(t, "test-key", r.Header.Get("X-API-Key"))
		assert.Equal(t, "test-secret", r.Header.Get("X-API-Secret"))

		// 解析 multipart form
		err := r.ParseMultipartForm(32 << 20)
		require.NoError(t, err)

		// 验证图片处理参数
		assert.Equal(t, "true", r.FormValue("compress"))
		assert.Equal(t, "80", r.FormValue("quality"))
		assert.Equal(t, "true", r.FormValue("watermark"))
		assert.Equal(t, "© Test", r.FormValue("watermarkText"))
		assert.Equal(t, "bottom-right", r.FormValue("watermarkPosition"))

		// 返回 LingStorage 格式的成功响应
		response := map[string]interface{}{
			"code": 200,
			"msg":  "File uploaded successfully",
			"data": map[string]interface{}{
				"key":          "images/test.jpg",
				"bucket":       "images",
				"filename":     "test.jpg",
				"size":         int64(len(testContent) - 5), // 模拟压缩后大小
				"originalSize": int64(len(testContent)),
				"compressed":   true,
				"watermarked":  true,
				"url":          "https://example.com/uploads/images/test.jpg",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// 创建客户端
	client := NewClient(&Config{
		BaseURL:   server.URL,
		APIKey:    "test-key",
		APISecret: "test-secret",
	})

	// 测试上传
	result, err := client.UploadFile(&UploadRequest{
		FilePath:          testFile,
		Bucket:            "images",
		Key:               "images/test.jpg",
		Compress:          true,
		Quality:           80,
		Watermark:         true,
		WatermarkText:     "© Test",
		WatermarkPosition: "bottom-right",
	})

	require.NoError(t, err)
	assert.True(t, result.Compressed)
	assert.True(t, result.Watermarked)
	assert.Less(t, result.Size, result.OriginalSize)
}

func TestBatchUpload(t *testing.T) {
	// 创建测试文件
	tempDir := t.TempDir()
	files := []string{}
	for i := 0; i < 3; i++ {
		filename := fmt.Sprintf("test%d.txt", i)
		filepath := filepath.Join(tempDir, filename)
		content := fmt.Sprintf("Content of file %d", i)
		err := os.WriteFile(filepath, []byte(content), 0644)
		require.NoError(t, err)
		files = append(files, filepath)
	}

	uploadCount := 0
	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uploadCount++

		// 验证认证头
		assert.Equal(t, "test-key", r.Header.Get("X-API-Key"))
		assert.Equal(t, "test-secret", r.Header.Get("X-API-Secret"))

		// 解析 multipart form
		err := r.ParseMultipartForm(32 << 20)
		require.NoError(t, err)

		// 获取文件信息
		file, header, err := r.FormFile("file")
		require.NoError(t, err)
		defer file.Close()

		// 返回 LingStorage 格式的成功响应
		response := map[string]interface{}{
			"code": 200,
			"msg":  "File uploaded successfully",
			"data": map[string]interface{}{
				"key":          "batch/" + header.Filename,
				"bucket":       "batch",
				"filename":     header.Filename,
				"size":         int64(20),
				"originalSize": int64(20),
				"compressed":   false,
				"watermarked":  false,
				"url":          "https://example.com/uploads/batch/" + header.Filename,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// 创建客户端
	client := NewClient(&Config{
		BaseURL:   server.URL,
		APIKey:    "test-key",
		APISecret: "test-secret",
	})

	// 测试批量上传
	progressCalls := 0
	result, err := client.BatchUpload(&BatchUploadRequest{
		Files:     files,
		Bucket:    "batch",
		KeyPrefix: "batch",
		OnProgress: func(completed, total int, current string) {
			progressCalls++
			assert.LessOrEqual(t, completed, total)
			assert.Equal(t, 3, total)
		},
	})

	require.NoError(t, err)
	assert.Equal(t, 3, result.Total)
	assert.Len(t, result.Success, 3)
	assert.Len(t, result.Failed, 0)
	assert.Equal(t, 3, uploadCount)
	assert.Greater(t, progressCalls, 0)
}

func TestAPIError(t *testing.T) {
	// 创建测试文件
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	// 创建模拟服务器返回错误
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		response := map[string]interface{}{
			"code": 400,
			"msg":  "File too large",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// 创建客户端
	client := NewClient(&Config{
		BaseURL:   server.URL,
		APIKey:    "test-key",
		APISecret: "test-secret",
	})

	// 测试上传
	_, err = client.UploadFile(&UploadRequest{
		FilePath: testFile,
		Bucket:   "test",
	})

	require.Error(t, err)
	apiErr, ok := err.(*APIError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
}

func TestProgressCallback(t *testing.T) {
	// 创建测试文件
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := strings.Repeat("Hello, World! ", 1000) // 较大的文件以便测试进度
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	require.NoError(t, err)

	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 读取整个请求体以模拟上传过程
		_, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		response := map[string]interface{}{
			"code": 200,
			"msg":  "File uploaded successfully",
			"data": map[string]interface{}{
				"key":          "test.txt",
				"bucket":       "default",
				"filename":     "test.txt",
				"size":         int64(len(testContent)),
				"originalSize": int64(len(testContent)),
				"compressed":   false,
				"watermarked":  false,
				"url":          "https://example.com/uploads/test.txt",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// 创建客户端
	client := NewClient(&Config{
		BaseURL:   server.URL,
		APIKey:    "test-key",
		APISecret: "test-secret",
	})

	// 测试进度回调
	progressCalls := 0
	var lastUploaded, lastTotal int64

	_, err = client.UploadFile(&UploadRequest{
		FilePath: testFile,
		Bucket:   "default",
		OnProgress: func(uploaded, total int64) {
			progressCalls++
			lastUploaded = uploaded
			lastTotal = total
			assert.LessOrEqual(t, uploaded, total)
			assert.Greater(t, total, int64(0))
		},
	})

	require.NoError(t, err)
	assert.Greater(t, progressCalls, 0)
	assert.Equal(t, lastTotal, lastUploaded) // 最后一次调用应该是完整上传
}

func TestRetryMechanism(t *testing.T) {
	// 创建测试文件
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	attempts := 0
	// 创建模拟服务器，前两次返回错误，第三次成功
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"code": 200,
			"msg":  "File uploaded successfully",
			"data": map[string]interface{}{
				"key":          "test.txt",
				"bucket":       "default",
				"filename":     "test.txt",
				"size":         int64(4),
				"originalSize": int64(4),
				"compressed":   false,
				"watermarked":  false,
				"url":          "https://example.com/uploads/test.txt",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// 创建客户端
	client := NewClient(&Config{
		BaseURL:    server.URL,
		APIKey:     "test-key",
		APISecret:  "test-secret",
		RetryCount: 3,
	})

	// 测试重试机制
	result, err := client.UploadFile(&UploadRequest{
		FilePath: testFile,
		Bucket:   "default",
	})

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 3, attempts) // 应该重试了3次
}

func TestDeleteFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/files/test-bucket/test-key" {
			t.Errorf("Expected path /api/public/files/test-bucket/test-key, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "File deleted successfully",
		})
	}))
	defer server.Close()

	client := NewClient(&Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})

	err := client.DeleteFile("test-bucket", "test-key")
	if err != nil {
		t.Errorf("DeleteFile failed: %v", err)
	}
}

func TestGetFileURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/files/test-bucket/test-key/url" {
			t.Errorf("Expected path /api/public/files/test-bucket/test-key/url, got %s", r.URL.Path)
		}

		expires := r.URL.Query().Get("expires")
		if expires != "1h0m0s" {
			t.Errorf("Expected expires=1h0m0s, got %s", expires)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"url": "https://example.com/test-bucket/test-key",
			},
		})
	}))
	defer server.Close()

	client := NewClient(&Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})

	url, err := client.GetFileURL("test-bucket", "test-key", time.Hour)
	if err != nil {
		t.Errorf("GetFileURL failed: %v", err)
	}

	expectedURL := "https://example.com/test-bucket/test-key"
	if url != expectedURL {
		t.Errorf("Expected URL %s, got %s", expectedURL, url)
	}
}

func TestGetFileInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/files/test-bucket/test-key/info" {
			t.Errorf("Expected path /api/public/files/test-bucket/test-key/info, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": FileInfo{
				Key:          "test-key",
				Size:         1024,
				LastModified: time.Now(),
				ETag:         "test-etag",
				ContentType:  "text/plain",
			},
		})
	}))
	defer server.Close()

	client := NewClient(&Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})

	fileInfo, err := client.GetFileInfo("test-bucket", "test-key")
	if err != nil {
		t.Errorf("GetFileInfo failed: %v", err)
	}

	if fileInfo.Key != "test-key" {
		t.Errorf("Expected key test-key, got %s", fileInfo.Key)
	}
	if fileInfo.Size != 1024 {
		t.Errorf("Expected size 1024, got %d", fileInfo.Size)
	}
}

func TestListBuckets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/buckets" {
			t.Errorf("Expected path /api/public/buckets, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"buckets": []string{"bucket1", "bucket2", "bucket3"},
			},
		})
	}))
	defer server.Close()

	client := NewClient(&Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})

	buckets, err := client.ListBuckets("", false)
	if err != nil {
		t.Errorf("ListBuckets failed: %v", err)
	}

	expectedBuckets := []string{"bucket1", "bucket2", "bucket3"}
	if len(buckets) != len(expectedBuckets) {
		t.Errorf("Expected %d buckets, got %d", len(expectedBuckets), len(buckets))
	}

	for i, bucket := range buckets {
		if bucket != expectedBuckets[i] {
			t.Errorf("Expected bucket %s, got %s", expectedBuckets[i], bucket)
		}
	}
}

func TestCreateBucket(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/buckets" {
			t.Errorf("Expected path /api/public/buckets, got %s", r.URL.Path)
		}

		var req CreateBucketRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		if req.BucketName != "test-bucket" {
			t.Errorf("Expected bucket name test-bucket, got %s", req.BucketName)
		}
		if req.Region != "us-east-1" {
			t.Errorf("Expected region us-east-1, got %s", req.Region)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Bucket created successfully",
		})
	}))
	defer server.Close()

	client := NewClient(&Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})

	err := client.CreateBucket(&CreateBucketRequest{
		BucketName: "test-bucket",
		Region:     "us-east-1",
	})
	if err != nil {
		t.Errorf("CreateBucket failed: %v", err)
	}
}

func TestCopyFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/files/src-bucket/src-key/copy" {
			t.Errorf("Expected path /api/public/files/src-bucket/src-key/copy, got %s", r.URL.Path)
		}

		var reqBody map[string]string
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		if reqBody["destBucket"] != "dest-bucket" {
			t.Errorf("Expected destBucket dest-bucket, got %s", reqBody["destBucket"])
		}
		if reqBody["destKey"] != "dest-key" {
			t.Errorf("Expected destKey dest-key, got %s", reqBody["destKey"])
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "File copied successfully",
		})
	}))
	defer server.Close()

	client := NewClient(&Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})

	err := client.CopyFile(&CopyFileRequest{
		SrcBucket:  "src-bucket",
		SrcKey:     "src-key",
		DestBucket: "dest-bucket",
		DestKey:    "dest-key",
	})
	if err != nil {
		t.Errorf("CopyFile failed: %v", err)
	}
}

func TestListFiles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/buckets/test-bucket/files" {
			t.Errorf("Expected path /api/public/buckets/test-bucket/files, got %s", r.URL.Path)
		}

		prefix := r.URL.Query().Get("prefix")
		if prefix != "uploads/" {
			t.Errorf("Expected prefix uploads/, got %s", prefix)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": ListFilesResult{
				Files: []FileInfo{
					{
						Key:          "uploads/file1.txt",
						Size:         1024,
						LastModified: time.Now(),
						ContentType:  "text/plain",
					},
					{
						Key:          "uploads/file2.jpg",
						Size:         2048,
						LastModified: time.Now(),
						ContentType:  "image/jpeg",
					},
				},
				Directories: []string{"uploads/images/", "uploads/docs/"},
				NextMarker:  "",
				IsTruncated: false,
			},
		})
	}))
	defer server.Close()

	client := NewClient(&Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})

	result, err := client.ListFiles(&ListFilesRequest{
		Bucket: "test-bucket",
		Prefix: "uploads/",
		Limit:  10,
	})
	if err != nil {
		t.Errorf("ListFiles failed: %v", err)
	}

	if len(result.Files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(result.Files))
	}

	if result.Files[0].Key != "uploads/file1.txt" {
		t.Errorf("Expected first file key uploads/file1.txt, got %s", result.Files[0].Key)
	}
}
