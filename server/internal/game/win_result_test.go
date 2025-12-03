package game

import (
	"testing"
)

// TestHandleHu_BasicWin 測試基本胡牌邏輯
func TestHandleHu_BasicWin(t *testing.T) {
	room := NewRoom("test-room")

	// 創建測試玩家（模擬真實遊戲：5張手牌含胡牌 + 4組吃碰槓 = 17張）
	player := &Player{
		ID:       "player1",
		Name:     "測試玩家",
		Position: 0,
		Hand: []string{
			"tong-7", "tong-8", "tong-9", // 自摸的 tong-7 組成順子
			"wan-2", "wan-2",              // 將牌
		},
		Melds: []Meld{
			{Type: "chow", Tiles: []string{"wan-1", "wan-2", "wan-3"}},
			{Type: "chow", Tiles: []string{"tiao-4", "tiao-5", "tiao-6"}},
			{Type: "pong", Tiles: []string{"wan-7", "wan-7", "wan-7"}},
			{Type: "chow", Tiles: []string{"tong-2", "tong-3", "tong-4"}},
		},
		Score:  0,
		Flowers: []string{},
	}

	room.Players = []*Player{player}
	room.Game = NewMahjongGame(room.Players)

	// 測試自摸胡牌（摸到 tong-7 組成順子）
	winTile := "tong-7"
	isSelfDrawn := true

	t.Logf("手牌總數: %d, 手牌: %v", len(player.Hand), player.Hand)
	t.Logf("吃碰槓總數: %d組", len(player.Melds))
	t.Logf("測試胡牌: winTile=%s, isSelfDrawn=%v", winTile, isSelfDrawn)

	result := room.HandleHu(player.ID, winTile, isSelfDrawn)

	// 驗證結果
	if result == nil {
		// 先檢查 CanHu 是否能判斷這手牌可以胡
		canWin := room.Game.CanHu(player.Hand, player.Melds)
		t.Logf("CanHu 結果: %v", canWin)
		t.Fatal("HandleHu 應該返回 WinResult，但返回 nil")
	}

	if result.TotalTai <= 0 {
		t.Errorf("TotalTai 應該大於 0，實際值: %d", result.TotalTai)
	}

	if result.BaseScore <= 0 {
		t.Errorf("BaseScore 應該大於 0，實際值: %d", result.BaseScore)
	}

	t.Logf("✅ 胡牌成功: TotalTai=%d, BaseScore=%d", result.TotalTai, result.BaseScore)
}

// TestHandleHu_WithFlowers 測試帶花牌的胡牌
func TestHandleHu_WithFlowers(t *testing.T) {
	room := NewRoom("test-room")

	// 創建測試玩家（帶花牌：5張手牌含胡牌 + 4組吃碰槓 = 17張）
	player := &Player{
		ID:       "player1",
		Name:     "測試玩家",
		Position: 0,
		Hand: []string{
			"tong-7", "tong-8", "tong-9", // 自摸的 tong-7 組成順子
			"wan-2", "wan-2",              // 將牌
		},
		Melds: []Meld{
			{Type: "chow", Tiles: []string{"wan-1", "wan-2", "wan-3"}},
			{Type: "chow", Tiles: []string{"tiao-4", "tiao-5", "tiao-6"}},
			{Type: "pong", Tiles: []string{"wan-7", "wan-7", "wan-7"}},
			{Type: "chow", Tiles: []string{"tong-2", "tong-3", "tong-4"}},
		},
		Score:  0,
		Flowers: []string{"flower-mei", "flower-lan", "flower-ju"}, // 3 張花牌
	}

	room.Players = []*Player{player}
	room.Game = NewMahjongGame(room.Players)

	winTile := "wan-5"
	isSelfDrawn := true

	result := room.HandleHu(player.ID, winTile, isSelfDrawn)

	if result == nil {
		t.Fatal("HandleHu 應該返回 WinResult，但返回 nil")
	}

	// 驗證花牌台數
	hasFlowerTai := false
	for _, ht := range result.HandTypes {
		if ht.Name == "花牌" && ht.Tai == 3 {
			hasFlowerTai = true
			break
		}
	}

	if !hasFlowerTai {
		t.Error("結果中應該包含 3 台花牌")
	}

	t.Logf("✅ 帶花牌胡牌成功: TotalTai=%d, BaseScore=%d, HandTypes=%v",
		result.TotalTai, result.BaseScore, result.HandTypes)
}

// TestHandleHu_InvalidHand 測試無效的胡牌
func TestHandleHu_InvalidHand(t *testing.T) {
	room := NewRoom("test-room")

	// 創建測試玩家（無法胡牌的手牌）
	player := &Player{
		ID:       "player1",
		Name:     "測試玩家",
		Position: 0,
		Hand: []string{
			"wan-1", "wan-2", "wan-4", // 不是順子
			"tong-4", "tong-5", "tong-7", // 不是順子
			"tiao-1", "tiao-2", "tiao-3",
			"dong", "dong", "xi",
			"wan-5",
		},
		Melds:  []Meld{},
		Score:  0,
		Flowers: []string{},
	}

	room.Players = []*Player{player}
	room.Game = NewMahjongGame(room.Players)

	winTile := "wan-5"
	isSelfDrawn := true

	result := room.HandleHu(player.ID, winTile, isSelfDrawn)

	// 無效手牌應該返回 nil
	if result != nil {
		t.Error("無效手牌應該返回 nil，但返回了 WinResult")
	}

	t.Log("✅ 無效手牌正確返回 nil")
}

// TestHandleHu_OtherPlayerDiscard 測試吃胡（非自摸）
func TestHandleHu_OtherPlayerDiscard(t *testing.T) {
	room := NewRoom("test-room")

	// 創建測試玩家（4張手牌 + 4組吃碰槓 + 別人打出的 wan-5 = 17張）
	player := &Player{
		ID:       "player1",
		Name:     "測試玩家",
		Position: 0,
		Hand: []string{
			"tong-7", "tong-8", "tong-9", // 順子
			"wan-5",                       // 單張將牌（等待別人打出的 wan-5）
		},
		Melds: []Meld{
			{Type: "chow", Tiles: []string{"wan-1", "wan-2", "wan-3"}},
			{Type: "chow", Tiles: []string{"tiao-4", "tiao-5", "tiao-6"}},
			{Type: "pong", Tiles: []string{"wan-7", "wan-7", "wan-7"}},
			{Type: "chow", Tiles: []string{"tong-2", "tong-3", "tong-4"}},
		},
		Score:  0,
		Flowers: []string{},
	}

	room.Players = []*Player{player}
	room.Game = NewMahjongGame(room.Players)

	// 別人打出 wan-5（吃胡）
	winTile := "wan-5"
	isSelfDrawn := false

	t.Logf("手牌總數: %d, 手牌: %v", len(player.Hand), player.Hand)
	t.Logf("吃碰槓總數: %d組", len(player.Melds))
	t.Logf("測試吃胡: winTile=%s, isSelfDrawn=%v", winTile, isSelfDrawn)

	result := room.HandleHu(player.ID, winTile, isSelfDrawn)

	if result == nil {
		t.Fatal("HandleHu 應該返回 WinResult，但返回 nil")
	}

	// 驗證是放炮（非自摸，台數應該不包含自摸台）
	if result.TotalTai <= 0 {
		t.Errorf("TotalTai 應該大於 0，實際值: %d", result.TotalTai)
	}

	t.Logf("✅ 吃胡成功: TotalTai=%d, BaseScore=%d", result.TotalTai, result.BaseScore)
}

// TestCalculateScore_FlowerBonus 測試花牌加台
func TestCalculateScore_FlowerBonus(t *testing.T) {
	player := &Player{
		ID:       "player1",
		Name:     "測試玩家",
		Position: 0,
		Hand: []string{
			"wan-1", "wan-2", "wan-3",
			"tong-4", "tong-5", "tong-6",
			"tiao-7", "tiao-8", "tiao-9",
			"dong", "dong", "dong",
			"wan-5", "wan-5",
		},
		Melds:   []Meld{},
		Flowers: []string{"flower-mei", "flower-lan", "flower-ju"},
	}

	game := NewMahjongGame([]*Player{player})

	lastTile := "wan-5"
	isSelfDrawn := true

	result := game.CalculateScore(player, lastTile, isSelfDrawn)

	if result == nil {
		t.Fatal("CalculateScore 應該返回結果")
	}

	// 驗證花牌台數
	expectedFlowerTai := len(player.Flowers)
	foundFlowerTai := false

	for _, ht := range result.HandTypes {
		if ht.Name == "花牌" {
			if ht.Tai != expectedFlowerTai {
				t.Errorf("花牌台數錯誤: 期望 %d，實際 %d", expectedFlowerTai, ht.Tai)
			}
			foundFlowerTai = true
			break
		}
	}

	if !foundFlowerTai {
		t.Error("結果中應該包含花牌台數")
	}

	t.Logf("✅ 花牌計分正確: %d 張花牌 = %d 台", len(player.Flowers), expectedFlowerTai)
}
