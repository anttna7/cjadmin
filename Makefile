.PHONY: help build run init-db init-admin clean test test-coverage test-race test-bench

help: ## 显示帮助信息
	@echo "可用命令："
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## 编译项目
	@echo "编译项目..."
	@go build -o bin/cjadmin cmd/api/main.go
	@echo "编译完成: bin/cjadmin"

run: ## 运行项目
	@echo "启动服务..."
	@go run cmd/api/main.go

init-db: ## 初始化数据库
	@echo "初始化数据库..."
	@psql -U postgres -d cjadmin -f migrations/001_init_schema.sql
	@echo "数据库初始化完成"

init-admin: ## 创建超级管理员
	@echo "创建超级管理员..."
	@go run cmd/api/main.go -init-admin -admin-user=admin -admin-pass=admin123
	@echo "管理员创建完成"

clean: ## 清理编译文件
	@echo "清理编译文件..."
	@rm -rf bin/
	@echo "清理完成"

test: ## 运行测试
	@echo "运行测试..."
	@go test -v ./...

test-coverage: ## 运行测试并生成覆盖率报告
	@echo "运行测试并生成覆盖率报告..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

test-race: ## 运行竞态检测
	@echo "运行竞态检测..."
	@go test -race ./...

test-bench: ## 运行性能基准测试
	@echo "运行性能基准测试..."
	@go test -bench=. -benchmem ./...

deps: ## 下载依赖
	@echo "下载依赖..."
	@go mod download
	@echo "依赖下载完成"

fmt: ## 格式化代码
	@echo "格式化代码..."
	@go fmt ./...
	@echo "格式化完成"

dev: ## 开发模式运行（热重载需要安装air）
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "请先安装 air: go install github.com/cosmtrek/air@latest"; \
	fi

docker-build: ## 构建Docker镜像
	@echo "构建Docker镜像..."
	@docker build -t cjadmin:latest .
	@echo "镜像构建完成"

docker-run: ## 运行Docker容器
	@echo "运行Docker容器..."
	@docker run -d -p 8080:8080 --name cjadmin cjadmin:latest
	@echo "容器启动完成"

docker-stop: ## 停止Docker容器
	@echo "停止Docker容器..."
	@docker stop cjadmin
	@docker rm cjadmin
	@echo "容器已停止"
