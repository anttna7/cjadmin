# CJAdmin 快速开始指南

欢迎使用CJAdmin客户管理+收款系统！本指南将帮助您在5分钟内快速启动并体验系统。

## 📋 目录

- [系统要求](#系统要求)
- [快速启动](#快速启动)
- [首次登录](#首次登录)
- [核心功能体验](#核心功能体验)
- [下一步](#下一步)

---

## 系统要求

### 最小配置

- **Docker**: 20.10+
- **Docker Compose**: 2.0+
- **内存**: 2GB+
- **磁盘**: 10GB+
- **操作系统**: Linux / macOS / Windows (with WSL2)

### 推荐配置

- **CPU**: 4核+
- **内存**: 4GB+
- **磁盘**: 20GB+ SSD
- **网络**: 10Mbps+

---

## 快速启动

### 方法一：Docker Compose（推荐）⚡

**只需一条命令即可启动整个系统！**

```bash
# 1. 克隆项目
git clone https://github.com/anttna7/cjadmin.git
cd cjadmin

# 2. 启动所有服务（应用+数据库+Redis+Nginx）
docker-compose up -d

# 3. 等待服务启动（约30秒）
docker-compose ps

# 4. 初始化数据库
docker-compose exec app /bin/sh -c "psql postgresql://postgres:postgres@postgres:5432/cjadmin < /app/migrations/001_init_schema.sql"

# 5. 访问系统
# 浏览器打开 http://localhost
```

**就这么简单！** 🎉

### 方法二：本地开发模式

如果你想在本地开发环境运行：

```bash
# 1. 安装依赖
go mod download

# 2. 启动数据库（使用Docker）
docker run -d --name cjadmin-postgres \
  -e POSTGRES_DB=cjadmin \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 \
  postgres:18-alpine

# 3. 初始化数据库
psql -h localhost -U postgres -d cjadmin -f migrations/001_init_schema.sql

# 4. 配置环境（可选，使用默认配置）
cp config.example.yaml config.yaml

# 5. 启动应用
go run cmd/api/main.go

# 6. 访问系统
# 浏览器打开 http://localhost:8080
```

---

## 首次登录

### 默认管理员账号

系统会自动创建超级管理员账号：

- **用户名**: `admin`
- **密码**: `admin123`
- **角色**: 平台超级管理员

### 登录步骤

1. 浏览器访问 `http://localhost` (Docker) 或 `http://localhost:8080` (本地)
2. 进入登录页面
3. 输入用户名和密码
4. 点击"登录"按钮

### 首次登录建议

⚠️ **安全提示**：首次登录后，请立即修改默认密码！

```bash
# 使用API修改密码
curl -X POST http://localhost/api/auth/change-password \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "old_password": "admin123",
    "new_password": "your_secure_password"
  }'
```

---

## 核心功能体验

### 1. 创建租户（多租户管理）

```bash
# 创建一个新租户
curl -X POST http://localhost/api/tenants \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_name": "测试公司",
    "contact_person": "张三",
    "contact_phone": "13800138000",
    "contact_email": "test@example.com",
    "status": "active"
  }'
```

### 2. 创建客户

```bash
# 创建客户
curl -X POST http://localhost/api/customers \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_name": "ABC科技有限公司",
    "customer_type": "enterprise",
    "industry": "IT",
    "source": "website",
    "level": "A",
    "country": "China",
    "province": "Beijing",
    "city": "Beijing",
    "address": "朝阳区xxx大厦",
    "website": "https://abc.com"
  }'
```

### 3. 创建订单

```bash
# 创建充值订单
curl -X POST http://localhost/api/orders \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": 1,
    "order_type": "recharge",
    "amount": 10000.00,
    "payment_method": "bank_transfer",
    "description": "首次充值"
  }'
```

### 4. 查看订单列表

```bash
# 获取订单列表
curl http://localhost/api/orders?page=1&page_size=10 \
  -H "Authorization: Bearer <your_token>"
```

### 5. 访问仪表板

浏览器访问：
- **仪表板**: http://localhost/
- **API文档**: http://localhost/api/docs (如已配置Swagger)

---

## 监控系统（可选）

想要监控系统运行状态？启动监控栈：

```bash
# 启动Prometheus + Grafana + Alertmanager
cd monitoring
docker-compose -f docker-compose.monitoring.yml up -d

# 访问监控界面
# Grafana: http://localhost:3000 (admin/admin123)
# Prometheus: http://localhost:9090
# Alertmanager: http://localhost:9093
```

---

## 常用命令

### Docker操作

```bash
# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f app

# 重启服务
docker-compose restart app

# 停止所有服务
docker-compose down

# 停止并删除数据
docker-compose down -v
```

### 数据库操作

```bash
# 进入数据库
docker-compose exec postgres psql -U postgres -d cjadmin

# 备份数据库
docker-compose exec postgres pg_dump -U postgres cjadmin > backup.sql

# 恢复数据库
docker-compose exec -T postgres psql -U postgres -d cjadmin < backup.sql
```

### 应用操作

```bash
# 使用Makefile（更简单）
make build        # 编译项目
make run          # 运行项目
make test         # 运行测试
make test-coverage # 测试覆盖率报告
```

---

## 故障排查

### 问题1：端口被占用

```bash
# 检查端口占用
lsof -i :8080  # Linux/Mac
netstat -ano | findstr :8080  # Windows

# 修改端口（编辑docker-compose.yml或config.yaml）
ports:
  - "8081:8080"  # 将外部端口改为8081
```

### 问题2：数据库连接失败

```bash
# 检查PostgreSQL是否运行
docker-compose ps postgres

# 查看数据库日志
docker-compose logs postgres

# 重启数据库
docker-compose restart postgres
```

### 问题3：服务无响应

```bash
# 检查应用日志
docker-compose logs app

# 检查健康状态
curl http://localhost/api/health

# 重启应用
docker-compose restart app
```

### 问题4：权限错误

```bash
# 检查文件权限
ls -la uploads/

# 修复权限
chmod -R 755 uploads/
chown -R $(whoami) uploads/
```

---

## API快速参考

### 认证

```bash
# 登录
POST /api/auth/login
{
  "username": "admin",
  "password": "admin123"
}

# 获取当前用户信息
GET /api/auth/me
Authorization: Bearer <token>
```

### 租户管理

```bash
# 创建租户
POST /api/tenants

# 获取租户列表
GET /api/tenants?page=1&page_size=10

# 获取租户详情
GET /api/tenants/:id

# 更新租户
PUT /api/tenants/:id

# 删除租户
DELETE /api/tenants/:id
```

### 客户管理

```bash
# 创建客户
POST /api/customers

# 获取客户列表
GET /api/customers?page=1&page_size=10

# 搜索客户
GET /api/customers/search?keyword=ABC

# 导出客户
GET /api/customers/export

# 导入客户
POST /api/customers/import
```

### 订单管理

```bash
# 创建充值订单
POST /api/orders

# 获取订单列表
GET /api/orders?page=1&page_size=10&order_type=recharge

# 审核订单
PUT /api/orders/:id/approve

# 获取订单统计
GET /api/orders/statistics
```

### 财务管理

```bash
# 创建发票
POST /api/invoices

# 获取发票列表
GET /api/invoices?page=1&page_size=10

# 开具发票
PUT /api/invoices/:id/issue

# 获取财务统计
GET /api/payments/statistics
```

完整API文档请参考：[API文档](./docs/api_documentation.md)

---

## 测试数据

系统提供测试数据生成脚本（可选）：

```bash
# 运行测试数据生成脚本（需要先实现）
# go run scripts/generate_test_data.go

# 或手动创建测试数据
# 1. 创建3个租户
# 2. 每个租户创建10个客户
# 3. 每个客户创建5个订单
# 4. 生成财务数据
```

---

## 下一步

恭喜！您已经成功启动CJAdmin系统。接下来您可以：

### 📚 深入学习

1. **阅读用户手册** - [用户操作手册](./docs/user_manual.md)
2. **阅读管理员手册** - [管理员手册](./docs/admin_manual.md)
3. **了解系统架构** - [数据库设计](./docs/database_design.md)
4. **学习API接口** - [API文档](./docs/api_documentation.md)

### 🔧 系统配置

1. **配置邮件通知** - 编辑 `config.yaml` 中的邮件配置
2. **配置支付接口** - 集成微信支付/支付宝
3. **配置短信服务** - 集成阿里云短信/腾讯云短信
4. **自定义表单** - 创建业务相关的自定义表单

### 🚀 部署生产环境

1. **生产环境部署** - [部署文档](./DEPLOYMENT.md)
2. **监控告警配置** - [监控指南](./docs/monitoring_guide.md)
3. **安全加固** - [管理员手册-安全管理](./docs/admin_manual.md#安全管理)
4. **性能优化** - [管理员手册-性能优化](./docs/admin_manual.md#性能优化)

### 🧪 开发和测试

1. **运行测试** - [测试指南](./docs/testing_guide.md)
2. **贡献代码** - [开发规划](./docs/development_plan.md)
3. **代码规范** - 遵循Go最佳实践
4. **提交Issue** - [GitHub Issues](https://github.com/anttna7/cjadmin/issues)

---

## 获取帮助

### 文档资源

- **README** - [README.md](./README.md)
- **开发规划** - [development_plan.md](./docs/development_plan.md)
- **进度报告** - [progress_report.md](./docs/progress_report.md)
- **部署指南** - [DEPLOYMENT.md](./DEPLOYMENT.md)

### 社区支持

- **GitHub Issues**: https://github.com/anttna7/cjadmin/issues
- **项目Wiki**: https://github.com/anttna7/cjadmin/wiki
- **邮件联系**: support@example.com

### 常见问题

更多常见问题请参考：
- [用户手册-常见问题](./docs/user_manual.md#常见问题)
- [管理员手册-故障排查](./docs/admin_manual.md#故障排查)

---

## 反馈与建议

我们非常重视您的反馈！如果您有任何问题或建议，请：

1. 提交 [GitHub Issue](https://github.com/anttna7/cjadmin/issues)
2. 发送邮件至：support@example.com
3. 加入社区讨论

---

**祝您使用愉快！** 🎉

---

© 2025 CJAdmin Team. All rights reserved.
