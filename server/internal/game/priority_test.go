package game

import (
	"mahjong/internal/model"
	"testing"
	"time"
)

// TestActionPriority 測試動作優先權排序
func TestActionPriority(t *testing.T) {
	// 建立房間
	room := NewRoom("test-room")

	// 添加 4 個玩家
	players := []struct {
		id   string
		name string
	}{
		{"player1", "玩家1"},
		{"player2", "玩家2"},
		{"player3", "玩家3"},
		{"player4", "玩家4"},
	}

	for _, p := range players {
		err := room.AddPlayer(p.id, p.name, false)
		if err != nil {
			t.Fatalf("添加玩家失敗: %v", err)
		}
	}

	// 開始遊戲
	room.StartGame()
	room.DealTiles()

	// 設置最後打出的牌
	room.LastDiscardTile = "dong"
	room.LastDiscardPlayer = 0

	t.Run("胡牌優先於碰牌", func(t *testing.T) {
		room.PendingActions = []PendingAction{}

		// 設置玩家 2 的手牌，使其可以碰東
		room.Players[1].Hand = []string{"dong", "dong", "wan-1", "wan-2"}

		// 設置玩家 3 的手牌為已經胡牌的狀態（5 組面子+1 對將 = 17 張）
		// 5 組碰 + 2 張 wan-5 作為將牌
		room.Players[2].Hand = []string{"wan-5", "wan-5"} // 一對將
		room.Players[2].Melds = []model.Meld{
			{Type: "pong", Tiles: []string{"wan-1", "wan-1", "wan-1"}},
			{Type: "pong", Tiles: []string{"wan-2", "wan-2", "wan-2"}},
			{Type: "pong", Tiles: []string{"wan-3", "wan-3", "wan-3"}},
			{Type: "pong", Tiles: []string{"wan-4", "wan-4", "wan-4"}},
			{Type: "pong", Tiles: []string{"dong", "dong", "dong"}},
		}

		// 玩家 2 想碰
		room.AddPendingAction("player2", "pong", "dong", nil)
		// 玩家 3 想胡（玩家 3 已經胡牌了）
		room.AddPendingAction("player3", "hu", "dong", nil)

		// 處理動作
		action := room.ProcessPendingActions()

		if action == nil {
			t.Fatal("ProcessPendingActions 返回 nil")
		}

		if action.ActionType != "hu" {
			t.Errorf("期望執行 hu，實際執行 %s", action.ActionType)
		}

		if action.PlayerID != "player3" {
			t.Errorf("期望玩家 3 執行動作，實際是 %s", action.PlayerID)
		}
	})

	t.Run("碰牌優先於吃牌", func(t *testing.T) {
		room.PendingActions = []PendingAction{}
		room.LastDiscardPlayer = 0
		room.LastDiscardTile = "wan-2"

		// 設置玩家 2（下家）的手牌，使其可以吃萬 2（需要萬 1 和萬 3）
		room.Players[1].Hand = []string{"wan-1", "wan-3", "tong-1", "tong-2"}

		// 設置玩家 4 的手牌，使其可以碰萬 2
		room.Players[3].Hand = []string{"wan-2", "wan-2", "tong-5", "tong-6"}

		// 玩家 2（下家）想吃
		room.AddPendingAction("player2", "chow", "wan-2", nil)
		// 玩家 4 想碰
		room.AddPendingAction("player4", "pong", "wan-2", nil)

		// 處理動作
		action := room.ProcessPendingActions()

		if action == nil {
			t.Fatal("ProcessPendingActions 返回 nil")
		}

		if action.ActionType != "pong" {
			t.Errorf("期望執行 pong，實際執行 %s", action.ActionType)
		}
	})

	t.Run("槓牌優先於碰牌", func(t *testing.T) {
		room.PendingActions = []PendingAction{}
		room.LastDiscardTile = "dong"

		// 設置玩家 2 的手牌，使其可以碰東
		room.Players[1].Hand = []string{"dong", "dong", "wan-5", "wan-6"}

		// 設置玩家 3 的手牌，使其可以槓東（手牌有 3 張，加上打出的 1 張）
		room.Players[2].Hand = []string{"dong", "dong", "dong", "tong-1", "tong-2"}

		// 玩家 2 想碰
		room.AddPendingAction("player2", "pong", "dong", nil)
		// 玩家 3 想槓
		room.AddPendingAction("player3", "kong", "dong", nil)

		// 處理動作
		action := room.ProcessPendingActions()

		if action == nil {
			t.Fatal("ProcessPendingActions 返回 nil")
		}

		if action.ActionType != "kong" {
			t.Errorf("期望執行 kong，實際執行 %s", action.ActionType)
		}
	})

	t.Run("優先權相同時先到先處理", func(t *testing.T) {
		room.PendingActions = []PendingAction{}
		room.LastDiscardTile = "dong"

		// 設置兩個玩家的手牌，使其都可以碰東
		room.Players[1].Hand = []string{"dong", "dong", "wan-1", "wan-2"}
		room.Players[2].Hand = []string{"dong", "dong", "tong-1", "tong-2"}

		// 兩個玩家都想碰，玩家 2 先提交
		room.AddPendingAction("player2", "pong", "dong", nil)
		time.Sleep(10 * time.Millisecond) // 確保時間戳不同
		room.AddPendingAction("player3", "pong", "dong", nil)

		// 處理動作
		action := room.ProcessPendingActions()

		if action == nil {
			t.Fatal("ProcessPendingActions 返回 nil")
		}

		if action.PlayerID != "player2" {
			t.Errorf("期望玩家 2 先執行（最早提交），實際是 %s", action.PlayerID)
		}
	})
}

// TestPriorityOrder 測試完整的優先權順序
func TestPriorityOrder(t *testing.T) {
	room := NewRoom("test-room")

	// 添加 4 個玩家
	for i := 1; i <= 4; i++ {
		err := room.AddPlayer("player"+string(rune('0'+i)), "玩家"+string(rune('0'+i)), false)
		if err != nil {
			t.Fatalf("添加玩家失敗: %v", err)
		}
	}

	room.StartGame()
	room.DealTiles()
	room.LastDiscardTile = "wan-5"
	room.LastDiscardPlayer = 0

	// 設置玩家 1（下家）的手牌，使其可以吃萬 5
	room.Players[0].Hand = []string{"wan-4", "wan-6", "tong-1"}

	// 設置玩家 2 的手牌，使其可以碰萬 5
	room.Players[1].Hand = []string{"wan-5", "wan-5", "tong-2"}

	// 設置玩家 3 的手牌，使其可以槓萬 5
	room.Players[2].Hand = []string{"wan-5", "wan-5", "wan-5", "tong-3"}

	// 設置玩家 4 的手牌，使其可以胡（已經胡牌的狀態 = 5 組面子 + 1 對眼 = 17 張）
	room.Players[3].Hand = []string{"wan-5", "wan-5"} // 需要 2 張作為將牌
	room.Players[3].Melds = []model.Meld{
		{Type: "pong", Tiles: []string{"dong", "dong", "dong"}},
		{Type: "pong", Tiles: []string{"nan", "nan", "nan"}},
		{Type: "pong", Tiles: []string{"xi", "xi", "xi"}},
		{Type: "pong", Tiles: []string{"bei", "bei", "bei"}},
		{Type: "pong", Tiles: []string{"zhong", "zhong", "zhong"}},
	}

	// 同時添加所有四種動作
	room.PendingActions = []PendingAction{}
	room.AddPendingAction("player1", "chow", "wan-5", nil)
	room.AddPendingAction("player2", "pong", "wan-5", nil)
	room.AddPendingAction("player3", "kong", "wan-5", nil)
	room.AddPendingAction("player4", "hu", "wan-5", nil)

	// 處理動作
	action := room.ProcessPendingActions()

	if action == nil {
		t.Fatal("ProcessPendingActions 返回 nil")
	}

	// 應該執行胡牌（優先權最高）
	if action.ActionType != "hu" {
		t.Errorf("期望執行 hu（優先權最高），實際執行 %s", action.ActionType)
	}

	if action.Priority != PriorityHu {
		t.Errorf("期望優先權為 %d，實際為 %d", PriorityHu, action.Priority)
	}
}

// TestClearPendingActions 測試清空待處理動作
func TestClearPendingActions(t *testing.T) {
	room := NewRoom("test-room")

	// 添加玩家
	room.AddPlayer("player1", "玩家1", false)
	room.AddPlayer("player2", "玩家2", false)
	room.AddPlayer("player3", "玩家3", false)
	room.AddPlayer("player4", "玩家4", false)

	room.StartGame()
	room.DealTiles()
	room.LastDiscardTile = "dong"
	room.LastDiscardPlayer = 0

	// 設置玩家手牌
	room.Players[0].Hand = []string{"dong", "dong"}
	room.Players[1].Hand = []string{"dong", "nan", "xi"}

	// 添加一些待處理動作
	room.AddPendingAction("player1", "pong", "dong", nil)
	room.AddPendingAction("player2", "chow", "dong", nil) // 吃牌會被忽略（字牌不能吃）

	// 只有碰牌是有效的
	if room.GetPendingActionCount() == 0 {
		t.Skip("沒有有效的待處理動作（字牌不能吃）")
	}

	// 清空
	room.ClearPendingActions()

	if len(room.PendingActions) != 0 {
		t.Errorf("期望清空後有 0 個待處理動作，實際有 %d 個", len(room.PendingActions))
	}

	if room.IsWaitingForActions {
		t.Error("期望 IsWaitingForActions 為 false")
	}
}

// TestNoActionsToProcess 測試沒有待處理動作的情況
func TestNoActionsToProcess(t *testing.T) {
	room := NewRoom("test-room")
	room.AddPlayer("player1", "玩家1", false)

	room.StartGame()
	room.DealTiles()

	// 沒有添加任何待處理動作
	action := room.ProcessPendingActions()

	if action != nil {
		t.Errorf("期望返回 nil（沒有待處理動作），實際返回 %+v", action)
	}
}

// TestMultipleSamePriorityActions 測試多個相同優先權的動作
func TestMultipleSamePriorityActions(t *testing.T) {
	room := NewRoom("test-room")

	// 添加玩家
	for i := 1; i <= 4; i++ {
		room.AddPlayer("player"+string(rune('0'+i)), "玩家"+string(rune('0'+i)), false)
	}

	room.StartGame()
	room.DealTiles()
	room.LastDiscardTile = "dong"
	room.LastDiscardPlayer = 0

	// 設置三個玩家的手牌，使其都可以碰東
	room.Players[0].Hand = []string{"dong", "dong"}
	room.Players[1].Hand = []string{"dong", "dong"}
	room.Players[2].Hand = []string{"dong", "dong"}

	// 三個玩家都想碰
	room.PendingActions = []PendingAction{}
	room.AddPendingAction("player1", "pong", "dong", nil)
	time.Sleep(5 * time.Millisecond)
	room.AddPendingAction("player2", "pong", "dong", nil)
	time.Sleep(5 * time.Millisecond)
	room.AddPendingAction("player3", "pong", "dong", nil)

	// 處理動作
	action := room.ProcessPendingActions()

	if action == nil {
		t.Fatal("ProcessPendingActions 返回 nil")
	}

	// 應該是第一個提交的玩家
	if action.PlayerID != "player1" {
		t.Errorf("期望玩家 1 執行（最早提交），實際是 %s", action.PlayerID)
	}
}