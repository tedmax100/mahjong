package game

import (
	"testing"
)

// TestHandleDiscardReturnsDrawStatus 測試 HandleDiscard 在流局時返回 true
func TestHandleDiscardReturnsDrawStatus(t *testing.T) {
	room := NewRoom("test-draw-room")

	// 添加玩家
	for i := 1; i <= 4; i++ {
		err := room.AddPlayer("player"+string(rune('0'+i)), "玩家"+string(rune('0'+i)), false)
		if err != nil {
			t.Fatalf("添加玩家失敗: %v", err)
		}
	}

	room.StartGame()
	room.DealTiles()

	t.Run("正常情況下 HandleDiscard 返回 false", func(t *testing.T) {
		player := room.Players[0]
		room.CurrentTurn = 0

		// 確保有足夠的牌
		if room.Game.GetRemainingTiles() <= 8 {
			t.Skip("牌山剩餘不足，跳過此測試")
		}

		// 給玩家一些手牌
		player.Hand = []string{
			"wan-1", "wan-1", "wan-1", "wan-2", "wan-2", "wan-2", "wan-3", "wan-3", "wan-3",
			"wan-4", "wan-4", "wan-4", "wan-5", "wan-5", "wan-5", "wan-6", "wan-7",
		}

		success, isDraw := room.HandleDiscard(player.ID, "wan-1")
		if !success {
			t.Error("HandleDiscard 應該成功")
		}
		if isDraw {
			t.Error("正常情況下 isDraw 應該返回 false")
		}
	})

	t.Run("流局時 HandleDiscard 返回 true", func(t *testing.T) {
		// 建立一個新房間來測試流局
		drawRoom := NewRoom("test-draw")
		for i := 1; i <= 4; i++ {
			drawRoom.AddPlayer("player"+string(rune('0'+i)), "玩家"+string(rune('0'+i)), false)
		}
		drawRoom.StartGame()

		// 手動設置牌山只剩 8 張或更少
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

		// 確認牌山剩餘 <= 8
		if drawRoom.Game.GetRemainingTiles() > 8 {
			t.Fatalf("預期牌山剩餘 <= 8，實際 %d", drawRoom.Game.GetRemainingTiles())
		}

		// 執行出牌，應該觸發流局
		success, isDraw := drawRoom.HandleDiscard(player.ID, "wan-1")
		if !success {
			t.Error("HandleDiscard 應該成功")
		}
		if !isDraw {
			t.Error("牌山剩餘 <= 8 時，isDraw 應該返回 true 表示流局")
		}

		// 確認遊戲已停止
		if drawRoom.GameStarted {
			t.Error("流局後 GameStarted 應該為 false")
		}
	})
}

// TestCheckDrawCondition 測試流局條件判斷
func TestCheckDrawCondition(t *testing.T) {
	t.Run("牌山剩餘 8 張時應該流局", func(t *testing.T) {
		game := &MahjongGame{
			Deck: make([]string, 8),
		}

		if !game.CheckDraw() {
			t.Error("牌山剩餘 8 張時應該判定為流局")
		}
	})

	t.Run("牌山剩餘 7 張時應該流局", func(t *testing.T) {
		game := &MahjongGame{
			Deck: make([]string, 7),
		}

		if !game.CheckDraw() {
			t.Error("牌山剩餘 7 張時應該判定為流局")
		}
	})

	t.Run("牌山剩餘 0 張時應該流局", func(t *testing.T) {
		game := &MahjongGame{
			Deck: make([]string, 0),
		}

		if !game.CheckDraw() {
			t.Error("牌山剩餘 0 張時應該判定為流局")
		}
	})

	t.Run("牌山剩餘 9 張時不應該流局", func(t *testing.T) {
		game := &MahjongGame{
			Deck: make([]string, 9),
		}

		if game.CheckDraw() {
			t.Error("牌山剩餘 9 張時不應該判定為流局")
		}
	})

	t.Run("牌山剩餘 50 張時不應該流局", func(t *testing.T) {
		game := &MahjongGame{
			Deck: make([]string, 50),
		}

		if game.CheckDraw() {
			t.Error("牌山剩餘 50 張時不應該判定為流局")
		}
	})
}

// TestGetRemainingTiles 測試獲取剩餘牌數
func TestGetRemainingTiles(t *testing.T) {
	t.Run("初始狀態應該有 144 張牌", func(t *testing.T) {
		// 建立測試玩家
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

		// 初始牌山應該有 144 張牌
		if remaining != 144 {
			t.Errorf("初始牌山應該有 144 張牌，實際 %d 張", remaining)
		}
	})

	t.Run("摸牌後剩餘牌數應該減少", func(t *testing.T) {
		// 建立測試玩家
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

		// 摸一張牌
		game.DrawTile()
		newCount := game.GetRemainingTiles()

		if newCount != initialCount-1 {
			t.Errorf("摸牌後剩餘牌數應該減 1，期望 %d，實際 %d", initialCount-1, newCount)
		}
	})

	t.Run("空牌山應該返回 0", func(t *testing.T) {
		game := &MahjongGame{
			Deck: make([]string, 0),
		}

		remaining := game.GetRemainingTiles()
		if remaining != 0 {
			t.Errorf("空牌山應該返回 0，實際 %d", remaining)
		}
	})
}

// TestDrawLogicWithRemainingTiles 測試摸牌時的剩餘牌數邏輯
func TestDrawLogicWithRemainingTiles(t *testing.T) {
	room := NewRoom("test-room")

	// 添加玩家
	for i := 1; i <= 4; i++ {
		room.AddPlayer("player"+string(rune('0'+i)), "玩家"+string(rune('0'+i)), false)
	}

	room.StartGame()
	room.DealTiles()

	t.Run("摸牌後剩餘牌數應該正確更新", func(t *testing.T) {
		initialRemaining := room.Game.GetRemainingTiles()

		// 摸一張牌
		tile := room.Game.DrawTile()
		if tile == "" {
			t.Skip("牌山已空，跳過測試")
		}

		newRemaining := room.Game.GetRemainingTiles()
		expectedRemaining := initialRemaining - 1

		if newRemaining != expectedRemaining {
			t.Errorf("摸牌後剩餘牌數應該是 %d，實際 %d", expectedRemaining, newRemaining)
		}
	})

	t.Run("摸牌直到接近流局邊界", func(t *testing.T) {
		// 建立新房間
		drawRoom := NewRoom("draw-boundary")
		for i := 1; i <= 4; i++ {
			drawRoom.AddPlayer("player"+string(rune('0'+i)), "玩家"+string(rune('0'+i)), false)
		}
		drawRoom.StartGame()

		// 設置牌山剩餘 10 張
		drawRoom.Game.Deck = make([]string, 10)
		for i := 0; i < 10; i++ {
			drawRoom.Game.Deck[i] = "wan-1"
		}

		// 摸 2 張牌，剩餘 8 張
		drawRoom.Game.DrawTile()
		drawRoom.Game.DrawTile()

		remaining := drawRoom.Game.GetRemainingTiles()
		if remaining != 8 {
			t.Fatalf("預期剩餘 8 張，實際 %d 張", remaining)
		}

		// 此時應該滿足流局條件
		if !drawRoom.Game.CheckDraw() {
			t.Error("剩餘 8 張時應該判定為流局")
		}
	})
}

// TestGameStateAfterDraw 測試流局後的遊戲狀態
func TestGameStateAfterDraw(t *testing.T) {
	room := NewRoom("test-draw-state")

	// 添加玩家
	for i := 1; i <= 4; i++ {
		room.AddPlayer("player"+string(rune('0'+i)), "玩家"+string(rune('0'+i)), false)
	}

	room.StartGame()

	// 設置牌山剩餘 8 張
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

	t.Run("流局後遊戲應該停止", func(t *testing.T) {
		// 觸發流局
		success, isDraw := room.HandleDiscard(player.ID, "wan-1")

		if !success {
			t.Fatal("HandleDiscard 應該成功")
		}
		if !isDraw {
			t.Fatal("應該觸發流局")
		}

		// 遊戲應該停止
		if room.GameStarted {
			t.Error("流局後 GameStarted 應該為 false")
		}
	})

	t.Run("流局後棄牌堆應該包含打出的牌", func(t *testing.T) {
		// 建立新房間
		newRoom := NewRoom("test-discard-pile")
		for i := 1; i <= 4; i++ {
			newRoom.AddPlayer("player"+string(rune('0'+i)), "玩家"+string(rune('0'+i)), false)
		}
		newRoom.StartGame()

		// 設置牌山剩餘 8 張
		newRoom.Game.Deck = make([]string, 8)
		newRoom.Game.DiscardPile = []string{} // 清空棄牌堆

		p := newRoom.Players[0]
		newRoom.CurrentTurn = 0
		p.Hand = []string{
			"wan-1", "wan-1", "wan-1", "wan-2", "wan-2", "wan-2", "wan-3", "wan-3", "wan-3",
			"wan-4", "wan-4", "wan-4", "wan-5", "wan-5", "wan-5", "wan-6", "wan-7",
		}

		// 打出一張牌觸發流局
		_, _ = newRoom.HandleDiscard(p.ID, "wan-1")

		// 棄牌堆應該包含打出的牌
		if len(newRoom.Game.DiscardPile) != 1 {
			t.Errorf("棄牌堆應該有 1 張牌，實際 %d 張", len(newRoom.Game.DiscardPile))
		}

		if newRoom.Game.DiscardPile[0] != "wan-1" {
			t.Errorf("棄牌堆的牌應該是 wan-1，實際 %s", newRoom.Game.DiscardPile[0])
		}
	})
}
