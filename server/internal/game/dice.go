package game

import (
	"crypto/rand"
	"log"
	"math/big"
	mathrand "math/rand"
	"time"
)

// DiceRollResult holds the result of rolling dice for dealer selection
type DiceRollResult struct {
	DiceResults     [3]int `json:"diceResults"`     // Individual dice values (1-6)
	TotalSum        int    `json:"totalSum"`        // Sum of all dice
	DealerSeatIndex int    `json:"dealerSeatIndex"` // Calculated dealer position (0-3)
	DealerPlayerID  string `json:"dealerPlayerId"`  // Player ID of the dealer
}

// RollDiceForDealer generates 3 random dice values and calculates dealer position
// using Taiwan Mahjong rules: (sum - 1) % 4
func RollDiceForDealer(players []*Player) *DiceRollResult {
	result := &DiceRollResult{}

	// Generate 3 cryptographically secure random dice values
	for i := 0; i < 3; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(6))
		if err != nil {
			// Fallback to less secure random if crypto/rand fails
			log.Printf("crypto/rand 失敗，使用 math/rand: %v", err)
			mathrand.Seed(time.Now().UnixNano())
			result.DiceResults[i] = mathrand.Intn(6) + 1
		} else {
			result.DiceResults[i] = int(n.Int64()) + 1 // 1-6
		}
	}

	// Calculate sum
	result.TotalSum = result.DiceResults[0] + result.DiceResults[1] + result.DiceResults[2]

	// Calculate dealer position using Taiwan Mahjong rules
	// Starting from East (position 0), count counter-clockwise
	// (sum - 1) % 4 maps sum to position 0-3
	// Sum 1,5,9,13,17  -> Position 0 (East)
	// Sum 2,6,10,14,18 -> Position 1 (South)
	// Sum 3,7,11,15    -> Position 2 (West)
	// Sum 4,8,12,16    -> Position 3 (North)
	result.DealerSeatIndex = (result.TotalSum - 1) % 4

	// Get dealer player ID
	if len(players) > result.DealerSeatIndex {
		result.DealerPlayerID = players[result.DealerSeatIndex].ID
	}

	log.Printf("擲骰結果: %v, 總和: %d, 莊家位置: %d",
		result.DiceResults, result.TotalSum, result.DealerSeatIndex)

	return result
}
