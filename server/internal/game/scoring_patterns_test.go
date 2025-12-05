package game

import (
	"testing"
)

// Helper function to check if a hand type is present
func hasHandType(result *WinResult, name string) bool {
	for _, ht := range result.HandTypes {
		if ht.Name == name {
			return true
		}
	}
	return false
}

// TestScoringPatterns 測試特殊牌型計分
func TestScoringPatterns(t *testing.T) {
	game := &MahjongGame{}

	t.Run("碰碰胡 (All Triplets)", func(t *testing.T) {
		player := &Player{
			Hand: []string{
				"wan-1", "wan-1", "wan-1", // 刻子
				"tong-2", "tong-2",        // 將牌
			},
			Melds: []Meld{
				{Type: "pong", Tiles: []string{"tiao-3", "tiao-3", "tiao-3"}},
				{Type: "pong", Tiles: []string{"dong", "dong", "dong"}},
				{Type: "pong", Tiles: []string{"nan", "nan", "nan"}},
				{Type: "kong_exposed", Tiles: []string{"xi", "xi", "xi", "xi"}},
			},
		}

		result := game.CalculateScore(player, "tong-2", true)
		
		if !hasHandType(result, "碰碰胡") {
			t.Error("應該識別為碰碰胡")
		}
	})

	t.Run("混一色 (Mixed One Suit)", func(t *testing.T) {
		player := &Player{
			Hand: []string{
				"wan-1", "wan-2", "wan-3",
				"dong", "dong",
			},
			Melds: []Meld{
				{Type: "chow", Tiles: []string{"wan-4", "wan-5", "wan-6"}},
				{Type: "pong", Tiles: []string{"wan-9", "wan-9", "wan-9"}},
				{Type: "pong", Tiles: []string{"nan", "nan", "nan"}},
				{Type: "pong", Tiles: []string{"zhong", "zhong", "zhong"}},
			},
		}

		result := game.CalculateScore(player, "dong", true)

		if !hasHandType(result, "混一色") {
			t.Error("應該識別為混一色")
		}
	})

	t.Run("清一色 (All One Suit)", func(t *testing.T) {
		player := &Player{
			Hand: []string{
				"wan-1", "wan-2", "wan-3",
				"wan-5", "wan-5",
			},
			Melds: []Meld{
				{Type: "chow", Tiles: []string{"wan-4", "wan-5", "wan-6"}},
				{Type: "pong", Tiles: []string{"wan-9", "wan-9", "wan-9"}},
				{Type: "pong", Tiles: []string{"wan-7", "wan-7", "wan-7"}},
				{Type: "chow", Tiles: []string{"wan-2", "wan-3", "wan-4"}},
			},
		}

		result := game.CalculateScore(player, "wan-5", true)

		if !hasHandType(result, "清一色") {
			t.Error("應該識別為清一色")
		}
	})

	t.Run("字一色 (All Honors)", func(t *testing.T) {
		player := &Player{
			Hand: []string{
				"dong", "dong",
			},
			Melds: []Meld{
				{Type: "pong", Tiles: []string{"nan", "nan", "nan"}},
				{Type: "pong", Tiles: []string{"xi", "xi", "xi"}},
				{Type: "pong", Tiles: []string{"bei", "bei", "bei"}},
				{Type: "pong", Tiles: []string{"zhong", "zhong", "zhong"}},
				{Type: "pong", Tiles: []string{"fa", "fa", "fa"}},
			},
		}

		result := game.CalculateScore(player, "dong", true)

		if !hasHandType(result, "字一色") {
			t.Error("應該識別為字一色")
		}
		// 字一色通常也隱含碰碰胡
		if !hasHandType(result, "碰碰胡") {
			t.Error("字一色應該也包含碰碰胡")
		}
	})

	t.Run("大三元 (Big Dragons)", func(t *testing.T) {
		player := &Player{
			Hand: []string{
				"wan-1", "wan-2", "wan-3",
				"tong-5", "tong-5",
			},
			Melds: []Meld{
				{Type: "pong", Tiles: []string{"zhong", "zhong", "zhong"}},
				{Type: "pong", Tiles: []string{"fa", "fa", "fa"}},
				{Type: "pong", Tiles: []string{"bai", "bai", "bai"}},
				{Type: "chow", Tiles: []string{"tiao-1", "tiao-2", "tiao-3"}},
			},
		}

		result := game.CalculateScore(player, "tong-5", true)

		if !hasHandType(result, "大三元") {
			t.Error("應該識別為大三元")
		}
	})

	t.Run("小三元 (Small Dragons)", func(t *testing.T) {
		player := &Player{
			Hand: []string{
				"zhong", "zhong", // 中為將牌
				"wan-1", "wan-2", "wan-3",
			},
			Melds: []Meld{
				{Type: "pong", Tiles: []string{"fa", "fa", "fa"}},
				{Type: "pong", Tiles: []string{"bai", "bai", "bai"}},
				{Type: "chow", Tiles: []string{"tiao-1", "tiao-2", "tiao-3"}},
				{Type: "pong", Tiles: []string{"tong-1", "tong-1", "tong-1"}},
			},
		}

		result := game.CalculateScore(player, "zhong", true)

		if !hasHandType(result, "小三元") {
			t.Error("應該識別為小三元")
		}
	})

	t.Run("大四喜 (Big Winds)", func(t *testing.T) {
		player := &Player{
			Hand: []string{
				"wan-1", "wan-1",
			},
			Melds: []Meld{
				{Type: "pong", Tiles: []string{"dong", "dong", "dong"}},
				{Type: "pong", Tiles: []string{"nan", "nan", "nan"}},
				{Type: "pong", Tiles: []string{"xi", "xi", "xi"}},
				{Type: "pong", Tiles: []string{"bei", "bei", "bei"}},
				{Type: "chow", Tiles: []string{"tong-1", "tong-2", "tong-3"}},
			},
		}

		result := game.CalculateScore(player, "wan-1", true)

		if !hasHandType(result, "大四喜") {
			t.Error("應該識別為大四喜")
		}
	})

	t.Run("小四喜 (Small Winds)", func(t *testing.T) {
		player := &Player{
			Hand: []string{
				"dong", "dong", // 東為將牌
				"wan-1", "wan-2", "wan-3",
			},
			Melds: []Meld{
				{Type: "pong", Tiles: []string{"nan", "nan", "nan"}},
				{Type: "pong", Tiles: []string{"xi", "xi", "xi"}},
				{Type: "pong", Tiles: []string{"bei", "bei", "bei"}},
				{Type: "pong", Tiles: []string{"tong-1", "tong-1", "tong-1"}},
			},
		}

		result := game.CalculateScore(player, "dong", true)

		if !hasHandType(result, "小四喜") {
			t.Error("應該識別為小四喜")
		}
	})
}
