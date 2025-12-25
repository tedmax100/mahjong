package main

import (
	"log"
	"os"
	"time"

	"mahjong/internal/auth"
	"mahjong/internal/lobby"
	"mahjong/internal/version"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 印出版本資訊
	version.Print("Mahjong Lobby Server")

	// 載入 .env 檔案（如果存在）
	if err := godotenv.Load(); err != nil {
		log.Println("提示: 未找到 .env 檔案，使用系統環境變數")
	}

	// 從環境變數取得設定
	port := getEnv("PORT", "3001")
	authProxyURL := getEnv("AUTH_PROXY_URL", "http://localhost:3001")
	gameServerURL := getEnv("GAME_SERVER_URL", "http://localhost:8080")
	gameClientURL := getEnv("GAME_CLIENT_URL", "") // 公開給前端的遊戲 URL（留空則使用 gameServerURL）

	// 取得內部密鑰
	lobbyInternalSecret := os.Getenv("LOBBY_INTERNAL_SECRET")
	if lobbyInternalSecret == "" {
		if gin.Mode() == gin.ReleaseMode {
			log.Fatal("錯誤: 在生產環境 (release mode) 下必須設定 LOBBY_INTERNAL_SECRET")
		}
		lobbyInternalSecret = "dev-internal-secret" // #nosec G101
		log.Println("⚠️ LOBBY_INTERNAL_SECRET 未設定，使用開發環境預設值")
	}

	externalServerSecret := os.Getenv("EXTERNAL_SERVER_SECRET")
	if externalServerSecret == "" {
		if gin.Mode() == gin.ReleaseMode {
			log.Fatal("錯誤: 在生產環境 (release mode) 下必須設定 EXTERNAL_SERVER_SECRET")
		}
		externalServerSecret = "dev-external-secret" // #nosec G101
		log.Println("⚠️ EXTERNAL_SERVER_SECRET 未設定，使用開發環境預設值")
	}

	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		if gin.Mode() == gin.ReleaseMode {
			log.Fatal("錯誤: 生產環境必須設定 SESSION_SECRET")
		}
		sessionSecret = "dev-secret" // #nosec G101
		log.Println("⚠️ SESSION_SECRET 未設定，使用開發環境預設值")
	}

	// 驗證必要設定
	if googleClientID == "" || googleClientSecret == "" {
		log.Fatal("錯誤: GOOGLE_CLIENT_ID 和 GOOGLE_CLIENT_SECRET 必須設定")
	}

	// 建立金鑰管理器 (Access Token 1 小時有效期)
	keyManager := auth.NewKeyManager("mahjong-auth-proxy", 1*time.Hour)
	if err := keyManager.Start(1 * time.Hour); err != nil {
		log.Fatalf("金鑰管理器啟動失敗: %v", err)
	}
	defer keyManager.Stop()

	// 建立 Refresh Token 儲存 (1 天有效期)
	refreshTokenStore := auth.NewRefreshTokenStore(24 * time.Hour)
	defer refreshTokenStore.Stop()

	// 建立 OAuth 處理器
	oauthConfig := &auth.OAuthConfig{
		ClientID:     googleClientID,
		ClientSecret: googleClientSecret,
		RedirectURI:  authProxyURL + "/auth/google/callback",
	}
	oauthHandler := auth.NewOAuthHandler(oauthConfig, keyManager, refreshTokenStore)

	// 建立 Refresh 和 Logout 處理器
	refreshHandler := auth.NewRefreshHandler(refreshTokenStore, keyManager)
	logoutHandler := auth.NewLogoutHandler(refreshTokenStore)

	// 建立 JWKS 處理器
	jwksHandler := auth.NewJWKSHandler(keyManager)

	// 建立大廳組件
	lobbyStore := lobby.NewLobbyStore()
	lobbyHub := lobby.NewLobbyHub(lobbyStore)
	lobbyHandler := lobby.NewHandler(lobbyStore, lobbyHub, gameServerURL, gameClientURL, lobbyInternalSecret)
	lobbyWsHandler := lobby.NewLobbyWsHandler(lobbyHub)

	// 建立外部伺服器組件
	externalServerStore := lobby.NewExternalServerStore(externalServerSecret)
	externalServerHandler := lobby.NewExternalServerHandler(
		externalServerStore,
		lobbyStore,
		lobbyHub,
		lobbyInternalSecret,
		externalServerSecret,
	)
	// 設置伺服器離線回調
	externalServerStore.SetOnOfflineCallback(externalServerHandler.OnServerOffline)
	// 啟動心跳監控
	externalServerStore.StartMonitor()

	// 啟動大廳 Hub
	go lobbyHub.Run()

	// 建立 Gin 路由
	r := gin.Default()

	// CORS 中介軟體
	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 驗證 origin 是否在白名單中
		if origin != "" && auth.ValidateOrigin(origin) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		} else if origin == "" {
			// 沒有 Origin header 的請求（例如同源請求）
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}

		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
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
			"status":              "ok",
			"service":             "mahjong-lobby",
			"onlineCount":         lobbyHub.GetOnlineCount(),
			"roomCount":           lobbyStore.PublicRoomCount(),
			"externalServerCount": externalServerStore.OnlineCount(),
		})
	})

	// Auth Proxy 路由
	r.GET("/login", oauthHandler.LoginHandler)
	r.GET("/auth/google/callback", oauthHandler.CallbackHandler)
	r.POST("/auth/refresh", refreshHandler.Handler)
	r.POST("/auth/logout", logoutHandler.Handler)
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

	// 外部伺服器 API
	externalGroup := r.Group("/internal/external-servers")
	{
		// 註冊需要共享密鑰
		externalGroup.POST("/register", externalServerHandler.Register)
		// 其他操作需要 JWT Token（在 handler 中驗證）
		externalGroup.POST("/:serverId/heartbeat", externalServerHandler.Heartbeat)
		externalGroup.POST("/:serverId/room-events", externalServerHandler.HandleRoomEvent)
		externalGroup.DELETE("/:serverId", externalServerHandler.Deregister)
		// 管理端點（需要內部密鑰）
		externalGroup.GET("", lobby.InternalAuthMiddleware(lobbyInternalSecret), externalServerHandler.ListServers)
	}

	// 大廳 WebSocket
	r.GET("/ws/lobby", lobbyWsHandler.ServeWs)

	// 提供靜態檔案（生產環境 Bundle 部署）
	staticDir := getEnv("STATIC_DIR", "")
	if staticDir != "" {
		r.Static("/assets", staticDir+"/assets")
		r.StaticFile("/", staticDir+"/index.html")
		// SPA fallback - 所有非 API/WS 路由都返回 index.html
		r.NoRoute(func(c *gin.Context) {
			c.File(staticDir + "/index.html")
		})
		log.Printf("   靜態文件:    %s", staticDir)
	}

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
	log.Printf("   外部伺服器:  http://localhost:%s/internal/external-servers", port)
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
