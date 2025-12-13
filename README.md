# 麻將專案 README

[![Security Scan](https://github.com/tedmax100/mahjong/actions/workflows/security.yml/badge.svg)](https://github.com/tedmax100/mahjong/actions/workflows/security.yml)
![Go Coverage](https://img.shields.io/badge/Go_Coverage-50.1%25-orange)
![Frontend Coverage](https://img.shields.io/badge/Frontend_Coverage-18.9%25-red)

這是一個16張台灣麻將的網頁遊戲專案。

## ✨ 功能亮點

*   **前後端分離**: 採用 Go 語言後端與現代 JavaScript 前端分離的架構。
*   **即時通訊**: 使用 WebSocket 技術實現玩家之間的即時互動。
*   **一鍵啟動**: 透過 `Makefile` 整合了完整的開發、測試、建置與部署流程。
*   **遠端測試**: 內建 Cloudflare Tunnel 整合，方便將本機服務暴露於公網進行測試。

## 🛠️ 技術棧 (Tech Stack)

*   **後端 (Backend)**: Go
*   **前端 (Frontend)**: JavaScript (使用 [Vite](https://vitejs.dev/) 作為建置工具)
*   **即時通訊 (Real-time Communication)**: WebSocket

## 🚀 如何開始 (Getting Started)

#### 1. 環境準備

請確保您已安裝以下軟體：

*   `make`
*   `Go` (建議版本 1.20+)
*   `Node.js` (建議版本 18+) 及 `npm`
*   (選用) `cloudflared` - 若要使用 `make start` 功能，會自動為您安裝。

執行 LLM
```bash
docker compose up -d

## 選用gemma3 270m
docker exec -it qwen_server ollama run gemma3:270m
```

#### 2. 安裝依賴

進入專案根目錄，執行以下指令來安裝前後端的所有依賴項目：

```bash
make install
```

#### 3. 運行開發環境

您可以根據需求選擇不同的啟動方式：

*   **純本機開發** (推薦日常開發使用)
    此模式會在本機啟動後端伺服器 (`:8080`) 與前端開發伺服器 (`:5173`)。

    ```bash
    make dev
    ```

*   **本機開發並建立公網通道**
    此模式除了啟動本機服務外，還會使用 Cloudflare Tunnel 建立一個臨時的公開網址，讓其他人可以存取您的本機前端服務。

    ```bash
    make start
    ```

#### 4. 停止所有服務

若要停止所有由 `make` 啟動的服務 (包含後端、前端與 tunnel)，請執行：

```bash
make stop
```

## 🧪 如何測試 (Running Tests)

使用以下指令來運行前端的單元測試：

```bash
make test
```

您也可以使用 Vitest 的 UI 模式來進行互動式測試：

```bash
make test-ui
```

## 📦 如何建置 (Building for Production)

執行以下指令來建置用於生產環境的前後端應用程式：

```bash
make build
```

*   後端應用程式會被編譯成執行檔 `mahjong-server` 並放置於專案根目錄。
*   前端的靜態檔案會被建置到 `client/dist` 目錄下。

## 📜 可用指令

本專案 `Makefile` 中包含了許多便利的指令，您可以執行 `make help` 來查看所有可用的指令及其說明。

```bash
make help
```
