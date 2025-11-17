# 客户管理+收款系统

一个功能完整的多租户SaaS客户管理与收款系统，基于Go、PostgreSQL和纯HTML5构建。

## 系统特性

### 核心功能

1. **多租户架构**
   - 完整的租户隔离机制
   - 支持创建和管理多个公司（租户）
   - 每个租户独立的数据和权限体系

2. **角色权限管理**
   - 平台角色：超级管理员及其他平台管理角色
   - 租户角色：租户自定义的管理角色
   - 基于RBAC的权限控制
   - 部门权限继承机制

3. **客户管理**
   - 客户信息的增删改查
   - 批量导入/导出客户数据
   - 合同归档管理
   - 自定义字段支持

4. **订单管理**
   - 充值订单：客户充值流程
   - 结算订单：业务结算
   - 转账订单：账户间转账
   - 完整的订单状态流转

5. **财务管理**
   - 充值订单审核
   - 发票管理（开具、查询、作废）
   - 账户余额管理（资金账户、消耗账户）

6. **订单流转流程**
   ```
   客户发起充值
   → 生成充值订单（pending）
   → 支付完成（paid）
   → 财务审核（verified）
   → 余额入账（completed）
   → 业务员转账（资金账户 → 消耗账户）
   → 客户使用广告投放余额
   ```

7. **账户体系**
   - 资金账户：充值后的资金存放
   - 消耗账户：可用于广告投放的余额
   - 现金余额和总余额分离管理

8. **表单管理**
   - 动态创建自定义表单
   - 表单数据存储和查询

9. **系统设置**
   - 安全设置
   - 通知设置
   - 登录设置

## 技术栈

- **后端**: Go 1.25+
- **Web框架**: Gin 1.11+
- **数据库**: PostgreSQL 18+
- **ORM**: GORM
- **认证**: JWT
- **密码加密**: bcrypt
- **前端**: 纯HTML5 + CSS + JavaScript

## 项目结构

```
cjadmin/
├── cmd/
│   └── api/
│       └── main.go              # 主程序入口
├── internal/
│   ├── config/                  # 配置管理
│   │   └── config.go
│   ├── database/                # 数据库连接
│   │   └── database.go
│   ├── middleware/              # 中间件
│   │   ├── auth.go             # 认证中间件
│   │   ├── tenant.go           # 多租户中间件
│   │   ├── permission.go       # 权限中间件
│   │   └── cors.go             # CORS中间件
│   ├── models/                  # 数据模型
│   │   ├── base.go
│   │   ├── tenant.go
│   │   ├── user.go
│   │   ├── customer.go
│   │   ├── order.go
│   │   ├── account.go
│   │   ├── finance.go
│   │   ├── form.go
│   │   └── system.go
│   ├── repository/              # 数据访问层（预留）
│   ├── service/                 # 业务逻辑层
│   │   ├── auth.go
│   │   ├── customer.go
│   │   └── order.go
│   ├── handler/                 # HTTP处理器
│   │   ├── auth.go
│   │   ├── customer.go
│   │   └── order.go
│   └── utils/                   # 工具函数
│       ├── jwt.go
│       ├── password.go
│       └── response.go
├── migrations/                  # 数据库迁移文件
│   └── 001_init_schema.sql
├── web/                         # 前端文件
│   ├── static/                  # 静态资源
│   │   ├── css/
│   │   ├── js/
│   │   └── images/
│   └── templates/               # HTML模板
│       ├── index.html
│       ├── login.html
│       └── dashboard.html
├── docs/                        # 文档
│   └── database_design.md
├── config.yaml                  # 配置文件
├── go.mod                       # Go模块文件
└── README.md                    # 本文件
```

## 快速开始

### 环境要求

- Go 1.25 或更高版本
- PostgreSQL 18 或更高版本
- Git

### 安装步骤

1. **克隆项目**
   ```bash
   git clone https://github.com/anttna7/cjadmin.git
   cd cjadmin
   ```

2. **安装依赖**
   ```bash
   go mod download
   ```

3. **配置数据库**

   创建PostgreSQL数据库：
   ```bash
   createdb cjadmin
   ```

   执行数据库迁移：
   ```bash
   psql -U postgres -d cjadmin -f migrations/001_init_schema.sql
   ```

4. **修改配置文件**

   编辑 `config.yaml`，修改数据库连接信息：
   ```yaml
   database:
     host: "localhost"
     port: 5432
     user: "postgres"
     password: "your_password"
     dbname: "cjadmin"
   ```

5. **创建超级管理员**
   ```bash
   go run cmd/api/main.go -init-admin -admin-user=admin -admin-pass=admin123
   ```

6. **启动服务**
   ```bash
   go run cmd/api/main.go
   ```

7. **访问系统**

   打开浏览器访问：http://localhost:8080

   默认管理员账号：
   - 用户名: admin
   - 密码: admin123

## API文档

### 认证接口

#### 登录
```http
POST /api/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}

响应：
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "username": "admin",
      "is_platform_admin": true
    }
  }
}
```

### 客户管理接口

#### 创建客户
```http
POST /api/customers
Authorization: Bearer {token}
Content-Type: application/json

{
  "customer_name": "测试客户",
  "customer_type": "企业",
  "phone": "13800138000",
  "email": "test@example.com"
}
```

#### 获取客户列表
```http
GET /api/customers?page=1&size=20
Authorization: Bearer {token}
```

#### 获取客户详情
```http
GET /api/customers/{id}
Authorization: Bearer {token}
```

#### 更新客户
```http
PUT /api/customers/{id}
Authorization: Bearer {token}
Content-Type: application/json

{
  "customer_name": "新客户名称"
}
```

#### 删除客户
```http
DELETE /api/customers/{id}
Authorization: Bearer {token}
```

### 订单管理接口

#### 创建充值订单
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

#### 财务审核订单
```http
POST /api/orders/{id}/verify
Authorization: Bearer {token}
Content-Type: application/json

{
  "approved": true,
  "notes": "审核通过"
}
```

#### 创建转账订单
```http
POST /api/orders/transfer
Authorization: Bearer {token}
Content-Type: application/json

{
  "customer_id": 1,
  "amount": 500.00
}
```

#### 获取订单列表
```http
GET /api/orders?page=1&size=20&order_type=recharge&status=paid
Authorization: Bearer {token}
```

## 数据库设计

详细的数据库设计文档请查看：[docs/database_design.md](docs/database_design.md)

### 核心表

- **tenants**: 租户表
- **users**: 用户表
- **roles**: 角色表
- **permissions**: 权限表
- **departments**: 部门表
- **customers**: 客户表
- **orders**: 订单表
- **customer_accounts**: 客户账户表
- **account_transactions**: 账户交易明细表
- **invoices**: 发票表

## 权限系统

系统实现了完整的RBAC权限控制：

1. **平台角色**：由平台管理员管理，全局有效
2. **租户角色**：由租户管理员管理，仅在租户内有效
3. **部门权限**：部门可以设置权限，同一部门的用户自动继承
4. **权限继承**：子部门继承父部门的权限

### 内置权限

- customer.create/read/update/delete/import/export
- order.create/read/update/delete/import/export
- finance.verify
- invoice.create/read/update/delete
- role.create/read/update/delete
- user.create/read/update/delete

## 订单状态流转

### 充值订单
```
pending → paid → verified → completed
                    ↓
                  failed
```

### 转账订单
```
pending → processing → completed
                    ↓
                  failed
```

## 部署说明

### 使用Docker部署

1. 创建 `Dockerfile`:
```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main cmd/api/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/config.yaml .
COPY --from=builder /app/web ./web
EXPOSE 8080
CMD ["./main"]
```

2. 构建镜像：
```bash
docker build -t cjadmin:latest .
```

3. 运行容器：
```bash
docker run -d -p 8080:8080 \
  -e DB_HOST=postgres \
  -e DB_PASSWORD=your_password \
  cjadmin:latest
```

### 生产环境建议

1. 使用环境变量管理敏感配置
2. 启用HTTPS
3. 配置Nginx反向代理
4. 使用PostgreSQL主从复制
5. 配置定期数据备份
6. 启用日志监控和告警

## 开发说明

### 添加新功能

1. 在 `internal/models/` 添加数据模型
2. 在 `migrations/` 添加数据库迁移
3. 在 `internal/service/` 实现业务逻辑
4. 在 `internal/handler/` 添加HTTP处理器
5. 在 `cmd/api/main.go` 注册路由

### 代码规范

- 遵循Go官方代码规范
- 使用gofmt格式化代码
- 变量和函数使用驼峰命名
- 添加必要的注释

## 许可证

MIT License

## 联系方式

- 项目地址：https://github.com/anttna7/cjadmin
- 问题反馈：https://github.com/anttna7/cjadmin/issues

## 更新日志

### v1.0.0 (2025-11-17)

- 初始版本发布
- 实现多租户架构
- 完成角色权限管理
- 实现客户管理功能
- 实现订单管理和流转
- 实现账户余额管理
- 完成基础前端页面
