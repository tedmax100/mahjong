package scoring

import (
	"mahjong/internal/model"
	"testing"
)

// Helper function to check if a breakdown item is present
func hasBreakdown(result *WinResult, key string) bool {
	for _, b := range result.Breakdown {
		if b.Key == key {
			return true
		}
	}
	return false
}

// Helper function to get breakdown value by key
func getBreakdownValue(result *WinResult, key string) int {
	for _, b := range result.Breakdown {
		if b.Key == key {
			return b.Value
		}
	}
	return 0
}

// TestRoundWindTriplet 測試場風刻
func TestRoundWindTriplet(t *testing.T) {
	t.Run("東風局有東風刻子", func(t *testing.T) {
		// 東風局，玩家有東風刻子
		input := &ScoreInput{
			RoundWind:   WindEast,
			SeatWind:    WindSouth, // 南家
			IsDealer:    false,
			IsSelfDrawn: true,
			Hand: []string{
				"wan-1", "wan-2", "wan-3",
				"tong-5", "tong-5",
			},
			Melds: []model.Meld{
				{Type: "pong", Tiles: []string{"dong", "dong", "dong"}}, // 東風碰
				{Type: "chow", Tiles: []string{"tiao-1", "tiao-2", "tiao-3"}},
				{Type: "chow", Tiles: []string{"wan-4", "wan-5", "wan-6"}},
				{Type: "pong", Tiles: []string{"tong-9", "tong-9", "tong-9"}},
			},
			Flowers:     []string{},
			WinningTile: "tong-5",
		}

		result := CalculateScoreWithInput(input)

		if !hasBreakdown(result, "ROUND_WIND_TRIPLET") {
			t.Error("應該識別場風刻")
		}
		if hasBreakdown(result, "SEAT_WIND_TRIPLET") {
			t.Error("不應該識別門風刻（南家碰東風）")
		}
	})

	t.Run("南風局有南風刻子", func(t *testing.T) {
		input := &ScoreInput{
			RoundWind:   WindSouth,
			SeatWind:    WindEast,
			IsDealer:    true,
			IsSelfDrawn: true,
			Hand: []string{
				"nan", "nan", "nan", // 南風暗刻
				"wan-5", "wan-5",
			},
			Melds: []model.Meld{
				{Type: "chow", Tiles: []string{"tiao-1", "tiao-2", "tiao-3"}},
				{Type: "chow", Tiles: []string{"wan-4", "wan-5", "wan-6"}},
				{Type: "pong", Tiles: []string{"tong-9", "tong-9", "tong-9"}},
				{Type: "pong", Tiles: []string{"tong-1", "tong-1", "tong-1"}},
			},
			Flowers:     []string{},
			WinningTile: "wan-5",
		}

		result := CalculateScoreWithInput(input)

		if !hasBreakdown(result, "ROUND_WIND_TRIPLET") {
			t.Error("應該識別場風刻（南風局南風暗刻）")
		}
	})
}

// TestSeatWindTriplet 測試門風刻
func TestSeatWindTriplet(t *testing.T) {
	t.Run("東家有東風刻子", func(t *testing.T) {
		input := &ScoreInput{
			RoundWind:   WindSouth, // 南風局
			SeatWind:    WindEast,  // 東家
			IsDealer:    true,
			IsSelfDrawn: true,
			Hand: []string{
				"wan-1", "wan-2", "wan-3",
				"tong-5", "tong-5",
			},
			Melds: []model.Meld{
				{Type: "pong", Tiles: []string{"dong", "dong", "dong"}}, // 東風碰
				{Type: "chow", Tiles: []string{"tiao-1", "tiao-2", "tiao-3"}},
				{Type: "chow", Tiles: []string{"wan-4", "wan-5", "wan-6"}},
				{Type: "pong", Tiles: []string{"tong-9", "tong-9", "tong-9"}},
			},
			Flowers:     []string{},
			WinningTile: "tong-5",
		}

		result := CalculateScoreWithInput(input)

		if !hasBreakdown(result, "SEAT_WIND_TRIPLET") {
			t.Error("應該識別門風刻（東家碰東風）")
		}
		if hasBreakdown(result, "ROUND_WIND_TRIPLET") {
			t.Error("不應該識別場風刻（南風局碰東風）")
		}
	})
}

// TestBothWindTriplets 測試場風=門風的情況
func TestBothWindTriplets(t *testing.T) {
	t.Run("東風局東家碰東風（同時計算場風刻+門風刻）", func(t *testing.T) {
		input := &ScoreInput{
			RoundWind:   WindEast,
			SeatWind:    WindEast,
			IsDealer:    true,
			IsSelfDrawn: true,
			Hand: []string{
				"wan-1", "wan-2", "wan-3",
				"tong-5", "tong-5",
			},
			Melds: []model.Meld{
				{Type: "pong", Tiles: []string{"dong", "dong", "dong"}}, // 東風碰
				{Type: "chow", Tiles: []string{"tiao-1", "tiao-2", "tiao-3"}},
				{Type: "chow", Tiles: []string{"wan-4", "wan-5", "wan-6"}},
				{Type: "pong", Tiles: []string{"tong-9", "tong-9", "tong-9"}},
			},
			Flowers:     []string{},
			WinningTile: "tong-5",
		}

		result := CalculateScoreWithInput(input)

		if !hasBreakdown(result, "ROUND_WIND_TRIPLET") {
			t.Error("應該識別場風刻")
		}
		if !hasBreakdown(result, "SEAT_WIND_TRIPLET") {
			t.Error("應該識別門風刻")
		}

		// 確認總共加了 2 台（場風刻 1 + 門風刻 1）
		windTai := getBreakdownValue(result, "ROUND_WIND_TRIPLET") + getBreakdownValue(result, "SEAT_WIND_TRIPLET")
		if windTai != 2 {
			t.Errorf("場風刻+門風刻應該共 2 台，實際 %d 台", windTai)
		}
	})
}

// TestNoWindTriplets 測試無風牌刻子
func TestNoWindTriplets(t *testing.T) {
	t.Run("沒有風牌刻子", func(t *testing.T) {
		input := &ScoreInput{
			RoundWind:   WindEast,
			SeatWind:    WindSouth,
			IsDealer:    false,
			IsSelfDrawn: false,
			Hand: []string{
				"wan-1", "wan-2", "wan-3",
				"tong-5", "tong-5",
			},
			Melds: []model.Meld{
				{Type: "chow", Tiles: []string{"tiao-1", "tiao-2", "tiao-3"}},
				{Type: "chow", Tiles: []string{"wan-4", "wan-5", "wan-6"}},
				{Type: "pong", Tiles: []string{"tong-9", "tong-9", "tong-9"}},
				{Type: "pong", Tiles: []string{"zhong", "zhong", "zhong"}}, // 三元牌，不是風牌
			},
			Flowers:     []string{},
			WinningTile: "tong-5",
		}

		result := CalculateScoreWithInput(input)

		if hasBreakdown(result, "ROUND_WIND_TRIPLET") {
			t.Error("不應該識別場風刻")
		}
		if hasBreakdown(result, "SEAT_WIND_TRIPLET") {
			t.Error("不應該識別門風刻")
		}
	})
}

// TestKongScoring 測試槓牌計分
func TestKongScoring(t *testing.T) {
	t.Run("暗槓 2 台", func(t *testing.T) {
		input := &ScoreInput{
			RoundWind:   WindEast,
			SeatWind:    WindEast,
			IsDealer:    true,
			IsSelfDrawn: true,
			Hand: []string{
				"wan-1", "wan-2", "wan-3",
				"tong-5", "tong-5",
			},
			Melds: []model.Meld{
				{Type: "kong_concealed", Tiles: []string{"dong", "dong", "dong", "dong"}}, // 暗槓
				{Type: "chow", Tiles: []string{"tiao-1", "tiao-2", "tiao-3"}},
				{Type: "chow", Tiles: []string{"wan-4", "wan-5", "wan-6"}},
				{Type: "pong", Tiles: []string{"tong-9", "tong-9", "tong-9"}},
			},
			Flowers:     []string{},
			WinningTile: "tong-5",
		}

		result := CalculateScoreWithInput(input)

		if !hasBreakdown(result, "KONG_CONCEALED") {
			t.Error("應該識別暗槓")
		}
		if getBreakdownValue(result, "KONG_CONCEALED") != 2 {
			t.Errorf("暗槓應該 2 台，實際 %d 台", getBreakdownValue(result, "KONG_CONCEALED"))
		}
	})

	t.Run("明槓 1 台", func(t *testing.T) {
		input := &ScoreInput{
			RoundWind:   WindEast,
			SeatWind:    WindSouth,
			IsDealer:    false,
			IsSelfDrawn: true,
			Hand: []string{
				"wan-1", "wan-2", "wan-3",
				"tong-5", "tong-5",
			},
			Melds: []model.Meld{
				{Type: "kong_exposed", Tiles: []string{"nan", "nan", "nan", "nan"}}, // 明槓
				{Type: "chow", Tiles: []string{"tiao-1", "tiao-2", "tiao-3"}},
				{Type: "chow", Tiles: []string{"wan-4", "wan-5", "wan-6"}},
				{Type: "pong", Tiles: []string{"tong-9", "tong-9", "tong-9"}},
			},
			Flowers:     []string{},
			WinningTile: "tong-5",
		}

		result := CalculateScoreWithInput(input)

		if !hasBreakdown(result, "KONG_EXPOSED") {
			t.Error("應該識別明槓")
		}
		if getBreakdownValue(result, "KONG_EXPOSED") != 1 {
			t.Errorf("明槓應該 1 台，實際 %d 台", getBreakdownValue(result, "KONG_EXPOSED"))
		}
	})

	t.Run("加槓 1 台", func(t *testing.T) {
		input := &ScoreInput{
			RoundWind:   WindEast,
			SeatWind:    WindSouth,
			IsDealer:    false,
			IsSelfDrawn: true,
			Hand: []string{
				"wan-1", "wan-2", "wan-3",
				"tong-5", "tong-5",
			},
			Melds: []model.Meld{
				{Type: "kong_promoted", Tiles: []string{"xi", "xi", "xi", "xi"}}, // 加槓
				{Type: "chow", Tiles: []string{"tiao-1", "tiao-2", "tiao-3"}},
				{Type: "chow", Tiles: []string{"wan-4", "wan-5", "wan-6"}},
				{Type: "pong", Tiles: []string{"tong-9", "tong-9", "tong-9"}},
			},
			Flowers:     []string{},
			WinningTile: "tong-5",
		}

		result := CalculateScoreWithInput(input)

		if !hasBreakdown(result, "KONG_PROMOTED") {
			t.Error("應該識別加槓")
		}
		if getBreakdownValue(result, "KONG_PROMOTED") != 1 {
			t.Errorf("加槓應該 1 台，實際 %d 台", getBreakdownValue(result, "KONG_PROMOTED"))
		}
	})

	t.Run("多個槓累計", func(t *testing.T) {
		input := &ScoreInput{
			RoundWind:   WindEast,
			SeatWind:    WindSouth,
			IsDealer:    false,
			IsSelfDrawn: true,
			Hand: []string{
				"wan-1", "wan-2", "wan-3",
				"tong-5", "tong-5",
			},
			Melds: []model.Meld{
				{Type: "kong_concealed", Tiles: []string{"dong", "dong", "dong", "dong"}}, // 暗槓 +2
				{Type: "kong_exposed", Tiles: []string{"nan", "nan", "nan", "nan"}},       // 明槓 +1
				{Type: "chow", Tiles: []string{"wan-4", "wan-5", "wan-6"}},
				{Type: "pong", Tiles: []string{"tong-9", "tong-9", "tong-9"}},
			},
			Flowers:     []string{},
			WinningTile: "tong-5",
		}

		result := CalculateScoreWithInput(input)

		kongTai := getBreakdownValue(result, "KONG_CONCEALED") + getBreakdownValue(result, "KONG_EXPOSED")
		if kongTai != 3 {
			t.Errorf("暗槓(2)+明槓(1)應該共 3 台，實際 %d 台", kongTai)
		}
	})
}

// TestSpecialConditions 測試特殊胡牌情境
func TestSpecialConditions(t *testing.T) {
	t.Run("海底撈月", func(t *testing.T) {
		input := &ScoreInput{
			RoundWind:   WindEast,
			SeatWind:    WindEast,
			IsDealer:    true,
			IsSelfDrawn: true,
			Hand: []string{
				"wan-1", "wan-2", "wan-3",
				"wan-4", "wan-5", "wan-6",
				"tong-1", "tong-2", "tong-3",
				"tiao-1", "tiao-2", "tiao-3",
				"tong-5", "tong-5",
				"tong-7", "tong-8", "tong-9",
			},
			Melds:             []model.Meld{},
			Flowers:           []string{},
			WinningTile:       "tong-5",
			SpecialConditions: []SpecialCondition{ConditionLastTile},
		}

		result := CalculateScoreWithInput(input)

		if !hasBreakdown(result, "LAST_TILE") {
			t.Error("應該識別海底撈月")
		}
	})

	t.Run("槓上開花", func(t *testing.T) {
		input := &ScoreInput{
			RoundWind:   WindEast,
			SeatWind:    WindSouth,
			IsDealer:    false,
			IsSelfDrawn: true,
			Hand: []string{
				"wan-1", "wan-2", "wan-3",
				"tong-5", "tong-5",
			},
			Melds: []model.Meld{
				{Type: "kong_concealed", Tiles: []string{"dong", "dong", "dong", "dong"}},
				{Type: "chow", Tiles: []string{"tiao-1", "tiao-2", "tiao-3"}},
				{Type: "chow", Tiles: []string{"wan-4", "wan-5", "wan-6"}},
				{Type: "pong", Tiles: []string{"tong-9", "tong-9", "tong-9"}},
			},
			Flowers:           []string{},
			WinningTile:       "tong-5",
			SpecialConditions: []SpecialCondition{ConditionKongBloom},
		}

		result := CalculateScoreWithInput(input)

		if !hasBreakdown(result, "KONG_BLOOM") {
			t.Error("應該識別槓上開花")
		}
	})

	t.Run("搶槓", func(t *testing.T) {
		input := &ScoreInput{
			RoundWind:   WindEast,
			SeatWind:    WindWest,
			IsDealer:    false,
			IsSelfDrawn: false, // 搶槓不算自摸
			Hand: []string{
				"wan-1", "wan-2", "wan-3",
				"wan-4", "wan-5", "wan-6",
				"tong-1", "tong-2", "tong-3",
				"tiao-1", "tiao-2", "tiao-3",
				"tong-5", "tong-5",
				"wan-7", "wan-8", "wan-9",
			},
			Melds:             []model.Meld{},
			Flowers:           []string{},
			WinningTile:       "tong-5",
			SpecialConditions: []SpecialCondition{ConditionRobbingKong},
		}

		result := CalculateScoreWithInput(input)

		if !hasBreakdown(result, "ROBBING_KONG") {
			t.Error("應該識別搶槓")
		}
	})
}

// TestWindKongCombined 測試風牌槓的綜合情況
func TestWindKongCombined(t *testing.T) {
	t.Run("東風局東家暗槓東風（場風刻+門風刻+暗槓+門清）", func(t *testing.T) {
		// 測試暗槓不破壞門清的情況
		// 手牌：完整的 17 張門清手牌，只有暗槓
		input := &ScoreInput{
			RoundWind:   WindEast,
			SeatWind:    WindEast,
			IsDealer:    true,
			IsSelfDrawn: true,
			Hand: []string{
				"wan-1", "wan-2", "wan-3", // 順子
				"wan-4", "wan-5", "wan-6", // 順子
				"tong-1", "tong-2", "tong-3", // 順子
				"tiao-7", "tiao-8", "tiao-9", // 順子
				"tong-5", "tong-5", // 將牌
			},
			Melds: []model.Meld{
				{Type: "kong_concealed", Tiles: []string{"dong", "dong", "dong", "dong"}}, // 東風暗槓
			},
			Flowers:     []string{},
			WinningTile: "tong-5",
		}

		result := CalculateScoreWithInput(input)

		// 應該有場風刻、門風刻、暗槓
		if !hasBreakdown(result, "ROUND_WIND_TRIPLET") {
			t.Error("應該識別場風刻")
		}
		if !hasBreakdown(result, "SEAT_WIND_TRIPLET") {
			t.Error("應該識別門風刻")
		}
		if !hasBreakdown(result, "KONG_CONCEALED") {
			t.Error("應該識別暗槓")
		}

		// 門清也應該保持（暗槓不破壞門清）
		if !hasBreakdown(result, "MEN_QING") {
			t.Error("暗槓不應該破壞門清")
		}

		// 計算總台數：自摸(1) + 門清(1) + 場風刻(1) + 門風刻(1) + 暗槓(2) = 6
		expectedTai := 6
		if result.TotalTai != expectedTai {
			t.Errorf("總台數應該是 %d，實際 %d", expectedTai, result.TotalTai)
		}
	})

	t.Run("東風局東家明槓東風（場風刻+門風刻+明槓，無門清）", func(t *testing.T) {
		input := &ScoreInput{
			RoundWind:   WindEast,
			SeatWind:    WindEast,
			IsDealer:    true,
			IsSelfDrawn: true,
			Hand: []string{
				"wan-1", "wan-2", "wan-3",
				"tong-5", "tong-5",
			},
			Melds: []model.Meld{
				{Type: "kong_exposed", Tiles: []string{"dong", "dong", "dong", "dong"}}, // 東風明槓
				{Type: "chow", Tiles: []string{"tiao-1", "tiao-2", "tiao-3"}},
				{Type: "chow", Tiles: []string{"wan-4", "wan-5", "wan-6"}},
				{Type: "pong", Tiles: []string{"tong-9", "tong-9", "tong-9"}},
			},
			Flowers:     []string{},
			WinningTile: "tong-5",
		}

		result := CalculateScoreWithInput(input)

		// 應該有場風刻、門風刻、明槓
		if !hasBreakdown(result, "ROUND_WIND_TRIPLET") {
			t.Error("應該識別場風刻")
		}
		if !hasBreakdown(result, "SEAT_WIND_TRIPLET") {
			t.Error("應該識別門風刻")
		}
		if !hasBreakdown(result, "KONG_EXPOSED") {
			t.Error("應該識別明槓")
		}

		// 明槓和吃碰會破壞門清
		if hasBreakdown(result, "MEN_QING") {
			t.Error("明槓和吃碰應該破壞門清")
		}

		// 計算總台數：自摸(1) + 場風刻(1) + 門風刻(1) + 明槓(1) = 4
		expectedTai := 4
		if result.TotalTai != expectedTai {
			t.Errorf("總台數應該是 %d，實際 %d", expectedTai, result.TotalTai)
		}
	})
}

// TestBackwardCompatibility 測試向後兼容
func TestBackwardCompatibility(t *testing.T) {
	t.Run("使用舊版 CalculateScore 函數", func(t *testing.T) {
		hand := []string{
			"wan-1", "wan-2", "wan-3",
			"tong-5", "tong-5",
		}
		melds := []model.Meld{
			{Type: "pong", Tiles: []string{"dong", "dong", "dong"}},
			{Type: "chow", Tiles: []string{"tiao-1", "tiao-2", "tiao-3"}},
			{Type: "chow", Tiles: []string{"wan-4", "wan-5", "wan-6"}},
			{Type: "pong", Tiles: []string{"tong-9", "tong-9", "tong-9"}},
		}
		flowers := []string{}

		result := CalculateScore(hand, melds, flowers, "tong-5", true)

		// 舊版函數仍應正常工作
		if result == nil {
			t.Error("CalculateScore 應該返回結果")
		}
		if len(result.HandTypes) == 0 {
			t.Error("應該有牌型")
		}
		if result.TotalTai <= 0 {
			t.Error("應該有台數")
		}

		// 預設會是東風東家，所以東風碰會同時算場風刻和門風刻
		if !hasHandType(result, "場風刻") {
			t.Error("預設東風局，東風碰應該算場風刻")
		}
		if !hasHandType(result, "門風刻") {
			t.Error("預設東家，東風碰應該算門風刻")
		}
	})
}
