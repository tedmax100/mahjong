package game

import (
	"testing"
)

// TestHandleDiscardReturnsDrawStatus 测试 HandleDiscard 在流局时返回 true
func TestHandleDiscardReturnsDrawStatus(t *testing.T) {
	room := NewRoom("test-draw-room")

	// 添加玩家
	for i := 1; i <= 4; i++ {
		err := room.AddPlayer("player"+string(rune('0'+i)), "玩家"+string(rune('0'+i)))
		if err != nil {
			t.Fatalf("添加玩家失败: %v", err)
		}
	}

	room.StartGame()
	room.DealTiles()

	t.Run("正常情况下 HandleDiscard 返回 false", func(t *testing.T) {
		player := room.Players[0]
		room.CurrentTurn = 0

		// 确保有足够的牌
		if room.Game.GetRemainingTiles() <= 8 {
			t.Skip("牌山剩余不足，跳过此测试")
		}

		// 给玩家一些手牌
		player.Hand = []string{
			"wan-1", "wan-1", "wan-1", "wan-2", "wan-2", "wan-2", "wan-3", "wan-3", "wan-3",
			"wan-4", "wan-4", "wan-4", "wan-5", "wan-5", "wan-5", "wan-6", "wan-7",
		}

		isDraw := room.HandleDiscard(player.ID, "wan-1")
		if isDraw {
			t.Error("正常情况下 HandleDiscard 应该返回 false")
		}
	})

	t.Run("流局时 HandleDiscard 返回 true", func(t *testing.T) {
		// 创建一个新房间来测试流局
		drawRoom := NewRoom("test-draw")
		for i := 1; i <= 4; i++ {
			drawRoom.AddPlayer("player"+string(rune('0'+i)), "玩家"+string(rune('0'+i)))
		}
		drawRoom.StartGame()

		// 手动设置牌山只剩8张或更少
		drawRoom.Game.Deck = make([]string, 8)
		for i := 0; i < 8; i++ {
			drawRoom.Game.Deck[i] = "wan-1"
		}

		player := drawRoom.Players[0]
		drawRoom.CurrentTurn = 0
		player.Hand = []string{
			"wan-1", "wan-1", "wan-1", "wan-2", "wan-2", "wan-2", "wan-3", "wan-3", "wan-3",
			"wan-4", "wan-4", "wan-4", "wan-5", "wan-5", "wan-5", "wan-6", "wan-7",
		}

		// 确认牌山剩余 <= 8
		if drawRoom.Game.GetRemainingTiles() > 8 {
			t.Fatalf("预期牌山剩余 <= 8，实际 %d", drawRoom.Game.GetRemainingTiles())
		}

		// 执行出牌，应该触发流局
		isDraw := drawRoom.HandleDiscard(player.ID, "wan-1")
		if !isDraw {
			t.Error("牌山剩余 <= 8 时，HandleDiscard 应该返回 true 表示流局")
		}

		// 确认游戏已停止
		if drawRoom.GameStarted {
			t.Error("流局后 GameStarted 应该为 false")
		}
	})
}

// TestCheckDrawCondition 测试流局条件判断
func TestCheckDrawCondition(t *testing.T) {
	t.Run("牌山剩余8张时应该流局", func(t *testing.T) {
		game := &MahjongGame{
			Deck: make([]string, 8),
		}

		if !game.CheckDraw() {
			t.Error("牌山剩余8张时应该判定为流局")
		}
	})

	t.Run("牌山剩余7张时应该流局", func(t *testing.T) {
		game := &MahjongGame{
			Deck: make([]string, 7),
		}

		if !game.CheckDraw() {
			t.Error("牌山剩余7张时应该判定为流局")
		}
	})

	t.Run("牌山剩余0张时应该流局", func(t *testing.T) {
		game := &MahjongGame{
			Deck: make([]string, 0),
		}

		if !game.CheckDraw() {
			t.Error("牌山剩余0张时应该判定为流局")
		}
	})

	t.Run("牌山剩余9张时不应该流局", func(t *testing.T) {
		game := &MahjongGame{
			Deck: make([]string, 9),
		}

		if game.CheckDraw() {
			t.Error("牌山剩余9张时不应该判定为流局")
		}
	})

	t.Run("牌山剩余50张时不应该流局", func(t *testing.T) {
		game := &MahjongGame{
			Deck: make([]string, 50),
		}

		if game.CheckDraw() {
			t.Error("牌山剩余50张时不应该判定为流局")
		}
	})
}

// TestGetRemainingTiles 测试获取剩余牌数
func TestGetRemainingTiles(t *testing.T) {
	t.Run("初始状态应该有144张牌", func(t *testing.T) {
		// 创建测试玩家
		players := make([]*Player, 4)
		for i := 0; i < 4; i++ {
			players[i] = &Player{
				ID:       "player" + string(rune('1'+i)),
				Name:     "玩家" + string(rune('1'+i)),
				Position: i,
			}
		}

		game := NewMahjongGame(players)
		remaining := game.GetRemainingTiles()

		// 初始牌山应该有144张牌
		if remaining != 144 {
			t.Errorf("初始牌山应该有144张牌，实际 %d 张", remaining)
		}
	})

	t.Run("摸牌后剩余牌数应该减少", func(t *testing.T) {
		// 创建测试玩家
		players := make([]*Player, 4)
		for i := 0; i < 4; i++ {
			players[i] = &Player{
				ID:       "player" + string(rune('1'+i)),
				Name:     "玩家" + string(rune('1'+i)),
				Position: i,
			}
		}

		game := NewMahjongGame(players)
		initialCount := game.GetRemainingTiles()

		// 摸一张牌
		game.DrawTile()
		newCount := game.GetRemainingTiles()

		if newCount != initialCount-1 {
			t.Errorf("摸牌后剩余牌数应该减1，期望 %d，实际 %d", initialCount-1, newCount)
		}
	})

	t.Run("空牌山应该返回0", func(t *testing.T) {
		game := &MahjongGame{
			Deck: make([]string, 0),
		}

		remaining := game.GetRemainingTiles()
		if remaining != 0 {
			t.Errorf("空牌山应该返回0，实际 %d", remaining)
		}
	})
}

// TestDrawLogicWithRemainingTiles 测试摸牌时的剩余牌数逻辑
func TestDrawLogicWithRemainingTiles(t *testing.T) {
	room := NewRoom("test-room")

	// 添加玩家
	for i := 1; i <= 4; i++ {
		room.AddPlayer("player"+string(rune('0'+i)), "玩家"+string(rune('0'+i)))
	}

	room.StartGame()
	room.DealTiles()

	t.Run("摸牌后剩余牌数应该正确更新", func(t *testing.T) {
		initialRemaining := room.Game.GetRemainingTiles()

		// 摸一张牌
		tile := room.Game.DrawTile()
		if tile == "" {
			t.Skip("牌山已空，跳过测试")
		}

		newRemaining := room.Game.GetRemainingTiles()
		expectedRemaining := initialRemaining - 1

		if newRemaining != expectedRemaining {
			t.Errorf("摸牌后剩余牌数应该是 %d，实际 %d", expectedRemaining, newRemaining)
		}
	})

	t.Run("摸牌直到接近流局边界", func(t *testing.T) {
		// 创建新房间
		drawRoom := NewRoom("draw-boundary")
		for i := 1; i <= 4; i++ {
			drawRoom.AddPlayer("player"+string(rune('0'+i)), "玩家"+string(rune('0'+i)))
		}
		drawRoom.StartGame()

		// 设置牌山剩余10张
		drawRoom.Game.Deck = make([]string, 10)
		for i := 0; i < 10; i++ {
			drawRoom.Game.Deck[i] = "wan-1"
		}

		// 摸2张牌，剩余8张
		drawRoom.Game.DrawTile()
		drawRoom.Game.DrawTile()

		remaining := drawRoom.Game.GetRemainingTiles()
		if remaining != 8 {
			t.Fatalf("预期剩余8张，实际 %d 张", remaining)
		}

		// 此时应该满足流局条件
		if !drawRoom.Game.CheckDraw() {
			t.Error("剩余8张时应该判定为流局")
		}
	})
}

// TestGameStateAfterDraw 测试流局后的游戏状态
func TestGameStateAfterDraw(t *testing.T) {
	room := NewRoom("test-draw-state")

	// 添加玩家
	for i := 1; i <= 4; i++ {
		room.AddPlayer("player"+string(rune('0'+i)), "玩家"+string(rune('0'+i)))
	}

	room.StartGame()

	// 设置牌山剩余8张
	room.Game.Deck = make([]string, 8)
	for i := 0; i < 8; i++ {
		room.Game.Deck[i] = "wan-1"
	}

	player := room.Players[0]
	room.CurrentTurn = 0
	player.Hand = []string{
		"wan-1", "wan-1", "wan-1", "wan-2", "wan-2", "wan-2", "wan-3", "wan-3", "wan-3",
		"wan-4", "wan-4", "wan-4", "wan-5", "wan-5", "wan-5", "wan-6", "wan-7",
	}

	t.Run("流局后游戏应该停止", func(t *testing.T) {
		// 触发流局
		isDraw := room.HandleDiscard(player.ID, "wan-1")

		if !isDraw {
			t.Fatal("应该触发流局")
		}

		// 游戏应该停止
		if room.GameStarted {
			t.Error("流局后 GameStarted 应该为 false")
		}
	})

	t.Run("流局后弃牌堆应该包含打出的牌", func(t *testing.T) {
		// 创建新房间
		newRoom := NewRoom("test-discard-pile")
		for i := 1; i <= 4; i++ {
			newRoom.AddPlayer("player"+string(rune('0'+i)), "玩家"+string(rune('0'+i)))
		}
		newRoom.StartGame()

		// 设置牌山剩余8张
		newRoom.Game.Deck = make([]string, 8)
		newRoom.Game.DiscardPile = []string{} // 清空弃牌堆

		p := newRoom.Players[0]
		newRoom.CurrentTurn = 0
		p.Hand = []string{
			"wan-1", "wan-1", "wan-1", "wan-2", "wan-2", "wan-2", "wan-3", "wan-3", "wan-3",
			"wan-4", "wan-4", "wan-4", "wan-5", "wan-5", "wan-5", "wan-6", "wan-7",
		}

		// 打出一张牌触发流局
		newRoom.HandleDiscard(p.ID, "wan-1")

		// 弃牌堆应该包含打出的牌
		if len(newRoom.Game.DiscardPile) != 1 {
			t.Errorf("弃牌堆应该有1张牌，实际 %d 张", len(newRoom.Game.DiscardPile))
		}

		if newRoom.Game.DiscardPile[0] != "wan-1" {
			t.Errorf("弃牌堆的牌应该是 wan-1，实际 %s", newRoom.Game.DiscardPile[0])
		}
	})
}
