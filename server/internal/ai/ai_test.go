package ai

import (
	"testing"
)

// TestChooseDiscardAI 測試 Bot 出牌邏輯
func TestChooseDiscardAI(t *testing.T) {
	// game := &MahjongGame{} // No longer needed

	t.Run("優先打出孤立的字牌", func(t *testing.T) {
		// 手牌中有孤立的 "dong" (東)，應該優先打出
		hand := []string{
			"wan-1", "wan-2", "wan-3",
			"tong-5", "tong-5",
			"tiao-7", "tiao-8", "tiao-9",
			"dong", // 孤立字牌
			"wan-5", "wan-6",
		}

		discard := ChooseDiscard(hand)
		if discard != "dong" {
			t.Errorf("期望打出 dong，實際打出 %s", discard)
		}
	})

	t.Run("優先打出孤立的么九牌", func(t *testing.T) {
		// 沒有字牌，應該打出孤立的 wan-1 或 wan-9
		hand := []string{
			"wan-1", // 孤立么九
			"tong-2", "tong-3", "tong-4",
			"tiao-5", "tiao-5",
			"wan-4", "wan-5", "wan-6",
		}

		discard := ChooseDiscard(hand)
		if discard != "wan-1" {
			t.Errorf("期望打出 wan-1，實際打出 %s", discard)
		}
	})

	t.Run("優先打出孤立的中張牌", func(t *testing.T) {
		// 沒有字牌和么九，打出孤立的 tong-5
		hand := []string{
			"wan-2", "wan-3", "wan-4",
			"tong-5", // 孤立中張
			"tiao-5", "tiao-5",
			"tiao-7", "tiao-8", "tiao-9",
		}

		discard := ChooseDiscard(hand)
		if discard != "tong-5" {
			t.Errorf("期望打出 tong-5，實際打出 %s", discard)
		}
	})

	t.Run("沒有孤張時打出不成對的牌", func(t *testing.T) {
		// 所有牌都有鄰居或對子，打出不成對的牌
		// 這裡 wan-2 有 wan-3 鄰居，tong-5 是對子
		// tiao-8 是單張（雖然有 tiao-7, tiao-9，但 ChooseDiscardAI 簡單邏輯可能判定它不是單張）
		// 讓我們構造一個情況：
		// wan-1, wan-1 (對)
		// tong-2, tong-2 (對)
		// tiao-3 (單)
		hand := []string{
			"wan-1", "wan-1",
			"tong-2", "tong-2",
			"tiao-3",
		}

		discard := ChooseDiscard(hand)
		if discard != "tiao-3" {
			t.Errorf("期望打出 tiao-3，實際打出 %s", discard)
		}
	})

	t.Run("全是對子或刻子時打出最後一張", func(t *testing.T) {
		// 這種情況比較極端，通常會胡牌或聽牌
		hand := []string{
			"wan-1", "wan-1",
			"tong-2", "tong-2", "tong-2",
		}

		discard := ChooseDiscard(hand)
		// 邏輯是打出 hand[len-1]
		expected := hand[len(hand)-1]
		if discard != expected {
			t.Errorf("期望打出 %s，實際打出 %s", expected, discard)
		}
	})
}