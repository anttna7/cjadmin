package handler

import (
	"path/filepath"
	"strconv"

	"github.com/anttna7/cjadmin/internal/middleware"
	"github.com/anttna7/cjadmin/internal/service"
	"github.com/anttna7/cjadmin/internal/utils"
	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	uploadService *service.UploadService
}

func NewUploadHandler() *UploadHandler {
	return &UploadHandler{
		uploadService: service.NewUploadService(),
	}
}

// UploadFile 上传文件
// 任务 14.3.1: 上传文件API
func (h *UploadHandler) UploadFile(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "请上传文件")
		return
	}

	// 获取资源类型和资源ID
	resourceType := c.PostForm("resource_type")
	if resourceType == "" {
		resourceType = "general"
	}

	var resourceID *int64
	if ridStr := c.PostForm("resource_id"); ridStr != "" {
		rid, err := strconv.ParseInt(ridStr, 10, 64)
		if err == nil {
			resourceID = &rid
		}
	}

	// 上传文件
	userID := middleware.GetUserID(c)
	fileInfo, err := h.uploadService.UploadFile(file, *tenantID, userID, resourceType, resourceID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "文件上传成功", fileInfo)
}

// UploadImage 上传图片
// 任务 14.3.2: 上传图片API
func (h *UploadHandler) UploadImage(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "请上传图片")
		return
	}

	// 获取资源类型和资源ID
	resourceType := c.PostForm("resource_type")
	if resourceType == "" {
		resourceType = "image"
	}

	var resourceID *int64
	if ridStr := c.PostForm("resource_id"); ridStr != "" {
		rid, err := strconv.ParseInt(ridStr, 10, 64)
		if err == nil {
			resourceID = &rid
		}
	}

	// 上传图片
	userID := middleware.GetUserID(c)
	fileInfo, err := h.uploadService.UploadImage(file, *tenantID, userID, resourceType, resourceID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "图片上传成功", fileInfo)
}

// UploadDocument 上传文档
// 任务 14.3.3: 上传文档API
func (h *UploadHandler) UploadDocument(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		utils.Error(c, utils.CodeInvalidParams, "请上传文档")
		return
	}

	// 获取资源类型和资源ID
	resourceType := c.PostForm("resource_type")
	if resourceType == "" {
		resourceType = "document"
	}

	var resourceID *int64
	if ridStr := c.PostForm("resource_id"); ridStr != "" {
		rid, err := strconv.ParseInt(ridStr, 10, 64)
		if err == nil {
			resourceID = &rid
		}
	}

	// 上传文档
	userID := middleware.GetUserID(c)
	fileInfo, err := h.uploadService.UploadDocument(file, *tenantID, userID, resourceType, resourceID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "文档上传成功", fileInfo)
}

// DownloadFile 下载文件
// 任务 14.3.4: 下载文件API
func (h *UploadHandler) DownloadFile(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	resourceType := c.Param("resource_type")
	fileName := c.Param("filename")

	// 构建文件路径
	filePath := filepath.Join("./uploads", "tenant_"+tenantID, resourceType, fileName)

	// 检查文件是否存在
	fileInfo, err := h.uploadService.GetFile(filePath)
	if err != nil {
		utils.Error(c, utils.CodeNotFound, "文件不存在")
		return
	}

	// 设置响应头
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+fileInfo.FileName)
	c.Header("Content-Type", "application/octet-stream")

	// 发送文件
	c.File(filePath)
}

// GetUploadStats 获取上传统计
func (h *UploadHandler) GetUploadStats(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == nil {
		utils.Error(c, utils.CodeForbidden, "租户ID不能为空")
		return
	}

	stats, err := h.uploadService.GetUploadStats(*tenantID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, stats)
}
