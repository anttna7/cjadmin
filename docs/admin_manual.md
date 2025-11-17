# 管理员手册

## 目录

1. [系统管理概述](#系统管理概述)
2. [租户管理](#租户管理)
3. [用户和权限管理](#用户和权限管理)
4. [部门管理](#部门管理)
5. [角色管理](#角色管理)
6. [系统配置](#系统配置)
7. [审计日志](#审计日志)
8. [数据备份与恢复](#数据备份与恢复)
9. [性能优化](#性能优化)
10. [故障排查](#故障排查)
11. [安全管理](#安全管理)

---

## 系统管理概述

### 管理员类型

#### 1. 平台管理员 (Platform Admin)
- **权限范围**：所有租户和平台配置
- **主要职责**：
  - 创建和管理租户
  - 配置平台级设置
  - 监控系统运行状态
  - 处理跨租户问题

#### 2. 租户管理员 (Tenant Admin)
- **权限范围**：本租户内的所有数据
- **主要职责**：
  - 管理租户用户
  - 配置租户权限
  - 管理业务数据
  - 生成业务报表

### 管理员初始化

#### 创建第一个超级管理员

```bash
curl -X POST http://localhost:8080/api/auth/create-admin \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "Admin@123456"
  }'
```

⚠️ **安全建议**：
- 首次登录后立即修改密码
- 使用强密码（至少12位，包含大小写字母、数字、特殊字符）
- 启用双因素认证（如已实现）
- 定期更换密码（建议90天）

---

## 租户管理

### 创建租户

**权限要求**：平台管理员

**步骤：**

```bash
curl -X POST http://localhost:8080/api/tenants \
  -H "Authorization: Bearer platform_admin_token" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_name": "示例企业",
    "tenant_code": "demo001",
    "contact_person": "张三",
    "contact_phone": "13800138000",
    "contact_email": "admin@example.com",
    "address": "北京市朝阳区"
  }'
```

**系统自动操作：**
1. 创建租户记录
2. 生成默认角色（管理员、普通用户）
3. 初始化租户配置
4. 创建独立数据空间

### 租户状态管理

**租户状态：**
- `active` - 正常运行
- `suspended` - 暂停（临时）
- `inactive` - 停用（长期）

**更新租户状态：**

```bash
# 暂停租户
curl -X PUT http://localhost:8080/api/tenants/1/status \
  -H "Authorization: Bearer platform_admin_token" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "suspended"
  }'

# 激活租户
curl -X PUT http://localhost:8080/api/tenants/1/status \
  -H "Authorization: Bearer platform_admin_token" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "active"
  }'
```

### 租户配额管理

**配额类型：**
- 用户数量限制
- 存储空间限制
- API请求频率限制
- 并发连接数限制

**设置配额：**

```bash
curl -X PUT http://localhost:8080/api/tenants/1 \
  -H "Authorization: Bearer platform_admin_token" \
  -H "Content-Type: application/json" \
  -d '{
    "max_users": 100,
    "max_storage": 10737418240,
    "max_api_calls": 10000
  }'
```

### 租户统计信息

```bash
curl -X GET http://localhost:8080/api/tenants/1/statistics \
  -H "Authorization: Bearer platform_admin_token"
```

**返回信息：**
- 用户总数
- 客户总数
- 订单总数
- 存储使用量
- API调用次数

---

## 用户和权限管理

### 用户生命周期管理

#### 1. 创建用户

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Authorization: Bearer admin_token" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "user001",
    "password": "User@123456",
    "email": "user001@example.com",
    "phone": "13900139000",
    "department_id": 1
  }'
```

#### 2. 禁用用户

```bash
curl -X PUT http://localhost:8080/api/users/1/status \
  -H "Authorization: Bearer admin_token" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "inactive"
  }'
```

#### 3. 重置用户密码

```bash
curl -X POST http://localhost:8080/api/users/1/reset-password \
  -H "Authorization: Bearer admin_token" \
  -H "Content-Type: application/json" \
  -d '{
    "new_password": "NewPassword@123"
  }'
```

### 权限体系

#### 权限结构

```
用户权限 = 角色权限 ∪ 部门权限

其中：
- 角色权限：通过分配角色获得
- 部门权限：从部门及其父部门继承
```

#### 权限列表

**用户管理（user）**
- `user.create` - 创建用户
- `user.read` - 查看用户
- `user.update` - 更新用户
- `user.delete` - 删除用户

**客户管理（customer）**
- `customer.create` - 创建客户
- `customer.read` - 查看客户
- `customer.update` - 更新客户
- `customer.delete` - 删除客户

**订单管理（order）**
- `order.create` - 创建订单
- `order.read` - 查看订单
- `order.update` - 更新订单
- `order.delete` - 删除订单

**财务管理（finance）**
- `finance.create` - 创建财务记录
- `finance.read` - 查看财务数据
- `finance.update` - 更新财务记录
- `finance.delete` - 删除财务记录
- `finance.verify` - 财务审核

**系统管理（system）**
- `system.create` - 创建系统配置
- `system.read` - 查看系统配置
- `system.update` - 更新系统配置
- `system.delete` - 删除系统配置

### 查看用户权限

```bash
# 查看用户的所有权限
curl -X GET http://localhost:8080/api/auth/permissions \
  -H "Authorization: Bearer user_token"

# 查看特定用户的权限（管理员）
curl -X GET http://localhost:8080/api/users/1/permissions \
  -H "Authorization: Bearer admin_token"
```

---

## 部门管理

### 部门树形结构

**示例结构：**
```
公司总部
├── 销售部
│   ├── 华东销售组
│   └── 华南销售组
├── 技术部
│   ├── 研发组
│   └── 运维组
└── 财务部
```

### 创建部门

```bash
curl -X POST http://localhost:8080/api/departments \
  -H "Authorization: Bearer admin_token" \
  -H "Content-Type: application/json" \
  -d '{
    "department_name": "销售部",
    "parent_id": null,
    "description": "负责销售业务"
  }'
```

### 创建子部门

```bash
curl -X POST http://localhost:8080/api/departments \
  -H "Authorization: Bearer admin_token" \
  -H "Content-Type: application/json" \
  -d '{
    "department_name": "华东销售组",
    "parent_id": 1,
    "description": "负责华东地区销售"
  }'
```

### 获取部门树

```bash
curl -X GET http://localhost:8080/api/departments/tree \
  -H "Authorization: Bearer admin_token"
```

### 设置部门权限

```bash
curl -X POST http://localhost:8080/api/departments/1/permissions \
  -H "Authorization: Bearer admin_token" \
  -H "Content-Type: application/json" \
  -d '{
    "permission_ids": [1, 2, 3, 4, 5]
  }'
```

**权限继承规则：**
- 子部门自动继承父部门的所有权限
- 可以为子部门添加额外权限
- 不能移除继承的权限

### 查看部门权限（含继承）

```bash
curl -X GET http://localhost:8080/api/departments/1/permissions/inherited \
  -H "Authorization: Bearer admin_token"
```

---

## 角色管理

### 角色类型

#### 1. 平台角色
- 作用域：所有租户
- 只能由平台管理员创建
- 用于跨租户管理

#### 2. 租户角色
- 作用域：单个租户
- 租户管理员可创建
- 权限不能超过创建者

### 创建平台角色

```bash
curl -X POST http://localhost:8080/api/roles/platform \
  -H "Authorization: Bearer platform_admin_token" \
  -H "Content-Type: application/json" \
  -d '{
    "role_name": "平台运维",
    "role_code": "platform_ops",
    "description": "负责平台运维工作"
  }'
```

### 创建租户角色

```bash
curl -X POST http://localhost:8080/api/roles/tenant \
  -H "Authorization: Bearer tenant_admin_token" \
  -H "Content-Type: application/json" \
  -d '{
    "role_name": "销售经理",
    "role_code": "sales_manager",
    "description": "销售部门经理"
  }'
```

### 为角色分配权限

```bash
curl -X POST http://localhost:8080/api/roles/1/permissions \
  -H "Authorization: Bearer admin_token" \
  -H "Content-Type: application/json" \
  -d '{
    "permission_ids": [1, 2, 3, 4, 5, 6, 7, 8]
  }'
```

### 为用户分配角色

```bash
# 分配单个角色
curl -X POST http://localhost:8080/api/user-roles/1/assign \
  -H "Authorization: Bearer admin_token" \
  -H "Content-Type: application/json" \
  -d '{
    "role_id": 1
  }'

# 批量分配角色
curl -X PUT http://localhost:8080/api/user-roles/1/roles/batch \
  -H "Authorization: Bearer admin_token" \
  -H "Content-Type: application/json" \
  -d '{
    "role_ids": [1, 2, 3]
  }'
```

### 查看角色权限

```bash
curl -X GET http://localhost:8080/api/roles/1/permissions \
  -H "Authorization: Bearer admin_token"
```

---

## 系统配置

### 配置类型

#### 1. 租户级配置
- 作用域：单个租户
- 租户管理员可修改
- 示例：企业信息、业务规则

#### 2. 平台级配置
- 作用域：整个平台
- 只有平台管理员可修改
- 示例：系统参数、全局设置

### 设置租户配置

```bash
curl -X POST http://localhost:8080/api/settings \
  -H "Authorization: Bearer tenant_admin_token" \
  -H "Content-Type: application/json" \
  -d '{
    "setting_key": "company.info",
    "setting_value": {
      "name": "示例科技有限公司",
      "address": "北京市朝阳区",
      "phone": "010-12345678",
      "email": "contact@example.com"
    },
    "description": "公司基本信息"
  }'
```

### 设置平台配置

```bash
curl -X POST http://localhost:8080/api/platform/settings \
  -H "Authorization: Bearer platform_admin_token" \
  -H "Content-Type: application/json" \
  -d '{
    "setting_key": "system.maintenance",
    "setting_value": {
      "enabled": false,
      "message": "系统维护中，预计1小时后恢复"
    },
    "description": "系统维护模式"
  }'
```

### 获取配置

```bash
# 获取租户配置
curl -X GET http://localhost:8080/api/settings/company.info \
  -H "Authorization: Bearer tenant_admin_token"

# 获取平台配置
curl -X GET http://localhost:8080/api/platform/settings/system.maintenance \
  -H "Authorization: Bearer platform_admin_token"
```

### 配置分类查询

```bash
curl -X GET "http://localhost:8080/api/settings?category=company" \
  -H "Authorization: Bearer tenant_admin_token"
```

---

## 审计日志

### 审计日志类型

1. **用户操作日志**：登录、登出、权限变更
2. **数据变更日志**：创建、更新、删除记录
3. **系统事件日志**：配置变更、系统错误
4. **安全事件日志**：登录失败、权限拒绝

### 查询审计日志

```bash
# 查询所有日志
curl -X GET "http://localhost:8080/api/audit-logs?page=1&size=20" \
  -H "Authorization: Bearer admin_token"

# 按用户查询
curl -X GET "http://localhost:8080/api/audit-logs?user_id=1" \
  -H "Authorization: Bearer admin_token"

# 按操作类型查询
curl -X GET "http://localhost:8080/api/audit-logs?action=create" \
  -H "Authorization: Bearer admin_token"

# 按时间范围查询
curl -X GET "http://localhost:8080/api/audit-logs?start_date=2024-01-01&end_date=2024-01-31" \
  -H "Authorization: Bearer admin_token"
```

### 审计统计

```bash
# 获取最近30天统计
curl -X GET "http://localhost:8080/api/audit-logs/statistics?days=30" \
  -H "Authorization: Bearer admin_token"
```

**统计指标：**
- 总操作数
- 按操作类型统计
- 按资源类型统计
- 活跃用户数

### 清理旧日志

```bash
# 清理30天前的日志
curl -X POST http://localhost:8080/api/audit-logs/clean \
  -H "Authorization: Bearer admin_token" \
  -H "Content-Type: application/json" \
  -d '{
    "days_to_keep": 30
  }'
```

⚠️ **建议**：
- 生产环境至少保留90天日志
- 定期导出重要日志
- 设置自动清理任务

---

## 数据备份与恢复

### 数据库备份

#### 手动备份

```bash
# Docker环境
docker-compose exec postgres pg_dump -U postgres cjadmin | gzip > backup_$(date +%Y%m%d_%H%M%S).sql.gz

# 本地环境
pg_dump -U postgres -d cjadmin | gzip > backup_$(date +%Y%m%d_%H%M%S).sql.gz
```

#### 自动备份脚本

创建 `backup.sh`:
```bash
#!/bin/bash
BACKUP_DIR="/backups"
DATE=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS=7

# 创建备份目录
mkdir -p "$BACKUP_DIR"

# 备份数据库
docker-compose exec -T postgres pg_dump -U postgres cjadmin | \
  gzip > "$BACKUP_DIR/db_$DATE.sql.gz"

# 备份上传文件
tar -czf "$BACKUP_DIR/uploads_$DATE.tar.gz" uploads/

# 备份配置文件
tar -czf "$BACKUP_DIR/config_$DATE.tar.gz" config.yaml .env

# 删除旧备份
find "$BACKUP_DIR" -name "*.gz" -mtime +$RETENTION_DAYS -delete

echo "Backup completed: $DATE"
```

#### 设置定时备份

```bash
# 编辑crontab
crontab -e

# 每天凌晨2点执行备份
0 2 * * * /path/to/backup.sh >> /var/log/backup.log 2>&1
```

### 数据恢复

#### 恢复数据库

```bash
# Docker环境
gunzip < backup_20240117_020000.sql.gz | \
  docker-compose exec -T postgres psql -U postgres cjadmin

# 本地环境
gunzip < backup_20240117_020000.sql.gz | \
  psql -U postgres -d cjadmin
```

#### 恢复文件

```bash
# 恢复上传文件
tar -xzf uploads_20240117_020000.tar.gz

# 恢复配置文件
tar -xzf config_20240117_020000.tar.gz
```

### 灾难恢复计划

1. **数据备份**：
   - 每日全量备份
   - 每小时增量备份（如需要）
   - 异地存储备份文件

2. **恢复时间目标(RTO)**：< 4小时
3. **恢复点目标(RPO)**：< 24小时

4. **恢复步骤**：
   ```bash
   # 1. 停止应用
   docker-compose down

   # 2. 恢复数据库
   gunzip < latest_backup.sql.gz | docker-compose exec -T postgres psql -U postgres cjadmin

   # 3. 恢复文件
   tar -xzf latest_uploads.tar.gz

   # 4. 启动应用
   docker-compose up -d

   # 5. 验证恢复
   curl http://localhost:8080/api/health
   ```

---

## 性能优化

### 数据库优化

#### 1. 索引优化

```sql
-- 查看缺失索引
SELECT schemaname, tablename, attname
FROM pg_stats
WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
  AND n_distinct > 0.01
  AND null_frac < 0.5
ORDER BY n_distinct DESC;

-- 查看未使用的索引
SELECT schemaname, tablename, indexname
FROM pg_stat_user_indexes
WHERE idx_scan = 0
  AND indexrelname NOT LIKE 'pg_toast%';
```

#### 2. 查询优化

```sql
-- 查看慢查询
SELECT query, calls, total_time, mean_time
FROM pg_stat_statements
ORDER BY total_time DESC
LIMIT 10;

-- 启用查询分析
EXPLAIN ANALYZE SELECT * FROM customers WHERE customer_name LIKE '%关键词%';
```

#### 3. 连接池配置

修改 `config.yaml`:
```yaml
database:
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 3600
```

### 应用优化

#### 1. 启用Redis缓存

修改 `docker-compose.yml`，启用Redis服务。

缓存策略：
- 用户权限信息：缓存24小时
- 系统配置：缓存1小时
- 热点数据：缓存10分钟

#### 2. API限流

配置Nginx限流：
```nginx
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;

location /api/ {
    limit_req zone=api_limit burst=20 nodelay;
    proxy_pass http://app_backend;
}
```

#### 3. 启用Gzip压缩

Nginx配置已包含Gzip，确保启用：
```nginx
gzip on;
gzip_min_length 1024;
gzip_types text/plain application/json;
```

### 监控指标

#### 系统指标
- CPU使用率 < 70%
- 内存使用率 < 80%
- 磁盘使用率 < 85%
- 网络流量

#### 应用指标
- API响应时间 < 200ms (P95)
- 错误率 < 0.1%
- 并发用户数
- 数据库连接数

#### 数据库指标
- 连接数 < 80%
- 慢查询数
- 锁等待时间
- 缓存命中率 > 95%

---

## 故障排查

### 常见问题诊断

#### 1. 应用无法启动

**检查步骤：**
```bash
# 查看应用日志
docker-compose logs app

# 检查配置文件
cat config.yaml

# 检查数据库连接
docker-compose exec app sh
nc -zv postgres 5432

# 检查端口占用
netstat -tulpn | grep 8080
```

#### 2. 数据库连接失败

**检查步骤：**
```bash
# 检查数据库状态
docker-compose exec postgres pg_isready

# 查看数据库日志
docker-compose logs postgres

# 测试连接
docker-compose exec postgres psql -U postgres -d cjadmin -c "SELECT 1;"
```

#### 3. 性能下降

**诊断命令：**
```bash
# 查看系统资源
docker stats

# 查看数据库连接
docker-compose exec postgres psql -U postgres -c "SELECT count(*) FROM pg_stat_activity;"

# 查看慢查询
docker-compose exec postgres psql -U postgres -d cjadmin -c "SELECT query, calls, total_time FROM pg_stat_statements ORDER BY total_time DESC LIMIT 10;"
```

#### 4. API响应慢

**排查步骤：**
1. 检查数据库查询性能
2. 查看API日志中的响应时间
3. 检查网络延迟
4. 分析是否有死锁

```bash
# 查看数据库锁
docker-compose exec postgres psql -U postgres -d cjadmin -c "SELECT * FROM pg_locks WHERE NOT granted;"
```

### 日志分析

#### 应用日志位置
- 容器日志：`docker-compose logs`
- 本地日志：`./logs/app.log`
- Nginx日志：`./logs/nginx/`

#### 日志级别
- DEBUG：详细调试信息
- INFO：一般信息
- WARNING：警告信息
- ERROR：错误信息
- FATAL：致命错误

#### 常用日志命令

```bash
# 实时查看日志
docker-compose logs -f app

# 查看最近100行
docker-compose logs --tail=100 app

# 搜索错误
docker-compose logs app | grep ERROR

# 按时间查询
docker-compose logs --since 2024-01-17T10:00:00 app
```

---

## 安全管理

### 安全检查清单

#### 系统安全
- [ ] 使用强密码策略
- [ ] 启用HTTPS/SSL
- [ ] 配置防火墙规则
- [ ] 定期更新系统补丁
- [ ] 限制SSH访问
- [ ] 配置fail2ban

#### 应用安全
- [ ] JWT密钥定期更换
- [ ] 数据库密码强度检查
- [ ] API限流配置
- [ ] CORS正确配置
- [ ] 输入验证和过滤
- [ ] SQL注入防护

#### 数据安全
- [ ] 敏感数据加密存储
- [ ] 定期数据备份
- [ ] 备份文件加密
- [ ] 访问日志审计
- [ ] 数据传输加密

### 安全配置

#### 1. 启用HTTPS

准备SSL证书：
```bash
mkdir -p nginx/ssl
# 将证书文件放入 nginx/ssl/
```

修改 `nginx/conf.d/default.conf`:
```nginx
server {
    listen 443 ssl http2;
    server_name yourdomain.com;

    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    # 其他配置...
}
```

#### 2. 配置防火墙

```bash
# UFW配置（Ubuntu）
sudo ufw allow 22/tcp      # SSH
sudo ufw allow 80/tcp      # HTTP
sudo ufw allow 443/tcp     # HTTPS
sudo ufw enable

# iptables配置
iptables -A INPUT -p tcp --dport 22 -j ACCEPT
iptables -A INPUT -p tcp --dport 80 -j ACCEPT
iptables -A INPUT -p tcp --dport 443 -j ACCEPT
iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
iptables -P INPUT DROP
```

#### 3. 限制数据库访问

修改 `docker-compose.yml`:
```yaml
services:
  postgres:
    environment:
      POSTGRES_HOST_AUTH_METHOD: md5
    # 不要暴露5432端口到外网
    # ports:
    #   - "5432:5432"
```

#### 4. 配置JWT安全

生成强JWT密钥：
```bash
openssl rand -base64 32
```

修改 `.env`:
```bash
JWT_SECRET=your_generated_secret_key_here
JWT_EXPIRE_HOURS=24
```

### 安全事件响应

#### 1. 发现异常登录

**响应步骤：**
1. 立即禁用相关账户
2. 强制所有用户重新登录
3. 检查审计日志
4. 修改密码和JWT密钥
5. 通知相关用户

```bash
# 禁用用户
curl -X PUT http://localhost:8080/api/users/1/status \
  -H "Authorization: Bearer admin_token" \
  -H "Content-Type: application/json" \
  -d '{"status": "inactive"}'

# 查看用户登录记录
curl -X GET "http://localhost:8080/api/audit-logs?user_id=1&action=login" \
  -H "Authorization: Bearer admin_token"
```

#### 2. 数据泄露

**响应步骤：**
1. 立即隔离受影响系统
2. 评估泄露范围
3. 通知相关方
4. 修复漏洞
5. 加强监控

#### 3. 服务攻击

**响应步骤：**
1. 启用DDoS防护
2. 增加限流规则
3. 分析攻击源
4. 封禁攻击IP
5. 联系服务提供商

---

## 附录

### 常用命令速查

#### Docker命令
```bash
docker-compose up -d           # 启动服务
docker-compose down            # 停止服务
docker-compose restart app     # 重启应用
docker-compose logs -f app     # 查看日志
docker-compose exec app sh     # 进入容器
docker-compose ps              # 查看状态
```

#### 数据库命令
```bash
# 进入数据库
docker-compose exec postgres psql -U postgres cjadmin

# 查看表
\dt

# 查看表结构
\d table_name

# 执行SQL
SELECT * FROM users LIMIT 10;

# 退出
\q
```

#### 系统维护
```bash
# 清理Docker
docker system prune -a

# 清理日志
find logs/ -name "*.log" -mtime +7 -delete

# 备份数据
./backup.sh

# 查看磁盘使用
df -h
du -sh uploads/ logs/
```

### 问题报告模板

```markdown
## 问题描述
[简要描述问题]

## 环境信息
- 系统版本：
- Docker版本：
- 数据库版本：

## 复现步骤
1.
2.
3.

## 期望结果
[描述期望的行为]

## 实际结果
[描述实际发生的情况]

## 日志信息
```
[粘贴相关日志]
```

## 已尝试的解决方法
[描述已经尝试的方法]
```

---

**版本**: v1.0
**最后更新**: 2025-11-17
**适用版本**: 客户管理+收款系统 v1.0+
