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