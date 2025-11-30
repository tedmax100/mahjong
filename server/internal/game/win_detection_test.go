package game

import (
	"testing"
)

// TestCanHu_PlayerNN_Bug 测试玩家 NN 无法胡牌的 bug 修复
// 这是真实游戏中发生的场景
func TestCanHu_PlayerNN_Bug(t *testing.T) {
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

	t.Run("玩家 NN 的胡牌手型应该被识别（2组吃碰槓 + 11张手牌）", func(t *testing.T) {
		// 玩家 NN 的手牌（11张）
		hand := []string{
			"tong-1", "tong-2", "tong-2", "tong-3", "tong-3", "tong-4",
			"tong-6", "tong-6", "tong-6",
			"xi", "xi",
		}

		// 玩家 NN 的吃碰槓（2组）
		melds := []Meld{
			{Type: "pong", Tiles: []string{"tiao-5", "tiao-5", "tiao-5"}},
			{Type: "chow", Tiles: []string{"wan-4", "wan-5", "wan-6"}},
		}

		// 验证胡牌
		if !game.CanHu(hand, melds) {
			t.Error("玩家 NN 应该能够胡牌")
		}
	})

	t.Run("验证手牌可以组成正确的牌型", func(t *testing.T) {
		hand := []string{
			"tong-1", "tong-2", "tong-2", "tong-3", "tong-3", "tong-4",
			"tong-6", "tong-6", "tong-6",
			"xi", "xi",
		}

		// 分析手牌结构：
		// - 顺子：tong-1 tong-2 tong-3
		// - 顺子：tong-2 tong-3 tong-4
		// - 刻子：tong-6 tong-6 tong-6
		// - 对子：xi xi
		// 总共：3组面子 + 1对眼 = 11张

		// 加上已有的2组吃碰槓：
		// - 刻子：tiao-5 tiao-5 tiao-5
		// - 顺子：wan-4 wan-5 wan-6
		// 总计：5组面子 + 1对眼 = 17张 ✓

		melds := []Meld{
			{Type: "pong", Tiles: []string{"tiao-5", "tiao-5", "tiao-5"}},
			{Type: "chow", Tiles: []string{"wan-4", "wan-5", "wan-6"}},
		}

		if !game.CanHu(hand, melds) {
			t.Error("完整的17张牌型应该能够胡牌")
		}
	})
}

// TestCanHu_GroupCounts 测试不同吃碰槓数量下的胡牌判定
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

	t.Run("0组吃碰槓 - 需要17张手牌（5组 + 1对）", func(t *testing.T) {
		// 完全没有吃碰槓，手牌需要17张
		hand := []string{
			"wan-1", "wan-2", "wan-3", // 顺子1
			"wan-4", "wan-5", "wan-6", // 顺子2
			"tiao-1", "tiao-2", "tiao-3", // 顺子3
			"tong-1", "tong-1", "tong-1", // 刻子1
			"tong-2", "tong-2", "tong-2", // 刻子2
			"xi", "xi", // 对子
		}

		melds := []Meld{}

		if !game.CanHu(hand, melds) {
			t.Error("17张手牌（5组+1对）应该能够胡牌")
		}
	})

	t.Run("1组吃碰槓 - 需要14张手牌（4组 + 1对）", func(t *testing.T) {
		hand := []string{
			"wan-1", "wan-2", "wan-3", // 顺子1
			"wan-4", "wan-5", "wan-6", // 顺子2
			"tiao-1", "tiao-2", "tiao-3", // 顺子3
			"tong-1", "tong-1", "tong-1", // 刻子1
			"xi", "xi", // 对子
		}

		melds := []Meld{
			{Type: "pong", Tiles: []string{"tong-2", "tong-2", "tong-2"}},
		}

		if !game.CanHu(hand, melds) {
			t.Error("1组吃碰槓 + 14张手牌应该能够胡牌")
		}
	})

	t.Run("2组吃碰槓 - 需要11张手牌（3组 + 1对）", func(t *testing.T) {
		hand := []string{
			"wan-1", "wan-2", "wan-3", // 顺子1
			"wan-4", "wan-5", "wan-6", // 顺子2
			"tiao-1", "tiao-2", "tiao-3", // 顺子3
			"xi", "xi", // 对子
		}

		melds := []Meld{
			{Type: "pong", Tiles: []string{"tong-1", "tong-1", "tong-1"}},
			{Type: "pong", Tiles: []string{"tong-2", "tong-2", "tong-2"}},
		}

		if !game.CanHu(hand, melds) {
			t.Error("2组吃碰槓 + 11张手牌应该能够胡牌")
		}
	})

	t.Run("3组吃碰槓 - 需要8张手牌（2组 + 1对）", func(t *testing.T) {
		hand := []string{
			"wan-1", "wan-2", "wan-3", // 顺子1
			"wan-4", "wan-5", "wan-6", // 顺子2
			"xi", "xi", // 对子
		}

		melds := []Meld{
			{Type: "pong", Tiles: []string{"tong-1", "tong-1", "tong-1"}},
			{Type: "pong", Tiles: []string{"tong-2", "tong-2", "tong-2"}},
			{Type: "chow", Tiles: []string{"tiao-1", "tiao-2", "tiao-3"}},
		}

		if !game.CanHu(hand, melds) {
			t.Error("3组吃碰槓 + 8张手牌应该能够胡牌")
		}
	})

	t.Run("4组吃碰槓 - 需要5张手牌（1组 + 1对）", func(t *testing.T) {
		hand := []string{
			"wan-1", "wan-2", "wan-3", // 顺子1
			"xi", "xi", // 对子
		}

		melds := []Meld{
			{Type: "pong", Tiles: []string{"tong-1", "tong-1", "tong-1"}},
			{Type: "pong", Tiles: []string{"tong-2", "tong-2", "tong-2"}},
			{Type: "chow", Tiles: []string{"tiao-1", "tiao-2", "tiao-3"}},
			{Type: "kong_exposed", Tiles: []string{"dong", "dong", "dong", "dong"}},
		}

		if !game.CanHu(hand, melds) {
			t.Error("4组吃碰槓 + 5张手牌应该能够胡牌")
		}
	})

	t.Run("5组吃碰槓 - 只需要1对（2张手牌）", func(t *testing.T) {
		hand := []string{
			"xi", "xi", // 对子
		}

		melds := []Meld{
			{Type: "pong", Tiles: []string{"tong-1", "tong-1", "tong-1"}},
			{Type: "pong", Tiles: []string{"tong-2", "tong-2", "tong-2"}},
			{Type: "chow", Tiles: []string{"tiao-1", "tiao-2", "tiao-3"}},
			{Type: "chow", Tiles: []string{"wan-1", "wan-2", "wan-3"}},
			{Type: "pong", Tiles: []string{"dong", "dong", "dong"}},
		}

		if !game.CanHu(hand, melds) {
			t.Error("5组吃碰槓 + 1对应该能够胡牌")
		}
	})
}

// TestCanHu_InvalidCases 测试无效的胡牌情况
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

	t.Run("手牌数量不正确（2组吃碰槓 + 10张手牌）", func(t *testing.T) {
		hand := []string{
			"wan-1", "wan-2", "wan-3",
			"wan-4", "wan-5", "wan-6",
			"tiao-1", "tiao-2", "tiao-3",
			"xi", // 缺少一张
		}

		melds := []Meld{
			{Type: "pong", Tiles: []string{"tong-1", "tong-1", "tong-1"}},
			{Type: "pong", Tiles: []string{"tong-2", "tong-2", "tong-2"}},
		}

		if game.CanHu(hand, melds) {
			t.Error("手牌数量不正确，不应该能够胡牌")
		}
	})

	t.Run("手牌无法组成有效牌型", func(t *testing.T) {
		hand := []string{
			"wan-1", "wan-3", "wan-5", // 无法组成顺子
			"wan-4", "wan-5", "wan-6",
			"tiao-1", "tiao-2", "tiao-3",
			"xi", "xi",
		}

		melds := []Meld{
			{Type: "pong", Tiles: []string{"tong-1", "tong-1", "tong-1"}},
			{Type: "pong", Tiles: []string{"tong-2", "tong-2", "tong-2"}},
		}

		if game.CanHu(hand, melds) {
			t.Error("无法组成有效牌型，不应该能够胡牌")
		}
	})

	t.Run("缺少对子", func(t *testing.T) {
		hand := []string{
			"wan-1", "wan-2", "wan-3",
			"wan-4", "wan-5", "wan-6",
			"tiao-1", "tiao-2", "tiao-3",
			"tong-5", "tong-6", // 不是对子
		}

		melds := []Meld{
			{Type: "pong", Tiles: []string{"tong-1", "tong-1", "tong-1"}},
			{Type: "pong", Tiles: []string{"tong-2", "tong-2", "tong-2"}},
		}

		if game.CanHu(hand, melds) {
			t.Error("缺少对子，不应该能够胡牌")
		}
	})
}

// TestCanHu_ComplexPatterns 测试复杂的牌型
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

		melds := []Meld{
			{Type: "pong", Tiles: []string{"tong-5", "tong-5", "tong-5"}},
			{Type: "pong", Tiles: []string{"xi", "xi", "xi"}},
		}

		if !game.CanHu(hand, melds) {
			t.Error("全是刻子的牌型应该能够胡牌")
		}
	})

	t.Run("全是顺子的牌型", func(t *testing.T) {
		hand := []string{
			"wan-1", "wan-2", "wan-3",
			"wan-4", "wan-5", "wan-6",
			"wan-7", "wan-8", "wan-9",
			"xi", "xi",
		}

		melds := []Meld{
			{Type: "chow", Tiles: []string{"tiao-1", "tiao-2", "tiao-3"}},
			{Type: "chow", Tiles: []string{"tong-1", "tong-2", "tong-3"}},
		}

		if !game.CanHu(hand, melds) {
			t.Error("全是顺子的牌型应该能够胡牌")
		}
	})

	t.Run("混合牌型 - 顺子和刻子", func(t *testing.T) {
		hand := []string{
			"wan-1", "wan-2", "wan-3", // 顺子
			"tiao-5", "tiao-5", "tiao-5", // 刻子
			"tong-1", "tong-2", "tong-3", // 顺子
			"dong", "dong", // 对子
		}

		melds := []Meld{
			{Type: "pong", Tiles: []string{"xi", "xi", "xi"}},
			{Type: "chow", Tiles: []string{"wan-7", "wan-8", "wan-9"}},
		}

		if !game.CanHu(hand, melds) {
			t.Error("混合牌型应该能够胡牌")
		}
	})
}
