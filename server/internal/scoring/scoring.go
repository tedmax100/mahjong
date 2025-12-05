package scoring

import (
	"log"
	"mahjong/internal/model"
	"mahjong/internal/tile"
)

// HandType 代表胡牌的牌型
type HandType struct {
	Name   string // 牌型名稱
	Tai    int    // 台數
	IsFaan bool   // 是否為番（特殊計分）
}

// WinResult 代表胡牌結果
type WinResult struct {
	HandTypes   []HandType   // 所有符合的牌型
	TotalTai    int          // 總台數
	BaseScore   int          // 基礎分數
	IsSelfDrawn bool         // 是否自摸
	WinningHand []string     // 胡牌時的手牌
	Melds       []model.Meld // 胡牌時的吃碰槓
	WinTile     string       // 胡的最後一張牌
}

// CalculateScore 計算台數和得分
func CalculateScore(hand []string, melds []model.Meld, flowers []string, lastTile string, isSelfDrawn bool) *WinResult {
	result := &WinResult{
		HandTypes:   make([]HandType, 0),
		IsSelfDrawn: isSelfDrawn,
		WinningHand: hand,
		Melds:       melds,
		WinTile:     lastTile,
	}

	// 基礎台數
	baseTai := 0

	// 1. 檢查自摸（1 台）
	if isSelfDrawn {
		result.HandTypes = append(result.HandTypes, HandType{
			Name: "自摸",
			Tai:  1,
		})
		baseTai += 1
	}

	// 2. 檢查門清（門前清，未碰、槓、吃）
	// 暗槓不算破壞門清
	isMenQing := true
	for _, meld := range melds {
		if meld.Type != "kong_concealed" {
			isMenQing = false
			break
		}
	}
	
	// 如果是吃胡（非自摸），門清要看是否有吃碰槓（暗槓除外），如果有則沒有門清
	// 但台灣麻將門清自摸是3台（門清1+自摸1+不求人1? 規則不一，這裡假設門清就是沒吃碰明槓）
	// 簡單判斷：len(melds) == 0 或者只有暗槓
	
	if isMenQing {
		result.HandTypes = append(result.HandTypes, HandType{
			Name: "門清",
			Tai:  1,
		})
		baseTai += 1
	}

	// 3. 檢查花牌（每朵 1 台）
	flowerTai := len(flowers)
	if flowerTai > 0 {
		result.HandTypes = append(result.HandTypes, HandType{
			Name: "花牌",
			Tai:  flowerTai,
		})
		baseTai += flowerTai
	}

	// 4. 檢查全求人（全部靠別人打的牌胡牌，0 台）
	// TODO: 實作全求人判斷

	// 5. 檢查平胡（基礎胡牌，1 台）
	if baseTai == 0 {
		result.HandTypes = append(result.HandTypes, HandType{
			Name: "平胡",
			Tai:  1,
		})
		baseTai = 1
	}

	// 6. 檢查特殊牌型
	specialTai := checkSpecialHandTypes(hand, melds, result)
	baseTai += specialTai

	// 7. 檢查番牌型（大三元、大四喜等）
	// TODO: 實作番牌型判斷

	result.TotalTai = baseTai

	// 計算基礎分數（底分 * 2^台數）
	// 台灣 16 張麻將常見規則：底分 10 元，每台翻倍
	baseAmount := 10
	// #nosec G115 -- baseTai 來自遊戲邏輯，範圍有限 (0-20)，不會溢出
	result.BaseScore = baseAmount << uint(baseTai) // 相當於 10 * (2 ^ baseTai)

	// log.Printf("胡牌: 台數=%d, 分數=%d", baseTai, result.BaseScore)
	// for _, ht := range result.HandTypes {
	// 	log.Printf("  - %s: %d 台", ht.Name, ht.Tai)
	// }

	return result
}

// checkSpecialHandTypes 檢查特殊牌型
func checkSpecialHandTypes(hand []string, melds []model.Meld, result *WinResult) int {
	totalTai := 0

	// 複製手牌+已展示的牌
	allTiles := make([]string, len(hand))
	copy(allTiles, hand)

	for _, meld := range melds {
		allTiles = append(allTiles, meld.Tiles...)
	}

	// 1. 檢查碰碰胡（全部都是刻子，3 台）
	if isPongPongHu(hand, melds) {
		result.HandTypes = append(result.HandTypes, HandType{
			Name: "碰碰胡",
			Tai:  3,
		})
		totalTai += 3
	}

	// 2. 檢查混一色（只有一種花色+字牌，3 台）
	if isMixedOneSuit(allTiles) {
		result.HandTypes = append(result.HandTypes, HandType{
			Name: "混一色",
			Tai:  3,
		})
		totalTai += 3
	}

	// 3. 檢查清一色（只有一種花色，5 台）
	if isOneSuit(allTiles) {
		result.HandTypes = append(result.HandTypes, HandType{
			Name: "清一色",
			Tai:  5,
		})
		totalTai += 5
	}

	// 4. 檢查字一色（全部是字牌，8 台）
	if isAllHonors(allTiles) {
		result.HandTypes = append(result.HandTypes, HandType{
			Name: "字一色",
			Tai:  8,
		})
		totalTai += 8
	}

	// 5. 檢查小三元（中發白有 2 組+1 對，2 台）
	if isSmallDragons(allTiles) {
		result.HandTypes = append(result.HandTypes, HandType{
			Name: "小三元",
			Tai:  2,
		})
		totalTai += 2
	}

	// 6. 檢查大三元（中發白全有，8 台）
	if isBigDragons(allTiles) {
		result.HandTypes = append(result.HandTypes, HandType{
			Name: "大三元",
			Tai:  8,
		})
		totalTai += 8
	}

	// 7. 檢查小四喜（東南西北有 3 組+1 對，2 台）
	if isSmallWinds(allTiles) {
		result.HandTypes = append(result.HandTypes, HandType{
			Name: "小四喜",
			Tai:  2,
		})
		totalTai += 2
	}

	// 8. 檢查大四喜（東南西北全有，8 台）
	if isBigWinds(allTiles) {
		result.HandTypes = append(result.HandTypes, HandType{
			Name: "大四喜",
			Tai:  8,
		})
		totalTai += 8
	}

	return totalTai
}

// isPongPongHu 檢查是否為碰碰胡
func isPongPongHu(hand []string, melds []model.Meld) bool {
	// 所有的牌組都必須是刻子（3 張或 4 張相同）
	// 手牌必須能組成刻子+一對眼

	// 檢查已展示的牌組
	for _, meld := range melds {
		if meld.Type == "chow" { // 如果有吃牌（順子），不是碰碰胡
			return false
		}
	}

	// 檢查手牌是否都能組成刻子
	tiles := make([]string, len(hand))
	copy(tiles, hand)

	return canFormPongPongHu(tiles)
}

// canFormPongPongHu 檢查手牌是否能組成碰碰胡
func canFormPongPongHu(tiles []string) bool {
	if len(tiles) == 0 {
		return true
	}

	if len(tiles) == 2 {
		// 最後的對眼
		return tiles[0] == tiles[1]
	}

	t := tiles[0]
	count := tile.Count(tiles, t)

	// 嘗試組成刻子（3 張）
	if count >= 3 {
		newTiles := tile.Remove(tiles, t, 3)
		if canFormPongPongHu(newTiles) {
			return true
		}
	}

	// 嘗試組成對眼（2 張）
	if count >= 2 && len(tiles) > 2 {
		newTiles := tile.Remove(tiles, t, 2)
		// 剩餘的必須都是刻子
		return canFormAllTriples(newTiles)
	}

	return false
}

// canFormAllTriples 檢查是否都能組成刻子
func canFormAllTriples(tiles []string) bool {
	if len(tiles) == 0 {
		return true
	}

	if len(tiles)%3 != 0 {
		return false
	}

	t := tiles[0]
	count := tile.Count(tiles, t)

	if count < 3 {
		return false
	}

	newTiles := tile.Remove(tiles, t, 3)
	return canFormAllTriples(newTiles)
}

// isMixedOneSuit 檢查是否為混一色
func isMixedOneSuit(tiles []string) bool {
	var mainSuit string
	hasHonor := false

	for _, t := range tiles {
		if tile.IsHonor(t) {
			hasHonor = true
			continue
		}

		suit, _ := tile.Parse(t)
		if mainSuit == "" {
			mainSuit = suit
		} else if suit != mainSuit {
			return false // 有超過一種花色
		}
	}

	return mainSuit != "" && hasHonor
}

// isOneSuit 檢查是否為清一色
func isOneSuit(tiles []string) bool {
	var mainSuit string

	for _, t := range tiles {
		if tile.IsHonor(t) {
			return false // 有字牌不是清一色
		}

		suit, _ := tile.Parse(t)
		if mainSuit == "" {
			mainSuit = suit
		} else if suit != mainSuit {
			return false
		}
	}

	return mainSuit != ""
}

// isAllHonors 檢查是否為字一色
func isAllHonors(tiles []string) bool {
	for _, t := range tiles {
		if !tile.IsHonor(t) {
			return false
		}
	}
	return true
}

// isSmallDragons 檢查是否為小三元
func isSmallDragons(tiles []string) bool {
	dragons := []string{"zhong", "fa", "bai"}
	dragonCount := make(map[string]int)

	for _, t := range tiles {
		for _, dragon := range dragons {
			if t == dragon {
				dragonCount[dragon]++
			}
		}
	}

	// 小三元：2 種三元牌各有 3 張或 4 張，1 種有 2 張
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

// isBigDragons 檢查是否為大三元
func isBigDragons(tiles []string) bool {
	dragons := []string{"zhong", "fa", "bai"}
	dragonCount := make(map[string]int)

	for _, t := range tiles {
		for _, dragon := range dragons {
			if t == dragon {
				dragonCount[dragon]++
			}
		}
	}

	// 大三元：3 種三元牌各有 3 張或 4 張
	for _, dragon := range dragons {
		if dragonCount[dragon] < 3 {
			return false
		}
	}

	return true
}

// isSmallWinds 檢查是否為小四喜
func isSmallWinds(tiles []string) bool {
	winds := []string{"dong", "nan", "xi", "bei"}
	windCount := make(map[string]int)

	for _, t := range tiles {
		for _, wind := range winds {
			if t == wind {
				windCount[wind]++
			}
		}
	}

	// 小四喜：3 種風牌各有 3 張或 4 張，1 種有 2 張
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

// isBigWinds 檢查是否為大四喜
func isBigWinds(tiles []string) bool {
	winds := []string{"dong", "nan", "xi", "bei"}
	windCount := make(map[string]int)

	for _, t := range tiles {
		for _, wind := range winds {
			if t == wind {
				windCount[wind]++
			}
		}
	}

	// 大四喜：4 種風牌各有 3 張或 4 張
	for _, wind := range winds {
		if windCount[wind] < 3 {
			return false
		}
	}

	return true
}
