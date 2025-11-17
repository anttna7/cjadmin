package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/anttna7/cjadmin/internal/config"
	"github.com/anttna7/cjadmin/internal/database"
	"github.com/anttna7/cjadmin/internal/handler"
	"github.com/anttna7/cjadmin/internal/middleware"
	"github.com/anttna7/cjadmin/internal/service"
	"github.com/anttna7/cjadmin/internal/utils"
	"github.com/gin-gonic/gin"
)

var (
	configPath = flag.String("config", "config.yaml", "配置文件路径")
	initAdmin  = flag.Bool("init-admin", false, "初始化超级管理员")
	adminUser  = flag.String("admin-user", "admin", "管理员用户名")
	adminPass  = flag.String("admin-pass", "admin123", "管理员密码")
)

func main() {
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 初始化JWT
	utils.InitJWT(cfg.JWT.Secret)

	// 连接数据库
	if err := database.InitDB(&cfg.Database); err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer database.CloseDB()

	// 如果需要初始化管理员
	if *initAdmin {
		authService := service.NewAuthService(cfg)
		if err := authService.CreateFirstAdmin(*adminUser, *adminPass); err != nil {
			log.Printf("创建管理员失败: %v", err)
		} else {
			log.Printf("管理员创建成功: %s", *adminUser)
		}
		return
	}

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 创建路由
	router := setupRouter(cfg)

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("服务器启动在 %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}

func setupRouter(cfg *config.Config) *gin.Engine {
	router := gin.Default()

	// 全局中间件
	router.Use(middleware.CORS())

	// 静态文件
	router.Static("/static", "./web/static")
	router.LoadHTMLGlob("web/templates/*")

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		utils.Success(c, gin.H{"status": "ok"})
	})

	// API路由组
	api := router.Group("/api")
	{
		// 认证相关（不需要登录）
		authHandler := handler.NewAuthHandler(cfg)
		api.POST("/login", authHandler.Login)
		api.POST("/register", authHandler.Register) // 注册路由，实际应该需要管理员权限

		// 支付回调（不需要登录，但应该有签名验证）
		orderHandler := handler.NewOrderHandler()
		api.POST("/payment/callback", orderHandler.PaymentCallback)

		// 需要认证的路由
		authorized := api.Group("")
		authorized.Use(middleware.AuthMiddleware())
		{
			// 当前用户信息
			authorized.GET("/user/current", authHandler.GetCurrentUser)

			// 客户管理
			customerHandler := handler.NewCustomerHandler()
			customers := authorized.Group("/customers")
			customers.Use(middleware.TenantMiddleware())
			{
				customers.POST("", middleware.RequirePermission("customer.create"), customerHandler.Create)
				customers.GET("", middleware.RequirePermission("customer.read"), customerHandler.List)
				customers.GET("/:id", middleware.RequirePermission("customer.read"), customerHandler.Get)
				customers.PUT("/:id", middleware.RequirePermission("customer.update"), customerHandler.Update)
				customers.DELETE("/:id", middleware.RequirePermission("customer.delete"), customerHandler.Delete)
				customers.GET("/:id/accounts", middleware.RequirePermission("customer.read"), customerHandler.GetAccounts)
			}

			// 订单管理
			orders := authorized.Group("/orders")
			orders.Use(middleware.TenantMiddleware())
			{
				orders.POST("/recharge", middleware.RequirePermission("order.create"), orderHandler.CreateRechargeOrder)
				orders.POST("/transfer", middleware.RequirePermission("order.create"), orderHandler.CreateTransferOrder)
				orders.POST("/:id/verify", middleware.RequirePermission("finance.verify"), orderHandler.VerifyRechargeOrder)
				orders.GET("", middleware.RequirePermission("order.read"), orderHandler.ListOrders)
				orders.GET("/:id", middleware.RequirePermission("order.read"), orderHandler.GetOrder)
			}
		}
	}

	// 前端页面路由
	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{
			"title": "客户管理+收款系统",
		})
	})

	router.GET("/login", func(c *gin.Context) {
		c.HTML(200, "login.html", gin.H{})
	})

	router.GET("/dashboard", func(c *gin.Context) {
		c.HTML(200, "dashboard.html", gin.H{})
	})

	return router
}
