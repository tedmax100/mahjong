package game

import (
	"encoding/json"
	"errors"
	"log"
	"reflect"
	"sort"
)

// Room 代表一个游戏房间
type Room struct {
	ID              string
	Players         []*Player
	Clients         map[string]interface{} // 存储WebSocket客户端
	Game            *MahjongGame
	GameStarted     bool
	CurrentTurn     int // 当前轮到谁（0-3）
	LastDiscardPlayer int // 最後打牌的玩家位置（用於檢查吃牌資格）
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
func (r *Room) HandleDiscard(userID, tile string) {
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
		return
	}

	// 检查是否轮到该玩家
	if player.Position != r.CurrentTurn {
		log.Printf("还没轮到玩家 %s（当前回合: %d，玩家位置: %d）", player.Name, r.CurrentTurn, player.Position)
		return
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
			return
		}
	}

	// 記錄最後打牌的玩家（用於檢查吃牌資格）
	r.LastDiscardPlayer = player.Position

	// 切换到下一个玩家
	r.NextTurn()
	log.Printf("轮到下一位玩家（位置: %d）", r.CurrentTurn)
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
func (r *Room) HandleKong(userID, tile string, isConcealed bool) bool {
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
			return false
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
			return false
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
		drawnTile := r.Game.DrawTileFromEnd()
		if drawnTile != "" {
			player.Hand = append(player.Hand, drawnTile)
			log.Printf("玩家 %s 杠牌后补牌: %s", player.Name, drawnTile)
		}
	}

	// 杠牌后轮到该玩家出牌
	r.CurrentTurn = player.Position
	log.Printf("杠牌成功，轮到玩家 %s 出牌", player.Name)

	return true
}

// HandleHu 处理胡牌
func (r *Room) HandleHu(userID string, isSelfDrawn bool) *WinResult {
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

	// 检查是否可以胡牌
	if !r.Game.CanHu(player.Hand, player.Melds) {
		log.Printf("玩家 %s 无法胡牌", player.Name)
		return nil
	}

	log.Printf("玩家 %s 胡牌成功！", player.Name)

	// 计算台数和得分
	lastTile := ""
	if len(player.Hand) > 0 {
		lastTile = player.Hand[len(player.Hand)-1]
	}

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
