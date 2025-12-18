package auth

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var (
	jwks       keyfunc.Keyfunc
	jwksCancel context.CancelFunc
	jwksMutex  sync.RWMutex
	initOnce   sync.Once
)

// UserClaims JWT 中的使用者資訊
type UserClaims struct {
	Sub     string `json:"sub"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
	jwt.RegisteredClaims
}

// InitJWKS 初始化 JWKS 客戶端
// 從 AUTH_PROXY_URL 環境變數取得 JWKS 端點
func InitJWKS() error {
	var initErr error

	initOnce.Do(func() {
		authProxyURL := os.Getenv("AUTH_PROXY_URL")
		if authProxyURL == "" {
			log.Println("[Auth] AUTH_PROXY_URL 未設定，JWT 驗證功能停用")
			return
		}

		jwksURL := authProxyURL + "/.well-known/jwks.json"
		log.Printf("[Auth] 初始化 JWKS 客戶端: %s", jwksURL)

		// 建立帶取消功能的 context
		ctx, cancel := context.WithCancel(context.Background())

		// 建立 JWKS 客戶端
		kf, err := keyfunc.NewDefaultCtx(ctx, []string{jwksURL})
		if err != nil {
			cancel() // 發生錯誤時取消 context
			initErr = err
			log.Printf("[Auth] JWKS 初始化失敗: %v", err)
			return
		}

		jwksMutex.Lock()
		jwks = kf
		jwksCancel = cancel
		jwksMutex.Unlock()

		log.Println("[Auth] JWKS 初始化成功")
	})

	return initErr
}

// ShutdownJWKS 關閉 JWKS 客戶端
func ShutdownJWKS() {
	jwksMutex.Lock()
	defer jwksMutex.Unlock()

	if jwksCancel != nil {
		jwksCancel()
		jwksCancel = nil
		jwks = nil
		log.Println("[Auth] JWKS 客戶端已關閉")
	}
}

// JWTMiddleware 必要的 JWT 驗證中介軟體
// 如果 token 無效或不存在，回傳 401
func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果 JWKS 未設定，跳過驗證
		jwksMutex.RLock()
		localJwks := jwks
		jwksMutex.RUnlock()

		if localJwks == nil {
			c.Next()
			return
		}

		// 取得 Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Missing authorization header",
				"message": "缺少認證標頭",
			})
			return
		}

		// 解析 Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid authorization header format",
				"message": "認證標頭格式錯誤",
			})
			return
		}

		tokenString := parts[1]

		// 驗證 JWT
		claims, err := validateToken(tokenString, localJwks)
		if err != nil {
			log.Printf("[Auth] Token 驗證失敗: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid token",
				"message": "Token 無效或已過期",
			})
			return
		}

		// 將使用者資訊存入 context
		c.Set("userId", claims.Sub)
		c.Set("userName", claims.Name)
		c.Set("userPicture", claims.Picture)
		c.Set("claims", claims)

		c.Next()
	}
}

// OptionalJWTMiddleware 可選的 JWT 驗證中介軟體
// 如果有 token 則驗證，無 token 或驗證失敗則繼續（不設定使用者資訊）
func OptionalJWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果 JWKS 未設定，跳過驗證
		jwksMutex.RLock()
		localJwks := jwks
		jwksMutex.RUnlock()

		if localJwks == nil {
			c.Next()
			return
		}

		// 取得 Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// 解析 Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]

		// 驗證 JWT（失敗時不中斷，只是不設定使用者資訊）
		claims, err := validateToken(tokenString, localJwks)
		if err != nil {
			log.Printf("[Auth] Optional token 驗證失敗: %v", err)
			c.Next()
			return
		}

		// 將使用者資訊存入 context
		c.Set("userId", claims.Sub)
		c.Set("userName", claims.Name)
		c.Set("userPicture", claims.Picture)
		c.Set("claims", claims)

		c.Next()
	}
}

// validateToken 驗證 JWT token
func validateToken(tokenString string, kf keyfunc.Keyfunc) (*UserClaims, error) {
	claims := &UserClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, kf.Keyfunc,
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithIssuer("mahjong-auth-proxy"),
		jwt.WithExpirationRequired(),
	)

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrTokenNotValidYet
	}

	return claims, nil
}

// GetUserFromContext 從 Gin context 取得使用者資訊
func GetUserFromContext(c *gin.Context) (userId, userName, userPicture string, ok bool) {
	userIdVal, exists := c.Get("userId")
	if !exists {
		return "", "", "", false
	}

	userId, ok = userIdVal.(string)
	if !ok {
		return "", "", "", false
	}

	userNameVal, _ := c.Get("userName")
	userName, _ = userNameVal.(string)

	userPictureVal, _ := c.Get("userPicture")
	userPicture, _ = userPictureVal.(string)

	return userId, userName, userPicture, true
}

// GetClaimsFromContext 從 Gin context 取得完整的 claims
func GetClaimsFromContext(c *gin.Context) (*UserClaims, bool) {
	claimsVal, exists := c.Get("claims")
	if !exists {
		return nil, false
	}

	claims, ok := claimsVal.(*UserClaims)
	return claims, ok
}
