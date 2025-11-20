package game

import (
	"math/rand"
	"time"
)

// MahjongGame 代表一局麻将游戏
type MahjongGame struct {
	Players      []*Player
	Deck         []string
	CurrentTurn  int
	DiscardPile  []string
	Dealer       int // 庄家位置
}

// NewMahjongGame 创建新游戏
func NewMahjongGame(players []*Player) *MahjongGame {
	game := &MahjongGame{
		Players:      players,
		CurrentTurn:  0,
		DiscardPile:  make([]string, 0),
		Dealer:       0,
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
				tile := g.Deck[0]
				g.Deck = g.Deck[1:]

				// 如果是花牌，立即补牌
				if isFlowerTile(tile) {
					// TODO: 处理花牌
					i-- // 重新抽一张
					continue
				}

				player.Hand = append(player.Hand, tile)
			}
		}

		// 庄家多一张
		if player.Position == g.Dealer && len(g.Deck) > 0 {
			player.Hand = append(player.Hand, g.Deck[0])
			g.Deck = g.Deck[1:]
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

// CanHu 检查是否可以胡牌（简化版）
func (g *MahjongGame) CanHu(hand []string) bool {
	// TODO: 实现完整的台湾16张麻将胡牌判断
	// 这里先用简化版本
	return len(hand) == 17 || len(hand) == 18
}
