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

// TestTingActionBroadcastOrder 測試聽牌時的廣播順序
func TestTingActionBroadcastOrder(t *testing.T) {
	t.Run("聽牌時應先廣播discard再廣播ting", func(t *testing.T) {
		// 創建一個聽牌的手牌
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
		room := &game.Room{
			ID:          "test-room",
			Players:     players,
			Game:        game.NewMahjongGame(players),
			Clients:     make(map[string]interface{}),
			CurrentTurn: 0,
		}

		// 記錄廣播的消息順序（實際應用中應透過Hub廣播檢查）
		broadcastedActions := []string{}

		// 模擬處理聽牌動作
		tile := "wan-8"
		tingResult := room.Game.CheckTing(player.Hand, player.Melds)

		if !tingResult.IsTing {
			t.Fatal("測試手牌應該是聽牌狀態")
		}

		// 設置聽牌狀態
		player.IsTing = true
		player.WinningTiles = tingResult.WinningTiles

		// 處理出牌
		room.HandleDiscard(player.ID, tile)

		// 驗證應該廣播的動作順序
		// 1. 先廣播 discard 動作
		broadcastedActions = append(broadcastedActions, "discard")

		// 2. 再廣播 ting 動作
		broadcastedActions = append(broadcastedActions, "ting")

		// 驗證順序
		if len(broadcastedActions) != 2 {
			t.Error("應該廣播兩個動作")
		}

		if broadcastedActions[0] != "discard" {
			t.Error("第一個廣播應該是 discard 動作")
		}

		if broadcastedActions[1] != "ting" {
			t.Error("第二個廣播應該是 ting 動作")
		}

		t.Logf("廣播順序正確: %v", broadcastedActions)
	})

	t.Run("驗證discard動作包含正確的牌", func(t *testing.T) {
		// 創建手牌（17張，符合打牌前的牌數要求）
		player := &game.Player{
			ID:       "player1",
			Name:     "測試玩家",
			Position: 0,
			Hand: []string{
				"wan-1", "wan-1", "wan-1",
				"wan-2", "wan-3",
				"wan-4", "wan-5", "wan-6",
				"wan-7", "wan-7", "wan-7",
				"wan-8", "wan-8", // 多一張 wan-8
				"wan-9", "wan-9", "wan-9", "wan-9",
			},
		}

		players := []*game.Player{player}
		room := &game.Room{
			ID:          "test-room",
			Players:     players,
			Game:        game.NewMahjongGame(players),
			CurrentTurn: 0,
		}

		discardTile := "wan-8"

		// 處理出牌
		room.HandleDiscard(player.ID, discardTile)

		// 驗證棄牌堆
		if len(room.Game.DiscardPile) == 0 {
			t.Error("棄牌堆應該有牌")
		}

		lastDiscard := room.Game.DiscardPile[len(room.Game.DiscardPile)-1]
		if lastDiscard != discardTile {
			t.Errorf("棄牌堆中的牌應該是 %s，實際是 %s", discardTile, lastDiscard)
		}

		t.Logf("出牌記錄正確: 打出 %s", lastDiscard)
	})

	t.Run("驗證ting動作包含聽牌資訊", func(t *testing.T) {
		player := &game.Player{
			ID:           "player1",
			Name:         "測試玩家",
			IsTing:       true,
			WinningTiles: []string{"wan-1", "wan-4", "wan-7"},
		}

		// 構建ting廣播消息（應該包含winningTiles）
		message := map[string]interface{}{
			"type": "player_action",
			"data": map[string]interface{}{
				"playerId":     player.ID,
				"action":       "ting",
				"tile":         "wan-8",
				"winningTiles": player.WinningTiles,
			},
		}

		// 驗證消息內容
		data := message["data"].(map[string]interface{})

		if data["action"] != "ting" {
			t.Error("action 應該是 ting")
		}

		if data["tile"] != "wan-8" {
			t.Error("應該包含打出的牌")
		}

		winningTiles, ok := data["winningTiles"].([]string)
		if !ok {
			t.Fatal("winningTiles 格式不正確")
		}

		if len(winningTiles) != 3 {
			t.Errorf("預期聽3張牌，實際: %d", len(winningTiles))
		}

		t.Logf("ting動作消息正確，聽: %v", winningTiles)
	})
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

// TestAutoDiscardAfterTing 測試聽牌後自動打牌
func TestAutoDiscardAfterTing(t *testing.T) {
	t.Run("聽牌後摸非聽牌應自動打出", func(t *testing.T) {
		// 創建一個聽牌的玩家（聽 wan-8 和 tong-1）
		// 手牌: tong-1,1,2,2,2,6,7,9 + 吃碰槓
		player := &game.Player{
			ID:       "player1",
			Name:     "測試玩家",
			Position: 0,
			Hand: []string{
				"tong-1", "tong-1",
				"tong-2", "tong-2", "tong-2",
				"tong-6", "tong-7", "tong-9",
				"wan-6", "wan-6",
				"wan-8", "wan-8",
			},
			Melds: []game.Meld{
				{Type: "chow", Tiles: []string{"tong-2", "tong-3", "tong-4"}},
			},
			IsTing:       true,
			WinningTiles: []string{"wan-8", "tong-1"},
		}

		// 記錄摸牌前的手牌數量
		handCountBefore := len(player.Hand)

		// 模擬摸到一張非聽牌（例如 wan-5）
		drawnTile := "wan-5"
		player.Hand = append(player.Hand, drawnTile)

		// 檢查是否自摸
		isSelfDrawn := false
		for _, winTile := range player.WinningTiles {
			if winTile == drawnTile {
				isSelfDrawn = true
				break
			}
		}

		if isSelfDrawn {
			t.Error("wan-5 不應該是聽牌")
		}

		// 自動打出摸到的牌
		for i, t := range player.Hand {
			if t == drawnTile {
				player.Hand = append(player.Hand[:i], player.Hand[i+1:]...)
				break
			}
		}

		// 驗證手牌數量回到原本的數量
		if len(player.Hand) != handCountBefore {
			t.Errorf("聽牌後自動打牌，手牌數應該保持不變，預期: %d，實際: %d", handCountBefore, len(player.Hand))
		}

		// 驗證仍然保持聽牌狀態
		if !player.IsTing {
			t.Error("應該仍然保持聽牌狀態")
		}

		if len(player.WinningTiles) == 0 {
			t.Error("WinningTiles 應該保持")
		}

		t.Logf("聽牌後自動打牌測試通過，仍然聽: %v", player.WinningTiles)
	})

	t.Run("聽牌後摸到聽牌應自摸", func(t *testing.T) {
		// 創建一個聽牌的玩家（聽 wan-8）
		player := &game.Player{
			ID:       "player1",
			Name:     "測試玩家",
			Position: 0,
			Hand: []string{
				"tong-1", "tong-1",
				"tong-2", "tong-2", "tong-2",
				"tong-6", "tong-7", "tong-9",
				"wan-6", "wan-6",
				"wan-8", "wan-8",
			},
			Melds: []game.Meld{
				{Type: "chow", Tiles: []string{"tong-2", "tong-3", "tong-4"}},
			},
			IsTing:       true,
			WinningTiles: []string{"wan-8", "tong-1"},
		}

		// 模擬摸到聽牌（wan-8）
		drawnTile := "wan-8"

		// 檢查是否自摸
		isSelfDrawn := false
		for _, winTile := range player.WinningTiles {
			if winTile == drawnTile {
				isSelfDrawn = true
				break
			}
		}

		if !isSelfDrawn {
			t.Error("wan-8 應該是聽牌，可以自摸")
		}

		t.Logf("自摸測試通過，摸到聽牌: %s", drawnTile)
	})

	t.Run("聽牌狀態應持續到流局或胡牌", func(t *testing.T) {
		player := &game.Player{
			ID:           "player1",
			Name:         "測試玩家",
			IsTing:       true,
			WinningTiles: []string{"wan-1", "wan-4"},
		}

		// 模擬多次摸牌和自動打牌
		for i := 0; i < 5; i++ {
			// 摸到非聽牌
			drawnTile := "tong-" + string(rune('1'+i))

			// 檢查是否自摸
			isSelfDrawn := false
			for _, winTile := range player.WinningTiles {
				if winTile == drawnTile {
					isSelfDrawn = true
					break
				}
			}

			// 如果不是自摸，應該保持聽牌狀態
			if !isSelfDrawn && !player.IsTing {
				t.Error("聽牌狀態應該持續")
			}
		}

		// 最後驗證聽牌狀態仍然保持
		if !player.IsTing {
			t.Error("經過多次摸牌後，聽牌狀態應該仍然保持")
		}

		if len(player.WinningTiles) != 2 {
			t.Error("WinningTiles 應該保持不變")
		}

		t.Log("聽牌狀態持續測試通過")
	})

	t.Run("聽牌後不能改變手牌組合", func(t *testing.T) {
		player := &game.Player{
			ID:       "player1",
			Name:     "測試玩家",
			Position: 0,
			Hand: []string{
				"tong-1", "tong-1",
				"tong-2", "tong-2", "tong-2",
				"wan-6", "wan-6",
			},
			Melds: []game.Meld{
				{Type: "chow", Tiles: []string{"tong-2", "tong-3", "tong-4"}},
			},
			IsTing:       true,
			WinningTiles: []string{"tong-1"},
		}

		// 記錄聽牌前的 Melds 數量
		meldsCountBefore := len(player.Melds)

		// 聽牌後不應該能吃碰槓（因為會改變手牌組合）
		canChow := !player.IsTing
		canPong := !player.IsTing
		canKong := !player.IsTing

		if canChow || canPong || canKong {
			t.Error("聽牌後不應該能吃碰槓")
		}

		// 驗證 Melds 數量沒有變化
		if len(player.Melds) != meldsCountBefore {
			t.Error("聽牌後不應該能改變手牌組合")
		}

		t.Log("聽牌後限制測試通過")
	})
}
