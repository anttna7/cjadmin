package service

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/anttna7/cjadmin/internal/config"
)

type UploadService struct{}

func NewUploadService() *UploadService {
	return &UploadService{}
}

// FileInfo 文件信息
type FileInfo struct {
	FileName     string    `json:"file_name"`
	FileSize     int64     `json:"file_size"`
	FileType     string    `json:"file_type"`
	FilePath     string    `json:"file_path"`
	FileURL      string    `json:"file_url"`
	FileMD5      string    `json:"file_md5"`
	UploadTime   time.Time `json:"upload_time"`
	TenantID     int64     `json:"tenant_id"`
	UploadedBy   int64     `json:"uploaded_by"`
	ResourceType string    `json:"resource_type"` // invoice, contract, customer等
	ResourceID   *int64    `json:"resource_id"`
}

// 允许的文件类型
var allowedFileTypes = map[string][]string{
	"image": {".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"},
	"document": {".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt"},
	"archive": {".zip", ".rar", ".7z", ".tar", ".gz"},
}

// 允许的MIME类型
var allowedMimeTypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/gif":       true,
	"image/bmp":       true,
	"image/webp":      true,
	"application/pdf": true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   true,
	"application/vnd.ms-excel":                                                  true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         true,
	"application/vnd.ms-powerpoint":                                             true,
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
	"text/plain": true,
	"application/zip":     true,
	"application/x-rar":   true,
	"application/x-7z-compressed": true,
}

// UploadFile 上传文件
// 任务 14.2.1: 实现文件上传
func (s *UploadService) UploadFile(file *multipart.FileHeader, tenantID, uploadedBy int64, resourceType string, resourceID *int64) (*FileInfo, error) {
	// 验证文件大小（默认最大20MB）
	maxSize := int64(20 * 1024 * 1024)
	if file.Size > maxSize {
		return nil, errors.New("文件大小超过限制（最大20MB）")
	}

	// 验证文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !isAllowedFileType(ext) {
		return nil, errors.New("不支持的文件类型")
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer src.Close()

	// 计算MD5
	hash := md5.New()
	if _, err := io.Copy(hash, src); err != nil {
		return nil, fmt.Errorf("计算文件MD5失败: %w", err)
	}
	fileMD5 := hex.EncodeToString(hash.Sum(nil))

	// 重置文件指针
	src.Seek(0, 0)

	// 生成存储路径
	uploadDir := getUploadDir(tenantID, resourceType)
	if err := ensureDir(uploadDir); err != nil {
		return nil, fmt.Errorf("创建上传目录失败: %w", err)
	}

	// 生成唯一文件名
	newFileName := generateFileName(file.Filename, fileMD5)
	filePath := filepath.Join(uploadDir, newFileName)

	// 保存文件
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("创建文件失败: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(filePath)
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}

	// 生成文件URL
	fileURL := generateFileURL(tenantID, resourceType, newFileName)

	fileInfo := &FileInfo{
		FileName:     file.Filename,
		FileSize:     file.Size,
		FileType:     ext,
		FilePath:     filePath,
		FileURL:      fileURL,
		FileMD5:      fileMD5,
		UploadTime:   time.Now(),
		TenantID:     tenantID,
		UploadedBy:   uploadedBy,
		ResourceType: resourceType,
		ResourceID:   resourceID,
	}

	return fileInfo, nil
}

// DeleteFile 删除文件
// 任务 14.2.2: 实现文件删除
func (s *UploadService) DeleteFile(filePath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return errors.New("文件不存在")
	}

	// 删除文件
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}

	return nil
}

// GetFile 获取文件信息
func (s *UploadService) GetFile(filePath string) (*FileInfo, error) {
	// 检查文件是否存在
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil, errors.New("文件不存在")
	}
	if err != nil {
		return nil, err
	}

	return &FileInfo{
		FileName: fileInfo.Name(),
		FileSize: fileInfo.Size(),
		FilePath: filePath,
	}, nil
}

// ValidateFile 验证文件
// 任务 14.2.3: 实现文件验证
func (s *UploadService) ValidateFile(file *multipart.FileHeader, allowedTypes []string, maxSize int64) error {
	// 验证文件大小
	if file.Size > maxSize {
		return fmt.Errorf("文件大小超过限制（最大%dMB）", maxSize/(1024*1024))
	}

	// 验证文件扩展名
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if len(allowedTypes) > 0 {
		allowed := false
		for _, t := range allowedTypes {
			if ext == t {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("不支持的文件类型: %s", ext)
		}
	}

	return nil
}

// UploadImage 上传图片（限制为图片类型）
func (s *UploadService) UploadImage(file *multipart.FileHeader, tenantID, uploadedBy int64, resourceType string, resourceID *int64) (*FileInfo, error) {
	// 验证是否为图片
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !isImageType(ext) {
		return nil, errors.New("只支持图片文件（jpg, jpeg, png, gif, bmp, webp）")
	}

	// 图片大小限制为5MB
	if file.Size > 5*1024*1024 {
		return nil, errors.New("图片大小超过限制（最大5MB）")
	}

	return s.UploadFile(file, tenantID, uploadedBy, resourceType, resourceID)
}

// UploadDocument 上传文档（限制为文档类型）
func (s *UploadService) UploadDocument(file *multipart.FileHeader, tenantID, uploadedBy int64, resourceType string, resourceID *int64) (*FileInfo, error) {
	// 验证是否为文档
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !isDocumentType(ext) {
		return nil, errors.New("只支持文档文件（pdf, doc, docx, xls, xlsx, ppt, pptx, txt）")
	}

	return s.UploadFile(file, tenantID, uploadedBy, resourceType, resourceID)
}

// 辅助函数

func isAllowedFileType(ext string) bool {
	for _, types := range allowedFileTypes {
		for _, t := range types {
			if ext == t {
				return true
			}
		}
	}
	return false
}

func isImageType(ext string) bool {
	for _, t := range allowedFileTypes["image"] {
		if ext == t {
			return true
		}
	}
	return false
}

func isDocumentType(ext string) bool {
	for _, t := range allowedFileTypes["document"] {
		if ext == t {
			return true
		}
	}
	return false
}

func getUploadDir(tenantID int64, resourceType string) string {
	baseDir := config.AppConfig.Upload.Path
	if baseDir == "" {
		baseDir = "./uploads"
	}
	return filepath.Join(baseDir, fmt.Sprintf("tenant_%d", tenantID), resourceType)
}

func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

func generateFileName(originalName, md5Hash string) string {
	ext := filepath.Ext(originalName)
	timestamp := time.Now().Format("20060102150405")
	// 使用时间戳 + MD5前8位 + 原扩展名
	return fmt.Sprintf("%s_%s%s", timestamp, md5Hash[:8], ext)
}

func generateFileURL(tenantID int64, resourceType, fileName string) string {
	baseURL := config.AppConfig.Server.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	return fmt.Sprintf("%s/api/files/tenant_%d/%s/%s", baseURL, tenantID, resourceType, fileName)
}

// GetUploadStats 获取上传统计信息
func (s *UploadService) GetUploadStats(tenantID int64) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	uploadDir := filepath.Join(config.AppConfig.Upload.Path, fmt.Sprintf("tenant_%d", tenantID))

	// 遍历目录计算统计信息
	var totalFiles int64
	var totalSize int64

	filepath.Walk(uploadDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			totalFiles++
			totalSize += info.Size()
		}
		return nil
	})

	stats["total_files"] = totalFiles
	stats["total_size"] = totalSize
	stats["total_size_mb"] = float64(totalSize) / (1024 * 1024)

	return stats, nil
}
