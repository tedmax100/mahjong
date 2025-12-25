package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// LogoutHandler handles logout requests
type LogoutHandler struct {
	tokenStore *RefreshTokenStore
}

// NewLogoutHandler creates a new logout handler
func NewLogoutHandler(tokenStore *RefreshTokenStore) *LogoutHandler {
	return &LogoutHandler{
		tokenStore: tokenStore,
	}
}

// Handler handles POST /auth/logout
func (h *LogoutHandler) Handler(c *gin.Context) {
	// Get refresh token from cookie
	refreshToken, err := c.Cookie(RefreshTokenCookieName)
	if err == nil && refreshToken != "" {
		// Revoke the token
		h.tokenStore.Revoke(refreshToken)
	}

	// Clear the cookie
	clearRefreshTokenCookie(c)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logged out successfully",
	})
}
