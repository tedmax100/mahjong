package game

import (
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Meld 代表一组已展示的牌（碰、杠）
type Meld struct {
	Type  string   // "pong"(碰), "kong"(杠), "chow"(吃)
	Tiles []string // 牌组
}

// MahjongGame 代表一局麻将游戏
type MahjongGame struct {
	Players      []*Player
	Deck         []string
	CurrentTurn  int
	DiscardPile  []string
	Dealer       int  // 庄家位置
	IsDraw       bool // 是否流局
}

// NewMahjongGame 创建新游戏
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

// initializeDeck 初始化牌组
func (g *MahjongGame) initializeDeck() {
	g.Deck = make([]string, 0, 144)

	// 万子 1-9，每种4张
	for i := 1; i <= 9; i++ {
		for j := 0; j < 4; j++ {
			g.Deck = append(g.Deck, tileID("wan", i))
		}
	}

	// 筒子 1-9，每种4张
	for i := 1; i <= 9; i++ {
		for j := 0; j < 4; j++ {
			g.Deck = append(g.Deck, tileID("tong", i))
		}
	}

	// 条子 1-9，每种4张
	for i := 1; i <= 9; i++ {
		for j := 0; j < 4; j++ {
			g.Deck = append(g.Deck, tileID("tiao", i))
		}
	}

	// 风牌，每种4张
	winds := []string{"dong", "nan", "xi", "bei"}
	for _, wind := range winds {
		for j := 0; j < 4; j++ {
			g.Deck = append(g.Deck, wind)
		}
	}

	// 三元牌，每种4张
	dragons := []string{"zhong", "fa", "bai"}
	for _, dragon := range dragons {
		for j := 0; j < 4; j++ {
			g.Deck = append(g.Deck, dragon)
		}
	}

	// 花牌，每种1张
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

// DealTiles 发牌（台湾16张麻将）
func (g *MahjongGame) DealTiles() {
	// 台湾16张麻将：每人发16张牌
	for _, player := range g.Players {
		player.Hand = make([]string, 0, 17)

		for i := 0; i < 16; i++ {
			if len(g.Deck) > 0 {
				tile := g.DrawTile()

				// 如果是花牌，加入花牌集合并补牌
				if isFlowerTile(tile) {
					player.Flowers = append(player.Flowers, tile)
					i-- // 重新抽一张（不计入手牌数量）
					continue
				}

				player.Hand = append(player.Hand, tile)
			}
		}

		// 庄家多一张
		if player.Position == g.Dealer && len(g.Deck) > 0 {
			tile := g.DrawTile()
			// 检查是否为花牌
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

// DrawTileFromEnd 从牌山尾部摸一张牌（用于补杠）
func (g *MahjongGame) DrawTileFromEnd() string {
	if len(g.Deck) > 0 {
		tile := g.Deck[len(g.Deck)-1]
		g.Deck = g.Deck[:len(g.Deck)-1]
		return tile
	}
	return ""
}

// DrawTileWithFlowerReplacement 摸牌（自动处理花牌补牌）
func (g *MahjongGame) DrawTileWithFlowerReplacement(player *Player) string {
	// 检查是否流局
	if g.CheckDraw() {
		return ""
	}

	tile := g.DrawTile()

	// 如果摸到花牌，加入花牌集合并继续补牌
	for isFlowerTile(tile) && len(g.Deck) > 0 {
		player.Flowers = append(player.Flowers, tile)

		// 再次检查流局
		if g.CheckDraw() {
			return ""
		}

		tile = g.DrawTile()
	}

	return tile
}

// CheckDraw 检查是否流局（海底）
func (g *MahjongGame) CheckDraw() bool {
	// 台湾16张麻将：当牌山剩余8张或更少时流局
	if len(g.Deck) <= 8 {
		g.IsDraw = true
		return true
	}
	return false
}

// GetRemainingTiles 获取牌山剩余张数
func (g *MahjongGame) GetRemainingTiles() int {
	return len(g.Deck)
}

// NextTurn 下一回合
func (g *MahjongGame) NextTurn() {
	g.CurrentTurn = (g.CurrentTurn + 1) % len(g.Players)
}

// tileID 生成牌ID
func tileID(typ string, num int) string {
	if num == 0 {
		return typ
	}
	return typ + "-" + string(rune('0'+num))
}

// isFlowerTile 判断是否为花牌
func isFlowerTile(tile string) bool {
	return len(tile) > 6 && tile[:6] == "flower"
}

// CanPong 检查是否可以碰
func (g *MahjongGame) CanPong(hand []string, tile string) bool {
	count := 0
	for _, t := range hand {
		if t == tile {
			count++
		}
	}
	return count >= 2
}

// CanKong 检查是否可以杠
func (g *MahjongGame) CanKong(hand []string, tile string) bool {
	count := 0
	for _, t := range hand {
		if t == tile {
			count++
		}
	}
	return count >= 3
}

// CanExposedKong 检查是否可以明杠（包含加杠）
func (g *MahjongGame) CanExposedKong(player *Player, tile string) bool {
	// 检查手牌中是否有三张
	handCount := 0
	for _, t := range player.Hand {
		if t == tile {
			handCount++
		}
	}
	if handCount >= 3 {
		return true
	}

	// 检查是否已有碰牌，可以加杠
	for _, meld := range player.Melds {
		if meld.Type == "pong" && meld.Tiles[0] == tile {
			return true
		}
	}

	return false
}

// CanChow 检查是否可以吃牌（只能吃上家打出的牌）
// 返回所有可能的吃牌组合
func (g *MahjongGame) CanChow(hand []string, tile string) [][]string {
	// 字牌（风牌、三元牌）不能吃
	if isHonorTile(tile) {
		return nil
	}

	typ, num := parseTile(tile)
	if num == 0 {
		return nil
	}

	validCombinations := make([][]string, 0)

	// 情况1: tile + tile+1 + tile+2 (例如: 1,2,3)
	if num <= 7 {
		tile2 := tileID(typ, num+1)
		tile3 := tileID(typ, num+2)
		if countTile(hand, tile2) > 0 && countTile(hand, tile3) > 0 {
			validCombinations = append(validCombinations, []string{tile, tile2, tile3})
		}
	}

	// 情况2: tile-1 + tile + tile+1 (例如: 2,3,4)
	if num >= 2 && num <= 8 {
		tile1 := tileID(typ, num-1)
		tile3 := tileID(typ, num+1)
		if countTile(hand, tile1) > 0 && countTile(hand, tile3) > 0 {
			validCombinations = append(validCombinations, []string{tile1, tile, tile3})
		}
	}

	// 情况3: tile-2 + tile-1 + tile (例如: 3,4,5)
	if num >= 3 {
		tile1 := tileID(typ, num-2)
		tile2 := tileID(typ, num-1)
		if countTile(hand, tile1) > 0 && countTile(hand, tile2) > 0 {
			validCombinations = append(validCombinations, []string{tile1, tile2, tile})
		}
	}

	return validCombinations
}

// CanHu 检查是否可以胡牌
func (g *MahjongGame) CanHu(hand []string, melds []Meld) bool {
	// 台湾16张麻将：胡牌牌型 = 5组面子（顺子/刻子） + 1对眼 = 17张牌
	// 已展示的牌组（碰、杠）
	meldCount := len(melds)

	// 需要在手牌中找到剩余的组合
	// 总共需要5组面子 + 1对眼
	needGroups := 5 - meldCount

	// 手牌数量检查
	expectedHandSize := needGroups*3 + 2 // 剩余组数*3 + 1对眼
	if len(hand) != expectedHandSize {
		return false
	}

	// 复制手牌进行判断
	tiles := make([]string, len(hand))
	copy(tiles, hand)
	sortTiles(tiles)

	return canFormWinningHand(tiles, needGroups, true)
}

// canFormWinningHand 递归检查是否能组成胡牌手型
func canFormWinningHand(tiles []string, needGroups int, needPair bool) bool {
	if len(tiles) == 0 {
		return needGroups == 0 && !needPair
	}

	if needGroups == 0 && needPair {
		// 只需要找一对眼
		if len(tiles) != 2 {
			return false
		}
		return tiles[0] == tiles[1]
	}

	// 尝试移除刻子（3张相同）
	tile := tiles[0]
	if countTile(tiles, tile) >= 3 {
		newTiles := removeTiles(tiles, tile, 3)
		if canFormWinningHand(newTiles, needGroups-1, needPair) {
			return true
		}
	}

	// 尝试移除顺子（3张连续）
	if isSequencePossible(tiles, tile) {
		newTiles := removeSequence(tiles, tile)
		if len(newTiles) < len(tiles) { // 确保移除成功
			if canFormWinningHand(newTiles, needGroups-1, needPair) {
				return true
			}
		}
	}

	// 如果还需要对眼，尝试移除一对
	if needPair && countTile(tiles, tile) >= 2 {
		newTiles := removeTiles(tiles, tile, 2)
		if canFormWinningHand(newTiles, needGroups, false) {
			return true
		}
	}

	// 当前牌无法组成任何组合
	return false
}

// countTile 计算某张牌在手牌中的数量
func countTile(tiles []string, tile string) int {
	count := 0
	for _, t := range tiles {
		if t == tile {
			count++
		}
	}
	return count
}

// removeTiles 从手牌中移除指定数量的牌
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

// isSequencePossible 检查是否可能组成顺子
func isSequencePossible(tiles []string, tile string) bool {
	// 字牌（风牌、三元牌）不能组成顺子
	if isHonorTile(tile) {
		return false
	}

	typ, num := parseTile(tile)
	if num == 0 || num > 7 {
		return false // 8、9不能作为顺子开头
	}

	// 检查是否有连续的三张
	tile2 := tileID(typ, num+1)
	tile3 := tileID(typ, num+2)

	return countTile(tiles, tile) > 0 &&
	       countTile(tiles, tile2) > 0 &&
	       countTile(tiles, tile3) > 0
}

// removeSequence 移除一组顺子
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

// isHonorTile 判断是否为字牌（风牌、三元牌）
func isHonorTile(tile string) bool {
	honors := []string{"dong", "nan", "xi", "bei", "zhong", "fa", "bai"}
	for _, h := range honors {
		if tile == h {
			return true
		}
	}
	return false
}

// parseTile 解析牌的类型和数字
func parseTile(tile string) (string, int) {
	if len(tile) < 5 { // 字牌
		return tile, 0
	}

	// 格式: "wan-1", "tong-5"等
	typ := tile[:len(tile)-2]
	num := int(tile[len(tile)-1] - '0')
	return typ, num
}

// tileValue returns a sortable value for a tile.
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

// sortTiles sorts a slice of tiles.
func sortTiles(tiles []string) {
	sort.Slice(tiles, func(i, j int) bool {
		return tileValue(tiles[i]) < tileValue(tiles[j])
	})
}

// TingResult holds the result of a Ting check.
type TingResult struct {
	IsTing       bool
	WinningTiles []string
}

// getUniqueTileTypes returns a slice of all 34 unique tile types.
func (g *MahjongGame) getUniqueTileTypes() []string {
	return []string{
		"wan-1", "wan-2", "wan-3", "wan-4", "wan-5", "wan-6", "wan-7", "wan-8", "wan-9",
		"tong-1", "tong-2", "tong-3", "tong-4", "tong-5", "tong-6", "tong-7", "tong-8", "tong-9",
		"tiao-1", "tiao-2", "tiao-3", "tiao-4", "tiao-5", "tiao-6", "tiao-7", "tiao-8", "tiao-9",
		"dong", "nan", "xi", "bei",
		"zhong", "fa", "bai",
	}
}

// CheckTing determines if a hand is one tile away from winning (Ting).
func (g *MahjongGame) CheckTing(hand []string, melds []Meld) TingResult {
	// A hand must have a specific number of tiles to be in a Ting state.
	// For a 16-tile hand, it's 4 sets + 1 pair = 17 tiles.
	// A Ting hand has 16 tiles (or 13, 10, 7, 4), which is (N*3 + 1).
	// This check is a simplification and might not cover all cases, but is good for a start.
	if len(hand)%3 != 1 {
		return TingResult{IsTing: false}
	}

	winningTiles := make([]string, 0)
	checkedTiles := make(map[string]bool)

	allTileTypes := g.getUniqueTileTypes()

	for _, potentialTile := range allTileTypes {
		// Avoid re-checking the same tile
		if checkedTiles[potentialTile] {
			continue
		}

		tempHand := append([]string{}, hand...) // Create a copy
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


// ChooseDiscardAI 为Bot选择要打出的牌
func (g *MahjongGame) ChooseDiscardAI(hand []string) string {
	if len(hand) == 0 {
		return ""
	}

	tileCounts := make(map[string]int)
	for _, tile := range hand {
		tileCounts[tile]++
	}

	// 评估每张牌的价值，分数越低越优先打出
	// 简单的AI：孤张 > 边张 > 中张
	// 字牌孤张最优先
	var bestTileToDiscard string
	lowestScore := 1000

	// 检查孤立的字牌
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

	// 检查孤立的幺九牌 (1和9)
	for _, tile := range hand {
		typ, num := parseTile(tile)
		if num == 1 || num == 9 {
			if tileCounts[tile] == 1 {
				// 检查邻近牌
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

	// 检查其他孤立的牌
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

	// 如果没有孤张，则打出一张不成对的牌
	for _, tile := range hand {
		if tileCounts[tile] == 1 {
			return tile
		}
	}

	// 如果都是对子或刻子，打出最后一张
	return hand[len(hand)-1]
}
