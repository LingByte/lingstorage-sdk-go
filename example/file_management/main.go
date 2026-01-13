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
		BaseURL:   "http://localhost:8080",
		APIKey:    "your-api-key",
		APISecret: "your-api-secret",
		Timeout:   30 * time.Second,
	})

	// 测试连接
	if err := client.Ping(); err != nil {
		log.Fatalf("Failed to ping server: %v", err)
	}
	fmt.Println("✓ Server connection successful")

	// 1. 列举存储桶
	fmt.Println("\n=== 列举存储桶 ===")
	buckets, err := client.ListBuckets("", false)
	if err != nil {
		log.Printf("Failed to list buckets: %v", err)
	} else {
		fmt.Printf("Found %d buckets: %v\n", len(buckets), buckets)
	}

	// 2. 创建存储桶
	fmt.Println("\n=== 创建存储桶 ===")
	err = client.CreateBucket(&lingstorage.CreateBucketRequest{
		BucketName: "test-bucket",
		Region:     "us-east-1",
	})
	if err != nil {
		log.Printf("Failed to create bucket: %v", err)
	} else {
		fmt.Println("✓ Bucket created successfully")
	}

	// 3. 上传文件（假设有一个测试文件）
	fmt.Println("\n=== 上传文件 ===")
	uploadResult, err := client.UploadFile(&lingstorage.UploadRequest{
		FilePath: "test.txt", // 需要存在的测试文件
		Bucket:   "test-bucket",
		Key:      "uploads/test.txt",
		OnProgress: func(uploaded, total int64) {
			fmt.Printf("Upload progress: %d/%d (%.2f%%)\n", uploaded, total, float64(uploaded)/float64(total)*100)
		},
	})
	if err != nil {
		log.Printf("Failed to upload file: %v", err)
	} else {
		fmt.Printf("✓ File uploaded: %+v\n", uploadResult)
	}

	// 4. 列举文件
	fmt.Println("\n=== 列举文件 ===")
	filesResult, err := client.ListFiles(&lingstorage.ListFilesRequest{
		Bucket: "test-bucket",
		Prefix: "uploads/",
		Limit:  10,
	})
	if err != nil {
		log.Printf("Failed to list files: %v", err)
	} else {
		fmt.Printf("Found %d files:\n", len(filesResult.Files))
		for _, file := range filesResult.Files {
			fmt.Printf("  - %s (size: %d, modified: %v)\n", file.Key, file.Size, file.LastModified)
		}
	}

	// 5. 获取文件信息
	fmt.Println("\n=== 获取文件信息 ===")
	fileInfo, err := client.GetFileInfo("test-bucket", "uploads/test.txt")
	if err != nil {
		log.Printf("Failed to get file info: %v", err)
	} else {
		fmt.Printf("File info: %+v\n", fileInfo)
	}

	// 6. 获取文件URL
	fmt.Println("\n=== 获取文件URL ===")
	fileURL, err := client.GetFileURL("test-bucket", "uploads/test.txt", time.Hour)
	if err != nil {
		log.Printf("Failed to get file URL: %v", err)
	} else {
		fmt.Printf("File URL: %s\n", fileURL)
	}

	// 7. 复制文件
	fmt.Println("\n=== 复制文件 ===")
	err = client.CopyFile(&lingstorage.CopyFileRequest{
		SrcBucket:  "test-bucket",
		SrcKey:     "uploads/test.txt",
		DestBucket: "test-bucket",
		DestKey:    "backups/test-copy.txt",
	})
	if err != nil {
		log.Printf("Failed to copy file: %v", err)
	} else {
		fmt.Println("✓ File copied successfully")
	}

	// 8. 移动文件
	fmt.Println("\n=== 移动文件 ===")
	err = client.MoveFile(&lingstorage.MoveFileRequest{
		SrcBucket:  "test-bucket",
		SrcKey:     "backups/test-copy.txt",
		DestBucket: "test-bucket",
		DestKey:    "archive/test-moved.txt",
	})
	if err != nil {
		log.Printf("Failed to move file: %v", err)
	} else {
		fmt.Println("✓ File moved successfully")
	}

	// 9. 获取存储桶域名
	fmt.Println("\n=== 获取存储桶域名 ===")
	domains, err := client.GetBucketDomains("test-bucket")
	if err != nil {
		log.Printf("Failed to get bucket domains: %v", err)
	} else {
		fmt.Printf("Bucket domains: %v\n", domains)
	}

	// 10. 设置存储桶权限
	fmt.Println("\n=== 设置存储桶权限 ===")
	err = client.SetBucketPrivate(&lingstorage.SetBucketPrivateRequest{
		BucketName: "test-bucket",
		IsPrivate:  true,
	})
	if err != nil {
		log.Printf("Failed to set bucket private: %v", err)
	} else {
		fmt.Println("✓ Bucket set to private successfully")
	}

	// 11. 删除文件
	fmt.Println("\n=== 删除文件 ===")
	err = client.DeleteFile("test-bucket", "uploads/test.txt")
	if err != nil {
		log.Printf("Failed to delete file: %v", err)
	} else {
		fmt.Println("✓ File deleted successfully")
	}

	// 12. 删除存储桶
	fmt.Println("\n=== 删除存储桶 ===")
	err = client.DeleteBucket("test-bucket")
	if err != nil {
		log.Printf("Failed to delete bucket: %v", err)
	} else {
		fmt.Println("✓ Bucket deleted successfully")
	}

	fmt.Println("\n=== 所有操作完成 ===")
}
