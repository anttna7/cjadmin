# Excel导入导出功能设置指南

## 概述

本系统支持Excel格式的数据导入和导出功能，可用于以下模块：
- 客户数据导入导出
- 合同数据导出
- 发票数据导出
- 订单数据导出
- 自定义表单数据导出

## 安装依赖

由于Excel功能依赖第三方库，需要先安装依赖：

```bash
go get github.com/xuri/excelize/v2
```

## 功能说明

### 1. 客户数据导出 (已实现接口)

**API接口**: `GET /api/customers/export`

**权限要求**: `customer.read`

**功能**: 导出客户列表为Excel文件

**Excel表头**:
- ID
- 客户名称
- 联系人
- 电话
- 邮箱
- 地址
- 客户类型
- 状态
- 创建时间

### 2. 客户数据导入 (已实现接口)

**API接口**: `POST /api/customers/import`

**权限要求**: `customer.create`

**功能**: 从Excel文件批量导入客户数据

**Excel模板**: 可通过 `GET /api/customers/template` 下载导入模板

**必填字段**:
- 客户名称

**可选字段**:
- 联系人
- 电话（与邮箱至少填一个）
- 邮箱（与电话至少填一个）
- 地址
- 客户类型（个人/企业/代理商）

**导入流程**:
1. 下载导入模板
2. 填写客户数据
3. 上传Excel文件
4. 系统验证数据
5. 批量创建客户记录
6. 返回导入结果（成功数、失败数、错误详情）

### 3. 合同数据导出

**API接口**: `GET /api/contracts/export`

**权限要求**: `customer.read`

**Excel表头**:
- ID
- 合同编号
- 客户名称
- 合同标题
- 合同金额
- 签订日期
- 开始日期
- 结束日期
- 状态
- 创建时间

### 4. 发票数据导出

**API接口**: `GET /api/invoices/export`

**权限要求**: `finance.read`

**Excel表头**:
- ID
- 发票号
- 客户名称
- 发票抬头
- 税号
- 金额
- 税额
- 状态
- 开具日期
- 创建时间

### 5. 订单数据导出

**API接口**: `GET /api/orders/export`

**权限要求**: `order.read`

**Excel表头**:
- ID
- 订单号
- 客户名称
- 订单类型
- 金额
- 状态
- 创建时间
- 更新时间

### 6. 表单数据导出

**API接口**: `GET /api/forms/export/:form_id`

**权限要求**: `customer.read`

**功能**: 导出指定表单的所有提交数据

## 使用示例

### 导出客户数据

```bash
# 使用curl
curl -H "Authorization: Bearer YOUR_TOKEN" \
     http://localhost:8080/api/customers/export \
     -o customers.xlsx

# 使用JavaScript
fetch('/api/customers/export', {
  headers: {
    'Authorization': 'Bearer ' + token
  }
})
.then(response => response.blob())
.then(blob => {
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = 'customers.xlsx';
  a.click();
});
```

### 下载导入模板

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     http://localhost:8080/api/customers/template \
     -o customer_template.xlsx
```

### 导入客户数据

```bash
curl -X POST \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -F "file=@customers.xlsx" \
     http://localhost:8080/api/customers/import
```

```javascript
// 使用JavaScript + FormData
const formData = new FormData();
formData.append('file', fileInput.files[0]);

fetch('/api/customers/import', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer ' + token
  },
  body: formData
})
.then(response => response.json())
.then(data => {
  console.log('导入成功:', data.success_count);
  console.log('导入失败:', data.fail_count);
  console.log('错误详情:', data.errors);
});
```

## 实现步骤

### 启用Excel功能的步骤

1. **安装依赖**
   ```bash
   go get github.com/xuri/excelize/v2
   ```

2. **更新 internal/utils/excel.go**
   - 取消注释所有Excel导出导入实现代码
   - 移除错误返回语句

3. **测试Excel功能**
   ```bash
   go test ./internal/utils -v
   ```

4. **重新编译项目**
   ```bash
   make build
   ```

## 数据验证

### 客户导入数据验证规则

1. **客户名称**: 必填，长度 1-100 字符
2. **联系人**: 可选，长度不超过 50 字符
3. **电话**: 与邮箱至少填一个，格式验证（中国手机号）
4. **邮箱**: 与电话至少填一个，邮箱格式验证
5. **地址**: 可选，长度不超过 200 字符
6. **客户类型**: 可选，限定值（个人/企业/代理商）

### 错误处理

导入时如果遇到错误，系统会：
1. 记录所有验证错误（行号 + 错误信息）
2. 跳过错误行，继续处理其他数据
3. 返回导入摘要：
   - 总行数
   - 成功导入数
   - 失败数
   - 详细错误列表

## 性能优化

### 大批量导入优化

1. **分批处理**: 每批次处理 500 条记录
2. **事务管理**: 使用数据库事务确保数据一致性
3. **异步处理**: 大文件（>10000行）使用异步任务
4. **进度通知**: 通过WebSocket实时推送导入进度

### 大批量导出优化

1. **流式写入**: 使用流式API避免内存溢出
2. **分页查询**: 分批从数据库读取数据
3. **压缩输出**: 超过1万行数据自动压缩为zip

## 安全考虑

1. **文件大小限制**: 最大 10MB
2. **文件类型验证**: 仅允许 .xlsx 和 .xls 格式
3. **权限验证**: 所有操作都需要相应的权限
4. **数据隔离**: 租户数据完全隔离
5. **审计日志**: 记录所有导入导出操作

## 故障排除

### 常见问题

1. **问题**: Excel导出返回错误 "Excel导出功能需要安装 github.com/xuri/excelize/v2 库"
   **解决**: 运行 `go get github.com/xuri/excelize/v2` 并重新编译

2. **问题**: 导入数据时所有行都失败
   **解决**: 检查Excel表头是否与模板完全一致

3. **问题**: 中文显示乱码
   **解决**: 确保Excel文件使用UTF-8编码

4. **问题**: 导出文件过大
   **解决**: 使用查询参数过滤数据，或分批导出

## 扩展功能

### 计划中的功能

1. **自定义导出字段**: 允许用户选择要导出的字段
2. **导出格式选择**: 支持CSV、JSON等多种格式
3. **导入预览**: 导入前预览数据
4. **导入模板管理**: 保存常用的导入模板
5. **计划任务导出**: 定时自动导出报表

## 技术细节

### Excel库选择

使用 `excelize` 的原因：
- 纯Go实现，无需CGO
- 支持最新的Excel格式（.xlsx）
- 性能优秀，支持大文件
- 活跃维护，文档完善
- MIT许可证

### 代码位置

- Excel工具类: `internal/utils/excel.go`
- 客户导入服务: `internal/service/customer.go` (ImportCustomers方法)
- 导出接口: 各模块的Handler（CustomerHandler, ContractHandler等）

## 参考链接

- Excelize官方文档: https://xuri.me/excelize/
- Excel格式规范: https://learn.microsoft.com/en-us/openspecs/office_standards/
