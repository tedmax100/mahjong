package tile

import (
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
	}

	for _, tt := range tests {
		typ, num := Parse(tt.tile)
		if typ != tt.expectedTyp || num != tt.expectedNum {
			t.Errorf("Parse(%s) = %s, %d; expected %s, %d", tt.tile, typ, num, tt.expectedTyp, tt.expectedNum)
		}
	}
}

func TestValue(t *testing.T) {
	// Basic ordering check
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
