package main

import (
	"log"
	"mahjong/internal/api"
	"mahjong/internal/auth"
	"mahjong/internal/lobby"
	"mahjong/internal/logger"
	"mahjong/internal/websocket"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化 logger（同時輸出到 stdout 和檔案）
	if err := logger.Init(); err != nil {
		log.Fatal("Logger 初始化失敗:", err)
	}
	defer logger.Close()

	// 初始化 JWKS 認證（如果 AUTH_PROXY_URL 有設定）
	if err := auth.InitJWKS(); err != nil {
		log.Printf("⚠️ JWKS 初始化失敗: %v（認證功能停用）", err)
	}
	defer auth.ShutdownJWKS()

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

	// 初始化 Lobby Notifier（如果 LOBBY_SERVICE_URL 有設定）
	lobbyServiceURL := os.Getenv("LOBBY_SERVICE_URL")
	lobbyInternalSecret := os.Getenv("LOBBY_INTERNAL_SECRET")
	if lobbyServiceURL == "" {
		lobbyServiceURL = "http://localhost:3001"
	}
	if lobbyInternalSecret == "" {
		log.Fatal("LOBBY_INTERNAL_SECRET environment variable not set. Please set it to a strong, unique secret.")
	}

	lobbyNotifier := lobby.NewHTTPLobbyNotifier(lobbyServiceURL, lobbyInternalSecret)
	hub.SetLobbyNotifier(lobbyNotifier)
	log.Printf("🔗 Lobby Notifier 已設置: %s", lobbyServiceURL)

	go hub.Run()

	// API 路由（使用可選 JWT 驗證）
	apiGroup := r.Group("/api")
	apiGroup.Use(auth.OptionalJWTMiddleware())
	{
		apiGroup.POST("/rooms/create", api.CreateRoom(hub))
		apiGroup.GET("/rooms/:roomId", api.GetRoom(hub))
	}

	// 受保護的 API 路由（需要 JWT）
	// protectedGroup := r.Group("/api/protected")
	// protectedGroup.Use(auth.JWTMiddleware())
	// {
	//     // 未來需要強制驗證的端點放這裡
	// }

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