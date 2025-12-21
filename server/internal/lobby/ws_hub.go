package lobby

import (
	"encoding/json"
	"log"
	"sync"
)

// LobbyHub 管理大廳 WebSocket 連接
type LobbyHub struct {
	// 所有客戶端
	clients map[*LobbyClient]bool

	// 廣播通道
	broadcast chan []byte

	// 註冊請求
	register chan *LobbyClient

	// 註銷請求
	unregister chan *LobbyClient

	// 存儲引用
	store *LobbyStore

	// 互斥鎖
	mu sync.RWMutex
}

// NewLobbyHub 創建新的大廳 Hub
func NewLobbyHub(store *LobbyStore) *LobbyHub {
	return &LobbyHub{
		clients:    make(map[*LobbyClient]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *LobbyClient),
		unregister: make(chan *LobbyClient),
		store:      store,
	}
}

// Run 運行 Hub 主迴圈
func (h *LobbyHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			count := len(h.clients)
			h.mu.Unlock()

			log.Printf("[LobbyHub] 客戶端連接: %s (%s), 當前在線: %d", client.userName, client.userID, count)

			// 發送初始房間列表
			h.sendRoomListTo(client)

			// 發送最近的聊天訊息
			h.sendRecentMessagesTo(client)

			// 廣播在線人數更新
			h.broadcastOnlineCount()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			count := len(h.clients)
			h.mu.Unlock()

			log.Printf("[LobbyHub] 客戶端斷開: %s (%s), 當前在線: %d", client.userName, client.userID, count)

			// 廣播在線人數更新
			h.broadcastOnlineCount()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// 客戶端緩衝區滿，關閉連接
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastRoomList 廣播房間列表更新
func (h *LobbyHub) BroadcastRoomList() {
	rooms := h.store.GetPublicRooms()

	msg := WSMessage{
		Type: WSTypeRoomList,
		Data: map[string]interface{}{
			"rooms": rooms,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[LobbyHub] 序列化房間列表失敗: %v", err)
		return
	}

	h.broadcast <- data
}

// BroadcastChatMessage 廣播聊天訊息
func (h *LobbyHub) BroadcastChatMessage(msg *ChatMessage) {
	wsMsg := WSMessage{
		Type: WSTypeChatMessage,
		Data: msg,
	}

	data, err := json.Marshal(wsMsg)
	if err != nil {
		log.Printf("[LobbyHub] 序列化聊天訊息失敗: %v", err)
		return
	}

	h.broadcast <- data
}

// sendRoomListTo 發送房間列表給單個客戶端
func (h *LobbyHub) sendRoomListTo(client *LobbyClient) {
	rooms := h.store.GetPublicRooms()

	msg := WSMessage{
		Type: WSTypeRoomList,
		Data: map[string]interface{}{
			"rooms": rooms,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[LobbyHub] 序列化房間列表失敗: %v", err)
		return
	}

	select {
	case client.send <- data:
	default:
		// 緩衝區滿，忽略
	}
}

// sendRecentMessagesTo 發送最近的聊天訊息給單個客戶端
func (h *LobbyHub) sendRecentMessagesTo(client *LobbyClient) {
	messages := h.store.GetRecentMessages(50)

	for _, msg := range messages {
		wsMsg := WSMessage{
			Type: WSTypeChatMessage,
			Data: msg,
		}

		data, err := json.Marshal(wsMsg)
		if err != nil {
			continue
		}

		select {
		case client.send <- data:
		default:
			// 緩衝區滿，停止發送
			return
		}
	}
}

// broadcastOnlineCount 廣播在線人數
func (h *LobbyHub) broadcastOnlineCount() {
	h.mu.RLock()
	count := len(h.clients)
	h.mu.RUnlock()

	msg := WSMessage{
		Type: WSTypeOnlineCount,
		Data: map[string]interface{}{
			"count": count,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[LobbyHub] 序列化在線人數失敗: %v", err)
		return
	}

	h.broadcast <- data
}

// GetOnlineCount 獲取在線人數
func (h *LobbyHub) GetOnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// Register 註冊客戶端
func (h *LobbyHub) Register(client *LobbyClient) {
	h.register <- client
}

// Unregister 註銷客戶端
func (h *LobbyHub) Unregister(client *LobbyClient) {
	h.unregister <- client
}
