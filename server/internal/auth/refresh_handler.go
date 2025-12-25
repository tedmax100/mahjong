package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const RefreshTokenCookieName = "refresh_token"

// RefreshHandler handles token refresh requests
type RefreshHandler struct {
	tokenStore *RefreshTokenStore
	keyManager *KeyManager
}

// NewRefreshHandler creates a new refresh handler
func NewRefreshHandler(tokenStore *RefreshTokenStore, keyManager *KeyManager) *RefreshHandler {
	return &RefreshHandler{
		tokenStore: tokenStore,
		keyManager: keyManager,
	}
}

// Handler handles POST /auth/refresh
func (h *RefreshHandler) Handler(c *gin.Context) {
	// Get refresh token from cookie
	refreshToken, err := c.Cookie(RefreshTokenCookieName)
	if err != nil || refreshToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "No refresh token provided",
		})
		return
	}

	// Validate refresh token
	tokenInfo, valid := h.tokenStore.Validate(refreshToken)
	if !valid {
		// Clear invalid cookie
		clearRefreshTokenCookie(c)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid or expired refresh token",
		})
		return
	}

	// Generate new access token
	claims := &AuthTokenClaims{
		Sub:     tokenInfo.UserID,
		Name:    tokenInfo.UserName,
		Picture: tokenInfo.Picture,
	}

	accessToken, err := h.keyManager.SignToken(claims)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate access token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken": accessToken,
	})
}

// clearRefreshTokenCookie clears the refresh token cookie
func clearRefreshTokenCookie(c *gin.Context) {
	isSecure := isSecureContext(c)

	c.SetCookie(
		RefreshTokenCookieName,
		"",
		-1, // MaxAge -1 deletes the cookie
		"/",
		"",
		isSecure,
		true, // HttpOnly
	)
}

// isSecureContext determines if the request is in a secure context
func isSecureContext(c *gin.Context) bool {
	// Check X-Forwarded-Proto header (for reverse proxies)
	if proto := c.GetHeader("X-Forwarded-Proto"); proto == "https" {
		return true
	}

	// Check if the request is using TLS
	if c.Request.TLS != nil {
		return true
	}

	// Check the host - localhost is considered secure for development
	host := c.Request.Host
	if host == "localhost" || len(host) > 10 && host[:10] == "localhost:" {
		return false // Use non-secure cookie for localhost
	}
	if host == "127.0.0.1" || len(host) > 10 && host[:10] == "127.0.0.1:" {
		return false
	}

	// Default to secure for production
	return true
}
