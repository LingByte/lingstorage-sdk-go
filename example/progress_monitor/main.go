package main

import (
	"fmt"
	"github.com/LingByte/lingstorage-sdk-go"
	"log"
	"os"
	"time"
)

func main() {
	baseURL := os.Getenv("LINGSTORAGE_BASE_URL")
	apiKey := os.Getenv("LINGSTORAGE_API_KEY")
	apiSecret := os.Getenv("LINGSTORAGE_API_SECRET")
	if baseURL == "" {
		baseURL = "http://localhost:7075"
		fmt.Printf("Using default server address: %s\n", baseURL)
	}

	if apiKey == "" || apiSecret == "" {
		fmt.Println("请设置 API 凭据环境变量")
		fmt.Println("export LINGSTORAGE_API_KEY=\"your-api-key\"")
		fmt.Println("export LINGSTORAGE_API_SECRET=\"your-api-secret\"")
		os.Exit(1)
	}

	client := lingstorage.NewClient(&lingstorage.Config{
		BaseURL:   baseURL,
		APIKey:    apiKey,
		APISecret: apiSecret,
	})

	if len(os.Args) < 2 {
		log.Fatal("用法: go run main.go <文件路径>")
	}

	filePath := os.Args[1]

	fmt.Printf("正在上传文件: %s\n", filePath)
	fmt.Printf("监控上传进度...\n\n")
	result, err := client.UploadFile(&lingstorage.UploadRequest{
		FilePath:   filePath,
		Bucket:     "progress-demo",
		OnProgress: lingstorage.NewProgressMonitor().OnProgress,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n\nUpload completed!\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("File information:\n")
	fmt.Printf("Filename: %s\n", result.Filename)
	fmt.Printf("File size: %s\n", lingstorage.FormatDuration(time.Duration(result.Size)))
	fmt.Printf("Bucket: %s\n", result.Bucket)
	fmt.Printf("File key: %s\n", result.Key)
	if result.URL != "" {
		fmt.Printf("\nAccess URL: %s\n", result.URL)
	}
}
