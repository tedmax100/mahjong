package game

import (
	"mahjong/internal/model"
	"testing"
)

// TestCanHu_PlayerNN_Bug 測試玩家 NN 無法胡牌的 bug 修復
// 這是真實遊戲中發生的場景
func TestCanHu_PlayerNN_Bug(t *testing.T) {
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

	t.Run("玩家 NN 的胡牌手型應該被識別（2 組吃碰槓 + 11 張手牌）", func(t *testing.T) {
		// 玩家 NN 的手牌（11 張）
		hand := []string{
			"tong-1", "tong-2", "tong-2", "tong-3", "tong-3", "tong-4",
			"tong-6", "tong-6", "tong-6",
			"xi", "xi",
		}

		// 玩家 NN 的吃碰槓（2 組）
		melds := []model.Meld{
			{Type: "pong", Tiles: []string{"tiao-5", "tiao-5", "tiao-5"}},
			{Type: "chow", Tiles: []string{"wan-4", "wan-5", "wan-6"}},
		}

		// 驗證胡牌
		if !game.CanHu(hand, melds) {
			t.Error("玩家 NN 應該能夠胡牌")
		}
	})

	t.Run("驗證手牌可以組成正確的牌型", func(t *testing.T) {
		hand := []string{
			"tong-1", "tong-2", "tong-2", "tong-3", "tong-3", "tong-4",
			"tong-6", "tong-6", "tong-6",
			"xi", "xi",
		}

		// 分析手牌結構：
		// - 順子：tong-1 tong-2 tong-3
		// - 順子：tong-2 tong-3 tong-4
		// - 刻子：tong-6 tong-6 tong-6
		// - 對子：xi xi
		// 總共：3 組面子 + 1 對眼 = 11 張

		// 加上已有的 2 組吃碰槓：
		// - 刻子：tiao-5 tiao-5 tiao-5
		// - 順子：wan-4 wan-5 wan-6
		// 總計：5 組面子 + 1 對眼 = 17 張 ✓

		melds := []model.Meld{
			{Type: "pong", Tiles: []string{"tiao-5", "tiao-5", "tiao-5"}},
			{Type: "chow", Tiles: []string{"wan-4", "wan-5", "wan-6"}},
		}

		if !game.CanHu(hand, melds) {
			t.Error("完整的 17 張牌型應該能夠胡牌")
		}
	})
}

// TestCanHu_GroupCounts 測試不同吃碰槓數量下的胡牌判定
func TestCanHu_GroupCounts(t *testing.T) {
	players := make([]*Player, 4)
	for i := 0; i < 4; i++ {
		players[i] = &Player{
			ID:       "player" + string(rune('1'+i)),
			Name:     "玩家" + string(rune('1'+i)),
			Position: i,
		}
	}

	game := NewMahjongGame(players)

	t.Run("0 組吃碰槓 - 需要 17 張手牌（5 組 + 1 對）", func(t *testing.T) {
		// 完全沒有吃碰槓，手牌需要 17 張
		hand := []string{
			"wan-1", "wan-2", "wan-3", // 順子 1
			"wan-4", "wan-5", "wan-6", // 順子 2
			"tiao-1", "tiao-2", "tiao-3", // 順子 3
			"tong-1", "tong-1", "tong-1", // 刻子 1
			"tong-2", "tong-2", "tong-2", // 刻子 2
			"xi", "xi", // 對子
		}

		melds := []model.Meld{}

		if !game.CanHu(hand, melds) {
			t.Error("17 張手牌（5 組+1 對）應該能夠胡牌")
		}
	})

	t.Run("1 組吃碰槓 - 需要 14 張手牌（4 組 + 1 對）", func(t *testing.T) {
		hand := []string{
			"wan-1", "wan-2", "wan-3", // 順子 1
			"wan-4", "wan-5", "wan-6", // 順子 2
			"tiao-1", "tiao-2", "tiao-3", // 順子 3
			"tong-1", "tong-1", "tong-1", // 刻子 1
			"xi", "xi", // 對子
		}

		melds := []model.Meld{
			{Type: "pong", Tiles: []string{"tong-2", "tong-2", "tong-2"}},
		}

		if !game.CanHu(hand, melds) {
			t.Error("1 組吃碰槓 + 14 張手牌應該能夠胡牌")
		}
	})

	t.Run("2 組吃碰槓 - 需要 11 張手牌（3 組 + 1 對）", func(t *testing.T) {
		hand := []string{
			"wan-1", "wan-2", "wan-3", // 順子 1
			"wan-4", "wan-5", "wan-6", // 順子 2
			"tiao-1", "tiao-2", "tiao-3", // 順子 3
			"xi", "xi", // 對子
		}

		melds := []model.Meld{
			{Type: "pong", Tiles: []string{"tong-1", "tong-1", "tong-1"}},
			{Type: "pong", Tiles: []string{"tong-2", "tong-2", "tong-2"}},
		}

		if !game.CanHu(hand, melds) {
			t.Error("2 組吃碰槓 + 11 張手牌應該能夠胡牌")
		}
	})

	t.Run("3 組吃碰槓 - 需要 8 張手牌（2 組 + 1 對）", func(t *testing.T) {
		hand := []string{
			"wan-1", "wan-2", "wan-3", // 順子 1
			"wan-4", "wan-5", "wan-6", // 順子 2
			"xi", "xi", // 對子
		}

		melds := []model.Meld{
			{Type: "pong", Tiles: []string{"tong-1", "tong-1", "tong-1"}},
			{Type: "pong", Tiles: []string{"tong-2", "tong-2", "tong-2"}},
			{Type: "chow", Tiles: []string{"tiao-1", "tiao-2", "tiao-3"}},
		}

		if !game.CanHu(hand, melds) {
			t.Error("3 組吃碰槓 + 8 張手牌應該能夠胡牌")
		}
	})

	t.Run("4 組吃碰槓 - 需要 5 張手牌（1 組 + 1 對）", func(t *testing.T) {
		hand := []string{
			"wan-1", "wan-2", "wan-3", // 順子 1
			"xi", "xi", // 對子
		}

		melds := []model.Meld{
			{Type: "pong", Tiles: []string{"tong-1", "tong-1", "tong-1"}},
			{Type: "pong", Tiles: []string{"tong-2", "tong-2", "tong-2"}},
			{Type: "chow", Tiles: []string{"tiao-1", "tiao-2", "tiao-3"}},
			{Type: "kong_exposed", Tiles: []string{"dong", "dong", "dong", "dong"}},
		}

		if !game.CanHu(hand, melds) {
			t.Error("4 組吃碰槓 + 5 張手牌應該能夠胡牌")
		}
	})

	t.Run("5 組吃碰槓 - 只需要 1 對（2 張手牌）", func(t *testing.T) {
		hand := []string{
			"xi", "xi", // 對子
		}

		melds := []model.Meld{
			{Type: "pong", Tiles: []string{"tong-1", "tong-1", "tong-1"}},
			{Type: "pong", Tiles: []string{"tong-2", "tong-2", "tong-2"}},
			{Type: "chow", Tiles: []string{"tiao-1", "tiao-2", "tiao-3"}},
			{Type: "chow", Tiles: []string{"wan-1", "wan-2", "wan-3"}},
			{Type: "pong", Tiles: []string{"dong", "dong", "dong"}},
		}

		if !game.CanHu(hand, melds) {
			t.Error("5 組吃碰槓 + 1 對應該能夠胡牌")
		}
	})
}

// TestCanHu_InvalidCases 測試無效的胡牌情況
func TestCanHu_InvalidCases(t *testing.T) {
	players := make([]*Player, 4)
	for i := 0; i < 4; i++ {
		players[i] = &Player{
			ID:       "player" + string(rune('1'+i)),
			Name:     "玩家" + string(rune('1'+i)),
			Position: i,
		}
	}

	game := NewMahjongGame(players)

	t.Run("手牌數量不正確（2 組吃碰槓 + 10 張手牌）", func(t *testing.T) {
		hand := []string{
			"wan-1", "wan-2", "wan-3",
			"wan-4", "wan-5", "wan-6",
			"tiao-1", "tiao-2", "tiao-3",
			"xi", // 缺少一張
		}

		melds := []model.Meld{
			{Type: "pong", Tiles: []string{"tong-1", "tong-1", "tong-1"}},
			{Type: "pong", Tiles: []string{"tong-2", "tong-2", "tong-2"}},
		}

		if game.CanHu(hand, melds) {
			t.Error("手牌數量不正確，不應該能夠胡牌")
		}
	})

	t.Run("手牌無法組成有效牌型", func(t *testing.T) {
		hand := []string{
			"wan-1", "wan-3", "wan-5", // 無法組成順子
			"wan-4", "wan-5", "wan-6",
			"tiao-1", "tiao-2", "tiao-3",
			"xi", "xi",
		}

		melds := []model.Meld{
			{Type: "pong", Tiles: []string{"tong-1", "tong-1", "tong-1"}},
			{Type: "pong", Tiles: []string{"tong-2", "tong-2", "tong-2"}},
		}

		if game.CanHu(hand, melds) {
			t.Error("無法組成有效牌型，不應該能夠胡牌")
		}
	})

	t.Run("缺少對子", func(t *testing.T) {
		hand := []string{
			"wan-1", "wan-2", "wan-3",
			"wan-4", "wan-5", "wan-6",
			"tiao-1", "tiao-2", "tiao-3",
			"tong-5", "tong-6", // 不是對子
		}

		melds := []model.Meld{
			{Type: "pong", Tiles: []string{"tong-1", "tong-1", "tong-1"}},
			{Type: "pong", Tiles: []string{"tong-2", "tong-2", "tong-2"}},
		}

		if game.CanHu(hand, melds) {
			t.Error("缺少對子，不應該能夠胡牌")
		}
	})
}

// TestCanHu_ComplexPatterns 測試複雜的牌型
func TestCanHu_ComplexPatterns(t *testing.T) {
	players := make([]*Player, 4)
	for i := 0; i < 4; i++ {
		players[i] = &Player{
			ID:       "player" + string(rune('1'+i)),
			Name:     "玩家" + string(rune('1'+i)),
			Position: i,
		}
	}

	game := NewMahjongGame(players)

	t.Run("全是刻子的牌型", func(t *testing.T) {
		hand := []string{
			"wan-1", "wan-1", "wan-1",
			"wan-2", "wan-2", "wan-2",
			"tiao-3", "tiao-3", "tiao-3",
			"tong-4", "tong-4",
		}

		melds := []model.Meld{
			{Type: "pong", Tiles: []string{"tong-5", "tong-5", "tong-5"}},
			{Type: "pong", Tiles: []string{"xi", "xi", "xi"}},
		}

		if !game.CanHu(hand, melds) {
			t.Error("全是刻子的牌型應該能夠胡牌")
		}
	})

	t.Run("全是順子的牌型", func(t *testing.T) {
		hand := []string{
			"wan-1", "wan-2", "wan-3",
			"wan-4", "wan-5", "wan-6",
			"wan-7", "wan-8", "wan-9",
			"xi", "xi",
		}

		melds := []model.Meld{
			{Type: "chow", Tiles: []string{"tiao-1", "tiao-2", "tiao-3"}},
			{Type: "chow", Tiles: []string{"tong-1", "tong-2", "tong-3"}},
		}

		if !game.CanHu(hand, melds) {
			t.Error("全是順子的牌型應該能夠胡牌")
		}
	})

	t.Run("混合牌型 - 順子和刻子", func(t *testing.T) {
		hand := []string{
			"wan-1", "wan-2", "wan-3", // 順子
			"tiao-5", "tiao-5", "tiao-5", // 刻子
			"tong-1", "tong-2", "tong-3", // 順子
			"dong", "dong", // 對子
		}

		melds := []model.Meld{
			{Type: "pong", Tiles: []string{"xi", "xi", "xi"}},
			{Type: "chow", Tiles: []string{"wan-7", "wan-8", "wan-9"}},
		}

		if !game.CanHu(hand, melds) {
			t.Error("混合牌型應該能夠胡牌")
		}
	})
}

// TestCanHu_BugFromLogs is the test case that reproduces the bug from the logs.
func TestCanHu_BugFromLogs(t *testing.T) {
	players := make([]*Player, 1)
	players[0] = &Player{ID: "player1", Name: "Player 1", Position: 0}
	game := NewMahjongGame(players)

	t.Run("Bug from game log", func(t *testing.T) {
		hand := []string{
			"wan-1", "wan-2", "wan-3",
			"tong-3", "tong-3", "tong-3",
			"tiao-5", "tiao-6", "tiao-6", "tiao-7", "tiao-7", "tiao-8", "tiao-9", "tiao-9",
		}
		melds := []model.Meld{
			{Type: "pong", Tiles: []string{"wan-8", "wan-8", "wan-8"}},
		}

		if !game.CanHu(hand, melds) {
			t.Errorf("Hand from logs should be a winning hand but was not detected.")
		}
	})
}