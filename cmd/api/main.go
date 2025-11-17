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

			// 租户管理（仅平台管理员）
			tenantHandler := handler.NewTenantHandler()
			tenants := authorized.Group("/tenants")
			tenants.Use(middleware.RequirePlatformAdmin())
			{
				tenants.POST("", tenantHandler.Create)
				tenants.GET("", tenantHandler.List)
				tenants.GET("/:id", tenantHandler.Get)
				tenants.PUT("/:id", tenantHandler.Update)
				tenants.DELETE("/:id", tenantHandler.Delete)
				tenants.PUT("/:id/status", tenantHandler.UpdateStatus)
				tenants.GET("/:id/statistics", tenantHandler.GetStatistics)
			}

			// 部门管理
			deptHandler := handler.NewDepartmentHandler()
			departments := authorized.Group("/departments")
			departments.Use(middleware.TenantMiddleware())
			{
				departments.POST("", middleware.RequirePermission("user.create"), deptHandler.Create)
				departments.GET("/tree", middleware.RequirePermission("user.read"), deptHandler.GetTree)
				departments.GET("", middleware.RequirePermission("user.read"), deptHandler.GetList)
				departments.PUT("/:id", middleware.RequirePermission("user.update"), deptHandler.Update)
				departments.DELETE("/:id", middleware.RequirePermission("user.delete"), deptHandler.Delete)
				departments.PUT("/:id/permissions", middleware.RequirePermission("user.update"), deptHandler.SetPermissions)
				departments.GET("/:id/permissions", middleware.RequirePermission("user.read"), deptHandler.GetPermissions)
				departments.GET("/:id/permissions/inherited", middleware.RequirePermission("user.read"), deptHandler.GetInheritedPermissions)
			}

			// 角色权限管理
			roleHandler := handler.NewRoleHandler()
			roles := authorized.Group("/roles")
			{
				// 平台角色（仅平台管理员）
				roles.POST("/platform", middleware.RequirePlatformAdmin(), roleHandler.CreatePlatformRole)

				// 租户角色
				roles.POST("", middleware.TenantMiddleware(), middleware.RequirePermission("role.create"), roleHandler.CreateTenantRole)
				roles.GET("", middleware.RequirePermission("role.read"), roleHandler.GetList)
				roles.GET("/:id", middleware.RequirePermission("role.read"), roleHandler.Get)
				roles.PUT("/:id", middleware.RequirePermission("role.update"), roleHandler.Update)
				roles.DELETE("/:id", middleware.RequirePermission("role.delete"), roleHandler.Delete)

				// 角色权限管理
				roles.PUT("/:id/permissions", middleware.RequirePermission("role.update"), roleHandler.SetPermissions)
				roles.GET("/:id/permissions", middleware.RequirePermission("role.read"), roleHandler.GetPermissions)
			}

			// 权限列表（所有登录用户可查看）
			permissions := authorized.Group("/permissions")
			{
				permissions.GET("", roleHandler.GetAllPermissions)
				permissions.GET("/by-resource", roleHandler.GetPermissionsByResource)
				permissions.GET("/grouped", roleHandler.GetGroupedPermissions)
			}

			// 用户角色管理
			userRoles := authorized.Group("/users")
			{
				userRoles.POST("/:user_id/roles", middleware.RequirePermission("user.update"), roleHandler.AssignRoleToUser)
				userRoles.DELETE("/:user_id/roles/:role_id", middleware.RequirePermission("user.update"), roleHandler.RemoveRoleFromUser)
				userRoles.GET("/:user_id/roles", middleware.RequirePermission("user.read"), roleHandler.GetUserRoles)
				userRoles.PUT("/:user_id/roles/batch", middleware.RequirePermission("user.update"), roleHandler.BatchAssignRoles)
			}

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

				// Excel导入导出
				customers.GET("/export", middleware.RequirePermission("customer.read"), customerHandler.ExportCustomers)
				customers.GET("/template", middleware.RequirePermission("customer.read"), customerHandler.GetImportTemplate)
				customers.POST("/import", middleware.RequirePermission("customer.create"), customerHandler.ImportCustomers)
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

			// 合同管理
			contractHandler := handler.NewContractHandler()
			contracts := authorized.Group("/contracts")
			contracts.Use(middleware.TenantMiddleware())
			{
				contracts.POST("", middleware.RequirePermission("customer.create"), contractHandler.Create)
				contracts.GET("", middleware.RequirePermission("customer.read"), contractHandler.List)
				contracts.GET("/:id", middleware.RequirePermission("customer.read"), contractHandler.Get)
				contracts.PUT("/:id", middleware.RequirePermission("customer.update"), contractHandler.Update)
				contracts.DELETE("/:id", middleware.RequirePermission("customer.delete"), contractHandler.Delete)
				contracts.PUT("/:id/status", middleware.RequirePermission("customer.update"), contractHandler.UpdateStatus)
				contracts.GET("/customer/:customer_id", middleware.RequirePermission("customer.read"), contractHandler.GetByCustomer)
				contracts.GET("/statistics", middleware.RequirePermission("customer.read"), contractHandler.GetStatistics)
			}

			// 发票管理
			invoiceHandler := handler.NewInvoiceHandler()
			invoices := authorized.Group("/invoices")
			invoices.Use(middleware.TenantMiddleware())
			{
				invoices.POST("", middleware.RequirePermission("finance.create"), invoiceHandler.Create)
				invoices.GET("", middleware.RequirePermission("finance.read"), invoiceHandler.List)
				invoices.GET("/:id", middleware.RequirePermission("finance.read"), invoiceHandler.Get)
				invoices.PUT("/:id", middleware.RequirePermission("finance.update"), invoiceHandler.Update)
				invoices.DELETE("/:id", middleware.RequirePermission("finance.delete"), invoiceHandler.Delete)
				invoices.PUT("/:id/issue", middleware.RequirePermission("finance.verify"), invoiceHandler.Issue)
				invoices.PUT("/:id/cancel", middleware.RequirePermission("finance.verify"), invoiceHandler.Cancel)
				invoices.GET("/customer/:customer_id", middleware.RequirePermission("finance.read"), invoiceHandler.GetByCustomer)
				invoices.GET("/order/:order_id", middleware.RequirePermission("finance.read"), invoiceHandler.GetByOrder)
				invoices.GET("/statistics", middleware.RequirePermission("finance.read"), invoiceHandler.GetStatistics)
			}

			// 自定义表单管理
			formHandler := handler.NewFormHandler()
			forms := authorized.Group("/forms")
			forms.Use(middleware.TenantMiddleware())
			{
				// 表单定义管理
				forms.POST("", middleware.RequirePermission("system.create"), formHandler.Create)
				forms.GET("", middleware.RequirePermission("system.read"), formHandler.List)
				forms.GET("/:id", middleware.RequirePermission("system.read"), formHandler.Get)
				forms.GET("/code/:code", middleware.RequirePermission("system.read"), formHandler.GetByCode)
				forms.PUT("/:id", middleware.RequirePermission("system.update"), formHandler.Update)
				forms.DELETE("/:id", middleware.RequirePermission("system.delete"), formHandler.Delete)
				forms.GET("/statistics", middleware.RequirePermission("system.read"), formHandler.GetStatistics)

				// 表单数据管理
				forms.POST("/customer/:customer_id/submit", middleware.RequirePermission("customer.create"), formHandler.SubmitData)
				forms.GET("/data", middleware.RequirePermission("customer.read"), formHandler.QueryData)
				forms.GET("/data/:data_id", middleware.RequirePermission("customer.read"), formHandler.GetFormData)
				forms.GET("/data/customer/:customer_id", middleware.RequirePermission("customer.read"), formHandler.GetByCustomer)
				forms.GET("/data/form/:form_id", middleware.RequirePermission("customer.read"), formHandler.GetByForm)
				forms.GET("/export/:form_id", middleware.RequirePermission("customer.read"), formHandler.ExportData)
			}

			// 系统设置管理
			systemHandler := handler.NewSystemHandler()
			settings := authorized.Group("/settings")
			settings.Use(middleware.TenantMiddleware())
			{
				settings.POST("", middleware.RequirePermission("system.update"), systemHandler.SetSetting)
				settings.GET("", middleware.RequirePermission("system.read"), systemHandler.ListSettings)
				settings.GET("/:key", middleware.RequirePermission("system.read"), systemHandler.GetSetting)
				settings.DELETE("/:key", middleware.RequirePermission("system.delete"), systemHandler.DeleteSetting)
			}

			// 平台级系统设置（仅平台管理员）
			platformSettings := authorized.Group("/platform/settings")
			platformSettings.Use(middleware.RequirePlatformAdmin())
			{
				platformSettings.POST("", systemHandler.SetPlatformSetting)
				platformSettings.GET("/:key", systemHandler.GetPlatformSetting)
			}

			// 审计日志管理
			auditLogs := authorized.Group("/audit-logs")
			auditLogs.Use(middleware.TenantMiddleware())
			{
				auditLogs.POST("", middleware.RequirePermission("system.create"), systemHandler.CreateAuditLog)
				auditLogs.GET("", middleware.RequirePermission("system.read"), systemHandler.QueryAuditLogs)
				auditLogs.GET("/:id", middleware.RequirePermission("system.read"), systemHandler.GetAuditLog)
				auditLogs.GET("/statistics", middleware.RequirePermission("system.read"), systemHandler.GetAuditStatistics)
				auditLogs.POST("/clean", middleware.RequirePermission("system.delete"), systemHandler.CleanOldAuditLogs)
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
