package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// KeyPair 代表一組 RSA 金鑰對
type KeyPair struct {
	KID        string
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
	CreatedAt  time.Time
}

// KeyManager 管理 RSA 金鑰輪替
type KeyManager struct {
	currentKey  *KeyPair
	previousKey *KeyPair
	mutex       sync.RWMutex
	stopChan    chan struct{}
	issuer      string
	expiration  time.Duration
}

// JWK JSON Web Key 結構
type JWK struct {
	KTY string `json:"kty"`
	Use string `json:"use"`
	ALG string `json:"alg"`
	KID string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKS JSON Web Key Set 結構
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// NewKeyManager 建立新的金鑰管理器
func NewKeyManager(issuer string, expiration time.Duration) *KeyManager {
	return &KeyManager{
		issuer:     issuer,
		expiration: expiration,
		stopChan:   make(chan struct{}),
	}
}

// Start 初始化並啟動金鑰輪替
func (km *KeyManager) Start(rotationInterval time.Duration) error {
	// 生成初始金鑰
	if err := km.rotateKeys(); err != nil {
		return err
	}

	// 啟動定期輪替
	go func() {
		ticker := time.NewTicker(rotationInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := km.rotateKeys(); err != nil {
					log.Printf("[KeyManager] 金鑰輪替失敗: %v", err)
				}
			case <-km.stopChan:
				log.Println("[KeyManager] 金鑰輪替已停止")
				return
			}
		}
	}()

	log.Printf("[KeyManager] 金鑰管理器已啟動，輪替間隔: %v", rotationInterval)
	return nil
}

// Stop 停止金鑰輪替
func (km *KeyManager) Stop() {
	close(km.stopChan)
}

// rotateKeys 執行金鑰輪替
func (km *KeyManager) rotateKeys() error {
	// 生成新的 RSA 金鑰對
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	newKey := &KeyPair{
		KID:        uuid.New().String(),
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
		CreatedAt:  time.Now(),
	}

	km.mutex.Lock()
	km.previousKey = km.currentKey
	km.currentKey = newKey
	km.mutex.Unlock()

	log.Printf("[KeyManager] 金鑰已輪替，新 KID: %s", newKey.KID)
	return nil
}

// GetJWKS 取得 JWKS 格式的公鑰列表
func (km *KeyManager) GetJWKS() *JWKS {
	km.mutex.RLock()
	defer km.mutex.RUnlock()

	keys := []JWK{}

	if km.currentKey != nil {
		keys = append(keys, km.publicKeyToJWK(km.currentKey))
	}

	if km.previousKey != nil {
		keys = append(keys, km.publicKeyToJWK(km.previousKey))
	}

	return &JWKS{Keys: keys}
}

// publicKeyToJWK 將公鑰轉換為 JWK 格式
func (km *KeyManager) publicKeyToJWK(kp *KeyPair) JWK {
	return JWK{
		KTY: "RSA",
		Use: "sig",
		ALG: "RS256",
		KID: kp.KID,
		N:   base64.RawURLEncoding.EncodeToString(kp.PublicKey.N.Bytes()),
		E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(kp.PublicKey.E)).Bytes()),
	}
}

// SignToken 使用當前私鑰簽發 JWT
func (km *KeyManager) SignToken(claims *AuthTokenClaims) (string, error) {
	km.mutex.RLock()
	currentKey := km.currentKey
	km.mutex.RUnlock()

	if currentKey == nil {
		return "", jwt.ErrTokenNotValidYet
	}

	// 設定標準 claims
	now := time.Now()
	claims.RegisteredClaims = jwt.RegisteredClaims{
		Issuer:    km.issuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(km.expiration)),
	}

	// 建立 token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = currentKey.KID

	// 簽名
	return token.SignedString(currentKey.PrivateKey)
}

// AuthTokenClaims JWT 中的使用者資訊（用於簽發）
type AuthTokenClaims struct {
	Sub     string `json:"sub"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
	jwt.RegisteredClaims
}

// GetJWKSJSON 取得 JSON 格式的 JWKS
func (km *KeyManager) GetJWKSJSON() ([]byte, error) {
	return json.Marshal(km.GetJWKS())
}
