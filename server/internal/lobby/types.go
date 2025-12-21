package lobby

import (
	"time"
)

// LobbyRoom 大廳中顯示的房間資訊
type LobbyRoom struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	HostID      string    `json:"hostId"`
	HostName    string    `json:"hostName"`
	PlayerCount int       `json:"playerCount"`
	MaxPlayers  int       `json:"maxPlayers"`
	IsPublic    bool      `json:"isPublic"`
	Status      string    `json:"status"` // waiting, playing, closed
	ServerAddr  string    `json:"serverAddr"`
	ServerID    string    `json:"serverId,omitempty"`   // 伺服器 ID（外部伺服器用）
	IsExternal  bool      `json:"isExternal,omitempty"` // 是否外部伺服器
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ChatMessage 大廳聊天訊息
type ChatMessage struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	UserName  string    `json:"userName"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"` // text, system
}

// RoomEvent Game Server 發送的房間事件
type RoomEvent struct {
	Event     string     `json:"event"` // room_created, player_joined, player_left, game_started, room_closed
	RoomID    string     `json:"roomId"`
	Room      *LobbyRoom `json:"room,omitempty"`
	Timestamp time.Time  `json:"timestamp"`
}

// RoomEventType 房間事件類型常量
const (
	EventRoomCreated  = "room_created"
	EventPlayerJoined = "player_joined"
	EventPlayerLeft   = "player_left"
	EventGameStarted  = "game_started"
	EventRoomClosed   = "room_closed"
)

// RoomStatus 房間狀態常量
const (
	StatusWaiting = "waiting"
	StatusPlaying = "playing"
	StatusClosed  = "closed"
)

// WSMessage WebSocket 訊息結構
type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// WSMessageType WebSocket 訊息類型常量
const (
	WSTypeRoomList    = "room_list"
	WSTypeRoomUpdate  = "room_update"
	WSTypeChatMessage = "chat_message"
	WSTypeOnlineCount = "online_count"
	WSTypeError       = "error"
)

// ChatRequest 聊天請求
type ChatRequest struct {
	Content string `json:"content"`
}

// CreateRoomRequest 創建房間請求（從大廳）
type CreateRoomRequest struct {
	UserID   string `json:"userId"`
	UserName string `json:"userName"`
	RoomName string `json:"roomName,omitempty"`
	IsPublic bool   `json:"isPublic"`
}

// CreateRoomResponse 創建房間響應
type CreateRoomResponse struct {
	Success    bool   `json:"success"`
	RoomID     string `json:"roomId,omitempty"`
	ServerAddr string `json:"serverAddr,omitempty"`
	Error      string `json:"error,omitempty"`
}

// ListRoomsResponse 房間列表響應
type ListRoomsResponse struct {
	Success bool        `json:"success"`
	Rooms   []LobbyRoom `json:"rooms"`
}
