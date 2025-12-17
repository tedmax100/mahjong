package game

import (
	"encoding/json"
	"errors"
	"log"
	"mahjong/internal/model"
	"mahjong/internal/scoring"
	"reflect"
	"sort"
	"sync"
	"time"
)

// ActionPriority 定義動作優先權等級
type ActionPriority int

const (
	PriorityChow ActionPriority = 1 // 吃牌（最低）
	PriorityPong ActionPriority = 2 // 碰牌
	PriorityKong ActionPriority = 3 // 槓牌
	PriorityHu   ActionPriority = 4 // 胡牌（最高）
)

// PendingAction 代表一個待處理的動作請求
type PendingAction struct {
	PlayerID   string
	PlayerPos  int
	ActionType string // "chow", "pong", "kong", "hu"
	Priority   ActionPriority
	Tile       string
	Data       map[string]interface{} // 額外數據（如吃牌組合）
	Timestamp  time.Time
}

// Room 代表一個遊戲房間
type Room struct {
	ID                  string
	Players             []*Player
	Clients             map[string]interface{} // 儲存 WebSocket 客戶端
	Game                *MahjongGame
	GameStarted         bool
	CurrentTurn         int             // 當前輪到誰（0-3）
	LastDiscardPlayer   int             // 最後打牌的玩家位置（用於檢查吃牌資格）
	LastDiscardTile     string          // 最後打出的牌
	PendingActions      []PendingAction // 待處理的動作請求
	ActionTimeout       *time.Timer     // 動作超時計時器
	ActionMutex         sync.Mutex      // 保護 PendingActions 的互斥鎖
	IsWaitingForActions bool            // 是否正在等待玩家動作回應
	noResponseTimer     *time.Timer
	timerMu             sync.Mutex
	DiceRollResult      *DiceRollResult // 擲骰結果（用於斷線重連）
}

// Player 代表一個玩家
type Player struct {
	ID           string
	Name         string
	Position     int          // 0=東, 1=南, 2=西, 3=北
	Hand         []string     // 手牌
	Score        int
	Melds        []model.Meld // 已展示的牌組（碰、槓等）
	Flowers      []string     // 花牌
	IsTing       bool         // 玩家是否已聽牌
	WinningTiles []string     // 聽牌所胡的牌
	IsBot        bool         // 是否為 Bot（包含斷線由 Bot 代打的玩家）
}

// StartNoResponseTimer 啟動一個計時器，如果未收到玩家動作則觸發
func (r *Room) StartNoResponseTimer(d time.Duration, cb func()) {
	r.timerMu.Lock()
	defer r.timerMu.Unlock()
	// Stop any existing timer
	if r.noResponseTimer != nil {
		r.noResponseTimer.Stop()
	}
	r.noResponseTimer = time.AfterFunc(d, cb)
}

// StopNoResponseTimer 停止無回應計時器
func (r *Room) StopNoResponseTimer() {
	r.timerMu.Lock()
	defer r.timerMu.Unlock()
	if r.noResponseTimer != nil {
		r.noResponseTimer.Stop()
		r.noResponseTimer = nil
	}
}

// GetTotalTiles 計算玩家的總牌數（手牌 + 吃碰槓）
func (p *Player) GetTotalTiles() int {
	total := len(p.Hand)
	for _, meld := range p.Melds {
		total += len(meld.Tiles)
	}
	return total
}

// LogPlayerHand 記錄單個玩家的手牌狀態（用於 debug）
func LogPlayerHand(player *Player, action string) {
	if player == nil {
		return
	}

	// 複製手牌並排序
	sortedHand := make([]string, len(player.Hand))
	copy(sortedHand, player.Hand)
	sort.Strings(sortedHand)

	// 格式化吃碰槓牌組
	melds := make([]string, 0)
	for _, meld := range player.Melds {
		melds = append(melds, meld.Type+": "+formatTiles(meld.Tiles))
	}

	actionPrefix := ""
	if action != "" {
		actionPrefix = "[" + action + "] "
	}

	log.Printf("📋 %s玩家 %s (位置 %d)", actionPrefix, player.Name, player.Position)
	log.Printf("   手牌 (%d 張): %s", len(sortedHand), formatTiles(sortedHand))
	if len(melds) > 0 {
		log.Printf("   吃碰槓: %v", melds)
	}
	log.Printf("   總牌數: %d 張", player.GetTotalTiles())
}

// LogAllPlayersHands 記錄所有玩家的手牌狀態
func LogAllPlayersHands(room *Room, action string) {
	log.Println("============================================================")
	if action != "" {
		log.Printf("📊 %s", action)
	} else {
		log.Println("📊 當前遊戲狀態")
	}
	log.Println("============================================================")

	for _, player := range room.Players {
		LogPlayerHand(player, "")
	}

	log.Println("============================================================")
}

// formatTiles 格式化牌列表為字串
func formatTiles(tiles []string) string {
	result := ""
	for i, tile := range tiles {
		if i > 0 {
			result += " "
		}
		result += tile
	}
	return result
}

// NewRoom 建立新房間
func NewRoom(id string) *Room {
	return &Room{
		ID:          id,
		Players:     make([]*Player, 0, 4),
		Clients:     make(map[string]interface{}),
		GameStarted: false,
	}
}

// AddPlayer 新增玩家到房間
func (r *Room) AddPlayer(userID, userName string, isBot bool) error {
	// 檢查遊戲是否已經開始
	if r.GameStarted {
		return errors.New("遊戲已開始")
	}

	if len(r.Players) >= 4 {
		return errors.New("房間已滿")
	}

	player := &Player{
		ID:       userID,
		Name:     userName,
		Position: len(r.Players),
		Hand:     make([]string, 0, 17),
		Score:    1000,
		Melds:    make([]model.Meld, 0),
		Flowers:  make([]string, 0),
		IsBot:    isBot,
	}

	r.Players = append(r.Players, player)
	log.Printf("玩家 %s 加入房間 %s (位置: %d, IsBot: %v)", userName, r.ID, player.Position, isBot)

	return nil
}

// RemovePlayer 從房間移除玩家
func (r *Room) RemovePlayer(userID string) {
	for i, player := range r.Players {
		if player.ID == userID {
			r.Players = append(r.Players[:i], r.Players[i+1:]...)
			log.Printf("玩家 %s 離開房間 %s", player.Name, r.ID)
			return
		}
	}
}

// StartGame 開始遊戲
func (r *Room) StartGame() {
	r.Game = NewMahjongGame(r.Players)
	r.GameStarted = true

	// 使用擲骰結果決定莊家位置（如果有的話）
	if r.DiceRollResult != nil {
		r.Game.Dealer = r.DiceRollResult.DealerSeatIndex
		r.CurrentTurn = r.DiceRollResult.DealerSeatIndex
		log.Printf("房間 %s 遊戲開始，依擲骰結果莊家為位置 %d", r.ID, r.DiceRollResult.DealerSeatIndex)
	} else {
		r.CurrentTurn = 0 // 預設莊家（位置 0）先出牌
		log.Printf("房間 %s 遊戲開始，預設莊家先出牌", r.ID)
	}
}

// DealTiles 發牌
func (r *Room) DealTiles() {
	if r.Game != nil {
		r.Game.DealTiles()

		// 📋 記錄所有玩家的初始手牌
		LogAllPlayersHands(r, "遊戲開始 - 所有玩家初始手牌")
	}
}

// GetRoomUpdateMessage 取得房間更新訊息
func (r *Room) GetRoomUpdateMessage() []byte {
	playersData := make([]map[string]interface{}, len(r.Players))
	for i, player := range r.Players {
		playersData[i] = map[string]interface{}{
			"id":       player.ID,
			"name":     player.Name,
			"position": player.Position,
			"score":    player.Score,
		}
	}

	message := map[string]interface{}{
		"type": "room_update",
		"data": map[string]interface{}{
			"playerCount": len(r.Players),
			"players":     playersData,
		},
	}

	data, _ := json.Marshal(message)
	return data
}

// GetGameStartMessage 取得遊戲開始訊息（不包含牌）
func (r *Room) GetGameStartMessage(playerIndex int) []byte {
	if playerIndex < 0 || playerIndex >= len(r.Players) {
		return []byte{}
	}
	player := r.Players[playerIndex]

	// 計算該玩家的門風
	seatWind := r.Game.GetSeatWind(player.Position)

	messageData := map[string]interface{}{
		"roomId":         r.ID,
		"currentTurn":    r.CurrentTurn,
		"myPosition":     player.Position,
		"dealerPosition": r.Game.Dealer,
		"roundWind":      r.Game.RoundWind,         // 場風
		"seatWind":       seatWind,                 // 該玩家的門風
		"allSeatWinds":   r.Game.GetAllSeatWinds(), // 所有玩家的門風
	}

	// 包含擲骰結果（用於斷線重連顯示）
	if r.DiceRollResult != nil {
		messageData["diceRollResult"] = r.DiceRollResult
	}

	message := map[string]interface{}{
		"type": "game_start",
		"data": messageData,
	}

	data, _ := json.Marshal(message)
	return data
}

// GetDealTilesMessage 取得發牌訊息
func (r *Room) GetDealTilesMessage(playerIndex int) []byte {
	if playerIndex < 0 || playerIndex >= len(r.Players) {
		return []byte{}
	}
	player := r.Players[playerIndex]

	message := map[string]interface{}{
		"type": "deal_tiles",
		"data": map[string]interface{}{
			"tiles":    player.Hand,
			"position": player.Position,
		},
	}

	data, _ := json.Marshal(message)
	return data
}

// HandleDiscard 處理出牌
// 返回值：(success bool, isDraw bool)
// - success: 打牌是否成功
// - isDraw: 是否流局（僅在 success=true 時有意義）
func (r *Room) HandleDiscard(userID, tile string) (bool, bool) {
	// 找到玩家
	var player *Player
	for _, p := range r.Players {
		if p.ID == userID {
			player = p
			break
		}
	}

	if player == nil {
		log.Printf("未找到玩家: %s", userID)
		return false, false
	}

	// 檢查是否輪到該玩家
	if player.Position != r.CurrentTurn {
		log.Printf("還沒輪到玩家 %s（當前回合: %d，玩家位置: %d）", player.Name, r.CurrentTurn, player.Position)
		return false, false
	}

	// 檢查是否正在等待動作回應（如胡、碰、槓等）
	if r.IsWaitingForActions {
		log.Printf("❌ 玩家 %s 無法打牌：正在等待動作回應（請選擇胡/碰/槓/過）", player.Name)
		return false, false
	}

	// 檢查手牌數量是否正確（應該有 17 張牌才能打牌，槓後補牌可能是 18 張）
	// 每個槓（kong）比普通碰多 1 張牌，所以需要調整預期值
	totalTiles := player.GetTotalTiles()

	// 計算槓的數量（每個槓多 1 張牌）
	kongCount := 0
	for _, meld := range player.Melds {
		if meld.Type == "kong_concealed" || meld.Type == "kong_promoted" || meld.Type == "kong_exposed" {
			kongCount++
		}
	}

	// 基本預期：17 或 18 張
	// 每個槓額外增加 1 張
	expectedMin := 17 + kongCount
	expectedMax := 18 + kongCount

	if totalTiles != expectedMin && totalTiles != expectedMax {
		log.Printf("❌ 玩家 %s 手牌數量錯誤！總牌數 %d（手牌 %d + 吃碰槓 %d 組，其中 %d 個槓），預期 %d 或 %d 張，拒絕打牌",
			player.Name, totalTiles, len(player.Hand), len(player.Melds), kongCount, expectedMin, expectedMax)
		return false, false
	}

	log.Printf("玩家 %s 打出 %s", player.Name, tile)

	// 從手牌中移除
	for i, t := range player.Hand {
		if t == tile {
			player.Hand = append(player.Hand[:i], player.Hand[i+1:]...)
			break
		}
	}

	// 將牌加入棄牌堆
	if r.Game != nil {
		r.Game.DiscardPile = append(r.Game.DiscardPile, tile)
		log.Printf("棄牌堆: %v (共 %d 張)", r.Game.DiscardPile, len(r.Game.DiscardPile))

		// 檢查流局
		if r.Game.CheckDraw() {
			log.Printf("流局！牌山剩餘 %d 張", r.Game.GetRemainingTiles())
			r.GameStarted = false
			return true, true // 打牌成功，且流局
		}
	}

	// 記錄最後打牌的玩家和牌（用於檢查吃牌資格）
	r.LastDiscardPlayer = player.Position
	r.LastDiscardTile = tile

	// 📋 記錄打牌後的手牌狀態
	LogPlayerHand(player, "打牌: "+tile)

	// 切換到下一個玩家
	// TODO: 當實作完整優先權處理後，這裡應該等待其他玩家響應後再切換
	r.NextTurn()
	log.Printf("輪到下一位玩家（位置: %d）", r.CurrentTurn)

	return true, false // 打牌成功，沒有流局
}

// NextTurn 切換到下一個玩家
func (r *Room) NextTurn() {
	r.CurrentTurn = (r.CurrentTurn + 1) % 4
}

// HandlePong 處理碰牌
func (r *Room) HandlePong(userID, tile string) bool {
	// 找到玩家
	var player *Player
	for _, p := range r.Players {
		if p.ID == userID {
			player = p
			break
		}
	}

	if player == nil {
		log.Printf("未找到玩家: %s", userID)
		return false
	}

	// 檢查是否可以碰（需要手牌中有至少 2 張相同的牌）
	if r.Game == nil || !r.Game.CanPong(player.Hand, tile) {
		log.Printf("玩家 %s 無法碰 %s", player.Name, tile)
		return false
	}

	log.Printf("玩家 %s 碰 %s", player.Name, tile)

	// 從手牌中移除 2 張
	removed := 0
	for i := len(player.Hand) - 1; i >= 0 && removed < 2; i-- {
		if player.Hand[i] == tile {
			player.Hand = append(player.Hand[:i], player.Hand[i+1:]...)
			removed++
		}
	}

	// 從棄牌堆中移除最後一張（即被碰的牌）
	if len(r.Game.DiscardPile) > 0 {
		r.Game.DiscardPile = r.Game.DiscardPile[:len(r.Game.DiscardPile)-1]
	}

	// 添加到已展示的牌組
	player.Melds = append(player.Melds, model.Meld{
		Type:  "pong",
		Tiles: []string{tile, tile, tile},
	})

	// 碰牌後輪到該玩家出牌
	r.CurrentTurn = player.Position
	log.Printf("碰牌成功，輪到玩家 %s 出牌", player.Name)

	// 📋 記錄碰牌後的手牌狀態
	LogPlayerHand(player, "碰牌: "+tile)

	return true
}

// HandleChow 處理吃牌（只能吃上家打出的牌）
func (r *Room) HandleChow(userID, tile string, chowTiles []string) bool {
	// 找到玩家
	var player *Player
	for _, p := range r.Players {
		if p.ID == userID {
			player = p
			break
		}
	}

	if player == nil {
		log.Printf("未找到玩家: %s", userID)
		return false
	}

	if r.Game == nil {
		return false
	}

	// 檢查是否是上家打出的牌
	// 台灣麻將規則：只能吃上家的牌
	previousPlayer := (player.Position + 3) % 4 // 上家位置
	if r.LastDiscardPlayer != previousPlayer {
		log.Printf("玩家 %s 只能吃上家的牌（最後出牌者位置: %d，上家位置: %d）",
			player.Name, r.LastDiscardPlayer, previousPlayer)
		return false
	}

	// 驗證吃牌組合是否有效
	validCombinations := r.Game.CanChow(player.Hand, tile)
	if len(validCombinations) == 0 {
		log.Printf("玩家 %s 無法吃 %s", player.Name, tile)
		return false
	}

	// 檢查提供的吃牌組合是否在有效組合中
	isValidCombination := false
	for _, combo := range validCombinations {
		if len(combo) == len(chowTiles) && isSameCombination(combo, chowTiles) {
			isValidCombination = true
			break
		}
	}

	if !isValidCombination {
		log.Printf("玩家 %s 提供的吃牌組合無效: %v", player.Name, chowTiles)
		return false
	}

	log.Printf("玩家 %s 吃 %s，組合: %v", player.Name, tile, chowTiles)

	// 從手牌中移除需要的牌（除了吃的那張）
	for _, chowTile := range chowTiles {
		if chowTile == tile {
			continue // 跳過要吃的牌（從棄牌堆拿）
		}

		// 從手牌中移除
		for i, t := range player.Hand {
			if t == chowTile {
				player.Hand = append(player.Hand[:i], player.Hand[i+1:]...)
				break
			}
		}
	}

	// 從棄牌堆中移除最後一張（即被吃的牌）
	if len(r.Game.DiscardPile) > 0 {
		r.Game.DiscardPile = r.Game.DiscardPile[:len(r.Game.DiscardPile)-1]
	}

	// 添加到已展示的牌組
	player.Melds = append(player.Melds, model.Meld{
		Type:  "chow",
		Tiles: chowTiles,
	})

	// 吃牌後輪到該玩家出牌
	r.CurrentTurn = player.Position
	log.Printf("吃牌成功，輪到玩家 %s 出牌", player.Name)

	// 📋 記錄吃牌後的手牌狀態
	LogPlayerHand(player, "吃牌: "+formatTiles(chowTiles))

	return true
}

// isSameCombination 檢查兩個牌組是否相同（忽略順序）
func isSameCombination(combo1, combo2 []string) bool {
	if len(combo1) != len(combo2) {
		return false
	}
	// 建立副本並排序
	c1 := make([]string, len(combo1))
	copy(c1, combo1)
	sort.Strings(c1)

	c2 := make([]string, len(combo2))
	copy(c2, combo2)
	sort.Strings(c2)

	return reflect.DeepEqual(c1, c2)
}

// HandleKong 處理槓牌
func (r *Room) HandleKong(userID, tile string, isConcealed bool) (bool, string) {
	// 找到玩家
	var player *Player
	for _, p := range r.Players {
		if p.ID == userID {
			player = p
			break
		}
	}

	if player == nil {
		log.Printf("未找到玩家: %s", userID)
		return false, ""
	}

	if r.Game == nil {
		return false, ""
	}

	var drawnTile string

	if isConcealed {
		// 暗槓：手牌中有 4 張相同的牌
		count := 0
		for _, t := range player.Hand {
			if t == tile {
				count++
			}
		}

		if count < 4 {
			log.Printf("玩家 %s 無法暗槓 %s（手牌不足 4 張）", player.Name, tile)
			return false, ""
		}

		log.Printf("玩家 %s 暗槓 %s", player.Name, tile)

		// 從手牌中移除 4 張
		removed := 0
		for i := len(player.Hand) - 1; i >= 0 && removed < 4; i-- {
			if player.Hand[i] == tile {
				player.Hand = append(player.Hand[:i], player.Hand[i+1:]...)
				removed++
			}
		}

		// 添加到已展示的牌組
		player.Melds = append(player.Melds, model.Meld{
			Type:  "kong_concealed",
			Tiles: []string{tile, tile, tile, tile},
		})

	} else {
		// 明槓：手牌中有 3 張，或已碰出，加上別人打出的 1 張
		// 判斷是自己摸牌還是別人打出牌：
		// 檢查棄牌堆最後一張是否是這張牌
		// 如果是，說明是回應別人打出的牌（明槓）
		// 如果不是，說明是自己摸牌後想補槓
		isSelfDrawn := len(r.Game.DiscardPile) == 0 || r.Game.DiscardPile[len(r.Game.DiscardPile)-1] != tile
		if !r.Game.CanExposedKong(player, tile, isSelfDrawn) {
			log.Printf("玩家 %s 無法明槓 %s (isSelfDrawn=%v)", player.Name, tile, isSelfDrawn)
			return false, ""
		}

		// 判斷是加槓還是明槓
		isPromotedKong := false
		for i, meld := range player.Melds {
			if meld.Type == "pong" && meld.Tiles[0] == tile {
				log.Printf("玩家 %s 加槓 %s", player.Name, tile)

				// 從手牌中移除 1 張（如果是自己摸牌補槓）
				// 如果是別人打出的牌，則不需要從手牌中移除
				if isSelfDrawn {
					for j := len(player.Hand) - 1; j >= 0; j-- {
						if player.Hand[j] == tile {
							player.Hand = append(player.Hand[:j], player.Hand[j+1:]...)
							break
						}
					}
				}

				player.Melds[i].Type = "kong_promoted"
				player.Melds[i].Tiles = append(player.Melds[i].Tiles, tile)
				isPromotedKong = true
				break
			}
		}

		if !isPromotedKong {
			log.Printf("玩家 %s 明槓 %s", player.Name, tile)
			// 從手牌中移除 3 張
			removed := 0
			for i := len(player.Hand) - 1; i >= 0 && removed < 3; i-- {
				if player.Hand[i] == tile {
					player.Hand = append(player.Hand[:i], player.Hand[i+1:]...)
					removed++
				}
			}
			// 添加到已展示的牌組
			player.Melds = append(player.Melds, model.Meld{
				Type:  "kong_exposed",
				Tiles: []string{tile, tile, tile, tile},
			})

			// 大明槓：從棄牌堆中移除最後一張（別人打出的牌）
			if !isSelfDrawn && len(r.Game.DiscardPile) > 0 {
				r.Game.DiscardPile = r.Game.DiscardPile[:len(r.Game.DiscardPile)-1]
			}
		}
		// 加槓（isPromotedKong=true）：不需要從棄牌堆移除，因為是自己摸到的牌
	}

	// 槓牌後需要補牌
	if len(r.Game.Deck) > 0 {
		// 槓牌補牌需要從牌尾拿
		drawnTile = r.Game.DrawTileFromEnd()
		if drawnTile != "" {
			player.Hand = append(player.Hand, drawnTile)
			log.Printf("玩家 %s 槓牌後補牌: %s", player.Name, drawnTile)
		}
	}

	// 槓牌後輪到該玩家出牌
	r.CurrentTurn = player.Position
	log.Printf("槓牌成功，輪到玩家 %s 出牌", player.Name)

	// 📋 記錄槓牌後的手牌狀態
	kongType := "kong"
	if isConcealed {
		kongType = "kong_concealed"
	}
	LogPlayerHand(player, "槓牌: "+tile+" ("+kongType+")")

	return true, drawnTile
}

// HandleHu 處理胡牌
func (r *Room) HandleHu(userID string, winTile string, isSelfDrawn bool) *scoring.WinResult {
	return r.HandleHuWithConditions(userID, winTile, isSelfDrawn, nil)
}

// HandleHuWithConditions 處理胡牌（支援特殊情境）
func (r *Room) HandleHuWithConditions(userID string, winTile string, isSelfDrawn bool, specialConditions []scoring.SpecialCondition) *scoring.WinResult {
	// 找到玩家
	var player *Player
	for _, p := range r.Players {
		if p.ID == userID {
			player = p
			break
		}
	}

	if player == nil {
		log.Printf("未找到玩家: %s", userID)
		return nil
	}

	if r.Game == nil {
		return nil
	}

	// 根據是否自摸，準備用於驗證的最終手牌
	var validationHand []string
	if isSelfDrawn {
		// 自摸時，牌已在 draw-logic 中加入手牌
		validationHand = player.Hand
	} else {
		// 吃胡時，臨時將胡的牌加入手牌
		validationHand = append([]string{}, player.Hand...)
		validationHand = append(validationHand, winTile)
	}

	// 使用最終手牌檢查是否可以胡牌
	if !r.Game.CanHu(validationHand, player.Melds) {
		log.Printf("玩家 %s 無法胡牌（手牌+%s）", player.Name, winTile)
		return nil
	}

	log.Printf("玩家 %s 胡牌成功！胡牌: %s", player.Name, winTile)

	// 在算分前，確保玩家的手牌是最終的胡牌狀態
	player.Hand = validationHand

	// 取得場風和門風
	roundWind := scoring.Wind(r.Game.RoundWind)
	seatWind := scoring.Wind(r.Game.GetSeatWind(player.Position))
	isDealer := player.Position == r.Game.Dealer

	// 建立計分輸入
	scoreInput := &scoring.ScoreInput{
		RoundWind:         roundWind,
		SeatWind:          seatWind,
		IsDealer:          isDealer,
		IsSelfDrawn:       isSelfDrawn,
		Hand:              player.Hand,
		Melds:             player.Melds,
		Flowers:           player.Flowers,
		WinningTile:       winTile,
		SpecialConditions: specialConditions,
	}

	// 使用新版 scoring 計算分數
	winResult := scoring.CalculateScoreWithInput(scoreInput)

	// 更新玩家分數
	if isSelfDrawn {
		// 自摸：其他三家各付分數
		for _, p := range r.Players {
			if p.ID != player.ID {
				p.Score -= winResult.BaseScore
				player.Score += winResult.BaseScore
			}
		}
	} else {
		// 放炮：放炮者付全部分數
		// TODO: 需要記錄是誰放炮
		// 暫時簡化處理
	}

	// 標記遊戲結束
	r.GameStarted = false

	return winResult
}

// AddPendingAction 添加待處理的動作請求
func (r *Room) AddPendingAction(playerID string, actionType string, tile string, data map[string]interface{}) {
	r.ActionMutex.Lock()
	defer r.ActionMutex.Unlock()

	// 找到玩家位置
	var playerPos int
	for i, p := range r.Players {
		if p.ID == playerID {
			playerPos = i
			break
		}
	}

	// 確定優先權
	var priority ActionPriority
	switch actionType {
	case "hu":
		priority = PriorityHu
	case "kong":
		priority = PriorityKong
	case "pong":
		priority = PriorityPong
	case "chow":
		priority = PriorityChow
	default:
		log.Printf("未知的動作類型: %s", actionType)
		return
	}

	// 驗證動作是否合法
	if !r.ValidateAction(playerID, actionType, tile, data) {
		log.Printf("玩家 %s 的動作 %s 不合法，忽略", playerID, actionType)
		return
	}

	action := PendingAction{
		PlayerID:   playerID,
		PlayerPos:  playerPos,
		ActionType: actionType,
		Priority:   priority,
		Tile:       tile,
		Data:       data,
		Timestamp:  time.Now(),
	}

	r.PendingActions = append(r.PendingActions, action)
	log.Printf("添加待處理動作: 玩家 %d, 動作 %s, 優先權 %d", playerPos, actionType, priority)
}

// ValidateAction 驗證動作是否合法
func (r *Room) ValidateAction(playerID string, actionType string, tile string, data map[string]interface{}) bool {
	// 找到玩家
	var player *Player
	for _, p := range r.Players {
		if p.ID == playerID {
			player = p
			break
		}
	}

	if player == nil {
		return false
	}

	if r.Game == nil {
		return false
	}

	switch actionType {
	case "hu":
		return r.Game.CanHu(player.Hand, player.Melds)
	case "kong":
		if data != nil {
			if isConcealed, ok := data["isConcealed"].(bool); ok && isConcealed {
				// 暗槓檢查
				count := 0
				for _, t := range player.Hand {
					if t == tile {
						count++
					}
				}
				return count >= 4
			}
		}
		// ValidateAction 用於驗證 pending actions，即別人打出牌後的回應
		// 所以這裡 isSelfDrawn 應該為 false
		return r.Game.CanExposedKong(player, tile, false)
	case "pong":
		return r.Game.CanPong(player.Hand, tile)
	case "chow":
		// 檢查是否是上家打出的牌
		previousPlayer := (player.Position + 3) % 4
		if r.LastDiscardPlayer != previousPlayer {
			return false
		}
		validCombinations := r.Game.CanChow(player.Hand, tile)
		return len(validCombinations) > 0
	}

	return false
}

// ProcessPendingActions 處理所有待處理的動作，執行優先權最高的
func (r *Room) ProcessPendingActions() *PendingAction {
	r.ActionMutex.Lock()
	defer r.ActionMutex.Unlock()

	if len(r.PendingActions) == 0 {
		log.Println("沒有待處理的動作")
		return nil
	}

	// 按優先權排序（優先權高的在前）
	sort.Slice(r.PendingActions, func(i, j int) bool {
		if r.PendingActions[i].Priority != r.PendingActions[j].Priority {
			return r.PendingActions[i].Priority > r.PendingActions[j].Priority
		}
		// 優先權相同時，按時間戳排序（先提交的在前）
		return r.PendingActions[i].Timestamp.Before(r.PendingActions[j].Timestamp)
	})

	// 取得優先權最高的動作
	highestAction := r.PendingActions[0]
	log.Printf("執行優先權最高的動作: 玩家 %d, 動作 %s, 優先權 %d",
		highestAction.PlayerPos, highestAction.ActionType, highestAction.Priority)

	// 清空待處理動作列表
	r.PendingActions = []PendingAction{}
	r.IsWaitingForActions = false

	// 取消計時器
	if r.ActionTimeout != nil {
		r.ActionTimeout.Stop()
		r.ActionTimeout = nil
	}

	return &highestAction
}

// ClearPendingActions 清空所有待處理的動作
func (r *Room) ClearPendingActions() {
	r.ActionMutex.Lock()
	defer r.ActionMutex.Unlock()

	r.PendingActions = []PendingAction{}
	r.IsWaitingForActions = false

	if r.ActionTimeout != nil {
		r.ActionTimeout.Stop()
		r.ActionTimeout = nil
	}

	log.Println("清空待處理動作列表")
}

// StartActionCollection 開始收集玩家動作（設定超時）
func (r *Room) StartActionCollection(timeoutCallback func()) {
	r.ActionMutex.Lock()
	defer r.ActionMutex.Unlock()

	r.IsWaitingForActions = true
	r.PendingActions = []PendingAction{}

	// 設定 3 秒超時
	r.ActionTimeout = time.AfterFunc(3*time.Second, func() {
		log.Println("動作收集超時，處理待處理動作")
		timeoutCallback()
	})

	log.Println("開始收集玩家動作（3 秒超時）")
}

// GetPendingActionCount 獲取待處理動作數量
func (r *Room) GetPendingActionCount() int {
	r.ActionMutex.Lock()
	defer r.ActionMutex.Unlock()
	return len(r.PendingActions)
}

// NextRound 為下一局遊戲準備房間
func (r *Room) NextRound() {
	if r.Game == nil {
		// 如果沒有遊戲，則從頭開始一個新遊戲
		r.StartGame()
		return
	}

	// 將莊家推進到下一位玩家
	currentDealer := r.Game.Dealer
	nextDealer := (currentDealer + 1) % len(r.Players)

	// 重置玩家手牌、面子和花牌
	for _, p := range r.Players {
		p.Hand = make([]string, 0, 17)
		p.Melds = make([]model.Meld, 0)
		p.Flowers = make([]string, 0)
		p.IsTing = false
		p.WinningTiles = nil
	}

	// 使用相同的玩家和新莊家建立一個新遊戲
	newGame := NewMahjongGame(r.Players)
	newGame.Dealer = nextDealer
	r.Game = newGame

	r.GameStarted = true
	r.CurrentTurn = r.Game.Dealer // 新莊家開始
	r.LastDiscardPlayer = -1
	r.LastDiscardTile = ""

	log.Printf("房間 %s 開始新的一局，莊家是: %s (位置 %d)", r.ID, r.Players[r.Game.Dealer].Name, r.Game.Dealer)
}