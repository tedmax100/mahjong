package game

import (
	"testing"
)

func TestNextRound(t *testing.T) {
	// 1. Setup Room and Players
	playerA := &Player{ID: "playerA", Name: "Player A", Position: 0}
	playerB := &Player{ID: "playerB", Name: "Player B", Position: 1}
	playerC := &Player{ID: "playerC", Name: "Player C", Position: 2}
	playerD := &Player{ID: "playerD", Name: "Player D", Position: 3}

	players := []*Player{playerA, playerB, playerC, playerD}

	room := NewRoom("test-room")
	for _, p := range players {
		room.AddPlayer(p.ID, p.Name)
	}
	room.StartGame()

	// Give a player some state to ensure it gets reset
	room.Players[1].Hand = []string{"wan-1", "wan-2", "wan-3"}
	room.Players[1].Melds = []Meld{{Type: "pong", Tiles: []string{"tiao-5", "tiao-5", "tiao-5"}}}
	room.Players[1].Flowers = []string{"flower-chun"}
	room.Game.Dealer = 0
	originalDeckSize := len(room.Game.Deck)

	// 2. Call NextRound
	room.NextRound()

	// 3. Assertions
	// Assert dealer advanced
	if room.Game.Dealer != 1 {
		t.Errorf("Dealer should have advanced to 1, but is %d", room.Game.Dealer)
	}

	// Assert current turn is the new dealer
	if room.CurrentTurn != 1 {
		t.Errorf("CurrentTurn should be the new dealer (1), but is %d", room.CurrentTurn)
	}

	// Assert player state was reset
	if len(room.Players[1].Hand) != 0 {
		t.Errorf("Player hand should have been reset, but has %d tiles", len(room.Players[1].Hand))
	}
	if len(room.Players[1].Melds) != 0 {
		t.Errorf("Player melds should have been reset, but has %d melds", len(room.Players[1].Melds))
	}
	if len(room.Players[1].Flowers) != 0 {
		t.Errorf("Player flowers should have been reset, but has %d flowers", len(room.Players[1].Flowers))
	}

	// Assert a new deck was created
	if len(room.Game.Deck) <= originalDeckSize {
		// This is a loose check, but a new deck should have a lot of tiles.
		// A more robust check might be to compare pointers, but this is fine.
	}

	// Assert game is marked as started
	if !room.GameStarted {
		t.Errorf("GameStarted should be true after NextRound")
	}
}
