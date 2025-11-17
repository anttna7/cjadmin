# Docker部署指南

## 概述

本文档提供了使用Docker和Docker Compose部署客户管理+收款系统的完整指南。

## 系统要求

### 最低配置
- CPU: 2核
- 内存: 4GB
- 磁盘: 20GB
- 操作系统: Linux/macOS/Windows (支持Docker)

### 推荐配置
- CPU: 4核
- 内存: 8GB
- 磁盘: 50GB SSD
- 操作系统: Linux (Ubuntu 20.04+ / CentOS 8+)

### 软件依赖
- Docker 20.10+
- Docker Compose 2.0+

## 快速开始

### 1. 克隆代码

```bash
git clone https://github.com/anttna7/cjadmin.git
cd cjadmin
```

### 2. 配置环境变量

```bash
cp .env.example .env
```

编辑 `.env` 文件，修改以下配置：

```bash
# 数据库配置
DB_NAME=cjadmin
DB_USER=postgres
DB_PASSWORD=your_secure_password_here  # 修改为安全密码
DB_PORT=5432

# JWT密钥
JWT_SECRET=your_jwt_secret_key_here    # 修改为随机密钥

# 其他配置...
```

### 3. 启动服务

```bash
# 构建并启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f app
```

### 4. 初始化数据库

```bash
# 进入应用容器
docker-compose exec app sh

# 运行数据库迁移（如果需要）
# ./main migrate
```

### 5. 访问应用

- 应用地址: http://localhost:8080
- API文档: http://localhost:8080/api/health
- Nginx代理: http://localhost:80 (如果启用)

### 6. 创建超级管理员

```bash
# 使用API创建第一个管理员
curl -X POST http://localhost:8080/api/auth/create-admin \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123456"
  }'
```

## 服务说明

### 服务架构

```
┌─────────────┐
│   Nginx     │ :80, :443
└──────┬──────┘
       │
┌──────▼──────┐
│     App     │ :8080
└──────┬──────┘
       │
┌──────▼──────┐
│  PostgreSQL │ :5432
└─────────────┘
       │
┌──────▼──────┐
│    Redis    │ :6379 (可选)
└─────────────┘
```

### 服务列表

#### 1. PostgreSQL (postgres)
- **镜像**: postgres:18-alpine
- **端口**: 5432
- **功能**: 主数据库
- **数据卷**: postgres_data

#### 2. Redis (redis)
- **镜像**: redis:7-alpine
- **端口**: 6379
- **功能**: 缓存和会话存储
- **数据卷**: redis_data

#### 3. 应用服务 (app)
- **镜像**: 自构建 (基于golang:1.25-alpine)
- **端口**: 8080
- **功能**: 主应用服务
- **数据卷**: uploads/, logs/

#### 4. Nginx (nginx) - 可选
- **镜像**: nginx:alpine
- **端口**: 80, 443
- **功能**: 反向代理、负载均衡
- **配置**: nginx/conf.d/

## Docker Compose命令

### 基本命令

```bash
# 启动所有服务
docker-compose up -d

# 停止所有服务
docker-compose down

# 重启服务
docker-compose restart

# 查看服务状态
docker-compose ps

# 查看服务日志
docker-compose logs -f [service_name]

# 进入容器
docker-compose exec [service_name] sh
```

### 服务管理

```bash
# 只启动特定服务
docker-compose up -d app postgres

# 停止特定服务
docker-compose stop app

# 重启特定服务
docker-compose restart app

# 查看特定服务日志
docker-compose logs -f app
```

### 数据管理

```bash
# 备份数据库
docker-compose exec postgres pg_dump -U postgres cjadmin > backup.sql

# 恢复数据库
cat backup.sql | docker-compose exec -T postgres psql -U postgres cjadmin

# 清理数据卷（谨慎操作！）
docker-compose down -v
```

## 生产环境部署

### 1. 安全配置

#### 修改默认密码
```bash
# .env文件
DB_PASSWORD=your_very_secure_password_123!@#
JWT_SECRET=$(openssl rand -base64 32)
```

#### 启用HTTPS
```bash
# 准备SSL证书
mkdir -p nginx/ssl
cp your-cert.pem nginx/ssl/
cp your-key.pem nginx/ssl/
```

修改 `nginx/conf.d/default.conf`:
```nginx
server {
    listen 443 ssl http2;
    server_name yourdomain.com;

    ssl_certificate /etc/nginx/ssl/your-cert.pem;
    ssl_certificate_key /etc/nginx/ssl/your-key.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    # ... 其他配置
}
```

### 2. 性能优化

#### 调整资源限制

修改 `docker-compose.yml`:
```yaml
services:
  app:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '1'
          memory: 1G
```

#### PostgreSQL优化

```yaml
services:
  postgres:
    environment:
      # 增加共享内存
      POSTGRES_INITDB_ARGS: "-E UTF8 --locale=C"
    command:
      - postgres
      - -c
      - shared_buffers=256MB
      - -c
      - max_connections=200
```

### 3. 监控和日志

#### 日志轮转

创建 `docker-compose.override.yml`:
```yaml
version: '3.8'
services:
  app:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

#### 健康检查

所有服务都配置了健康检查：
- PostgreSQL: 每10秒检查一次
- Redis: 每10秒检查一次
- App: 每30秒检查一次

查看健康状态：
```bash
docker-compose ps
```

### 4. 备份策略

#### 自动备份脚本

创建 `backup.sh`:
```bash
#!/bin/bash
BACKUP_DIR="/backups"
DATE=$(date +%Y%m%d_%H%M%S)

# 备份数据库
docker-compose exec -T postgres pg_dump -U postgres cjadmin | gzip > "$BACKUP_DIR/db_$DATE.sql.gz"

# 备份上传文件
tar -czf "$BACKUP_DIR/uploads_$DATE.tar.gz" uploads/

# 保留最近7天的备份
find "$BACKUP_DIR" -name "*.gz" -mtime +7 -delete

echo "Backup completed: $DATE"
```

设置定时任务：
```bash
# 每天凌晨2点执行备份
0 2 * * * /path/to/backup.sh >> /var/log/backup.log 2>&1
```

## 故障排查

### 常见问题

#### 1. 服务无法启动

```bash
# 查看详细日志
docker-compose logs [service_name]

# 检查端口占用
netstat -tulpn | grep LISTEN

# 检查Docker资源
docker system df
docker system prune
```

#### 2. 数据库连接失败

```bash
# 检查数据库是否就绪
docker-compose exec postgres pg_isready -U postgres

# 查看数据库日志
docker-compose logs postgres

# 重启数据库
docker-compose restart postgres
```

#### 3. 应用无法访问

```bash
# 检查应用健康状态
curl http://localhost:8080/api/health

# 进入容器排查
docker-compose exec app sh
ps aux
netstat -tulpn
```

#### 4. 文件上传失败

```bash
# 检查uploads目录权限
ls -la uploads/

# 调整权限
chmod -R 755 uploads/
chown -R 1000:1000 uploads/
```

### 性能问题排查

```bash
# 查看资源使用情况
docker stats

# 查看容器进程
docker-compose top

# 分析慢查询（PostgreSQL）
docker-compose exec postgres psql -U postgres -d cjadmin -c "SELECT * FROM pg_stat_statements ORDER BY total_time DESC LIMIT 10;"
```

## 更新和维护

### 更新应用

```bash
# 1. 备份数据
./backup.sh

# 2. 拉取最新代码
git pull origin main

# 3. 重新构建镜像
docker-compose build app

# 4. 重启服务
docker-compose up -d app

# 5. 验证更新
docker-compose logs -f app
```

### 数据库迁移

```bash
# 进入应用容器
docker-compose exec app sh

# 运行迁移脚本
# ./migrate up
```

### 清理和维护

```bash
# 清理未使用的镜像
docker image prune -a

# 清理未使用的容器
docker container prune

# 清理未使用的网络
docker network prune

# 清理未使用的数据卷（谨慎！）
docker volume prune
```

## 扩展部署

### 多实例部署

修改 `docker-compose.yml`:
```yaml
services:
  app:
    deploy:
      replicas: 3
```

### 负载均衡

使用Nginx upstream:
```nginx
upstream app_backend {
    least_conn;
    server app1:8080;
    server app2:8080;
    server app3:8080;
}
```

### 数据库主从复制

参考PostgreSQL官方文档配置主从复制。

## 安全建议

1. **定期更新镜像**
   ```bash
   docker-compose pull
   docker-compose up -d
   ```

2. **使用secrets管理敏感信息**
   - 使用Docker secrets或环境变量加密工具

3. **限制容器权限**
   - 使用非root用户运行容器
   - 启用只读文件系统

4. **网络隔离**
   - 使用内部网络隔离服务
   - 只暴露必要的端口

5. **日志审计**
   - 启用审计日志
   - 定期检查异常访问

## 监控集成

### Prometheus + Grafana

可以集成Prometheus和Grafana进行监控：
```yaml
services:
  prometheus:
    image: prom/prometheus
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    ports:
      - "9090:9090"

  grafana:
    image: grafana/grafana
    ports:
      - "3000:3000"
```

## 参考资料

- [Docker官方文档](https://docs.docker.com/)
- [Docker Compose文档](https://docs.docker.com/compose/)
- [PostgreSQL Docker](https://hub.docker.com/_/postgres)
- [Redis Docker](https://hub.docker.com/_/redis)
- [Nginx Docker](https://hub.docker.com/_/nginx)

## 技术支持

如遇问题，请访问：
- GitHub Issues: https://github.com/anttna7/cjadmin/issues
- 项目文档: ./docs/
