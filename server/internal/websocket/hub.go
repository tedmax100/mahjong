package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"mahjong/internal/ai"
	"mahjong/internal/game"
	"mahjong/internal/logger"
	"mahjong/internal/model"
	"mahjong/internal/scoring"
	"math/rand"
	"strings"
	"sync"
	"time"
)

// Hub 管理所有 WebSocket 連接和房間
type Hub struct {
	// 所有房間
	rooms map[string]*game.Room

	// 註冊請求
	register chan *Client

	// 註銷請求
	unregister chan *Client

	// 互斥鎖
	mu sync.RWMutex
}

// NewHub 建立新的 Hub
func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]*game.Room),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run 運行 Hub 主迴圈
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

// registerClient 註冊新客戶端
func (h *Hub) registerClient(client *Client) {
	var roomToUpdate *game.Room
	var shouldStartGame bool

	h.mu.Lock()

	room, exists := h.rooms[client.RoomID]
	if !exists {
		room = game.NewRoom(client.RoomID)
		h.rooms[client.RoomID] = room
		log.Printf("建立新房間: %s", client.RoomID)
	}

	if err := room.AddPlayer(client.UserID, client.UserName, false); err != nil {
		log.Printf("玩家加入房間失敗: %v", err)
		h.mu.Unlock() // Unlock before sending to channel
		// 發送具體的錯誤訊息
		errorMsg := fmt.Sprintf(`{"type":"error","message":"%s"}`, err.Error())
		client.Send <- []byte(errorMsg)
		return
	}

	client.Room = room
	room.Clients[client.UserID] = client
	log.Printf("玩家 %s 加入房間 %s", client.UserName, client.RoomID)

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
		log.Printf("房間 %s 已滿 4 人，準備開始遊戲", roomToUpdate.ID)
		h.startGame(roomToUpdate)
	}
}

// unregisterClient 註銷客戶端
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()

	var roomToCheck *game.Room
	var shouldCheckBotTurn bool

	if client.Room != nil {
		room := client.Room

		// 從房間移除客戶端連接
		delete(room.Clients, client.UserID)

		log.Printf("玩家 %s 離開房間 %s", client.UserName, room.ID)

		// 如果遊戲已開始，將玩家轉為 Bot 代打
		if room.GameStarted {
			playerName := client.UserName
			// 將玩家標記為 Bot
			for _, player := range room.Players {
				if player.ID == client.UserID {
					player.IsBot = true
					log.Printf("玩家 %s 斷線，由 Bot 代打", playerName)
					break
				}
			}

			// 廣播玩家離開通知
			h.broadcastPlayerLeftNoLock(room, client.UserID, playerName)

			// 標記需要檢查 Bot 回合
			roomToCheck = room
			shouldCheckBotTurn = true
		} else {
			// 遊戲未開始，直接移除玩家
			room.RemovePlayer(client.UserID)
		}

		// 檢查是否還有真實玩家連線（有 WebSocket 連接的玩家）
		// 如果沒有任何真實玩家連線，刪除房間
		if len(room.Clients) == 0 {
			delete(h.rooms, room.ID)
			log.Printf("房間 %s 已無真實玩家連線，刪除房間", room.ID)
			shouldCheckBotTurn = false // 房間已刪除，不需要檢查
		} else {
			// 還有玩家，廣播房間更新
			h.broadcastRoomUpdateNoLock(room)
		}
	}

	close(client.Send)
	h.mu.Unlock()

	// 在鎖外檢查 Bot 回合，避免死鎖
	if shouldCheckBotTurn && roomToCheck != nil {
		h.CheckAndPlayBotTurn(roomToCheck, false)
	}
}

// broadcastRoomUpdate 廣播房間狀態更新
func (h *Hub) broadcastRoomUpdate(room *game.Room) {
	message := room.GetRoomUpdateMessage()
	h.broadcast(room, message)
}

// broadcastRoomUpdateNoLock 廣播房間狀態更新（調用者已持有鎖）
func (h *Hub) broadcastRoomUpdateNoLock(room *game.Room) {
	message := room.GetRoomUpdateMessage()
	h.broadcastNoLock(room, message)
}

// broadcastPlayerLeft 廣播玩家離開通知
func (h *Hub) broadcastPlayerLeft(room *game.Room, playerID, playerName string) {
	message := map[string]interface{}{
		"type": "player_left",
		"data": map[string]interface{}{
			"playerId":   playerID,
			"playerName": playerName,
		},
	}

	data, _ := json.Marshal(message)
	h.broadcast(room, data)
	log.Printf("廣播玩家離開通知: %s", playerName)
}

// broadcastPlayerLeftNoLock 廣播玩家離開通知（調用者已持有鎖）
func (h *Hub) broadcastPlayerLeftNoLock(room *game.Room, playerID, playerName string) {
	message := map[string]interface{}{
		"type": "player_left",
		"data": map[string]interface{}{
			"playerId":   playerID,
			"playerName": playerName,
		},
	}

	data, _ := json.Marshal(message)
	h.broadcastNoLock(room, data)
	log.Printf("廣播玩家離開通知: %s", playerName)
}

// BroadcastDiceRoll 廣播擲骰結果
func (h *Hub) BroadcastDiceRoll(room *game.Room, diceResult *game.DiceRollResult) {
	message := map[string]interface{}{
		"type": "dice_roll",
		"data": map[string]interface{}{
			"diceResults":     diceResult.DiceResults,
			"totalSum":        diceResult.TotalSum,
			"dealerPlayerId":  diceResult.DealerPlayerID,
			"dealerSeatIndex": diceResult.DealerSeatIndex,
		},
	}

	msgBytes, _ := json.Marshal(message)
	log.Printf("廣播擲骰結果: %v, 總和: %d, 莊家位置: %d",
		diceResult.DiceResults, diceResult.TotalSum, diceResult.DealerSeatIndex)

	h.broadcast(room, msgBytes)
}

// startGame 開始遊戲
func (h *Hub) startGame(room *game.Room) {
	// 清空 log 檔案，開始新局
	if err := logger.ClearLog(); err != nil {
		log.Printf("清空 log 檔案失敗: %v", err)
	}

	log.Printf("房間 %s 開始遊戲流程", room.ID)

	// 擲骰決定莊家
	diceResult := game.RollDiceForDealer(room.Players)
	room.DiceRollResult = diceResult

	// 廣播擲骰事件（在 game_start 之前）
	h.BroadcastDiceRoll(room, diceResult)

	// 等待前端播放擲骰動畫（約 5 秒）
	time.Sleep(5 * time.Second)

	room.GameStarted = true
	room.StartGame()

	// 發送遊戲開始訊息（為每個玩家單獨發送，包含該玩家的位置）
	for i, player := range room.Players {
		if clientInterface, ok := room.Clients[player.ID]; ok {
			client, ok := clientInterface.(*Client)
			if !ok {
				continue
			}
			startMessage := room.GetGameStartMessage(i)
			client.Send <- startMessage
		}
	}

	// 等待一小段時間，確保 game_start 訊息先被客戶端處理
	time.Sleep(100 * time.Millisecond)

	// 發牌
	h.dealTiles(room)

	// 檢查第一個回合是否是 Bot
	h.CheckAndPlayBotTurn(room, false) // 遊戲開始時莊家直接打牌，不應該有延遲
}

// dealTiles 發牌
func (h *Hub) dealTiles(room *game.Room) {
	room.DealTiles()
	log.Printf("房間 %s 發牌完成", room.ID)

	// 向每個玩家發送他們的手牌
	for i, player := range room.Players {
		log.Printf("玩家 %s 收到 %d 張牌", player.Name, len(player.Hand))

		// 只向有 WebSocket 連接的真實玩家發送
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

// GetRoom 獲取房間
func (h *Hub) GetRoom(roomID string) *game.Room {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.rooms[roomID]
}

// CreateRoom 建立房間
func (h *Hub) CreateRoom(roomID string) *game.Room {
	h.mu.Lock()
	defer h.mu.Unlock()

	room := game.NewRoom(roomID)
	h.rooms[roomID] = room

	log.Printf("建立房間: %s", roomID)
	return room
}

// addBot 添加 Bot 玩家
func (h *Hub) addBot(room *game.Room) {
	if room == nil {
		return
	}

	var shouldStartGame bool

	h.mu.Lock()

	if len(room.Players) >= 4 {
		log.Printf("房間 %s 已滿，無法添加 Bot", room.ID)
		h.mu.Unlock()
		return
	}

	botNames := []string{"機器人·小智", "機器人·阿爾法", "機器人·貝塔", "機器人·伽馬"}
	playerCount := len(room.Players)
	if playerCount < 0 {
		playerCount = 0
	}
	botName := botNames[playerCount%len(botNames)]
	botID := "bot_" + room.ID + "_" + string(rune('A'+len(room.Players)))

	if err := room.AddPlayer(botID, botName, true); err != nil {
		log.Printf("添加 Bot 失敗: %v", err)
		h.mu.Unlock()
		return
	}
	log.Printf("Bot %s 加入房間 %s", botName, room.ID)

	if len(room.Players) == 4 && !room.GameStarted {
		shouldStartGame = true
		room.GameStarted = true // Mark as started under lock
	}
	
	h.mu.Unlock() // --- Lock Released ---

	h.broadcastRoomUpdate(room)

	if shouldStartGame {
		log.Printf("房間 %s 已滿 4 人（包含 Bot），準備開始遊戲", room.ID)
		h.startGame(room)
	}
}

// CheckAndPlayBotTurn 檢查並執行 Bot 回合
func (h *Hub) CheckAndPlayBotTurn(room *game.Room, withDelay bool) {
	log.Printf("🎮 [CheckAndPlayBotTurn] 開始檢查並執行回合，withDelay: %v", withDelay)

	if room == nil || room.Game == nil || !room.GameStarted {
		log.Printf("❌ [CheckAndPlayBotTurn] 房間或遊戲狀態無效，返回")
		return
	}

	log.Printf("🔒 [CheckAndPlayBotTurn] 獲取讀鎖")
	h.mu.RLock()
	if room.CurrentTurn < 0 || room.CurrentTurn >= len(room.Players) {
		log.Printf("❌ [CheckAndPlayBotTurn] 當前回合無效: %d", room.CurrentTurn)
		h.mu.RUnlock()
		return
	}
	currentPlayer := room.Players[room.CurrentTurn]
	currentTurnAtCheck := room.CurrentTurn
	h.mu.RUnlock()
	log.Printf("🔓 [CheckAndPlayBotTurn] 釋放讀鎖，當前玩家: %s (位置 %d), IsBot: %v", currentPlayer.Name, currentTurnAtCheck, currentPlayer.IsBot)

	// 判斷是否為 Bot（包括原生 Bot 和斷線轉為 Bot 的玩家）
	isBot := strings.HasPrefix(currentPlayer.ID, "bot_") || currentPlayer.IsBot

	// 如果是真實玩家，根據情況決定是否延遲發牌
	if !isBot {
		log.Printf("🧑 [CheckAndPlayBotTurn] 當前玩家是真實玩家")
		if withDelay {
			go func() {
				log.Printf("檢測到玩家 %s (位置 %d) 有可執行動作，開始 10 秒等待...", currentPlayer.Name, currentTurnAtCheck)
				time.Sleep(10 * time.Second)
				h.mu.Lock()
				defer h.mu.Unlock()
				log.Printf("玩家 %s 的等待時間結束。當前輪次: %d, 期望輪次: %d", currentPlayer.Name, room.CurrentTurn, currentTurnAtCheck)
				if room.GameStarted && room.CurrentTurn == currentTurnAtCheck {
					log.Printf("輪次未變，為玩家 %s 執行摸牌", currentPlayer.Name)
					h.drawForRealPlayer_needsLock(room)
				} else {
					log.Printf("輪次已從 %d 變為 %d，取消為玩家 %s 的自動摸牌", currentTurnAtCheck, room.CurrentTurn, currentPlayer.Name)
				}
			}()
		} else {
			log.Printf("沒有檢測到可執行動作，立即為玩家 %s 執行摸牌", currentPlayer.Name)
			h.mu.Lock()
			defer h.mu.Unlock()
			h.drawForRealPlayer_needsLock(room)
		}
		return
	}

	// Bot 的邏輯（包括原生 Bot 和斷線轉為 Bot 的玩家）
	if isBot {
		log.Printf("🤖 [CheckAndPlayBotTurn] 當前玩家是 Bot (IsBot: %v)", currentPlayer.IsBot)
		roomID := room.ID // 保存房間 ID 用於後續檢查
		go func() {
			log.Printf("🤖 Bot %s 的回合已開始，等待出牌...", currentPlayer.Name)
			// #nosec G404 -- Bot 延遲時間不需要加密安全隨機數
			time.Sleep(time.Duration(1000+rand.Intn(1000)) * time.Millisecond)

			log.Printf("🔒 [CheckAndPlayBotTurn-Bot] 準備獲取鎖")
			h.mu.Lock()
			log.Printf("✅ [CheckAndPlayBotTurn-Bot] 已獲取鎖")

			// 檢查房間是否還存在（可能已被刪除）
			if _, exists := h.rooms[roomID]; !exists {
				log.Printf("🚫 Bot %s 的房間 %s 已不存在，取消執行", currentPlayer.Name, roomID)
				h.mu.Unlock()
				return
			}

			if !room.GameStarted || room.CurrentTurn != currentTurnAtCheck {
				log.Printf("⚠️ Bot %s 的回合被中斷", currentPlayer.Name)
				h.mu.Unlock()
				return
			}

			// 計算總牌數（手牌 + 已展示的吃碰槓）
			totalTiles := len(currentPlayer.Hand) + len(currentPlayer.Melds)*3

			// 只有在正常輪次（總牌數 16 張）才摸牌
			// 如果剛吃/碰/槓完（總牌數 17 張），不摸牌直接出牌
			if totalTiles == 16 {
				// 記錄摸牌前的花牌數量
				flowerCountBefore := len(currentPlayer.Flowers)

				// 使用 DrawTileWithFlowerReplacement 自動處理花牌
				drawnTile := room.Game.DrawTileWithFlowerReplacement(currentPlayer)
				if drawnTile != "" {
					log.Printf("Bot %s 摸到了 %s（手牌 %d 張 + 吃碰槓 %d 組 + 花牌 %d 張）",
						currentPlayer.Name, drawnTile, len(currentPlayer.Hand), len(currentPlayer.Melds), len(currentPlayer.Flowers))
					currentPlayer.Hand = append(currentPlayer.Hand, drawnTile)

					// 如果摸牌過程中補了花牌，廣播給前端
					if len(currentPlayer.Flowers) > flowerCountBefore {
						newFlowers := currentPlayer.Flowers[flowerCountBefore:]
						log.Printf("Bot %s 的花牌: %v", currentPlayer.Name, currentPlayer.Flowers)
						h.BroadcastFlowerTiles(room, currentPlayer.ID, newFlowers)
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
							log.Printf("🎉 Bot %s 自摸 %s！準備廣播勝利", currentPlayer.Name, drawnTile)

							// 使用 room.HandleHu 處理胡牌邏輯
							winResult := room.HandleHu(currentPlayer.ID, drawnTile, true)
							h.mu.Unlock() // 在廣播前解鎖

							if winResult != nil {
								// 廣播胡牌事件
								h.BroadcastGameWin(room, currentPlayer.ID, winResult)
							}
							return
						} else {
							// 不是自摸，自動打出摸到的牌，保持聽牌狀態
							log.Printf("Bot %s 聽牌中，自動打出 %s", currentPlayer.Name, drawnTile)

							discarderPosition := currentPlayer.Position
							h.mu.Unlock()

							// 處理出牌（HandleDiscard 會自己移除手牌）
							success, isDraw := room.HandleDiscard(currentPlayer.ID, drawnTile)
							if !success {
								log.Printf("警告：Bot %s 打牌失敗", currentPlayer.Name)
								return
							}
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
									log.Printf("等待真實玩家回應可執行動作，10 秒後自動繼續...")
									currentTurnSnapshot := room.CurrentTurn
									room.StartNoResponseTimer(10*time.Second, func() {
										h.mu.Lock()
										shouldContinue := room.CurrentTurn == currentTurnSnapshot
										if shouldContinue {
											log.Printf("真實玩家未回應，繼續遊戲")
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
				// 使用新的 ai 套件
				tileToDiscard := ai.ChooseDiscard(currentPlayer.Hand)
				discarderPosition := currentPlayer.Position
				h.mu.Unlock() // 在調用回應和下一個回合檢查前解鎖

				log.Printf("Bot %s 自動打出 %s", currentPlayer.Name, tileToDiscard)

				success, isDraw := room.HandleDiscard(currentPlayer.ID, tileToDiscard)
				if !success {
					log.Printf("警告：Bot %s 打牌失敗", currentPlayer.Name)
					return
				}
				h.BroadcastPlayerAction(room, currentPlayer.ID, "discard", tileToDiscard)

				// 檢查是否流局
				if isDraw {
					h.BroadcastGameDraw(room)
					return
				}

				actionTaken := h.botsReactToDiscard(room, tileToDiscard, discarderPosition)
				if !actionTaken {
					// 廣播可執行動作給真實玩家
					h.BroadcastPossibleActions(room, tileToDiscard, discarderPosition)

					hasHumanAction := h.HasHumanAction(room, tileToDiscard, discarderPosition)

					// 如果有真實玩家可以執行動作，等待 10 秒後再繼續
					if hasHumanAction {
						log.Printf("等待真實玩家回應可執行動作，10 秒後自動繼續...")
						currentTurnSnapshot := room.CurrentTurn
						room.StartNoResponseTimer(10*time.Second, func() {
							h.mu.Lock()
							// 檢查輪次是否改變（玩家是否已經執行了動作）
							shouldContinue := room.CurrentTurn == currentTurnSnapshot
							if shouldContinue {
								log.Printf("真實玩家未回應，繼續遊戲")
								// 確保在繼續遊戲前停止計時器
								room.StopNoResponseTimer()
							}
							h.mu.Unlock()
							// 在鎖外調用 CheckAndPlayBotTurn，避免死鎖
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

// drawForRealPlayer_needsLock 為真實玩家自動發牌（需要持有鎖）
func (h *Hub) drawForRealPlayer_needsLock(room *game.Room) {
	if room == nil || room.Game == nil || !room.GameStarted {
		return
	}

	if room.CurrentTurn < 0 || room.CurrentTurn >= len(room.Players) {
		return
	}

	currentPlayer := room.Players[room.CurrentTurn]

	// 確認是真實玩家
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

			// 廣播摸牌事件
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
					log.Printf("🎉 玩家 %s 自摸 %s！準備廣播勝利", currentPlayer.Name, drawnTile)

					// 使用 room.HandleHu 處理胡牌邏輯
					winResult := room.HandleHu(currentPlayer.ID, drawnTile, true)
					if winResult != nil {
						// 廣播胡牌事件
						h.BroadcastGameWin(room, currentPlayer.ID, winResult)
					}
					return
				} else {
					// 不是自摸，自動打出摸到的牌，保持聽牌狀態
					log.Printf("玩家 %s 聽牌中，自動打出 %s", currentPlayer.Name, drawnTile)

					// 處理出牌（HandleDiscard 會自己移除手牌）
					success, isDraw := room.HandleDiscard(currentPlayer.ID, drawnTile)
					if !success {
						log.Printf("警告：玩家 %s 聽牌打牌失敗", currentPlayer.Name)
						return
					}

					// 廣播出牌動作
					h.BroadcastPlayerAction(room, currentPlayer.ID, "discard", drawnTile)

					// 檢查是否流局
					if isDraw {
						log.Printf("🎲 聽牌自動打牌後流局")
						h.BroadcastGameDraw(room)
						return
					}

					// 釋放鎖後再檢查 Bot 回應（避免死鎖）
					log.Printf("🔓 [聽牌自動打牌] 釋放鎖，準備檢查 Bot 回應...")
					h.mu.Unlock()
					log.Printf("🔍 [聽牌自動打牌] 檢查 Bot 是否回應 %s 的打牌 %s", currentPlayer.Name, drawnTile)
					actionTaken := h.botsReactToDiscard(room, drawnTile, currentPlayer.Position)
					log.Printf("🔒 [聽牌自動打牌] 重新獲取鎖，Bot 回應結果: %v", actionTaken)
					h.mu.Lock()

					if !actionTaken {
						log.Printf("✅ [聽牌自動打牌] 沒有 Bot 回應，繼續檢查真實玩家動作")
						// 廣播可執行動作給真實玩家
						h.BroadcastPossibleActions(room, drawnTile, currentPlayer.Position)

						hasHumanAction := h.HasHumanAction(room, drawnTile, currentPlayer.Position)
						log.Printf("🧑 [聽牌自動打牌] 真實玩家動作檢查結果: %v", hasHumanAction)

						// 如果有真實玩家可以執行動作，等待 10 秒後再繼續
						if hasHumanAction {
							log.Printf("⏰ [聽牌自動打牌] 等待真實玩家回應可執行動作，10 秒後自動繼續...")
							currentTurnSnapshot := room.CurrentTurn
							// 釋放鎖後再啟動計時器
							log.Printf("🔓 [聽牌自動打牌] 釋放鎖，啟動等待計時器")
							h.mu.Unlock()
							room.StartNoResponseTimer(10*time.Second, func() {
								h.mu.Lock()
								shouldContinue := room.CurrentTurn == currentTurnSnapshot
								if shouldContinue {
									log.Printf("⏰ [聽牌自動打牌] 真實玩家未回應，繼續遊戲")
									room.StopNoResponseTimer()
								}
								h.mu.Unlock()
								if shouldContinue {
									log.Printf("▶️ [聽牌自動打牌] 調用 CheckAndPlayBotTurn")
									h.CheckAndPlayBotTurn(room, false)
								}
							})
							log.Printf("🔒 [聽牌自動打牌] 重新獲取鎖（等待計時器路徑）")
							h.mu.Lock() // 重新獲取鎖以保持函數退出時的鎖狀態一致
						} else {
							// 釋放鎖後再調用 CheckAndPlayBotTurn（避免死鎖）
							log.Printf("🔓 [聽牌自動打牌] 沒有真實玩家動作，釋放鎖並調用 CheckAndPlayBotTurn")
							h.mu.Unlock()
							h.CheckAndPlayBotTurn(room, false)
							log.Printf("🔒 [聽牌自動打牌] CheckAndPlayBotTurn 完成，重新獲取鎖")
							h.mu.Lock() // 重新獲取鎖以保持函數退出時的鎖狀態一致
						}
					} else {
						log.Printf("🤖 [聽牌自動打牌] Bot 已回應，結束處理")
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

// HasHumanAction 檢查是否有任何人類玩家可以對棄牌做出反應
func (h *Hub) HasHumanAction(room *game.Room, discardedTile string, discarderPosition int) bool {
	if room == nil || room.Game == nil {
		return false
	}

	for _, p := range room.Players {
		// 只檢查人類玩家，且不是出牌者自己
		if strings.HasPrefix(p.ID, "bot_") || p.Position == discarderPosition {
			continue
		}

		// 如果玩家聽牌，只檢查胡牌
		if p.IsTing {
			tempHand := append([]string{}, p.Hand...)
			tempHand = append(tempHand, discardedTile)
			if room.Game.CanHu(tempHand, p.Melds) {
				return true
			}
			continue // 聽牌後不檢查其他動作
		}

		// 檢查胡牌
		tempHand := append([]string{}, p.Hand...)
		tempHand = append(tempHand, discardedTile)
		if room.Game.CanHu(tempHand, p.Melds) {
			return true
		}

		// 檢查碰或槓
		if room.Game.CanPong(p.Hand, discardedTile) {
			return true
		}
		// 檢查槓（別人打出牌，所以 isSelfDrawn=false）
		if room.Game.CanExposedKong(p, discardedTile, false) {
			return true
		}

		// 檢查吃（只能是上家）
		isPreviousPlayer := (p.Position + 3) % 4 == discarderPosition
		if isPreviousPlayer {
			if chowCombinations := room.Game.CanChow(p.Hand, discardedTile); len(chowCombinations) > 0 {
				return true
			}
		}
	}

	return false
}

// BroadcastDrawTile 廣播摸牌事件
func (h *Hub) BroadcastDrawTile(room *game.Room, playerID, tile string) {
	// 獲取剩餘牌數
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

	log.Printf("廣播玩家摸牌: %s 摸 %s (當前輪次: %d, 剩餘: %d)", playerID, tile, room.CurrentTurn, remainingTiles)

	// 向所有玩家發送
	for _, clientInterface := range room.Clients {
		client, ok := clientInterface.(*Client)
		if !ok {
			continue
		}
		select {
		case client.Send <- msgBytes:
		default:
			log.Printf("警告：客戶端 %s 的發送緩衝區已滿，摸牌訊息被丟棄", client.UserName)
		}
	}
}

// BroadcastGameDraw 廣播流局事件
func (h *Hub) BroadcastGameDraw(room *game.Room) {
	// 獲取剩餘牌數
	remainingTiles := 0
	if room.Game != nil {
		remainingTiles = room.Game.GetRemainingTiles()
	}

	message := map[string]interface{}{
		"type": "game_draw",
		"data": map[string]interface{}{
			"remainingTiles": remainingTiles,
			"countdown":      5, // 5 秒倒計時
		},
	}

	msgBytes, _ := json.Marshal(message)

	log.Printf("廣播流局事件，剩餘牌數: %d", remainingTiles)

	// 向所有玩家發送
	for _, clientInterface := range room.Clients {
		client, ok := clientInterface.(*Client)
		if !ok {
			continue
		}
		select {
		case client.Send <- msgBytes:
		default:
			log.Printf("警告：客戶端 %s 的發送緩衝區已滿，流局訊息被丟棄", client.UserName)
		}
	}
}

// botsReactToDiscard 讓所有機器人對一個棄牌做出反應
func (h *Hub) botsReactToDiscard(room *game.Room, discardedTile string, discarderPosition int) bool {
	log.Printf("🔍 [botsReactToDiscard] 開始檢查 Bot 回應，牌: %s，打牌者位置: %d", discardedTile, discarderPosition)

	if room == nil || room.Game == nil || !room.GameStarted {
		log.Printf("❌ [botsReactToDiscard] 房間或遊戲狀態無效，返回 false")
		return false
	}

	var bestAction = "none"
	var bestBot *game.Player
	var bestChowCombination []string

	log.Printf("🔒 [botsReactToDiscard] 獲取讀鎖，複製玩家列表")
	h.mu.RLock()
	playersCopy := make([]*game.Player, len(room.Players))
	copy(playersCopy, room.Players)
	h.mu.RUnlock()
	log.Printf("🔓 [botsReactToDiscard] 釋放讀鎖")

	for _, p := range playersCopy {
		// 判斷是否為 Bot（包括原生 Bot 和斷線轉為 Bot 的玩家）
		isBot := strings.HasPrefix(p.ID, "bot_") || p.IsBot
		if p.Position == discarderPosition || !isBot {
			continue
		}

		// 如果 Bot 已經聽牌，則只檢查是否可以胡牌
		if p.IsTing {
			tempHand := append([]string{}, p.Hand...)
			tempHand = append(tempHand, discardedTile)
			if room.Game.CanHu(tempHand, p.Melds) {
				bestAction = "hu"
				bestBot = p
			}
			continue // 跳過其他動作檢查
		}

		// 檢查胡牌（優先級最高）
		tempHand := append([]string{}, p.Hand...)
		tempHand = append(tempHand, discardedTile)
		if room.Game.CanHu(tempHand, p.Melds) {
			bestAction = "hu"
			bestBot = p
			continue // 胡牌優先級最高，直接跳出繼續
		}

		// Simplified action checking...
		// 檢查槓（別人打出牌，所以 isSelfDrawn=false）
		if room.Game.CanExposedKong(p, discardedTile, false) {
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
		log.Printf("🤖 [botsReactToDiscard] Bot %s 決定執行動作: %s on %s", bestBot.Name, bestAction, discardedTile)
		time.Sleep(500 * time.Millisecond) // Shorter sleep for reaction

		var success bool
		var actionToBroadcast string

		log.Printf("🔒 [botsReactToDiscard] 準備獲取寫鎖執行 Bot 動作")
		h.mu.Lock()

		// 檢查房間是否還存在（可能已被刪除）
		if _, exists := h.rooms[room.ID]; !exists {
			log.Printf("🚫 [botsReactToDiscard] 房間 %s 已不存在，取消執行 Bot 動作", room.ID)
			h.mu.Unlock()
			return false
		}

		log.Printf("✅ [botsReactToDiscard] 已獲取寫鎖，執行動作: %s", bestAction)
		var drawnTile string
		switch bestAction {
		case "hu":
			// Bot 胡牌（別人打出的牌）
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
		log.Printf("🔓 [botsReactToDiscard] 釋放寫鎖")
		h.mu.Unlock() // 狀態改變後解鎖

		if success {
			log.Printf("✅ [botsReactToDiscard] Bot 動作執行成功，廣播動作")
			if actionToBroadcast == "chow" {
				h.BroadcastChowAction(room, bestBot.ID, discardedTile, bestChowCombination)
			} else if actionToBroadcast == "kong" {
				var kongMeld model.Meld
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
					// 後備
					h.BroadcastPlayerAction(room, bestBot.ID, "kong", discardedTile)
				}
				// 廣播補牌
				if drawnTile != "" {
					h.BroadcastDrawTile(room, bestBot.ID, drawnTile)
				}
			} else {
				h.BroadcastPlayerAction(room, bestBot.ID, actionToBroadcast, discardedTile)
			}
			log.Printf("▶️ [botsReactToDiscard] 調用 CheckAndPlayBotTurn 觸發 Bot 自己的回合")
			h.CheckAndPlayBotTurn(room, false) // 現在觸發 Bot 自己的回合，無延遲
			log.Printf("✅ [botsReactToDiscard] Bot 回應完成，返回 true")
			return true
		} else {
			log.Printf("❌ [botsReactToDiscard] Bot 動作執行失敗")
		}
	} else {
		log.Printf("🚫 [botsReactToDiscard] 沒有 Bot 決定執行動作")
	}

	log.Printf("✅ [botsReactToDiscard] 完成檢查，返回 false")
	return false
}


// BroadcastPlayerAction 廣播玩家動作（包括 Bot 和真實玩家）
func (h *Hub) BroadcastPlayerAction(room *game.Room, playerID, action, tile string) {
	message := map[string]interface{}{
		"type": "player_action",
		"data": map[string]interface{}{
			"playerId":    playerID,
			"action":      action,
			"tile":        tile,
			"currentTurn": room.CurrentTurn, // 添加當前輪次資訊
		},
	}

	msgBytes, _ := json.Marshal(message)

	log.Printf("廣播玩家動作: %s %s %s (下一輪: %d)", playerID, action, tile, room.CurrentTurn)

	// 向所有玩家發送
	for _, clientInterface := range room.Clients {
		client, ok := clientInterface.(*Client)
		if !ok {
			continue
		}
		select {
		case client.Send <- msgBytes:
		default:
			log.Printf("警告：客戶端 %s 的發送緩衝區已滿，廣播玩家動作訊息被丟棄", client.UserName)
		}
	}
}

// broadcastBotAction 廣播 Bot 動作（保留向後相容）
func (h *Hub) broadcastBotAction(room *game.Room, botID, action, tile string) {
	h.BroadcastPlayerAction(room, botID, action, tile)
}

// BroadcastPlayerTingAction 廣播玩家聽牌的動作
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
			"tile":         discardedTile, // 打出以進入聽牌狀態的牌
			"winningTiles": player.WinningTiles,
			"currentTurn":  room.CurrentTurn,
		},
	}
	msgBytes, _ := json.Marshal(message)

	log.Printf("廣播玩家聽牌動作: %s, 聽 %v", player.Name, player.WinningTiles)

	for _, clientInterface := range room.Clients {
		client, ok := clientInterface.(*Client)
		if !ok {
			continue
		}
		select {
		case client.Send <- msgBytes:
		default:
			log.Printf("警告：客戶端 %s 的發送緩衝區已滿，廣播聽牌訊息被丟棄", client.UserName)
		}
	}
}


// BroadcastChowAction 廣播吃牌動作（包含完整的吃牌組合）
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

	log.Printf("廣播玩家吃牌動作: %s 吃 %s，組合: %v (下一輪: %d)", playerID, tile, chowTiles, room.CurrentTurn)

	// 向所有玩家發送
	for _, clientInterface := range room.Clients {
		client, ok := clientInterface.(*Client)
		if !ok {
			continue
		}
		select {
		case client.Send <- msgBytes:
		default:
			log.Printf("警告：客戶端 %s 的發送緩衝區已滿，廣播吃牌訊息被丟棄", client.UserName)
		}
	}
}

// BroadcastKongAction 廣播槓牌動作（包含完整的槓牌組合和類型）
func (h *Hub) BroadcastKongAction(room *game.Room, playerID string, meld model.Meld) {
	message := map[string]interface{}{
		"type": "player_action",
		"data": map[string]interface{}{
			"playerId":    playerID,
			"action":      "kong",
			"tile":        meld.Tiles[0], // 主要牌
			"meld":        meld,
			"currentTurn": room.CurrentTurn,
		},
	}

	msgBytes, _ := json.Marshal(message)

	log.Printf("廣播玩家槓牌動作: %s 槓 %s, 類型: %s (下一輪: %d)", playerID, meld.Tiles[0], meld.Type, room.CurrentTurn)

	// 向所有玩家發送
	for _, clientInterface := range room.Clients {
		client, ok := clientInterface.(*Client)
		if !ok {
			continue
		}
		select {
		case client.Send <- msgBytes:
		default:
			log.Printf("警告：客戶端 %s 的發送緩衝區已滿，廣播槓牌訊息被丟棄", client.UserName)
		}
	}
}

// BroadcastFlowerTiles 廣播花牌事件
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

	log.Printf("廣播玩家花牌: %s 摸到花牌 %v", playerID, flowers)

	for _, clientInterface := range room.Clients {
		client, ok := clientInterface.(*Client)
		if !ok {
			continue
		}
		select {
		case client.Send <- msgBytes:
		default:
			log.Printf("警告：客戶端 %s 的發送緩衝區已滿，廣播花牌訊息被丟棄", client.UserName)
		}
	}
}

// broadcast 是一個幫助函數，用於向房間內的所有客戶端發送訊息
func (h *Hub) broadcast(room *game.Room, message []byte) {
	for _, clientInterface := range room.Clients {
		client, ok := clientInterface.(*Client)
		if !ok {
			continue
		}
		select {
		case client.Send <- message:
		default:
			log.Printf("警告：客戶端 %s 的發送緩衝區已滿，訊息被丟棄", client.UserName)
		}
	}
}

// broadcastNoLock 廣播訊息（與 broadcast 相同，用於區分調用場景）
func (h *Hub) broadcastNoLock(room *game.Room, message []byte) {
	for _, clientInterface := range room.Clients {
		client, ok := clientInterface.(*Client)
		if !ok {
			continue
		}
		select {
		case client.Send <- message:
		default:
			log.Printf("警告：客戶端 %s 的發送緩衝區已滿，訊息被丟棄", client.UserName)
		}
	}
}

// BroadcastPossibleActions 廣播可執行動作給玩家
func (h *Hub) BroadcastPossibleActions(room *game.Room, discardedTile string, discarderPosition int) {
	for _, player := range room.Players {
		// 跳過 Bot 和出牌者自己
		if strings.HasPrefix(player.ID, "bot_") || player.Position == discarderPosition {
			continue
		}

		// 檢測可執行動作
		possibleActions := make(map[string]interface{})

		if player.IsTing {
			// 聽牌狀態下只檢查是否可以胡牌
			// 將打出的牌加入手牌進行檢查
			tempHand := append([]string{}, player.Hand...)
			tempHand = append(tempHand, discardedTile)
			if room.Game.CanHu(tempHand, player.Melds) {
				possibleActions["hu"] = true
			}
		} else {
			// 非聽牌狀態下檢查所有動作
			// 檢查碰
			if room.Game.CanPong(player.Hand, discardedTile) {
				possibleActions["pong"] = true
			}

			// 檢查吃（只能吃上家的牌）
			isPreviousPlayer := (player.Position + 3) % 4 == discarderPosition
			if isPreviousPlayer {
				chowCombinations := room.Game.CanChow(player.Hand, discardedTile)
				if len(chowCombinations) > 0 {
					possibleActions["chow"] = chowCombinations
				}
			}

			// 檢查槓（明槓，別人打出牌，所以 isSelfDrawn=false）
			if room.Game.CanExposedKong(player, discardedTile, false) {
				possibleActions["kong"] = true
			}

			// 檢查胡（將打出的牌加入手牌進行檢查）
			tempHand := append([]string{}, player.Hand...)
			tempHand = append(tempHand, discardedTile)
			if room.Game.CanHu(tempHand, player.Melds) {
				possibleActions["hu"] = true
			}
		}

		// 如果有可執行動作，廣播給該玩家
		if len(possibleActions) > 0 {
			message := map[string]interface{}{
				"type": "possible_actions",
				"data": map[string]interface{}{
					"playerId": player.ID,
					"tile":     discardedTile,
					"actions":  possibleActions,
					"timeout":  10, // 10 秒超時
				},
			}

			msgBytes, _ := json.Marshal(message)

			// 只發送給該玩家
			for _, clientInterface := range room.Clients {
				client, ok := clientInterface.(*Client)
				if !ok || client.UserID != player.ID {
					continue
				}
				select {
				case client.Send <- msgBytes:
					log.Printf("廣播可執行動作給玩家 %s: %v", player.Name, possibleActions)
				default:
					log.Printf("警告：無法發送可執行動作給玩家 %s", player.Name)
				}
			}
		}
	}
}

// BroadcastGameWin 廣播遊戲勝利事件並準備下一局
func (h *Hub) BroadcastGameWin(room *game.Room, winnerID string, result *scoring.WinResult) {
	var winnerName string
	for _, p := range room.Players {
		if p.ID == winnerID {
			winnerName = p.Name
			break
		}
	}

	countdown := 7 // 倒計時秒數
	animationDelay := 5 // 前端動畫延遲秒數

	message := map[string]interface{}{
		"type": "game_win",
		"data": map[string]interface{}{
			"winnerId":    winnerID,
			"winnerName":  winnerName,
			"winResult":   result,
			"countdown":   countdown, // 告知前端倒計時秒數
		},
	}

	msgBytes, _ := json.Marshal(message)
	log.Printf("廣播遊戲勝利: 玩家 %s (%s) 胡牌", winnerName, winnerID)
	h.broadcast(room, msgBytes)

	// 等待前端動畫 + 倒計時完成後開始新的一局
	totalWaitTime := countdown + animationDelay
	go func() {
		time.Sleep(time.Duration(totalWaitTime) * time.Second)

		h.mu.Lock()
		// 檢查遊戲是否還未開始新的一局
		if !room.GameStarted {
			// 清空 log 檔案，開始新局
			if err := logger.ClearLog(); err != nil {
				log.Printf("清空 log 檔案失敗: %v", err)
			}
			log.Printf("%d 秒等待結束（動畫 %d 秒 + 倒計時 %d 秒），開始新的一局...", totalWaitTime, animationDelay, countdown)
			room.NextRound()
		} else {
			log.Printf("新的一局已在倒計時期間開始，取消自動開始")
			h.mu.Unlock()
			return
		}
		h.mu.Unlock()

		// 廣播房間更新（確保客戶端有最新的玩家資訊）
		h.broadcastRoomUpdate(room)

		// 發送遊戲開始訊息（為每個玩家單獨發送，包含該玩家的位置）
		for i, player := range room.Players {
			if clientInterface, ok := room.Clients[player.ID]; ok {
				client, ok := clientInterface.(*Client)
				if !ok {
					continue
				}
				startMessage := room.GetGameStartMessage(i)
				client.Send <- startMessage
			}
		}

		// 發牌
		h.dealTiles(room)

		// 檢查 Bot 回合
		h.CheckAndPlayBotTurn(room, false)
	}()
}
