package tile

import (
	"sort"
	"testing"
)

func TestID(t *testing.T) {
	tests := []struct {
		typ      string
		num      int
		expected string
	}{
		{"wan", 1, "wan-1"},
		{"tong", 9, "tong-9"},
		{"tiao", 5, "tiao-5"},
		{"dong", 0, "dong"},
		{"zhong", 0, "zhong"},
	}

	for _, tt := range tests {
		result := ID(tt.typ, tt.num)
		if result != tt.expected {
			t.Errorf("ID(%s, %d) = %s; expected %s", tt.typ, tt.num, result, tt.expected)
		}
	}
}

func TestIsFlower(t *testing.T) {
	tests := []struct {
		tile     string
		expected bool
	}{
		{"flower-chun", true},
		{"flower-xia", true},
		{"flower-qiu", true},
		{"flower-dong", true},
		{"flower-mei", true},
		{"flower-lan", true},
		{"flower-zhu", true},
		{"flower-ju", true},
		{"dong", false},
		{"wan-1", false},
		{"tong-5", false},
		{"zhong", false},
	}

	for _, tt := range tests {
		result := IsFlower(tt.tile)
		if result != tt.expected {
			t.Errorf("IsFlower(%s) = %v; expected %v", tt.tile, result, tt.expected)
		}
	}
}

func TestIsHonor(t *testing.T) {
	tests := []struct {
		tile     string
		expected bool
	}{
		{"dong", true},
		{"nan", true},
		{"xi", true},
		{"bei", true},
		{"zhong", true},
		{"fa", true},
		{"bai", true},
		{"wan-1", false},
		{"tong-9", false},
		{"tiao-5", false},
		{"flower-chun", false},
	}

	for _, tt := range tests {
		result := IsHonor(tt.tile)
		if result != tt.expected {
			t.Errorf("IsHonor(%s) = %v; expected %v", tt.tile, result, tt.expected)
		}
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		tile        string
		expectedTyp string
		expectedNum int
	}{
		{"wan-1", "wan", 1},
		{"tong-9", "tong", 9},
		{"tiao-5", "tiao", 5},
		{"dong", "dong", 0},
		{"zhong", "zhong", 0},
		{"flower-chun", "flower-chun", 0}, // Flower tile case
		{"invalid-tile", "invalid-tile", 0}, // Malformed (no '-')
		{"invalid", "invalid", 0}, // Malformed
	}

	for _, tt := range tests {
		typ, num := Parse(tt.tile)
		if typ != tt.expectedTyp || num != tt.expectedNum {
			t.Errorf("Parse(%s) = %s, %d; expected %s, %d", tt.tile, typ, num, tt.expectedTyp, tt.expectedNum)
		}
	}
}

func TestValue(t *testing.T) {
	tests := []struct {
		tile     string
		expected int
	}{
		// 萬子
		{"wan-1", 11}, {"wan-5", 15}, {"wan-9", 19},
		// 筒子
		{"tong-1", 21}, {"tong-5", 25}, {"tong-9", 29},
		// 條子
		{"tiao-1", 31}, {"tiao-5", 35}, {"tiao-9", 39},
		// 風牌 (dong=1, nan=2, xi=3, bei=4)
		{"dong", 41}, {"nan", 42}, {"xi", 43}, {"bei", 44},
		// 三元牌 (zhong=1, fa=2, bai=3)
		{"zhong", 51}, {"fa", 52}, {"bai", 53},
		// 花牌 (應返回 0)
		{"flower-chun", 0},
		{"flower-ju", 0},
		// 無效牌 (應返回 0)
		{"invalid-tile", 0},
		{"invalid", 0},
	}

	for _, tt := range tests {
		result := Value(tt.tile)
		if result != tt.expected {
			t.Errorf("Value(%s) = %d; expected %d", tt.tile, result, tt.expected)
		}
	}

	// Basic ordering check (redundant but good for sanity)
	if Value("wan-1") >= Value("wan-2") {
		t.Error("wan-1 should be < wan-2")
	}
	if Value("wan-9") >= Value("tong-1") {
		t.Error("wan-9 should be < tong-1")
	}
	if Value("tong-9") >= Value("tiao-1") {
		t.Error("tong-9 should be < tiao-1")
	}
	if Value("tiao-9") >= Value("dong") {
		t.Error("tiao-9 should be < dong")
	}
	if Value("dong") >= Value("nan") {
		t.Error("dong should be < nan")
	}
	if Value("bei") >= Value("zhong") {
		t.Error("bei should be < zhong")
	}
	if Value("zhong") >= Value("fa") {
		t.Error("zhong should be < fa")
	}
	if Value("fa") >= Value("bai") {
		t.Error("fa should be < bai")
	}
}

func TestSort(t *testing.T) {
	tiles := []string{"wan-3", "wan-1", "tong-2", "dong", "tiao-5", "wan-2"}
	Sort(tiles)

	expected := []string{"wan-1", "wan-2", "wan-3", "tong-2", "tiao-5", "dong"}
	for i, tile := range tiles {
		if tile != expected[i] {
			t.Errorf("Index %d: expected %s, got %s", i, expected[i], tile)
		}
	}
}

func TestGetUniqueTypes(t *testing.T) {
	expectedTypes := []string{
		"wan-1", "wan-2", "wan-3", "wan-4", "wan-5", "wan-6", "wan-7", "wan-8", "wan-9",
		"tong-1", "tong-2", "tong-3", "tong-4", "tong-5", "tong-6", "tong-7", "tong-8", "tong-9",
		"tiao-1", "tiao-2", "tiao-3", "tiao-4", "tiao-5", "tiao-6", "tiao-7", "tiao-8", "tiao-9",
		"dong", "nan", "xi", "bei",
		"zhong", "fa", "bai",
	}

	result := GetUniqueTypes()

	if len(result) != len(expectedTypes) {
		t.Errorf("GetUniqueTypes returned %d types, expected %d", len(result), len(expectedTypes))
		return // Return to prevent index out of range panic
	}

	for i, typ := range result {
		if typ != expectedTypes[i] {
			t.Errorf("Index %d: expected %s, got %s", i, expectedTypes[i], typ)
		}
	}
}

func TestCount(t *testing.T) {
	tiles := []string{"wan-1", "wan-1", "wan-2", "tong-3", "wan-1"}

	tests := []struct {
		tile     string
		expected int
	}{
		{"wan-1", 3},
		{"wan-2", 1},
		{"tong-3", 1},
		{"wan-5", 0}, // Not in slice
	}

	for _, tt := range tests {
		result := Count(tiles, tt.tile)
		if result != tt.expected {
			t.Errorf("Count(%v, %s) = %d; expected %d", tiles, tt.tile, result, tt.expected)
		}
	}
}

func TestRemove(t *testing.T) {
	initialTiles := []string{"wan-1", "wan-1", "wan-2", "tong-3", "wan-1", "tiao-5"}

	tests := []struct {
		name         string
		tileToRemove string
		count        int
		expected     []string
		expectedLen  int
	}{
		{
			name:         "Remove one existing tile",
			tileToRemove: "wan-2",
			count:        1,
			expected:     []string{"wan-1", "wan-1", "tong-3", "wan-1", "tiao-5"},
			expectedLen:  5,
		},
		{
			name:         "Remove multiple existing tiles",
			tileToRemove: "wan-1",
			count:        2,
			expected:     []string{"wan-1", "wan-2", "tong-3", "tiao-5"},
			expectedLen:  4,
		},
		{
			name:         "Remove non-existing tile",
			tileToRemove: "wan-5",
			count:        1,
			expected:     initialTiles, // Should remain unchanged
			expectedLen:  len(initialTiles),
		},
		{
			name:         "Remove more than available",
			tileToRemove: "wan-1",
			count:        5, // Only 3 wan-1 are available
			expected:     []string{"wan-2", "tong-3", "tiao-5"},
			expectedLen:  3,
		},
		{
			name:         "Remove zero tiles",
			tileToRemove: "wan-1",
			count:        0,
			expected:     initialTiles,
			expectedLen:  len(initialTiles),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy to avoid modifying the original slice for subsequent tests
			tilesCopy := make([]string, len(initialTiles))
			copy(tilesCopy, initialTiles)

			result := Remove(tilesCopy, tt.tileToRemove, tt.count)

			if len(result) != tt.expectedLen {
				t.Errorf("Length mismatch: expected %d, got %d", tt.expectedLen, len(result))
			}

			// Sort both for comparison, as order might not be strictly preserved
			// depending on internal Remove implementation (though it should be here)
			// For precise removal, compare directly if order is guaranteed
			// For simplicity and general correctness, a map-based comparison can be used
			// Here, assuming internal order is preserved by `Remove`'s implementation
			sort.Strings(result)
			sort.Strings(tt.expected)

			if !compareSlices(result, tt.expected) {
				t.Errorf("Result mismatch: expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Helper to compare string slices
func compareSlices(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}
