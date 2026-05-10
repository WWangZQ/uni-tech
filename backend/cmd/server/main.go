package main

import (
	"fmt"
	"log"

	"smart-campus/internal/handler/user"
	"smart-campus/internal/handler/space"
	"smart-campus/internal/handler/seckill"
	"smart-campus/internal/handler/order"
	"smart-campus/internal/pkg/config"
	"smart-campus/internal/pkg/database"
	"smart-campus/internal/pkg/jwt"
	"smart-campus/internal/pkg/middleware"
	redisClient "smart-campus/internal/pkg/redis"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化数据库
	db, err := database.NewDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	// 初始化 Redis
	rdb, err := redisClient.NewClient(&cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect redis: %v", err)
	}

	// 初始化 JWT
	jwtManager := jwt.New(cfg.JWT.Secret, cfg.JWT.ExpireHour)

	// 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 创建路由
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 用户模块
		userHandler := user.NewHandler(db, jwtManager)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
		}

		users := v1.Group("/users")
		users.Use(middleware.AuthMiddleware(jwtManager))
		{
			users.GET("/me", userHandler.GetCurrentUser)
			users.GET("/credits", userHandler.GetCreditScore)
			users.GET("/quota", userHandler.GetQuota)
		}

		// 空间模块
		spaceHandler := space.NewHandler(db)
		spaces := v1.Group("/spaces")
		spaces.Use(middleware.AuthMiddleware(jwtManager))
		{
			spaces.GET("", spaceHandler.ListSpaces)
			spaces.GET("/:id", spaceHandler.GetSpace)
			spaces.GET("/:id/slots", spaceHandler.GetAvailableSlots)
			spaces.POST("/bookings", spaceHandler.CreateBooking)
		}
		bookings := v1.Group("/bookings")
		bookings.Use(middleware.AuthMiddleware(jwtManager))
		{
			bookings.GET("", spaceHandler.ListBookings)
			bookings.GET("/:id", spaceHandler.GetBooking)
			bookings.DELETE("/:id", spaceHandler.CancelBooking)
		}

		// 活动秒杀模块
		seckillHandler := seckill.NewHandler(db, rdb)
		activities := v1.Group("/activities")
		{
			activities.GET("", seckillHandler.ListActivities)
			activities.GET("/:id", seckillHandler.GetActivity)
		}
		activitiesSecured := activities.Group("")
		activitiesSecured.Use(middleware.AuthMiddleware(jwtManager))
		{
			activitiesSecured.POST("/:id/seckill", seckillHandler.DoSeckill)
			activitiesSecured.GET("/:id/ticket", seckillHandler.GetMyTicket)
		}

		// 订单模块
		orderHandler := order.NewHandler(db)
		orders := v1.Group("/orders")
		orders.Use(middleware.AuthMiddleware(jwtManager))
		{
			orders.GET("", orderHandler.ListOrders)
			orders.GET("/:id", orderHandler.GetOrder)
			orders.POST("/:id/pay", orderHandler.PayOrder)
			orders.POST("/:id/cancel", orderHandler.CancelOrder)
		}
	}

	// 启动服务
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
