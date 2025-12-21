# 台灣十六張麻將專案 README

[![Security Scan](https://github.com/tedmax100/mahjong/actions/workflows/security.yml/badge.svg)](https://github.com/tedmax100/mahjong/actions/workflows/security.yml)

這是一個功能完整的台灣十六張麻將網頁遊戲專案，具備前後端分離、即時連線對戰、以及可選的 AI 提示功能。

## ✨ 架構亮點

*   **微前端 (Micro-Frontends)**: 將使用者介面拆分為`大廳 (Lobby)`與`遊戲 (Game)`兩個獨立的前端應用，提高開發效率與可維護性。
*   **服務導向後端 (Service-Oriented Backend)**: 後端拆分為`Lobby 認證服務`與`遊戲邏輯服務`，讓系統各部分職責分明。
*   **即時遊戲體驗**: 使用 WebSocket 實現玩家間的即時出牌、吃、碰、槓、胡牌等操作。
*   **AI 輔助**: 整合 [Ollama](https://ollama.com/)，在本機運行大型語言模型 (LLM)，為玩家提供出牌建議。
*   **一鍵式開發環境**: 透過 `Makefile` 整合了所有服務的啟動、安裝、測試與建置流程，簡化了開發設定。
*   **容器化部署**: 提供 `Dockerfile`，可將遊戲服務打包成獨立的 Docker 映像檔，方便部署。

## 🛠️ 技術棧 (Tech Stack)

*   **後端 (Backend)**: Go
*   **前端 (Frontend)**:
    *   **遊戲端**: JavaScript + [PixiJS](https://pixijs.com/) (用於 2D 渲染)
    *   **大廳端**: 純 JavaScript
    *   **建置工具**: [Vite](https://vitejs.dev/)
*   **即時通訊 (Real-time)**: WebSocket
*   **AI 模型服務**: Ollama + Docker
*   **認證 (Authentication)**: Google OAuth 2.0 + JWT

## 📂 專案結構

```
/
├── server/          # Go 後端原始碼
│   ├── cmd/main.go      # 遊戲邏輯主服務 (WebSocket)
│   └── cmd/lobby/main.go  # Lobby 與認證代理服務
├── lobby-client/    # 大廳介面前端專案
├── game-client/     # 遊戲介面前端專案
├── game-bundle/     # 遊戲獨立部署包 (整合後端與 game-client)
├── docker-compose.yaml # AI 服務 (Ollama) 的設定
└── Makefile         # 自動化指令腳本
```

## 🚀 如何開始 (Getting Started)

推薦使用分離式前端的模式進行開發，此模式會完整啟動所有服務。

#### 1. 環境準備

請確保您已安裝以下軟體：
*   `make`
*   `go` (建議版本 1.20+)
*   `node` & `npm` (建議版本 18+)
*   `docker` & `docker-compose` (若要啟用 AI 功能)

#### 2. 設定環境變數

Lobby 伺服器需要 Google OAuth 的憑證。
```bash
cp server/.env.example server/.env
```
然後編輯 `server/.env` 檔案，填入您的 `GOOGLE_CLIENT_ID` 和 `GOOGLE_CLIENT_SECRET`。

#### 3. 安裝所有依賴

此指令會安裝後端、大廳前端、遊戲前端的所有依賴項目。
```bash
make install-all
```

#### 4. 啟動開發環境

*   **啟動所有服務（推薦）**
    此指令會同時啟動後端兩個服務、前端兩個開發伺服器。
    ```bash
    make dev-all
    ```
    啟動後，各服務的位址為：
    *   **大廳前端**: `http://localhost:5174` (由此進入遊戲)
    *   **遊戲前端**: `http://localhost:5175`
    *   **Lobby 後端**: `http://localhost:3001`
    *   **遊戲後端**: `http://localhost:8080`

*   **啟動 AI 建議功能 (可選)**
    ```bash
    docker compose up -d
    ```
    第一次啟動後，需要手動下載模型：
    ```bash
    docker exec -it qwen_server ollama pull gemma:2b
    ```

#### 5. 停止所有服務

```bash
make stop
```

## 📦 如何建置與部署

本專案支援將`遊戲服務`與`遊戲前端`打包成一個獨立的應用程式進行部署。

#### 1. 建立部署包

此指令會建置 `game-client`，並將其與編譯後的 `mahjong-server` 執行檔一起放入 `game-bundle/dist` 目錄。
```bash
make build-game-bundle
```

#### 2. 使用 Docker 進行部署 (推薦)

專案已預設 `game-bundle/Dockerfile`，可直接用來建置與運行遊戲服務。
```bash
# 步驟 1: 建置 Docker 映像
make docker-build-game

# 步驟 2: 運行 Docker 容器
make docker-run-game
```
服務將會運行在 `http://localhost:8080`。

## 🧪 如何測試

目前專案提供針對 `game-client` 的單元測試。
```bash
# 運行所有測試
make test

# 啟動互動式測試介面
make test-ui
```

## 📜 可用指令

`Makefile` 提供了豐富的指令來簡化開發流程，您可以執行 `make help` 來查看所有可用的指令與其說明。
```bash
make help
```
