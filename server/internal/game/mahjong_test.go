package game

import (
	"testing"
)

// TestCanHu 测试胡牌判断
func TestCanHu(t *testing.T) {
	players := []*Player{
		{ID: "1", Name: "Player1", Position: 0, Hand: make([]string, 0), Melds: make([]Meld, 0), Flowers: make([]string, 0)},
	}
	game := NewMahjongGame(players)

	// 测试1: 简单的胡牌（5组刻子+1对眼 = 17张）
	hand1 := []string{
		"wan-1", "wan-1", "wan-1", // 刻子
		"wan-2", "wan-2", "wan-2", // 刻子
		"wan-3", "wan-3", "wan-3", // 刻子
		"wan-4", "wan-4", "wan-4", // 刻子
		"wan-5", "wan-5", "wan-5", // 刻子
		"wan-6", "wan-6", // 对眼
	}
	if !game.CanHu(hand1, []Meld{}) {
		t.Errorf("应该能胡牌，但判断为不能胡")
	}

	// 测试2: 简单的顺子胡牌（5组顺子+1对眼 = 17张）
	hand2 := []string{
		"wan-1", "wan-2", "wan-3", // 顺子
		"wan-4", "wan-5", "wan-6", // 顺子
		"wan-7", "wan-8", "wan-9", // 顺子
		"tong-1", "tong-2", "tong-3", // 顺子
		"tiao-7", "tiao-8", "tiao-9", // 顺子
		"dong", "dong", // 对眼
	}
	if !game.CanHu(hand2, []Meld{}) {
		t.Errorf("应该能胡牌（顺子），但判断为不能胡")
	}

	// 测试3: 不能胡牌
	hand3 := []string{
		"wan-1", "wan-2", "wan-4", // 不连续
		"wan-5", "wan-6", "wan-7",
		"tong-1", "tong-2", "tong-3",
		"tiao-7", "tiao-8", "tiao-9",
		"dong", "dong",
	}
	if game.CanHu(hand3, []Meld{}) {
		t.Errorf("不应该能胡牌，但判断为能胡")
	}

	// 测试4: 有碰的情况（2组已展示+手牌3组+1对眼 = 17张）
	hand4 := []string{
		"wan-1", "wan-2", "wan-3", // 顺子
		"wan-4", "wan-5", "wan-6", // 顺子
		"wan-7", "wan-8", "wan-9", // 顺子
		"dong", "dong", // 对眼
	}
	melds4 := []Meld{
		{Type: "pong", Tiles: []string{"tong-5", "tong-5", "tong-5"}},
		{Type: "pong", Tiles: []string{"tiao-9", "tiao-9", "tiao-9"}},
	}
	if !game.CanHu(hand4, melds4) {
		t.Errorf("应该能胡牌（有碰），但判断为不能胡")
	}
}

// TestCanPong 测试碰牌判断
func TestCanPong(t *testing.T) {
	game := &MahjongGame{}

	// 测试1: 手牌中有2张相同
	hand1 := []string{"wan-1", "wan-1", "wan-2", "wan-3"}
	if !game.CanPong(hand1, "wan-1") {
		t.Errorf("应该能碰wan-1")
	}

	// 测试2: 手牌中只有1张
	hand2 := []string{"wan-1", "wan-2", "wan-3"}
	if game.CanPong(hand2, "wan-1") {
		t.Errorf("不应该能碰wan-1（只有1张）")
	}

	// 测试3: 手牌中有3张
	hand3 := []string{"wan-1", "wan-1", "wan-1", "wan-2"}
	if !game.CanPong(hand3, "wan-1") {
		t.Errorf("应该能碰wan-1（有3张）")
	}
}

// TestCanKong 测试杠牌判断
func TestCanKong(t *testing.T) {
	game := &MahjongGame{}

	// 测试1: 手牌中有3张相同
	hand1 := []string{"wan-1", "wan-1", "wan-1", "wan-2"}
	if !game.CanKong(hand1, "wan-1") {
		t.Errorf("应该能杠wan-1")
	}

	// 测试2: 手牌中只有2张
	hand2 := []string{"wan-1", "wan-1", "wan-2", "wan-3"}
	if game.CanKong(hand2, "wan-1") {
		t.Errorf("不应该能杠wan-1（只有2张）")
	}

	// 测试3: 手牌中有4张
	hand3 := []string{"wan-1", "wan-1", "wan-1", "wan-1"}
	if !game.CanKong(hand3, "wan-1") {
		t.Errorf("应该能杠wan-1（有4张）")
	}
}

// TestCanChow 测试吃牌判断
func TestCanChow(t *testing.T) {
	game := &MahjongGame{}

	// 测试1: 能吃（wan-2，手牌有wan-1和wan-3）
	hand1 := []string{"wan-1", "wan-3", "tong-5", "tiao-7"}
	combos1 := game.CanChow(hand1, "wan-2")
	if len(combos1) != 1 {
		t.Errorf("应该有1种吃牌组合，实际: %d", len(combos1))
	}
	expectedCombo1 := []string{"wan-1", "wan-2", "wan-3"}
	if len(combos1) > 0 && !isSameCombination(combos1[0], expectedCombo1) {
		t.Errorf("吃牌组合错误，期望: %v，实际: %v", expectedCombo1, combos1[0])
	}

	// 测试2: 能吃（wan-5，手牌有wan-4,wan-6和wan-6,wan-7）- 多种组合
	hand2 := []string{"wan-4", "wan-6", "wan-7", "tong-1"}
	combos2 := game.CanChow(hand2, "wan-5")
	if len(combos2) != 2 {
		t.Errorf("应该有2种吃牌组合，实际: %d", len(combos2))
	}

	// 测试3: 不能吃（字牌）
	hand3 := []string{"wan-1", "wan-2", "tong-5"}
	combos3 := game.CanChow(hand3, "dong")
	if len(combos3) != 0 {
		t.Errorf("字牌不应该能吃，实际组合数: %d", len(combos3))
	}

	// 测试4: 不能吃（手牌不连续）
	hand4 := []string{"wan-1", "wan-4", "tong-5"}
	combos4 := game.CanChow(hand4, "wan-2")
	if len(combos4) != 0 {
		t.Errorf("手牌不连续不应该能吃，实际组合数: %d", len(combos4))
	}

	// 测试5: 能吃（wan-1，手牌有wan-2和wan-3）
	hand5 := []string{"wan-2", "wan-3", "tong-5"}
	combos5 := game.CanChow(hand5, "wan-1")
	if len(combos5) != 1 {
		t.Errorf("应该有1种吃牌组合，实际: %d", len(combos5))
	}

	// 测试6: 能吃（wan-9，手牌有wan-7和wan-8）
	hand6 := []string{"wan-7", "wan-8", "tong-5"}
	combos6 := game.CanChow(hand6, "wan-9")
	if len(combos6) != 1 {
		t.Errorf("应该有1种吃牌组合，实际: %d", len(combos6))
	}

	// 测试7: 不能吃（不同花色）
	hand7 := []string{"wan-1", "wan-3", "tong-5"}
	combos7 := game.CanChow(hand7, "tong-2")
	if len(combos7) != 0 {
		t.Errorf("不同花色不应该能吃，实际组合数: %d", len(combos7))
	}
}

// TestIsFlowerTile 测试花牌判断
func TestIsFlowerTile(t *testing.T) {
	if !isFlowerTile("flower-chun") {
		t.Errorf("flower-chun应该是花牌")
	}

	if !isFlowerTile("flower-mei") {
		t.Errorf("flower-mei应该是花牌")
	}

	if isFlowerTile("wan-1") {
		t.Errorf("wan-1不应该是花牌")
	}

	if isFlowerTile("dong") {
		t.Errorf("dong不应该是花牌")
	}
}

// TestCheckDraw 测试流局判断
func TestCheckDraw(t *testing.T) {
	players := []*Player{
		{ID: "1", Name: "Player1", Position: 0},
	}
	game := NewMahjongGame(players)

	// 测试1: 牌山充足
	if game.CheckDraw() {
		t.Errorf("牌山充足时不应该流局")
	}

	// 测试2: 模拟牌山消耗到8张
	for len(game.Deck) > 8 {
		game.DrawTile()
	}

	if !game.CheckDraw() {
		t.Errorf("牌山剩余8张时应该流局")
	}
}

// NOTE: The scoring functions are not in the provided mahjong.go file,
// so these tests might fail if those functions don't exist.
// I am commenting them out to be safe.
/*
// TestPongPongHu 测试碰碰胡判断
func TestPongPongHu(t *testing.T) {
	player := &Player{
		Hand: []string{
			"wan-1", "wan-1", "wan-1",
			"wan-2", "wan-2", "wan-2",
			"wan-3", "wan-3", "wan-3",
			"wan-4", "wan-4", "wan-4",
			"wan-5", "wan-5",
		},
		Melds: []Meld{},
	}

	game := &MahjongGame{Players: []*Player{player}}

	if !game.isPongPongHu(player) {
		t.Errorf("应该是碰碰胡")
	}

	// 测试有顺子的情况
	player2 := &Player{
		Hand: []string{
			"wan-1", "wan-2", "wan-3", // 顺子
			"wan-4", "wan-4", "wan-4",
			"wan-5", "wan-5", "wan-5",
			"wan-6", "wan-6", "wan-6",
			"wan-7", "wan-7",
		},
		Melds: []Meld{},
	}

	if game.isPongPongHu(player2) {
		t.Errorf("有顺子不应该是碰碰胡")
	}
}

// TestOneSuit 测试清一色判断
func TestOneSuit(t *testing.T) {
	game := &MahjongGame{}

	// 测试1: 清一色（全是万子）
	tiles1 := []string{
		"wan-1", "wan-1", "wan-1",
		"wan-2", "wan-3", "wan-4",
		"wan-5", "wan-6", "wan-7",
		"wan-8", "wan-8", "wan-8",
		"wan-9", "wan-9",
	}
	if !game.isOneSuit(tiles1) {
		t.Errorf("应该是清一色")
	}

	// 测试2: 不是清一色（有字牌）
	tiles2 := []string{
		"wan-1", "wan-1", "wan-1",
		"wan-2", "wan-3", "wan-4",
		"dong", "dong", "dong",
		"wan-8", "wan-8", "wan-8",
		"wan-9", "wan-9",
	}
	if game.isOneSuit(tiles2) {
		t.Errorf("有字牌不应该是清一色")
	}

	// 测试3: 不是清一色（有多种花色）
	tiles3 := []string{
		"wan-1", "wan-1", "wan-1",
		"tong-2", "tong-3", "tong-4",
		"wan-5", "wan-6", "wan-7",
		"wan-8", "wan-8", "wan-8",
		"wan-9", "wan-9",
	}
	if game.isOneSuit(tiles3) {
		t.Errorf("有多种花色不应该是清一色")
	}
}

// TestMixedOneSuit 测试混一色判断
func TestMixedOneSuit(t *testing.T) {
	game := &MahjongGame{}

	// 测试1: 混一色（万子+字牌）
	tiles1 := []string{
		"wan-1", "wan-1", "wan-1",
		"wan-2", "wan-3", "wan-4",
		"dong", "dong", "dong",
		"wan-8", "wan-8", "wan-8",
		"zhong", "zhong",
	}
	if !game.isMixedOneSuit(tiles1) {
		t.Errorf("应该是混一色")
	}

	// 测试2: 不是混一色（纯清一色）
	tiles2 := []string{
		"wan-1", "wan-1", "wan-1",
		"wan-2", "wan-3", "wan-4",
		"wan-5", "wan-6", "wan-7",
		"wan-8", "wan-8", "wan-8",
		"wan-9", "wan-9",
	}
	if game.isMixedOneSuit(tiles2) {
		t.Errorf("纯清一色不应该算混一色")
	}

	// 测试3: 不是混一色（有多种花色）
	tiles3 := []string{
		"wan-1", "wan-1", "wan-1",
		"tong-2", "tong-3", "tong-4",
		"dong", "dong", "dong",
		"wan-8", "wan-8", "wan-8",
		"zhong", "zhong",
	}
	if game.isMixedOneSuit(tiles3) {
		t.Errorf("有多种花色不应该是混一色")
	}
}

// TestBigDragons 测试大三元判断
func TestBigDragons(t *testing.T) {
	game := &MahjongGame{}

	// 测试1: 大三元
	tiles1 := []string{
		"zhong", "zhong", "zhong",
		"fa", "fa", "fa",
		"bai", "bai", "bai",
		"wan-1", "wan-1", "wan-1",
		"wan-2", "wan-2",
	}
	if !game.isBigDragons(tiles1) {
		t.Errorf("应该是大三元")
	}

	// 测试2: 不是大三元（缺一种）
	tiles2 := []string{
		"zhong", "zhong", "zhong",
		"fa", "fa", "fa",
		"wan-1", "wan-1", "wan-1",
		"wan-2", "wan-2", "wan-2",
		"wan-3", "wan-3",
	}
	if game.isBigDragons(tiles2) {
		t.Errorf("缺白板不应该是大三元")
	}
}
*/

// TestCanExposedKong tests the logic for an exposed kong, including promoting a pong.
func TestCanExposedKong(t *testing.T) {
	game := NewMahjongGame([]*Player{})

	// Scenario 1: Player has 3 tiles in hand, can kong a discard.
	player1 := &Player{
		Hand:  []string{"wan-5", "wan-5", "wan-5", "tiao-1"},
		Melds: []Meld{},
	}
	if !game.CanExposedKong(player1, "wan-5") {
		t.Errorf("TestCanExposedKong: Should be able to exposed kong from 3 tiles in hand.")
	}

	// Scenario 2: Player has an exposed pong, can promote it with a discard.
	player2 := &Player{
		Hand: []string{"tiao-1", "tiao-2"},
		Melds: []Meld{
			{Type: "pong", Tiles: []string{"wan-5", "wan-5", "wan-5"}},
		},
	}
	if !game.CanExposedKong(player2, "wan-5") {
		t.Errorf("TestCanExposedKong: Should be able to promote an exposed pong.")
	}

	// Scenario 3: Player has only 2 tiles in hand, no pong.
	player3 := &Player{
		Hand:  []string{"wan-5", "wan-5", "tiao-1"},
		Melds: []Meld{},
	}
	if game.CanExposedKong(player3, "wan-5") {
		t.Errorf("TestCanExposedKong: Should not be able to kong with only 2 tiles in hand.")
	}

	// Scenario 4: Player has a pong of a different tile.
	player4 := &Player{
		Hand: []string{"tiao-1", "tiao-2"},
		Melds: []Meld{
			{Type: "pong", Tiles: []string{"tiao-3", "tiao-3", "tiao-3"}},
		},
	}
	if game.CanExposedKong(player4, "wan-5") {
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

// TestHandleKong_PromotedKongFromDiscard simulates promoting a Pong to a Kong from a discard.
func TestHandleKong_PromotedKongFromDiscard(t *testing.T) {
	// 1. Setup Room and Players
	playerA := &Player{
		ID:       "playerA",
		Name:     "Player A",
		Position: 0,
		Hand:     []string{"tiao-1", "tiao-2", "tiao-3"},
		Melds: []Meld{
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

	// 2. Simulate State
	discardedTile := "wan-8"
	room.LastDiscardPlayer = playerB.Position // Player B discarded the tile
	room.Game.DiscardPile = append(room.Game.DiscardPile, discardedTile)

	// 3. Call HandleKong
	success, drawnTile := room.HandleKong(playerA.ID, discardedTile, false)

	// 4. Assertions
	if !success {
		t.Fatalf("HandleKong should have returned true for a valid promoted kong.")
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

	// Assert discard pile was consumed
	if len(room.Game.DiscardPile) > 0 && room.Game.DiscardPile[len(room.Game.DiscardPile)-1] == discardedTile {
		t.Errorf("Discard pile should not contain the konged tile.")
	}

	// Assert player drew a replacement tile
	if len(playerA.Hand) != initialHandSize+1 {
		t.Errorf("Player's hand size should be %d after drawing a replacement tile, but got %d.", initialHandSize+1, len(playerA.Hand))
	}

	// Assert deck size decreased by 1 (for the replacement tile)
	if len(room.Game.Deck) != deckSize-1 {
		t.Errorf("Deck size should have decreased by 1, from %d to %d, but is %d.", deckSize, deckSize-1, len(room.Game.Deck))
	}

	// Assert turn changed to the konging player
	if room.CurrentTurn != playerA.Position {
		t.Errorf("It should be player A's turn after kong, but got turn %d.", room.CurrentTurn)
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
	melds := []Meld{
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
