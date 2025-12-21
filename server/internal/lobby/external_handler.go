package lobby

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// ExternalServerClaims 外部伺服器 JWT Claims
type ExternalServerClaims struct {
	ServerID string `json:"server_id"`
	jwt.RegisteredClaims
}

// TokenExpiration Token 有效期
const TokenExpiration = 24 * time.Hour

// ExternalServerHandler 處理外部伺服器相關的 HTTP 請求
type ExternalServerHandler struct {
	store         *ExternalServerStore
	lobbyStore    *LobbyStore
	hub           *LobbyHub
	internalSecret string // 用於驗證註冊請求
	jwtSecret     string  // 用於簽發/驗證 JWT Token
}

// NewExternalServerHandler 創建新的外部伺服器處理器
func NewExternalServerHandler(
	store *ExternalServerStore,
	lobbyStore *LobbyStore,
	hub *LobbyHub,
	internalSecret string,
	jwtSecret string,
) *ExternalServerHandler {
	return &ExternalServerHandler{
		store:          store,
		lobbyStore:     lobbyStore,
		hub:            hub,
		internalSecret: internalSecret,
		jwtSecret:      jwtSecret,
	}
}

// Register 註冊外部伺服器
// POST /internal/external-servers/register
func (h *ExternalServerHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, RegisterResponse{
			Success: false,
			Error:   "無效的請求格式",
		})
		return
	}

	// 驗證共享密鑰
	if req.Secret != h.internalSecret {
		c.JSON(http.StatusUnauthorized, RegisterResponse{
			Success: false,
			Error:   "未授權：密鑰無效",
		})
		return
	}

	// 檢查是否已註冊
	if h.store.Exists(req.ServerID) {
		// 已存在，更新資訊
		log.Printf("[ExternalServer] 伺服器 %s 重新註冊", req.ServerID)
	}

	// 生成 JWT Token
	token, err := h.generateServerToken(req.ServerID)
	if err != nil {
		log.Printf("[ExternalServer] 生成 Token 失敗: %v", err)
		c.JSON(http.StatusInternalServerError, RegisterResponse{
			Success: false,
			Error:   "無法生成 Token",
		})
		return
	}

	// 儲存伺服器資訊
	server := &ExternalServer{
		ID:            req.ServerID,
		DisplayName:   req.DisplayName,
		IP:            req.IP,
		Port:          req.Port,
		WebURL:        req.WebURL,
		MaxRooms:      req.MaxRooms,
		CurrentRooms:  0,
		Status:        ServerStatusOnline,
		LastHeartbeat: time.Now(),
		RegisteredAt:  time.Now(),
		Token:         token,
	}
	h.store.Add(server)

	log.Printf("[ExternalServer] 伺服器已註冊: %s (%s) @ %s", req.ServerID, req.DisplayName, req.WebURL)

	c.JSON(http.StatusOK, RegisterResponse{
		Success:   true,
		Token:     token,
		ExpiresIn: int64(TokenExpiration.Seconds()),
	})
}

// Heartbeat 心跳更新
// POST /internal/external-servers/:serverId/heartbeat
func (h *ExternalServerHandler) Heartbeat(c *gin.Context) {
	serverID := c.Param("serverId")

	// 驗證 JWT Token
	claims, err := h.validateToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, HeartbeatResponse{
			Success: false,
			Error:   "Token 無效：" + err.Error(),
		})
		return
	}

	// 驗證 serverID 匹配
	if claims.ServerID != serverID {
		c.JSON(http.StatusForbidden, HeartbeatResponse{
			Success: false,
			Error:   "Token 與伺服器 ID 不匹配",
		})
		return
	}

	// 更新心跳
	if !h.store.UpdateHeartbeat(serverID) {
		c.JSON(http.StatusNotFound, HeartbeatResponse{
			Success: false,
			Error:   "伺服器未註冊",
		})
		return
	}

	// 更新房間數量（如果提供）
	var req HeartbeatRequest
	if err := c.ShouldBindJSON(&req); err == nil {
		h.store.UpdateRoomCount(serverID, req.CurrentRooms)
	}

	c.JSON(http.StatusOK, HeartbeatResponse{
		Success:   true,
		ExpiresIn: int64(TokenExpiration.Seconds()),
	})
}

// HandleRoomEvent 處理外部伺服器的房間事件
// POST /internal/external-servers/:serverId/room-events
func (h *ExternalServerHandler) HandleRoomEvent(c *gin.Context) {
	serverID := c.Param("serverId")

	// 驗證 JWT Token
	claims, err := h.validateToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Token 無效：" + err.Error(),
		})
		return
	}

	// 驗證 serverID 匹配
	if claims.ServerID != serverID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Token 與伺服器 ID 不匹配",
		})
		return
	}

	// 獲取伺服器資訊
	server := h.store.Get(serverID)
	if server == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "伺服器未註冊",
		})
		return
	}

	var event RoomEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "無效的請求格式",
		})
		return
	}

	log.Printf("[ExternalServer] 收到伺服器 %s 的房間事件: %s (房間: %s)", serverID, event.Event, event.RoomID)

	// 設定房間的伺服器資訊
	if event.Room != nil {
		event.Room.ServerAddr = server.WebURL
		event.Room.ServerID = serverID
		event.Room.IsExternal = true
	}

	// 根據事件類型更新存儲
	switch event.Event {
	case EventRoomCreated:
		if event.Room != nil {
			h.lobbyStore.AddRoom(event.Room)
			if h.hub != nil {
				h.hub.BroadcastRoomList()
			}
		}

	case EventPlayerJoined, EventPlayerLeft:
		if event.Room != nil {
			h.lobbyStore.UpdateRoom(event.RoomID, func(room *LobbyRoom) {
				room.PlayerCount = event.Room.PlayerCount
				room.Status = event.Room.Status
			})
			if h.hub != nil {
				h.hub.BroadcastRoomList()
			}
		}

	case EventGameStarted:
		h.lobbyStore.UpdateRoom(event.RoomID, func(room *LobbyRoom) {
			room.Status = StatusPlaying
		})
		if h.hub != nil {
			h.hub.BroadcastRoomList()
		}

	case EventRoomClosed:
		h.lobbyStore.RemoveRoom(event.RoomID)
		if h.hub != nil {
			h.hub.BroadcastRoomList()
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

// Deregister 註銷外部伺服器
// DELETE /internal/external-servers/:serverId
func (h *ExternalServerHandler) Deregister(c *gin.Context) {
	serverID := c.Param("serverId")

	// 驗證 JWT Token
	claims, err := h.validateToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Token 無效：" + err.Error(),
		})
		return
	}

	// 驗證 serverID 匹配
	if claims.ServerID != serverID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Token 與伺服器 ID 不匹配",
		})
		return
	}

	// 移除伺服器的所有房間
	h.removeServerRooms(serverID)

	// 移除伺服器
	h.store.Remove(serverID)

	log.Printf("[ExternalServer] 伺服器已註銷: %s", serverID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

// ListServers 列出所有外部伺服器（管理用）
// GET /internal/external-servers
func (h *ExternalServerHandler) ListServers(c *gin.Context) {
	servers := h.store.GetAllServers()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"servers": servers,
		"count":   len(servers),
	})
}

// generateServerToken 生成伺服器 JWT Token
func (h *ExternalServerHandler) generateServerToken(serverID string) (string, error) {
	claims := ExternalServerClaims{
		ServerID: serverID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "mahjong-lobby",
			Subject:   serverID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExpiration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}

// validateToken 驗證 JWT Token
func (h *ExternalServerHandler) validateToken(c *gin.Context) (*ExternalServerClaims, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil, jwt.ErrTokenMalformed
	}

	// 移除 "Bearer " 前綴
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return nil, jwt.ErrTokenMalformed
	}

	token, err := jwt.ParseWithClaims(tokenString, &ExternalServerClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(h.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*ExternalServerClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrTokenInvalidClaims
}

// removeServerRooms 移除指定伺服器的所有房間
func (h *ExternalServerHandler) removeServerRooms(serverID string) {
	rooms := h.lobbyStore.GetPublicRooms()
	for _, room := range rooms {
		if room.ServerID == serverID {
			h.lobbyStore.RemoveRoom(room.ID)
		}
	}

	// 廣播更新
	if h.hub != nil {
		h.hub.BroadcastRoomList()
	}
}

// OnServerOffline 伺服器離線時的處理
func (h *ExternalServerHandler) OnServerOffline(serverID string) {
	log.Printf("[ExternalServer] 伺服器 %s 心跳超時，標記為離線", serverID)
	h.removeServerRooms(serverID)
}
