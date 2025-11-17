# 测试指南

## 概述

本项目使用Go的原生testing包进行单元测试和集成测试，并使用testify库提供断言和测试辅助功能。

## 测试框架结构

```
cjadmin/
├── internal/
│   ├── test/
│   │   └── testutil.go          # 测试工具和辅助函数
│   └── service/
│       ├── auth_test.go         # 认证服务测试
│       ├── tenant_test.go       # 租户服务测试
│       ├── customer_test.go     # 客户服务测试
│       └── ...                  # 其他服务测试
```

## 安装依赖

```bash
# 安装testify断言库
go get github.com/stretchr/testify

# 下载所有测试依赖
go mod download
```

## 运行测试

### 运行所有测试

```bash
# 运行所有测试
go test ./...

# 运行测试并显示详细输出
go test -v ./...

# 运行测试并生成覆盖率报告
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# 查看覆盖率报告
go tool cover -html=coverage.out
```

### 运行特定包的测试

```bash
# 运行service包的测试
go test ./internal/service/...

# 运行handler包的测试
go test ./internal/handler/...
```

### 运行特定测试函数

```bash
# 运行特定的测试函数
go test -v ./internal/service -run TestAuthService_Login

# 运行匹配模式的测试
go test -v ./internal/service -run "Auth.*"
```

### 使用Makefile

```bash
# 运行测试
make test

# 运行测试并生成覆盖率
make test-coverage

# 运行竞态检测
make test-race
```

## 测试环境配置

### 数据库配置

测试使用独立的测试数据库，通过环境变量配置：

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=cjadmin_test
```

### 使用Docker运行测试数据库

```bash
# 启动PostgreSQL测试数据库
docker run -d \
  --name cjadmin-test-db \
  -e POSTGRES_DB=cjadmin_test \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 \
  postgres:18-alpine

# 停止并删除测试数据库
docker stop cjadmin-test-db
docker rm cjadmin-test-db
```

## 编写测试

### 基本测试结构

```go
package service

import (
    "testing"
    "cjadmin/internal/test"
    "github.com/stretchr/testify/assert"
)

func TestServiceName_MethodName(t *testing.T) {
    // 1. 设置测试环境
    testDB := test.SetupTestDB(t)
    defer testDB.TearDown(t)

    // 2. 创建服务实例
    service := NewServiceName(testDB.DB)

    // 3. 准备测试数据
    tenant := testDB.CreateTestTenant(t, "test-tenant")

    // 4. 定义测试用例
    tests := []struct {
        name    string
        input   interface{}
        wantErr bool
    }{
        {
            name:    "successful case",
            input:   validInput,
            wantErr: false,
        },
        {
            name:    "error case",
            input:   invalidInput,
            wantErr: true,
        },
    }

    // 5. 运行测试用例
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := service.Method(tt.input)

            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, result)
            }
        })
    }
}
```

### 使用测试工具

#### 创建测试租户

```go
tenant := testDB.CreateTestTenant(t, "test-tenant")
```

#### 创建测试用户

```go
user := testDB.CreateTestUser(t, tenant.ID, "testuser", "admin")
```

#### 创建测试客户

```go
customer := testDB.CreateTestCustomer(t, tenant.ID, "Test Customer")
```

#### 创建测试部门

```go
dept := testDB.CreateTestDepartment(t, tenant.ID, "IT部门", nil)
```

### 常用断言

```go
// 基本断言
assert.Equal(t, expected, actual)
assert.NotEqual(t, expected, actual)
assert.Nil(t, object)
assert.NotNil(t, object)

// 错误断言
assert.NoError(t, err)
assert.Error(t, err)
assert.Contains(t, err.Error(), "expected message")

// 布尔断言
assert.True(t, value)
assert.False(t, value)

// 集合断言
assert.Len(t, list, expectedLength)
assert.Empty(t, list)
assert.NotEmpty(t, list)
assert.Contains(t, list, element)

// 数值断言
assert.Greater(t, actual, expected)
assert.Less(t, actual, expected)
assert.InDelta(t, expected, actual, delta)
```

## 测试覆盖率目标

| 组件 | 目标覆盖率 | 当前状态 |
|------|-----------|---------|
| Service层 | 80%+ | 🟡 进行中 |
| Handler层 | 70%+ | ⚪ 待开始 |
| Utils层 | 90%+ | ⚪ 待开始 |
| Models层 | 60%+ | ⚪ 待开始 |
| 整体 | 75%+ | ⚪ 待开始 |

## 测试最佳实践

### 1. 测试隔离

- 每个测试应该是独立的，不依赖其他测试
- 使用 `testDB.TearDown(t)` 清理测试数据
- 避免使用全局变量

### 2. 测试命名

- 函数名: `Test<ServiceName>_<MethodName>`
- 测试用例名: 描述测试场景，如 "successful creation"、"invalid input"

### 3. 表驱动测试

优先使用表驱动测试方式，提高测试覆盖率和可维护性：

```go
tests := []struct {
    name    string
    input   interface{}
    want    interface{}
    wantErr bool
}{
    // 测试用例...
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // 测试逻辑...
    })
}
```

### 4. 错误处理测试

- 测试正常流程
- 测试边界条件
- 测试错误情况
- 验证错误消息

### 5. 数据库事务测试

对于涉及事务的操作，确保测试：
- 事务提交后的状态
- 事务回滚的情况
- 并发操作的安全性

### 6. Mock和Stub

对于外部依赖（如第三方API、文件系统等），使用mock或stub：

```go
// 使用testify/mock
type MockService struct {
    mock.Mock
}

func (m *MockService) Method(args) (result, error) {
    args := m.Called(args)
    return args.Get(0).(result), args.Error(1)
}
```

## CI/CD集成

测试已集成到GitHub Actions工作流中：

```yaml
- name: Run tests
  run: |
    go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

- name: Upload coverage to Codecov
  uses: codecov/codecov-action@v3
  with:
    files: ./coverage.out
```

## 性能测试

### 基准测试

```go
func BenchmarkServiceMethod(b *testing.B) {
    // 设置
    testDB := test.SetupTestDB(&testing.T{})
    service := NewService(testDB.DB)

    // 重置计时器
    b.ResetTimer()

    // 运行N次
    for i := 0; i < b.N; i++ {
        service.Method(args)
    }
}
```

运行基准测试：

```bash
# 运行基准测试
go test -bench=. ./...

# 生成CPU profile
go test -bench=. -cpuprofile=cpu.prof ./...

# 生成内存profile
go test -bench=. -memprofile=mem.prof ./...

# 查看profile
go tool pprof cpu.prof
```

## 集成测试

### API集成测试

```go
func TestAPI_CreateCustomer(t *testing.T) {
    // 启动测试服务器
    router := setupRouter()
    w := httptest.NewRecorder()

    // 构造请求
    body := `{"customer_name": "Test Customer"}`
    req, _ := http.NewRequest("POST", "/api/customers", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+token)

    // 执行请求
    router.ServeHTTP(w, req)

    // 验证响应
    assert.Equal(t, 200, w.Code)
}
```

### 数据库集成测试

已通过 `internal/test/testutil.go` 提供完整的数据库测试工具。

## 故障排查

### 测试失败常见原因

1. **数据库连接失败**
   - 检查PostgreSQL是否运行
   - 验证环境变量配置
   - 确认测试数据库已创建

2. **测试数据污染**
   - 确保每个测试后调用 `TearDown`
   - 检查是否有未清理的数据

3. **并发问题**
   - 使用 `-race` 标志检测竞态条件
   - 检查共享资源的访问

4. **超时**
   - 增加测试超时时间: `go test -timeout 30s`
   - 检查是否有死锁或无限循环

### 调试技巧

```bash
# 运行单个测试并显示详细日志
go test -v -run TestSpecificTest ./internal/service

# 在测试中打印调试信息
t.Logf("Debug info: %+v", variable)

# 使用Delve调试器
dlv test ./internal/service -- -test.run TestSpecificTest
```

## 参考资源

- [Go Testing 官方文档](https://golang.org/pkg/testing/)
- [Testify 文档](https://github.com/stretchr/testify)
- [表驱动测试指南](https://github.com/golang/go/wiki/TableDrivenTests)
- [Go测试最佳实践](https://golang.org/doc/effective_go#testing)

## 下一步计划

- [ ] 完成所有Service层测试（目标80%+覆盖率）
- [ ] 添加Handler层集成测试
- [ ] 添加并发测试用例
- [ ] 添加性能基准测试
- [ ] 集成测试覆盖率报告到CI/CD
- [ ] 添加端到端测试
