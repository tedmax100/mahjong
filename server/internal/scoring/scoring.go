package scoring

import (
	"mahjong/internal/model"
	"mahjong/internal/tile"
)

// Wind 代表風牌方向
type Wind string

const (
	WindEast  Wind = "E" // 東
	WindSouth Wind = "S" // 南
	WindWest  Wind = "W" // 西
	WindNorth Wind = "N" // 北
)

// SpecialCondition 代表特殊胡牌情境
type SpecialCondition string

const (
	ConditionLastTile       SpecialCondition = "LAST_TILE"        // 海底撈月
	ConditionKongBloom      SpecialCondition = "KONG_BLOOM"       // 槓上開花
	ConditionRobbingKong    SpecialCondition = "ROBBING_KONG"     // 搶槓
	ConditionSelfDrawnAfter SpecialCondition = "SELF_DRAWN_AFTER" // 妙手回春（自摸最後一張）
)

// ScoreInput 計分輸入結構
type ScoreInput struct {
	RoundWind         Wind               // 場風
	SeatWind          Wind               // 胡牌玩家的門風（自風）
	IsDealer          bool               // 是否為莊家
	IsSelfDrawn       bool               // 胡牌方式：自摸 / 榮和（吃胡）
	Hand              []string           // 胡牌玩家手牌
	Melds             []model.Meld       // 吃/碰/槓等結構化資料
	Flowers           []string           // 花牌
	WinningTile       string             // 胡的那張牌
	SpecialConditions []SpecialCondition // 特殊情境標記
}

// HandType 代表胡牌的牌型
type HandType struct {
	Name   string // 牌型名稱
	Tai    int    // 台數
	IsFaan bool   // 是否為番（特殊計分）
}

// ScoreBreakdown 計分明細
type ScoreBreakdown struct {
	Key      string // 內部用代號，例如 "MEN_QING_ZIMO", "ROUND_WIND_KOUTSU"
	Label    string // 給 UI 顯示的文字，例如「門清自摸」「場風刻」
	Value    int    // 該項加成的數值（台數）
	Category string // 分類，例如 "yaku", "bonus", "wind", "kong"
}

// WinResult 代表胡牌結果
type WinResult struct {
	HandTypes   []HandType       // 所有符合的牌型（向後兼容）
	Breakdown   []ScoreBreakdown // 詳細計分明細
	TotalTai    int              // 總台數
	BaseScore   int              // 基礎分數
	IsSelfDrawn bool             // 是否自摸
	WinningHand []string         // 胡牌時的手牌
	Melds       []model.Meld     // 胡牌時的吃碰槓
	WinTile     string           // 胡的最後一張牌
	RoundWind   Wind             // 場風
	SeatWind    Wind             // 門風
	IsDealer    bool             // 是否為莊家
}

// addScore 添加計分項目
func (r *WinResult) addScore(key, label string, value int, category string) {
	// 添加到 HandTypes（向後兼容）
	r.HandTypes = append(r.HandTypes, HandType{
		Name: label,
		Tai:  value,
	})

	// 添加到 Breakdown（新格式）
	r.Breakdown = append(r.Breakdown, ScoreBreakdown{
		Key:      key,
		Label:    label,
		Value:    value,
		Category: category,
	})
}

// windToTile 將 Wind 轉換為牌面字串
func windToTile(w Wind) string {
	switch w {
	case WindEast:
		return "dong"
	case WindSouth:
		return "nan"
	case WindWest:
		return "xi"
	case WindNorth:
		return "bei"
	default:
		return ""
	}
}

// countTileInHand 計算指定牌在手牌中的數量
func countTileInHand(hand []string, t string) int {
	count := 0
	for _, h := range hand {
		if h == t {
			count++
		}
	}
	return count
}

// hasTripletOrKong 檢查是否有指定牌的刻子或槓子（包含手牌和已亮牌組）
func hasTripletOrKong(hand []string, melds []model.Meld, t string) bool {
	// 檢查已亮牌組（碰、槓）
	for _, meld := range melds {
		if len(meld.Tiles) > 0 && meld.Tiles[0] == t {
			if meld.Type == "pong" || meld.Type == "kong_concealed" ||
				meld.Type == "kong_exposed" || meld.Type == "kong_promoted" {
				return true
			}
		}
	}

	// 檢查手牌中的暗刻（3 張或以上相同）
	if countTileInHand(hand, t) >= 3 {
		return true
	}

	return false
}

// checkWindTriplets 檢查場風刻和門風刻
func checkWindTriplets(hand []string, melds []model.Meld, roundWind, seatWind Wind, result *WinResult) int {
	totalTai := 0

	roundWindTile := windToTile(roundWind)
	seatWindTile := windToTile(seatWind)

	// 檢查場風刻
	if roundWindTile != "" && hasTripletOrKong(hand, melds, roundWindTile) {
		result.addScore("ROUND_WIND_TRIPLET", "場風刻", 1, "wind")
		totalTai += 1
	}

	// 檢查門風刻
	if seatWindTile != "" && hasTripletOrKong(hand, melds, seatWindTile) {
		result.addScore("SEAT_WIND_TRIPLET", "門風刻", 1, "wind")
		totalTai += 1
	}

	return totalTai
}

// checkKongScoring 檢查槓牌台數
func checkKongScoring(melds []model.Meld, result *WinResult) int {
	totalTai := 0

	for _, meld := range melds {
		switch meld.Type {
		case "kong_concealed":
			// 暗槓：2 台
			result.addScore("KONG_CONCEALED", "暗槓", 2, "kong")
			totalTai += 2
		case "kong_exposed":
			// 明槓：1 台
			result.addScore("KONG_EXPOSED", "明槓", 1, "kong")
			totalTai += 1
		case "kong_promoted":
			// 加槓（補槓）：1 台
			result.addScore("KONG_PROMOTED", "加槓", 1, "kong")
			totalTai += 1
		}
	}

	return totalTai
}

// checkSpecialConditions 檢查特殊胡牌情境
func checkSpecialConditions(conditions []SpecialCondition, result *WinResult) int {
	totalTai := 0

	for _, condition := range conditions {
		switch condition {
		case ConditionLastTile:
			// 海底撈月：1 台
			result.addScore("LAST_TILE", "海底撈月", 1, "special")
			totalTai += 1
		case ConditionKongBloom:
			// 槓上開花：1 台
			result.addScore("KONG_BLOOM", "槓上開花", 1, "special")
			totalTai += 1
		case ConditionRobbingKong:
			// 搶槓：1 台
			result.addScore("ROBBING_KONG", "搶槓", 1, "special")
			totalTai += 1
		case ConditionSelfDrawnAfter:
			// 妙手回春：1 台（與海底撈月類似，但通常用於自摸最後一張）
			result.addScore("SELF_DRAWN_AFTER", "妙手回春", 1, "special")
			totalTai += 1
		}
	}

	return totalTai
}

// CalculateScoreWithInput 使用 ScoreInput 計算台數和得分（新版本）
func CalculateScoreWithInput(input *ScoreInput) *WinResult {
	result := &WinResult{
		HandTypes:   make([]HandType, 0),
		Breakdown:   make([]ScoreBreakdown, 0),
		IsSelfDrawn: input.IsSelfDrawn,
		WinningHand: input.Hand,
		Melds:       input.Melds,
		WinTile:     input.WinningTile,
		RoundWind:   input.RoundWind,
		SeatWind:    input.SeatWind,
		IsDealer:    input.IsDealer,
	}

	baseTai := 0

	// 1. 檢查自摸（1 台）
	if input.IsSelfDrawn {
		result.addScore("SELF_DRAWN", "自摸", 1, "basic")
		baseTai += 1
	}

	// 2. 檢查門清（門前清，未碰、槓、吃，暗槓除外）
	isMenQing := true
	for _, meld := range input.Melds {
		if meld.Type != "kong_concealed" {
			isMenQing = false
			break
		}
	}

	if isMenQing {
		result.addScore("MEN_QING", "門清", 1, "basic")
		baseTai += 1
	}

	// 3. 檢查花牌（每朵 1 台）
	flowerTai := len(input.Flowers)
	if flowerTai > 0 {
		result.addScore("FLOWERS", "花牌", flowerTai, "flower")
		baseTai += flowerTai
	}

	// 4. 檢查場風刻 / 門風刻
	windTai := checkWindTriplets(input.Hand, input.Melds, input.RoundWind, input.SeatWind, result)
	baseTai += windTai

	// 5. 檢查槓牌台數
	kongTai := checkKongScoring(input.Melds, result)
	baseTai += kongTai

	// 6. 檢查特殊情境
	specialTai := checkSpecialConditions(input.SpecialConditions, result)
	baseTai += specialTai

	// 7. 檢查平胡（基礎胡牌，1 台）- 只在沒有任何台數時才給
	if baseTai == 0 {
		result.addScore("PING_HU", "平胡", 1, "basic")
		baseTai = 1
	}

	// 8. 檢查特殊牌型
	patternTai := checkSpecialHandTypes(input.Hand, input.Melds, result)
	baseTai += patternTai

	result.TotalTai = baseTai

	// 計算基礎分數（底分 * 2^台數）
	baseAmount := 10
	// #nosec G115 -- baseTai 來自遊戲邏輯，範圍有限 (0-20)，不會溢出
	result.BaseScore = baseAmount << uint(baseTai)

	return result
}

// CalculateScore 計算台數和得分（向後兼容版本）
func CalculateScore(hand []string, melds []model.Meld, flowers []string, lastTile string, isSelfDrawn bool) *WinResult {
	// 使用新版本計算，預設東風東家
	input := &ScoreInput{
		RoundWind:         WindEast,
		SeatWind:          WindEast,
		IsDealer:          false,
		IsSelfDrawn:       isSelfDrawn,
		Hand:              hand,
		Melds:             melds,
		Flowers:           flowers,
		WinningTile:       lastTile,
		SpecialConditions: nil,
	}
	return CalculateScoreWithInput(input)
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

	// 1. 檢查碰碰胡（全部都是刻子，4 台）
	if isPongPongHu(hand, melds) {
		result.addScore("PONG_PONG_HU", "碰碰胡", 4, "pattern")
		totalTai += 4
	}

	// 2. 檢查混一色（只有一種花色+字牌，4 台）
	if isMixedOneSuit(allTiles) {
		result.addScore("MIXED_ONE_SUIT", "混一色", 4, "pattern")
		totalTai += 4
	}

	// 3. 檢查清一色（只有一種花色，8 台）
	if isOneSuit(allTiles) {
		result.addScore("ONE_SUIT", "清一色", 8, "pattern")
		totalTai += 8
	}

	// 4. 檢查字一色（全部是字牌，8 台）
	if isAllHonors(allTiles) {
		result.addScore("ALL_HONORS", "字一色", 8, "pattern")
		totalTai += 8
	}

	// 5. 檢查小三元（中發白有 2 組+1 對，4 台）
	if isSmallDragons(allTiles) {
		result.addScore("SMALL_DRAGONS", "小三元", 4, "pattern")
		totalTai += 4
	}

	// 6. 檢查大三元（中發白全有，8 台）
	if isBigDragons(allTiles) {
		result.addScore("BIG_DRAGONS", "大三元", 8, "pattern")
		totalTai += 8
	}

	// 7. 檢查小四喜（東南西北有 3 組+1 對，8 台）
	if isSmallWinds(allTiles) {
		result.addScore("SMALL_WINDS", "小四喜", 8, "pattern")
		totalTai += 8
	}

	// 8. 檢查大四喜（東南西北全有，16 台）
	if isBigWinds(allTiles) {
		result.addScore("BIG_WINDS", "大四喜", 16, "pattern")
		totalTai += 16
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
	return len(tiles) > 0
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
