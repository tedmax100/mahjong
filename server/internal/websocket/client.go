package websocket

import (
	"encoding/json"
	"log"
	"mahjong/internal/game"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源（生产环境应该限制）
	},
}

// Client 代表一个WebSocket客户端
type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan []byte
	RoomID   string
	UserID   string
	UserName string
	Room     *game.Room
}

// Message 表示客户端消息
type Message struct {
	Type   string                 `json:"type"`
	Action string                 `json:"action,omitempty"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

// readPump 从WebSocket连接读取消息
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket错误: %v", err)
			}
			break
		}

		// 处理消息
		c.handleMessage(message)
	}
}

// writePump 向WebSocket连接写入消息
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage 处理收到的消息
func (c *Client) handleMessage(data []byte) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("解析消息失败: %v", err)
		return
	}

	log.Printf("收到消息: %s from %s", msg.Type, c.UserName)

	switch msg.Type {
	case "join":
		// 加入房间消息在连接时已处理
		log.Printf("玩家 %s 加入", c.UserName)

	case "action":
		// 处理游戏动作
		c.handleGameAction(msg.Action, msg.Data)

	default:
		log.Printf("未知消息类型: %s", msg.Type)
	}
}

// handleGameAction 处理游戏动作
func (c *Client) handleGameAction(action string, data map[string]interface{}) {
	if c.Room == nil {
		return
	}

	switch action {
	case "discard":
		// 处理出牌
		tile, ok := data["tile"].(string)
		if !ok {
			return
		}
		c.Room.HandleDiscard(c.UserID, tile)
		// 广播玩家出牌动作到所有客户端
		c.Hub.BroadcastPlayerAction(c.Room, c.UserID, "discard", tile)

	case "pong":
		// 处理碰牌
		tile, ok := data["tile"].(string)
		if !ok {
			return
		}
		c.Room.HandlePong(c.UserID, tile)
		// 广播玩家碰牌动作
		c.Hub.BroadcastPlayerAction(c.Room, c.UserID, "pong", tile)

	case "kong":
		// 处理杠牌
		tile, ok := data["tile"].(string)
		if !ok {
			return
		}
		c.Room.HandleKong(c.UserID, tile)
		// 广播玩家杠牌动作
		c.Hub.BroadcastPlayerAction(c.Room, c.UserID, "kong", tile)

	case "hu":
		// 处理胡牌
		c.Room.HandleHu(c.UserID)
		// 广播玩家胡牌动作
		c.Hub.BroadcastPlayerAction(c.Room, c.UserID, "hu", "")

	case "add_bot":
		// 添加Bot玩家
		c.Hub.addBot(c.Room)

	default:
		log.Printf("未知游戏动作: %s", action)
	}
}

// ServeWs 处理WebSocket请求
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	// 从查询参数获取房间ID和用户信息
	roomID := r.URL.Query().Get("room")
	userID := r.URL.Query().Get("userId")
	userName := r.URL.Query().Get("userName")

	if roomID == "" || userID == "" {
		conn.Close()
		return
	}

	client := &Client{
		Hub:      hub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		RoomID:   roomID,
		UserID:   userID,
		UserName: userName,
	}

	hub.register <- client

	// 启动读写协程
	go client.writePump()
	go client.readPump()
}
