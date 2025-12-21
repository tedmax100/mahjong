package main

import (
	"log"
	"os"
	"time"

	"mahjong/internal/auth"
	"mahjong/internal/lobby"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 載入 .env 檔案（如果存在）
	if err := godotenv.Load(); err != nil {
		log.Println("提示: 未找到 .env 檔案，使用系統環境變數")
	}

	// 從環境變數取得設定
	port := getEnv("PORT", "3001")
	authProxyURL := getEnv("AUTH_PROXY_URL", "http://localhost:3001")
	gameServerURL := getEnv("GAME_SERVER_URL", "http://localhost:8080")
	lobbyInternalSecret := getEnv("LOBBY_INTERNAL_SECRET", "dev-internal-secret")
	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	sessionSecret := getEnv("SESSION_SECRET", "dev-secret")

	// 驗證必要設定
	if googleClientID == "" || googleClientSecret == "" {
		log.Fatal("錯誤: GOOGLE_CLIENT_ID 和 GOOGLE_CLIENT_SECRET 必須設定")
	}

	if sessionSecret == "dev-secret" && os.Getenv("GIN_MODE") == "release" {
		log.Fatal("錯誤: 生產環境必須設定 SESSION_SECRET")
	}

	// 建立金鑰管理器
	keyManager := auth.NewKeyManager("mahjong-auth-proxy", 12*time.Hour)
	if err := keyManager.Start(1 * time.Hour); err != nil {
		log.Fatalf("金鑰管理器啟動失敗: %v", err)
	}
	defer keyManager.Stop()

	// 建立 OAuth 處理器
	oauthConfig := &auth.OAuthConfig{
		ClientID:     googleClientID,
		ClientSecret: googleClientSecret,
		RedirectURI:  authProxyURL + "/auth/google/callback",
	}
	oauthHandler := auth.NewOAuthHandler(oauthConfig, keyManager)

	// 建立 JWKS 處理器
	jwksHandler := auth.NewJWKSHandler(keyManager)

	// 建立大廳組件
	lobbyStore := lobby.NewLobbyStore()
	lobbyHub := lobby.NewLobbyHub(lobbyStore)
	lobbyHandler := lobby.NewHandler(lobbyStore, lobbyHub, gameServerURL, lobbyInternalSecret)
	lobbyWsHandler := lobby.NewLobbyWsHandler(lobbyHub)

	// 啟動大廳 Hub
	go lobbyHub.Run()

	// 建立 Gin 路由
	r := gin.Default()

	// CORS 中介軟體
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 健康檢查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":      "ok",
			"service":     "mahjong-lobby",
			"onlineCount": lobbyHub.GetOnlineCount(),
			"roomCount":   lobbyStore.PublicRoomCount(),
		})
	})

	// Auth Proxy 路由
	r.GET("/login", oauthHandler.LoginHandler)
	r.GET("/auth/google/callback", oauthHandler.CallbackHandler)
	r.GET("/.well-known/jwks.json", jwksHandler.Handler)

	// 大廳 API 路由（使用可選認證）
	lobbyGroup := r.Group("/api/lobby")
	lobbyGroup.Use(auth.OptionalJWTMiddleware())
	{
		lobbyGroup.GET("/rooms", lobbyHandler.ListRooms)
		lobbyGroup.POST("/rooms", lobbyHandler.CreateRoom)
		lobbyGroup.GET("/rooms/:id", lobbyHandler.GetRoom)
		lobbyGroup.GET("/messages", lobbyHandler.GetRecentMessages)
	}

	// 內部 API（驗證共享密鑰）
	internalGroup := r.Group("/internal")
	internalGroup.Use(lobby.InternalAuthMiddleware(lobbyInternalSecret))
	{
		internalGroup.POST("/room-events", lobbyHandler.HandleRoomEvent)
	}

	// 大廳 WebSocket
	r.GET("/ws/lobby", lobbyWsHandler.ServeWs)

	// 啟動伺服器
	log.Println("")
	log.Println("🏠 Lobby & Auth Proxy Server 啟動成功")
	log.Println("=====================================")
	log.Printf("   URL:         http://localhost:%s", port)
	log.Printf("   Proxy:       %s", authProxyURL)
	log.Printf("   Game Server: %s", gameServerURL)
	log.Printf("   健康檢查:    http://localhost:%s/health", port)
	log.Printf("   JWKS:        http://localhost:%s/.well-known/jwks.json", port)
	log.Printf("   大廳 API:    http://localhost:%s/api/lobby/rooms", port)
	log.Printf("   大廳 WS:     ws://localhost:%s/ws/lobby", port)
	log.Println("=====================================")
	log.Println("")

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("伺服器啟動失敗: %v", err)
	}
}

// getEnv 取得環境變數，如果不存在則回傳預設值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
