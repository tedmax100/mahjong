package model

import (
	"encoding/json"
	"testing"
)

func TestMeld_JSON(t *testing.T) {
	meld := Meld{
		Type:  "pong",
		Tiles: []string{"wan-1", "wan-1", "wan-1"},
	}

	data, err := json.Marshal(meld)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Meld
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Type != meld.Type {
		t.Errorf("Expected Type %s, got %s", meld.Type, decoded.Type)
	}
	if len(decoded.Tiles) != len(meld.Tiles) {
		t.Errorf("Expected Tiles length %d, got %d", len(meld.Tiles), len(decoded.Tiles))
	}
	for i, tile := range meld.Tiles {
		if decoded.Tiles[i] != tile {
			t.Errorf("Expected Tile[%d] %s, got %s", i, tile, decoded.Tiles[i])
		}
	}
}
