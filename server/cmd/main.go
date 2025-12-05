package main

import (
	"log"
	"mahjong/internal/api"
	"mahjong/internal/websocket"

	"github.com/gin-gonic/gin"
)

func main() {
	// 建立 Gin 路由
	r := gin.Default()

	// CORS 中介軟體
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

	// 建立遊戲伺服器
	hub := websocket.NewHub()
	go hub.Run()

	// API 路由
	apiGroup := r.Group("/api")
	{
		apiGroup.POST("/rooms/create", api.CreateRoom(hub))
		apiGroup.GET("/rooms/:roomId", api.GetRoom(hub))
	}

	// WebSocket 路由
	r.GET("/ws", func(c *gin.Context) {
		websocket.ServeWs(hub, c.Writer, c.Request)
	})

	// 提供靜態檔案（生產環境）
	r.Static("/assets", "./public/assets")
	r.StaticFile("/", "./public/index.html")

	// 啟動伺服器
	log.Println("🀄 麻將伺服器啟動於 http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("伺服器啟動失敗:", err)
	}
}