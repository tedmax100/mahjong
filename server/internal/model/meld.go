package model

// Meld 代表一組已展示的牌（碰、槓）
type Meld struct {
	Type  string   // "pong"(碰), "kong"(槓), "chow"(吃), "kong_exposed", "kong_concealed", "kong_promoted"
	Tiles []string // 牌組
}
