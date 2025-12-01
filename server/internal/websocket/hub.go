package websocket

import (
	"encoding/json"
	"log"
	"mahjong/internal/game"
	"math/rand"
	"strings"
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
	var roomToUpdate *game.Room
	var shouldStartGame bool

	h.mu.Lock()

	room, exists := h.rooms[client.RoomID]
	if !exists {
		room = game.NewRoom(client.RoomID)
		h.rooms[client.RoomID] = room
		log.Printf("创建新房间: %s", client.RoomID)
	}

	if err := room.AddPlayer(client.UserID, client.UserName); err != nil {
		log.Printf("玩家加入房间失败: %v", err)
		h.mu.Unlock() // Unlock before sending to channel
		client.Send <- []byte(`{"type":"error","message":"房间已满"}`)
		return
	}

	client.Room = room
	room.Clients[client.UserID] = client
	log.Printf("玩家 %s 加入房间 %s", client.UserName, client.RoomID)

	roomToUpdate = room
	if len(room.Players) == 4 && !room.GameStarted {
		shouldStartGame = true
		room.GameStarted = true // Mark as started under lock
	}

	h.mu.Unlock() // --- Lock Released ---

	if roomToUpdate != nil {
		h.broadcastRoomUpdate(roomToUpdate)
	}

	if shouldStartGame {
		log.Printf("房间 %s 已满4人，准备开始游戏", roomToUpdate.ID)
		h.startGame(roomToUpdate)
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
	h.broadcast(room, message)
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

	// 检查第一个回合是否是Bot
	h.CheckAndPlayBotTurn(room, false) // 游戏开始时庄家直接打牌，不应该有延迟
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

	var shouldStartGame bool

	h.mu.Lock()

	if len(room.Players) >= 4 {
		log.Printf("房间 %s 已满，无法添加Bot", room.ID)
		h.mu.Unlock()
		return
	}

	botNames := []string{"机器人·小智", "机器人·阿尔法", "机器人·贝塔", "机器人·伽马"}
	botName := botNames[len(room.Players)%len(botNames)]
	botID := "bot_" + room.ID + "_" + string(rune('A'+len(room.Players)))

	if err := room.AddPlayer(botID, botName); err != nil {
		log.Printf("添加Bot失败: %v", err)
		h.mu.Unlock()
		return
	}
	log.Printf("Bot %s 加入房间 %s", botName, room.ID)

	if len(room.Players) == 4 && !room.GameStarted {
		shouldStartGame = true
		room.GameStarted = true // Mark as started under lock
	}
	
	h.mu.Unlock() // --- Lock Released ---

	h.broadcastRoomUpdate(room)

	if shouldStartGame {
		log.Printf("房间 %s 已满4人（包含Bot），准备开始游戏", room.ID)
		h.startGame(room)
	}
}

// CheckAndPlayBotTurn 检查并执行Bot回合
func (h *Hub) CheckAndPlayBotTurn(room *game.Room, withDelay bool) {
	if room == nil || room.Game == nil || !room.GameStarted {
		return
	}

	h.mu.RLock()
	if room.CurrentTurn < 0 || room.CurrentTurn >= len(room.Players) {
		h.mu.RUnlock()
		return
	}
	currentPlayer := room.Players[room.CurrentTurn]
	currentTurnAtCheck := room.CurrentTurn
	h.mu.RUnlock()

	// 如果是真实玩家，根据情况决定是否延迟发牌
	if !strings.HasPrefix(currentPlayer.ID, "bot_") {
		if withDelay {
			go func() {
				log.Printf("检测到玩家 %s (位置 %d) 有可执行动作，开始 10 秒等待...", currentPlayer.Name, currentTurnAtCheck)
				time.Sleep(10 * time.Second)
				h.mu.Lock()
				defer h.mu.Unlock()
				log.Printf("玩家 %s 的等待时间结束。当前轮次: %d, 期望轮次: %d", currentPlayer.Name, room.CurrentTurn, currentTurnAtCheck)
				if room.GameStarted && room.CurrentTurn == currentTurnAtCheck {
					log.Printf("轮次未变，为玩家 %s 执行摸牌", currentPlayer.Name)
					h.drawForRealPlayer_needsLock(room)
				} else {
					log.Printf("轮次已从 %d 变为 %d，取消为玩家 %s 的自动摸牌", currentTurnAtCheck, room.CurrentTurn, currentPlayer.Name)
				}
			}()
		} else {
			log.Printf("没有检测到可执行动作，立即为玩家 %s 执行摸牌", currentPlayer.Name)
			h.mu.Lock()
			defer h.mu.Unlock()
			h.drawForRealPlayer_needsLock(room)
		}
		return
	}

	// Bot 的逻辑
	if strings.HasPrefix(currentPlayer.ID, "bot_") {
		go func() {
			log.Printf("Bot %s 的回合已开始，等待出牌...", currentPlayer.Name)
			time.Sleep(time.Duration(1000+rand.Intn(1000)) * time.Millisecond)

			h.mu.Lock()

			if !room.GameStarted || room.CurrentTurn != currentTurnAtCheck {
				log.Printf("Bot %s 的回合被中断", currentPlayer.Name)
				h.mu.Unlock()
				return
			}

			// 計算總牌數（手牌 + 已展示的吃碰槓）
			totalTiles := len(currentPlayer.Hand) + len(currentPlayer.Melds)*3

			// 只有在正常轮次（總牌數 16 張）才摸牌
			// 如果剛吃/碰/槓完（總牌數 17 張），不摸牌直接出牌
			if totalTiles == 16 {
				// 使用 DrawTileWithFlowerReplacement 自動處理花牌
				drawnTile := room.Game.DrawTileWithFlowerReplacement(currentPlayer)
				if drawnTile != "" {
					log.Printf("Bot %s 摸到了 %s（手牌 %d 張 + 吃碰槓 %d 組 + 花牌 %d 張）",
						currentPlayer.Name, drawnTile, len(currentPlayer.Hand), len(currentPlayer.Melds), len(currentPlayer.Flowers))
					currentPlayer.Hand = append(currentPlayer.Hand, drawnTile)

					// 如果摸牌過程中補了花牌，記錄並廣播
					if len(currentPlayer.Flowers) > 0 {
						log.Printf("Bot %s 的花牌: %v", currentPlayer.Name, currentPlayer.Flowers)
						// TODO: 廣播花牌給前端
					}

					// 檢查 Bot 是否已聽牌
					if currentPlayer.IsTing {
						// 檢查是否自摸（摸到的牌是否在聽牌列表中）
						isSelfDrawn := false
						for _, winTile := range currentPlayer.WinningTiles {
							if winTile == drawnTile {
								isSelfDrawn = true
								break
							}
						}

						if isSelfDrawn {
							// 自摸！廣播胡牌
							log.Printf("Bot %s 自摸 %s！", currentPlayer.Name, drawnTile)
							// TODO: 廣播胡牌事件
							// h.mu.Unlock()
							// h.BroadcastGameWin(room, currentPlayer.ID, drawnTile, true)
							// return
						} else {
							// 不是自摸，自動打出摸到的牌，保持聽牌狀態
							log.Printf("Bot %s 聽牌中，自動打出 %s", currentPlayer.Name, drawnTile)

							discarderPosition := currentPlayer.Position
							h.mu.Unlock()

							// 處理出牌（HandleDiscard 會自己移除手牌）
							isDraw := room.HandleDiscard(currentPlayer.ID, drawnTile)
							h.BroadcastPlayerAction(room, currentPlayer.ID, "discard", drawnTile)

							// 檢查是否流局
							if isDraw {
								h.BroadcastGameDraw(room)
								return
							}

							actionTaken := h.botsReactToDiscard(room, drawnTile, discarderPosition)
							if !actionTaken {
								h.BroadcastPossibleActions(room, drawnTile, discarderPosition)
								hasHumanAction := h.HasHumanAction(room, drawnTile, discarderPosition)
								if hasHumanAction {
									log.Printf("等待真實玩家響應可執行動作，10秒後自動繼續...")
									currentTurnSnapshot := room.CurrentTurn
									room.StartNoResponseTimer(10*time.Second, func() {
										h.mu.Lock()
										shouldContinue := room.CurrentTurn == currentTurnSnapshot
										if shouldContinue {
											log.Printf("真實玩家未響應，繼續遊戲")
											room.StopNoResponseTimer()
										}
										h.mu.Unlock()
										if shouldContinue {
											h.CheckAndPlayBotTurn(room, false)
										}
									})
								} else {
									h.CheckAndPlayBotTurn(room, false)
								}
							}
							return
						}
					}
				}
			} else if totalTiles == 17 {
				log.Printf("Bot %s 剛吃/碰/槓完，不摸牌直接出牌（手牌 %d 張 + 吃碰槓 %d 組）",
					currentPlayer.Name, len(currentPlayer.Hand), len(currentPlayer.Melds))
			} else if totalTiles < 16 {
				log.Printf("警告：Bot %s 牌數異常！總牌數 %d（手牌 %d + 吃碰槓 %d 組），預期 16 或 17",
					currentPlayer.Name, totalTiles, len(currentPlayer.Hand), len(currentPlayer.Melds))
			}

			if len(currentPlayer.Hand) > 0 {
				tileToDiscard := room.Game.ChooseDiscardAI(currentPlayer.Hand)
				discarderPosition := currentPlayer.Position
				h.mu.Unlock() // 在调用响应和下一个回合检查前解锁

				log.Printf("Bot %s 自动打出 %s", currentPlayer.Name, tileToDiscard)

				isDraw := room.HandleDiscard(currentPlayer.ID, tileToDiscard)
				h.BroadcastPlayerAction(room, currentPlayer.ID, "discard", tileToDiscard)

				// 检查是否流局
				if isDraw {
					h.BroadcastGameDraw(room)
					return
				}

				actionTaken := h.botsReactToDiscard(room, tileToDiscard, discarderPosition)
				if !actionTaken {
					// 广播可执行动作给真实玩家
					h.BroadcastPossibleActions(room, tileToDiscard, discarderPosition)

					hasHumanAction := h.HasHumanAction(room, tileToDiscard, discarderPosition)

					// 如果有真实玩家可以执行动作，等待10秒后再继续
					if hasHumanAction {
						log.Printf("等待真实玩家响应可执行动作，10秒后自动继续...")
						currentTurnSnapshot := room.CurrentTurn
						room.StartNoResponseTimer(10*time.Second, func() {
							h.mu.Lock()
							// 检查轮次是否改变（玩家是否已经执行了动作）
							shouldContinue := room.CurrentTurn == currentTurnSnapshot
							if shouldContinue {
								log.Printf("真实玩家未响应，继续游戏")
								// 确保在继续游戏前停止定时器
								room.StopNoResponseTimer()
							}
							h.mu.Unlock()
							// 在锁外调用 CheckAndPlayBotTurn，避免死锁
							if shouldContinue {
								h.CheckAndPlayBotTurn(room, false)
							}
						})
					} else {
						h.CheckAndPlayBotTurn(room, false)
					}
				}
			} else {
				h.mu.Unlock()
			}
		}()
	}
}

// drawForRealPlayer_needsLock 为真实玩家自动发牌（需要持有锁）
func (h *Hub) drawForRealPlayer_needsLock(room *game.Room) {
	if room == nil || room.Game == nil || !room.GameStarted {
		return
	}

	if room.CurrentTurn < 0 || room.CurrentTurn >= len(room.Players) {
		return
	}

	currentPlayer := room.Players[room.CurrentTurn]

	// 确认是真实玩家
	if strings.HasPrefix(currentPlayer.ID, "bot_") {
		return
	}

	// 計算總牌數（手牌 + 已展示的吃碰槓）
	totalTiles := len(currentPlayer.Hand) + len(currentPlayer.Melds)*3

	// 判斷是否需要摸牌：
	// - 總牌數 15 或 16：需要摸牌（玩家已打牌，輪到回合需要摸牌）
	// - 總牌數 17：剛吃/碰/槓完，不摸牌直接等待出牌
	if totalTiles == 17 {
		log.Printf("玩家 %s 剛吃/碰/槓完，不摸牌等待出牌（手牌 %d 張 + 吃碰槓 %d 組）",
			currentPlayer.Name, len(currentPlayer.Hand), len(currentPlayer.Melds))
	} else if totalTiles == 16 || totalTiles == 15 {
		// 記錄摸牌前的花牌數量
		flowerCountBefore := len(currentPlayer.Flowers)

		// 使用 DrawTileWithFlowerReplacement 自動處理花牌
		drawnTile := room.Game.DrawTileWithFlowerReplacement(currentPlayer)
		if drawnTile != "" {
			log.Printf("玩家 %s 摸到了 %s（手牌 %d 張 + 吃碰槓 %d 組 + 花牌 %d 張）",
				currentPlayer.Name, drawnTile, len(currentPlayer.Hand), len(currentPlayer.Melds), len(currentPlayer.Flowers))
			currentPlayer.Hand = append(currentPlayer.Hand, drawnTile)

			// 檢查是否有新的花牌
			if len(currentPlayer.Flowers) > flowerCountBefore {
				newFlowers := currentPlayer.Flowers[flowerCountBefore:]
				log.Printf("玩家 %s 摸到花牌: %v，補牌後摸到 %s", currentPlayer.Name, newFlowers, drawnTile)
				// 廣播花牌事件
				h.BroadcastFlowerTiles(room, currentPlayer.ID, newFlowers)
			}

			// 广播摸牌事件
			h.BroadcastDrawTile(room, currentPlayer.ID, drawnTile)

			// 記錄摸牌後的手牌狀態
			game.LogPlayerHand(currentPlayer, "摸牌: "+drawnTile)

			// 檢查玩家是否已聽牌
			if currentPlayer.IsTing {
				// 檢查是否自摸（摸到的牌是否在聽牌列表中）
				isSelfDrawn := false
				for _, winTile := range currentPlayer.WinningTiles {
					if winTile == drawnTile {
						isSelfDrawn = true
						break
					}
				}

				if isSelfDrawn {
					// 自摸！廣播胡牌
					log.Printf("玩家 %s 自摸 %s！", currentPlayer.Name, drawnTile)
					// TODO: 廣播胡牌事件
					// h.BroadcastGameWin(room, currentPlayer.ID, drawnTile, true)
				} else {
					// 不是自摸，自動打出摸到的牌，保持聽牌狀態
					log.Printf("玩家 %s 聽牌中，自動打出 %s", currentPlayer.Name, drawnTile)

					// 處理出牌（HandleDiscard 會自己移除手牌）
					isDraw := room.HandleDiscard(currentPlayer.ID, drawnTile)

					// 廣播出牌動作
					h.BroadcastPlayerAction(room, currentPlayer.ID, "discard", drawnTile)

					// 檢查是否流局
					if isDraw {
						h.BroadcastGameDraw(room)
						return
					}

					// 檢查是否有Bot響應
					actionTaken := h.botsReactToDiscard(room, drawnTile, currentPlayer.Position)
					if !actionTaken {
						// 廣播可執行動作給真實玩家
						h.BroadcastPossibleActions(room, drawnTile, currentPlayer.Position)

						hasHumanAction := h.HasHumanAction(room, drawnTile, currentPlayer.Position)

						// 如果有真實玩家可以執行動作，等待10秒後再繼續
						if hasHumanAction {
							log.Printf("等待真實玩家響應可執行動作，10秒後自動繼續...")
							currentTurnSnapshot := room.CurrentTurn
							room.StartNoResponseTimer(10*time.Second, func() {
								h.mu.Lock()
								shouldContinue := room.CurrentTurn == currentTurnSnapshot
								if shouldContinue {
									log.Printf("真實玩家未響應，繼續遊戲")
									room.StopNoResponseTimer()
								}
								h.mu.Unlock()
								if shouldContinue {
									h.CheckAndPlayBotTurn(room, false)
								}
							})
						} else {
							h.CheckAndPlayBotTurn(room, false)
						}
					}
				}
			}
		}
	} else {
		// 總牌數 < 15 或 > 17 才是真正的異常
		log.Printf("警告：玩家 %s 牌數異常！總牌數 %d（手牌 %d + 吃碰槓 %d 組），預期 15-17",
			currentPlayer.Name, totalTiles, len(currentPlayer.Hand), len(currentPlayer.Melds))
	}
}

// HasHumanAction 检查是否有任何人类玩家可以对弃牌做出反应
func (h *Hub) HasHumanAction(room *game.Room, discardedTile string, discarderPosition int) bool {
	if room == nil || room.Game == nil {
		return false
	}

	for _, p := range room.Players {
		// 只检查人类玩家，且不是出牌者自己
		if strings.HasPrefix(p.ID, "bot_") || p.Position == discarderPosition {
			continue
		}

		// 如果玩家听牌，只检查胡牌
		if p.IsTing {
			tempHand := append([]string{}, p.Hand...)
			tempHand = append(tempHand, discardedTile)
			if room.Game.CanHu(tempHand, p.Melds) {
				return true
			}
			continue // 听牌后不检查其他动作
		}

		// 检查胡牌
		tempHand := append([]string{}, p.Hand...)
		tempHand = append(tempHand, discardedTile)
		if room.Game.CanHu(tempHand, p.Melds) {
			return true
		}

		// 检查碰或杠
		if room.Game.CanPong(p.Hand, discardedTile) {
			return true
		}
		if room.Game.CanExposedKong(p, discardedTile) {
			return true
		}

		// 检查吃（只能是上家）
		isPreviousPlayer := (p.Position + 3) % 4 == discarderPosition
		if isPreviousPlayer {
			if chowCombinations := room.Game.CanChow(p.Hand, discardedTile); len(chowCombinations) > 0 {
				return true
			}
		}
	}

	return false
}

// BroadcastDrawTile 广播摸牌事件
func (h *Hub) BroadcastDrawTile(room *game.Room, playerID, tile string) {
	// 获取剩余牌数
	remainingTiles := 0
	if room.Game != nil {
		remainingTiles = room.Game.GetRemainingTiles()
	}

	message := map[string]interface{}{
		"type": "player_action",
		"data": map[string]interface{}{
			"playerId":       playerID,
			"action":         "draw",
			"tile":           tile,
			"currentTurn":    room.CurrentTurn,
			"remainingTiles": remainingTiles,
		},
	}

	msgBytes, _ := json.Marshal(message)

	log.Printf("广播玩家摸牌: %s 摸 %s (当前轮次: %d, 剩余: %d)", playerID, tile, room.CurrentTurn, remainingTiles)

	// 向所有玩家发送
	for _, clientInterface := range room.Clients {
		client, ok := clientInterface.(*Client)
		if !ok {
			continue
		}
		select {
		case client.Send <- msgBytes:
		default:
			log.Printf("警告：客户端 %s 的发送缓冲区已满，摸牌消息被丢弃", client.UserName)
		}
	}
}

// BroadcastGameDraw 广播流局事件
func (h *Hub) BroadcastGameDraw(room *game.Room) {
	// 获取剩余牌数
	remainingTiles := 0
	if room.Game != nil {
		remainingTiles = room.Game.GetRemainingTiles()
	}

	message := map[string]interface{}{
		"type": "game_draw",
		"data": map[string]interface{}{
			"remainingTiles": remainingTiles,
			"countdown":      5, // 5秒倒计时
		},
	}

	msgBytes, _ := json.Marshal(message)

	log.Printf("广播流局事件，剩余牌数: %d", remainingTiles)

	// 向所有玩家发送
	for _, clientInterface := range room.Clients {
		client, ok := clientInterface.(*Client)
		if !ok {
			continue
		}
		select {
		case client.Send <- msgBytes:
		default:
			log.Printf("警告：客户端 %s 的发送缓冲区已满，流局消息被丢弃", client.UserName)
		}
	}
}

// botsReactToDiscard 让所有机器人对一个弃牌做出反应
func (h *Hub) botsReactToDiscard(room *game.Room, discardedTile string, discarderPosition int) bool {
	if room == nil || room.Game == nil || !room.GameStarted {
		return false
	}

	var bestAction = "none"
	var bestBot *game.Player
	var bestChowCombination []string

	h.mu.RLock()
	playersCopy := make([]*game.Player, len(room.Players))
	copy(playersCopy, room.Players)
	h.mu.RUnlock()

	for _, p := range playersCopy {
		if p.Position == discarderPosition || !strings.HasPrefix(p.ID, "bot_") {
			continue
		}

		// 如果 Bot 已经听牌，则只检查是否可以胡牌
		if p.IsTing {
			tempHand := append([]string{}, p.Hand...)
			tempHand = append(tempHand, discardedTile)
			if room.Game.CanHu(tempHand, p.Melds) {
				bestAction = "hu"
				bestBot = p
			}
			continue // 跳过其他动作检查
		}

		// 检查胡牌（优先级最高）
		tempHand := append([]string{}, p.Hand...)
		tempHand = append(tempHand, discardedTile)
		if room.Game.CanHu(tempHand, p.Melds) {
			bestAction = "hu"
			bestBot = p
			continue // 胡牌优先级最高，直接跳出继续
		}

		// Simplified action checking...
		if room.Game.CanExposedKong(p, discardedTile) {
			if bestAction != "hu" {
				bestAction = "kong"
				bestBot = p
			}
		}
		if room.Game.CanPong(p.Hand, discardedTile) {
			if bestAction != "hu" && bestAction != "kong" {
				bestAction = "pong"
				bestBot = p
			}
		}
		isPreviousPlayer := (p.Position+3)%4 == discarderPosition
		if isPreviousPlayer {
			chowCombinations := room.Game.CanChow(p.Hand, discardedTile)
			if len(chowCombinations) > 0 {
				if bestAction == "none" {
					bestAction = "chow"
					bestBot = p
					bestChowCombination = chowCombinations[0]
				}
			}
		}
	}

	if bestBot != nil {
		time.Sleep(500 * time.Millisecond) // Shorter sleep for reaction

		var success bool
		var actionToBroadcast string
		
		h.mu.Lock()
		log.Printf("Bot %s 决定执行动作: %s on %s", bestBot.Name, bestAction, discardedTile)
		var drawnTile string
		switch bestAction {
		case "hu":
			// Bot 胡牌（别人打出的牌）
			winResult := room.HandleHu(bestBot.ID, discardedTile, false)
			h.mu.Unlock()
			if winResult != nil {
				h.BroadcastGameWin(room, bestBot.ID, winResult)
			}
			return true
		case "kong":
			success, drawnTile = room.HandleKong(bestBot.ID, discardedTile, false)
			actionToBroadcast = "kong"
		case "pong":
			success = room.HandlePong(bestBot.ID, discardedTile)
			actionToBroadcast = "pong"
		case "chow":
			success = room.HandleChow(bestBot.ID, discardedTile, bestChowCombination)
			actionToBroadcast = "chow"
		}
		h.mu.Unlock() // Unlock after state change

		if success {
			if actionToBroadcast == "chow" {
				h.BroadcastChowAction(room, bestBot.ID, discardedTile, bestChowCombination)
			} else if actionToBroadcast == "kong" {
				var kongMeld game.Meld
				for _, meld := range bestBot.Melds {
					isKong := (meld.Type == "kong_exposed" || meld.Type == "kong_concealed" || meld.Type == "kong_promoted")
					if isKong && meld.Tiles[0] == discardedTile {
						kongMeld = meld
						break
					}
				}
				if kongMeld.Type != "" {
					h.BroadcastKongAction(room, bestBot.ID, kongMeld)
				} else {
					// Fallback
					h.BroadcastPlayerAction(room, bestBot.ID, "kong", discardedTile)
				}
				// 广播补牌
				if drawnTile != "" {
					h.BroadcastDrawTile(room, bestBot.ID, drawnTile)
				}
			} else {
				h.BroadcastPlayerAction(room, bestBot.ID, actionToBroadcast, discardedTile)
			}
			h.CheckAndPlayBotTurn(room, false) // Now trigger the bot's own turn, no delay
			return true
		}
	}

	return false
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
			log.Printf("警告：客户端 %s 的发送缓冲区已满，广播玩家动作消息被丢弃", client.UserName)
		}
	}
}

// broadcastBotAction 广播Bot动作（保留向后兼容）
func (h *Hub) broadcastBotAction(room *game.Room, botID, action, tile string) {
	h.BroadcastPlayerAction(room, botID, action, tile)
}

// BroadcastPlayerTingAction 广播玩家听牌的动作
func (h *Hub) BroadcastPlayerTingAction(room *game.Room, playerID string, discardedTile string) {
	var player *game.Player
	for _, p := range room.Players {
		if p.ID == playerID {
			player = p
			break
		}
	}
	if player == nil || !player.IsTing {
		return
	}

	message := map[string]interface{}{
		"type": "player_action",
		"data": map[string]interface{}{
			"playerId":     playerID,
			"action":       "ting",
			"tile":         discardedTile, // The tile that was discarded to enter Ting
			"winningTiles": player.WinningTiles,
			"currentTurn":  room.CurrentTurn,
		},
	}
	msgBytes, _ := json.Marshal(message)

	log.Printf("广播玩家听牌动作: %s, 听 %v", player.Name, player.WinningTiles)

	for _, clientInterface := range room.Clients {
		client, ok := clientInterface.(*Client)
		if !ok {
			continue
		}
		select {
		case client.Send <- msgBytes:
		default:
			log.Printf("警告：客户端 %s 的发送缓冲区已满，广播听牌消息被丢弃", client.UserName)
		}
	}
}


// BroadcastChowAction 广播吃牌动作（包含完整的吃牌组合）
func (h *Hub) BroadcastChowAction(room *game.Room, playerID, tile string, chowTiles []string) {
	message := map[string]interface{}{
		"type": "player_action",
		"data": map[string]interface{}{
			"playerId":    playerID,
			"action":      "chow",
			"tile":        tile,
			"chowTiles":   chowTiles,
			"currentTurn": room.CurrentTurn,
		},
	}

	msgBytes, _ := json.Marshal(message)

	log.Printf("广播玩家吃牌动作: %s 吃 %s，组合: %v (下一轮: %d)", playerID, tile, chowTiles, room.CurrentTurn)

	// 向所有玩家发送
	for _, clientInterface := range room.Clients {
		client, ok := clientInterface.(*Client)
		if !ok {
			continue
		}
		select {
		case client.Send <- msgBytes:
		default:
			log.Printf("警告：客户端 %s 的发送缓冲区已满，广播吃牌消息被丢弃", client.UserName)
		}
	}
}

// BroadcastKongAction 广播杠牌动作（包含完整的杠牌组合和类型）
func (h *Hub) BroadcastKongAction(room *game.Room, playerID string, meld game.Meld) {
	message := map[string]interface{}{
		"type": "player_action",
		"data": map[string]interface{}{
			"playerId":    playerID,
			"action":      "kong",
			"tile":        meld.Tiles[0], // The primary tile
			"meld":        meld,
			"currentTurn": room.CurrentTurn,
		},
	}

	msgBytes, _ := json.Marshal(message)

	log.Printf("广播玩家杠牌动作: %s 杠 %s, 类型: %s (下一轮: %d)", playerID, meld.Tiles[0], meld.Type, room.CurrentTurn)

	// 向所有玩家发送
	for _, clientInterface := range room.Clients {
		client, ok := clientInterface.(*Client)
		if !ok {
			continue
		}
		select {
		case client.Send <- msgBytes:
		default:
			log.Printf("警告：客户端 %s 的发送缓冲区已满，广播杠牌消息被丢弃", client.UserName)
		}
	}
}

// BroadcastFlowerTiles 广播花牌事件
func (h *Hub) BroadcastFlowerTiles(room *game.Room, playerID string, flowers []string) {
	message := map[string]interface{}{
		"type": "player_action",
		"data": map[string]interface{}{
			"playerId": playerID,
			"action":   "flower",
			"flowers":  flowers,
		},
	}

	msgBytes, _ := json.Marshal(message)

	log.Printf("广播玩家花牌: %s 摸到花牌 %v", playerID, flowers)

	for _, clientInterface := range room.Clients {
		client, ok := clientInterface.(*Client)
		if !ok {
			continue
		}
		select {
		case client.Send <- msgBytes:
		default:
			log.Printf("警告：客户端 %s 的发送缓冲区已满，广播花牌消息被丢弃", client.UserName)
		}
	}
}

// broadcast is a helper to send a message to all clients in a room
func (h *Hub) broadcast(room *game.Room, message []byte) {
	for _, clientInterface := range room.Clients {
		client, ok := clientInterface.(*Client)
		if !ok {
			continue
		}
		select {
		case client.Send <- message:
		default:
			log.Printf("警告：客户端 %s 的发送缓冲区已满，消息被丢弃", client.UserName)
		}
	}
}

// BroadcastPossibleActions 广播可执行动作给玩家
func (h *Hub) BroadcastPossibleActions(room *game.Room, discardedTile string, discarderPosition int) {
	for _, player := range room.Players {
		// 跳过Bot和出牌者自己
		if strings.HasPrefix(player.ID, "bot_") || player.Position == discarderPosition {
			continue
		}

		// 检测可执行动作
		possibleActions := make(map[string]interface{})

		if player.IsTing {
			// 听牌状态下只检查是否可以胡牌
			// 将打出的牌加入手牌进行检查
			tempHand := append([]string{}, player.Hand...)
			tempHand = append(tempHand, discardedTile)
			if room.Game.CanHu(tempHand, player.Melds) {
				possibleActions["hu"] = true
			}
		} else {
			// 非听牌状态下检查所有动作
			// 检查碰
			if room.Game.CanPong(player.Hand, discardedTile) {
				possibleActions["pong"] = true
			}

			// 检查吃（只能吃上家的牌）
			isPreviousPlayer := (player.Position + 3) % 4 == discarderPosition
			if isPreviousPlayer {
				chowCombinations := room.Game.CanChow(player.Hand, discardedTile)
				if len(chowCombinations) > 0 {
					possibleActions["chow"] = chowCombinations
				}
			}

			// 检查槓（明槓）
			if room.Game.CanExposedKong(player, discardedTile) {
				possibleActions["kong"] = true
			}

			// 检查胡（将打出的牌加入手牌进行检查）
			tempHand := append([]string{}, player.Hand...)
			tempHand = append(tempHand, discardedTile)
			if room.Game.CanHu(tempHand, player.Melds) {
				possibleActions["hu"] = true
			}
		}

		// 如果有可执行动作，广播给该玩家
		if len(possibleActions) > 0 {
			message := map[string]interface{}{
				"type": "possible_actions",
				"data": map[string]interface{}{
					"playerId": player.ID,
					"tile":     discardedTile,
					"actions":  possibleActions,
					"timeout":  10, // 10秒超时
				},
			}

			msgBytes, _ := json.Marshal(message)

			// 只发送给该玩家
			for _, clientInterface := range room.Clients {
				client, ok := clientInterface.(*Client)
				if !ok || client.UserID != player.ID {
					continue
				}
				select {
				case client.Send <- msgBytes:
					log.Printf("广播可执行动作给玩家 %s: %v", player.Name, possibleActions)
				default:
					log.Printf("警告：无法发送可执行动作给玩家 %s", player.Name)
				}
			}
		}
	}
}

// BroadcastGameWin 广播游戏胜利事件并准备下一局
func (h *Hub) BroadcastGameWin(room *game.Room, winnerID string, result *game.WinResult) {
	var winnerName string
	for _, p := range room.Players {
		if p.ID == winnerID {
			winnerName = p.Name
			break
		}
	}

	message := map[string]interface{}{
		"type": "game_win",
		"data": map[string]interface{}{
			"winnerId":    winnerID,
			"winnerName":  winnerName,
			"winResult":   result,
			"countdown":   8, // 告知前端倒计时秒数
		},
	}

	msgBytes, _ := json.Marshal(message)
	log.Printf("广播游戏胜利: 玩家 %s (%s) 胡牌", winnerName, winnerID)
	h.broadcast(room, msgBytes)

	// 8秒后开始新的一局
	go func() {
		time.Sleep(8 * time.Second)

		h.mu.Lock()
		// 检查游戏是否还未开始新的一局
		if !room.GameStarted {
			log.Printf("8秒倒计时结束，开始新的一局...")
			room.NextRound()
		} else {
			log.Printf("新的一局已在倒计时期间开始，取消自动开始")
			h.mu.Unlock()
			return
		}
		h.mu.Unlock()

		// 发牌
		h.dealTiles(room)

		// 发送游戏开始消息（通知新回合）
		startMessage := room.GetGameStartMessage()
		h.broadcast(room, startMessage)

		// 检查Bot回合
		h.CheckAndPlayBotTurn(room, false)
	}()
}

