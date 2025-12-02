package game

import "log"

// HandType 代表胡牌的牌型
type HandType struct {
	Name   string // 牌型名称
	Tai    int    // 台数
	IsFaan bool   // 是否为番（特殊计分）
}

// WinResult 代表胡牌结果
type WinResult struct {
	HandTypes   []HandType // 所有符合的牌型
	TotalTai    int        // 总台数
	BaseScore   int        // 基础分数
	IsSelfDrawn bool       // 是否自摸
	WinningHand []string   // 胡牌时的手牌
	Melds       []Meld     // 胡牌时的吃碰杠
	WinTile     string     // 胡的最后一张牌
}

// CalculateScore 计算台数和得分
func (g *MahjongGame) CalculateScore(player *Player, lastTile string, isSelfDrawn bool) *WinResult {
	result := &WinResult{
		HandTypes:   make([]HandType, 0),
		IsSelfDrawn: isSelfDrawn,
		WinningHand: player.Hand,
		Melds:       player.Melds,
		WinTile:     lastTile,
	}

	// 基础台数
	baseTai := 0

	// 1. 检查自摸（1台）
	if isSelfDrawn {
		result.HandTypes = append(result.HandTypes, HandType{
			Name: "自摸",
			Tai:  1,
		})
		baseTai += 1
	}

	// 2. 检查门清（门前清，未碰、杠、吃）
	if len(player.Melds) == 0 {
		result.HandTypes = append(result.HandTypes, HandType{
			Name: "门清",
			Tai:  1,
		})
		baseTai += 1
	}

	// 3. 检查花牌（每朵1台）
	flowerTai := len(player.Flowers)
	if flowerTai > 0 {
		result.HandTypes = append(result.HandTypes, HandType{
			Name: "花牌",
			Tai:  flowerTai,
		})
		baseTai += flowerTai
	}

	// 4. 检查全求人（全部靠别人打的牌胡牌，0台）
	// TODO: 实现全求人判断

	// 5. 检查平胡（基础胡牌，1台）
	if baseTai == 0 {
		result.HandTypes = append(result.HandTypes, HandType{
			Name: "平胡",
			Tai:  1,
		})
		baseTai = 1
	}

	// 6. 检查特殊牌型
	specialTai := g.checkSpecialHandTypes(player, result)
	baseTai += specialTai

	// 7. 检查番牌型（大三元、大四喜等）
	// TODO: 实现番牌型判断

	result.TotalTai = baseTai

	// 计算基础分数（底分 * 2^台数）
	// 台湾16张麻将常见规则：底分10元，每台翻倍
	baseAmount := 10
	result.BaseScore = baseAmount << uint(baseTai) // 相当于 10 * (2 ^ baseTai)

	log.Printf("玩家 %s 胡牌: 台数=%d, 分数=%d", player.Name, baseTai, result.BaseScore)
	for _, ht := range result.HandTypes {
		log.Printf("  - %s: %d台", ht.Name, ht.Tai)
	}

	return result
}

// checkSpecialHandTypes 检查特殊牌型
func (g *MahjongGame) checkSpecialHandTypes(player *Player, result *WinResult) int {
	totalTai := 0

	// 复制手牌+已展示的牌
	allTiles := make([]string, len(player.Hand))
	copy(allTiles, player.Hand)

	for _, meld := range player.Melds {
		allTiles = append(allTiles, meld.Tiles...)
	}

	// 1. 检查碰碰胡（全部都是刻子，3台）
	if g.isPongPongHu(player) {
		result.HandTypes = append(result.HandTypes, HandType{
			Name: "碰碰胡",
			Tai:  3,
		})
		totalTai += 3
	}

	// 2. 检查混一色（只有一种花色+字牌，3台）
	if g.isMixedOneSuit(allTiles) {
		result.HandTypes = append(result.HandTypes, HandType{
			Name: "混一色",
			Tai:  3,
		})
		totalTai += 3
	}

	// 3. 检查清一色（只有一种花色，5台）
	if g.isOneSuit(allTiles) {
		result.HandTypes = append(result.HandTypes, HandType{
			Name: "清一色",
			Tai:  5,
		})
		totalTai += 5
	}

	// 4. 检查字一色（全部是字牌，8台）
	if g.isAllHonors(allTiles) {
		result.HandTypes = append(result.HandTypes, HandType{
			Name: "字一色",
			Tai:  8,
		})
		totalTai += 8
	}

	// 5. 检查小三元（中发白有2组+1对，2台）
	if g.isSmallDragons(allTiles) {
		result.HandTypes = append(result.HandTypes, HandType{
			Name: "小三元",
			Tai:  2,
		})
		totalTai += 2
	}

	// 6. 检查大三元（中发白全有，8台）
	if g.isBigDragons(allTiles) {
		result.HandTypes = append(result.HandTypes, HandType{
			Name: "大三元",
			Tai:  8,
		})
		totalTai += 8
	}

	// 7. 检查小四喜（东南西北有3组+1对，2台）
	if g.isSmallWinds(allTiles) {
		result.HandTypes = append(result.HandTypes, HandType{
			Name: "小四喜",
			Tai:  2,
		})
		totalTai += 2
	}

	// 8. 检查大四喜（东南西北全有，8台）
	if g.isBigWinds(allTiles) {
		result.HandTypes = append(result.HandTypes, HandType{
			Name: "大四喜",
			Tai:  8,
		})
		totalTai += 8
	}

	return totalTai
}

// isPongPongHu 检查是否为碰碰胡
func (g *MahjongGame) isPongPongHu(player *Player) bool {
	// 所有的牌组都必须是刻子（3张或4张相同）
	// 手牌必须能组成刻子+一对眼

	// 检查已展示的牌组
	for _, meld := range player.Melds {
		if meld.Type == "chow" { // 如果有吃牌（顺子），不是碰碰胡
			return false
		}
	}

	// 检查手牌是否都能组成刻子
	tiles := make([]string, len(player.Hand))
	copy(tiles, player.Hand)

	return canFormPongPongHu(tiles)
}

// canFormPongPongHu 检查手牌是否能组成碰碰胡
func canFormPongPongHu(tiles []string) bool {
	if len(tiles) == 0 {
		return true
	}

	if len(tiles) == 2 {
		// 最后的对眼
		return tiles[0] == tiles[1]
	}

	tile := tiles[0]
	count := countTile(tiles, tile)

	// 尝试组成刻子（3张）
	if count >= 3 {
		newTiles := removeTiles(tiles, tile, 3)
		if canFormPongPongHu(newTiles) {
			return true
		}
	}

	// 尝试组成对眼（2张）
	if count >= 2 && len(tiles) > 2 {
		newTiles := removeTiles(tiles, tile, 2)
		// 剩余的必须都是刻子
		return canFormAllTriples(newTiles)
	}

	return false
}

// canFormAllTriples 检查是否都能组成刻子
func canFormAllTriples(tiles []string) bool {
	if len(tiles) == 0 {
		return true
	}

	if len(tiles)%3 != 0 {
		return false
	}

	tile := tiles[0]
	count := countTile(tiles, tile)

	if count < 3 {
		return false
	}

	newTiles := removeTiles(tiles, tile, 3)
	return canFormAllTriples(newTiles)
}

// isMixedOneSuit 检查是否为混一色
func (g *MahjongGame) isMixedOneSuit(tiles []string) bool {
	var mainSuit string
	hasHonor := false

	for _, tile := range tiles {
		if isHonorTile(tile) {
			hasHonor = true
			continue
		}

		suit, _ := parseTile(tile)
		if mainSuit == "" {
			mainSuit = suit
		} else if suit != mainSuit {
			return false // 有超过一种花色
		}
	}

	return mainSuit != "" && hasHonor
}

// isOneSuit 检查是否为清一色
func (g *MahjongGame) isOneSuit(tiles []string) bool {
	var mainSuit string

	for _, tile := range tiles {
		if isHonorTile(tile) {
			return false // 有字牌不是清一色
		}

		suit, _ := parseTile(tile)
		if mainSuit == "" {
			mainSuit = suit
		} else if suit != mainSuit {
			return false
		}
	}

	return mainSuit != ""
}

// isAllHonors 检查是否为字一色
func (g *MahjongGame) isAllHonors(tiles []string) bool {
	for _, tile := range tiles {
		if !isHonorTile(tile) {
			return false
		}
	}
	return true
}

// isSmallDragons 检查是否为小三元
func (g *MahjongGame) isSmallDragons(tiles []string) bool {
	dragons := []string{"zhong", "fa", "bai"}
	dragonCount := make(map[string]int)

	for _, tile := range tiles {
		for _, dragon := range dragons {
			if tile == dragon {
				dragonCount[dragon]++
			}
		}
	}

	// 小三元：2种三元牌各有3张或4张，1种有2张
	tripletCount := 0
	pairCount := 0

	for _, count := range dragonCount {
		if count >= 3 {
			tripletCount++
		} else if count == 2 {
			pairCount++
		}
	}

	return tripletCount == 2 && pairCount == 1
}

// isBigDragons 检查是否为大三元
func (g *MahjongGame) isBigDragons(tiles []string) bool {
	dragons := []string{"zhong", "fa", "bai"}
	dragonCount := make(map[string]int)

	for _, tile := range tiles {
		for _, dragon := range dragons {
			if tile == dragon {
				dragonCount[dragon]++
			}
		}
	}

	// 大三元：3种三元牌各有3张或4张
	for _, dragon := range dragons {
		if dragonCount[dragon] < 3 {
			return false
		}
	}

	return true
}

// isSmallWinds 检查是否为小四喜
func (g *MahjongGame) isSmallWinds(tiles []string) bool {
	winds := []string{"dong", "nan", "xi", "bei"}
	windCount := make(map[string]int)

	for _, tile := range tiles {
		for _, wind := range winds {
			if tile == wind {
				windCount[wind]++
			}
		}
	}

	// 小四喜：3种风牌各有3张或4张，1种有2张
	tripletCount := 0
	pairCount := 0

	for _, count := range windCount {
		if count >= 3 {
			tripletCount++
		} else if count == 2 {
			pairCount++
		}
	}

	return tripletCount == 3 && pairCount == 1
}

// isBigWinds 检查是否为大四喜
func (g *MahjongGame) isBigWinds(tiles []string) bool {
	winds := []string{"dong", "nan", "xi", "bei"}
	windCount := make(map[string]int)

	for _, tile := range tiles {
		for _, wind := range winds {
			if tile == wind {
				windCount[wind]++
			}
		}
	}

	// 大四喜：4种风牌各有3张或4张
	for _, wind := range winds {
		if windCount[wind] < 3 {
			return false
		}
	}

	return true
}
