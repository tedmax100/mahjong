package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// OAuthConfig Google OAuth 設定
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string // Auth Proxy 的 callback URL
}

// OAuthSession OAuth 流程中的 session 資料
type OAuthSession struct {
	State     string
	Origin    string
	CreatedAt time.Time
}

// OAuthHandler 處理 OAuth 流程
type OAuthHandler struct {
	config     *OAuthConfig
	keyManager *KeyManager
	tokenStore *RefreshTokenStore
	sessions   map[string]*OAuthSession
	mutex      sync.RWMutex
}

// GoogleTokenResponse Google token 端點回應
type GoogleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
}

// GoogleUserInfo Google userinfo 端點回應
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// NewOAuthHandler 建立新的 OAuth 處理器
func NewOAuthHandler(config *OAuthConfig, keyManager *KeyManager, tokenStore *RefreshTokenStore) *OAuthHandler {
	handler := &OAuthHandler{
		config:     config,
		keyManager: keyManager,
		tokenStore: tokenStore,
		sessions:   make(map[string]*OAuthSession),
	}

	// 啟動 session 清理
	go handler.cleanupSessions()

	return handler
}

// cleanupSessions 定期清理過期的 session
func (h *OAuthHandler) cleanupSessions() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		h.mutex.Lock()
		now := time.Now()
		for state, session := range h.sessions {
			// Session 10 分鐘後過期
			if now.Sub(session.CreatedAt) > 10*time.Minute {
				delete(h.sessions, state)
			}
		}
		h.mutex.Unlock()
	}
}

// LoginHandler GET /login - 發起 OAuth 流程
func (h *OAuthHandler) LoginHandler(c *gin.Context) {
	origin := c.Query("origin")

	// 驗證 origin
	if origin == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing origin parameter",
			"message": "缺少 origin 參數",
		})
		return
	}

	if !ValidateOrigin(origin) {
		log.Printf("[OAuth] 拒絕不允許的 origin: %s", origin)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Origin not allowed",
			"message": fmt.Sprintf("不允許的來源: %s", origin),
		})
		return
	}

	// 生成 state
	state := uuid.New().String()

	// 儲存 session
	h.mutex.Lock()
	h.sessions[state] = &OAuthSession{
		State:     state,
		Origin:    origin,
		CreatedAt: time.Now(),
	}
	h.mutex.Unlock()

	// 構建 Google OAuth URL
	params := url.Values{
		"client_id":     {h.config.ClientID},
		"redirect_uri":  {h.config.RedirectURI},
		"response_type": {"code"},
		"scope":         {"openid email profile"},
		"state":         {state},
		"prompt":        {"select_account"},
	}

	googleAuthURL := "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()

	log.Printf("[OAuth] 導向 Google OAuth，origin: %s, state: %s...", origin, state[:8])

	c.Redirect(http.StatusFound, googleAuthURL)
}

// CallbackHandler GET /auth/google/callback - OAuth 回調
func (h *OAuthHandler) CallbackHandler(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	errorParam := c.Query("error")

	// 設定不快取
	c.Header("Cache-Control", "no-store")

	// 檢查 Google 是否回傳錯誤
	if errorParam != "" {
		log.Printf("[OAuth] Google 回傳錯誤: %s", errorParam)
		h.renderErrorPage(c, "登入取消", "您已取消 Google 登入，請關閉此視窗後重試。")
		return
	}

	// 驗證必要參數
	if code == "" || state == "" {
		log.Println("[OAuth] 缺少必要參數")
		h.renderErrorPage(c, "參數錯誤", "登入請求缺少必要參數，請重新登入。")
		return
	}

	// 驗證 state
	h.mutex.RLock()
	session, exists := h.sessions[state]
	h.mutex.RUnlock()

	if !exists {
		log.Printf("[OAuth] State 不存在或已過期: %s", state)
		h.renderErrorPage(c, "安全驗證失敗", "登入請求的安全驗證失敗（State 不匹配），這可能是 CSRF 攻擊或 Session 已過期。請重新登入。")
		return
	}

	origin := session.Origin

	// 刪除已使用的 session
	h.mutex.Lock()
	delete(h.sessions, state)
	h.mutex.Unlock()

	// 交換 code 為 access token
	tokenResp, err := h.exchangeCode(code)
	if err != nil {
		log.Printf("[OAuth] 交換 token 失敗: %v", err)
		h.renderErrorPage(c, "登入失敗", "與 Google 伺服器通訊時發生錯誤，請稍後重試。")
		return
	}

	// 取得使用者資訊
	userInfo, err := h.getUserInfo(tokenResp.AccessToken)
	if err != nil {
		log.Printf("[OAuth] 取得使用者資訊失敗: %v", err)
		h.renderErrorPage(c, "登入失敗", "無法取得使用者資訊，請稍後重試。")
		return
	}

	// 簽發 JWT
	claims := &AuthTokenClaims{
		Sub:     userInfo.ID,
		Name:    userInfo.Name,
		Picture: userInfo.Picture,
	}

	token, err := h.keyManager.SignToken(claims)
	if err != nil {
		log.Printf("[OAuth] 簽發 JWT 失敗: %v", err)
		h.renderErrorPage(c, "登入失敗", "無法產生認證 Token，請稍後重試。")
		return
	}

	log.Printf("[OAuth] 登入成功，使用者: %s (%s)", userInfo.Name, userInfo.ID)

	// 生成 Refresh Token 並儲存
	refreshTokenID := h.tokenStore.Store(userInfo.ID, userInfo.Name, userInfo.Picture)

	// 設定 Refresh Token Cookie
	h.setRefreshTokenCookie(c, refreshTokenID)

	// 導回 origin，access token 放在 URL hash
	redirectURL := fmt.Sprintf("%s#token=%s", origin, token)
	c.Redirect(http.StatusFound, redirectURL)
}

// setRefreshTokenCookie 設定 Refresh Token Cookie
func (h *OAuthHandler) setRefreshTokenCookie(c *gin.Context, refreshTokenID string) {
	isSecure := isSecureContext(c)
	sameSite := http.SameSiteLaxMode // 使用 Lax 以允許重導向時攜帶 Cookie

	c.SetSameSite(sameSite)
	c.SetCookie(
		RefreshTokenCookieName,
		refreshTokenID,
		86400, // 1 天 (秒)
		"/",   // 根路徑，確保所有請求都能攜帶
		"",
		isSecure,
		true, // HttpOnly
	)
}

// exchangeCode 用授權碼交換 access token
func (h *OAuthHandler) exchangeCode(code string) (*GoogleTokenResponse, error) {
	data := url.Values{
		"client_id":     {h.config.ClientID},
		"client_secret": {h.config.ClientSecret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {h.config.RedirectURI},
	}

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tokenResp GoogleTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("empty access token, response: %s", string(body))
	}

	return &tokenResp, nil
}

// getUserInfo 取得 Google 使用者資訊
func (h *OAuthHandler) getUserInfo(accessToken string) (*GoogleUserInfo, error) {
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo GoogleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	if userInfo.ID == "" {
		return nil, fmt.Errorf("empty user ID, response: %s", string(body))
	}

	return &userInfo, nil
}

// renderErrorPage 渲染錯誤頁面
func (h *OAuthHandler) renderErrorPage(c *gin.Context, title, message string) {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-TW">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>%s - 麻將遊戲認證</title>
  <style>
    body {
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
      background: linear-gradient(135deg, #1a1a2e 0%%, #16213e 100%%);
      color: #fff;
      min-height: 100vh;
      display: flex;
      align-items: center;
      justify-content: center;
      margin: 0;
      padding: 20px;
    }
    .error-container {
      background: rgba(255, 255, 255, 0.1);
      border-radius: 16px;
      padding: 40px;
      max-width: 400px;
      text-align: center;
      backdrop-filter: blur(10px);
    }
    .error-icon { font-size: 64px; margin-bottom: 20px; }
    h1 { margin: 0 0 16px 0; font-size: 24px; }
    p { margin: 0; color: rgba(255, 255, 255, 0.7); line-height: 1.6; }
    .back-link {
      display: inline-block;
      margin-top: 24px;
      color: #4fc3f7;
      text-decoration: none;
    }
    .back-link:hover { text-decoration: underline; }
  </style>
</head>
<body>
  <div class="error-container">
    <div class="error-icon">⚠️</div>
    <h1>%s</h1>
    <p>%s</p>
    <a href="javascript:window.close()" class="back-link">關閉視窗</a>
  </div>
</body>
</html>`, title, title, message)

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusBadRequest, html)
}
