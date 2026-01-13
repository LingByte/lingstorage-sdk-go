# LingStorage SDK

LingStorage SDK is a Go language client library for interacting with LingStorage services. It provides easy-to-use APIs for uploading files, managing buckets, and other functions.

## Features

- **文件操作**
  - 文件上传（支持图片压缩和水印）
  - 文件删除
  - 文件复制和移动
  - 获取文件信息和访问URL
  - 批量上传
- **存储桶管理**
  - 创建和删除存储桶
  - 列举存储桶和文件
  - 设置存储桶权限
  - 获取存储桶域名
- **其他功能**
  - 多种存储后端支持（本地、七牛云、阿里云OSS、AWS S3等）
  - API Key 认证
  - 文件类型和大小限制
  - 上传进度回调
  - 错误重试机制
  - 全面的测试覆盖

## 快速开始

### 安装

```bash
go get github.com/LingByte/lingstorage-sdk-go
```

### 基本使用

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    lingstorage "github.com/LingByte/lingstorage-sdk-go"
)

func main() {
    // 创建客户端
    client := lingstorage.NewClient(&lingstorage.Config{
        BaseURL:   "https://your-lingstorage-server.com",
        APIKey:    "your-api-key",
        APISecret: "your-api-secret",
    })
    
    // 测试连接
    if err := client.Ping(); err != nil {
        log.Fatal("服务器连接失败:", err)
    }
    
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

## API 文档

### 文件操作

#### 上传文件

```go
// 基本上传
result, err := client.UploadFile(&lingstorage.UploadRequest{
    FilePath: "./photo.jpg",
    Bucket:   "images",
    Key:      "uploads/photo.jpg",
})

// 图片压缩和水印
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
    
    // 进度回调
    OnProgress: func(uploaded, total int64) {
        fmt.Printf("上传进度: %.2f%%\n", float64(uploaded)/float64(total)*100)
    },
})
```

#### 删除文件

```go
err := client.DeleteFile("bucket-name", "file-key")
if err != nil {
    log.Printf("删除文件失败: %v", err)
}
```

#### 获取文件信息

```go
fileInfo, err := client.GetFileInfo("bucket-name", "file-key")
if err != nil {
    log.Printf("获取文件信息失败: %v", err)
} else {
    fmt.Printf("文件大小: %d bytes\n", fileInfo.Size)
    fmt.Printf("最后修改时间: %v\n", fileInfo.LastModified)
}
```

#### 获取文件访问URL

```go
// 获取1小时有效期的访问URL
fileURL, err := client.GetFileURL("bucket-name", "file-key", time.Hour)
if err != nil {
    log.Printf("获取文件URL失败: %v", err)
} else {
    fmt.Printf("文件URL: %s\n", fileURL)
}
```

#### 复制文件

```go
err := client.CopyFile(&lingstorage.CopyFileRequest{
    SrcBucket:  "source-bucket",
    SrcKey:     "source/file.jpg",
    DestBucket: "dest-bucket",
    DestKey:    "backup/file.jpg",
})
```

#### 移动文件

```go
err := client.MoveFile(&lingstorage.MoveFileRequest{
    SrcBucket:  "source-bucket",
    SrcKey:     "temp/file.jpg",
    DestBucket: "dest-bucket",
    DestKey:    "final/file.jpg",
})
```

### 存储桶管理

#### 列举存储桶

```go
buckets, err := client.ListBuckets("", false)
if err != nil {
    log.Printf("列举存储桶失败: %v", err)
} else {
    fmt.Printf("找到 %d 个存储桶: %v\n", len(buckets), buckets)
}
```

#### 创建存储桶

```go
err := client.CreateBucket(&lingstorage.CreateBucketRequest{
    BucketName: "my-new-bucket",
    Region:     "us-east-1",
})
```

#### 删除存储桶

```go
err := client.DeleteBucket("bucket-to-delete")
```

#### 列举文件

```go
result, err := client.ListFiles(&lingstorage.ListFilesRequest{
    Bucket:    "my-bucket",
    Prefix:    "uploads/",
    Limit:     100,
    Delimiter: "/",
})
if err != nil {
    log.Printf("列举文件失败: %v", err)
} else {
    fmt.Printf("找到 %d 个文件\n", len(result.Files))
    for _, file := range result.Files {
        fmt.Printf("  - %s (%d bytes)\n", file.Key, file.Size)
    }
}
```

#### 获取存储桶域名

```go
domains, err := client.GetBucketDomains("my-bucket")
if err != nil {
    log.Printf("获取域名失败: %v", err)
} else {
    fmt.Printf("存储桶域名: %v\n", domains)
}
```

#### 设置存储桶权限

```go
err := client.SetBucketPrivate(&lingstorage.SetBucketPrivateRequest{
    BucketName: "my-bucket",
    IsPrivate:  true,
})
```

### 高级功能

#### 批量上传

```go
files := []string{"file1.jpg", "file2.png", "file3.pdf"}
results, err := client.BatchUpload(&lingstorage.BatchUploadRequest{
    Files:  files,
    Bucket: "documents",
    
    // 批量上传进度回调
    OnProgress: func(completed, total int, current string) {
        fmt.Printf("批量上传进度: %d/%d - 当前文件: %s\n", completed, total, current)
    },
    
    // 单个文件上传进度回调
    OnFileProgress: func(uploaded, total int64) {
        fmt.Printf("文件上传进度: %.2f%%\n", float64(uploaded)/float64(total)*100)
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
    Key:      "greetings/hello.txt",
})
```

#### 从 io.Reader 上传

```go
file, err := os.Open("large-file.zip")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

result, err := client.UploadFromReader(&lingstorage.UploadFromReaderRequest{
    Reader:   file,
    Filename: "large-file.zip",
    Size:     fileSize, // 如果已知文件大小
    Bucket:   "archives",
    Key:      "uploads/large-file.zip",
})
```

## 数据结构

### 客户端配置

```go
type Config struct {
    BaseURL    string        // LingStorage 服务器地址
    APIKey     string        // API 密钥
    APISecret  string        // API 密钥对应的 Secret
    Timeout    time.Duration // 请求超时时间（默认30秒）
    RetryCount int           // 重试次数（默认3次）
    UserAgent  string        // 用户代理（可选）
}
```

### 上传请求

```go
type UploadRequest struct {
    FilePath          string   // 本地文件路径
    Bucket            string   // 存储桶名称
    Key               string   // 文件键名（可选，自动生成）
    AllowedTypes      []string // 允许的文件类型（可选）
    
    // 图片处理选项
    Compress          bool     // 是否压缩图片
    Quality           int      // 压缩质量 1-100
    Watermark         bool     // 是否添加水印
    WatermarkText     string   // 水印文本
    WatermarkPosition string   // 水印位置
    
    // 回调函数
    OnProgress        func(uploaded, total int64) // 上传进度回调
}
```

### 上传响应

```go
type UploadResult struct {
    Key          string `json:"key"`          // 文件键名
    Bucket       string `json:"bucket"`       // 存储桶名称
    Filename     string `json:"filename"`     // 原始文件名
    Size         int64  `json:"size"`         // 文件大小
    OriginalSize int64  `json:"originalSize"` // 原始文件大小
    Compressed   bool   `json:"compressed"`   // 是否已压缩
    Watermarked  bool   `json:"watermarked"`  // 是否已添加水印
    URL          string `json:"url"`          // 访问URL
}
```

### 文件信息

```go
type FileInfo struct {
    Key          string    `json:"key"`          // 文件键名
    Size         int64     `json:"size"`         // 文件大小
    LastModified time.Time `json:"lastModified"` // 最后修改时间
    ETag         string    `json:"etag"`         // ETag
    ContentType  string    `json:"contentType"`  // 内容类型
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

查看 `examples/` 目录获取更多使用示例：

- [基本上传](examples/basic_upload/main.go)
- [批量上传](examples/batch_upload/main.go)
- [图片处理](examples/image_processing/main.go)
- [进度监控](examples/progress_monitoring/main.go)
- [文件管理](examples/file_management/main.go) - **新增**

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

## 更新日志

### v1.1.0 (最新)
- ✨ 新增文件删除功能
- ✨ 新增文件信息获取功能
- ✨ 新增文件URL获取功能
- ✨ 新增文件复制和移动功能
- ✨ 新增存储桶管理功能（创建、删除、列举）
- ✨ 新增文件列举功能
- ✨ 新增存储桶域名获取功能
- ✨ 新增存储桶权限设置功能
- 🔧 优化错误处理和重试机制
- 📚 完善文档和示例

### v1.0.0
- 🎉 初始版本
- ✨ 文件上传功能
- ✨ 图片压缩和水印支持
- ✨ 批量上传功能
- ✨ 进度回调支持

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License