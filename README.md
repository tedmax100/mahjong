# 台灣十六張麻將專案 README

[![Security Scan](https://github.com/tedmax100/mahjong/actions/workflows/security.yml/badge.svg)](https://github.com/tedmax100/mahjong/actions/workflows/security.yml)
![Go Coverage](https://img.shields.io/badge/Go_Coverage-34.5%25-red)
![Frontend Coverage](https://img.shields.io/badge/Frontend_Coverage-14.08%25-red)
![Game Version](https://img.shields.io/badge/Game-v0.0.8-blue)
![Lobby Version](https://img.shields.io/badge/Lobby-v0.0.10-blue)

這是一個功能完整的台灣十六張麻將網頁遊戲專案，旨在提供玩家一個即時、互動的線上麻將體驗，並可選擇 AI 輔助來提升遊戲策略。

## 🎮 專案特色與功能

*   **即時線上對戰**: 玩家可以透過網路與朋友或其他玩家進行即時的台灣十六張麻將對局。
*   **靈活的遊戲大廳**: 方便建立私人房間或加入公開對局。
*   **AI 輔助決策 (可選)**: 整合大型語言模型 (LLM)，為您提供出牌建議，幫助您學習和提升麻將技巧。
*   **跨平台網頁體驗**: 無需安裝，在任何支援現代瀏覽器的裝置上都能遊玩。

## 🗺️ 系統架構概覽

本專案採用了模組化的架構設計，將不同的功能拆分為獨立的服務，以提升可擴展性和維護性。

*   **前端應用**:
    *   **大廳客戶端 (`lobby-client`)**: 提供使用者登入、註冊、管理遊戲房間（建立/加入）的介面。它是您進入遊戲的門戶。
    *   **遊戲客戶端 (`game-client`)**: 負責呈現實際的麻將牌桌、玩家手牌、出牌區等所有遊戲視覺和互動邏輯。
*   **後端服務**:
    *   **Lobby/認證代理伺服器 (`mahjong-lobby`)**: 處理使用者身份驗證 (透過 Google OAuth)、管理使用者會話、以及協調遊戲房間的創建與狀態。
    *   **麻將遊戲伺服器 (`mahjong-server`)**: 透過 WebSocket 技術，處理所有即時的麻將遊戲邏輯，包括牌局初始化、牌面交換、玩家動作驗證、勝負判定等。
*   **AI 輔助服務 (可選)**:
    *   透過 Docker 運行一個本地大型語言模型 (`Ollama`)，為遊戲客戶端提供智慧出牌建議。

## ⚡ 核心技術一覽

*   **後端**: Go 語言
*   **前端**: Vite, PixiJS (遊戲渲染), 純 JavaScript (大廳)
*   **即時通訊**: WebSocket
*   **認證**: Google OAuth 2.0
*   **AI**: Ollama

## 🔐 認證架構：Auth Proxy 與 JWKS

本專案採用了**認證代理 (Auth Proxy)** 模式來處理使用者登入，旨在解決一個核心挑戰：如何在一個允許使用者自行託管（產生**動態、臨時網址**）的系統中，整合一個需要**固定、預先註冊網址**的第三方認證服務（如 Google OAuth）。

#### 為何需要 Auth Proxy？

*   **集中處理認證**：所有與 Google OAuth 的複雜互動都由`mahjong-lobby`服務（即 Auth Proxy）統一處理。此代理伺服器部署在一個固定的公開網址上。
*   **簡化其他服務**：遊戲主客戶端 (`game-client`) 和遊戲伺服器 (`mahjong-server`) 無需知道任何 Google OAuth 的實作細節。它們只認得由 Auth Proxy 簽發的內部 JWT (JSON Web Token)。
*   **解決動態網址問題**：當使用者從一個臨時的遊戲網址（例如 `https://random-name.trycloudflare.com`）點擊登入時，他們會被導向固定的 Auth Proxy。完成認證後，Auth Proxy 會再將他們安全地導回原來的臨時網址，並附上一個認證權杖 (JWT)。

#### JWKS 如何保障安全？

*   **非對稱金鑰驗證**：Auth Proxy 使用一對**私鑰/公鑰**來簽發 JWT。私鑰被嚴格保管在 Auth Proxy 內部，而公鑰則透過一個公開的 **JWKS (JSON Web Key Set) 端點**提供。
*   **去中心化驗證**：當遊戲伺服器 (`mahjong-server`) 收到來自客戶端的請求時，它會從請求中提取 JWT。然後，它會從 Auth Proxy 的 JWKS 端點獲取公鑰，用以驗證該 JWT 的簽名是否有效。
*   **提升安全性與靈活性**：這種機制意味著遊戲伺服器無需存取或共享任何敏感的私鑰。它只需要知道去哪裡獲取公鑰即可驗證權杖的真偽，確保了服務之間的信任關係，同時也方便了金鑰的輪換與管理。

總結來說，Auth Proxy + JWKS 的架構不僅解決了動態託管環境下的認證難題，還建立了一套標準化、安全且可擴展的微服務認證流程。

## 🚀 快速開始：如何啟動開發環境

本節將指導您如何快速設定並啟動所有必要的服務，以便進行開發或體驗完整的專案功能。

#### 1. 前置條件

請確保您的系統已安裝以下軟體：
*   `make` (用於自動化指令)
*   `Go` (版本 1.20 或更高)
*   `Node.js` & `npm` (版本 18 或更高)
*   `Docker` & `Docker Compose` (若要啟用 AI 輔助功能)

#### 2. 設定環境變數

Lobby 伺服器需要 Google OAuth 憑證才能運作。
```bash
cp server/.env.example server/.env
```
然後請編輯 `server/.env` 檔案，填入您的 `GOOGLE_CLIENT_ID` 和 `GOOGLE_CLIENT_SECRET`。

#### 3. 安裝所有依賴

執行此命令將自動安裝所有後端 Go 模組以及前端應用 (大廳客戶端與遊戲客戶端) 的 npm 依賴。
```bash
make install-all
```

#### 4. 啟動所有服務

這是啟動完整開發環境的推薦方式，它將同時啟動兩個後端服務和兩個前端應用。
```bash
make dev-all
```
當所有服務成功啟動後，您可以透過以下位址存取：
*   **遊戲大廳入口**: `http://localhost:5174` (請由此進入，並透過 Google 帳號登入)
*   **遊戲客戶端 (直接訪問，通常由大廳引導)**: `http://localhost:5175`
*   **Lobby/認證後端**: `http://localhost:3001`
*   **麻將遊戲後端**: `http://localhost:8080`

#### 5. 啟動 AI 輔助功能 (可選)

如果您想在遊戲中體驗 AI 出牌建議，請執行以下步驟：
```bash
# 啟動 Ollama 容器
docker compose up -d

# 第一次運行需下載麻將AI模型 (以 Gemma 2B 為例)
docker exec -it qwen_server ollama pull gemma:2b
```

#### 6. 停止所有服務

若要停止所有由 `make dev-all` 或其他 `make` 指令啟動的服務，請執行：
```bash
make stop
```

## 📜 主要指令一覽

以下是專案中常用的 `Makefile` 指令，幫助您更有效地開發與管理專案：

*   `make install-all`: 安裝所有後端與前端的專案依賴。
*   `make dev-all`: 啟動完整的開發環境 (兩個後端服務 + 兩個前端應用)。
*   `make stop`: 停止所有運行中的專案服務。
*   `make test`: 運行前端遊戲客戶端的單元測試。
*   `make test-ui`: 啟動互動式前端測試介面。
*   `make build-game-bundle`: 將遊戲後端與遊戲前端打包成一個獨立的部署包。
*   `make docker-build-game`: 建置遊戲服務的 Docker 映像檔。
*   `make docker-run-game`: 運行遊戲服務的 Docker 容器。
*   `make help`: 顯示所有可用的 `Makefile` 指令及其詳細說明。

## 📦 版本管理與發布

### 版本管理指令

| 指令 | 說明 |
|------|------|
| `make version` | 顯示當前 Game 與 Lobby 版本 |
| `make version-bump-game V=v0.0.9` | 更新 Game 版本 |
| `make version-bump-lobby V=v0.0.11` | 更新 Lobby 版本 |
| `make version-bump-all GAME_V=v0.0.9 LOBBY_V=v0.0.11` | 同時更新兩者版本 |

### Docker 版本化推送

| 指令 | 說明 |
|------|------|
| `make docker-push-game-versioned` | 推送 Game 映像（含版本 tag + latest） |
| `make docker-push-lobby-versioned` | 推送 Lobby 映像（含版本 tag + latest） |
| `make docker-push-all-versioned` | 推送所有映像（含版本 tag + latest） |

### GitHub Release

| 指令 | 說明 |
|------|------|
| `make release-game` | 使用當前版本創建 Game Release |
| `make release-game V=v0.0.9` | 指定版本創建 Game Release |
| `make release-lobby` | 使用當前版本創建 Lobby Release |
| `make release-lobby V=v0.0.11` | 指定版本創建 Lobby Release |
| `make release-all` | 創建兩個 Release |

### 完整部署流程（一鍵發布）

| 指令 | 說明 |
|------|------|
| `make deploy-game V=v0.0.9` | 版本更新 → Docker 推送 → GitHub Release |
| `make deploy-lobby V=v0.0.11` | 版本更新 → Docker 推送 → GitHub Release |
| `make deploy-all GAME_V=v0.0.9 LOBBY_V=v0.0.11` | 完整部署所有服務 |

### GitHub Actions 自動化 CD

專案已配置 GitHub Actions 工作流程 (`.github/workflows/release.yml`)，支援：

**觸發方式：**
1. **Tag 推送觸發** - 推送 `game-v*` 或 `lobby-v*` tag 時自動執行
   ```bash
   git tag -a game-v0.0.9 -m "Game v0.0.9"
   git push origin game-v0.0.9
   ```
2. **手動觸發** - 在 GitHub Actions 頁面選擇 service (game/lobby/all) 和版本號

**自動執行流程：**
- Cross-compile Go binary (linux/amd64)
- Build Docker image
- Push to DockerHub（版本 tag + latest）
- 創建 GitHub Release

**所需 Secrets 設定：**
- `DOCKERHUB_USERNAME` - DockerHub 使用者名稱
- `DOCKERHUB_TOKEN` - DockerHub Access Token

**所需 Variables 設定（給 Lobby）：**
- `VITE_AUTH_PROXY_URL` - Auth Proxy URL
- `VITE_GAME_SERVER_URL` - Game Server URL

## 部署說明

本專案支援將遊戲服務打包為一個可獨立運行的部署包或 Docker 容器。詳細步驟請參考 `make build-game-bundle` 及 `make docker-build-game` 相關指令。

### Docker Hub 映像

| 服務 | 映像 | 版本 |
|------|------|------|
| Game Server | `tedmax100/mahjong-game` | `v0.0.8` |
| Lobby Server | `tedmax100/mahjong-lobby` | `v0.0.10` |

```bash
# 拉取最新版本
docker pull tedmax100/mahjong-game:latest
docker pull tedmax100/mahjong-lobby:latest

# 或指定版本
docker pull tedmax100/mahjong-game:v0.0.8
docker pull tedmax100/mahjong-lobby:v0.0.10
```
