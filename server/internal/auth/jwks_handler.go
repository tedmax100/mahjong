package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// JWKSHandler 處理 JWKS 端點請求
type JWKSHandler struct {
	keyManager *KeyManager
}

// NewJWKSHandler 建立新的 JWKS 處理器
func NewJWKSHandler(keyManager *KeyManager) *JWKSHandler {
	return &JWKSHandler{
		keyManager: keyManager,
	}
}

// Handler GET /.well-known/jwks.json - 回傳 JWKS
func (h *JWKSHandler) Handler(c *gin.Context) {
	jwks := h.keyManager.GetJWKS()

	// 設定快取標頭
	c.Header("Content-Type", "application/json")
	c.Header("Cache-Control", "public, max-age=3600") // 1 小時

	c.JSON(http.StatusOK, jwks)
}
