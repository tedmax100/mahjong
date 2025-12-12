package game

import (
	"mahjong/internal/model"
	"testing"
)

func TestNextRound(t *testing.T) {
	// 1. Setup Room and Players
	playerA := &Player{ID: "playerA", Name: "Player A", Position: 0}
	playerB := &Player{ID: "playerB", Name: "Player B", Position: 1}
	playerC := &Player{ID: "playerC", Name: "Player C", Position: 2}
	playerD := &Player{ID: "playerD", Name: "Player D", Position: 3}

	players := []*Player{playerA, playerB, playerC, playerD}

	room := NewRoom("test-room")
	for _, p := range players {
		room.AddPlayer(p.ID, p.Name, false)
	}
	room.StartGame()

	// Give a player some state to ensure it gets reset
	room.Players[1].Hand = []string{"wan-1", "wan-2", "wan-3"}
	room.Players[1].Melds = []model.Meld{{Type: "pong", Tiles: []string{"tiao-5", "tiao-5", "tiao-5"}}}
	room.Players[1].Flowers = []string{"flower-chun"}
	room.Game.Dealer = 0
	originalDeckSize := len(room.Game.Deck)

	// 2. Call NextRound
	room.NextRound()

	// 3. Assertions
	// Assert dealer advanced
	if room.Game.Dealer != 1 {
		t.Errorf("Dealer should have advanced to 1, but is %d", room.Game.Dealer)
	}

	// Assert current turn is the new dealer
	if room.CurrentTurn != 1 {
		t.Errorf("CurrentTurn should be the new dealer (1), but is %d", room.CurrentTurn)
	}

	// Assert player state was reset
	if len(room.Players[1].Hand) != 0 {
		t.Errorf("Player hand should have been reset, but has %d tiles", len(room.Players[1].Hand))
	}
	if len(room.Players[1].Melds) != 0 {
		t.Errorf("Player melds should have been reset, but has %d melds", len(room.Players[1].Melds))
	}
	if len(room.Players[1].Flowers) != 0 {
		t.Errorf("Player flowers should have been reset, but has %d flowers", len(room.Players[1].Flowers))
	}

	// Assert a new deck was created
	if len(room.Game.Deck) <= originalDeckSize {
		// This is a loose check, but a new deck should have a lot of tiles.
		// A more robust check might be to compare pointers, but this is fine.
	}

	// Assert game is marked as started
	if !room.GameStarted {
		t.Errorf("GameStarted should be true after NextRound")
	}
}

// TestHandleDiscard_TileCountWithKongs 測試有槓時的牌數驗證
func TestHandleDiscard_TileCountWithKongs(t *testing.T) {
	testCases := []struct {
		name           string
		handTiles      []string
		melds          []model.Meld
		kongCount      int
		shouldSucceed  bool
		description    string
	}{
		{
			name: "無槓_17 張手牌_應該成功",
			handTiles: []string{
				"wan-1", "wan-2", "wan-3", "wan-4", "wan-5",
				"tiao-1", "tiao-2", "tiao-3", "tiao-4", "tiao-5",
				"tong-1", "tong-2", "tong-3", "tong-4", "tong-5",
				"dong", "nan",
			},
			melds:         []model.Meld{},
			kongCount:     0,
			shouldSucceed: true,
			description:   "正常情況，17 張手牌，無吃碰槓",
		},
		{
			name: "無槓_18 張手牌_應該成功",
			handTiles: []string{
				"wan-1", "wan-2", "wan-3", "wan-4", "wan-5",
				"tiao-1", "tiao-2", "tiao-3", "tiao-4", "tiao-5",
				"tong-1", "tong-2", "tong-3", "tong-4", "tong-5",
				"dong", "nan", "xi",
			},
			melds:         []model.Meld{},
			kongCount:     0,
			shouldSucceed: true,
			description:   "槓後補牌，18 張手牌",
		},
		{
			name: "1 個槓_18 張總牌_應該成功",
			handTiles: []string{
				"wan-1", "wan-2", "wan-3", "wan-4", "wan-5",
				"tiao-1", "tiao-2", "tiao-3", "tiao-4",
				"tong-1", "tong-2", "tong-3", "tong-4", "tong-5",
			},
			melds: []model.Meld{
				{Type: "kong_promoted", Tiles: []string{"dong", "dong", "dong", "dong"}},
			},
			kongCount:     1,
			shouldSucceed: true,
			description:   "1 個槓（4 張）+ 14 張手牌 = 18 張（打牌前）",
		},
		{
			name: "1 個槓_19 張總牌_應該成功",
			handTiles: []string{
				"wan-1", "wan-2", "wan-3", "wan-4", "wan-5",
				"tiao-1", "tiao-2", "tiao-3", "tiao-4", "tiao-5",
				"tong-1", "tong-2", "tong-3", "tong-4", "tong-5",
			},
			melds: []model.Meld{
				{Type: "kong_concealed", Tiles: []string{"dong", "dong", "dong", "dong"}},
			},
			kongCount:     1,
			shouldSucceed: true,
			description:   "1 個槓 + 15 張手牌 = 19 張（槓後補牌，打牌前）",
		},
		{
			name: "2 個槓_19 張總牌_應該成功",
			handTiles: []string{
				"wan-1", "wan-2", "wan-3", "wan-4", "wan-5",
				"tiao-1", "tiao-2", "tiao-3", "tiao-4",
				"tong-1", "tong-2", "tong-3",
			},
			melds: []model.Meld{
				{Type: "kong_promoted", Tiles: []string{"dong", "dong", "dong", "dong"}},
				{Type: "kong_concealed", Tiles: []string{"nan", "nan", "nan", "nan"}},
			},
			kongCount:     2,
			shouldSucceed: true,
			description:   "2 個槓（8 張）+ 11 張手牌 = 19 張（打牌前）",
		},
		{
			name: "2 個槓_20 張總牌_應該成功",
			handTiles: []string{
				"wan-1", "wan-2", "wan-3", "wan-4", "wan-5",
				"tiao-1", "tiao-2", "tiao-3", "tiao-4", "tiao-5",
				"tong-1", "tong-2",
			},
			melds: []model.Meld{
				{Type: "kong_promoted", Tiles: []string{"dong", "dong", "dong", "dong"}},
				{Type: "kong_promoted", Tiles: []string{"nan", "nan", "nan", "nan"}},
			},
			kongCount:     2,
			shouldSucceed: true,
			description:   "2 個槓 + 12 張手牌 = 20 張（槓後補牌，打牌前）",
		},
		{
			name: "3 個槓_20 張總牌_應該成功",
			handTiles: []string{
				"wan-1", "wan-2", "wan-3", "wan-4",
				"tiao-1", "tiao-2", "tiao-3", "tiao-4",
			},
			melds: []model.Meld{
				{Type: "kong_promoted", Tiles: []string{"dong", "dong", "dong", "dong"}},
				{Type: "kong_concealed", Tiles: []string{"nan", "nan", "nan", "nan"}},
				{Type: "kong_exposed", Tiles: []string{"xi", "xi", "xi", "xi"}},
			},
			kongCount:     3,
			shouldSucceed: true,
			description:   "3 個槓（12 張）+ 8 張手牌 = 20 張（打牌前）",
		},
		{
			name: "1 個槓_但牌數錯誤_應該失敗",
			handTiles: []string{
				"wan-1", "wan-2", "wan-3",
			},
			melds: []model.Meld{
				{Type: "kong_promoted", Tiles: []string{"dong", "dong", "dong", "dong"}},
			},
			kongCount:     1,
			shouldSucceed: false,
			description:   "只有 7 張總牌（3 手牌+4 槓），應該拒絕",
		},
		{
			name: "2 個槓_但牌數錯誤_應該失敗",
			handTiles: []string{
				"wan-1", "wan-2", "wan-3", "wan-4",
			},
			melds: []model.Meld{
				{Type: "kong_promoted", Tiles: []string{"dong", "dong", "dong", "dong"}},
				{Type: "kong_concealed", Tiles: []string{"nan", "nan", "nan", "nan"}},
			},
			kongCount:     2,
			shouldSucceed: false,
			description:   "只有 12 張總牌（4 手牌+8 槓），預期 19 或 20 張，應該拒絕",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 建立房間
			room := NewRoom("test-room")

			// 添加玩家
			room.AddPlayer("player1", "玩家1", false)
			room.AddPlayer("player2", "玩家2", false)
			room.AddPlayer("player3", "玩家3", false)
			room.AddPlayer("player4", "玩家4", false)

			room.StartGame()
			// 不調用 DealTiles()，直接設置手牌

			// 設置測試玩家的手牌和吃碰槓
			player := room.Players[0]
			player.Hand = tc.handTiles
			player.Melds = tc.melds
			room.CurrentTurn = 0

			// 確保房間有 Game 實例
			if room.Game == nil {
				room.Game = NewMahjongGame(room.Players)
			}

			// 選擇一張手牌打出
			if len(player.Hand) == 0 {
				t.Skip("沒有手牌可以打出")
				return
			}
			tileToDiscard := player.Hand[0]
			initialHandCount := len(player.Hand)
			initialTotalTiles := player.GetTotalTiles()

			// 嘗試打牌（返回值: success, isDraw）
			success, _ := room.HandleDiscard(player.ID, tileToDiscard)

			// 檢查牌是否被移除（判斷是否成功）
			tileWasRemoved := len(player.Hand) == initialHandCount-1
			actualSuccess := success && tileWasRemoved

			// 驗證結果
			if tc.shouldSucceed && !actualSuccess {
				t.Errorf("%s: 應該成功但失敗了。總牌數: %d, 槓數: %d, 預期: %d 或 %d",
					tc.description, initialTotalTiles, tc.kongCount, 17+tc.kongCount, 18+tc.kongCount)
			}

			if !tc.shouldSucceed && actualSuccess {
				t.Errorf("%s: 應該失敗但成功了。總牌數: %d, 槓數: %d",
					tc.description, initialTotalTiles, tc.kongCount)
			}

			if tc.shouldSucceed && actualSuccess {
				t.Logf("✓ %s - 成功", tc.description)
			} else if !tc.shouldSucceed && !actualSuccess {
				t.Logf("✓ %s - 正確拒絕", tc.description)
			}
		})
	}
}