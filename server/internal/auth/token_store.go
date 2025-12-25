package auth

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// RefreshTokenInfo stores the information associated with a refresh token
type RefreshTokenInfo struct {
	UserID    string
	UserName  string
	Picture   string
	ExpiresAt time.Time
}

// RefreshTokenStore manages refresh tokens in memory
type RefreshTokenStore struct {
	tokens     sync.Map
	expiration time.Duration
	stopCh     chan struct{}
}

// NewRefreshTokenStore creates a new refresh token store
func NewRefreshTokenStore(expiration time.Duration) *RefreshTokenStore {
	store := &RefreshTokenStore{
		expiration: expiration,
		stopCh:     make(chan struct{}),
	}

	// Start cleanup goroutine
	go store.cleanupLoop()

	return store
}

// Store creates and stores a new refresh token, returns the token ID
func (s *RefreshTokenStore) Store(userID, userName, picture string) string {
	tokenID := uuid.New().String()

	info := RefreshTokenInfo{
		UserID:    userID,
		UserName:  userName,
		Picture:   picture,
		ExpiresAt: time.Now().Add(s.expiration),
	}

	s.tokens.Store(tokenID, info)

	return tokenID
}

// Validate checks if a refresh token is valid and returns the associated user info
func (s *RefreshTokenStore) Validate(tokenID string) (*RefreshTokenInfo, bool) {
	value, ok := s.tokens.Load(tokenID)
	if !ok {
		return nil, false
	}

	info := value.(RefreshTokenInfo)

	// Check if token is expired
	if time.Now().After(info.ExpiresAt) {
		s.tokens.Delete(tokenID)
		return nil, false
	}

	return &info, true
}

// Revoke removes a refresh token from the store
func (s *RefreshTokenStore) Revoke(tokenID string) {
	s.tokens.Delete(tokenID)
}

// RevokeAllForUser removes all refresh tokens for a specific user
func (s *RefreshTokenStore) RevokeAllForUser(userID string) {
	s.tokens.Range(func(key, value interface{}) bool {
		info := value.(RefreshTokenInfo)
		if info.UserID == userID {
			s.tokens.Delete(key)
		}
		return true
	})
}

// cleanupLoop periodically removes expired tokens
func (s *RefreshTokenStore) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cleanup()
		case <-s.stopCh:
			return
		}
	}
}

// cleanup removes all expired tokens
func (s *RefreshTokenStore) cleanup() {
	now := time.Now()
	s.tokens.Range(func(key, value interface{}) bool {
		info := value.(RefreshTokenInfo)
		if now.After(info.ExpiresAt) {
			s.tokens.Delete(key)
		}
		return true
	})
}

// Stop stops the cleanup goroutine
func (s *RefreshTokenStore) Stop() {
	close(s.stopCh)
}

// Count returns the number of tokens in the store (for debugging)
func (s *RefreshTokenStore) Count() int {
	count := 0
	s.tokens.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}
