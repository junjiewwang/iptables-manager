package main

import (
	"log"
	"os"
	"strings"

	"iptables-management-backend/config"
	"iptables-management-backend/handlers"
	"iptables-management-backend/middleware"
	"iptables-management-backend/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// 初始化数据库
	if err := config.InitDB(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// 创建服务实例
	authService := services.NewAuthService()
	ruleService := services.NewRuleService()
	logService := services.NewLogService()
	tableService := services.NewTableService()
	topologyService := services.NewTopologyService()
	networkService := services.NewNetworkService()

	// 创建默认用户
	if err := authService.CreateDefaultUsers(); err != nil {
		log.Printf("Failed to create default users: %v", err)
	}

	// 创建处理器实例
	authHandler := handlers.NewAuthHandler(authService, logService)
	ruleHandler := handlers.NewRuleHandler(ruleService, logService)
	logHandler := handlers.NewLogHandler(logService)
	tableHandler := handlers.NewTableHandler(tableService, logService)
	topologyHandler := handlers.NewTopologyHandler(topologyService)
	networkHandler := handlers.NewNetworkHandler(networkService, logService)

	// 创建Gin路由器
	r := gin.Default()

	// 配置CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"*"}
	r.Use(cors.New(config))

	// 静态文件服务 - 提供前端构建后的文件
	r.Static("/assets", "./dist/assets")
	r.StaticFile("/", "./dist/index.html")
	r.StaticFile("/favicon.ico", "./dist/favicon.ico")
	r.StaticFile("/vite.svg", "./dist/vite.svg")

	// 处理前端路由 - SPA 路由支持
	r.NoRoute(func(c *gin.Context) {
		// 如果请求的是 API 路径，返回 404
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(404, gin.H{"error": "API endpoint not found"})
			return
		}
		// 如果请求的是静态资源路径，返回 404
		if strings.HasPrefix(c.Request.URL.Path, "/assets/") {
			c.JSON(404, gin.H{"error": "Static file not found"})
			return
		}
		// 否则返回 index.html，让前端路由处理
		c.File("./dist/index.html")
	})

	// API路由组
	api := r.Group("/api")
	{
		// 认证路由
		api.POST("/login", authHandler.Login)

		// 需要认证的路由
		auth := api.Group("/")
		auth.Use(middleware.AuthMiddleware())
		{
			// 规则管理
			auth.GET("/rules", ruleHandler.GetRules)
			auth.GET("/rules/:id", ruleHandler.GetRule)
			auth.POST("/rules", ruleHandler.CreateRule)
			auth.PUT("/rules/:id", ruleHandler.UpdateRule)
			auth.DELETE("/rules/:id", ruleHandler.DeleteRule)

			auth.GET("/rules/system", ruleHandler.GetSystemRules)
			auth.POST("/rules/sync", ruleHandler.SyncSystemRules)

			// 统计信息
			auth.GET("/statistics", ruleHandler.GetStatistics)

			// 操作日志
			auth.GET("/logs", logHandler.GetLogs)

			// 表管理
			auth.GET("/tables", tableHandler.GetAllTables)
			auth.GET("/tables/:table", tableHandler.GetTableInfo)
			auth.GET("/tables/:table/chains/:chain", tableHandler.GetChainVerbose)
			auth.GET("/special-chains", tableHandler.GetSpecialChains)

			// 拓扑图
			auth.GET("/topology", topologyHandler.GetTopology)
			auth.GET("/topology/stats", topologyHandler.GetTopologyStats)
			auth.POST("/topology/refresh", topologyHandler.RefreshTopology)
			auth.GET("/topology/export", topologyHandler.ExportTopology)
			auth.GET("/topology/health", topologyHandler.GetTopologyHealth)

			// 网络接口管理
			auth.GET("/network/interfaces", networkHandler.GetInterfaces)
			auth.GET("/docker/bridges", networkHandler.GetDockerBridges)
			auth.GET("/bridges/:name/rules", networkHandler.GetBridgeRules)
			auth.GET("/network/connections", networkHandler.GetNetworkConnections)
			auth.GET("/network/routes", networkHandler.GetRouteTable)

			// 测试规则（模拟）
			auth.POST("/test-rule", func(c *gin.Context) {
				c.JSON(200, gin.H{"result": "规则测试通过"})
			})

			// 链管理（模拟）
			auth.POST("/chains/list", func(c *gin.Context) {
				c.JSON(200, gin.H{"chains": []string{"INPUT", "OUTPUT", "FORWARD"}})
			})
			auth.POST("/chains/create", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "链创建成功"})
			})
			auth.DELETE("/chains/:name", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "链删除成功"})
			})
			auth.POST("/chains/flush", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "链清空成功"})
			})
			auth.POST("/chains/policy", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "策略设置成功"})
			})

			// 备份和恢复（模拟）
			auth.POST("/backup", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "备份成功"})
			})
			auth.POST("/restore", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "恢复成功"})
			})
		}
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 获取端口
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
