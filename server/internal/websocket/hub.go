package websocket

import (
	"encoding/json"
	"log"
	"mahjong/internal/game"
	"math/rand"
	"sync"
	"time"
)

// Hub 管理所有WebSocket连接和房间
type Hub struct {
	// 所有房间
	rooms map[string]*game.Room

	// 注册请求
	register chan *Client

	// 注销请求
	unregister chan *Client

	// 互斥锁
	mu sync.RWMutex
}

// NewHub 创建新的Hub
func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]*game.Room),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run 运行Hub主循环
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)
		}
	}
}

// registerClient 注册新客户端
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	roomID := client.RoomID

	// 获取或创建房间
	room, exists := h.rooms[roomID]
	if !exists {
		room = game.NewRoom(roomID)
		h.rooms[roomID] = room
		log.Printf("创建新房间: %s", roomID)
	}

	// 将客户端加入房间
	if err := room.AddPlayer(client.UserID, client.UserName); err != nil {
		log.Printf("玩家加入房间失败: %v", err)
		client.Send <- []byte(`{"type":"error","message":"房间已满"}`)
		return
	}

	client.Room = room
	room.Clients[client.UserID] = client

	log.Printf("玩家 %s 加入房间 %s", client.UserName, roomID)

	// 广播房间更新
	log.Printf("广播房间更新: 房间 %s 现有 %d 名玩家", roomID, len(room.Players))
	h.broadcastRoomUpdate(room)

	// 如果房间满4人，开始游戏
	if len(room.Players) == 4 && !room.GameStarted {
		log.Printf("房间 %s 已满4人，准备开始游戏", roomID)
		h.startGame(room)
	}
}

// unregisterClient 注销客户端
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if client.Room != nil {
		room := client.Room

		// 从房间移除客户端
		delete(room.Clients, client.UserID)
		room.RemovePlayer(client.UserID)

		log.Printf("玩家 %s 离开房间 %s", client.UserName, room.ID)

		// 如果房间为空，删除房间
		if len(room.Clients) == 0 {
			delete(h.rooms, room.ID)
			log.Printf("删除空房间: %s", room.ID)
		} else {
			// 广播房间更新
			h.broadcastRoomUpdate(room)
		}
	}

	close(client.Send)
}

// broadcastRoomUpdate 广播房间状态更新
func (h *Hub) broadcastRoomUpdate(room *game.Room) {
	message := room.GetRoomUpdateMessage()

	for userID, clientInterface := range room.Clients {
		client, ok := clientInterface.(*Client)
		if !ok {
			continue
		}
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(room.Clients, userID)
		}
	}
}

// startGame 开始游戏
func (h *Hub) startGame(room *game.Room) {
	log.Printf("房间 %s 开始游戏", room.ID)

	room.GameStarted = true
	room.StartGame()

	// 发送游戏开始消息
	startMessage := room.GetGameStartMessage()
	for _, clientInterface := range room.Clients {
		client, ok := clientInterface.(*Client)
		if !ok {
			continue
		}
		client.Send <- startMessage
	}

	// 发牌
	h.dealTiles(room)
}

// dealTiles 发牌
func (h *Hub) dealTiles(room *game.Room) {
	room.DealTiles()
	log.Printf("房间 %s 发牌完成", room.ID)

	// 向每个玩家发送他们的手牌
	for i, player := range room.Players {
		log.Printf("玩家 %s 收到 %d 张牌", player.Name, len(player.Hand))

		// 只向有WebSocket连接的真实玩家发送
		if clientInterface, ok := room.Clients[player.ID]; ok {
			client, ok := clientInterface.(*Client)
			if !ok {
				continue
			}
			dealMessage := room.GetDealTilesMessage(i)
			client.Send <- dealMessage
		}
	}

	// Bot玩家自动出牌
	go h.botAutoPlay(room)
}

// GetRoom 获取房间
func (h *Hub) GetRoom(roomID string) *game.Room {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.rooms[roomID]
}

// CreateRoom 创建房间
func (h *Hub) CreateRoom(roomID string) *game.Room {
	h.mu.Lock()
	defer h.mu.Unlock()

	room := game.NewRoom(roomID)
	h.rooms[roomID] = room

	log.Printf("创建房间: %s", roomID)
	return room
}

// addBot 添加Bot玩家
func (h *Hub) addBot(room *game.Room) {
	if room == nil {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// 检查房间是否已满
	if len(room.Players) >= 4 {
		log.Printf("房间 %s 已满，无法添加Bot", room.ID)
		return
	}

	// 生成Bot信息
	botNames := []string{"机器人·小智", "机器人·阿尔法", "机器人·贝塔", "机器人·伽马"}
	botName := botNames[len(room.Players)%len(botNames)]
	botID := "bot_" + room.ID + "_" + string(rune('A'+len(room.Players)))

	// 添加Bot玩家
	if err := room.AddPlayer(botID, botName); err != nil {
		log.Printf("添加Bot失败: %v", err)
		return
	}

	log.Printf("Bot %s 加入房间 %s", botName, room.ID)

	// 广播房间更新
	h.broadcastRoomUpdate(room)

	// 如果房间满4人，开始游戏
	if len(room.Players) == 4 && !room.GameStarted {
		log.Printf("房间 %s 已满4人（包含Bot），准备开始游戏", room.ID)
		h.startGame(room)
	}
}

// botAutoPlay Bot自动出牌（简单AI）
func (h *Hub) botAutoPlay(room *game.Room) {
	log.Printf("房间 %s 的Bot玩家AI已激活", room.ID)

	// 简单AI：检查轮到Bot时自动出牌
	go func() {
		time.Sleep(2 * time.Second) // 等待游戏稳定

		for {
			if room.Game == nil || !room.GameStarted {
				break
			}

			// 检查当前轮到谁
			currentPlayer := room.Players[room.CurrentTurn]

			// 只有轮到Bot时才出牌
			if len(currentPlayer.ID) > 4 && currentPlayer.ID[:4] == "bot_" {
				// Bot有手牌就自动打出第一张
				if len(currentPlayer.Hand) > 0 {
					tile := currentPlayer.Hand[0]

					log.Printf("Bot %s 的回合，准备打出 %s", currentPlayer.Name, tile)

					// 等待1-2秒（模拟思考）
					time.Sleep(time.Duration(1000+rand.Intn(1000)) * time.Millisecond)

					// 调用出牌逻辑（会自动切换回合）
					room.HandleDiscard(currentPlayer.ID, tile)

					// 广播Bot出牌消息给所有真实玩家
					h.broadcastBotAction(room, currentPlayer.ID, "discard", tile)
				}
			}

			// 每500ms检查一次是否轮到Bot
			time.Sleep(500 * time.Millisecond)
		}
	}()
}

// BroadcastPlayerAction 广播玩家动作（包括Bot和真实玩家）
func (h *Hub) BroadcastPlayerAction(room *game.Room, playerID, action, tile string) {
	message := map[string]interface{}{
		"type": "player_action",
		"data": map[string]interface{}{
			"playerId":    playerID,
			"action":      action,
			"tile":        tile,
			"currentTurn": room.CurrentTurn, // 添加当前轮次信息
		},
	}

	msgBytes, _ := json.Marshal(message)

	log.Printf("广播玩家动作: %s %s %s (下一轮: %d)", playerID, action, tile, room.CurrentTurn)

	// 向所有玩家发送
	for _, clientInterface := range room.Clients {
		client, ok := clientInterface.(*Client)
		if !ok {
			continue
		}
		select {
		case client.Send <- msgBytes:
		default:
			log.Printf("发送玩家动作消息失败")
		}
	}
}

// broadcastBotAction 广播Bot动作（保留向后兼容）
func (h *Hub) broadcastBotAction(room *game.Room, botID, action, tile string) {
	h.BroadcastPlayerAction(room, botID, action, tile)
}
