package api

import (
	"mahjong/internal/websocket"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateRoomRequest 创建房间请求
type CreateRoomRequest struct {
	UserID   string `json:"userId"`
	UserName string `json:"userName"`
}

// CreateRoom 创建房间API
func CreateRoom(hub *websocket.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateRoomRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "无效的请求",
			})
			return
		}

		// 生成房间ID（6位数字）
		roomID := generateRoomID()

		// 创建房间
		hub.CreateRoom(roomID)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"roomId":  roomID,
		})
	}
}

// GetRoom 获取房间信息API
func GetRoom(hub *websocket.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		roomID := c.Param("roomId")

		room := hub.GetRoom(roomID)
		if room == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "房间不存在",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"room":    room,
		})
	}
}

// generateRoomID 生成6位房间号
func generateRoomID() string {
	// 使用UUID的前6个字符作为房间号
	id := uuid.New().String()
	return id[:6]
}
