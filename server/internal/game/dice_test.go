package game

import (
	"testing"
)

func TestRollDiceForDealer(t *testing.T) {
	// 創建測試玩家
	players := []*Player{
		{ID: "player0", Name: "東家", Position: 0},
		{ID: "player1", Name: "南家", Position: 1},
		{ID: "player2", Name: "西家", Position: 2},
		{ID: "player3", Name: "北家", Position: 3},
	}

	// 執行多次測試以確保隨機性和正確性
	for i := 0; i < 100; i++ {
		result := RollDiceForDealer(players)

		// 驗證骰子結果在有效範圍內
		for j, dice := range result.DiceResults {
			if dice < 1 || dice > 6 {
				t.Errorf("骰子 %d 的值 %d 不在有效範圍 1-6 內", j, dice)
			}
		}

		// 驗證總和計算正確
		expectedSum := result.DiceResults[0] + result.DiceResults[1] + result.DiceResults[2]
		if result.TotalSum != expectedSum {
			t.Errorf("總和計算錯誤: 預期 %d, 實際 %d", expectedSum, result.TotalSum)
		}

		// 驗證總和在有效範圍內 (3-18)
		if result.TotalSum < 3 || result.TotalSum > 18 {
			t.Errorf("總和 %d 不在有效範圍 3-18 內", result.TotalSum)
		}

		// 驗證莊家位置計算正確
		expectedDealer := (result.TotalSum - 1) % 4
		if result.DealerSeatIndex != expectedDealer {
			t.Errorf("莊家位置計算錯誤: 總和 %d 預期位置 %d, 實際 %d",
				result.TotalSum, expectedDealer, result.DealerSeatIndex)
		}

		// 驗證莊家位置在有效範圍內
		if result.DealerSeatIndex < 0 || result.DealerSeatIndex > 3 {
			t.Errorf("莊家位置 %d 不在有效範圍 0-3 內", result.DealerSeatIndex)
		}

		// 驗證莊家 ID 正確對應
		if result.DealerPlayerID != players[result.DealerSeatIndex].ID {
			t.Errorf("莊家 ID 不匹配: 預期 %s, 實際 %s",
				players[result.DealerSeatIndex].ID, result.DealerPlayerID)
		}
	}
}

func TestDealerPositionFormula(t *testing.T) {
	// 台灣麻將莊家計算規則測試
	// 從東位開始逆時針數
	testCases := []struct {
		sum              int
		expectedPosition int
		description      string
	}{
		{3, 2, "3點 -> 西家 (位置2)"},
		{4, 3, "4點 -> 北家 (位置3)"},
		{5, 0, "5點 -> 東家 (位置0)"},
		{6, 1, "6點 -> 南家 (位置1)"},
		{7, 2, "7點 -> 西家 (位置2)"},
		{8, 3, "8點 -> 北家 (位置3)"},
		{9, 0, "9點 -> 東家 (位置0)"},
		{10, 1, "10點 -> 南家 (位置1)"},
		{11, 2, "11點 -> 西家 (位置2)"},
		{12, 3, "12點 -> 北家 (位置3)"},
		{13, 0, "13點 -> 東家 (位置0)"},
		{14, 1, "14點 -> 南家 (位置1)"},
		{15, 2, "15點 -> 西家 (位置2)"},
		{16, 3, "16點 -> 北家 (位置3)"},
		{17, 0, "17點 -> 東家 (位置0)"},
		{18, 1, "18點 -> 南家 (位置1)"},
	}

	for _, tc := range testCases {
		result := (tc.sum - 1) % 4
		if result != tc.expectedPosition {
			t.Errorf("%s: 預期 %d, 實際 %d", tc.description, tc.expectedPosition, result)
		}
	}
}
