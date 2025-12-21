package lobby

import (
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	// MaxChatMessages 保留的最大聊天訊息數量
	MaxChatMessages = 100
)

// LobbyStore 大廳數據存儲
type LobbyStore struct {
	rooms    map[string]*LobbyRoom
	messages []ChatMessage
	mu       sync.RWMutex
}

// NewLobbyStore 創建新的大廳存儲
func NewLobbyStore() *LobbyStore {
	return &LobbyStore{
		rooms:    make(map[string]*LobbyRoom),
		messages: make([]ChatMessage, 0, MaxChatMessages),
	}
}

// AddRoom 添加房間
func (s *LobbyStore) AddRoom(room *LobbyRoom) {
	s.mu.Lock()
	defer s.mu.Unlock()

	room.UpdatedAt = time.Now()
	s.rooms[room.ID] = room
	log.Printf("[LobbyStore] 添加房間: %s (公開: %t, 狀態: %s, 玩家數: %d)", room.ID, room.IsPublic, room.Status, room.PlayerCount)
}

// UpdateRoom 更新房間資訊
func (s *LobbyStore) UpdateRoom(roomID string, update func(*LobbyRoom)) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	room, exists := s.rooms[roomID]
	if !exists {
		return false
	}

	update(room)
	room.UpdatedAt = time.Now()
	return true
}

// RemoveRoom 移除房間
func (s *LobbyStore) RemoveRoom(roomID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.rooms[roomID]; exists {
		delete(s.rooms, roomID)
		return true
	}
	return false
}

// GetRoom 獲取單個房間
func (s *LobbyStore) GetRoom(roomID string) (*LobbyRoom, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	room, exists := s.rooms[roomID]
	if !exists {
		return nil, false
	}

	// 返回副本以避免並發修改
	roomCopy := *room
	return &roomCopy, true
}

// GetPublicRooms 獲取所有可加入的公開房間
// 過濾條件: IsPublic=true AND Status=waiting AND PlayerCount < MaxPlayers
func (s *LobbyStore) GetPublicRooms() []LobbyRoom {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]LobbyRoom, 0)
	for _, room := range s.rooms {
		if room.IsPublic && room.Status == StatusWaiting && room.PlayerCount < room.MaxPlayers {
			result = append(result, *room)
		}
	}

	return result
}

// GetAllRooms 獲取所有房間（用於調試）
func (s *LobbyStore) GetAllRooms() []LobbyRoom {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]LobbyRoom, 0, len(s.rooms))
	for _, room := range s.rooms {
		result = append(result, *room)
	}

	return result
}

// AddChatMessage 添加聊天訊息
func (s *LobbyStore) AddChatMessage(userID, userName, content string) *ChatMessage {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg := ChatMessage{
		ID:        uuid.New().String()[:8],
		UserID:    userID,
		UserName:  userName,
		Content:   content,
		Timestamp: time.Now(),
		Type:      "text",
	}

	s.messages = append(s.messages, msg)

	// 保持訊息數量在限制內
	if len(s.messages) > MaxChatMessages {
		s.messages = s.messages[len(s.messages)-MaxChatMessages:]
	}

	return &msg
}

// AddSystemMessage 添加系統訊息
func (s *LobbyStore) AddSystemMessage(content string) *ChatMessage {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg := ChatMessage{
		ID:        uuid.New().String()[:8],
		UserID:    "system",
		UserName:  "系統",
		Content:   content,
		Timestamp: time.Now(),
		Type:      "system",
	}

	s.messages = append(s.messages, msg)

	// 保持訊息數量在限制內
	if len(s.messages) > MaxChatMessages {
		s.messages = s.messages[len(s.messages)-MaxChatMessages:]
	}

	return &msg
}

// GetRecentMessages 獲取最近的聊天訊息
func (s *LobbyStore) GetRecentMessages(limit int) []ChatMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 || limit > len(s.messages) {
		limit = len(s.messages)
	}

	start := len(s.messages) - limit
	if start < 0 {
		start = 0
	}

	result := make([]ChatMessage, limit)
	copy(result, s.messages[start:])
	return result
}

// RoomCount 獲取房間總數
func (s *LobbyStore) RoomCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.rooms)
}

// PublicRoomCount 獲取可加入的公開房間數量
func (s *LobbyStore) PublicRoomCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, room := range s.rooms {
		if room.IsPublic && room.Status == StatusWaiting && room.PlayerCount < room.MaxPlayers {
			count++
		}
	}
	return count
}
