# 监控系统配置指南

## 概述

本系统使用Prometheus + Grafana + Alertmanager构建完整的监控和告警体系。

## 监控架构

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│   CJAdmin   │────▶│  Prometheus  │────▶│   Grafana   │
│     App     │     │              │     │             │
└─────────────┘     └──────┬───────┘     └─────────────┘
                           │
┌─────────────┐            │
│  PostgreSQL │◀───────────┤
│  + Exporter │            │
└─────────────┘            │
                           │
┌─────────────┐            │
│    Redis    │◀───────────┤
│  + Exporter │            │
└─────────────┘            │
                           │
┌─────────────┐            │             ┌─────────────┐
│    Node     │◀───────────┼────────────▶│ Alertmanager│
│  Exporter   │            │             │             │
└─────────────┘            │             └─────────────┘
                           │
┌─────────────┐            │
│  cAdvisor   │◀───────────┘
│             │
└─────────────┘
```

## 监控组件

### 1. Prometheus
- **端口**: 9090
- **功能**: 时序数据库，指标收集和存储
- **保留期**: 30天
- **Web界面**: http://localhost:9090

### 2. Grafana
- **端口**: 3000
- **功能**: 数据可视化和仪表板
- **默认账号**: admin / admin123
- **Web界面**: http://localhost:3000

### 3. Alertmanager
- **端口**: 9093
- **功能**: 告警路由和通知
- **Web界面**: http://localhost:9093

### 4. Exporters
- **Node Exporter** (9100): 系统指标（CPU、内存、磁盘等）
- **PostgreSQL Exporter** (9187): 数据库指标
- **Redis Exporter** (9121): Redis指标
- **cAdvisor** (8081): 容器指标

## 快速开始

### 1. 启动监控服务

```bash
cd monitoring

# 启动所有监控服务
docker-compose -f docker-compose.monitoring.yml up -d

# 查看服务状态
docker-compose -f docker-compose.monitoring.yml ps

# 查看日志
docker-compose -f docker-compose.monitoring.yml logs -f
```

### 2. 访问监控界面

#### Prometheus
```bash
# 浏览器访问
http://localhost:9090

# 查看目标状态
http://localhost:9090/targets

# 查看告警规则
http://localhost:9090/alerts
```

#### Grafana
```bash
# 浏览器访问
http://localhost:3000

# 默认账号密码
用户名: admin
密码: admin123
```

首次登录后，系统会自动加载预配置的仪表板：
- **CJAdmin 应用监控**: 应用性能、请求、错误等

### 3. 配置告警通知

编辑 `monitoring/alertmanager.yml`:

```yaml
receivers:
  - name: 'critical-alerts'
    # 配置邮件通知
    email_configs:
      - to: 'your-email@example.com'
        from: 'alertmanager@example.com'
        smarthost: 'smtp.example.com:587'
        auth_username: 'alertmanager@example.com'
        auth_password: 'your_password'

    # 配置Webhook通知
    webhook_configs:
      - url: 'http://your-webhook-url'

    # 配置钉钉通知（需要第三方工具）
    # webhook_configs:
    #   - url: 'http://dingtalk-webhook-proxy:8060/dingtalk/webhook'
```

重启Alertmanager:
```bash
docker-compose -f docker-compose.monitoring.yml restart alertmanager
```

## 监控指标说明

### 应用指标

#### HTTP请求指标
- `http_requests_total`: 总请求数
  - 标签: method, path, status
  - 类型: Counter

- `http_request_duration_seconds`: 请求耗时
  - 标签: method, path, status
  - 类型: Histogram

#### 业务指标
- `cjadmin_customers_total`: 客户总数
- `cjadmin_orders_total`: 订单总数
- `cjadmin_payments_total`: 支付总额
- `cjadmin_active_tenants`: 活跃租户数

#### 系统指标
- `process_cpu_seconds_total`: 进程CPU时间
- `process_resident_memory_bytes`: 进程内存使用
- `go_goroutines`: Goroutine数量
- `go_memstats_*`: Go内存统计

### 数据库指标

#### PostgreSQL
- `pg_stat_database_numbackends`: 当前连接数
- `pg_stat_database_xact_commit`: 事务提交数
- `pg_stat_database_xact_rollback`: 事务回滚数
- `pg_stat_database_deadlocks`: 死锁数
- `pg_stat_database_tup_*`: 元组操作统计
- `pg_database_size_bytes`: 数据库大小

#### Redis
- `redis_connected_clients`: 连接数
- `redis_memory_used_bytes`: 内存使用
- `redis_memory_max_bytes`: 最大内存
- `redis_keyspace_hits_total`: 键命中数
- `redis_keyspace_misses_total`: 键未命中数
- `redis_evicted_keys_total`: 驱逐键数

### 系统指标

#### Node Exporter
- `node_cpu_seconds_total`: CPU使用时间
- `node_memory_MemTotal_bytes`: 总内存
- `node_memory_MemAvailable_bytes`: 可用内存
- `node_filesystem_size_bytes`: 文件系统大小
- `node_filesystem_avail_bytes`: 可用空间
- `node_network_receive_bytes_total`: 网络接收字节
- `node_network_transmit_bytes_total`: 网络发送字节

## 告警规则

### 应用告警

| 告警名称 | 触发条件 | 级别 | 持续时间 |
|---------|---------|------|---------|
| AppServiceDown | 服务停机 | critical | 1分钟 |
| HighErrorRate | 5xx错误率 > 5% | warning | 5分钟 |
| SlowResponseTime | P95响应时间 > 1秒 | warning | 5分钟 |
| HighCPUUsage | CPU使用率 > 80% | warning | 10分钟 |
| HighMemoryUsage | 内存使用 > 2GB | warning | 10分钟 |
| HighGoroutineCount | Goroutine数 > 1000 | warning | 10分钟 |

### 数据库告警

| 告警名称 | 触发条件 | 级别 | 持续时间 |
|---------|---------|------|---------|
| HighDatabaseConnections | 连接数 > 80 | warning | 5分钟 |
| DatabaseDeadlocks | 检测到死锁 | critical | 1分钟 |
| DatabaseReplicationLag | 复制延迟 > 10秒 | warning | 5分钟 |
| RedisHighMemoryUsage | 内存使用率 > 90% | warning | 5分钟 |
| RedisHighConnections | 连接数 > 100 | warning | 5分钟 |
| RedisEvictedKeys | 发生键驱逐 | warning | 5分钟 |

### 系统告警

| 告警名称 | 触发条件 | 级别 | 持续时间 |
|---------|---------|------|---------|
| DiskSpaceLow | 磁盘剩余 < 20% | warning | 10分钟 |
| SystemHighCPU | CPU使用率 > 85% | warning | 10分钟 |
| SystemHighMemory | 内存使用率 > 90% | critical | 10分钟 |

## 自定义指标

### 1. 在应用中暴露指标

安装Prometheus Go客户端:
```bash
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promhttp
```

在代码中定义指标:
```go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // 请求计数器
    RequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )

    // 请求耗时直方图
    RequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path", "status"},
    )

    // 业务指标
    CustomersTotal = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "cjadmin_customers_total",
        Help: "Total number of customers",
    })
)
```

暴露指标端点:
```go
import (
    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func SetupMetricsRoute(router *gin.Engine) {
    router.GET("/api/metrics", gin.WrapH(promhttp.Handler()))
}
```

使用指标:
```go
// 记录请求
RequestsTotal.WithLabelValues(method, path, status).Inc()

// 记录耗时
timer := prometheus.NewTimer(RequestDuration.WithLabelValues(method, path, status))
defer timer.ObserveDuration()

// 更新业务指标
CustomersTotal.Set(float64(count))
```

### 2. 添加自定义告警规则

创建新的告警规则文件 `monitoring/alerts/custom_alerts.yml`:

```yaml
groups:
  - name: custom_alerts
    interval: 30s
    rules:
      - alert: HighOrderVolume
        expr: rate(cjadmin_orders_total[5m]) > 100
        for: 5m
        labels:
          severity: info
        annotations:
          summary: "订单量激增"
          description: "过去5分钟订单速率超过100/分钟"
```

重启Prometheus:
```bash
docker-compose -f docker-compose.monitoring.yml restart prometheus
```

### 3. 创建自定义Grafana仪表板

1. 登录Grafana: http://localhost:3000
2. 点击 "+" → "Create Dashboard"
3. 添加面板并配置查询
4. 保存仪表板

常用查询示例:
```promql
# HTTP请求速率
rate(http_requests_total[5m])

# P95响应时间
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# CPU使用率
rate(process_cpu_seconds_total[5m]) * 100

# 内存使用
process_resident_memory_bytes / 1024 / 1024

# 活跃用户数
count(user_last_login_timestamp > (time() - 300))
```

## 最佳实践

### 1. 指标命名规范

- 使用小写字母和下划线
- 前缀表示应用名: `cjadmin_`
- 后缀表示单位: `_bytes`, `_seconds`, `_total`
- Counter类型使用 `_total` 后缀
- 示例: `cjadmin_http_requests_total`

### 2. 标签使用

- 避免使用高基数标签（如用户ID）
- 使用有意义的标签名
- 保持标签一致性
- 限制标签值的数量

### 3. 告警配置

- 设置合理的阈值
- 避免告警风暴
- 使用告警分组
- 配置告警抑制规则
- 设置合理的重复间隔

### 4. 性能优化

- 合理设置抓取间隔（15s-60s）
- 限制指标保留期（30天）
- 使用recording rules预计算
- 定期清理无用指标

## 故障排查

### 常见问题

#### 1. Prometheus无法抓取目标

```bash
# 检查目标状态
curl http://localhost:9090/api/v1/targets

# 检查网络连通性
docker-compose -f docker-compose.monitoring.yml exec prometheus wget -O- http://app:8080/api/metrics

# 查看Prometheus日志
docker-compose -f docker-compose.monitoring.yml logs prometheus
```

#### 2. Grafana无法连接Prometheus

```bash
# 检查数据源配置
http://localhost:3000/datasources

# 测试连接
curl http://prometheus:9090/api/v1/query?query=up

# 查看Grafana日志
docker-compose -f docker-compose.monitoring.yml logs grafana
```

#### 3. 告警未触发

```bash
# 检查告警规则
http://localhost:9090/alerts

# 检查Alertmanager状态
http://localhost:9093/#/alerts

# 查看Alertmanager日志
docker-compose -f docker-compose.monitoring.yml logs alertmanager
```

### 调试技巧

```bash
# 查询指标是否存在
curl 'http://localhost:9090/api/v1/query?query=http_requests_total'

# 验证PromQL查询
http://localhost:9090/graph

# 检查告警规则语法
promtool check rules monitoring/alerts/*.yml

# 验证Prometheus配置
promtool check config monitoring/prometheus.yml
```

## 数据备份

### Prometheus数据备份

```bash
# 创建快照
curl -XPOST http://localhost:9090/api/v1/admin/tsdb/snapshot

# 备份数据目录
docker-compose -f docker-compose.monitoring.yml exec prometheus \
  tar czf /tmp/prometheus-backup.tar.gz /prometheus

# 复制到主机
docker cp cjadmin-prometheus:/tmp/prometheus-backup.tar.gz ./
```

### Grafana配置备份

```bash
# 备份Grafana数据
docker-compose -f docker-compose.monitoring.yml exec grafana \
  tar czf /tmp/grafana-backup.tar.gz /var/lib/grafana

# 复制到主机
docker cp cjadmin-grafana:/tmp/grafana-backup.tar.gz ./
```

## 扩展集成

### 集成钉钉通知

使用 [prometheus-webhook-dingtalk](https://github.com/timonwong/prometheus-webhook-dingtalk):

```yaml
dingtalk:
  image: timonwong/prometheus-webhook-dingtalk:latest
  environment:
    - WEBHOOK_URL=https://oapi.dingtalk.com/robot/send?access_token=YOUR_TOKEN
  ports:
    - "8060:8060"
```

### 集成企业微信通知

修改 `alertmanager.yml`:
```yaml
receivers:
  - name: 'wechat'
    wechat_configs:
      - corp_id: 'YOUR_CORP_ID'
        api_secret: 'YOUR_SECRET'
        to_user: '@all'
        agent_id: 'YOUR_AGENT_ID'
```

## 监控清单

- [ ] Prometheus正常运行并收集指标
- [ ] Grafana仪表板正常显示
- [ ] 所有Exporter正常工作
- [ ] 告警规则已配置
- [ ] Alertmanager正确路由告警
- [ ] 告警通知渠道已测试
- [ ] 数据备份策略已建立
- [ ] 监控数据保留期已设置
- [ ] 团队成员已培训

## 参考资源

- [Prometheus官方文档](https://prometheus.io/docs/)
- [Grafana官方文档](https://grafana.com/docs/)
- [PromQL查询语言](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [Alertmanager配置](https://prometheus.io/docs/alerting/latest/configuration/)
- [Node Exporter指标](https://github.com/prometheus/node_exporter)
