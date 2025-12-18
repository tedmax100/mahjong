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
		return true // 允許所有來源（生產環境應該限制）
	},
}

// Client 代表一個 WebSocket 客戶端
type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan []byte
	RoomID   string
	UserID   string
	UserName string
	Room     *game.Room
}

// Message 表示客戶端訊息
type Message struct {
	Type   string                 `json:"type"`
	Action string                 `json:"action,omitempty"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

// readPump 從 WebSocket 連接讀取訊息
func (c *Client) readPump() {
	log.Printf("啟動讀協程 for %s", c.UserName)
	defer func() {
		c.Hub.unregister <- c
		_ = c.Conn.Close()
	}()

	_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		return c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		log.Printf("等待來自 %s 的訊息...", c.UserName)
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket 錯誤: %v", err)
			}
			break
		}

		// 處理訊息
		c.handleMessage(message)
	}
}

// writePump 向 WebSocket 連接寫入訊息
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage 處理收到的訊息
func (c *Client) handleMessage(data []byte) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("解析訊息失敗: %v", err)
		return
	}

	log.Printf("收到訊息: %s from %s", msg.Type, c.UserName)

	switch msg.Type {
	case "join":
		// 加入房間訊息在連接時已處理
		log.Printf("玩家 %s 加入", c.UserName)

	case "action":
		// 處理遊戲動作
		c.handleGameAction(msg.Action, msg.Data)

	case "webrtc_signal":
		// 處理 WebRTC 信令轉發
		c.handleWebRTCSignal(msg.Data)

	// [TEMP] IP 偵測功能 - 之後要移除時，刪除這個 case
	case "client_ip":
		if ip, ok := msg.Data["ip"].(string); ok {
			log.Printf("[TEMP] 玩家 %s 的 IP: %s", c.UserName, ip)
		}
	// [TEMP] END

	default:
		log.Printf("未知訊息類型: %s", msg.Type)
	}
}

// handleWebRTCSignal 處理 WebRTC 信令轉發
func (c *Client) handleWebRTCSignal(data map[string]interface{}) {
	if c.Room == nil {
		log.Printf("WebRTC 信令失敗：玩家 %s 不在房間中", c.UserName)
		return
	}

	targetId, ok := data["targetId"].(string)
	if !ok || targetId == "" {
		log.Printf("WebRTC 信令缺少 targetId")
		return
	}

	signalType, ok := data["signalType"].(string)
	if !ok || signalType == "" {
		log.Printf("WebRTC 信令缺少 signalType")
		return
	}

	payload := data["payload"]

	log.Printf("WebRTC 信令轉發: %s -> %s (type: %s)", c.UserID, targetId, signalType)

	// 轉發信令給目標玩家
	c.Hub.SendToPlayer(c.Room, targetId, map[string]interface{}{
		"type": "webrtc_signal",
		"data": map[string]interface{}{
			"fromId":     c.UserID,
			"signalType": signalType,
			"payload":    payload,
		},
	})
}

// handleGameAction 處理遊戲動作
func (c *Client) handleGameAction(action string, data map[string]interface{}) {
	if c.Room == nil {
		return
	}

	switch action {
	case "discard":
		// 處理出牌
		tile, ok := data["tile"].(string)
		if !ok {
			return
		}
		success, isDraw := c.Room.HandleDiscard(c.UserID, tile)

		// 如果打牌失敗，不廣播也不繼續處理
		if !success {
			return
		}

		c.Hub.BroadcastPlayerAction(c.Room, c.UserID, "discard", tile)

		// 檢查是否流局
		if isDraw {
			c.Hub.BroadcastGameDraw(c.Room)
			return
		}

		// 獲取出牌者的位置
		var discarderPosition int
		for _, p := range c.Room.Players {
			if p.ID == c.UserID {
				discarderPosition = p.Position
				break
			}
		}

		// 檢查是否有 Bot 回應

		actionTaken := c.Hub.botsReactToDiscard(c.Room, tile, discarderPosition)

		if !actionTaken {

			// 檢查是否有人類玩家可以回應

			hasHumanAction := c.Hub.HasHumanAction(c.Room, tile, discarderPosition)

			// 如果沒有 Bot 回應，則輪到下一個玩家

			c.Hub.CheckAndPlayBotTurn(c.Room, hasHumanAction)

		}

	case "chow":

		c.Room.StopNoResponseTimer()

		// 處理吃牌

		tile, ok := data["tile"].(string)

		if !ok {

			return

		}

		// 獲取吃牌組合

		var chowTiles []string

		if chowTilesInterface, ok := data["chowTiles"].([]interface{}); ok {

			for _, t := range chowTilesInterface {

				if tileStr, ok := t.(string); ok {

					chowTiles = append(chowTiles, tileStr)

				}

			}

		}

		if len(chowTiles) != 3 {

			log.Printf("吃牌組合無效，長度: %d", len(chowTiles))

			return

		}

		success := c.Room.HandleChow(c.UserID, tile, chowTiles)

		if success {

			// 廣播玩家吃牌動作（需要包含完整的組合資訊）

			c.Hub.BroadcastChowAction(c.Room, c.UserID, tile, chowTiles)

		}

	case "pong":

		c.Room.StopNoResponseTimer()

		// 處理碰牌

		tile, ok := data["tile"].(string)

		if !ok {

			return

		}

		// 只有成功碰牌才廣播

		success := c.Room.HandlePong(c.UserID, tile)

		if success {

			// 廣播玩家碰牌動作

			c.Hub.BroadcastPlayerAction(c.Room, c.UserID, "pong", tile)

		}

	case "kong":

		c.Room.StopNoResponseTimer()

		// 處理槓牌

		tile, ok := data["tile"].(string)

		if !ok {

			return

		}

		isConcealed := false

		if concealed, ok := data["concealed"].(bool); ok {

			isConcealed = concealed

		}

		success, drawnTile := c.Room.HandleKong(c.UserID, tile, isConcealed)

		if success {

			// 槓牌成功後，找到槓牌的玩家和牌組

			var player *game.Player

			for _, p := range c.Room.Players {

				if p.ID == c.UserID {

					player = p

					break

				}

			}

			if player != nil {

				// 使用 model.Meld 替代 game.Meld
				var kongMeld interface{}

				// 找到對應的槓牌組
				for _, meld := range player.Melds {

					isKong := (meld.Type == "kong_exposed" || meld.Type == "kong_concealed" || meld.Type == "kong_promoted")

					// 假設槓的牌是唯一的，或者最近的一個

					if isKong && meld.Tiles[0] == tile {

						kongMeld = meld

						break // 找到就跳出

					}

				}

				// 這裡需要特別處理，因為 BroadcastKongAction 需要 model.Meld
				// 但我們現在沒有引入 model 包，而 game.Meld 已經是 model.Meld
				// 所以我們可以透過 player.Melds[i] 直接傳遞
				// 但是為了保持代碼簡潔，我們在 hub.go 已經修改了 BroadcastKongAction 的簽名
				// 這裡需要確保傳遞正確的類型
				
				// 由於我們還沒修改 client.go 引入 model 包
				// 但 player.Melds 本身就是 []model.Meld (透過 game.Player 定義)
				// 所以 kongMeld 應該是 model.Meld 類型
				// 不過這檔案還沒引入 mahjong/internal/model
				// 所以我等下要加上 import

				if kongMeld != nil {
					// 因為 kongMeld 是 interface{}，我們需要斷言或者更改上面的變量類型
					// 但這裡最簡單的是直接傳遞，因為 Hub 方法接受 game.Meld (即 model.Meld)
					// 等等，hub.go 已經更新為接受 game.Meld (它是 model.Meld 的別名嗎？不是，它是 model.Meld)
					// 在 room.go 中: Melds []model.Meld
					// 所以 player.Melds 是 []model.Meld
					// 所以 meld 是 model.Meld
					
					// 為了避免類型問題，我應該讓 hub.go 的 BroadcastKongAction 接受 model.Meld
					// 而這已經在 hub.go 中完成了 (import model, meld model.Meld)
					
					// 所以這裡只需要從 player.Melds 中取出即可
					
					// 重新寫這一段邏輯:
					found := false
					for _, meld := range player.Melds {
						isKong := (meld.Type == "kong_exposed" || meld.Type == "kong_concealed" || meld.Type == "kong_promoted")
						if isKong && meld.Tiles[0] == tile {
							c.Hub.BroadcastKongAction(c.Room, c.UserID, meld)
							found = true
							break
						}
					}

					if !found {
						log.Printf("找不到槓牌組，使用舊版廣播: %s", tile)
						c.Hub.BroadcastPlayerAction(c.Room, c.UserID, "kong", tile)
					}
				} else {
					log.Printf("找不到槓牌組，使用舊版廣播: %s", tile)
					c.Hub.BroadcastPlayerAction(c.Room, c.UserID, "kong", tile)
				}

				// 廣播補牌

				if drawnTile != "" {

					c.Hub.BroadcastDrawTile(c.Room, c.UserID, drawnTile)

				}

			}

		}

	case "pass":

		c.Room.StopNoResponseTimer()

		// 處理"過"動作（玩家選擇不執行吃/碰/槓/胡）

		tile, _ := data["tile"].(string)

		log.Printf("玩家 %s 選擇過，不執行動作 (tile: %s)", c.UserName, tile)

		// 不需要做任何特殊處理，遊戲將繼續正常流程

		// 輪到下一個玩家

		c.Hub.CheckAndPlayBotTurn(c.Room, false)

	case "check_ting":

		// 檢查聽牌

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

		// 遍歷玩家手牌，嘗試打出每一張並檢查是否聽牌

		var possibleTingDiscards = make(map[string][]string)

		for _, discardCandidate := range player.Hand {

			// 建立一個不包含候選棄牌的臨時手牌

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

		// 發送聽牌結果給客戶端

		tingResultMsg := map[string]interface{}{

			"type": "ting_result",

			"data": possibleTingDiscards,
		}

		tingResultBytes, _ := json.Marshal(tingResultMsg)

		c.Send <- tingResultBytes

	case "ting":

		// 宣布聽牌

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

		// 驗證這是否是一個有效的聽牌棄牌 (與上面類似)

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

			log.Printf("玩家 %s 宣布聽牌，打出 %s", c.UserName, tile)

			player.IsTing = true

			player.WinningTiles = tingResult.WinningTiles

			// 處理出牌

			success, isDraw := c.Room.HandleDiscard(c.UserID, tile)

			// 如果打牌失敗，不廣播也不繼續處理
			if !success {
				return
			}

			// 先廣播出牌動作（讓客戶端移除手牌）
			c.Hub.BroadcastPlayerAction(c.Room, c.UserID, "discard", tile)

			// 再廣播聽牌動作（顯示聽牌狀態）

			c.Hub.BroadcastPlayerTingAction(c.Room, c.UserID, tile)

			// 檢查是否流局
			if isDraw {
				c.Hub.BroadcastGameDraw(c.Room)
				return
			}

			// 檢查是否有 Bot 回應

			actionTaken := c.Hub.botsReactToDiscard(c.Room, tile, player.Position)

			if !actionTaken {

				hasHumanAction := c.Hub.HasHumanAction(c.Room, tile, player.Position)

				c.Hub.CheckAndPlayBotTurn(c.Room, hasHumanAction)

			}

		} else {

			log.Printf("玩家 %s 嘗試聽牌失敗，打出 %s", c.UserName, tile)

			// 可選：發送一個錯誤訊息給客戶端

		}

	case "declare_ting":
		// 宣告聽牌（不打牌，只標記狀態）
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

		// 檢查當前手牌是否聽牌
		tingResult := c.Room.Game.CheckTing(player.Hand, player.Melds)
		if tingResult.IsTing {
			log.Printf("玩家 %s 宣告聽牌", c.UserName)
			player.IsTing = true
			player.WinningTiles = tingResult.WinningTiles

			// 廣播聽牌狀態給所有玩家
			message := map[string]interface{}{
				"type": "player_action",
				"data": map[string]interface{}{
					"playerId":     c.UserID,
					"action":       "ting",
					"winningTiles": player.WinningTiles,
				},
			}
			msgBytes, _ := json.Marshal(message)
			for _, clientInterface := range c.Room.Clients {
				client, ok := clientInterface.(*Client)
				if !ok {
					continue
				}
				select {
				case client.Send <- msgBytes:
				default:
					log.Printf("發送聽牌訊息失敗")
				}
			}
		} else {
			log.Printf("玩家 %s 嘗試宣告聽牌失敗，當前手牌不是聽牌", c.UserName)
		}

	case "hu":

		// 處理胡牌

		isSelfDrawn := false

		if selfDrawn, ok := data["selfDrawn"].(bool); ok {
			isSelfDrawn = selfDrawn
		}
		tile, _ := data["tile"].(string)
		log.Printf("🎯 [DEBUG] 處理胡牌請求: userID=%s, tile=%s, isSelfDrawn=%v", c.UserID, tile, isSelfDrawn)
		winResult := c.Room.HandleHu(c.UserID, tile, isSelfDrawn)
		log.Printf("🎯 [DEBUG] HandleHu 返回結果: winResult != nil = %v", winResult != nil)
		if winResult != nil {
			log.Printf("🎯 [DEBUG] winResult 詳情: TotalTai=%d, BaseScore=%d", winResult.TotalTai, winResult.BaseScore)
			// 廣播玩家胡牌動作
			log.Printf("🎯 [DEBUG] 準備廣播玩家胡牌動作")
			c.Hub.BroadcastPlayerAction(c.Room, c.UserID, "hu", "")
			log.Printf("🎯 [DEBUG] 玩家胡牌動作已廣播")
			// 廣播遊戲勝利
			log.Printf("🎯 [DEBUG] 準備廣播遊戲勝利")
			c.Hub.BroadcastGameWin(c.Room, c.UserID, winResult)
			log.Printf("🎯 [DEBUG] 遊戲勝利已廣播")
		} else {
			log.Printf("❌ [DEBUG] winResult 為 nil，不廣播勝利訊息")
		}

	case "add_bot":
		// 添加 Bot 玩家
		c.Hub.addBot(c.Room)

	default:
		log.Printf("未知遊戲動作: %s", action)
	}
}

// ServeWs 處理 WebSocket 請求
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket 升級失敗: %v", err)
		return
	}

	// 從查詢參數獲取房間 ID 和使用者資訊
	roomID := r.URL.Query().Get("room")
	userID := r.URL.Query().Get("userId")
	userName := r.URL.Query().Get("userName")

	if roomID == "" || userID == "" {
		_ = conn.Close()
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

	// 啟動讀寫協程
	go client.writePump()
	go client.readPump()
}
