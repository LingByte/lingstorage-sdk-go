# LingStorage SDK

LingStorage SDK is a Go language client library for interacting with LingStorage services. It provides easy-to-use APIs for uploading files, managing buckets, and other functions.

## Features

- File upload (with image compression and watermark support)
- Multiple storage backend support (local, Qiniu Cloud, Alibaba Cloud OSS, AWS S3, etc.)
- API Key authentication
- File type and size restrictions
- Batch upload
- Upload progress callbacks
- Error retry mechanism
- Comprehensive test coverage

## 快速开始

### 安装

```bash
go get github.com/LingByte/lingstorage-sdk
```

### 基本使用

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/LingByte/lingstorage-sdk"
)

func main() {
    // 创建客户端
    client := lingstorage.NewClient(&lingstorage.Config{
        BaseURL:   "https://your-lingstorage-server.com",
        APIKey:    "your-api-key",
        APISecret: "your-api-secret",
    })
    
    // 上传文件
    result, err := client.UploadFile(&lingstorage.UploadRequest{
        FilePath: "./example.jpg",
        Bucket:   "default",
        Key:      "uploads/example.jpg",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("文件上传成功: %s\n", result.URL)
}
```

### 高级功能

#### 图片压缩和水印

```go
result, err := client.UploadFile(&lingstorage.UploadRequest{
    FilePath: "./photo.jpg",
    Bucket:   "images",
    
    // 图片压缩
    Compress: true,
    Quality:  80,
    
    // 添加水印
    Watermark: true,
    WatermarkText: "© 2024 My Company",
    WatermarkPosition: "bottom-right",
})
```

#### 批量上传

```go
files := []string{"file1.jpg", "file2.png", "file3.pdf"}
results, err := client.BatchUpload(&lingstorage.BatchUploadRequest{
    Files:  files,
    Bucket: "documents",
    
    // 上传进度回调
    OnProgress: func(completed, total int, current string) {
        fmt.Printf("进度: %d/%d - 当前文件: %s\n", completed, total, current)
    },
})
```

#### 从内存上传

```go
data := []byte("Hello, World!")
result, err := client.UploadBytes(&lingstorage.UploadBytesRequest{
    Data:     data,
    Filename: "hello.txt",
    Bucket:   "text-files",
})
```

## API 文档

### 客户端配置

```go
type Config struct {
    BaseURL   string        // LingStorage 服务器地址
    APIKey    string        // API 密钥
    APISecret string        // API 密钥对应的 Secret
    Timeout   time.Duration // 请求超时时间（默认30秒）
    RetryCount int          // 重试次数（默认3次）
    UserAgent string        // 用户代理（可选）
}
```

### Upload Request

```go
type UploadRequest struct {
    FilePath          string   // Local file path
    Bucket            string   // Bucket name
    Key               string   // File key name (optional, auto-generated)
    AllowedTypes      []string // Allowed file types (optional)
    
    // Image processing options
    Compress          bool     // Whether to compress the image
    Quality           int      // Compression quality 1-100
    Watermark         bool     // Whether to add watermark
    WatermarkText     string   // Watermark text
    WatermarkPosition string   // Watermark position
    
    // Callback function
    OnProgress        func(uploaded, total int64) // Upload progress callback
}
```

### 上传响应

```go
type UploadResult struct {
    Key          string `json:"key"`
    Bucket       string `json:"bucket"`
    Filename     string `json:"filename"`
    Size         int64  `json:"size"`
    OriginalSize int64  `json:"originalSize"`
    Compressed   bool   `json:"compressed"`
    Watermarked  bool   `json:"watermarked"`
    URL          string `json:"url"`
}
```

## 错误处理

SDK 提供了详细的错误信息：

```go
result, err := client.UploadFile(req)
if err != nil {
    if apiErr, ok := err.(*lingstorage.APIError); ok {
        fmt.Printf("API 错误: %s (状态码: %d)\n", apiErr.Message, apiErr.StatusCode)
    } else {
        fmt.Printf("其他错误: %s\n", err.Error())
    }
}
```

## Examples

Check the `examples/` directory for more usage examples:

- [Basic Upload](examples/basic_upload/main.go)
- [Batch Upload](examples/batch_upload/main.go)
- [Image Processing](examples/image_processing/main.go)
- [Progress Monitoring](examples/progress_monitoring/main.go)

## 测试

```bash
# 运行所有测试
go test ./...

# 运行测试并显示覆盖率
go test -cover ./...

# 生成测试报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License