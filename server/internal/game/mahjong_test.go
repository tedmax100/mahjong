package game

import (
	"mahjong/internal/model"
	"mahjong/internal/tile"
	"testing"
)

// TestCanHu 測試胡牌判斷
func TestCanHu(t *testing.T) {
	players := []*Player{
		{ID: "1", Name: "Player1", Position: 0, Hand: make([]string, 0), Melds: make([]model.Meld, 0), Flowers: make([]string, 0)},
	}
	game := NewMahjongGame(players)

	// 測試 1: 簡單的胡牌（5 組刻子+1 對眼 = 17 張）
	hand1 := []string{
		"wan-1", "wan-1", "wan-1", // 刻子
		"wan-2", "wan-2", "wan-2", // 刻子
		"wan-3", "wan-3", "wan-3", // 刻子
		"wan-4", "wan-4", "wan-4", // 刻子
		"wan-5", "wan-5", "wan-5", // 刻子
		"wan-6", "wan-6", // 對眼
	}
	if !game.CanHu(hand1, []model.Meld{}) {
		t.Errorf("應該能胡牌，但判斷為不能胡")
	}

	// 測試 2: 簡單的順子胡牌（5 組順子+1 對眼 = 17 張）
	hand2 := []string{
		"wan-1", "wan-2", "wan-3", // 順子
		"wan-4", "wan-5", "wan-6", // 順子
		"wan-7", "wan-8", "wan-9", // 順子
		"tong-1", "tong-2", "tong-3", // 順子
		"tiao-7", "tiao-8", "tiao-9", // 順子
		"dong", "dong", // 對眼
	}
	if !game.CanHu(hand2, []model.Meld{}) {
		t.Errorf("應該能胡牌（順子），但判斷為不能胡")
	}

	// 測試 3: 不能胡牌
	hand3 := []string{
		"wan-1", "wan-2", "wan-4", // 不連續
		"wan-5", "wan-6", "wan-7",
		"tong-1", "tong-2", "tong-3",
		"tiao-7", "tiao-8", "tiao-9",
		"dong", "dong",
	}
	if game.CanHu(hand3, []model.Meld{}) {
		t.Errorf("不應該能胡牌，但判斷為能胡")
	}

	// 測試 4: 有碰的情況（2 組已展示+手牌 3 組+1 對眼 = 17 張）
	hand4 := []string{
		"wan-1", "wan-2", "wan-3", // 順子
		"wan-4", "wan-5", "wan-6", // 順子
		"wan-7", "wan-8", "wan-9", // 順子
		"dong", "dong", // 對眼
	}
	melds4 := []model.Meld{
		{Type: "pong", Tiles: []string{"tong-5", "tong-5", "tong-5"}},
		{Type: "pong", Tiles: []string{"tiao-9", "tiao-9", "tiao-9"}},
	}
	if !game.CanHu(hand4, melds4) {
		t.Errorf("應該能胡牌（有碰），但判斷為不能胡")
	}
}

// TestCanPong 測試碰牌判斷
func TestCanPong(t *testing.T) {
	game := &MahjongGame{}

	// 測試 1: 手牌中有 2 張相同
	hand1 := []string{"wan-1", "wan-1", "wan-2", "wan-3"}
	if !game.CanPong(hand1, "wan-1") {
		t.Errorf("應該能碰 wan-1")
	}

	// 測試 2: 手牌中只有 1 張
	hand2 := []string{"wan-1", "wan-2", "wan-3"}
	if game.CanPong(hand2, "wan-1") {
		t.Errorf("不應該能碰 wan-1（只有 1 張）")
	}

	// 測試 3: 手牌中有 3 張
	hand3 := []string{"wan-1", "wan-1", "wan-1", "wan-2"}
	if !game.CanPong(hand3, "wan-1") {
		t.Errorf("應該能碰 wan-1（有 3 張）")
	}
}

// TestCanKong 測試槓牌判斷
func TestCanKong(t *testing.T) {
	game := &MahjongGame{}

	// 測試 1: 手牌中有 3 張相同
	hand1 := []string{"wan-1", "wan-1", "wan-1", "wan-2"}
	if !game.CanKong(hand1, "wan-1") {
		t.Errorf("應該能槓 wan-1")
	}

	// 測試 2: 手牌中只有 2 張
	hand2 := []string{"wan-1", "wan-1", "wan-2", "wan-3"}
	if game.CanKong(hand2, "wan-1") {
		t.Errorf("不應該能槓 wan-1（只有 2 張）")
	}

	// 測試 3: 手牌中有 4 張
	hand3 := []string{"wan-1", "wan-1", "wan-1", "wan-1"}
	if !game.CanKong(hand3, "wan-1") {
		t.Errorf("應該能槓 wan-1（有 4 張）")
	}
}

// TestCanChow 測試吃牌判斷
func TestCanChow(t *testing.T) {
	game := &MahjongGame{}

	// 測試 1: 能吃（wan-2，手牌有 wan-1 和 wan-3）
	hand1 := []string{"wan-1", "wan-3", "tong-5", "tiao-7"}
	combos1 := game.CanChow(hand1, "wan-2")
	if len(combos1) != 1 {
		t.Errorf("應該有 1 種吃牌組合，實際: %d", len(combos1))
	}
	expectedCombo1 := []string{"wan-1", "wan-2", "wan-3"}
	if len(combos1) > 0 && !isSameCombination(combos1[0], expectedCombo1) {
		t.Errorf("吃牌組合錯誤，期望: %v，實際: %v", expectedCombo1, combos1[0])
	}

	// 測試 2: 能吃（wan-5，手牌有 wan-4,wan-6 和 wan-6,wan-7）- 多種組合
	hand2 := []string{"wan-4", "wan-6", "wan-7", "tong-1"}
	combos2 := game.CanChow(hand2, "wan-5")
	if len(combos2) != 2 {
		t.Errorf("應該有 2 種吃牌組合，實際: %d", len(combos2))
	}

	// 測試 3: 不能吃（字牌）
	hand3 := []string{"wan-1", "wan-2", "tong-5"}
	combos3 := game.CanChow(hand3, "dong")
	if len(combos3) != 0 {
		t.Errorf("字牌不應該能吃，實際組合數: %d", len(combos3))
	}

	// 測試 4: 不能吃（手牌不連續）
	hand4 := []string{"wan-1", "wan-4", "tong-5"}
	combos4 := game.CanChow(hand4, "wan-2")
	if len(combos4) != 0 {
		t.Errorf("手牌不連續不應該能吃，實際組合數: %d", len(combos4))
	}

	// 測試 5: 能吃（wan-1，手牌有 wan-2 和 wan-3）
	hand5 := []string{"wan-2", "wan-3", "tong-5"}
	combos5 := game.CanChow(hand5, "wan-1")
	if len(combos5) != 1 {
		t.Errorf("應該有 1 種吃牌組合，實際: %d", len(combos5))
	}

	// 測試 6: 能吃（wan-9，手牌有 wan-7 和 wan-8）
	hand6 := []string{"wan-7", "wan-8", "tong-5"}
	combos6 := game.CanChow(hand6, "wan-9")
	if len(combos6) != 1 {
		t.Errorf("應該有 1 種吃牌組合，實際: %d", len(combos6))
	}

	// 測試 7: 不能吃（不同花色）
	hand7 := []string{"wan-1", "wan-3", "tong-5"}
	combos7 := game.CanChow(hand7, "tong-2")
	if len(combos7) != 0 {
		t.Errorf("不同花色不應該能吃，實際組合數: %d", len(combos7))
	}
}

// TestIsFlowerTile 測試花牌判斷
func TestIsFlowerTile(t *testing.T) {
	if !tile.IsFlower("flower-chun") {
		t.Errorf("flower-chun 應該是花牌")
	}

	if !tile.IsFlower("flower-mei") {
		t.Errorf("flower-mei 應該是花牌")
	}

	if tile.IsFlower("wan-1") {
		t.Errorf("wan-1 不應該是花牌")
	}

	if tile.IsFlower("dong") {
		t.Errorf("dong 不應該是花牌")
	}
}

// TestCheckDraw 測試流局判斷
func TestCheckDraw(t *testing.T) {
	players := []*Player{
		{ID: "1", Name: "Player1", Position: 0},
	}
	game := NewMahjongGame(players)

	// 測試 1: 牌山充足
	if game.CheckDraw() {
		t.Errorf("牌山充足時不應該流局")
	}

	// 測試 2: 模擬牌山消耗到 8 張
	for len(game.Deck) > 8 {
		game.DrawTile()
	}

	if !game.CheckDraw() {
		t.Errorf("牌山剩餘 8 張時應該流局")
	}
}

// TestCanExposedKong tests the logic for an exposed kong, including promoting a pong.
func TestCanExposedKong(t *testing.T) {
	game := NewMahjongGame([]*Player{})

	// Scenario 1: Player has 3 tiles in hand, can kong a discard (from others).
	player1 := &Player{
		Hand:  []string{"wan-5", "wan-5", "wan-5", "tiao-1"},
		Melds: []model.Meld{},
	}
	if !game.CanExposedKong(player1, "wan-5", false) {
		t.Errorf("TestCanExposedKong: Should be able to exposed kong from 3 tiles in hand when others discard.")
	}

	// Scenario 2a: Player has an exposed pong, can promote it when self-drawn.
	player2 := &Player{
		Hand: []string{"wan-5", "tiao-1", "tiao-2"},
		Melds: []model.Meld{
			{Type: "pong", Tiles: []string{"wan-5", "wan-5", "wan-5"}},
		},
	}
	if !game.CanExposedKong(player2, "wan-5", true) {
		t.Errorf("TestCanExposedKong: Should be able to promote an exposed pong when self-drawn.")
	}

	// Scenario 2b: Player has an exposed pong, CANNOT promote it with others' discard.
	if game.CanExposedKong(player2, "wan-5", false) {
		t.Errorf("TestCanExposedKong: Should NOT be able to promote an exposed pong when others discard.")
	}

	// Scenario 3: Player has only 2 tiles in hand, no pong.
	player3 := &Player{
		Hand:  []string{"wan-5", "wan-5", "tiao-1"},
		Melds: []model.Meld{},
	}
	if game.CanExposedKong(player3, "wan-5", false) {
		t.Errorf("TestCanExposedKong: Should not be able to kong with only 2 tiles in hand.")
	}

	// Scenario 4: Player has a pong of a different tile.
	player4 := &Player{
		Hand: []string{"tiao-1", "tiao-2"},
		Melds: []model.Meld{
			{Type: "pong", Tiles: []string{"tiao-3", "tiao-3", "tiao-3"}},
		},
	}
	if game.CanExposedKong(player4, "wan-5", false) {
		t.Errorf("TestCanExposedKong: Should not be able to kong tile that is not in hand or melds.")
	}
}

// TestDrawTileFromEnd tests drawing a tile from the end of the deck.
func TestDrawTileFromEnd(t *testing.T) {
	game := NewMahjongGame([]*Player{})
	originalDeckSize := len(game.Deck)
	if originalDeckSize == 0 {
		t.Fatal("TestDrawTileFromEnd: Deck is empty, cannot run test.")
	}

	lastTile := game.Deck[originalDeckSize-1]
	drawnTile := game.DrawTileFromEnd()

	if drawnTile != lastTile {
		t.Errorf("TestDrawTileFromEnd: Expected to draw tile %s, but got %s", lastTile, drawnTile)
	}

	if len(game.Deck) != originalDeckSize-1 {
		t.Errorf("TestDrawTileFromEnd: Deck size should be %d, but is %d", originalDeckSize-1, len(game.Deck))
	}
}

// TestHandleKong_CannotPromotePongWithOthersDiscard 測試不能用別人打出的牌來補槓
func TestHandleKong_CannotPromotePongWithOthersDiscard(t *testing.T) {
	// 1. Setup Room and Players
	playerA := &Player{
		ID:       "playerA",
		Name:     "Player A",
		Position: 0,
		Hand:     []string{"tiao-1", "tiao-2", "tiao-3"},
		Melds: []model.Meld{
			{Type: "pong", Tiles: []string{"wan-8", "wan-8", "wan-8"}},
		},
	}
	playerB := &Player{ID: "playerB", Name: "Player B", Position: 1}
	room := &Room{
		ID:      "test-room",
		Players: []*Player{playerA, playerB},
		Game:    NewMahjongGame([]*Player{playerA, playerB}),
	}
	room.GameStarted = true

	// 2. Simulate State: Player B discards wan-8
	discardedTile := "wan-8"
	room.LastDiscardPlayer = playerB.Position
	room.Game.DiscardPile = append(room.Game.DiscardPile, discardedTile)

	// 3. Call HandleKong - should fail because cannot promote pong with others' discard
	success, _ := room.HandleKong(playerA.ID, discardedTile, false)

	// 4. Assertions - should return false
	if success {
		t.Errorf("HandleKong should have returned false - cannot promote pong with others' discard")
	}

	// Assert meld was NOT updated
	meld := playerA.Melds[0]
	if meld.Type != "pong" {
		t.Errorf("Meld type should still be 'pong', but got '%s'", meld.Type)
	}
	if len(meld.Tiles) != 3 {
		t.Errorf("Meld should still have 3 tiles, but has %d", len(meld.Tiles))
	}
}

// TestHandleKong_PromotePongWithSelfDraw 測試可以用自己摸到的牌來補槓
func TestHandleKong_PromotePongWithSelfDraw(t *testing.T) {
	// 1. Setup Room and Players
	playerA := &Player{
		ID:       "playerA",
		Name:     "Player A",
		Position: 0,
		Hand:     []string{"tiao-1", "tiao-2", "tiao-3", "wan-8"}, // 手上有一張 wan-8
		Melds: []model.Meld{
			{Type: "pong", Tiles: []string{"wan-8", "wan-8", "wan-8"}},
		},
	}
	playerB := &Player{ID: "playerB", Name: "Player B", Position: 1}
	room := &Room{
		ID:      "test-room",
		Players: []*Player{playerA, playerB},
		Game:    NewMahjongGame([]*Player{playerA, playerB}),
	}
	room.GameStarted = true
	initialHandSize := len(playerA.Hand)
	deckSize := len(room.Game.Deck)

	// 2. Simulate State: No discard pile (self-drawn)
	tileToKong := "wan-8"
	// DiscardPile is empty or doesn't contain wan-8

	// 3. Call HandleKong
	success, drawnTile := room.HandleKong(playerA.ID, tileToKong, false)

	// 4. Assertions
	if !success {
		t.Fatalf("HandleKong should have returned true for a valid promoted kong from self-draw")
	}

	// 檢查槓牌後有補牌
	if drawnTile == "" {
		t.Error("Expected a supplemental tile to be drawn after kong")
	} else {
		t.Logf("Supplemental tile drawn after kong: %s", drawnTile)
	}

	// Assert meld was updated
	updatedMeld := playerA.Melds[0]
	if updatedMeld.Type != "kong_promoted" {
		t.Errorf("Meld type should be 'kong_promoted', but got '%s'", updatedMeld.Type)
	}
	if len(updatedMeld.Tiles) != 4 {
		t.Errorf("Meld should have 4 tiles after promotion, but has %d", len(updatedMeld.Tiles))
	}

	// Assert player drew a replacement tile (hand should be same size, since we removed 1 and drew 1)
	if len(playerA.Hand) != initialHandSize {
		t.Errorf("Player's hand size should be %d after kong, but got %d", initialHandSize, len(playerA.Hand))
	}

	// Assert deck size decreased by 1 (for the replacement tile)
	if len(room.Game.Deck) != deckSize-1 {
		t.Errorf("Deck size should have decreased by 1, from %d to %d, but is %d", deckSize, deckSize-1, len(room.Game.Deck))
	}

	// Assert turn changed to the konging player
	if room.CurrentTurn != playerA.Position {
		t.Errorf("It should be player A's turn after kong, but got turn %d", room.CurrentTurn)
	}
}

// TestCheckTing_UnsortedHand tests the CheckTing function with an unsorted hand
func TestCheckTing_UnsortedHand(t *testing.T) {
	players := []*Player{
		{ID: "1", Name: "Player1"},
	}
	game := NewMahjongGame(players)

	// This hand is unsorted and is waiting for "tong-5" to form a pair,
	// while "wan-1", "wan-2", "wan-3" will form a chow.
	hand := []string{"wan-3", "wan-1", "tong-5", "wan-2"}
	melds := []model.Meld{
		{Type: "pong", Tiles: []string{"tiao-1", "tiao-1", "tiao-1"}},
		{Type: "pong", Tiles: []string{"tiao-2", "tiao-2", "tiao-2"}},
		{Type: "pong", Tiles: []string{"tiao-3", "tiao-3", "tiao-3"}},
		{Type: "pong", Tiles: []string{"tiao-4", "tiao-4", "tiao-4"}},
	}

	result := game.CheckTing(hand, melds)

	if !result.IsTing {
		t.Errorf("Hand should be in Ting state, but CheckTing returned false.")
	}

	found := false
	for _, tile := range result.WinningTiles {
		if tile == "tong-5" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected winning tile 'tong-5', but it was not found in %v", result.WinningTiles)
	}
}

// TestConsecutiveFlowerReplacement 測試連續補花
func TestConsecutiveFlowerReplacement(t *testing.T) {
	player := &Player{
		ID:      "p1",
		Flowers: make([]string, 0),
		Hand:    make([]string, 0),
	}
	game := NewMahjongGame([]*Player{player})

	// 手動設置牌山：花牌 -> 花牌 -> 萬子 -> 填充牌 (避免流局檢查)
	// 台灣麻將剩餘 8 張流局，所以總數至少要 3 + 8 + 1 = 12 張
	game.Deck = []string{"flower-chun", "flower-xia", "wan-1"}
	for i := 0; i < 20; i++ {
		game.Deck = append(game.Deck, "tong-1")
	}

	// 執行補花摸牌
	drawnTile := game.DrawTileWithFlowerReplacement(player)

	// 驗證
	if drawnTile != "wan-1" {
		t.Errorf("應該摸到 wan-1，但摸到 %s", drawnTile)
	}

	if len(player.Flowers) != 2 {
		t.Errorf("應該有 2 張花牌，實際 %d 張", len(player.Flowers))
	}

	if len(player.Flowers) >= 2 {
		if player.Flowers[0] != "flower-chun" || player.Flowers[1] != "flower-xia" {
			t.Errorf("花牌順序或內容錯誤: %v", player.Flowers)
		}
	}
}

// TestNineGatesTing 測試九蓮寶燈聽牌（9 面聽）
// 1112345678999 聽 1-9
func TestNineGatesTing(t *testing.T) {
	// 注意：台灣麻將 16 張，九蓮寶燈通常指清一色且聽牌狀態下聽很多張
	// 標準 13 張九蓮是 1112345678999
	// 16 張要湊成胡牌型（5 組 + 1 對），九蓮寶燈可能定義不同，
	// 這裡我們測試一個清一色且聽很多張的型態，驗證 CheckTing 的能力
	
	// 構造一個清一色聽多張的牌型 (16張)
	// 例如：111 234 567 888 234 5 (聽 2, 5, 8?) -> 3刻+2順+單吊?
	// 讓我們用更簡單的邏輯：測試 CheckTing 是否能找出所有聽牌
	
	// 構造：111 234 456 678 999 + 聽一張
	// 111 (pong)
	// 234 (chow)
	// 456 (chow)
	// 678 (chow)
	// 999 (pong)
	// 剩餘一張：假設是 5，則聽 2, 5, 8 (如果 456 是 45 聽 3,6? 不對)
	
	// 讓我們測試一個標準的多面聽：
	// 手牌：wan-2, wan-3, wan-4, wan-5, wan-6
	// 聽：wan-1, wan-4, wan-7
	// 因為 234 + 56 (聽4,7) 或 23 + 456 (聽1,4) -> 合集 1,4,7
	
	player := &Player{ID: "p1"}
	game := NewMahjongGame([]*Player{player})
	
	hand := []string{
		"tong-1", "tong-1", "tong-1", // 刻
		"tiao-1", "tiao-1", "tiao-1", // 刻
		"wan-9", "wan-9",             // 眼
		"wan-2", "wan-3", "wan-4", "wan-5", "wan-6", // 5 連張
		// 總共 3+3+2+5 = 13 張，還差 3 張？ 
		// 16 張麻將聽牌時手牌數應為 16 張 (4*3+1+3) ? 
		// 不，聽牌時手牌數是 16 張，加一張成 17 張胡。
		// 所以我們需要 16 張手牌。
		"xi", "xi", "xi", // 刻
	}
	// 手牌：
	// 刻：tong-1
	// 刻：tiao-1
	// 刻：xi
	// 眼：wan-9
	// 剩下：wan-2, wan-3, wan-4, wan-5, wan-6 (5張)
	// 總共 3+3+3+2+5 = 16 張
	// 聽牌分析：
	// wan-234 (順) + wan-56 (搭) -> 聽 4, 7
	// wan-23 (搭) + wan-456 (順) -> 聽 1, 4
	// 所以應該聽 1, 4, 7
	
	result := game.CheckTing(hand, []model.Meld{})
	
	if !result.IsTing {
		t.Fatalf("應該聽牌")
	}
	
	expectedWins := []string{"wan-1", "wan-4", "wan-7"}
	
	// 驗證是否包含所有期望的聽牌
	for _, exp := range expectedWins {
		found := false
		for _, act := range result.WinningTiles {
			if act == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("應該聽 %s，但結果未包含。實際: %v", exp, result.WinningTiles)
		}
	}
	
	// 驗證沒有多餘的（在這個例子中應該只有 1,4,7）
	// 注意：wan-1 (123, 456), wan-4 (234, 456 or 44 + 2356X or 234 456), wan-7 (234, 567)
	if len(result.WinningTiles) != 3 {
		t.Errorf("應該只聽 3 張牌，實際: %v", result.WinningTiles)
	}
}

// TestHandleKong_ConcealedKong 測試暗槓（手牌有4張相同的牌）
func TestHandleKong_ConcealedKong(t *testing.T) {
	// Setup: 玩家手上有 4 張相同的牌
	playerA := &Player{
		ID:       "playerA",
		Name:     "Player A",
		Position: 0,
		Hand:     []string{"wan-5", "wan-5", "wan-5", "wan-5", "tiao-1", "tiao-2", "tiao-3"},
		Melds:    []model.Meld{},
	}
	playerB := &Player{ID: "playerB", Name: "Player B", Position: 1}
	room := &Room{
		ID:      "test-room",
		Players: []*Player{playerA, playerB},
		Game:    NewMahjongGame([]*Player{playerA, playerB}),
	}
	room.GameStarted = true
	room.CurrentTurn = playerA.Position
	initialHandSize := len(playerA.Hand)
	deckSize := len(room.Game.Deck)

	// 呼叫 HandleKong（暗槓 = isConcealed: true）
	success, drawnTile := room.HandleKong(playerA.ID, "wan-5", true)

	// 驗證
	if !success {
		t.Fatalf("暗槓應該成功")
	}

	// 檢查槓牌後有補牌
	if drawnTile == "" {
		t.Error("暗槓後應該補牌")
	}

	// 驗證 meld 類型是 kong_concealed
	if len(playerA.Melds) != 1 {
		t.Fatalf("應該有 1 組 meld，實際: %d", len(playerA.Melds))
	}

	meld := playerA.Melds[0]
	if meld.Type != "kong_concealed" {
		t.Errorf("Meld 類型應該是 'kong_concealed'，但是 '%s'", meld.Type)
	}
	if len(meld.Tiles) != 4 {
		t.Errorf("暗槓 Meld 應該有 4 張牌，實際: %d", len(meld.Tiles))
	}
	for _, tile := range meld.Tiles {
		if tile != "wan-5" {
			t.Errorf("暗槓的牌應該都是 'wan-5'，但有 '%s'", tile)
		}
	}

	// 手牌應該減少 4 張，再補 1 張 = 減少 3 張
	if len(playerA.Hand) != initialHandSize-3 {
		t.Errorf("手牌數應該是 %d，實際: %d", initialHandSize-3, len(playerA.Hand))
	}

	// 牌山應該減少 1 張（補牌用）
	if len(room.Game.Deck) != deckSize-1 {
		t.Errorf("牌山應該減少 1 張，從 %d 到 %d，實際: %d", deckSize, deckSize-1, len(room.Game.Deck))
	}

	// 輪到槓牌的玩家
	if room.CurrentTurn != playerA.Position {
		t.Errorf("暗槓後應該輪到玩家 A，但輪到 %d", room.CurrentTurn)
	}
}

// TestHandleKong_ExposedKong 測試明槓（手牌有3張，別人打出1張）
func TestHandleKong_ExposedKong(t *testing.T) {
	// Setup: 玩家手上有 3 張相同的牌
	playerA := &Player{
		ID:       "playerA",
		Name:     "Player A",
		Position: 0,
		Hand:     []string{"wan-5", "wan-5", "wan-5", "tiao-1", "tiao-2", "tiao-3"},
		Melds:    []model.Meld{},
	}
	playerB := &Player{ID: "playerB", Name: "Player B", Position: 1}
	room := &Room{
		ID:      "test-room",
		Players: []*Player{playerA, playerB},
		Game:    NewMahjongGame([]*Player{playerA, playerB}),
	}
	room.GameStarted = true
	room.CurrentTurn = playerB.Position // 輪到 B 打牌後 A 槓
	initialHandSize := len(playerA.Hand)
	deckSize := len(room.Game.Deck)

	// 模擬 B 打出 wan-5
	room.LastDiscardPlayer = playerB.Position
	room.Game.DiscardPile = append(room.Game.DiscardPile, "wan-5")

	// 呼叫 HandleKong（明槓 = isConcealed: false）
	success, drawnTile := room.HandleKong(playerA.ID, "wan-5", false)

	// 驗證
	if !success {
		t.Fatalf("明槓應該成功")
	}

	// 檢查槓牌後有補牌
	if drawnTile == "" {
		t.Error("明槓後應該補牌")
	}

	// 驗證 meld 類型是 kong_exposed
	if len(playerA.Melds) != 1 {
		t.Fatalf("應該有 1 組 meld，實際: %d", len(playerA.Melds))
	}

	meld := playerA.Melds[0]
	if meld.Type != "kong_exposed" {
		t.Errorf("Meld 類型應該是 'kong_exposed'，但是 '%s'", meld.Type)
	}
	if len(meld.Tiles) != 4 {
		t.Errorf("明槓 Meld 應該有 4 張牌，實際: %d", len(meld.Tiles))
	}

	// 手牌應該減少 3 張，再補 1 張 = 減少 2 張
	if len(playerA.Hand) != initialHandSize-2 {
		t.Errorf("手牌數應該是 %d，實際: %d", initialHandSize-2, len(playerA.Hand))
	}

	// 牌山應該減少 1 張（補牌用）
	if len(room.Game.Deck) != deckSize-1 {
		t.Errorf("牌山應該減少 1 張，從 %d 到 %d，實際: %d", deckSize, deckSize-1, len(room.Game.Deck))
	}

	// 輪到槓牌的玩家
	if room.CurrentTurn != playerA.Position {
		t.Errorf("明槓後應該輪到玩家 A，但輪到 %d", room.CurrentTurn)
	}

	// 棄牌堆應該移除被槓的牌
	for _, tile := range room.Game.DiscardPile {
		if tile == "wan-5" {
			t.Error("棄牌堆不應該還有被槓的 wan-5")
		}
	}
}

// TestHandleKong_ConcealedKong_NotEnoughTiles 測試暗槓失敗（不足4張）
func TestHandleKong_ConcealedKong_NotEnoughTiles(t *testing.T) {
	// Setup: 玩家手上只有 3 張
	playerA := &Player{
		ID:       "playerA",
		Name:     "Player A",
		Position: 0,
		Hand:     []string{"wan-5", "wan-5", "wan-5", "tiao-1"},
		Melds:    []model.Meld{},
	}
	room := &Room{
		ID:      "test-room",
		Players: []*Player{playerA},
		Game:    NewMahjongGame([]*Player{playerA}),
	}
	room.GameStarted = true

	// 嘗試暗槓
	success, _ := room.HandleKong(playerA.ID, "wan-5", true)

	// 應該失敗
	if success {
		t.Error("手牌只有 3 張，暗槓應該失敗")
	}

	// Melds 應該還是空的
	if len(playerA.Melds) != 0 {
		t.Error("暗槓失敗後 Melds 應該是空的")
	}
}

// TestKongTypeDistinction 測試明槓和暗槓的類型區分
func TestKongTypeDistinction(t *testing.T) {
	tests := []struct {
		name         string
		isConcealed  bool
		handCount    int  // 手牌中有幾張要槓的牌
		hasDiscard   bool // 棄牌堆是否有這張牌
		expectedType string
		shouldPass   bool
	}{
		{
			name:         "暗槓-手牌4張",
			isConcealed:  true,
			handCount:    4,
			hasDiscard:   false,
			expectedType: "kong_concealed",
			shouldPass:   true,
		},
		{
			name:         "明槓-手牌3張+棄牌1張",
			isConcealed:  false,
			handCount:    3,
			hasDiscard:   true,
			expectedType: "kong_exposed",
			shouldPass:   true,
		},
		{
			name:         "暗槓失敗-只有3張",
			isConcealed:  true,
			handCount:    3,
			hasDiscard:   false,
			expectedType: "",
			shouldPass:   false,
		},
		{
			name:         "明槓失敗-只有2張",
			isConcealed:  false,
			handCount:    2,
			hasDiscard:   true,
			expectedType: "",
			shouldPass:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 建立手牌
			hand := []string{"tiao-1", "tiao-2"} // 基礎手牌
			for i := 0; i < tt.handCount; i++ {
				hand = append(hand, "wan-8")
			}

			playerA := &Player{
				ID:       "playerA",
				Name:     "Player A",
				Position: 0,
				Hand:     hand,
				Melds:    []model.Meld{},
			}
			playerB := &Player{ID: "playerB", Name: "Player B", Position: 1}
			room := &Room{
				ID:      "test-room",
				Players: []*Player{playerA, playerB},
				Game:    NewMahjongGame([]*Player{playerA, playerB}),
			}
			room.GameStarted = true

			if tt.hasDiscard {
				room.Game.DiscardPile = []string{"wan-8"}
			}

			success, _ := room.HandleKong(playerA.ID, "wan-8", tt.isConcealed)

			if success != tt.shouldPass {
				t.Errorf("預期成功=%v，實際=%v", tt.shouldPass, success)
			}

			if tt.shouldPass {
				if len(playerA.Melds) != 1 {
					t.Fatalf("應該有 1 組 meld，實際: %d", len(playerA.Melds))
				}
				if playerA.Melds[0].Type != tt.expectedType {
					t.Errorf("預期 meld 類型是 '%s'，實際是 '%s'", tt.expectedType, playerA.Melds[0].Type)
				}
			}
		})
	}
}

// TestHandleKong_PromotedKongDoesNotRemoveFromDiscardPile 測試加槓不會從棄牌堆移除牌
func TestHandleKong_PromotedKongDoesNotRemoveFromDiscardPile(t *testing.T) {
	// Setup: 玩家已碰 wan-5，然後自己摸到第4張 wan-5
	playerA := &Player{
		ID:       "playerA",
		Name:     "Player A",
		Position: 0,
		Hand:     []string{"tiao-1", "tiao-2", "wan-5"}, // 手上有一張 wan-5（自己摸到的）
		Melds: []model.Meld{
			{Type: "pong", Tiles: []string{"wan-5", "wan-5", "wan-5"}},
		},
	}
	playerB := &Player{ID: "playerB", Name: "Player B", Position: 1}
	room := &Room{
		ID:      "test-room",
		Players: []*Player{playerA, playerB},
		Game:    NewMahjongGame([]*Player{playerA, playerB}),
	}
	room.GameStarted = true

	// 棄牌堆有其他牌（但不是 wan-5）
	room.Game.DiscardPile = []string{"tong-1", "tong-2", "tong-3"}
	originalDiscardCount := len(room.Game.DiscardPile)

	// 執行加槓
	success, _ := room.HandleKong(playerA.ID, "wan-5", false)

	if !success {
		t.Fatalf("加槓應該成功")
	}

	// 驗證棄牌堆沒有被修改（因為是自己摸到的牌，不是別人打出的）
	if len(room.Game.DiscardPile) != originalDiscardCount {
		t.Errorf("加槓不應該修改棄牌堆，原本 %d 張，現在 %d 張", originalDiscardCount, len(room.Game.DiscardPile))
	}

	// 驗證 meld 類型是 kong_promoted
	if playerA.Melds[0].Type != "kong_promoted" {
		t.Errorf("Meld 類型應該是 'kong_promoted'，但是 '%s'", playerA.Melds[0].Type)
	}
}

// TestHandleKong_ExposedKongRemovesFromDiscardPile 測試大明槓會從棄牌堆移除牌
func TestHandleKong_ExposedKongRemovesFromDiscardPile(t *testing.T) {
	// Setup: 玩家手上有3張 wan-5，別人打出第4張
	playerA := &Player{
		ID:       "playerA",
		Name:     "Player A",
		Position: 0,
		Hand:     []string{"tiao-1", "wan-5", "wan-5", "wan-5"},
		Melds:    []model.Meld{},
	}
	playerB := &Player{ID: "playerB", Name: "Player B", Position: 1}
	room := &Room{
		ID:      "test-room",
		Players: []*Player{playerA, playerB},
		Game:    NewMahjongGame([]*Player{playerA, playerB}),
	}
	room.GameStarted = true

	// 棄牌堆最後一張是 wan-5（別人打出的）
	room.Game.DiscardPile = []string{"tong-1", "tong-2", "wan-5"}
	room.LastDiscardPlayer = playerB.Position

	// 執行大明槓
	success, _ := room.HandleKong(playerA.ID, "wan-5", false)

	if !success {
		t.Fatalf("大明槓應該成功")
	}

	// 驗證棄牌堆的 wan-5 被移除
	if len(room.Game.DiscardPile) != 2 {
		t.Errorf("大明槓應該從棄牌堆移除牌，預期 2 張，實際 %d 張", len(room.Game.DiscardPile))
	}
	for _, tile := range room.Game.DiscardPile {
		if tile == "wan-5" {
			t.Error("棄牌堆不應該還有 wan-5")
		}
	}

	// 驗證 meld 類型是 kong_exposed
	if playerA.Melds[0].Type != "kong_exposed" {
		t.Errorf("Meld 類型應該是 'kong_exposed'，但是 '%s'", playerA.Melds[0].Type)
	}
}

// TestHandleKong_ConcealedKongDoesNotTouchDiscardPile 測試暗槓不會影響棄牌堆
func TestHandleKong_ConcealedKongDoesNotTouchDiscardPile(t *testing.T) {
	// Setup: 玩家手上有4張 wan-5
	playerA := &Player{
		ID:       "playerA",
		Name:     "Player A",
		Position: 0,
		Hand:     []string{"tiao-1", "wan-5", "wan-5", "wan-5", "wan-5"},
		Melds:    []model.Meld{},
	}
	room := &Room{
		ID:      "test-room",
		Players: []*Player{playerA},
		Game:    NewMahjongGame([]*Player{playerA}),
	}
	room.GameStarted = true

	// 棄牌堆有其他牌
	room.Game.DiscardPile = []string{"tong-1", "tong-2", "tong-3"}
	originalDiscardCount := len(room.Game.DiscardPile)

	// 執行暗槓
	success, _ := room.HandleKong(playerA.ID, "wan-5", true)

	if !success {
		t.Fatalf("暗槓應該成功")
	}

	// 驗證棄牌堆沒有被修改
	if len(room.Game.DiscardPile) != originalDiscardCount {
		t.Errorf("暗槓不應該修改棄牌堆，原本 %d 張，現在 %d 張", originalDiscardCount, len(room.Game.DiscardPile))
	}

	// 驗證 meld 類型是 kong_concealed
	if playerA.Melds[0].Type != "kong_concealed" {
		t.Errorf("Meld 類型應該是 'kong_concealed'，但是 '%s'", playerA.Melds[0].Type)
	}
}