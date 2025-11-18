package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{
			name:   "6 characters",
			length: 6,
		},
		{
			name:   "16 characters",
			length: 16,
		},
		{
			name:   "32 characters",
			length: 32,
		},
		{
			name:   "64 characters",
			length: 64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateRandomString(tt.length)
			assert.Len(t, result, tt.length)

			// 验证是十六进制字符串
			for _, c := range result {
				assert.True(t, (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'))
			}

			// 生成第二个字符串，应该不同
			result2 := GenerateRandomString(tt.length)
			assert.NotEqual(t, result, result2)
		})
	}
}

func TestSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := map[string]string{"key": "value"}
	Success(c, data)

	assert.Equal(t, http.StatusOK, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 0, response.Code)
	assert.Equal(t, "success", response.Message)
}

func TestBadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	BadRequest(c, "参数错误")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeInvalidParams, response.Code)
	assert.Equal(t, "参数错误", response.Message)
}

func TestUnauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Unauthorized(c, "未授权")

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeUnauthorized, response.Code)
	assert.Equal(t, "未授权", response.Message)
}

func TestForbidden(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Forbidden(c, "禁止访问")

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeForbidden, response.Code)
	assert.Equal(t, "禁止访问", response.Message)
}

func TestNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	NotFound(c, "资源不存在")

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeNotFound, response.Code)
	assert.Equal(t, "资源不存在", response.Message)
}

func TestTooManyRequests(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	TooManyRequests(c, "请求过于频繁")

	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeTooManyRequests, response.Code)
	assert.Equal(t, "请求过于频繁", response.Message)
}

func TestServerError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ServerError(c, "服务器错误")

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, CodeInternalError, response.Code)
	assert.Equal(t, "服务器错误", response.Message)
}

func TestPageSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := []map[string]string{
		{"id": "1", "name": "item1"},
		{"id": "2", "name": "item2"},
	}
	PageSuccess(c, data, 100, 1, 20)

	assert.Equal(t, http.StatusOK, w.Code)

	var response PageResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 0, response.Code)
	assert.Equal(t, "success", response.Message)
	assert.Equal(t, int64(100), response.Total)
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 20, response.Size)
}

func TestSuccessWithPage(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := []string{"a", "b", "c"}
	SuccessWithPage(c, data, 50, 2, 10)

	assert.Equal(t, http.StatusOK, w.Code)

	var response PageResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, int64(50), response.Total)
	assert.Equal(t, 2, response.Page)
	assert.Equal(t, 10, response.Size)
}

func TestSuccessWithMessage(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := gin.H{"result": "ok"}
	SuccessWithMessage(c, "操作成功", data)

	assert.Equal(t, http.StatusOK, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 0, response.Code)
	assert.Equal(t, "操作成功", response.Message)
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Error(c, 1001, "自定义错误")

	assert.Equal(t, http.StatusOK, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 1001, response.Code)
	assert.Equal(t, "自定义错误", response.Message)
}

func TestErrorWithStatus(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ErrorWithStatus(c, http.StatusServiceUnavailable, 503, "服务不可用")

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 503, response.Code)
	assert.Equal(t, "服务不可用", response.Message)
}

func TestGenerateRandomString_Uniqueness(t *testing.T) {
	// 生成多个随机字符串并验证唯一性
	generated := make(map[string]bool)
	iterations := 1000

	for i := 0; i < iterations; i++ {
		str := GenerateRandomString(16)
		assert.False(t, generated[str], "生成了重复的随机字符串")
		generated[str] = true
	}

	assert.Len(t, generated, iterations)
}
