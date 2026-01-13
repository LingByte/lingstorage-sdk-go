package lingstorage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/LingByte/lingstorage-sdk-go/constants"
)

// Client LingStorage SDK Client
type Client struct {
	config     *Config
	httpClient *http.Client
}

// Config LingStorage client config
type Config struct {
	BaseURL    string        // LingStorage server address
	APIKey     string        // API Key
	APISecret  string        // API Secret
	Timeout    time.Duration // Request Timeout
	RetryCount int           // retry times
	UserAgent  string        // user agent
}

// NewClient create new lingStorage client
func NewClient(config *Config) *Client {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.RetryCount == 0 {
		config.RetryCount = 3
	}
	if config.UserAgent == "" {
		config.UserAgent = constants.DEFAULT_USER_AGENT
	}

	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// UploadRequest upload request
type UploadRequest struct {
	FilePath          string                      // file path
	Bucket            string                      // bucket name
	Key               string                      // file key name
	AllowedTypes      []string                    // all file types
	Compress          bool                        // if compress file
	Quality           int                         // quality 1-100 - default 100
	Watermark         bool                        // if watermark
	WatermarkText     string                      // watermark text
	WatermarkPosition string                      // watermark position
	OnProgress        func(uploaded, total int64) // upload progress callback
}

// UploadBytesRequest upload request from  bytes
type UploadBytesRequest struct {
	Data              []byte                      // file data
	Filename          string                      // file name
	Bucket            string                      // bucket name
	Key               string                      // file key
	AllowedTypes      []string                    // all types
	Compress          bool                        // if compress file
	Quality           int                         // quality 1-100 - default 100
	Watermark         bool                        // if watermark
	WatermarkText     string                      // watermark text
	WatermarkPosition string                      // watermark position
	OnProgress        func(uploaded, total int64) // upload progress callback
}

// BatchUploadRequest batch upload request
type BatchUploadRequest struct {
	Files             []string                                   // file list
	Bucket            string                                     // bucket name
	KeyPrefix         string                                     // key prefix
	AllowedTypes      []string                                   // all types
	Compress          bool                                       // if compress file
	Quality           int                                        // quality 1-100 - default 100
	Watermark         bool                                       // if watermark
	WatermarkText     string                                     // watermark text
	WatermarkPosition string                                     // watermark position
	OnProgress        func(completed, total int, current string) // batch upload progress callback
	OnFileProgress    func(uploaded, total int64)                // signal file upload progress
}

// UploadFromReaderRequest read from io.Reader
type UploadFromReaderRequest struct {
	Reader            io.Reader
	Filename          string
	Size              int64
	Bucket            string
	Key               string
	AllowedTypes      []string
	Compress          bool
	Quality           int
	Watermark         bool
	WatermarkText     string
	WatermarkPosition string
	OnProgress        func(uploaded, total int64)
}

// FileInfo 文件信息
type FileInfo struct {
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"lastModified"`
	ETag         string    `json:"etag"`
	ContentType  string    `json:"contentType"`
}

// ListFilesRequest 列举文件请求
type ListFilesRequest struct {
	Bucket    string `json:"bucket"`
	Prefix    string `json:"prefix"`
	Marker    string `json:"marker"`
	Delimiter string `json:"delimiter"`
	Limit     int    `json:"limit"`
}

// ListFilesResult 列举文件结果
type ListFilesResult struct {
	Files       []FileInfo `json:"files"`
	Directories []string   `json:"directories"`
	NextMarker  string     `json:"nextMarker"`
	IsTruncated bool       `json:"isTruncated"`
}

// CreateBucketRequest 创建存储桶请求
type CreateBucketRequest struct {
	BucketName string `json:"bucketName"`
	Region     string `json:"region"`
}

// CopyFileRequest 复制文件请求
type CopyFileRequest struct {
	SrcBucket  string `json:"srcBucket"`
	SrcKey     string `json:"srcKey"`
	DestBucket string `json:"destBucket"`
	DestKey    string `json:"destKey"`
}

// MoveFileRequest 移动文件请求
type MoveFileRequest struct {
	SrcBucket  string `json:"srcBucket"`
	SrcKey     string `json:"srcKey"`
	DestBucket string `json:"destBucket"`
	DestKey    string `json:"destKey"`
}

// SetBucketPrivateRequest 设置存储桶权限请求
type SetBucketPrivateRequest struct {
	BucketName string `json:"bucketName"`
	IsPrivate  bool   `json:"isPrivate"`
}

// UploadResult upload result
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

// UploadError upload error
type UploadError struct {
	File  string `json:"file"`
	Error string `json:"error"`
}

// BatchUploadResult batch upload result
type BatchUploadResult struct {
	Success []UploadResult `json:"success"`
	Failed  []UploadError  `json:"failed"`
	Total   int            `json:"total"`
}

// APIError API Error
type APIError struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Details    string `json:"details"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("ling storage api error %d: %s", e.StatusCode, e.Message)
}

// UploadFile upload single files
func (c *Client) UploadFile(req *UploadRequest) (*UploadResult, error) {
	file, err := os.Open(req.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	var reader io.Reader = file
	if req.OnProgress != nil {
		reader = &progressReader{
			reader:   file,
			total:    fileInfo.Size(),
			callback: req.OnProgress,
		}
	}

	return c.uploadReader(reader, filepath.Base(req.FilePath), fileInfo.Size(), req)
}

// UploadFromReader upload from io.Reader
func (c *Client) UploadFromReader(req *UploadFromReaderRequest) (*UploadResult, error) {
	size := req.Size
	if size <= 0 {
		if seeker, ok := req.Reader.(io.Seeker); ok {
			currentPos, _ := seeker.Seek(0, io.SeekCurrent)
			size, _ = seeker.Seek(0, io.SeekEnd)
			seeker.Seek(currentPos, io.SeekStart)
		}
	}
	var reader io.Reader = req.Reader
	if req.OnProgress != nil && size > 0 {
		reader = &progressReader{
			reader:   req.Reader,
			total:    size,
			callback: req.OnProgress,
		}
	}
	uploadReq := &UploadRequest{
		Bucket:            req.Bucket,
		Key:               req.Key,
		AllowedTypes:      req.AllowedTypes,
		Compress:          req.Compress,
		Quality:           req.Quality,
		Watermark:         req.Watermark,
		WatermarkText:     req.WatermarkText,
		WatermarkPosition: req.WatermarkPosition,
	}

	return c.uploadReader(reader, req.Filename, size, uploadReq)
}

// UploadBytes upload file from bytes
func (c *Client) UploadBytes(req *UploadBytesRequest) (*UploadResult, error) {
	reader := bytes.NewReader(req.Data)
	var readerWithProgress io.Reader = reader
	if req.OnProgress != nil {
		readerWithProgress = &progressReader{
			reader:   reader,
			total:    int64(len(req.Data)),
			callback: req.OnProgress,
		}
	}

	uploadReq := &UploadRequest{
		Bucket:            req.Bucket,
		Key:               req.Key,
		AllowedTypes:      req.AllowedTypes,
		Compress:          req.Compress,
		Quality:           req.Quality,
		Watermark:         req.Watermark,
		WatermarkText:     req.WatermarkText,
		WatermarkPosition: req.WatermarkPosition,
	}

	return c.uploadReader(readerWithProgress, req.Filename, int64(len(req.Data)), uploadReq)
}

// BatchUpload batch upload files
func (c *Client) BatchUpload(req *BatchUploadRequest) (*BatchUploadResult, error) {
	result := &BatchUploadResult{
		Success: make([]UploadResult, 0),
		Failed:  make([]UploadError, 0),
		Total:   len(req.Files),
	}

	for i, filePath := range req.Files {
		if req.OnProgress != nil {
			req.OnProgress(i, len(req.Files), filePath)
		}
		uploadReq := &UploadRequest{
			FilePath:          filePath,
			Bucket:            req.Bucket,
			AllowedTypes:      req.AllowedTypes,
			Compress:          req.Compress,
			Quality:           req.Quality,
			Watermark:         req.Watermark,
			WatermarkText:     req.WatermarkText,
			WatermarkPosition: req.WatermarkPosition,
			OnProgress:        req.OnFileProgress,
		}
		if req.KeyPrefix != "" {
			filename := filepath.Base(filePath)
			uploadReq.Key = req.KeyPrefix + "/" + filename
		}
		uploadResult, err := c.UploadFile(uploadReq)
		if err != nil {
			result.Failed = append(result.Failed, UploadError{
				File:  filePath,
				Error: err.Error(),
			})
		} else {
			result.Success = append(result.Success, *uploadResult)
		}
	}
	if req.OnProgress != nil {
		req.OnProgress(len(req.Files), len(req.Files), "")
	}

	return result, nil
}

// Ping check server if is alive
func (c *Client) Ping() error {
	url := strings.TrimRight(c.config.BaseURL, "/")

	httpReq, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create ping request: %w", err)
	}

	httpReq.Header.Set(constants.USER_AGENT, c.config.UserAgent)
	if c.config.APIKey != "" {
		httpReq.Header.Set(constants.XAPIKEY, c.config.APIKey)
	}
	if c.config.APISecret != "" {
		httpReq.Header.Set(constants.XAPISECRET, c.config.APISecret)
	}

	var resp *http.Response
	var lastErr error
	for i := 0; i <= c.config.RetryCount; i++ {
		resp, lastErr = c.httpClient.Do(httpReq)
		if lastErr == nil && resp.StatusCode < 500 {
			break
		}
		if i < c.config.RetryCount {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	if lastErr != nil {
		return fmt.Errorf("ping request failed after %d retries: %w", c.config.RetryCount, lastErr)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("ping failed with status code: %d", resp.StatusCode)
	}

	return nil
}

// DeleteFile 删除文件
func (c *Client) DeleteFile(bucket, key string) error {
	url := fmt.Sprintf("%s/api/public/files/%s/%s", strings.TrimRight(c.config.BaseURL, "/"), bucket, key)

	httpReq, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set(constants.USER_AGENT, c.config.UserAgent)
	if c.config.APIKey != "" {
		httpReq.Header.Set(constants.XAPIKEY, c.config.APIKey)
	}
	if c.config.APISecret != "" {
		httpReq.Header.Set(constants.XAPISECRET, c.config.APISecret)
	}

	resp, err := c.doRequestWithRetry(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp)
	}

	return nil
}

// GetFileURL 获取文件访问URL
func (c *Client) GetFileURL(bucket, key string, expires time.Duration) (string, error) {
	url := fmt.Sprintf("%s/api/public/files/%s/%s/url", strings.TrimRight(c.config.BaseURL, "/"), bucket, key)

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// 添加过期时间参数
	if expires > 0 {
		q := httpReq.URL.Query()
		q.Set("expires", expires.String())
		httpReq.URL.RawQuery = q.Encode()
	}

	httpReq.Header.Set(constants.USER_AGENT, c.config.UserAgent)
	if c.config.APIKey != "" {
		httpReq.Header.Set(constants.XAPIKEY, c.config.APIKey)
	}
	if c.config.APISecret != "" {
		httpReq.Header.Set(constants.XAPISECRET, c.config.APISecret)
	}

	resp, err := c.doRequestWithRetry(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", c.handleErrorResponse(resp)
	}

	var apiResp struct {
		Success bool `json:"success"`
		Data    struct {
			URL string `json:"url"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return apiResp.Data.URL, nil
}

// GetFileInfo 获取文件信息
func (c *Client) GetFileInfo(bucket, key string) (*FileInfo, error) {
	url := fmt.Sprintf("%s/api/public/files/%s/%s/info", strings.TrimRight(c.config.BaseURL, "/"), bucket, key)

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set(constants.USER_AGENT, c.config.UserAgent)
	if c.config.APIKey != "" {
		httpReq.Header.Set(constants.XAPIKEY, c.config.APIKey)
	}
	if c.config.APISecret != "" {
		httpReq.Header.Set(constants.XAPISECRET, c.config.APISecret)
	}

	resp, err := c.doRequestWithRetry(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var apiResp struct {
		Success bool     `json:"success"`
		Data    FileInfo `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &apiResp.Data, nil
}

// ListFiles 列举文件
func (c *Client) ListFiles(req *ListFilesRequest) (*ListFilesResult, error) {
	url := fmt.Sprintf("%s/api/public/buckets/%s/files", strings.TrimRight(c.config.BaseURL, "/"), req.Bucket)

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 添加查询参数
	q := httpReq.URL.Query()
	if req.Prefix != "" {
		q.Set("prefix", req.Prefix)
	}
	if req.Marker != "" {
		q.Set("marker", req.Marker)
	}
	if req.Delimiter != "" {
		q.Set("delimiter", req.Delimiter)
	}
	if req.Limit > 0 {
		q.Set("limit", strconv.Itoa(req.Limit))
	}
	httpReq.URL.RawQuery = q.Encode()

	httpReq.Header.Set(constants.USER_AGENT, c.config.UserAgent)
	if c.config.APIKey != "" {
		httpReq.Header.Set(constants.XAPIKEY, c.config.APIKey)
	}
	if c.config.APISecret != "" {
		httpReq.Header.Set(constants.XAPISECRET, c.config.APISecret)
	}

	resp, err := c.doRequestWithRetry(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var apiResp struct {
		Success bool            `json:"success"`
		Data    ListFilesResult `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &apiResp.Data, nil
}

// ListBuckets 列举存储桶
func (c *Client) ListBuckets(tagCondition string, shared bool) ([]string, error) {
	url := fmt.Sprintf("%s/api/public/buckets", strings.TrimRight(c.config.BaseURL, "/"))

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 添加查询参数
	q := httpReq.URL.Query()
	if tagCondition != "" {
		q.Set("tagCondition", tagCondition)
	}
	if shared {
		q.Set("shared", "true")
	}
	httpReq.URL.RawQuery = q.Encode()

	httpReq.Header.Set(constants.USER_AGENT, c.config.UserAgent)
	if c.config.APIKey != "" {
		httpReq.Header.Set(constants.XAPIKEY, c.config.APIKey)
	}
	if c.config.APISecret != "" {
		httpReq.Header.Set(constants.XAPISECRET, c.config.APISecret)
	}

	resp, err := c.doRequestWithRetry(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var apiResp struct {
		Success bool `json:"success"`
		Data    struct {
			Buckets []string `json:"buckets"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return apiResp.Data.Buckets, nil
}

// CreateBucket 创建存储桶
func (c *Client) CreateBucket(req *CreateBucketRequest) error {
	url := fmt.Sprintf("%s/api/public/buckets", strings.TrimRight(c.config.BaseURL, "/"))

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set(constants.CONETENT_TYPE, "application/json")
	httpReq.Header.Set(constants.USER_AGENT, c.config.UserAgent)
	if c.config.APIKey != "" {
		httpReq.Header.Set(constants.XAPIKEY, c.config.APIKey)
	}
	if c.config.APISecret != "" {
		httpReq.Header.Set(constants.XAPISECRET, c.config.APISecret)
	}

	resp, err := c.doRequestWithRetry(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp)
	}

	return nil
}

// DeleteBucket 删除存储桶
func (c *Client) DeleteBucket(bucketName string) error {
	url := fmt.Sprintf("%s/api/public/buckets/%s", strings.TrimRight(c.config.BaseURL, "/"), bucketName)

	httpReq, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set(constants.USER_AGENT, c.config.UserAgent)
	if c.config.APIKey != "" {
		httpReq.Header.Set(constants.XAPIKEY, c.config.APIKey)
	}
	if c.config.APISecret != "" {
		httpReq.Header.Set(constants.XAPISECRET, c.config.APISecret)
	}

	resp, err := c.doRequestWithRetry(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp)
	}

	return nil
}

// GetBucketDomains 获取存储桶域名
func (c *Client) GetBucketDomains(bucketName string) ([]string, error) {
	url := fmt.Sprintf("%s/api/public/buckets/%s/domains", strings.TrimRight(c.config.BaseURL, "/"), bucketName)

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set(constants.USER_AGENT, c.config.UserAgent)
	if c.config.APIKey != "" {
		httpReq.Header.Set(constants.XAPIKEY, c.config.APIKey)
	}
	if c.config.APISecret != "" {
		httpReq.Header.Set(constants.XAPISECRET, c.config.APISecret)
	}

	resp, err := c.doRequestWithRetry(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var apiResp struct {
		Success bool `json:"success"`
		Data    struct {
			Domains []string `json:"domains"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return apiResp.Data.Domains, nil
}

// SetBucketPrivate 设置存储桶权限
func (c *Client) SetBucketPrivate(req *SetBucketPrivateRequest) error {
	url := fmt.Sprintf("%s/api/public/buckets/%s/private", strings.TrimRight(c.config.BaseURL, "/"), req.BucketName)

	jsonData, err := json.Marshal(map[string]bool{"isPrivate": req.IsPrivate})
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set(constants.CONETENT_TYPE, "application/json")
	httpReq.Header.Set(constants.USER_AGENT, c.config.UserAgent)
	if c.config.APIKey != "" {
		httpReq.Header.Set(constants.XAPIKEY, c.config.APIKey)
	}
	if c.config.APISecret != "" {
		httpReq.Header.Set(constants.XAPISECRET, c.config.APISecret)
	}

	resp, err := c.doRequestWithRetry(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp)
	}

	return nil
}

// CopyFile 复制文件
func (c *Client) CopyFile(req *CopyFileRequest) error {
	url := fmt.Sprintf("%s/api/public/files/%s/%s/copy", strings.TrimRight(c.config.BaseURL, "/"), req.SrcBucket, req.SrcKey)

	jsonData, err := json.Marshal(map[string]string{
		"destBucket": req.DestBucket,
		"destKey":    req.DestKey,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set(constants.CONETENT_TYPE, "application/json")
	httpReq.Header.Set(constants.USER_AGENT, c.config.UserAgent)
	if c.config.APIKey != "" {
		httpReq.Header.Set(constants.XAPIKEY, c.config.APIKey)
	}
	if c.config.APISecret != "" {
		httpReq.Header.Set(constants.XAPISECRET, c.config.APISecret)
	}

	resp, err := c.doRequestWithRetry(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp)
	}

	return nil
}

// MoveFile 移动文件
func (c *Client) MoveFile(req *MoveFileRequest) error {
	url := fmt.Sprintf("%s/api/public/files/%s/%s/move", strings.TrimRight(c.config.BaseURL, "/"), req.SrcBucket, req.SrcKey)

	jsonData, err := json.Marshal(map[string]string{
		"destBucket": req.DestBucket,
		"destKey":    req.DestKey,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set(constants.CONETENT_TYPE, "application/json")
	httpReq.Header.Set(constants.USER_AGENT, c.config.UserAgent)
	if c.config.APIKey != "" {
		httpReq.Header.Set(constants.XAPIKEY, c.config.APIKey)
	}
	if c.config.APISecret != "" {
		httpReq.Header.Set(constants.XAPISECRET, c.config.APISecret)
	}

	resp, err := c.doRequestWithRetry(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp)
	}

	return nil
}

// uploadReader common upload method
func (c *Client) uploadReader(reader io.Reader, filename string, size int64, req *UploadRequest) (*UploadResult, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	fileWriter, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(fileWriter, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file data: %w", err)
	}
	if req.Bucket != "" {
		writer.WriteField("bucket", req.Bucket)
	}
	if req.Key != "" {
		writer.WriteField("key", req.Key)
	}
	if req.Compress {
		writer.WriteField("compress", "true")
		if req.Quality > 0 {
			writer.WriteField("quality", strconv.Itoa(req.Quality))
		}
	}
	if req.Watermark {
		writer.WriteField("watermark", "true")
		if req.WatermarkText != "" {
			writer.WriteField("watermarkText", req.WatermarkText)
		}
		if req.WatermarkPosition != "" {
			writer.WriteField("watermarkPosition", req.WatermarkPosition)
		}
	}
	writer.Close()
	url := strings.TrimRight(c.config.BaseURL, "/") + "/api/public/upload"
	httpReq, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set(constants.CONETENT_TYPE, writer.FormDataContentType())
	httpReq.Header.Set(constants.USER_AGENT, c.config.UserAgent)
	if c.config.APIKey != "" {
		httpReq.Header.Set(constants.XAPIKEY, c.config.APIKey)
	}
	if c.config.APISecret != "" {
		httpReq.Header.Set(constants.XAPISECRET, c.config.APISecret)
	}
	if len(req.AllowedTypes) > 0 {
		q := httpReq.URL.Query()
		for _, t := range req.AllowedTypes {
			q.Add("allowedTypes", t)
		}
		httpReq.URL.RawQuery = q.Encode()
	}
	var resp *http.Response
	var lastErr error
	for i := 0; i <= c.config.RetryCount; i++ {
		resp, lastErr = c.httpClient.Do(httpReq)
		if lastErr == nil && resp.StatusCode < 500 {
			break
		}
		if i < c.config.RetryCount {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}
	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d retries: %w", c.config.RetryCount, lastErr)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		if json.Unmarshal(respBody, &apiErr) == nil {
			apiErr.StatusCode = resp.StatusCode
			return nil, &apiErr
		}
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}
	var apiResp struct {
		Success bool         `json:"success"`
		Message string       `json:"message"`
		Data    UploadResult `json:"data"`
		Code    int          `json:"code"`
		Msg     string       `json:"msg"`
	}
	err = json.Unmarshal(respBody, &apiResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	isSuccess := apiResp.Success || (apiResp.Code == 200)
	if !isSuccess {
		message := apiResp.Message
		if message == "" {
			message = apiResp.Msg
		}
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    message,
		}
	}
	return &apiResp.Data, nil
}

// doRequestWithRetry 执行带重试的HTTP请求
func (c *Client) doRequestWithRetry(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var lastErr error

	for i := 0; i <= c.config.RetryCount; i++ {
		resp, lastErr = c.httpClient.Do(req)
		if lastErr == nil && resp.StatusCode < 500 {
			break
		}
		if i < c.config.RetryCount {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d retries: %w", c.config.RetryCount, lastErr)
	}

	return resp, nil
}

// handleErrorResponse 处理错误响应
func (c *Client) handleErrorResponse(resp *http.Response) error {
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read error response: %w", err)
	}

	var apiErr APIError
	if json.Unmarshal(respBody, &apiErr) == nil {
		apiErr.StatusCode = resp.StatusCode
		return &apiErr
	}

	return &APIError{
		StatusCode: resp.StatusCode,
		Message:    string(respBody),
	}
}

// progressReader func reader from io.Reader
type progressReader struct {
	reader   io.Reader
	total    int64
	read     int64
	callback func(uploaded, total int64)
}

func (pr *progressReader) Read(p []byte) (n int, err error) {
	n, err = pr.reader.Read(p)
	pr.read += int64(n)
	if pr.callback != nil {
		pr.callback(pr.read, pr.total)
	}
	return
}
