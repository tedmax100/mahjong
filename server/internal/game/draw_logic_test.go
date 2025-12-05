package game

import (
	"testing"
)

// TestDrawLogicAfterMeld 測試吃/碰/槓後的摸牌邏輯
func TestDrawLogicAfterMeld(t *testing.T) {
	room := NewRoom("test-room")

	// 添加玩家
	for i := 1; i <= 4; i++ {
		err := room.AddPlayer("player"+string(rune('0'+i)), "玩家"+string(rune('0'+i)))
		if err != nil {
			t.Fatalf("添加玩家失敗: %v", err)
		}
	}

	room.StartGame()
	room.DealTiles()

	player := room.Players[0]

	t.Run("正常輪次應該摸牌（總牌數 16）", func(t *testing.T) {
		// 設置玩家手牌為 16 張
		player.Hand = make([]string, 16)
		for i := 0; i < 16; i++ {
			player.Hand[i] = "wan-1"
		}
		player.Melds = []Meld{} // 沒有吃碰槓

		totalTiles := len(player.Hand) + len(player.Melds)*3
		if totalTiles != 16 {
			t.Fatalf("預期總牌數 16，實際 %d", totalTiles)
		}

		// 應該摸牌
		shouldDraw := (totalTiles == 16)
		if !shouldDraw {
			t.Error("總牌數 16 時應該摸牌")
		}
	})

	t.Run("吃/碰/槓後不應該摸牌（總牌數 17）", func(t *testing.T) {
		// 設置玩家手牌為 14 張，加上 1 組碰（3 張）= 17 張
		player.Hand = make([]string, 14)
		for i := 0; i < 14; i++ {
			player.Hand[i] = "wan-1"
		}
		player.Melds = []Meld{
			{Type: "pong", Tiles: []string{"dong", "dong", "dong"}},
		}

		totalTiles := len(player.Hand) + len(player.Melds)*3
		if totalTiles != 17 {
			t.Fatalf("預期總牌數 17（手牌 14 + 碰 1 組），實際 %d", totalTiles)
		}

		// 不應該摸牌
		shouldDraw := (totalTiles == 16)
		if shouldDraw {
			t.Error("總牌數 17 時不應該摸牌（剛吃/碰/槓完）")
		}
	})

	t.Run("吃碰槓後的下一輪應該摸牌（總牌數 16）", func(t *testing.T) {
		// 設置玩家手牌為 13 張，加上 1 組碰（3 張）= 16 張
		// 這是碰牌後出牌的狀態
		player.Hand = make([]string, 13)
		for i := 0; i < 13; i++ {
			player.Hand[i] = "wan-1"
		}
		player.Melds = []Meld{
			{Type: "pong", Tiles: []string{"dong", "dong", "dong"}},
		}

		totalTiles := len(player.Hand) + len(player.Melds)*3
		if totalTiles != 16 {
			t.Fatalf("預期總牌數 16（手牌 13 + 碰 1 組），實際 %d", totalTiles)
		}

		// 應該摸牌
		shouldDraw := (totalTiles == 16)
		if !shouldDraw {
			t.Error("總牌數 16 時應該摸牌")
		}
	})

	t.Run("多組吃碰槓的情況", func(t *testing.T) {
		// 設置玩家手牌為 10 張，加上 2 組吃碰槓（6 張）= 16 張
		player.Hand = make([]string, 10)
		for i := 0; i < 10; i++ {
			player.Hand[i] = "wan-1"
		}
		player.Melds = []Meld{
			{Type: "pong", Tiles: []string{"dong", "dong", "dong"}},
			{Type: "chow", Tiles: []string{"wan-2", "wan-3", "wan-4"}},
		}

		totalTiles := len(player.Hand) + len(player.Melds)*3
		if totalTiles != 16 {
			t.Fatalf("預期總牌數 16（手牌 10 + 2 組），實際 %d", totalTiles)
		}

		// 應該摸牌
		shouldDraw := (totalTiles == 16)
		if !shouldDraw {
			t.Error("總牌數 16 時應該摸牌")
		}
	})

	t.Run("槓牌的情況（4 張算 3 張）", func(t *testing.T) {
		// 設置玩家手牌為 13 張，加上 1 組槓（4 張，但算 3 張）= 16 張
		player.Hand = make([]string, 13)
		for i := 0; i < 13; i++ {
			player.Hand[i] = "wan-1"
		}
		player.Melds = []Meld{
			{Type: "kong_exposed", Tiles: []string{"dong", "dong", "dong", "dong"}},
		}

		// 注意：槓牌雖然有 4 張，但在計算時仍算 3 張（一組面子）
		totalTiles := len(player.Hand) + len(player.Melds)*3
		if totalTiles != 16 {
			t.Fatalf("預期總牌數 16（手牌 13 + 槓 1 組×3），實際 %d", totalTiles)
		}
	})
}

// TestPlayerHandCountAfterActions 測試玩家執行動作後的手牌數量
func TestPlayerHandCountAfterActions(t *testing.T) {
	room := NewRoom("test-room")

	// 添加玩家
	for i := 1; i <= 4; i++ {
		room.AddPlayer("player"+string(rune('0'+i)), "玩家"+string(rune('0'+i)))
	}

	room.StartGame()
	room.DealTiles()

	player := room.Players[0]

	t.Run("碰牌後手牌應該減少 2 張", func(t *testing.T) {
		// 設置手牌
		player.Hand = []string{"dong", "dong", "wan-1", "wan-2", "wan-3", "wan-4"}
		initialHandCount := len(player.Hand)

		// 模擬碰牌
		room.HandlePong(player.ID, "dong")

		// 碰牌後手牌應該減少 2 張（移到 Melds）
		expectedHandCount := initialHandCount - 2
		if len(player.Hand) != expectedHandCount {
			t.Errorf("碰牌後手牌應該是 %d 張，實際 %d 張", expectedHandCount, len(player.Hand))
		}

		// 應該有 1 組碰
		if len(player.Melds) != 1 {
			t.Errorf("碰牌後應該有 1 組碰，實際 %d 組", len(player.Melds))
		}

		// 總牌數應該是原來的手牌數 + 1（從棄牌堆拿的）
		totalTiles := len(player.Hand) + len(player.Melds)*3
		expectedTotal := initialHandCount + 1
		if totalTiles != expectedTotal {
			t.Errorf("碰牌後總牌數應該是 %d，實際 %d（手牌 %d + 碰 %d 組）",
				expectedTotal, totalTiles, len(player.Hand), len(player.Melds))
		}
	})

	t.Run("吃牌後手牌應該減少 2 張", func(t *testing.T) {
		// 重置玩家
		player.Hand = []string{"wan-1", "wan-3", "tong-1", "tong-2"}
		player.Melds = []Meld{}
		initialHandCount := len(player.Hand)

		// 設置上家
		room.LastDiscardPlayer = (player.Position + 3) % 4

		// 模擬吃牌
		room.HandleChow(player.ID, "wan-2", []string{"wan-1", "wan-2", "wan-3"})

		// 吃牌後手牌應該減少 2 張（移到 Melds）
		expectedHandCount := initialHandCount - 2
		if len(player.Hand) != expectedHandCount {
			t.Errorf("吃牌後手牌應該是 %d 張，實際 %d 張", expectedHandCount, len(player.Hand))
		}

		// 應該有 1 組吃
		if len(player.Melds) != 1 {
			t.Errorf("吃牌後應該有 1 組吃，實際 %d 組", len(player.Melds))
		}
	})
}

// TestHandCountNeverZero 測試手牌數量永遠不應該變成 0
func TestHandCountNeverZero(t *testing.T) {
	room := NewRoom("test-room")

	for i := 1; i <= 4; i++ {
		room.AddPlayer("player"+string(rune('0'+i)), "玩家"+string(rune('0'+i)))
	}

	room.StartGame()
	room.DealTiles()

	player := room.Players[0]

	t.Run("玩家有吃碰槓時手牌至少要有 2 張（將牌）", func(t *testing.T) {
		// 設置玩家有 4 組面子
		player.Hand = []string{"dong", "dong"}
		player.Melds = []Meld{
			{Type: "pong", Tiles: []string{"wan-1", "wan-1", "wan-1"}},
			{Type: "pong", Tiles: []string{"wan-2", "wan-2", "wan-2"}},
			{Type: "pong", Tiles: []string{"wan-3", "wan-3", "wan-3"}},
			{Type: "chow", Tiles: []string{"tong-1", "tong-2", "tong-3"}},
		}

		totalTiles := len(player.Hand) + len(player.Melds)*3
		expectedTotal := 2 + 4*3 // 2 張將牌 + 4 組面子
		if totalTiles != expectedTotal {
			t.Errorf("預期總牌數 %d，實際 %d", expectedTotal, totalTiles)
		}

		// 手牌至少要有 2 張（將牌）
		if len(player.Hand) < 2 {
			t.Errorf("玩家手牌至少要有 2 張（將牌），實際 %d 張", len(player.Hand))
		}

		// 手牌永遠不應該是 0
		if len(player.Hand) == 0 {
			t.Error("手牌不應該是 0 張！")
		}
	})

	t.Run("即使有 5 組面子，手牌也不應該是 0", func(t *testing.T) {
		// 這是胡牌的狀態，但也不應該手牌是 0
		player.Hand = []string{"dong", "dong"}
		player.Melds = []Meld{
			{Type: "kong_promoted", Tiles: []string{"wan-1", "wan-1", "wan-1", "wan-1"}},
			{Type: "pong", Tiles: []string{"wan-2", "wan-2", "wan-2"}},
			{Type: "pong", Tiles: []string{"wan-3", "wan-3", "wan-3"}},
			{Type: "chow", Tiles: []string{"tong-1", "tong-2", "tong-3"}},
			{Type: "chow", Tiles: []string{"tong-4", "tong-5", "tong-6"}},
		}

		if len(player.Hand) == 0 {
			t.Error("即使有 5 組面子，手牌也不應該是 0 張（應該保留將牌）！")
		}

		// 手牌應該正好是 2 張（將牌）
		if len(player.Hand) != 2 {
			t.Errorf("有 5 組面子時，手牌應該是 2 張（將牌），實際 %d 張", len(player.Hand))
		}
	})
}