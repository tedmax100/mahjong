package lobby

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	// 寫入超時
	writeWait = 10 * time.Second

	// 讀取超時（Pong 超時）
	pongWait = 60 * time.Second

	// Ping 間隔（必須小於 pongWait）
	pingPeriod = (pongWait * 9) / 10

	// 最大訊息大小
	maxMessageSize = 512

	// 聊天頻率限制
	chatRateLimit = 1 * time.Second

	// 聊天訊息最大長度
	chatMaxLength = 200
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允許所有來源（生產環境應該檢查）
	},
}

// LobbyClient 大廳 WebSocket 客戶端
type LobbyClient struct {
	hub      *LobbyHub
	conn     *websocket.Conn
	send     chan []byte
	userID   string
	userName string
	lastChat time.Time
}

// ClientMessage 客戶端發送的訊息
type ClientMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// readPump 讀取訊息
func (c *LobbyClient) readPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[LobbyClient] 讀取錯誤: %v", err)
			}
			break
		}

		c.handleMessage(message)
	}
}

// writePump 寫入訊息
func (c *LobbyClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 批量發送隊列中的訊息
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage 處理客戶端訊息
func (c *LobbyClient) handleMessage(message []byte) {
	var msg ClientMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("[LobbyClient] 解析訊息失敗: %v", err)
		return
	}

	switch msg.Type {
	case "chat":
		c.handleChat(msg.Data)
	case "get_rooms":
		// 重新獲取房間列表
		c.hub.sendRoomListTo(c)
	}
}

// handleChat 處理聊天訊息
func (c *LobbyClient) handleChat(data json.RawMessage) {
	var req ChatRequest
	if err := json.Unmarshal(data, &req); err != nil {
		c.sendError("無效的聊天請求")
		return
	}

	// 檢查發言頻率
	if time.Since(c.lastChat) < chatRateLimit {
		c.sendError("發言太頻繁，請稍後再試")
		return
	}

	// 檢查訊息長度
	content := req.Content
	if utf8.RuneCountInString(content) > chatMaxLength {
		c.sendError("訊息太長")
		return
	}

	// 檢查訊息內容
	if content == "" {
		return
	}

	// 更新最後發言時間
	c.lastChat = time.Now()

	// 添加聊天訊息到存儲
	chatMsg := c.hub.store.AddChatMessage(c.userID, c.userName, content)

	// 廣播聊天訊息
	c.hub.BroadcastChatMessage(chatMsg)

	log.Printf("[LobbyClient] 聊天訊息: %s: %s", c.userName, content)
}

// sendError 發送錯誤訊息
func (c *LobbyClient) sendError(message string) {
	msg := WSMessage{
		Type: WSTypeError,
		Data: map[string]interface{}{
			"message": message,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	select {
	case c.send <- data:
	default:
	}
}

// LobbyWsHandler WebSocket 處理器
type LobbyWsHandler struct {
	hub *LobbyHub
}

// NewLobbyWsHandler 創建 WebSocket 處理器
func NewLobbyWsHandler(hub *LobbyHub) *LobbyWsHandler {
	return &LobbyWsHandler{hub: hub}
}

// ServeWs 處理 WebSocket 升級請求
func (h *LobbyWsHandler) ServeWs(c *gin.Context) {
	userID := c.Query("userId")
	userName := c.Query("userName")

	if userID == "" {
		userID = "anonymous_" + time.Now().Format("20060102150405")
	}
	if userName == "" {
		userName = "匿名玩家"
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[LobbyWs] 升級連接失敗: %v", err)
		return
	}

	client := &LobbyClient{
		hub:      h.hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		userID:   userID,
		userName: userName,
	}

	h.hub.Register(client)

	go client.writePump()
	go client.readPump()
}
