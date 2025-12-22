# Zeabur 部署指南

本專案包含兩個需要部署的服務：

| 服務 | 說明 | Port | Dockerfile |
|------|------|------|------------|
| **Lobby** | 大廳 + Auth Proxy + 前端 | 3001 | `lobby-bundle/Dockerfile` |
| **Game** | 遊戲伺服器 + 前端 | 8080 | `game-bundle/Dockerfile` |

## 步驟 1：建立 Zeabur 專案

1. 登入 [Zeabur Console](https://dash.zeabur.com)
2. 點擊「New Project」建立新專案
3. 選擇你的 Git 供應商（GitHub/GitLab）並授權
4. 選擇這個 repository

## 步驟 2：部署 Lobby 服務

1. 在專案中點擊「Add Service」→「Git」
2. 選擇這個 repository
3. **重要**：在「Root Directory」填入 `.`（專案根目錄）
4. **重要**：在「Dockerfile Path」填入 `lobby-bundle/Dockerfile`
5. 設定服務名稱為 `mahjong-lobby`

### Lobby 環境變數

在 Zeabur 的「Variables」頁面設定以下環境變數：

| 變數名稱 | 說明 | 範例值 |
|----------|------|--------|
| `PORT` | 服務埠號 | `3001`（Zeabur 會自動設定） |
| `GIN_MODE` | Gin 運行模式 | `release` |
| `AUTH_PROXY_URL` | 認證服務公開 URL | `https://lobby.你的網域.zeabur.app` |
| `GAME_SERVER_URL` | 遊戲伺服器內部 URL | `https://game.你的網域.zeabur.app` |
| `GAME_CLIENT_URL` | 遊戲前端公開 URL | `https://game.你的網域.zeabur.app` |
| `GOOGLE_CLIENT_ID` | Google OAuth Client ID | `xxx.apps.googleusercontent.com` |
| `GOOGLE_CLIENT_SECRET` | Google OAuth Secret | `GOCSPX-xxx` |
| `SESSION_SECRET` | Session 加密金鑰 | 使用 `openssl rand -hex 32` 生成 |
| `LOBBY_INTERNAL_SECRET` | 內部通訊密鑰 | 使用 `openssl rand -hex 32` 生成 |
| `EXTERNAL_SERVER_SECRET` | 外部伺服器密鑰 | 使用 `openssl rand -hex 32` 生成 |

### 設定 Lobby 網域

1. 在服務的「Networking」頁面
2. 點擊「Generate Domain」或設定自訂網域
3. 記下這個 URL，需要設定到 `AUTH_PROXY_URL`

## 步驟 3：部署 Game 服務

1. 點擊「Add Service」→「Git」
2. 選擇同一個 repository
3. **重要**：在「Root Directory」填入 `.`（專案根目錄）
4. **重要**：在「Dockerfile Path」填入 `game-bundle/Dockerfile`
5. 設定服務名稱為 `mahjong-game`

### Game 環境變數

| 變數名稱 | 說明 | 範例值 |
|----------|------|--------|
| `PORT` | 服務埠號 | `8080`（Zeabur 會自動設定） |
| `GIN_MODE` | Gin 運行模式 | `release` |
| `AUTH_PROXY_URL` | Lobby 服務 URL（用於驗證 JWT） | `https://lobby.你的網域.zeabur.app` |
| `LOBBY_SERVICE_URL` | Lobby 服務 URL（用於回報房間狀態） | `https://lobby.你的網域.zeabur.app` |
| `LOBBY_INTERNAL_SECRET` | 內部通訊密鑰（需與 Lobby 一致） | 與上面相同 |

### 設定 Game 網域

1. 在服務的「Networking」頁面
2. 點擊「Generate Domain」或設定自訂網域
3. 記下這個 URL，需要設定到 Lobby 的 `GAME_SERVER_URL` 和 `GAME_CLIENT_URL`

## 步驟 4：更新 Google OAuth 設定

1. 前往 [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
2. 編輯你的 OAuth 2.0 Client
3. 在「Authorized redirect URIs」新增：
   ```
   https://lobby.你的網域.zeabur.app/auth/google/callback
   ```

## 步驟 5：驗證部署

1. 訪問 Lobby 服務的健康檢查端點：
   ```
   https://lobby.你的網域.zeabur.app/health
   ```

2. 訪問 Game 服務的健康檢查端點：
   ```
   https://game.你的網域.zeabur.app/health
   ```

3. 訪問大廳首頁開始遊戲：
   ```
   https://lobby.你的網域.zeabur.app
   ```

## 架構說明

```
┌─────────────────────────────────────────────────────────────┐
│                        使用者瀏覽器                           │
└─────────────────────────────────────────────────────────────┘
                    │                           │
                    ▼                           ▼
┌─────────────────────────────┐    ┌─────────────────────────────┐
│     Lobby Service           │    │     Game Service            │
│  lobby.xxx.zeabur.app       │    │  game.xxx.zeabur.app        │
├─────────────────────────────┤    ├─────────────────────────────┤
│  - 大廳前端 (Vue/React)      │    │  - 遊戲前端 (PixiJS)         │
│  - Auth Proxy (OAuth)       │    │  - 遊戲後端 (WebSocket)      │
│  - 大廳 API & WebSocket     │    │  - JWT 驗證                  │
│  - JWKS 公鑰端點            │    │                             │
└─────────────────────────────┘    └─────────────────────────────┘
         │                                     │
         └──────── 內部 API 通訊 ───────────────┘
              (LOBBY_INTERNAL_SECRET)
```

## 本地測試 Docker 構建

```bash
# 測試 Lobby bundle
cd lobby-bundle
docker compose up --build

# 測試 Game bundle
cd game-bundle
docker compose up --build
```

## 常見問題

### Q: 部署後無法登入？
A: 檢查 Google OAuth 的 redirect URI 是否正確設定，並確保 `AUTH_PROXY_URL` 與實際網域一致。

### Q: 遊戲無法連接？
A: 確認 `GAME_SERVER_URL` 和 `GAME_CLIENT_URL` 設定正確，並且 `LOBBY_INTERNAL_SECRET` 在兩個服務中一致。

### Q: WebSocket 連接失敗？
A: Zeabur 預設支持 WebSocket，確保你使用的是 `wss://` 協定（HTTPS 網域會自動升級）。
