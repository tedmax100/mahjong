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
	"github.com/joho/godotenv"
)

func main() {
	// 載入 .env 檔案（如果存在）
	if err := godotenv.Load(); err != nil {
		log.Println("提示: 未找到 .env 檔案，使用系統環境變數")
	}

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
	lobbyServiceURL := getEnv("LOBBY_SERVICE_URL", "http://localhost:3001")
	lobbyInternalSecret := getEnv("LOBBY_INTERNAL_SECRET", "dev-internal-secret")

	// 檢查生產環境安全性
	if lobbyInternalSecret == "dev-internal-secret" {
		if gin.Mode() == gin.ReleaseMode {
			log.Fatal("❌ 錯誤: 在生產環境 (release mode) 下必須設定 LOBBY_INTERNAL_SECRET")
		}
		log.Println("⚠️ LOBBY_INTERNAL_SECRET 未設定，使用開發環境預設值")
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

// getEnv 取得環境變數，如果不存在則回傳預設值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}