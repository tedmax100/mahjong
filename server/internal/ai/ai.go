package ai

import (
	"mahjong/internal/tile"
)

// ChooseDiscard 為 Bot 選擇要打出的牌
func ChooseDiscard(hand []string) string {
	if len(hand) == 0 {
		return ""
	}

	tileCounts := make(map[string]int)
	for _, t := range hand {
		tileCounts[t]++
	}

	// 評估每張牌的價值，分數越低越優先打出
	// 簡單的 AI：孤張 > 邊張 > 中張
	// 字牌孤張最優先
	var bestTileToDiscard string
	lowestScore := 1000

	// 檢查孤立的字牌
	for _, t := range hand {
		if tile.IsHonor(t) && tileCounts[t] == 1 {
			if 10 < lowestScore {
				lowestScore = 10
				bestTileToDiscard = t
			}
		}
	}
	if bestTileToDiscard != "" {
		return bestTileToDiscard
	}

	// 檢查孤立的么九牌 (1 和 9)
	for _, t := range hand {
		typ, num := tile.Parse(t)
		if num == 1 || num == 9 {
			if tileCounts[t] == 1 {
				// 檢查鄰近牌
				hasNeighbor := false
				if num == 1 && (tileCounts[tile.ID(typ, 2)] > 0) {
					hasNeighbor = true
				}
				if num == 9 && (tileCounts[tile.ID(typ, 8)] > 0) {
					hasNeighbor = true
				}

				if !hasNeighbor && 20 < lowestScore {
					lowestScore = 20
					bestTileToDiscard = t
				}
			}
		}
	}
	if bestTileToDiscard != "" {
		return bestTileToDiscard
	}

	// 檢查其他孤立的牌
	for _, t := range hand {
		typ, num := tile.Parse(t)
		if num > 1 && num < 9 {
			if tileCounts[t] == 1 {
				hasNeighbor := (tileCounts[tile.ID(typ, num-1)] > 0) || (tileCounts[tile.ID(typ, num+1)] > 0)
				if !hasNeighbor && 30 < lowestScore {
					lowestScore = 30
					bestTileToDiscard = t
				}
			}
		}
	}
	if bestTileToDiscard != "" {
		return bestTileToDiscard
	}

	// 如果沒有孤張，則打出一張不成對的牌
	for _, t := range hand {
		if tileCounts[t] == 1 {
			return t
		}
	}

	// 如果都是對子或刻子，打出最後一張
	return hand[len(hand)-1]
}
