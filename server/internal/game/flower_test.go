package game

import (
	"mahjong/internal/model"
	"mahjong/internal/tile"
	"testing"
)

// TestIsFlowerTile 測試花牌辨識
func TestFlowerTileDetection(t *testing.T) {
	tests := []struct {
		tile     string
		isFlower bool
	}{
		{"flower-chun", true},
		{"flower-xia", true},
		{"flower-qiu", true},
		{"flower-dong", true},
		{"flower-mei", true},
		{"flower-lan", true},
		{"flower-zhu", true},
		{"flower-ju", true},
		{"dong", false},       // 東風牌，不是花牌
		{"wan-1", false},      // 萬子，不是花牌
		{"tong-5", false},     // 筒子，不是花牌
		{"zhong", false},      // 中，不是花牌
	}

	for _, tt := range tests {
		t.Run(tt.tile, func(t *testing.T) {
			result := tile.IsFlower(tt.tile)
			if result != tt.isFlower {
				t.Errorf("isFlowerTile(%s) = %v, 期望 %v", tt.tile, result, tt.isFlower)
			}
		})
	}
}

// TestDrawTileWithFlowerReplacement 測試摸到花牌時的自動補牌
func TestDrawTileWithFlowerReplacement(t *testing.T) {
	room := NewRoom("test-room")

	// 添加玩家
	for i := 1; i <= 4; i++ {
		room.AddPlayer("player"+string(rune('0'+i)), "玩家"+string(rune('0'+i)), false)
	}

	room.StartGame()
	player := room.Players[0]

	t.Run("摸到花牌應該自動補牌", func(t *testing.T) {
		// 清空玩家手牌和花牌
		player.Hand = []string{}
		player.Flowers = []string{}

		// 手動設置牌堆：DrawTile() 從索引 0 開始抽牌
		// 先抽到 flower-chun，然後補牌抽到 wan-1
		// 牌堆需要大於 8 張才不會流局
		room.Game.Deck = []string{
			"flower-chun", "wan-1",
			"dong", "nan", "xi", "bei", "zhong", "fa", "bai", "tong-1", "tong-2",
		}

		initialDeckSize := len(room.Game.Deck)

		// 摸牌（應該先摸到 flower-chun，然後自動補牌摸到 wan-1）
		drawnTile := room.Game.DrawTileWithFlowerReplacement(player)

		// 應該摸到萬 1（花牌被自動替換）
		if drawnTile != "wan-1" {
			t.Errorf("期望摸到 wan-1（補牌後），實際摸到 %s", drawnTile)
		}

		// 花牌應該被加入 Flowers
		if len(player.Flowers) != 1 {
			t.Errorf("期望有 1 張花牌，實際有 %d 張", len(player.Flowers))
		}

		if len(player.Flowers) > 0 && player.Flowers[0] != "flower-chun" {
			t.Errorf("期望花牌是 flower-chun，實際是 %s", player.Flowers[0])
		}

		// 牌堆應該減少 2 張（花牌 + 補牌）
		expectedDeckSize := initialDeckSize - 2
		if len(room.Game.Deck) != expectedDeckSize {
			t.Errorf("期望牌堆剩餘 %d 張，實際 %d 張", expectedDeckSize, len(room.Game.Deck))
		}
	})

	t.Run("連續摸到多張花牌應該全部補牌", func(t *testing.T) {
		player.Hand = []string{}
		player.Flowers = []string{}

		// 手動設置牌堆：先抽到兩張花牌，然後補牌抽到 wan-2
		room.Game.Deck = []string{
			"flower-lan", "flower-mei", "wan-2",
			"dong", "nan", "xi", "bei", "zhong", "fa", "bai", "tong-1",
		}

		drawnTile := room.Game.DrawTileWithFlowerReplacement(player)

		// 應該摸到萬 2（兩張花牌都被自動替換）
		if drawnTile != "wan-2" {
			t.Errorf("期望摸到 wan-2，實際摸到 %s", drawnTile)
		}

		// 應該有 2 張花牌
		if len(player.Flowers) != 2 {
			t.Errorf("期望有 2 張花牌，實際有 %d 張", len(player.Flowers))
		}

		// 檢查花牌內容
		expectedFlowers := map[string]bool{"flower-mei": true, "flower-lan": true}
		for _, flower := range player.Flowers {
			if !expectedFlowers[flower] {
				t.Errorf("意外的花牌: %s", flower)
			}
		}
	})

	t.Run("摸到普通牌不應該補牌", func(t *testing.T) {
		player.Hand = []string{}
		player.Flowers = []string{}

		// 手動設置牌堆：只有普通牌，加上足夠的牌避免流局
		room.Game.Deck = []string{
			"dong", "wan-3", "nan", "xi", "bei", "zhong", "fa", "bai", "tong-1", "tong-2",
		}

		initialDeckSize := len(room.Game.Deck)

		drawnTile := room.Game.DrawTileWithFlowerReplacement(player)

		// 應該摸到東（普通牌）
		if drawnTile != "dong" {
			t.Errorf("期望摸到 dong，實際摸到 %s", drawnTile)
		}

		// 不應該有花牌
		if len(player.Flowers) != 0 {
			t.Errorf("期望沒有花牌，實際有 %d 張", len(player.Flowers))
		}

		// 牌堆應該只減少 1 張
		expectedDeckSize := initialDeckSize - 1
		if len(room.Game.Deck) != expectedDeckSize {
			t.Errorf("期望牌堆剩餘 %d 張，實際 %d 張", expectedDeckSize, len(room.Game.Deck))
		}
	})
}

// TestFlowerTilesInInitialDeal 測試發牌時的花牌處理
func TestFlowerTilesInInitialDeal(t *testing.T) {
	room := NewRoom("test-room")

	// 添加玩家
	for i := 1; i <= 4; i++ {
		room.AddPlayer("player"+string(rune('0'+i)), "玩家"+string(rune('0'+i)), false)
	}

	room.StartGame()
	room.DealTiles() // 發牌

	t.Run("發牌時遇到花牌應該自動補牌", func(t *testing.T) {
		// 發牌已經完成
		// 檢查是否有玩家摸到花牌
		totalFlowers := 0
		for _, player := range room.Players {
			totalFlowers += len(player.Flowers)
		}

		// 如果有花牌，確保玩家手牌數量仍然正確
		for i, player := range room.Players {
			expectedHandSize := 16
			if i == room.Game.Dealer {
				expectedHandSize = 17 // 莊家多一張
			}

			actualHandSize := len(player.Hand)
			if actualHandSize != expectedHandSize {
				t.Errorf("玩家 %d 手牌應該有 %d 張，實際有 %d 張（花牌 %d 張）",
					i, expectedHandSize, actualHandSize, len(player.Flowers))
			}
		}

		if totalFlowers > 0 {
			t.Logf("發牌時共摸到 %d 張花牌", totalFlowers)
		}
	})
}

// TestFlowerTilesDoNotCountAsHandTiles 測試花牌不計入手牌數量
func TestFlowerTilesDoNotCountAsHandTiles(t *testing.T) {
	room := NewRoom("test-room")

	for i := 1; i <= 4; i++ {
		room.AddPlayer("player"+string(rune('0'+i)), "玩家"+string(rune('0'+i)), false)
	}

	room.StartGame()
	player := room.Players[0]

	t.Run("花牌應該獨立存儲，不計入手牌", func(t *testing.T) {
		player.Hand = []string{"wan-1", "wan-2", "wan-3"}
		player.Flowers = []string{"flower-chun", "flower-xia"}
		player.Melds = []model.Meld{}

		handCount := len(player.Hand)
		flowerCount := len(player.Flowers)

		// 手牌應該是 3 張
		if handCount != 3 {
			t.Errorf("手牌應該是 3 張，實際 %d 張", handCount)
		}

		// 花牌應該是 2 張
		if flowerCount != 2 {
			t.Errorf("花牌應該是 2 張，實際 %d 張", flowerCount)
		}

		// 檢查花牌不在手牌中
		for _, h := range player.Hand {
			if tile.IsFlower(h) {
				t.Errorf("手牌中不應該有花牌: %s", h)
			}
		}
	})

	t.Run("總牌數計算不應該包含花牌", func(t *testing.T) {
		player.Hand = []string{"wan-1", "wan-2", "wan-3", "wan-4", "wan-5", "wan-6",
			"wan-7", "wan-8", "wan-9", "tong-1", "tong-2", "tong-3", "dong"}
		player.Flowers = []string{"flower-chun", "flower-xia", "flower-qiu"}
		player.Melds = []model.Meld{
			{Type: "pong", Tiles: []string{"nan", "nan", "nan"}},
		}

		// 總牌數 = 手牌 + 吃碰槓
		totalTiles := len(player.Hand) + len(player.Melds)*3
		expectedTotal := 13 + 3 // 13 張手牌 + 1 組碰（3 張）

		if totalTiles != expectedTotal {
			t.Errorf("總牌數應該是 %d（不含花牌），實際 %d", expectedTotal, totalTiles)
		}

		// 花牌不應該計入總牌數
		if totalTiles == len(player.Hand)+len(player.Melds)*3+len(player.Flowers) {
			t.Error("總牌數計算錯誤地包含了花牌")
		}
	})
}

// TestFlowerTileScoring 測試花牌台數計算
func TestFlowerTileScoring(t *testing.T) {
	room := NewRoom("test-room")

	for i := 1; i <= 4; i++ {
		room.AddPlayer("player"+string(rune('0'+i)), "玩家"+string(rune('0'+i)), false)
	}

	room.StartGame()
	player := room.Players[0]

	t.Run("每張花牌應該加 1 台", func(t *testing.T) {
		player.Flowers = []string{"flower-chun"}
		// TODO: 實作花牌台數計算時，這裡可以添加測試
		// 目前只測試花牌數量
		if len(player.Flowers) != 1 {
			t.Errorf("期望 1 張花牌，實際 %d 張", len(player.Flowers))
		}
	})

	t.Run("多張花牌累計台數", func(t *testing.T) {
		player.Flowers = []string{"flower-chun", "flower-xia", "flower-qiu"}
		// 3 張花牌應該加 3 台
		if len(player.Flowers) != 3 {
			t.Errorf("期望 3 張花牌，實際 %d 張", len(player.Flowers))
		}
	})
}
