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
	log.Printf("启动读协程 for %s", c.UserName)
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
		log.Printf("等待来自 %s 的消息...", c.UserName)
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
		c.Hub.BroadcastPlayerAction(c.Room, c.UserID, "discard", tile)

		// 获取出牌者的位置
		var discarderPosition int
		for _, p := range c.Room.Players {
			if p.ID == c.UserID {
				discarderPosition = p.Position
				break
			}
		}

				// 检查是否有Bot响应

				actionTaken := c.Hub.botsReactToDiscard(c.Room, tile, discarderPosition)

				if !actionTaken {

					// 检查是否有人类玩家可以响应

					hasHumanAction := c.Hub.HasHumanAction(c.Room, tile, discarderPosition)

					// 如果没有Bot响应，则轮到下一个玩家

					c.Hub.CheckAndPlayBotTurn(c.Room, hasHumanAction)

				}

		

		

			case "chow":

				// 处理吃牌

				tile, ok := data["tile"].(string)

				if !ok {

					return

				}

		

				// 获取吃牌组合

				var chowTiles []string

				if chowTilesInterface, ok := data["chowTiles"].([]interface{}); ok {

					for _, t := range chowTilesInterface {

						if tileStr, ok := t.(string); ok {

							chowTiles = append(chowTiles, tileStr)

						}

					}

				}

		

				if len(chowTiles) != 3 {

					log.Printf("吃牌组合无效，长度: %d", len(chowTiles))

					return

				}

		

				success := c.Room.HandleChow(c.UserID, tile, chowTiles)

				if success {

					// 广播玩家吃牌动作（需要包含完整的组合信息）

					c.Hub.BroadcastChowAction(c.Room, c.UserID, tile, chowTiles)

				}

		

			case "pong":

				// 处理碰牌

				tile, ok := data["tile"].(string)

				if !ok {

					return

				}

				// 只有成功碰牌才廣播

				success := c.Room.HandlePong(c.UserID, tile)

				if success {

					// 广播玩家碰牌动作

					c.Hub.BroadcastPlayerAction(c.Room, c.UserID, "pong", tile)

				}

		

			case "kong":

				// 处理杠牌

				tile, ok := data["tile"].(string)

				if !ok {

					return

				}

				isConcealed := false

				if concealed, ok := data["concealed"].(bool); ok {

					isConcealed = concealed

				}

				success := c.Room.HandleKong(c.UserID, tile, isConcealed)

				if success {

					// 杠牌成功后，找到杠牌的玩家和牌组

					var player *game.Player

					for _, p := range c.Room.Players {

						if p.ID == c.UserID {

							player = p

							break

						}

					}

		

					if player != nil {

						var kongMeld game.Meld

						// 找到对应的杠牌组

						for _, meld := range player.Melds {

							isKong := (meld.Type == "kong_exposed" || meld.Type == "kong_concealed" || meld.Type == "kong_promoted")

							// 假设杠的牌是唯一的，或者最近的一个

							if isKong && meld.Tiles[0] == tile {

								kongMeld = meld

								break // 找到就跳出

							}

						}

		

						if kongMeld.Type != "" {

							c.Hub.BroadcastKongAction(c.Room, c.UserID, kongMeld)

						} else {

							// Fallback

							log.Printf("找不到杠牌组，使用旧版广播: %s", tile)

							c.Hub.BroadcastPlayerAction(c.Room, c.UserID, "kong", tile)

						}

					}

				}

		

			case "check_ting":

				// 检查听牌

				if c.Room.Game == nil {

					return

				}

				var player *game.Player

				for _, p := range c.Room.Players {

					if p.ID == c.UserID {

						player = p

						break

					}

				}

				if player == nil {

					return

				}

				

				// 遍历玩家手牌，尝试打出每一张并检查是否听牌

				var possibleTingDiscards = make(map[string][]string)

				for _, discardCandidate := range player.Hand {

					

					// 创建一个不包含候选弃牌的临时手牌

					tempHand := []string{}

					removed := false

					for _, tile := range player.Hand {

						if tile == discardCandidate && !removed {

							removed = true

							continue

						}

						tempHand = append(tempHand, tile)

					}

		

					tingResult := c.Room.Game.CheckTing(tempHand, player.Melds)

					if tingResult.IsTing {

						possibleTingDiscards[discardCandidate] = tingResult.WinningTiles

					}

				}

		

				// 发送听牌结果给客户端

				tingResultMsg := map[string]interface{}{

					"type": "ting_result",

					"data": possibleTingDiscards,

				}

				tingResultBytes, _ := json.Marshal(tingResultMsg)

				c.Send <- tingResultBytes

		

		

			case "ting":

				// 宣布听牌

				tile, ok := data["tile"].(string)

				if !ok {

					return

				}

		

				var player *game.Player

				for _, p := range c.Room.Players {

					if p.ID == c.UserID {

						player = p

						break

					}

				}

				if player == nil {

					return

				}

				

				// 验证这是否是一个有效的听牌弃牌 (与上面类似)

				tempHand := []string{}

				removed := false

				for _, t := range player.Hand {

					if t == tile && !removed {

						removed = true

						continue

					}

					tempHand = append(tempHand, t)

				}

				tingResult := c.Room.Game.CheckTing(tempHand, player.Melds)

		

				if tingResult.IsTing {

					log.Printf("玩家 %s 宣布听牌，打出 %s", c.UserName, tile)

					player.IsTing = true

					player.WinningTiles = tingResult.WinningTiles

		

					// 处理出牌

					c.Room.HandleDiscard(c.UserID, tile)

					// 广播听牌动作

					c.Hub.BroadcastPlayerTingAction(c.Room, c.UserID, tile)

		

					// 检查是否有Bot响应

					actionTaken := c.Hub.botsReactToDiscard(c.Room, tile, player.Position)

					if !actionTaken {

						hasHumanAction := c.Hub.HasHumanAction(c.Room, tile, player.Position)

						c.Hub.CheckAndPlayBotTurn(c.Room, hasHumanAction)

					}

				} else {

					log.Printf("玩家 %s 尝试听牌失败，打出 %s", c.UserName, tile)

					// 可选：发送一个错误消息给客户端

				}

		

			case "hu":

				// 处理胡牌

				isSelfDrawn := false

				if selfDrawn, ok := data["selfDrawn"].(bool); ok {
			isSelfDrawn = selfDrawn
		}
		winResult := c.Room.HandleHu(c.UserID, isSelfDrawn)
		if winResult != nil {
			// 广播玩家胡牌动作
			c.Hub.BroadcastPlayerAction(c.Room, c.UserID, "hu", "")
		}

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
