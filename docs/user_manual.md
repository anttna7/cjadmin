# 用户操作手册

## 目录

1. [系统介绍](#系统介绍)
2. [快速开始](#快速开始)
3. [用户管理](#用户管理)
4. [客户管理](#客户管理)
5. [合同管理](#合同管理)
6. [订单管理](#订单管理)
7. [财务管理](#财务管理)
8. [数据导入导出](#数据导入导出)
9. [文件管理](#文件管理)
10. [常见问题](#常见问题)

---

## 系统介绍

### 什么是客户管理+收款系统？

客户管理+收款系统是一个多租户SaaS应用，帮助企业管理客户信息、合同、订单和财务流程。

### 核心功能

- **客户管理**：维护客户基本信息、联系方式
- **合同管理**：管理客户合同的全生命周期
- **订单管理**：创建充值订单、处理支付
- **财务管理**：发票开具、账户余额管理
- **数据管理**：Excel批量导入导出、文件上传

### 用户角色

- **平台管理员**：管理所有租户和系统配置
- **租户管理员**：管理本租户的所有数据和用户
- **普通用户**：根据分配的权限操作系统

---

## 快速开始

### 1. 登录系统

访问系统地址：`http://your-domain.com`

1. 输入用户名和密码
2. 点击"登录"按钮
3. 首次登录建议修改密码

**API登录示例：**
```bash
curl -X POST http://your-domain.com/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "your_username",
    "password": "your_password"
  }'
```

### 2. 获取访问令牌

登录成功后会返回JWT令牌，后续所有请求需要在Header中携带：

```
Authorization: Bearer your_token_here
```

### 3. 查看个人权限

```bash
curl -X GET http://your-domain.com/api/auth/permissions \
  -H "Authorization: Bearer your_token"
```

---

## 用户管理

### 创建用户

**权限要求**：`user.create`

**步骤：**
1. 导航到"用户管理"页面
2. 点击"新建用户"按钮
3. 填写用户信息：
   - 用户名（必填，唯一）
   - 密码（必填，至少6位）
   - 邮箱（可选）
   - 手机号（可选）
   - 所属部门（可选）
4. 点击"保存"

**API示例：**
```bash
curl -X POST http://your-domain.com/api/auth/register \
  -H "Authorization: Bearer your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "newuser",
    "password": "password123",
    "email": "user@example.com"
  }'
```

### 分配角色

**权限要求**：`user.update`

**步骤：**
1. 在用户列表中找到目标用户
2. 点击"分配角色"
3. 选择一个或多个角色
4. 点击"确定"

**API示例：**
```bash
curl -X POST http://your-domain.com/api/user-roles/{user_id}/assign \
  -H "Authorization: Bearer your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "role_id": 1
  }'
```

---

## 客户管理

### 创建客户

**权限要求**：`customer.create`

**步骤：**
1. 导航到"客户管理"页面
2. 点击"新建客户"
3. 填写客户信息：
   - **客户名称**（必填）
   - **联系人**（可选）
   - **联系电话**（建议填写）
   - **联系邮箱**（建议填写）
   - **地址**（可选）
   - **客户类型**：个人/企业/代理商
   - **自定义字段**（JSON格式）
4. 点击"保存"

**API示例：**
```bash
curl -X POST http://your-domain.com/api/customers \
  -H "Authorization: Bearer your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_name": "示例公司",
    "contact_person": "张三",
    "contact_phone": "13800138000",
    "contact_email": "zhangsan@example.com",
    "address": "北京市朝阳区",
    "customer_type": "enterprise"
  }'
```

### 查看客户列表

**权限要求**：`customer.read`

**筛选条件：**
- 客户名称（模糊搜索）
- 客户类型
- 创建时间范围

**API示例：**
```bash
# 获取第1页，每页20条
curl -X GET "http://your-domain.com/api/customers?page=1&size=20" \
  -H "Authorization: Bearer your_token"

# 搜索客户
curl -X GET "http://your-domain.com/api/customers?keyword=示例&customer_type=enterprise" \
  -H "Authorization: Bearer your_token"
```

### 更新客户信息

**权限要求**：`customer.update`

**步骤：**
1. 在客户列表中找到目标客户
2. 点击"编辑"按钮
3. 修改需要更新的信息
4. 点击"保存"

**API示例：**
```bash
curl -X PUT http://your-domain.com/api/customers/1 \
  -H "Authorization: Bearer your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "contact_phone": "13900139000",
    "address": "上海市浦东新区"
  }'
```

### 删除客户

**权限要求**：`customer.delete`

⚠️ **注意**：删除客户前请确保：
- 客户没有未完成的订单
- 客户没有未开具的发票
- 已备份重要数据

**API示例：**
```bash
curl -X DELETE http://your-domain.com/api/customers/1 \
  -H "Authorization: Bearer your_token"
```

---

## 合同管理

### 创建合同

**权限要求**：`customer.create`

**步骤：**
1. 导航到"合同管理"页面
2. 点击"新建合同"
3. 填写合同信息：
   - **客户**（必填，选择已有客户）
   - **合同标题**（必填）
   - **合同编号**（自动生成或手动输入）
   - **合同金额**（必填）
   - **签订日期**（可选）
   - **开始日期**（可选）
   - **结束日期**（可选）
   - **合同内容**（可选）
   - **附件**（可选，支持上传PDF、Word等）
4. 点击"保存"

**API示例：**
```bash
curl -X POST http://your-domain.com/api/contracts \
  -H "Authorization: Bearer your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": 1,
    "contract_title": "2024年度广告投放合同",
    "contract_amount": 100000.00,
    "sign_date": "2024-01-01",
    "start_date": "2024-01-01",
    "end_date": "2024-12-31"
  }'
```

### 变更合同状态

**合同状态：**
- `draft` - 草稿
- `signed` - 已签订
- `executing` - 执行中
- `completed` - 已完成
- `terminated` - 已终止

**API示例：**
```bash
curl -X PUT http://your-domain.com/api/contracts/1/status \
  -H "Authorization: Bearer your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "signed"
  }'
```

---

## 订单管理

### 创建充值订单

**权限要求**：`order.create`

**订单流程：**
```
创建订单 → 待支付 → 已支付 → 财务审核 → 审核通过 → 余额入账
```

**步骤：**
1. 导航到"订单管理"页面
2. 点击"创建充值订单"
3. 填写订单信息：
   - **客户**（必填）
   - **充值金额**（必填）
   - **支付方式**：支付宝/微信/银行转账
   - **备注**（可选）
4. 点击"创建订单"
5. 系统生成订单号
6. 客户完成支付后，更新订单状态

**API示例：**
```bash
curl -X POST http://your-domain.com/api/orders/recharge \
  -H "Authorization: Bearer your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": 1,
    "amount": 10000.00,
    "payment_method": "alipay",
    "notes": "2024年Q1充值"
  }'
```

### 财务审核订单

**权限要求**：`finance.verify`

**步骤：**
1. 导航到"订单管理" → "待审核订单"
2. 查看订单详情和支付凭证
3. 确认金额无误
4. 点击"审核通过"或"审核拒绝"
5. 填写审核意见
6. 提交审核结果

**API示例：**
```bash
# 审核通过
curl -X POST http://your-domain.com/api/orders/1/verify \
  -H "Authorization: Bearer your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "approved": true,
    "notes": "金额无误，审核通过"
  }'

# 审核拒绝
curl -X POST http://your-domain.com/api/orders/1/verify \
  -H "Authorization: Bearer your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "approved": false,
    "notes": "金额不符，请重新确认"
  }'
```

### 创建转账订单

**权限要求**：`order.create`

**用途**：将资金账户余额转到消耗账户

**API示例：**
```bash
curl -X POST http://your-domain.com/api/orders/transfer \
  -H "Authorization: Bearer your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": 1,
    "amount": 5000.00
  }'
```

### 查看订单详情

**API示例：**
```bash
curl -X GET http://your-domain.com/api/orders/1 \
  -H "Authorization: Bearer your_token"
```

---

## 财务管理

### 创建发票

**权限要求**：`finance.create`

**步骤：**
1. 导航到"财务管理" → "发票管理"
2. 点击"创建发票"
3. 填写发票信息：
   - **客户**（必填）
   - **关联订单**（可选）
   - **发票类型**：增值税普通发票/增值税专用发票
   - **发票抬头**（必填）
   - **税号**（必填）
   - **发票金额**（必填）
   - **税额**（可选）
4. 点击"保存"

**API示例：**
```bash
curl -X POST http://your-domain.com/api/invoices \
  -H "Authorization: Bearer your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": 1,
    "order_id": 1,
    "invoice_type": "vat_normal",
    "invoice_title": "示例公司",
    "tax_no": "91110000000000000X",
    "amount": 10000.00,
    "tax_amount": 600.00
  }'
```

### 开具发票

**权限要求**：`finance.verify`

**步骤：**
1. 在发票列表中找到待开具的发票
2. 上传发票PDF文件
3. 点击"开具发票"
4. 系统自动更新发票状态为"已开具"

**API示例：**
```bash
curl -X PUT http://your-domain.com/api/invoices/1/issue \
  -H "Authorization: Bearer your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "file_url": "/uploads/invoices/invoice_001.pdf"
  }'
```

### 作废发票

**权限要求**：`finance.verify`

⚠️ **注意**：发票作废后无法恢复，请谨慎操作

**API示例：**
```bash
curl -X PUT http://your-domain.com/api/invoices/1/cancel \
  -H "Authorization: Bearer your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "客户要求重新开具"
  }'
```

### 查看账户余额

**权限要求**：`customer.read`

**账户类型：**
- **资金账户**：充值后的资金存放处
- **消耗账户**：可用于业务消耗的余额

**API示例：**
```bash
curl -X GET http://your-domain.com/api/customers/1/accounts \
  -H "Authorization: Bearer your_token"
```

---

## 数据导入导出

### 导出客户数据

**权限要求**：`customer.read`

**支持格式**：Excel (.xlsx)

**步骤：**
1. 导航到"客户管理"页面
2. 点击"导出"按钮
3. 选择导出条件（可选）
4. 点击"确认导出"
5. 系统生成Excel文件并自动下载

**API示例：**
```bash
# 导出所有客户
curl -X GET http://your-domain.com/api/customers/export \
  -H "Authorization: Bearer your_token" \
  -o customers.xlsx

# 导出特定类型客户
curl -X GET "http://your-domain.com/api/customers/export?customer_type=enterprise" \
  -H "Authorization: Bearer your_token" \
  -o enterprise_customers.xlsx
```

### 下载导入模板

**步骤：**
1. 导航到"客户管理"页面
2. 点击"导入"按钮
3. 点击"下载模板"
4. 获得Excel模板文件

**API示例：**
```bash
curl -X GET http://your-domain.com/api/customers/template \
  -H "Authorization: Bearer your_token" \
  -o customer_template.xlsx
```

### 批量导入客户

**权限要求**：`customer.create`

**步骤：**
1. 下载并填写导入模板
2. 确保数据格式正确：
   - 客户名称：必填
   - 联系电话或邮箱：至少填一个
   - 客户类型：个人/企业/代理商
3. 点击"选择文件"上传Excel
4. 点击"开始导入"
5. 查看导入结果：
   - 成功导入数量
   - 失败数量
   - 错误详情

**API示例：**
```bash
curl -X POST http://your-domain.com/api/customers/import \
  -H "Authorization: Bearer your_token" \
  -F "file=@customers.xlsx"
```

**导入结果示例：**
```json
{
  "code": 200,
  "message": "导入完成",
  "data": {
    "total_count": 100,
    "success_count": 95,
    "fail_count": 5,
    "errors": [
      "第3行: 客户名称不能为空",
      "第7行: 电话和邮箱至少填写一个",
      "第12行: 客户类型无效"
    ]
  }
}
```

---

## 文件管理

### 上传文件

**权限要求**：根据资源类型而定

**支持的文件类型：**
- **图片**：jpg, jpeg, png, gif, bmp, webp（最大5MB）
- **文档**：pdf, doc, docx, xls, xlsx, ppt, pptx, txt（最大20MB）
- **压缩包**：zip, rar, 7z, tar, gz（最大20MB）

**步骤：**
1. 在需要上传文件的页面（如发票、合同）
2. 点击"上传文件"或"选择文件"
3. 选择本地文件
4. 等待上传完成
5. 文件自动关联到对应资源

**API示例：**
```bash
# 上传通用文件
curl -X POST http://your-domain.com/api/files/upload \
  -H "Authorization: Bearer your_token" \
  -F "file=@document.pdf" \
  -F "resource_type=invoice" \
  -F "resource_id=1"

# 上传图片
curl -X POST http://your-domain.com/api/files/upload/image \
  -H "Authorization: Bearer your_token" \
  -F "file=@photo.jpg" \
  -F "resource_type=customer" \
  -F "resource_id=1"

# 上传文档
curl -X POST http://your-domain.com/api/files/upload/document \
  -H "Authorization: Bearer your_token" \
  -F "file=@contract.pdf" \
  -F "resource_type=contract" \
  -F "resource_id=1"
```

### 下载文件

**API示例：**
```bash
curl -X GET http://your-domain.com/api/files/tenant_1/invoice/20240101120000_a1b2c3d4.pdf \
  -H "Authorization: Bearer your_token" \
  -o downloaded_file.pdf
```

---

## 常见问题

### 1. 忘记密码怎么办？

目前系统不支持自助重置密码，请联系管理员重置。

管理员可以通过以下方式重置用户密码：
1. 登录管理后台
2. 找到对应用户
3. 点击"重置密码"

### 2. 如何修改个人信息？

```bash
curl -X PUT http://your-domain.com/api/users/profile \
  -H "Authorization: Bearer your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newemail@example.com",
    "phone": "13900139000"
  }'
```

### 3. 为什么某些功能无法访问？

可能原因：
- 权限不足：联系管理员分配相应权限
- Token过期：重新登录获取新Token
- 账户被禁用：联系管理员启用账户

查看当前权限：
```bash
curl -X GET http://your-domain.com/api/auth/permissions \
  -H "Authorization: Bearer your_token"
```

### 4. 导入数据时出现错误怎么办？

常见错误及解决方法：
- **"客户名称不能为空"**：检查Excel中客户名称列是否填写
- **"电话和邮箱至少填写一个"**：至少填写联系电话或邮箱之一
- **"客户类型无效"**：只能填写：个人、企业、代理商
- **"文件格式错误"**：确保上传的是.xlsx或.xls格式

### 5. 如何查看操作历史？

系统会自动记录所有重要操作的审计日志：

```bash
curl -X GET "http://your-domain.com/api/audit-logs?page=1&size=20" \
  -H "Authorization: Bearer your_token"
```

### 6. 文件上传失败怎么办？

检查项：
- 文件大小：图片不超过5MB，其他文件不超过20MB
- 文件类型：确保文件类型在允许列表中
- 网络连接：检查网络是否稳定

### 7. API返回401错误？

原因：Token无效或已过期

解决方法：
1. 重新登录获取新Token
2. 检查Token是否正确携带在Header中
3. 确认Token格式：`Authorization: Bearer your_token`

### 8. 如何批量操作？

目前支持的批量操作：
- 批量导入客户（Excel）
- 批量分配角色
- 批量导出数据

更多批量操作功能持续开发中。

---

## 技术支持

### 联系方式
- 邮箱：support@example.com
- 电话：400-xxx-xxxx
- 在线客服：工作日 9:00-18:00

### 相关文档
- [管理员手册](admin_manual.md)
- [API文档](api_examples.md)
- [部署指南](docker_deployment.md)

### 问题反馈
- GitHub Issues: https://github.com/anttna7/cjadmin/issues
- 功能建议：欢迎提交Issue

---

**版本**: v1.0
**最后更新**: 2025-11-17
**适用版本**: 客户管理+收款系统 v1.0+
