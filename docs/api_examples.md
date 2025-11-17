# API使用示例

本文档提供系统API的详细使用示例。

## 基础说明

- **Base URL**: `http://localhost:8080/api`
- **认证方式**: JWT Bearer Token
- **请求格式**: JSON
- **响应格式**: JSON

## 通用响应格式

### 成功响应
```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

### 分页响应
```json
{
  "code": 0,
  "message": "success",
  "data": [...],
  "total": 100,
  "page": 1,
  "size": 20
}
```

### 错误响应
```json
{
  "code": 400,
  "message": "错误信息"
}
```

## 1. 认证接口

### 1.1 用户登录

**请求**
```http
POST /api/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}
```

**响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6ImFkbWluIiwidGVuYW50X2lkIjpudWxsLCJpc19wbGF0Zm9ybV9hZG1pbiI6dHJ1ZSwiZXhwIjoxNzMyMjI4ODAwfQ.xxx",
    "user": {
      "id": 1,
      "username": "admin",
      "real_name": "超级管理员",
      "email": "",
      "tenant_id": null,
      "is_platform_admin": true,
      "permissions": [
        "customer.create",
        "customer.read",
        "customer.update",
        "customer.delete",
        "order.create",
        "order.read",
        "finance.verify"
      ]
    }
  }
}
```

**cURL示例**
```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'
```

### 1.2 获取当前用户信息

**请求**
```http
GET /api/user/current
Authorization: Bearer {token}
```

**响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "username": "admin",
    "tenant_id": null,
    "is_platform_admin": true
  }
}
```

**cURL示例**
```bash
curl -X GET http://localhost:8080/api/user/current \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 2. 客户管理接口

### 2.1 创建客户

**请求**
```http
POST /api/customers
Authorization: Bearer {token}
Content-Type: application/json

{
  "customer_code": "CUS20250001",
  "customer_name": "测试科技有限公司",
  "customer_type": "企业",
  "contact_person": "张三",
  "phone": "13800138000",
  "email": "zhangsan@test.com",
  "address": "北京市朝阳区xxx路xxx号",
  "industry": "互联网",
  "source": "电话营销",
  "custom_fields": {
    "备注": "重要客户",
    "标签": ["VIP", "重点关注"]
  }
}
```

**响应**
```json
{
  "code": 0,
  "message": "客户创建成功",
  "data": {
    "id": 1,
    "tenant_id": 1,
    "customer_code": "CUS20250001",
    "customer_name": "测试科技有限公司",
    "customer_type": "企业",
    "contact_person": "张三",
    "phone": "13800138000",
    "email": "zhangsan@test.com",
    "address": "北京市朝阳区xxx路xxx号",
    "industry": "互联网",
    "source": "电话营销",
    "status": "active",
    "custom_fields": {
      "备注": "重要客户",
      "标签": ["VIP", "重点关注"]
    },
    "created_at": "2025-11-17T10:00:00Z",
    "updated_at": "2025-11-17T10:00:00Z"
  }
}
```

**cURL示例**
```bash
curl -X POST http://localhost:8080/api/customers \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_name": "测试科技有限公司",
    "customer_type": "企业",
    "phone": "13800138000"
  }'
```

### 2.2 获取客户列表

**请求**
```http
GET /api/customers?page=1&size=20&customer_name=测试&status=active
Authorization: Bearer {token}
```

**参数说明**
- `page`: 页码（默认1）
- `size`: 每页数量（默认20，最大100）
- `customer_name`: 客户名称（模糊搜索）
- `status`: 状态筛选（active/inactive）
- `assigned_to`: 负责人ID

**响应**
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "customer_code": "CUS20250001",
      "customer_name": "测试科技有限公司",
      "customer_type": "企业",
      "phone": "13800138000",
      "status": "active",
      "assigned_user": {
        "id": 2,
        "username": "sales01",
        "real_name": "销售员A"
      },
      "created_at": "2025-11-17T10:00:00Z"
    }
  ],
  "total": 100,
  "page": 1,
  "size": 20
}
```

**cURL示例**
```bash
curl -X GET "http://localhost:8080/api/customers?page=1&size=20" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 2.3 获取客户详情

**请求**
```http
GET /api/customers/1
Authorization: Bearer {token}
```

**响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "customer_code": "CUS20250001",
    "customer_name": "测试科技有限公司",
    "accounts": [
      {
        "id": 1,
        "account_type": "fund",
        "balance": 10000.00,
        "cash_balance": 8000.00
      },
      {
        "id": 2,
        "account_type": "consumption",
        "balance": 5000.00,
        "cash_balance": 4000.00
      }
    ]
  }
}
```

### 2.4 更新客户

**请求**
```http
PUT /api/customers/1
Authorization: Bearer {token}
Content-Type: application/json

{
  "customer_name": "新公司名称",
  "phone": "13900139000",
  "status": "active"
}
```

**响应**
```json
{
  "code": 0,
  "message": "客户更新成功"
}
```

### 2.5 删除客户

**请求**
```http
DELETE /api/customers/1
Authorization: Bearer {token}
```

**响应**
```json
{
  "code": 0,
  "message": "客户删除成功"
}
```

### 2.6 获取客户账户信息

**请求**
```http
GET /api/customers/1/accounts
Authorization: Bearer {token}
```

**响应**
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "customer_id": 1,
      "account_type": "fund",
      "balance": 10000.00,
      "cash_balance": 8000.00,
      "created_at": "2025-11-17T10:00:00Z"
    },
    {
      "id": 2,
      "customer_id": 1,
      "account_type": "consumption",
      "balance": 5000.00,
      "cash_balance": 4000.00,
      "created_at": "2025-11-17T10:00:00Z"
    }
  ]
}
```

## 3. 订单管理接口

### 3.1 创建充值订单

**请求**
```http
POST /api/orders/recharge
Authorization: Bearer {token}
Content-Type: application/json

{
  "customer_id": 1,
  "amount": 1000.00,
  "payment_method": "支付宝",
  "payment_channel": "alipay"
}
```

**响应**
```json
{
  "code": 0,
  "message": "充值订单创建成功",
  "data": {
    "id": 1,
    "order_no": "RCH20251117abcd1234",
    "order_type": "recharge",
    "customer_id": 1,
    "amount": 1000.00,
    "status": "pending",
    "payment_method": "支付宝",
    "payment_channel": "alipay",
    "created_at": "2025-11-17T10:00:00Z"
  }
}
```

**业务流程说明**
1. 客户在前端页面发起充值
2. 系统创建充值订单，状态为`pending`（待支付）
3. 客户完成支付后，支付平台回调通知系统
4. 订单状态变更为`paid`（已支付）
5. 财务人员审核订单
6. 审核通过后，余额入账，订单状态变为`completed`

### 3.2 支付回调（模拟）

**请求**
```http
POST /api/payment/callback
Content-Type: application/json

{
  "order_no": "RCH20251117abcd1234",
  "transaction_id": "2025111712345678"
}
```

**响应**
```json
{
  "code": 0,
  "message": "支付回调处理成功"
}
```

**说明**: 实际生产环境中，此接口应该验证支付平台的签名。

### 3.3 财务审核充值订单

**请求**
```http
POST /api/orders/1/verify
Authorization: Bearer {token}
Content-Type: application/json

{
  "approved": true,
  "notes": "审核通过，资金已到账"
}
```

**响应**
```json
{
  "code": 0,
  "message": "审核完成"
}
```

**业务说明**
- 审核通过（approved: true）：
  - 订单状态：paid → verified → completed
  - 客户资金账户余额增加
  - 记录账户交易明细

- 审核不通过（approved: false）：
  - 订单状态：paid → failed
  - 余额不变

### 3.4 创建转账订单

**请求**
```http
POST /api/orders/transfer
Authorization: Bearer {token}
Content-Type: application/json

{
  "customer_id": 1,
  "amount": 500.00
}
```

**响应**
```json
{
  "code": 0,
  "message": "转账成功",
  "data": {
    "id": 2,
    "order_no": "TRF20251117efgh5678",
    "order_type": "transfer",
    "customer_id": 1,
    "amount": 500.00,
    "status": "completed",
    "created_at": "2025-11-17T11:00:00Z"
  }
}
```

**业务说明**
- 从客户的资金账户转账到消耗账户
- 转账完成后，消耗账户余额可用于广告投放
- 转账会扣减资金账户余额，增加消耗账户余额

### 3.5 获取订单列表

**请求**
```http
GET /api/orders?page=1&size=20&order_type=recharge&status=paid&customer_id=1
Authorization: Bearer {token}
```

**参数说明**
- `page`: 页码
- `size`: 每页数量
- `order_type`: 订单类型（recharge/transfer/settlement）
- `status`: 订单状态（pending/paid/verified/completed/failed）
- `customer_id`: 客户ID

**响应**
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "order_no": "RCH20251117abcd1234",
      "order_type": "recharge",
      "amount": 1000.00,
      "status": "paid",
      "customer": {
        "id": 1,
        "customer_name": "测试科技有限公司"
      },
      "payment_time": "2025-11-17T10:05:00Z",
      "created_at": "2025-11-17T10:00:00Z"
    }
  ],
  "total": 50,
  "page": 1,
  "size": 20
}
```

### 3.6 获取订单详情

**请求**
```http
GET /api/orders/1
Authorization: Bearer {token}
```

**响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "order_no": "RCH20251117abcd1234",
    "order_type": "recharge",
    "amount": 1000.00,
    "status": "completed",
    "customer": {
      "id": 1,
      "customer_name": "测试科技有限公司"
    },
    "payment_method": "支付宝",
    "payment_time": "2025-11-17T10:05:00Z",
    "verified_time": "2025-11-17T10:30:00Z",
    "status_logs": [
      {
        "id": 1,
        "from_status": "",
        "to_status": "pending",
        "notes": "创建充值订单",
        "created_at": "2025-11-17T10:00:00Z"
      },
      {
        "id": 2,
        "from_status": "pending",
        "to_status": "paid",
        "notes": "支付完成",
        "created_at": "2025-11-17T10:05:00Z"
      },
      {
        "id": 3,
        "from_status": "paid",
        "to_status": "verified",
        "notes": "财务审核通过",
        "created_at": "2025-11-17T10:30:00Z"
      },
      {
        "id": 4,
        "from_status": "verified",
        "to_status": "completed",
        "notes": "余额入账完成",
        "created_at": "2025-11-17T10:30:01Z"
      }
    ]
  }
}
```

## 完整业务流程示例

### 场景：客户充值并使用余额

```bash
# 1. 登录获取token
TOKEN=$(curl -s -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r '.data.token')

# 2. 创建客户
CUSTOMER_ID=$(curl -s -X POST http://localhost:8080/api/customers \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"customer_name":"测试公司","phone":"13800138000"}' | jq -r '.data.id')

# 3. 创建充值订单
ORDER_NO=$(curl -s -X POST http://localhost:8080/api/orders/recharge \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"customer_id\":$CUSTOMER_ID,\"amount\":1000.00}" | jq -r '.data.order_no')

# 4. 模拟支付回调
curl -X POST http://localhost:8080/api/payment/callback \
  -H "Content-Type: application/json" \
  -d "{\"order_no\":\"$ORDER_NO\",\"transaction_id\":\"TX123456\"}"

# 5. 获取订单ID
ORDER_ID=$(curl -s -X GET "http://localhost:8080/api/orders?order_type=recharge&status=paid" \
  -H "Authorization: Bearer $TOKEN" | jq -r '.data[0].id')

# 6. 财务审核通过
curl -X POST http://localhost:8080/api/orders/$ORDER_ID/verify \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"approved":true,"notes":"审核通过"}'

# 7. 转账到消耗账户
curl -X POST http://localhost:8080/api/orders/transfer \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"customer_id\":$CUSTOMER_ID,\"amount\":500.00}"

# 8. 查看客户账户余额
curl -X GET http://localhost:8080/api/customers/$CUSTOMER_ID/accounts \
  -H "Authorization: Bearer $TOKEN"
```

## 错误码说明

| 错误码 | 说明 |
|-------|------|
| 0 | 成功 |
| 400 | 参数错误 |
| 401 | 未认证 |
| 403 | 无权限 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |
| 501 | 数据库错误 |
| 502 | 验证错误 |

## 注意事项

1. **Token过期**: 默认24小时过期，需要重新登录
2. **权限控制**: 不同角色拥有不同权限，调用接口时会验证
3. **租户隔离**: 非平台管理员只能访问自己租户的数据
4. **金额精度**: 金额字段保留2位小数
5. **状态流转**: 订单状态流转有严格的顺序，不能跨状态变更
