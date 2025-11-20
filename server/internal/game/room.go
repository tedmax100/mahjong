package game

import (
	"encoding/json"
	"errors"
	"log"
)

// Room 代表一个游戏房间
type Room struct {
	ID          string
	Players     []*Player
	Clients     map[string]interface{} // 存储WebSocket客户端
	Game        *MahjongGame
	GameStarted bool
	CurrentTurn int // 当前轮到谁（0-3）
}

// Player 代表一个玩家
type Player struct {
	ID       string
	Name     string
	Position int // 0=东, 1=南, 2=西, 3=北
	Hand     []string
	Score    int
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

	// 切换到下一个玩家
	r.NextTurn()
	log.Printf("轮到下一位玩家（位置: %d）", r.CurrentTurn)
}

// NextTurn 切换到下一个玩家
func (r *Room) NextTurn() {
	r.CurrentTurn = (r.CurrentTurn + 1) % 4
}

// HandlePong 处理碰牌
func (r *Room) HandlePong(userID, tile string) {
	log.Printf("玩家 %s 碰 %s", userID, tile)
	// TODO: 实现碰牌逻辑
}

// HandleKong 处理杠牌
func (r *Room) HandleKong(userID, tile string) {
	log.Printf("玩家 %s 杠 %s", userID, tile)
	// TODO: 实现杠牌逻辑
}

// HandleHu 处理胡牌
func (r *Room) HandleHu(userID string) {
	log.Printf("玩家 %s 胡牌", userID)
	// TODO: 实现胡牌逻辑
}
