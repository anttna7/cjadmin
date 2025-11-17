# 部署文档

本文档介绍如何在生产环境中部署客户管理+收款系统。

## 部署方式

### 1. 直接部署（推荐用于开发/测试）

#### 环境准备

1. 安装Go 1.25+
   ```bash
   wget https://golang.org/dl/go1.25.linux-amd64.tar.gz
   sudo tar -C /usr/local -xzf go1.25.linux-amd64.tar.gz
   export PATH=$PATH:/usr/local/go/bin
   ```

2. 安装PostgreSQL 18+
   ```bash
   sudo apt-get update
   sudo apt-get install postgresql-18 postgresql-contrib
   ```

#### 部署步骤

1. 克隆代码
   ```bash
   git clone https://github.com/anttna7/cjadmin.git
   cd cjadmin
   ```

2. 配置数据库
   ```bash
   sudo -u postgres createdb cjadmin
   sudo -u postgres psql -d cjadmin -f migrations/001_init_schema.sql
   ```

3. 修改配置
   ```bash
   cp .env.example .env
   vim config.yaml  # 修改数据库连接等配置
   ```

4. 编译程序
   ```bash
   make build
   ```

5. 初始化管理员
   ```bash
   ./bin/cjadmin -init-admin -admin-user=admin -admin-pass=YourSecurePassword
   ```

6. 启动服务
   ```bash
   ./bin/cjadmin
   ```

### 2. Docker部署（推荐用于生产）

#### 创建docker-compose.yml

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:18-alpine
    container_name: cjadmin-postgres
    environment:
      POSTGRES_DB: cjadmin
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    ports:
      - "5432:5432"
    restart: unless-stopped

  app:
    build: .
    container_name: cjadmin-app
    depends_on:
      - postgres
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: postgres
      DB_PASSWORD: ${DB_PASSWORD}
      DB_NAME: cjadmin
    ports:
      - "8080:8080"
    volumes:
      - ./uploads:/app/uploads
      - ./logs:/app/logs
    restart: unless-stopped

  nginx:
    image: nginx:alpine
    container_name: cjadmin-nginx
    depends_on:
      - app
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./nginx/ssl:/etc/nginx/ssl
    restart: unless-stopped

volumes:
  postgres_data:
```

#### Dockerfile

```dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app

# 安装依赖
RUN apk add --no-cache git

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/api/main.go

# 运行阶段
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

# 从构建阶段复制文件
COPY --from=builder /app/main .
COPY --from=builder /app/config.yaml .
COPY --from=builder /app/web ./web

# 创建必要的目录
RUN mkdir -p uploads logs

EXPOSE 8080

CMD ["./main"]
```

#### 部署步骤

1. 创建配置文件
   ```bash
   cp .env.example .env
   vim .env  # 修改密码等敏感信息
   ```

2. 启动服务
   ```bash
   docker-compose up -d
   ```

3. 初始化管理员
   ```bash
   docker exec -it cjadmin-app ./main -init-admin -admin-user=admin -admin-pass=YourSecurePassword
   ```

4. 查看日志
   ```bash
   docker-compose logs -f
   ```

### 3. Nginx反向代理配置

创建 `nginx/nginx.conf`:

```nginx
events {
    worker_connections 1024;
}

http {
    upstream app {
        server app:8080;
    }

    # HTTP服务器
    server {
        listen 80;
        server_name your-domain.com;

        # 重定向到HTTPS
        return 301 https://$server_name$request_uri;
    }

    # HTTPS服务器
    server {
        listen 443 ssl http2;
        server_name your-domain.com;

        ssl_certificate /etc/nginx/ssl/cert.pem;
        ssl_certificate_key /etc/nginx/ssl/key.pem;

        # SSL配置
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers HIGH:!aNULL:!MD5;
        ssl_prefer_server_ciphers on;

        # 日志
        access_log /var/log/nginx/access.log;
        error_log /var/log/nginx/error.log;

        # 代理设置
        location / {
            proxy_pass http://app;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # 静态文件
        location /static/ {
            proxy_pass http://app;
            expires 30d;
            add_header Cache-Control "public, immutable";
        }
    }
}
```

### 4. 使用Systemd管理服务

创建 `/etc/systemd/system/cjadmin.service`:

```ini
[Unit]
Description=Customer Management System
After=network.target postgresql.service

[Service]
Type=simple
User=cjadmin
WorkingDirectory=/opt/cjadmin
ExecStart=/opt/cjadmin/bin/cjadmin
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

启用服务：
```bash
sudo systemctl daemon-reload
sudo systemctl enable cjadmin
sudo systemctl start cjadmin
sudo systemctl status cjadmin
```

## 性能优化

### 1. 数据库优化

#### PostgreSQL配置优化

编辑 `/etc/postgresql/18/main/postgresql.conf`:

```conf
# 内存配置
shared_buffers = 256MB
effective_cache_size = 1GB
work_mem = 16MB
maintenance_work_mem = 128MB

# 连接配置
max_connections = 200

# 日志配置
log_min_duration_statement = 1000

# 查询优化
random_page_cost = 1.1
```

#### 创建索引

```sql
-- 为常用查询字段创建索引
CREATE INDEX CONCURRENTLY idx_customers_name_search ON customers USING gin(to_tsvector('simple', customer_name));
CREATE INDEX CONCURRENTLY idx_orders_created_at_desc ON orders(created_at DESC);
CREATE INDEX CONCURRENTLY idx_transactions_created_at_desc ON account_transactions(created_at DESC);
```

### 2. 应用优化

#### 连接池配置

在 `config.yaml` 中调整：

```yaml
database:
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 3600
```

#### 启用缓存

可以考虑使用Redis缓存热点数据：

```go
// 示例：缓存用户权限
func (s *AuthService) GetUserPermissions(userID int64) ([]string, error) {
    // 先从Redis获取
    cacheKey := fmt.Sprintf("user:permissions:%d", userID)
    cached, err := redis.Get(cacheKey)
    if err == nil {
        return cached, nil
    }

    // 从数据库查询
    permissions := s.queryFromDB(userID)

    // 写入缓存
    redis.Set(cacheKey, permissions, 1*time.Hour)

    return permissions, nil
}
```

## 监控和日志

### 1. 应用监控

#### 使用Prometheus监控

```go
// 添加Prometheus监控
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// 注册指标
var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
)

// 在main函数中
prometheus.MustRegister(httpRequestsTotal)
router.GET("/metrics", gin.WrapH(promhttp.Handler()))
```

### 2. 日志管理

#### 使用ELK Stack

1. 配置Filebeat收集日志
2. 发送到Elasticsearch
3. 使用Kibana可视化

#### 日志格式

```go
// 结构化日志
log.WithFields(log.Fields{
    "user_id": userID,
    "action": "create_order",
    "order_id": orderID,
}).Info("Order created successfully")
```

## 备份策略

### 1. 数据库备份

#### 每日全量备份

```bash
#!/bin/bash
# backup.sh
BACKUP_DIR="/backup/postgres"
DATE=$(date +%Y%m%d_%H%M%S)
pg_dump -U postgres cjadmin > $BACKUP_DIR/cjadmin_$DATE.sql
find $BACKUP_DIR -name "*.sql" -mtime +7 -delete
```

#### 配置定时任务

```bash
# 每天凌晨2点执行备份
0 2 * * * /opt/cjadmin/scripts/backup.sh
```

### 2. 文件备份

```bash
# 备份上传文件
rsync -avz /opt/cjadmin/uploads/ /backup/uploads/
```

## 安全加固

### 1. 防火墙配置

```bash
# 只开放必要端口
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

### 2. SSL证书

使用Let's Encrypt免费证书：

```bash
sudo apt-get install certbot
sudo certbot certonly --standalone -d your-domain.com
```

### 3. 环境变量管理

不要在代码中硬编码敏感信息，使用环境变量：

```bash
export JWT_SECRET=$(openssl rand -base64 32)
export DB_PASSWORD=$(openssl rand -base64 16)
```

### 4. 定期更新

```bash
# 更新系统
sudo apt-get update && sudo apt-get upgrade

# 更新Go依赖
go get -u all
go mod tidy
```

## 故障排查

### 常见问题

1. **无法连接数据库**
   - 检查PostgreSQL服务状态
   - 验证数据库连接配置
   - 检查防火墙规则

2. **服务启动失败**
   - 查看日志：`journalctl -u cjadmin -n 50`
   - 检查端口占用：`lsof -i :8080`

3. **性能问题**
   - 检查数据库慢查询
   - 查看系统资源使用
   - 分析应用日志

### 日志查看

```bash
# 查看应用日志
tail -f logs/app.log

# 查看Nginx日志
tail -f /var/log/nginx/access.log
tail -f /var/log/nginx/error.log

# 查看PostgreSQL日志
tail -f /var/log/postgresql/postgresql-18-main.log
```

## 扩展部署

### 水平扩展

1. 使用负载均衡器（如HAProxy、Nginx）
2. 部署多个应用实例
3. 使用共享存储（NFS、S3）存储上传文件
4. 使用Redis集群共享会话

### 高可用部署

1. PostgreSQL主从复制
2. 应用多实例部署
3. Nginx负载均衡
4. 健康检查和自动故障转移

## 联系支持

如有部署问题，请：
1. 查看项目文档
2. 提交Issue：https://github.com/anttna7/cjadmin/issues
3. 发送邮件（如有）
