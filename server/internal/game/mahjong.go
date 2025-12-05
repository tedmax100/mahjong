package game

import (
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Meld 代表一組已展示的牌（碰、槓）
type Meld struct {
	Type  string   // "pong"(碰), "kong"(槓), "chow"(吃)
	Tiles []string // 牌組
}

// MahjongGame 代表一局麻將遊戲
type MahjongGame struct {
	Players      []*Player
	Deck         []string
	CurrentTurn  int
	DiscardPile  []string
	Dealer       int  // 莊家位置
	IsDraw       bool // 是否流局
}

// NewMahjongGame 建立新遊戲
func NewMahjongGame(players []*Player) *MahjongGame {
	game := &MahjongGame{
		Players:      players,
		CurrentTurn:  0,
		DiscardPile:  make([]string, 0),
		Dealer:       0,
		IsDraw:       false,
	}

	game.initializeDeck()
	return game
}

// initializeDeck 初始化牌組
func (g *MahjongGame) initializeDeck() {
	g.Deck = make([]string, 0, 144)

	// 萬子 1-9，每種 4 張
	for i := 1; i <= 9; i++ {
		for j := 0; j < 4; j++ {
			g.Deck = append(g.Deck, tileID("wan", i))
		}
	}

	// 筒子 1-9，每種 4 張
	for i := 1; i <= 9; i++ {
		for j := 0; j < 4; j++ {
			g.Deck = append(g.Deck, tileID("tong", i))
		}
	}

	// 条子 1-9，每種 4 張
	for i := 1; i <= 9; i++ {
		for j := 0; j < 4; j++ {
			g.Deck = append(g.Deck, tileID("tiao", i))
		}
	}

	// 風牌，每種 4 張
	winds := []string{"dong", "nan", "xi", "bei"}
	for _, wind := range winds {
		for j := 0; j < 4; j++ {
			g.Deck = append(g.Deck, wind)
		}
	}

	// 三元牌，每種 4 張
	dragons := []string{"zhong", "fa", "bai"}
	for _, dragon := range dragons {
		for j := 0; j < 4; j++ {
			g.Deck = append(g.Deck, dragon)
		}
	}

	// 花牌，每種 1 張
	flowers := []string{
		"flower-chun", "flower-xia", "flower-qiu", "flower-dong",
		"flower-mei", "flower-lan", "flower-zhu", "flower-ju",
	}
	g.Deck = append(g.Deck, flowers...)

	// 洗牌
	g.shuffleDeck()
}

// shuffleDeck 洗牌
func (g *MahjongGame) shuffleDeck() {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(g.Deck), func(i, j int) {
		g.Deck[i], g.Deck[j] = g.Deck[j], g.Deck[i]
	})
}

// DealTiles 發牌（台灣 16 張麻將）
func (g *MahjongGame) DealTiles() {
	// 台灣 16 張麻將：每人發 16 張牌
	for _, player := range g.Players {
		player.Hand = make([]string, 0, 17)

		for i := 0; i < 16; i++ {
			if len(g.Deck) > 0 {
				tile := g.DrawTile()

				// 如果是花牌，加入花牌集合並補牌
				if isFlowerTile(tile) {
					player.Flowers = append(player.Flowers, tile)
					i-- // 重新抽一張（不計入手牌數量）
					continue
				}

				player.Hand = append(player.Hand, tile)
			}
		}

		// 莊家多一張
		if player.Position == g.Dealer && len(g.Deck) > 0 {
			tile := g.DrawTile()
			// 檢查是否為花牌
			for isFlowerTile(tile) && len(g.Deck) > 0 {
				player.Flowers = append(player.Flowers, tile)
				tile = g.DrawTile()
			}
			if !isFlowerTile(tile) {
				player.Hand = append(player.Hand, tile)
			}
		}
	}
}

// DrawTile 摸牌
func (g *MahjongGame) DrawTile() string {
	if len(g.Deck) > 0 {
		tile := g.Deck[0]
		g.Deck = g.Deck[1:]
		return tile
	}
	return ""
}

// DrawTileFromEnd 從牌山尾部摸一張牌（用於補槓）
func (g *MahjongGame) DrawTileFromEnd() string {
	if len(g.Deck) > 0 {
		tile := g.Deck[len(g.Deck)-1]
		g.Deck = g.Deck[:len(g.Deck)-1]
		return tile
	}
	return ""
}

// DrawTileWithFlowerReplacement 摸牌（自動處理花牌補牌）
func (g *MahjongGame) DrawTileWithFlowerReplacement(player *Player) string {
	// 檢查是否流局
	if g.CheckDraw() {
		return ""
	}

	tile := g.DrawTile()

	// 如果摸到花牌，加入花牌集合並繼續補牌
	for isFlowerTile(tile) && len(g.Deck) > 0 {
		player.Flowers = append(player.Flowers, tile)

		// 再次檢查是否流局
		if g.CheckDraw() {
			return ""
		}

		tile = g.DrawTile()
	}

	return tile
}

// CheckDraw 檢查是否流局（海底）
func (g *MahjongGame) CheckDraw() bool {
	// 台灣 16 張麻將：當牌山剩餘 8 張或更少時流局
	if len(g.Deck) <= 8 {
		g.IsDraw = true
		return true
	}
	return false
}

// GetRemainingTiles 取得牌山剩餘張數
func (g *MahjongGame) GetRemainingTiles() int {
	return len(g.Deck)
}

// NextTurn 下一回合
func (g *MahjongGame) NextTurn() {
	g.CurrentTurn = (g.CurrentTurn + 1) % len(g.Players)
}

// tileID 生成牌 ID
func tileID(typ string, num int) string {
	if num == 0 {
		return typ
	}
	return typ + "-" + string(rune('0'+num))
}

// isFlowerTile 判斷是否為花牌
func isFlowerTile(tile string) bool {
	return len(tile) > 6 && tile[:6] == "flower"
}

// CanPong 檢查是否可以碰
func (g *MahjongGame) CanPong(hand []string, tile string) bool {
	count := 0
	for _, t := range hand {
		if t == tile {
			count++
		}
	}
	return count >= 2
}

// CanKong 檢查是否可以槓
func (g *MahjongGame) CanKong(hand []string, tile string) bool {
	count := 0
	for _, t := range hand {
		if t == tile {
			count++
		}
	}
	return count >= 3
}

// CanExposedKong 檢查是否可以明槓（包含加槓）
// isSelfDrawn: true 表示玩家自己摸到牌，false 表示別人打出的牌
func (g *MahjongGame) CanExposedKong(player *Player, tile string, isSelfDrawn bool) bool {
	// 檢查手牌中是否有三張（明槓：手上 3 張+別人打出 1 張）
	handCount := 0
	for _, t := range player.Hand {
		if t == tile {
			handCount++
		}
	}
	if handCount >= 3 {
		return true
	}

	// 檢查是否已有碰牌，可以加槓
	// 重要：加槓只能在自己摸到牌時進行，不能透過別人打出的牌來加槓
	if isSelfDrawn {
		for _, meld := range player.Melds {
			if meld.Type == "pong" && meld.Tiles[0] == tile {
				return true
			}
		}
	}

	return false
}

// CanChow 檢查是否可以吃牌（只能吃上家打出的牌）
// 返回所有可能的吃牌組合
func (g *MahjongGame) CanChow(hand []string, tile string) [][]string {
	// 字牌（風牌、三元牌）不能吃
	if isHonorTile(tile) {
		return nil
	}

	typ, num := parseTile(tile)
	if num == 0 {
		return nil
	}

	validCombinations := make([][]string, 0)

	// 情況 1: tile + tile+1 + tile+2 (例如: 1,2,3)
	if num <= 7 {
		tile2 := tileID(typ, num+1)
		tile3 := tileID(typ, num+2)
		if countTile(hand, tile2) > 0 && countTile(hand, tile3) > 0 {
			validCombinations = append(validCombinations, []string{tile, tile2, tile3})
		}
	}

	// 情況 2: tile-1 + tile + tile+1 (例如: 2,3,4)
	if num >= 2 && num <= 8 {
		tile1 := tileID(typ, num-1)
		tile3 := tileID(typ, num+1)
		if countTile(hand, tile1) > 0 && countTile(hand, tile3) > 0 {
			validCombinations = append(validCombinations, []string{tile1, tile, tile3})
		}
	}

	// 情況 3: tile-2 + tile-1 + tile (例如: 3,4,5)
	if num >= 3 {
		tile1 := tileID(typ, num-2)
		tile2 := tileID(typ, num-1)
		if countTile(hand, tile1) > 0 && countTile(hand, tile2) > 0 {
			validCombinations = append(validCombinations, []string{tile1, tile2, tile})
		}
	}

	return validCombinations
}

// CanHu 檢查是否可以胡牌
func (g *MahjongGame) CanHu(hand []string, melds []Meld) bool {
	// 台灣 16 張麻將：胡牌牌型 = 5 組面子（順子/刻子） + 1 對眼 = 17 張牌
	meldCount := len(melds)
	needGroups := 5 - meldCount

	// 手牌數量檢查
	expectedHandSize := needGroups*3 + 2
	if len(hand) != expectedHandSize {
		return false
	}

	tiles := make([]string, len(hand))
	copy(tiles, hand)
	sortTiles(tiles)

	// 遍歷所有可能的對子
	for i := 0; i < len(tiles)-1; i++ {
		// 跳過重複的對子檢查，提高效率
		if i > 0 && tiles[i] == tiles[i-1] {
			continue
		}

		if tiles[i] == tiles[i+1] {
			// 找到一個潛在的對子
			remaining := make([]string, 0, len(tiles)-2)
			remaining = append(remaining, tiles[:i]...)
			remaining = append(remaining, tiles[i+2:]...)

			// 檢查剩下的牌是否能組成 needGroups 個面子
			if canFormGroups(remaining, needGroups) {
				return true
			}
		}
	}

	return false
}

// canFormGroups 遞迴檢查手牌是否能完全由刻子或順子組成
func canFormGroups(tiles []string, needGroups int) bool {
	if len(tiles) == 0 {
		return needGroups == 0
	}
	// 如果需要的組數和牌數不匹配，則失敗
	if needGroups*3 != len(tiles) {
		return false
	}

	// 總是從第一張牌開始嘗試組合
	tile := tiles[0]

	// 嘗試移除刻子
	if countTile(tiles, tile) >= 3 {
		newTiles := removeTiles(tiles, tile, 3)
		if canFormGroups(newTiles, needGroups-1) {
			return true
		}
	}

	// 嘗試移除順子
	if isSequencePossible(tiles, tile) {
		newTiles := removeSequence(tiles, tile)
		if canFormGroups(newTiles, needGroups-1) {
			return true
		}
	}

	// 如果第一張牌無法組成任何組合，則此路不通
	return false
}

// countTile 計算某張牌在手牌中的數量
func countTile(tiles []string, tile string) int {
	count := 0
	for _, t := range tiles {
		if t == tile {
			count++
		}
	}
	return count
}

// removeTiles 從手牌中移除指定數量的牌
func removeTiles(tiles []string, tile string, count int) []string {
	result := make([]string, 0, len(tiles))
	removed := 0
	for _, t := range tiles {
		if t == tile && removed < count {
			removed++
			continue
		}
		result = append(result, t)
	}
	return result
}

// isSequencePossible 檢查是否可能組成順子
func isSequencePossible(tiles []string, tile string) bool {
	// 字牌（風牌、三元牌）不能組成順子
	if isHonorTile(tile) {
		return false
	}

	typ, num := parseTile(tile)
	if num == 0 || num > 7 {
		return false // 8、9 不能作為順子開頭
	}

	// 檢查是否有連續的三張
	tile2 := tileID(typ, num+1)
	tile3 := tileID(typ, num+2)

	return countTile(tiles, tile) > 0 &&
	       countTile(tiles, tile2) > 0 &&
	       countTile(tiles, tile3) > 0
}

// removeSequence 移除一組順子
func removeSequence(tiles []string, tile string) []string {
	typ, num := parseTile(tile)
	if num == 0 {
		return tiles
	}

	tile2 := tileID(typ, num+1)
	tile3 := tileID(typ, num+2)

	result := removeTiles(tiles, tile, 1)
	result = removeTiles(result, tile2, 1)
	result = removeTiles(result, tile3, 1)

	return result
}

// isHonorTile 判斷是否為字牌（風牌、三元牌）
func isHonorTile(tile string) bool {
	honors := []string{"dong", "nan", "xi", "bei", "zhong", "fa", "bai"}
	for _, h := range honors {
		if tile == h {
			return true
		}
	}
	return false
}

// parseTile 解析牌的類型和數字
func parseTile(tile string) (string, int) {
	if len(tile) < 5 { // 字牌
		return tile, 0
	}

	// 格式: "wan-1", "tong-5" 等
	typ := tile[:len(tile)-2]
	num := int(tile[len(tile)-1] - '0')
	return typ, num
}

// tileValue 返回牌的可排序值
func tileValue(tile string) int {
	parts := strings.Split(tile, "-")
	suit := parts[0]
	num := 0
	if len(parts) > 1 {
		num, _ = strconv.Atoi(parts[1])
	}

	suitOrder := 0
	switch suit {
	case "wan":
		suitOrder = 1
	case "tong":
		suitOrder = 2
	case "tiao":
		suitOrder = 3
	case "dong":
		suitOrder = 4
		num = 1
	case "nan":
		suitOrder = 4
		num = 2
	case "xi":
		suitOrder = 4
		num = 3
	case "bei":
		suitOrder = 4
		num = 4
	case "zhong":
		suitOrder = 5
		num = 1
	case "fa":
		suitOrder = 5
		num = 2
	case "bai":
		suitOrder = 5
		num = 3
	}

	return suitOrder*10 + num
}

// sortTiles 對牌的切片進行排序
func sortTiles(tiles []string) {
	sort.Slice(tiles, func(i, j int) bool {
		return tileValue(tiles[i]) < tileValue(tiles[j])
	})
}

// TingResult 保存聽牌檢查的結果
type TingResult struct {
	IsTing       bool
	WinningTiles []string
}

// getUniqueTileTypes 返回所有 34 種唯一牌型的切片
func (g *MahjongGame) getUniqueTileTypes() []string {
	return []string{
		"wan-1", "wan-2", "wan-3", "wan-4", "wan-5", "wan-6", "wan-7", "wan-8", "wan-9",
		"tong-1", "tong-2", "tong-3", "tong-4", "tong-5", "tong-6", "tong-7", "tong-8", "tong-9",
		"tiao-1", "tiao-2", "tiao-3", "tiao-4", "tiao-5", "tiao-6", "tiao-7", "tiao-8", "tiao-9",
		"dong", "nan", "xi", "bei",
		"zhong", "fa", "bai",
	}
}

// CheckTing 確定手牌是否還差一張牌就能胡牌（聽牌）
func (g *MahjongGame) CheckTing(hand []string, melds []Meld) TingResult {
	// 手牌必須有特定數量的牌才能處於聽牌狀態
	// 對於 16 張牌的手牌，它是 4 組 + 1 對 = 17 張牌
	// 聽牌的手牌有 16 張牌（或 13、10、7、4），即 (N*3 + 1)
	// 這個檢查是一個簡化，可能無法涵蓋所有情況，但作為起點是不錯的
	if len(hand)%3 != 1 {
		return TingResult{IsTing: false}
	}

	winningTiles := make([]string, 0)
	checkedTiles := make(map[string]bool)

	allTileTypes := g.getUniqueTileTypes()

	for _, potentialTile := range allTileTypes {
		// 避免重複檢查同一張牌
		if checkedTiles[potentialTile] {
			continue
		}

		tempHand := append([]string{}, hand...) // 建立副本
		tempHand = append(tempHand, potentialTile)

		if g.CanHu(tempHand, melds) {
			winningTiles = append(winningTiles, potentialTile)
		}
		checkedTiles[potentialTile] = true
	}

	return TingResult{
		IsTing:       len(winningTiles) > 0,
		WinningTiles: winningTiles,
	}
}


// ChooseDiscardAI 為 Bot 選擇要打出的牌
func (g *MahjongGame) ChooseDiscardAI(hand []string) string {
	if len(hand) == 0 {
		return ""
	}

	tileCounts := make(map[string]int)
	for _, tile := range hand {
		tileCounts[tile]++
	}

	// 評估每張牌的價值，分數越低越優先打出
	// 簡單的 AI：孤張 > 邊張 > 中張
	// 字牌孤張最優先
	var bestTileToDiscard string
	lowestScore := 1000

	// 檢查孤立的字牌
	for _, tile := range hand {
		if isHonorTile(tile) && tileCounts[tile] == 1 {
			if 10 < lowestScore {
				lowestScore = 10
				bestTileToDiscard = tile
			}
		}
	}
	if bestTileToDiscard != "" {
		return bestTileToDiscard
	}

	// 檢查孤立的么九牌 (1 和 9)
	for _, tile := range hand {
		typ, num := parseTile(tile)
		if num == 1 || num == 9 {
			if tileCounts[tile] == 1 {
				// 檢查鄰近牌
				hasNeighbor := false
				if num == 1 && (tileCounts[tileID(typ, 2)] > 0) {
					hasNeighbor = true
				}
				if num == 9 && (tileCounts[tileID(typ, 8)] > 0) {
					hasNeighbor = true
				}

				if !hasNeighbor && 20 < lowestScore {
					lowestScore = 20
					bestTileToDiscard = tile
				}
			}
		}
	}
	if bestTileToDiscard != "" {
		return bestTileToDiscard
	}

	// 檢查其他孤立的牌
	for _, tile := range hand {
		typ, num := parseTile(tile)
		if num > 1 && num < 9 {
			if tileCounts[tile] == 1 {
				hasNeighbor := (tileCounts[tileID(typ, num-1)] > 0) || (tileCounts[tileID(typ, num+1)] > 0)
				if !hasNeighbor && 30 < lowestScore {
					lowestScore = 30
					bestTileToDiscard = tile
				}
			}
		}
	}
	if bestTileToDiscard != "" {
		return bestTileToDiscard
	}

	// 如果沒有孤張，則打出一張不成對的牌
	for _, tile := range hand {
		if tileCounts[tile] == 1 {
			return tile
		}
	}

	// 如果都是對子或刻子，打出最後一張
	return hand[len(hand)-1]
}