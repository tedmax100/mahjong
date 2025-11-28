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

	// 测试1: 简单的胡牌（4组刻子+1对眼）
	hand1 := []string{
		"wan-1", "wan-1", "wan-1", // 刻子
		"wan-2", "wan-2", "wan-2", // 刻子
		"wan-3", "wan-3", "wan-3", // 刻子
		"wan-4", "wan-4", "wan-4", // 刻子
		"wan-5", "wan-5", // 对眼
	}
	if !game.CanHu(hand1, []Meld{}) {
		t.Errorf("应该能胡牌，但判断为不能胡")
	}

	// 测试2: 简单的顺子胡牌
	hand2 := []string{
		"wan-1", "wan-2", "wan-3", // 顺子
		"wan-4", "wan-5", "wan-6", // 顺子
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

	// 测试4: 有碰的情况（2组已展示+手牌2组+1对眼）
	hand4 := []string{
		"wan-1", "wan-2", "wan-3", // 顺子
		"wan-4", "wan-5", "wan-6", // 顺子
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
