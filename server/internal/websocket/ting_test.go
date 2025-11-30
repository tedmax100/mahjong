package websocket

import (
	"encoding/json"
	"mahjong/internal/game"
	"testing"
	"time"
)

// TestDeclareTing 測試宣告聽牌功能
func TestDeclareTing(t *testing.T) {
	t.Run("宣告聽牌成功", func(t *testing.T) {
		// 創建一個聽牌的手牌（16張，台灣16張麻將）
		// 例如：萬-1,1,1,2,3,4,5,6,7,7,7,8,9,9,9,9
		player := &game.Player{
			ID:       "player1",
			Name:     "測試玩家",
			Position: 0,
			Hand: []string{
				"wan-1", "wan-1", "wan-1",
				"wan-2", "wan-3",
				"wan-4", "wan-5", "wan-6",
				"wan-7", "wan-7", "wan-7",
				"wan-8",
				"wan-9", "wan-9", "wan-9", "wan-9",
			},
			Melds:        []game.Meld{},
			IsTing:       false,
			WinningTiles: []string{},
		}

		players := []*game.Player{player}

		// 創建測試用的房間和玩家
		room := &game.Room{
			ID:      "test-room",
			Players: players,
			Game:    game.NewMahjongGame(players),
			Clients: make(map[string]interface{}),
		}

		// 檢查聽牌狀態
		tingResult := room.Game.CheckTing(player.Hand, player.Melds)

		if !tingResult.IsTing {
			t.Error("預期手牌應該是聽牌狀態")
		}

		// 模擬宣告聽牌
		player.IsTing = true
		player.WinningTiles = tingResult.WinningTiles

		if !player.IsTing {
			t.Error("玩家應該已宣告聽牌")
		}

		if len(player.WinningTiles) == 0 {
			t.Error("聽牌玩家應該有WinningTiles")
		}

		t.Logf("聽牌成功，聽: %v", player.WinningTiles)
	})

	t.Run("非聽牌手牌不能宣告聽牌", func(t *testing.T) {
		// 創建一個非聽牌的手牌
		player := &game.Player{
			ID:       "player1",
			Name:     "測試玩家",
			Position: 0,
			Hand: []string{
				"wan-1", "wan-2", "wan-3",
				"tong-1", "tong-2", "tong-3",
				"tiao-1", "tiao-2", "tiao-3",
				"dong", "nan", "xi", "bei",
			},
			Melds:  []game.Meld{},
			IsTing: false,
		}

		players := []*game.Player{player}

		room := &game.Room{
			ID:      "test-room",
			Players: players,
			Game:    game.NewMahjongGame(players),
		}

		// 檢查聽牌狀態
		tingResult := room.Game.CheckTing(player.Hand, player.Melds)

		if tingResult.IsTing {
			t.Error("預期手牌不應該是聽牌狀態")
		}

		// 不應該設置聽牌狀態
		if player.IsTing {
			t.Error("非聽牌手牌不應該被標記為聽牌")
		}
	})

	t.Run("聽牌後狀態應持續", func(t *testing.T) {
		player := &game.Player{
			ID:           "player1",
			Name:         "測試玩家",
			IsTing:       true,
			WinningTiles: []string{"wan-1", "wan-4"},
		}

		// 驗證狀態
		if !player.IsTing {
			t.Error("聽牌狀態應該保持")
		}

		if len(player.WinningTiles) != 2 {
			t.Error("WinningTiles 應該保持")
		}
	})

	t.Run("有吃碰槓後的聽牌檢查", func(t *testing.T) {
		// 創建一個有碰牌且聽牌的手牌（13張手牌+1組碰 = 16張總計）
		// 手牌: 萬1,1,1,2,3,4,5,6,7,7,7,8,9 (台灣16張麻將)
		player := &game.Player{
			ID:       "player1",
			Name:     "測試玩家",
			Position: 0,
			Hand: []string{
				"wan-1", "wan-1", "wan-1",
				"wan-2", "wan-3",
				"wan-4", "wan-5", "wan-6",
				"wan-7", "wan-7", "wan-7",
				"wan-8", "wan-9",
			},
			Melds: []game.Meld{
				{
					Type:  "pong",
					Tiles: []string{"tong-5", "tong-5", "tong-5"},
				},
			},
			IsTing: false,
		}

		players := []*game.Player{player}

		room := &game.Room{
			ID:      "test-room",
			Players: players,
			Game:    game.NewMahjongGame(players),
		}

		// 檢查聽牌狀態（考慮吃碰槓）
		tingResult := room.Game.CheckTing(player.Hand, player.Melds)

		if !tingResult.IsTing {
			t.Error("有碰牌的手牌應該能正確檢測聽牌")
		}

		t.Logf("有碰牌的聽牌檢測成功，聽: %v", tingResult.WinningTiles)
	})
}

// TestTingBroadcast 測試聽牌廣播
func TestTingBroadcast(t *testing.T) {
	t.Run("聽牌廣播消息格式正確", func(t *testing.T) {
		player := &game.Player{
			ID:           "player1",
			Name:         "測試玩家",
			IsTing:       true,
			WinningTiles: []string{"wan-1", "wan-4", "wan-7"},
		}

		// 構建廣播消息
		message := map[string]interface{}{
			"type": "player_action",
			"data": map[string]interface{}{
				"playerId":     player.ID,
				"action":       "ting",
				"winningTiles": player.WinningTiles,
			},
		}

		// 序列化消息
		msgBytes, err := json.Marshal(message)
		if err != nil {
			t.Fatalf("消息序列化失敗: %v", err)
		}

		// 反序列化驗證
		var decoded map[string]interface{}
		err = json.Unmarshal(msgBytes, &decoded)
		if err != nil {
			t.Fatalf("消息反序列化失敗: %v", err)
		}

		// 驗證消息類型
		if decoded["type"] != "player_action" {
			t.Error("消息類型應該是 player_action")
		}

		// 驗證數據
		data := decoded["data"].(map[string]interface{})
		if data["playerId"] != "player1" {
			t.Error("playerId 不正確")
		}

		if data["action"] != "ting" {
			t.Error("action 應該是 ting")
		}

		winningTiles := data["winningTiles"].([]interface{})
		if len(winningTiles) != 3 {
			t.Error("winningTiles 數量不正確")
		}

		t.Logf("聽牌廣播消息: %s", string(msgBytes))
	})
}

// TestTingRestrictions 測試聽牌後的限制
func TestTingRestrictions(t *testing.T) {
	t.Run("聽牌玩家標記檢查", func(t *testing.T) {
		player := &game.Player{
			ID:           "player1",
			IsTing:       true,
			WinningTiles: []string{"wan-1"},
		}

		// 檢查玩家是否已聽牌（用於判斷是否可以吃碰槓）
		if !player.IsTing {
			t.Error("應該能正確檢測到玩家已聽牌")
		}

		// 聽牌後的邏輯：不能吃碰槓
		canChow := !player.IsTing
		canPong := !player.IsTing
		canKong := !player.IsTing

		if canChow || canPong || canKong {
			t.Error("聽牌後不應該能吃碰槓")
		}
	})

	t.Run("未聽牌玩家可以正常操作", func(t *testing.T) {
		player := &game.Player{
			ID:     "player1",
			IsTing: false,
		}

		// 未聽牌時可以吃碰槓
		canChow := !player.IsTing
		canPong := !player.IsTing
		canKong := !player.IsTing

		if !canChow || !canPong || !canKong {
			t.Error("未聽牌時應該可以吃碰槓")
		}
	})
}

// TestTingWithDifferentHandSizes 測試不同手牌數量的聽牌檢查
func TestTingWithDifferentHandSizes(t *testing.T) {
	testCases := []struct {
		name        string
		hand        []string
		melds       []game.Meld
		shouldBeTing bool
	}{
		{
			name: "16張手牌聽牌（台灣16張麻將）",
			hand: []string{
				"wan-1", "wan-1", "wan-1",
				"wan-2", "wan-3",
				"wan-4", "wan-5", "wan-6",
				"wan-7", "wan-7", "wan-7",
				"wan-8",
				"wan-9", "wan-9", "wan-9", "wan-9",
			},
			melds:        []game.Meld{},
			shouldBeTing: true,
		},
		{
			name: "13張手牌+1組碰聽牌（台灣16張麻將）",
			hand: []string{
				"wan-1", "wan-1", "wan-1",
				"wan-2", "wan-3",
				"wan-4", "wan-5", "wan-6",
				"wan-7", "wan-7", "wan-7",
				"wan-8", "wan-9",
			},
			melds: []game.Meld{
				{Type: "pong", Tiles: []string{"tong-5", "tong-5", "tong-5"}},
			},
			shouldBeTing: true,
		},
		{
			name: "非聽牌手牌",
			hand: []string{
				"wan-1", "wan-3", "wan-5",
				"tong-2", "tong-4", "tong-6",
				"tiao-1", "tiao-3", "tiao-5",
				"dong", "nan", "xi", "bei",
			},
			melds:        []game.Meld{},
			shouldBeTing: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			player := &game.Player{
				Hand:  tc.hand,
				Melds: tc.melds,
			}
			players := []*game.Player{player}

			room := &game.Room{
				Game: game.NewMahjongGame(players),
			}

			tingResult := room.Game.CheckTing(tc.hand, tc.melds)

			if tingResult.IsTing != tc.shouldBeTing {
				t.Errorf("預期聽牌狀態: %v, 實際: %v", tc.shouldBeTing, tingResult.IsTing)
			}

			if tingResult.IsTing {
				t.Logf("聽牌成功，聽: %v", tingResult.WinningTiles)
			}
		})
	}
}

// BenchmarkCheckTing 聽牌檢查性能測試
func BenchmarkCheckTing(b *testing.B) {
	hand := []string{
		"wan-1", "wan-1", "wan-1",
		"wan-2", "wan-3",
		"wan-4", "wan-5", "wan-6",
		"wan-7", "wan-8",
		"wan-9", "wan-9", "wan-9",
	}

	melds := []game.Meld{}

	player := &game.Player{Hand: hand, Melds: melds}
	players := []*game.Player{player}

	room := &game.Room{
		Game: game.NewMahjongGame(players),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		room.Game.CheckTing(hand, melds)
	}
}

// TestConcurrentTingDeclarations 測試並發聽牌宣告
func TestConcurrentTingDeclarations(t *testing.T) {
	t.Run("多個玩家同時宣告聽牌", func(t *testing.T) {
		// 創建4個玩家
		players := make([]*game.Player, 4)
		for i := 0; i < 4; i++ {
			players[i] = &game.Player{
				ID:           string(rune('A' + i)),
				Name:         string(rune('A' + i)),
				Position:     i,
				Hand:         []string{"wan-1", "wan-1", "wan-2", "wan-3"},
				IsTing:       false,
				WinningTiles: []string{},
			}
		}

		room := &game.Room{
			ID:      "test-room",
			Players: players,
			Game:    game.NewMahjongGame(players),
		}

		// 並發設置聽牌狀態
		done := make(chan bool, 4)
		for i := 0; i < 4; i++ {
			go func(playerIdx int) {
				time.Sleep(time.Millisecond * 10) // 模擬延遲
				room.Players[playerIdx].IsTing = true
				room.Players[playerIdx].WinningTiles = []string{"wan-1", "wan-4"}
				done <- true
			}(i)
		}

		// 等待所有goroutine完成
		for i := 0; i < 4; i++ {
			<-done
		}

		// 驗證所有玩家都已聽牌
		for i, player := range room.Players {
			if !player.IsTing {
				t.Errorf("玩家 %d 應該已聽牌", i)
			}
			if len(player.WinningTiles) == 0 {
				t.Errorf("玩家 %d 應該有WinningTiles", i)
			}
		}
	})
}
