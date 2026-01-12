package main

import (
	"fmt"
	"log"
	"os"

	"github.com/LingByte/lingstorage-sdk"
)

func main() {
	baseURL := os.Getenv("LINGSTORAGE_BASE_URL")
	apiKey := os.Getenv("LINGSTORAGE_API_KEY")
	apiSecret := os.Getenv("LINGSTORAGE_API_SECRET")
	if baseURL == "" {
		baseURL = "http://localhost:7075"
		fmt.Printf("使用默认服务器地址: %s\n", baseURL)
	}

	if apiKey == "" || apiSecret == "" {
		fmt.Println("请设置 API 凭据环境变量")
		fmt.Println("export LINGSTORAGE_API_KEY=\"your-api-key\"")
		fmt.Println("export LINGSTORAGE_API_SECRET=\"your-api-secret\"")
		os.Exit(1)
	}

	// 创建客户端
	client := lingstorage.NewClient(&lingstorage.Config{
		BaseURL:   baseURL,
		APIKey:    apiKey,
		APISecret: apiSecret,
	})

	// Check command line arguments
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <file path>")
	}

	filePath := os.Args[1]

	// 上传文件
	fmt.Printf("正在上传文件: %s\n", filePath)

	result, err := client.UploadFile(&lingstorage.UploadRequest{
		FilePath: filePath,
		Bucket:   "cetide",
		OnProgress: func(uploaded, total int64) {
			percentage := float64(uploaded) / float64(total) * 100
			fmt.Printf("\r上传进度: %.1f%% (%d/%d bytes)", percentage, uploaded, total)
		},
	})

	if err != nil {
		log.Fatalf("上传失败: %v", err)
	}

	fmt.Printf("\n\nUpload successful!\n")
	fmt.Printf("File key: %s\n", result.Key)
	fmt.Printf("Bucket: %s\n", result.Bucket)
	fmt.Printf("Filename: %s\n", result.Filename)
	fmt.Printf("File size: %d bytes\n", result.Size)
	if result.URL != "" {
		fmt.Printf("Access URL: %s\n", result.URL)
	}
}
