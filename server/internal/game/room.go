package game

import (
	"encoding/json"
	"errors"
	"log"
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

// Room 代表一个游戏房间
type Room struct {
	ID                string
	Players           []*Player
	Clients           map[string]interface{} // 存储WebSocket客户端
	Game              *MahjongGame
	GameStarted       bool
	CurrentTurn       int // 当前轮到谁（0-3）
	LastDiscardPlayer int // 最後打牌的玩家位置（用於檢查吃牌資格）
	LastDiscardTile   string // 最後打出的牌
	PendingActions    []PendingAction // 待處理的動作請求
		ActionTimeout       *time.Timer // 動作超時計時器
		ActionMutex       sync.Mutex // 保護 PendingActions 的互斥鎖
		IsWaitingForActions bool // 是否正在等待玩家動作回應
		noResponseTimer *time.Timer
		timerMu         sync.Mutex
	}
	
	// Player 代表一个玩家
	type Player struct {
		ID       string
		Name     string
		Position int // 0=东, 1=南, 2=西, 3=北
		Hand     []string
		Score    int
		Melds    []Meld   // 已展示的牌组（碰、杠等）
		Flowers  []string // 花牌
		IsTing   bool     // 玩家是否已听牌
		WinningTiles []string // 听牌所胡的牌
	}
	
	// StartNoResponseTimer starts a timer that fires if no player action is received.
	func (r *Room) StartNoResponseTimer(d time.Duration, cb func()) {
		r.timerMu.Lock()
		defer r.timerMu.Unlock()
		// Stop any existing timer
		if r.noResponseTimer != nil {
			r.noResponseTimer.Stop()
		}
		r.noResponseTimer = time.AfterFunc(d, cb)
	}
	
	// StopNoResponseTimer stops the no-response timer.
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

	log.Printf("📋 %s玩家 %s (位置%d)", actionPrefix, player.Name, player.Position)
	log.Printf("   手牌 (%d張): %s", len(sortedHand), formatTiles(sortedHand))
	if len(melds) > 0 {
		log.Printf("   吃碰槓: %v", melds)
	}
	log.Printf("   總牌數: %d張", player.GetTotalTiles())
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

// formatTiles 格式化牌列表為字符串
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

// NewRoom 创建新房间
func NewRoom(id string) *Room {
	return &Room{
		ID:          id,
		Players:     make([]*Player, 0, 4),
		Clients:     make(map[string]interface{}),
		GameStarted: false,
	}
}

// AddPlayer 添加玩家到房间
func (r *Room) AddPlayer(userID, userName string) error {
	if len(r.Players) >= 4 {
		return errors.New("房间已满")
	}

	player := &Player{
		ID:       userID,
		Name:     userName,
		Position: len(r.Players),
		Hand:     make([]string, 0, 17),
		Score:    1000,
		Melds:    make([]Meld, 0),
		Flowers:  make([]string, 0),
	}

	r.Players = append(r.Players, player)
	log.Printf("玩家 %s 加入房间 %s (位置: %d)", userName, r.ID, player.Position)

	return nil
}

// RemovePlayer 从房间移除玩家
func (r *Room) RemovePlayer(userID string) {
	for i, player := range r.Players {
		if player.ID == userID {
			r.Players = append(r.Players[:i], r.Players[i+1:]...)
			log.Printf("玩家 %s 离开房间 %s", player.Name, r.ID)
			return
		}
	}
}

// StartGame 开始游戏
func (r *Room) StartGame() {
	r.Game = NewMahjongGame(r.Players)
	r.GameStarted = true
	r.CurrentTurn = 0 // 庄家（位置0）先出牌
	log.Printf("房间 %s 游戏开始，庄家先出牌", r.ID)
}

// DealTiles 发牌
func (r *Room) DealTiles() {
	if r.Game != nil {
		r.Game.DealTiles()

		// 📋 記錄所有玩家的初始手牌
		LogAllPlayersHands(r, "遊戲開始 - 所有玩家初始手牌")
	}
}

// GetRoomUpdateMessage 获取房间更新消息
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

// GetGameStartMessage 获取游戏开始消息
func (r *Room) GetGameStartMessage() []byte {
	message := map[string]interface{}{
		"type": "game_start",
		"data": map[string]interface{}{
			"roomId":      r.ID,
			"currentTurn": r.CurrentTurn,
		},
	}

	data, _ := json.Marshal(message)
	return data
}

// GetDealTilesMessage 获取发牌消息
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

// HandleDiscard 处理出牌
func (r *Room) HandleDiscard(userID, tile string) bool {
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

	// 检查是否轮到该玩家
	if player.Position != r.CurrentTurn {
		log.Printf("还没轮到玩家 %s（当前回合: %d，玩家位置: %d）", player.Name, r.CurrentTurn, player.Position)
		return false
	}

	// 检查是否正在等待动作回应（如胡、碰、杠等）
	if r.IsWaitingForActions {
		log.Printf("❌ 玩家 %s 无法打牌：正在等待动作回应（请选择胡/碰/杠/过）", player.Name)
		return false
	}

	// 检查手牌数量是否正确（应该有17张牌才能打牌，槓後補牌可能是18張）
	// 每個槓（kong）比普通碰多1張牌，所以需要調整預期值
	totalTiles := player.GetTotalTiles()

	// 計算槓的數量（每個槓多1張牌）
	kongCount := 0
	for _, meld := range player.Melds {
		if meld.Type == "kong_concealed" || meld.Type == "kong_promoted" || meld.Type == "kong_exposed" {
			kongCount++
		}
	}

	// 基本預期：17或18張
	// 每個槓額外增加1張
	expectedMin := 17 + kongCount
	expectedMax := 18 + kongCount

	if totalTiles != expectedMin && totalTiles != expectedMax {
		log.Printf("❌ 玩家 %s 手牌数量错误！总牌数 %d（手牌 %d + 吃碰槓 %d 組，其中 %d 個槓），预期 %d或%d 张，拒绝打牌",
			player.Name, totalTiles, len(player.Hand), len(player.Melds), kongCount, expectedMin, expectedMax)
		return false
	}

	log.Printf("玩家 %s 打出 %s", player.Name, tile)

	// 从手牌中移除
	for i, t := range player.Hand {
		if t == tile {
			player.Hand = append(player.Hand[:i], player.Hand[i+1:]...)
			break
		}
	}

	// 将牌加入弃牌堆
	if r.Game != nil {
		r.Game.DiscardPile = append(r.Game.DiscardPile, tile)
		log.Printf("弃牌堆: %v (共 %d 张)", r.Game.DiscardPile, len(r.Game.DiscardPile))

		// 检查流局
		if r.Game.CheckDraw() {
			log.Printf("流局！牌山剩余 %d 张", r.Game.GetRemainingTiles())
			r.GameStarted = false
			return true // 返回 true 表示流局
		}
	}

	// 記錄最後打牌的玩家和牌（用於檢查吃牌資格）
	r.LastDiscardPlayer = player.Position
	r.LastDiscardTile = tile

	// 📋 記錄打牌後的手牌狀態
	LogPlayerHand(player, "打牌: "+tile)

	// 切换到下一个玩家
	// TODO: 當實作完整優先權處理後，這裡應該等待其他玩家響應後再切換
	r.NextTurn()
	log.Printf("轮到下一位玩家（位置: %d）", r.CurrentTurn)

	return false // 返回 false 表示没有流局
}

// NextTurn 切换到下一个玩家
func (r *Room) NextTurn() {
	r.CurrentTurn = (r.CurrentTurn + 1) % 4
}

// HandlePong 处理碰牌
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

	// 检查是否可以碰（需要手牌中有至少2张相同的牌）
	if r.Game == nil || !r.Game.CanPong(player.Hand, tile) {
		log.Printf("玩家 %s 无法碰 %s", player.Name, tile)
		return false
	}

	log.Printf("玩家 %s 碰 %s", player.Name, tile)

	// 从手牌中移除2张
	removed := 0
	for i := len(player.Hand) - 1; i >= 0 && removed < 2; i-- {
		if player.Hand[i] == tile {
			player.Hand = append(player.Hand[:i], player.Hand[i+1:]...)
			removed++
		}
	}

	// 从弃牌堆中移除最后一张（即被碰的牌）
	if len(r.Game.DiscardPile) > 0 {
		r.Game.DiscardPile = r.Game.DiscardPile[:len(r.Game.DiscardPile)-1]
	}

	// 添加到已展示的牌组
	player.Melds = append(player.Melds, Meld{
		Type:  "pong",
		Tiles: []string{tile, tile, tile},
	})

	// 碰牌后轮到该玩家出牌
	r.CurrentTurn = player.Position
	log.Printf("碰牌成功，轮到玩家 %s 出牌", player.Name)

	// 📋 記錄碰牌後的手牌狀態
	LogPlayerHand(player, "碰牌: "+tile)

	return true
}

// HandleChow 处理吃牌（只能吃上家打出的牌）
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

	// 检查是否是上家打出的牌
	// 台湾麻将规则：只能吃上家的牌
	previousPlayer := (player.Position + 3) % 4 // 上家位置
	if r.LastDiscardPlayer != previousPlayer {
		log.Printf("玩家 %s 只能吃上家的牌（最後出牌者位置: %d，上家位置: %d）",
			player.Name, r.LastDiscardPlayer, previousPlayer)
		return false
	}

	// 验证吃牌组合是否有效
	validCombinations := r.Game.CanChow(player.Hand, tile)
	if len(validCombinations) == 0 {
		log.Printf("玩家 %s 无法吃 %s", player.Name, tile)
		return false
	}

	// 检查提供的吃牌组合是否在有效组合中
	isValidCombination := false
	for _, combo := range validCombinations {
		if len(combo) == len(chowTiles) && isSameCombination(combo, chowTiles) {
			isValidCombination = true
			break
		}
	}

	if !isValidCombination {
		log.Printf("玩家 %s 提供的吃牌组合无效: %v", player.Name, chowTiles)
		return false
	}

	log.Printf("玩家 %s 吃 %s，组合: %v", player.Name, tile, chowTiles)

	// 从手牌中移除需要的牌（除了吃的那张）
	for _, chowTile := range chowTiles {
		if chowTile == tile {
			continue // 跳过要吃的牌（从弃牌堆拿）
		}

		// 从手牌中移除
		for i, t := range player.Hand {
			if t == chowTile {
				player.Hand = append(player.Hand[:i], player.Hand[i+1:]...)
				break
			}
		}
	}

	// 从弃牌堆中移除最后一张（即被吃的牌）
	if len(r.Game.DiscardPile) > 0 {
		r.Game.DiscardPile = r.Game.DiscardPile[:len(r.Game.DiscardPile)-1]
	}

	// 添加到已展示的牌组
	player.Melds = append(player.Melds, Meld{
		Type:  "chow",
		Tiles: chowTiles,
	})

	// 吃牌后轮到该玩家出牌
	r.CurrentTurn = player.Position
	log.Printf("吃牌成功，轮到玩家 %s 出牌", player.Name)

	// 📋 記錄吃牌後的手牌狀態
	LogPlayerHand(player, "吃牌: "+formatTiles(chowTiles))

	return true
}

// isSameCombination 检查两个牌组是否相同（忽略顺序）
func isSameCombination(combo1, combo2 []string) bool {
	if len(combo1) != len(combo2) {
		return false
	}
	// 创建副本并排序
	c1 := make([]string, len(combo1))
	copy(c1, combo1)
	sort.Strings(c1)

	c2 := make([]string, len(combo2))
	copy(c2, combo2)
	sort.Strings(c2)

	return reflect.DeepEqual(c1, c2)
}

// HandleKong 处理杠牌
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
		// 暗杠：手牌中有4张相同的牌
		count := 0
		for _, t := range player.Hand {
			if t == tile {
				count++
			}
		}

		if count < 4 {
			log.Printf("玩家 %s 无法暗杠 %s（手牌不足4张）", player.Name, tile)
			return false, ""
		}

		log.Printf("玩家 %s 暗杠 %s", player.Name, tile)

		// 从手牌中移除4张
		removed := 0
		for i := len(player.Hand) - 1; i >= 0 && removed < 4; i-- {
			if player.Hand[i] == tile {
				player.Hand = append(player.Hand[:i], player.Hand[i+1:]...)
				removed++
			}
		}

		// 添加到已展示的牌组
		player.Melds = append(player.Melds, Meld{
			Type:  "kong_concealed",
			Tiles: []string{tile, tile, tile, tile},
		})

	} else {
		// 明杠：手牌中有3张，或已碰出，加上别人打出的1张
		if !r.Game.CanExposedKong(player, tile) {
			log.Printf("玩家 %s 无法明杠 %s", player.Name, tile)
			return false, ""
		}

		// 判断是加杠还是明杠
		isPromotedKong := false
		for i, meld := range player.Melds {
			if meld.Type == "pong" && meld.Tiles[0] == tile {
				log.Printf("玩家 %s 加杠 %s", player.Name, tile)
				player.Melds[i].Type = "kong_promoted"
				player.Melds[i].Tiles = append(player.Melds[i].Tiles, tile)
				isPromotedKong = true
				break
			}
		}

		if !isPromotedKong {
			log.Printf("玩家 %s 明杠 %s", player.Name, tile)
			// 从手牌中移除3张
			removed := 0
			for i := len(player.Hand) - 1; i >= 0 && removed < 3; i-- {
				if player.Hand[i] == tile {
					player.Hand = append(player.Hand[:i], player.Hand[i+1:]...)
					removed++
				}
			}
			// 添加到已展示的牌组
			player.Melds = append(player.Melds, Meld{
				Type:  "kong_exposed",
				Tiles: []string{tile, tile, tile, tile},
			})
		}

		// 从弃牌堆中移除最后一张
		if len(r.Game.DiscardPile) > 0 {
			r.Game.DiscardPile = r.Game.DiscardPile[:len(r.Game.DiscardPile)-1]
		}
	}

	// 杠牌后需要补牌
	if len(r.Game.Deck) > 0 {
		// 杠牌补牌需要从牌尾拿
		drawnTile = r.Game.DrawTileFromEnd()
		if drawnTile != "" {
			player.Hand = append(player.Hand, drawnTile)
			log.Printf("玩家 %s 杠牌后补牌: %s", player.Name, drawnTile)
		}
	}

	// 杠牌后轮到该玩家出牌
	r.CurrentTurn = player.Position
	log.Printf("杠牌成功，轮到玩家 %s 出牌", player.Name)

	// 📋 記錄槓牌後的手牌狀態
	kongType := "kong"
	if isConcealed {
		kongType = "kong_concealed"
	}
	LogPlayerHand(player, "槓牌: "+tile+" ("+kongType+")")

	return true, drawnTile
}

// HandleHu 处理胡牌
func (r *Room) HandleHu(userID string, winTile string, isSelfDrawn bool) *WinResult {
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

	// 检查是否可以胡牌（将胡牌的牌加入手牌进行检查）
	tempHand := append([]string{}, player.Hand...)
	tempHand = append(tempHand, winTile)
	if !r.Game.CanHu(tempHand, player.Melds) {
		log.Printf("玩家 %s 无法胡牌（手牌+%s）", player.Name, winTile)
		return nil
	}

	log.Printf("玩家 %s 胡牌成功！胡牌: %s", player.Name, winTile)

	// 使用胡牌的牌作为 lastTile
	lastTile := winTile

	winResult := r.Game.CalculateScore(player, lastTile, isSelfDrawn)

	// 更新玩家分数
	if isSelfDrawn {
		// 自摸：其他三家各付分数
		for _, p := range r.Players {
			if p.ID != player.ID {
				p.Score -= winResult.BaseScore
				player.Score += winResult.BaseScore
			}
		}
	} else {
		// 放炮：放炮者付全部分数
		// TODO: 需要记录是谁放炮
		// 暂时简化处理
	}

	// 标记游戏结束
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
		return r.Game.CanExposedKong(player, tile)
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

	// 設定3秒超時
	r.ActionTimeout = time.AfterFunc(3*time.Second, func() {
		log.Println("動作收集超時，處理待處理動作")
		timeoutCallback()
	})

	log.Println("開始收集玩家動作（3秒超時）")
}

// GetPendingActionCount 獲取待處理動作數量
func (r *Room) GetPendingActionCount() int {
	r.ActionMutex.Lock()
	defer r.ActionMutex.Unlock()
	return len(r.PendingActions)
}

// NextRound prepares the room for the next round of the game.
func (r *Room) NextRound() {
	if r.Game == nil {
		// If there's no game, start a new one from scratch
		r.StartGame()
		return
	}

	// Advance the dealer to the next player
	currentDealer := r.Game.Dealer
	nextDealer := (currentDealer + 1) % len(r.Players)

	// Reset player hands, melds, and flowers
	for _, p := range r.Players {
		p.Hand = make([]string, 0, 17)
		p.Melds = make([]Meld, 0)
		p.Flowers = make([]string, 0)
		p.IsTing = false
		p.WinningTiles = nil
	}

	// Create a new game with the same players and the new dealer
	newGame := NewMahjongGame(r.Players)
	newGame.Dealer = nextDealer
	r.Game = newGame

	r.GameStarted = true
	r.CurrentTurn = r.Game.Dealer // The new dealer starts
	r.LastDiscardPlayer = -1
	r.LastDiscardTile = ""

	log.Printf("房间 %s 开始新的一局，庄家是: %s (位置 %d)", r.ID, r.Players[r.Game.Dealer].Name, r.Game.Dealer)
}
