package api

import (
	"mahjong/internal/websocket"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateRoomRequest 建立房間請求
type CreateRoomRequest struct {
	UserID   string `json:"userId"`
	UserName string `json:"userName"`
}

// CreateRoom 建立房間 API
func CreateRoom(hub *websocket.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateRoomRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "無效的請求",
			})
			return
		}

		// 生成房間 ID（6 位數字）
		roomID := generateRoomID()

		// 建立房間
		hub.CreateRoom(roomID)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"roomId":  roomID,
		})
	}
}

// GetRoom 取得房間資訊 API
func GetRoom(hub *websocket.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		roomID := c.Param("roomId")

		room := hub.GetRoom(roomID)
		if room == nil {
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
}

// generateRoomID 生成 6 位房間號
func generateRoomID() string {
	// 使用 UUID 的前 6 個字元作為房間號
	id := uuid.New().String()
	return id[:6]
}