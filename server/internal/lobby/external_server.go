package lobby

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// ExternalServer 外部遊戲伺服器資訊
type ExternalServer struct {
	ID            string    `json:"id"`
	DisplayName   string    `json:"displayName"`
	IP            string    `json:"ip"`
	Port          int       `json:"port"`
	WebURL        string    `json:"webUrl"`        // 遊戲客戶端 URL
	MaxRooms      int       `json:"maxRooms"`      // 最大房間數
	CurrentRooms  int       `json:"currentRooms"`  // 當前房間數
	Status        string    `json:"status"`        // online, offline
	LastHeartbeat time.Time `json:"lastHeartbeat"` // 上次心跳時間
	RegisteredAt  time.Time `json:"registeredAt"`  // 註冊時間
	Token         string    `json:"-"`             // JWT Token（不序列化）
}

// ExternalServerStatus 伺服器狀態常量
const (
	ServerStatusOnline  = "online"
	ServerStatusOffline = "offline"
)

// 心跳相關常量
const (
	HeartbeatInterval   = 30 * time.Second // 心跳間隔
	HeartbeatTimeout    = 90 * time.Second // 心跳超時（3 次失敗）
	MonitorTickInterval = 10 * time.Second // 監控檢查間隔
)

// ExternalServerStore 外部伺服器存儲
type ExternalServerStore struct {
	servers    map[string]*ExternalServer
	mu         sync.RWMutex
	jwtSecret  string
	onOffline  func(serverID string) // 伺服器離線時的回調
	stopCh     chan struct{}
	monitorWg  sync.WaitGroup
}

// NewExternalServerStore 創建新的外部伺服器存儲
func NewExternalServerStore(jwtSecret string) *ExternalServerStore {
	return &ExternalServerStore{
		servers:   make(map[string]*ExternalServer),
		jwtSecret: jwtSecret,
		stopCh:    make(chan struct{}),
	}
}

// SetOnOfflineCallback 設置伺服器離線時的回調
func (s *ExternalServerStore) SetOnOfflineCallback(callback func(serverID string)) {
	s.onOffline = callback
}

// StartMonitor 啟動心跳監控
func (s *ExternalServerStore) StartMonitor() {
	s.monitorWg.Add(1)
	go func() {
		defer s.monitorWg.Done()
		ticker := time.NewTicker(MonitorTickInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.checkHeartbeats()
			case <-s.stopCh:
				return
			}
		}
	}()
}

// StopMonitor 停止心跳監控
func (s *ExternalServerStore) StopMonitor() {
	close(s.stopCh)
	s.monitorWg.Wait()
}

// checkHeartbeats 檢查所有伺服器的心跳
func (s *ExternalServerStore) checkHeartbeats() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, server := range s.servers {
		if server.Status == ServerStatusOnline && now.Sub(server.LastHeartbeat) > HeartbeatTimeout {
			server.Status = ServerStatusOffline
			if s.onOffline != nil {
				go s.onOffline(id)
			}
		}
	}
}

// Add 添加外部伺服器
func (s *ExternalServerStore) Add(server *ExternalServer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.servers[server.ID] = server
}

// Get 獲取外部伺服器
func (s *ExternalServerStore) Get(serverID string) *ExternalServer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.servers[serverID]
}

// Remove 移除外部伺服器
func (s *ExternalServerStore) Remove(serverID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.servers, serverID)
}

// UpdateHeartbeat 更新心跳時間
func (s *ExternalServerStore) UpdateHeartbeat(serverID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	server, exists := s.servers[serverID]
	if !exists {
		return false
	}

	server.LastHeartbeat = time.Now()
	server.Status = ServerStatusOnline
	return true
}

// UpdateRoomCount 更新房間數量
func (s *ExternalServerStore) UpdateRoomCount(serverID string, count int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	server, exists := s.servers[serverID]
	if !exists {
		return false
	}

	server.CurrentRooms = count
	return true
}

// GetOnlineServers 獲取所有在線伺服器
func (s *ExternalServerStore) GetOnlineServers() []*ExternalServer {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*ExternalServer
	for _, server := range s.servers {
		if server.Status == ServerStatusOnline {
			result = append(result, server)
		}
	}
	return result
}

// GetAllServers 獲取所有伺服器
func (s *ExternalServerStore) GetAllServers() []*ExternalServer {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*ExternalServer
	for _, server := range s.servers {
		result = append(result, server)
	}
	return result
}

// Exists 檢查伺服器是否存在
func (s *ExternalServerStore) Exists(serverID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.servers[serverID]
	return exists
}

// Count 返回伺服器數量
func (s *ExternalServerStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.servers)
}

// OnlineCount 返回在線伺服器數量
func (s *ExternalServerStore) OnlineCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, server := range s.servers {
		if server.Status == ServerStatusOnline {
			count++
		}
	}
	return count
}

// RegisterRequest 外部伺服器註冊請求
type RegisterRequest struct {
	ServerID    string `json:"serverId" binding:"required"`
	DisplayName string `json:"displayName" binding:"required"`
	IP          string `json:"ip" binding:"required"`
	Port        int    `json:"port" binding:"required"`
	WebURL      string `json:"webUrl" binding:"required"`
	MaxRooms    int    `json:"maxRooms"`
	Secret      string `json:"secret" binding:"required"`
}

// RegisterResponse 外部伺服器註冊回應
type RegisterResponse struct {
	Success   bool   `json:"success"`
	Token     string `json:"token,omitempty"`
	ExpiresIn int64  `json:"expiresIn,omitempty"` // Token 有效期（秒）
	Error     string `json:"error,omitempty"`
}

// HeartbeatRequest 心跳請求
type HeartbeatRequest struct {
	CurrentRooms int `json:"currentRooms"`
}

// HeartbeatResponse 心跳回應
type HeartbeatResponse struct {
	Success   bool  `json:"success"`
	ExpiresIn int64 `json:"expiresIn,omitempty"`
	Error     string `json:"error,omitempty"`
}

// GenerateServerSignature 生成伺服器簽名（用於驗證請求）
func GenerateServerSignature(serverID, secret, timestamp string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(serverID + timestamp))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyServerSignature 驗證伺服器簽名
func VerifyServerSignature(serverID, secret, timestamp, signature string) bool {
	expected := GenerateServerSignature(serverID, secret, timestamp)
	return hmac.Equal([]byte(expected), []byte(signature))
}
