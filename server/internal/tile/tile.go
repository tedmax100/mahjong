package tile

import (
	"sort"
	"strconv"
	"strings"
)

// ID 生成牌 ID
func ID(typ string, num int) string {
	if num == 0 {
		return typ
	}
	return typ + "-" + string(rune('0'+num))
}

// IsFlower 判斷是否為花牌
func IsFlower(tile string) bool {
	return len(tile) > 6 && tile[:6] == "flower"
}

// IsHonor 判斷是否為字牌（風牌、三元牌）
func IsHonor(tile string) bool {
	honors := []string{"dong", "nan", "xi", "bei", "zhong", "fa", "bai"}
	for _, h := range honors {
		if tile == h {
			return true
		}
	}
	return false
}

// Parse 解析牌的類型和數字
func Parse(tile string) (string, int) {
	// 如果不包含 "-", 則視為字牌
	if !strings.Contains(tile, "-") {
		return tile, 0
	}

	// 格式: "wan-1", "tong-5" 等
	// 倒數第二個字符應該是 "-"
	if len(tile) < 2 || tile[len(tile)-2] != '-' {
		// Fallback or error handling, though for now we assume valid format if it contains '-'
		// But to be safe against things like "flower-chun"
		if IsFlower(tile) {
			return tile, 0
		}
		return tile, 0
	}

	typ := tile[:len(tile)-2]
	num := int(tile[len(tile)-1] - '0')
	return typ, num
}

// Value 返回牌的可排序值
func Value(tile string) int {
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

// Sort 對牌的切片進行排序
func Sort(tiles []string) {
	sort.Slice(tiles, func(i, j int) bool {
		return Value(tiles[i]) < Value(tiles[j])
	})
}

// GetUniqueTypes 返回所有 34 種唯一牌型的切片
func GetUniqueTypes() []string {
	return []string{
		"wan-1", "wan-2", "wan-3", "wan-4", "wan-5", "wan-6", "wan-7", "wan-8", "wan-9",
		"tong-1", "tong-2", "tong-3", "tong-4", "tong-5", "tong-6", "tong-7", "tong-8", "tong-9",
		"tiao-1", "tiao-2", "tiao-3", "tiao-4", "tiao-5", "tiao-6", "tiao-7", "tiao-8", "tiao-9",
		"dong", "nan", "xi", "bei",
		"zhong", "fa", "bai",
	}
}

// Count 計算某張牌在牌組中的數量
func Count(tiles []string, t string) int {
	count := 0
	for _, h := range tiles {
		if h == t {
			count++
		}
	}
	return count
}

// Remove 從牌組中移除指定數量的牌，返回新的牌組
func Remove(tiles []string, t string, count int) []string {
	result := make([]string, 0, len(tiles))
	removed := 0
	for _, h := range tiles {
		if h == t && removed < count {
			removed++
			continue
		}
		result = append(result, h)
	}
	return result
}
