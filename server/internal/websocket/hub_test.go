package websocket

import (
	"mahjong/internal/game"
	"mahjong/internal/model"
	"testing"
)

// TestDrawForRealPlayer 测试真实玩家自动摸牌功能
func TestDrawForRealPlayer(t *testing.T) {
	// 1. 创建测试用的 Hub
	hub := NewHub()

	// 2. 创建测试房间和玩家
	realPlayer := &game.Player{
		ID:       "player_123",
		Name:     "TestPlayer",
		Position: 0,
		Hand:     make([]string, 16), // 正常轮次（16张手牌）
		Melds:    []model.Meld{},
	}

	// 填充手牌（16张）
	for i := 0; i < 16; i++ {
		realPlayer.Hand[i] = "wan-1"
	}

	bot := &game.Player{
		ID:       "bot_456",
		Name:     "TestBot",
		Position: 1,
		Hand:     []string{},
		Melds:    []model.Meld{},
	}

	room := &game.Room{
		ID:          "test-room",
		Players:     []*game.Player{realPlayer, bot},
		Game:        game.NewMahjongGame([]*game.Player{realPlayer, bot}),
		GameStarted: true,
		CurrentTurn: 0, // 轮到真实玩家
		Clients:     make(map[string]interface{}),
	}

	// Force deck to have no flowers at the top to avoid flaky test
	// Move any flowers to the end of the deck
	nonFlowers := []string{}
	flowers := []string{}
	for _, t := range room.Game.Deck {
		if len(t) > 6 && t[:6] == "flower" {
			flowers = append(flowers, t)
		} else {
			nonFlowers = append(nonFlowers, t)
		}
	}
	room.Game.Deck = append(nonFlowers, flowers...)

	hub.rooms["test-room"] = room

	// 记录初始状态
	initialHandSize := len(realPlayer.Hand)
	initialDeckSize := len(room.Game.Deck)

	// 3. 执行 drawForRealPlayer_needsLock (with lock)
	hub.mu.Lock()
	hub.drawForRealPlayer_needsLock(room)
	hub.mu.Unlock()

	// 4. 验证结果
	// 验证手牌数量增加了1
	if len(realPlayer.Hand) != initialHandSize+1 {
		t.Errorf("真实玩家摸牌后手牌数应该是 %d，但实际是 %d", initialHandSize+1, len(realPlayer.Hand))
	}

	// 验证牌山减少了1
	if len(room.Game.Deck) != initialDeckSize-1 {
		t.Errorf("摸牌后牌山数量应该是 %d，但实际是 %d", initialDeckSize-1, len(room.Game.Deck))
	}

	// 验证手牌应该有17张
	if len(realPlayer.Hand) != 17 {
		t.Errorf("摸牌后真实玩家应该有17张手牌，但实际有 %d 张", len(realPlayer.Hand))
	}
}

// TestDrawForRealPlayer_BotShouldNotDraw 测试机器人不应该通过 DrawForRealPlayer 摸牌
func TestDrawForRealPlayer_BotShouldNotDraw(t *testing.T) {
	hub := NewHub()

	bot := &game.Player{
		ID:       "bot_123",
		Name:     "TestBot",
		Position: 0,
		Hand:     make([]string, 16),
		Melds:    []model.Meld{},
	}

	for i := 0; i < 16; i++ {
		bot.Hand[i] = "wan-1"
	}

	room := &game.Room{
		ID:          "test-room",
		Players:     []*game.Player{bot},
		Game:        game.NewMahjongGame([]*game.Player{bot}),
		GameStarted: true,
		CurrentTurn: 0,
		Clients:     make(map[string]interface{}),
	}

	hub.rooms["test-room"] = room
	initialHandSize := len(bot.Hand)

	// 执行 drawForRealPlayer_needsLock
	hub.mu.Lock()
	hub.drawForRealPlayer_needsLock(room)
	hub.mu.Unlock()

	// 验证机器人手牌数量没有变化
	if len(bot.Hand) != initialHandSize {
		t.Errorf("机器人不应该通过 DrawForRealPlayer 摸牌，手牌数应该保持 %d，但实际是 %d", initialHandSize, len(bot.Hand))
	}
}

// TestDrawForRealPlayer_WrongHandSize 测试手牌不是16张时不应该摸牌
func TestDrawForRealPlayer_WrongHandSize(t *testing.T) {
	hub := NewHub()

	realPlayer := &game.Player{
		ID:       "player_123",
		Name:     "TestPlayer",
		Position: 0,
		Hand:     []string{"wan-1", "wan-2", "wan-3"}, // 只有3张（异常状态）
		Melds:    []model.Meld{},
	}

	room := &game.Room{
		ID:          "test-room",
		Players:     []*game.Player{realPlayer},
		Game:        game.NewMahjongGame([]*game.Player{realPlayer}),
		GameStarted: true,
		CurrentTurn: 0,
		Clients:     make(map[string]interface{}),
	}

	hub.rooms["test-room"] = room
	initialHandSize := len(realPlayer.Hand)

	// 执行 drawForRealPlayer_needsLock
	hub.mu.Lock()
	hub.drawForRealPlayer_needsLock(room)
	hub.mu.Unlock()

	// 验证手牌数量没有变化（因为不是16张）
	if len(realPlayer.Hand) != initialHandSize {
		t.Errorf("手牌不是16张时不应该摸牌，手牌数应该保持 %d，但实际是 %d", initialHandSize, len(realPlayer.Hand))
	}
}

// TestDrawForRealPlayer_GameNotStarted 测试游戏未开始时不应该摸牌
func TestDrawForRealPlayer_GameNotStarted(t *testing.T) {
	hub := NewHub()

	realPlayer := &game.Player{
		ID:       "player_123",
		Name:     "TestPlayer",
		Position: 0,
		Hand:     make([]string, 16),
		Melds:    []model.Meld{},
	}

	for i := 0; i < 16; i++ {
		realPlayer.Hand[i] = "wan-1"
	}

	room := &game.Room{
		ID:          "test-room",
		Players:     []*game.Player{realPlayer},
		Game:        game.NewMahjongGame([]*game.Player{realPlayer}),
		GameStarted: false, // 游戏未开始
		CurrentTurn: 0,
		Clients:     make(map[string]interface{}),
	}

	hub.rooms["test-room"] = room
	initialHandSize := len(realPlayer.Hand)

	// 执行 drawForRealPlayer_needsLock
	hub.mu.Lock()
	hub.drawForRealPlayer_needsLock(room)
	hub.mu.Unlock()

	// 验证手牌数量没有变化
	if len(realPlayer.Hand) != initialHandSize {
		t.Errorf("游戏未开始时不应该摸牌，手牌数應該保持 %d，但實際是 %d", initialHandSize, len(realPlayer.Hand))
	}
}

// TestCheckAndPlayBotTurn_CallsDrawForRealPlayer 测试 CheckAndPlayBotTurn 会为真实玩家调用 DrawForRealPlayer
func TestCheckAndPlayBotTurn_CallsDrawForRealPlayer(t *testing.T) {
	hub := NewHub()

	realPlayer := &game.Player{
		ID:       "player_123",
		Name:     "TestPlayer",
		Position: 0,
		Hand:     make([]string, 16),
		Melds:    []model.Meld{},
	}

	for i := 0; i < 16; i++ {
		realPlayer.Hand[i] = "wan-1"
	}

	room := &game.Room{
		ID:          "test-room",
		Players:     []*game.Player{realPlayer},
		Game:        game.NewMahjongGame([]*game.Player{realPlayer}),
		GameStarted: true,
		CurrentTurn: 0,
		Clients:     make(map[string]interface{}),
	}

	hub.rooms["test-room"] = room
	initialHandSize := len(realPlayer.Hand)

	// 执行 CheckAndPlayBotTurn（应该检测到是真实玩家并调用 DrawForRealPlayer）
	// 注意：withDelay = false 以避免 goroutine 和 sleep
	hub.CheckAndPlayBotTurn(room, false)

	// 验证真实玩家摸到了牌
	if len(realPlayer.Hand) != initialHandSize+1 {
		t.Errorf("CheckAndPlayBotTurn 应该为真实玩家摸牌，手牌数应该是 %d，但实际是 %d", initialHandSize+1, len(realPlayer.Hand))
	}
}
