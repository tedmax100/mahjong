# 台灣16張麻將 - 獨立部署包

這是麻將遊戲的獨立部署包，可以在自己的電腦或伺服器上運行。

## 快速開始

### 使用 Docker Compose（推薦）

```bash
# 在 game-bundle 目錄下
docker-compose up -d
```

遊戲將在 http://localhost:8080 啟動。

### 使用 Docker

```bash
# 在專案根目錄下
docker build -f game-bundle/Dockerfile -t mahjong-game .
docker run -d -p 8080:8080 mahjong-game
```

## 配置選項

### 環境變數

| 變數 | 說明 | 預設值 |
|------|------|--------|
| `PORT` | 伺服器埠號 | 8080 |
| `STANDALONE_MODE` | 獨立模式 | true |
| `STATIC_DIR` | 靜態檔案目錄 | ./public |

### 向中央大廳註冊（可選）

如果你想讓你的房間顯示在中央大廳，需要設定以下環境變數：

```bash
LOBBY_SERVICE_URL=https://lobby.example.com
LOBBY_INTERNAL_SECRET=your-secret
EXTERNAL_SERVER_ID=my-game-server
EXTERNAL_SERVER_SECRET=your-external-secret
EXTERNAL_SERVER_WEB_URL=https://your-game-server.example.com
```

## 使用方式

1. 開啟瀏覽器訪問 http://localhost:8080
2. 輸入名字或使用訪客模式
3. 創建房間或加入現有房間
4. 分享房間連結給朋友

## 技術規格

- 最低螢幕解析度：1280x720
- 支援的瀏覽器：Chrome、Firefox、Safari、Edge
- 需要 WebSocket 連接

## 常見問題

### Q: 如何讓其他人連接到我的伺服器？

1. 確保你的防火牆允許 8080 埠的入站連接
2. 取得你的公開 IP 或使用內網穿透工具（如 ngrok、Cloudflare Tunnel）
3. 分享連結：`http://你的IP:8080`

### Q: 為什麼無法連接？

- 檢查防火牆設定
- 確認 Docker 容器正在運行：`docker ps`
- 查看日誌：`docker-compose logs`
