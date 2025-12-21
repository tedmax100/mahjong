package lobby

import (
	"bytes"
	"encoding/json"
	"log"
	"mahjong/internal/game"
	"net/http"
	"time"
)

// HTTPLobbyNotifier 實現 LobbyNotifier 介面，透過 HTTP 通知 Lobby Service
type HTTPLobbyNotifier struct {
	lobbyURL   string
	secretKey  string
	httpClient *http.Client
}

// NewHTTPLobbyNotifier 創建新的 HTTP Lobby 通知器
func NewHTTPLobbyNotifier(lobbyURL, secretKey string) *HTTPLobbyNotifier {
	return &HTTPLobbyNotifier{
		lobbyURL:  lobbyURL,
		secretKey: secretKey,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// NotifyRoomCreated 通知房間已創建
func (n *HTTPLobbyNotifier) NotifyRoomCreated(room *game.Room) error {
	event := RoomEvent{
		Event:  EventRoomCreated,
		RoomID: room.ID,
		Room: &LobbyRoom{
			ID:          room.ID,
			Name:        room.Name,
			HostID:      room.HostID,
			HostName:    room.HostName,
			PlayerCount: room.GetPlayerCount(),
			MaxPlayers:  4,
			IsPublic:    room.IsPublic,
			Status:      room.GetStatus(),
			CreatedAt:   room.CreatedAt,
			UpdatedAt:   time.Now(),
		},
		Timestamp: time.Now(),
	}
	return n.sendEvent(event)
}

// NotifyPlayerJoined 通知玩家加入
func (n *HTTPLobbyNotifier) NotifyPlayerJoined(room *game.Room) error {
	event := RoomEvent{
		Event:  EventPlayerJoined,
		RoomID: room.ID,
		Room: &LobbyRoom{
			ID:          room.ID,
			PlayerCount: room.GetPlayerCount(),
			Status:      room.GetStatus(),
			UpdatedAt:   time.Now(),
		},
		Timestamp: time.Now(),
	}
	return n.sendEvent(event)
}

// NotifyPlayerLeft 通知玩家離開
func (n *HTTPLobbyNotifier) NotifyPlayerLeft(room *game.Room) error {
	event := RoomEvent{
		Event:  EventPlayerLeft,
		RoomID: room.ID,
		Room: &LobbyRoom{
			ID:          room.ID,
			PlayerCount: room.GetPlayerCount(),
			Status:      room.GetStatus(),
			UpdatedAt:   time.Now(),
		},
		Timestamp: time.Now(),
	}
	return n.sendEvent(event)
}

// NotifyGameStarted 通知遊戲開始
func (n *HTTPLobbyNotifier) NotifyGameStarted(roomID string) error {
	event := RoomEvent{
		Event:  EventGameStarted,
		RoomID: roomID,
		Room: &LobbyRoom{
			ID:        roomID,
			Status:    StatusPlaying,
			UpdatedAt: time.Now(),
		},
		Timestamp: time.Now(),
	}
	return n.sendEvent(event)
}

// NotifyRoomClosed 通知房間關閉
func (n *HTTPLobbyNotifier) NotifyRoomClosed(roomID string) error {
	event := RoomEvent{
		Event:     EventRoomClosed,
		RoomID:    roomID,
		Timestamp: time.Now(),
	}
	return n.sendEvent(event)
}

// sendEvent 發送事件到 Lobby Service
func (n *HTTPLobbyNotifier) sendEvent(event RoomEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		log.Printf("[LobbyNotifier] 序列化事件失敗: %v", err)
		return err
	}

	url := n.lobbyURL + "/internal/room-events"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("[LobbyNotifier] 創建請求失敗: %v", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Secret", n.secretKey)

	resp, err := n.httpClient.Do(req)
	if err != nil {
		log.Printf("[LobbyNotifier] 發送請求失敗: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[LobbyNotifier] Lobby Service 返回錯誤狀態碼: %d", resp.StatusCode)
	} else {
		log.Printf("[LobbyNotifier] 事件 %s 已發送到 Lobby Service (房間: %s)", event.Event, event.RoomID)
	}

	return nil
}
