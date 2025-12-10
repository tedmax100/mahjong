package websocket

import (
	"encoding/json"
	"mahjong/internal/game"
	"mahjong/internal/model"
	"testing"
	"time"
)

// TestHandleGameAction_Discard 測試出牌動作
func TestHandleGameAction_Discard(t *testing.T) {
	// 1. Setup Hub and Room
	hub := NewHub()
	// Fill hand with 17 cards (16 + 1 to discard)
	hand := []string{"wan-1"}
	for i := 0; i < 16; i++ {
		hand = append(hand, "wan-2")
	}
	p1 := &game.Player{ID: "p1", Name: "Player1", Position: 0, Hand: hand, Melds: []model.Meld{}}
	p2 := &game.Player{ID: "p2", Name: "Player2", Position: 1, Hand: []string{"tong-1"}, Melds: []model.Meld{}}
	
	room := &game.Room{
		ID:      "test-room",
		Players: []*game.Player{p1, p2},
		Game:    game.NewMahjongGame([]*game.Player{p1, p2}),
		Clients: make(map[string]interface{}),
	}
	room.GameStarted = true
	room.CurrentTurn = 0 // p1 turn

	hub.rooms["test-room"] = room

	// 2. Setup Client
	client := &Client{
		Hub:      hub,
		RoomID:   "test-room",
		UserID:   "p1",
		UserName: "Player1",
		Room:     room,
		Send:     make(chan []byte, 10), // Buffered channel
	}
	// Register client in room (needed for broadcasting)
	room.Clients["p1"] = client

	// 3. Action Data
	data := map[string]interface{}{
		"tile": "wan-1",
	}

	// 4. Execute
	client.handleGameAction("discard", data)

	// 5. Verify Hand
	if len(p1.Hand) != 16 {
		t.Errorf("Hand size should be 16 after discard, got %d", len(p1.Hand))
	}
	// wan-1 should be removed, only wan-2s left
	for _, tile := range p1.Hand {
		if tile == "wan-1" {
			t.Errorf("wan-1 should be removed")
		}
	}

	// 6. Verify Broadcast (check channel)
	select {
	case msg := <-client.Send:
		var message map[string]interface{}
		json.Unmarshal(msg, &message)
		if message["type"] != "player_action" {
			t.Errorf("Expected player_action message")
		}
		data := message["data"].(map[string]interface{})
		if data["action"] != "discard" {
			t.Errorf("Expected discard action")
		}
		if data["tile"] != "wan-1" {
			t.Errorf("Expected wan-1 tile")
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for broadcast message")
	}
}

// TestHandleGameAction_Pong 測試碰牌動作
func TestHandleGameAction_Pong(t *testing.T) {
	// 1. Setup Hub and Room
	hub := NewHub()
	p1 := &game.Player{
		ID:       "p1",
		Name:     "Player1",
		Position: 0,
		Hand:     []string{"wan-1", "wan-1", "wan-2"}, // Pair of wan-1
		Melds:    []model.Meld{},
	}
	p2 := &game.Player{ID: "p2", Name: "Player2", Position: 1}
	
	room := &game.Room{
		ID:      "test-room",
		Players: []*game.Player{p1, p2},
		Game:    game.NewMahjongGame([]*game.Player{p1, p2}),
		Clients: make(map[string]interface{}),
	}
	room.GameStarted = true
	
	// Simulate p2 discarded wan-1
	room.CurrentTurn = 1
	room.LastDiscardPlayer = 1
	room.Game.DiscardPile = append(room.Game.DiscardPile, "wan-1")

	hub.rooms["test-room"] = room

	// 2. Setup Client for p1
	client := &Client{
		Hub:      hub,
		RoomID:   "test-room",
		UserID:   "p1",
		UserName: "Player1",
		Room:     room,
		Send:     make(chan []byte, 10),
	}
	room.Clients["p1"] = client

	// 3. Action Data
	data := map[string]interface{}{
		"tile": "wan-1",
	}

	// 4. Execute
	client.handleGameAction("pong", data)

	// 5. Verify State
	// Hand should reduce by 2 (2 removed for pong)
	if len(p1.Hand) != 1 { // 3 - 2 = 1
		t.Errorf("Hand size should be 1 after pong, got %d", len(p1.Hand))
	}
	// Melds should increase
	if len(p1.Melds) != 1 {
		t.Errorf("Should have 1 meld")
	}
	if p1.Melds[0].Type != "pong" {
		t.Errorf("Meld type should be pong")
	}
	// Turn should be p1
	if room.CurrentTurn != 0 {
		t.Errorf("Turn should be 0 (p1), got %d", room.CurrentTurn)
	}

	// 6. Verify Broadcast
	select {
	case msg := <-client.Send:
		var message map[string]interface{}
		json.Unmarshal(msg, &message)
		if message["type"] != "player_action" {
			t.Errorf("Expected player_action message")
		}
		data := message["data"].(map[string]interface{})
		if data["action"] != "pong" {
			t.Errorf("Expected pong action")
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for broadcast message")
	}
}

// TestHandleGameAction_Chow 測試吃牌動作
func TestHandleGameAction_Chow(t *testing.T) {
	// 1. Setup
	hub := NewHub()
	p1 := &game.Player{
		ID:       "p1",
		Position: 0,
		Hand:     []string{"wan-2", "wan-3", "tong-1"},
		Melds:    []model.Meld{},
	}
	p2 := &game.Player{ID: "p2", Position: 3} // p2 is previous player (3) for p1 (0) in 4 players? No, 2 players.
	// Chow rule: only from previous player. (pos+3)%4 == discarder.
	// If 2 players, logic might be different or standard logic applies.
	// Standard logic: (p.Position + 3) % 4 == discarderPosition.
	// If p1 is 0, (0+3)%4 = 3. So discarder must be 3.
	// Let's create 4 players to be safe.
	p2 = &game.Player{ID: "p2", Position: 1}
	p3 := &game.Player{ID: "p3", Position: 2}
	p4 := &game.Player{ID: "p4", Position: 3} // Discarder

	room := &game.Room{
		ID:      "test-room",
		Players: []*game.Player{p1, p2, p3, p4},
		Game:    game.NewMahjongGame([]*game.Player{p1, p2, p3, p4}),
		Clients: make(map[string]interface{}),
	}
	room.GameStarted = true
	
	// Simulate p4 discarded wan-1
	room.CurrentTurn = 3
	room.LastDiscardPlayer = 3
	room.Game.DiscardPile = append(room.Game.DiscardPile, "wan-1")

	hub.rooms["test-room"] = room

	client := &Client{
		Hub:    hub,
		UserID: "p1",
		Room:   room,
		Send:   make(chan []byte, 10),
	}
	room.Clients["p1"] = client

	// 2. Action Data
	data := map[string]interface{}{
		"tile":      "wan-1",
		"chowTiles": []interface{}{"wan-1", "wan-2", "wan-3"},
	}

	// 3. Execute
	client.handleGameAction("chow", data)

	// 4. Verify
	if len(p1.Melds) != 1 {
		t.Errorf("Should have 1 meld")
	} else {
		if p1.Melds[0].Type != "chow" {
			t.Errorf("Meld type should be chow")
		}
	}
}

// TestHandleGameAction_Hu 測試胡牌動作
func TestHandleGameAction_Hu(t *testing.T) {
	// 1. Setup
	hub := NewHub()
	// P1 hand: 16 cards, waiting for wan-1 to win
	// Simple hand: 5 triplets + 1 single (waiting for pair)
	hand := []string{"wan-1"} // waiting for wan-1
	for i := 0; i < 5; i++ {
		hand = append(hand, "tong-1", "tong-1", "tong-1")
	}
	
	p1 := &game.Player{ID: "p1", Name: "P1", Position: 0, Hand: hand, Melds: []model.Meld{}}
	p2 := &game.Player{ID: "p2", Name: "P2", Position: 1}

	room := &game.Room{
		ID:      "test-room",
		Players: []*game.Player{p1, p2},
		Game:    game.NewMahjongGame([]*game.Player{p1, p2}),
		Clients: make(map[string]interface{}),
	}
	room.GameStarted = true
	room.CurrentTurn = 1 // P2 turn

	hub.rooms["test-room"] = room

	client := &Client{
		Hub:    hub,
		UserID: "p1",
		Room:   room,
		Send:   make(chan []byte, 10),
	}
	room.Clients["p1"] = client

	// 2. Action Data
	// Simulate P2 discarded "wan-1"
	data := map[string]interface{}{
		"tile":      "wan-1",
		"selfDrawn": false,
	}

	// 3. Execute
	client.handleGameAction("hu", data)

	// 4. Verify
	// Game should end
	if room.GameStarted {
		t.Errorf("Game should have ended (GameStarted=false)")
	}
	
	// Verify Broadcast
	select {
	case msg := <-client.Send:
		var message map[string]interface{}
		json.Unmarshal(msg, &message)
		if message["type"] == "player_action" {
			data := message["data"].(map[string]interface{})
			if data["action"] != "hu" {
				t.Errorf("Expected hu action")
			}
			// Check for next message (game_win)
			select {
			case msg2 := <-client.Send:
				var message2 map[string]interface{}
				json.Unmarshal(msg2, &message2)
				if message2["type"] != "game_win" {
					t.Errorf("Expected game_win message, got %s", message2["type"])
				}
			case <-time.After(1 * time.Second):
				t.Error("Timeout waiting for game_win message")
			}
		} else if message["type"] == "game_win" {
			// Okay too
		} else {
			t.Errorf("Unexpected message type: %s", message["type"])
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for broadcast message")
	}
}
