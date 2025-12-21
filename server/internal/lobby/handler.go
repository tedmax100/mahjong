package lobby

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler 處理大廳相關的 HTTP 請求
type Handler struct {
	store         *LobbyStore
	hub           *LobbyHub
	gameServerURL string
	secretKey     string
}

// NewHandler 創建新的 Handler
func NewHandler(store *LobbyStore, hub *LobbyHub, gameServerURL, secretKey string) *Handler {
	return &Handler{
		store:         store,
		hub:           hub,
		gameServerURL: gameServerURL,
		secretKey:     secretKey,
	}
}

// ListRooms 獲取公開房間列表
// GET /api/lobby/rooms
func (h *Handler) ListRooms(c *gin.Context) {
	rooms := h.store.GetPublicRooms()

	c.JSON(http.StatusOK, ListRoomsResponse{
		Success: true,
		Rooms:   rooms,
	})
}

// CreateRoom 創建房間（代理到 Game Server）
// POST /api/lobby/rooms
func (h *Handler) CreateRoom(c *gin.Context) {
	var req CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, CreateRoomResponse{
			Success: false,
			Error:   "無效的請求",
		})
		return
	}

	// 代理請求到 Game Server
	body, _ := json.Marshal(req)
	proxyReq, err := http.NewRequest("POST", h.gameServerURL+"/api/rooms/create", bytes.NewBuffer(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, CreateRoomResponse{
			Success: false,
			Error:   "無法創建請求",
		})
		return
	}

	proxyReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(proxyReq)
	if err != nil {
		log.Printf("[LobbyHandler] 代理請求失敗: %v", err)
		c.JSON(http.StatusServiceUnavailable, CreateRoomResponse{
			Success: false,
			Error:   "遊戲伺服器無法連線",
		})
		return
	}
	defer resp.Body.Close()

	// 讀取並返回 Game Server 的響應
	respBody, _ := io.ReadAll(resp.Body)

	var gameResp struct {
		Success  bool   `json:"success"`
		RoomID   string `json:"roomId"`
		IsPublic bool   `json:"isPublic"`
		Error    string `json:"error,omitempty"`
	}
	if err := json.Unmarshal(respBody, &gameResp); err != nil {
		c.JSON(http.StatusInternalServerError, CreateRoomResponse{
			Success: false,
			Error:   "無法解析響應",
		})
		return
	}

	c.JSON(resp.StatusCode, CreateRoomResponse{
		Success:    gameResp.Success,
		RoomID:     gameResp.RoomID,
		ServerAddr: h.gameServerURL,
		Error:      gameResp.Error,
	})
}

// GetRoom 獲取單個房間資訊
// GET /api/lobby/rooms/:id
func (h *Handler) GetRoom(c *gin.Context) {
	roomID := c.Param("id")

	room, exists := h.store.GetRoom(roomID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "房間不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"room":    room,
	})
}

// HandleRoomEvent 接收 Game Server 的房間事件
// POST /internal/room-events
func (h *Handler) HandleRoomEvent(c *gin.Context) {
	// 驗證內部請求
	secret := c.GetHeader("X-Internal-Secret")
	if secret != h.secretKey {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "未授權",
		})
		return
	}

	var event RoomEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "無效的請求",
		})
		return
	}

	log.Printf("[LobbyHandler] 收到房間事件: %s (房間: %s)", event.Event, event.RoomID)

	// 根據事件類型更新存儲
	switch event.Event {
	case EventRoomCreated:
		if event.Room != nil {
			h.store.AddRoom(event.Room)
			// 廣播房間列表更新
			if h.hub != nil {
				h.hub.BroadcastRoomList()
			}
		}

	case EventPlayerJoined, EventPlayerLeft:
		if event.Room != nil {
			h.store.UpdateRoom(event.RoomID, func(room *LobbyRoom) {
				room.PlayerCount = event.Room.PlayerCount
				room.Status = event.Room.Status
			})
			// 廣播房間列表更新
			if h.hub != nil {
				h.hub.BroadcastRoomList()
			}
		}

	case EventGameStarted:
		// 遊戲開始，更新狀態（會從公開列表中移除，因為狀態不再是 waiting）
		h.store.UpdateRoom(event.RoomID, func(room *LobbyRoom) {
			room.Status = StatusPlaying
		})
		// 廣播房間列表更新
		if h.hub != nil {
			h.hub.BroadcastRoomList()
		}

	case EventRoomClosed:
		h.store.RemoveRoom(event.RoomID)
		// 廣播房間列表更新
		if h.hub != nil {
			h.hub.BroadcastRoomList()
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

// GetRecentMessages 獲取最近的聊天訊息
// GET /api/lobby/messages
func (h *Handler) GetRecentMessages(c *gin.Context) {
	messages := h.store.GetRecentMessages(50)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"messages": messages,
	})
}

// InternalAuthMiddleware 內部 API 認證中間件
func InternalAuthMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		secret := c.GetHeader("X-Internal-Secret")
		if secret != secretKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "未授權",
			})
			return
		}
		c.Next()
	}
}
