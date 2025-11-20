package main

import (
	"log"
	"mahjong/internal/api"
	"mahjong/internal/websocket"

	"github.com/gin-gonic/gin"
)

func main() {
	// 创建Gin路由
	r := gin.Default()

	// CORS中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 创建游戏服务器
	hub := websocket.NewHub()
	go hub.Run()

	// API路由
	apiGroup := r.Group("/api")
	{
		apiGroup.POST("/rooms/create", api.CreateRoom(hub))
		apiGroup.GET("/rooms/:roomId", api.GetRoom(hub))
	}

	// WebSocket路由
	r.GET("/ws", func(c *gin.Context) {
		websocket.ServeWs(hub, c.Writer, c.Request)
	})

	// 提供静态文件（生产环境）
	r.Static("/assets", "./public/assets")
	r.StaticFile("/", "./public/index.html")

	// 启动服务器
	log.Println("🀄 麻将服务器启动在 http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}
