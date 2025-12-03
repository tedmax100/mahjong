# Cloudflare Tunnel 設置指南

## 概述
使用 Cloudflare Tunnel 將本地開發服務器暴露到互聯網，無需配置防火牆或 NAT 穿透。

## 快速開始

### 方法 1: 使用啟動腳本（推薦）
```bash
./start-tunnel.sh
```

這個腳本會：
1. ✅ 自動啟動後端服務器 (Port 8080)
2. ✅ 自動啟動前端開發服務器 (Port 5173)
3. ✅ 建立 Cloudflare Tunnel
4. ✅ 自動提取並顯示公網 URL
5. ✅ 在瀏覽器中自動打開 URL

### 方法 2: 使用 Makefile

#### 啟動所有服務 + Tunnel
```bash
make dev-tunnel
```

#### 快速啟動（推薦）
```bash
make tunnel-quick
```

#### 僅啟動本地開發服務器（不使用 tunnel）
```bash
make dev
```

#### 僅為後端建立 tunnel
```bash
make tunnel-backend
```

## 可用的命令

### Makefile 命令
```bash
make help                # 顯示所有可用命令
make install-cloudflared # 安裝 cloudflared
make dev                 # 啟動本地開發（無 tunnel）
make dev-tunnel          # 啟動開發 + Cloudflare Tunnel
make tunnel-quick        # 快速啟動 tunnel
make stop                # 停止所有服務
make clean               # 清理並停止所有服務
make status              # 檢查服務狀態
make logs                # 查看服務器日誌
make test                # 運行測試
make build               # 構建項目
make install             # 安裝所有依賴
```

### 直接命令
```bash
# 啟動 tunnel
./start-tunnel.sh

# 停止所有服務
make stop

# 檢查狀態
make status
```

## 工作原理

### 架構
```
Internet (公網)
    ↓
Cloudflare Tunnel (https://xxx.trycloudflare.com)
    ↓
本地前端開發服務器 (http://localhost:5173)
    ↓ (API 請求)
本地後端服務器 (http://localhost:8080)
```

### Cloudflare Tunnel 是什麼？
Cloudflare Tunnel 提供：
- 🔒 **安全**: 通過 Cloudflare 的加密隧道連接
- 🌐 **公網訪問**: 無需公網 IP 或端口轉發
- 🚀 **快速**: 利用 Cloudflare 的全球 CDN 網絡
- 🆓 **免費**: 使用 TryCloudflare 免費服務

## 配置說明

### 端口配置
- **前端**: `5173` (Vite 默認端口)
- **後端**: `8080` (Go 服務器)

### 文件說明
- `Makefile` - 包含所有開發命令
- `start-tunnel.sh` - 自動啟動腳本
- `cloudflare-tunnel.yml` - Tunnel 配置文件（自動生成）
- `.tunnel-url` - 存儲當前 Tunnel URL
- `tunnel.log` - Tunnel 日誌
- `server.log` - 後端服務器日誌
- `frontend.log` - 前端服務器日誌

## 常見問題

### Q: 如何安裝 cloudflared？
```bash
make install-cloudflared
```

或手動安裝：
- **Linux (Debian/Ubuntu)**:
  ```bash
  wget https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
  sudo dpkg -i cloudflared-linux-amd64.deb
  ```
- **macOS**:
  ```bash
  brew install cloudflared
  ```
- **其他系統**: 請訪問 https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/install-and-setup/installation/

### Q: Tunnel URL 每次都會變嗎？
是的，使用 TryCloudflare 的快速 Tunnel 每次啟動都會生成新的隨機 URL。如果需要固定 URL，可以註冊 Cloudflare Zero Trust 並創建命名的 Tunnel。

### Q: 如何查看日誌？
```bash
# 查看所有日誌
tail -f server.log frontend.log tunnel.log

# 或使用 make 命令
make logs
```

### Q: 如何停止所有服務？
```bash
# 在啟動腳本運行時按 Ctrl+C

# 或使用 make 命令
make stop
```

### Q: 如何檢查服務是否運行？
```bash
make status
```

### Q: 如何在遠程服務器上運行？
1. 確保遠程服務器已安裝 Go、Node.js 和 cloudflared
2. 克隆項目到遠程服務器
3. 運行 `./start-tunnel.sh`
4. 使用顯示的 Cloudflare URL 訪問

### Q: 如何分享給團隊成員？
直接分享 Cloudflare Tunnel URL 給團隊成員，他們可以通過這個 URL 訪問你的本地開發服務器。

**注意**: TryCloudflare 的 URL 是臨時的，每次重啟都會改變。

### Q: 性能如何？
Cloudflare Tunnel 會增加一些延遲（通常 50-200ms），但對於開發和測試來說完全足夠。生產環境建議使用正式的部署方案。

## 安全建議

⚠️ **重要**: Cloudflare Tunnel 會將你的本地服務器暴露到公網：

1. **僅用於開發/測試**: 不要在生產環境中使用 TryCloudflare
2. **保護敏感數據**: 確保不要暴露包含敏感信息的 API 或數據庫
3. **短期使用**: 測試完成後記得停止 Tunnel
4. **不要分享私密信息**: 通過 Tunnel URL 分享時要注意安全
5. **使用認證**: 考慮為你的應用添加基本認證

## 進階配置

### 使用配置文件
創建 `cloudflare-tunnel.yml`:
```yaml
url: http://localhost:5173
ingress:
  - hostname: "*"
    service: http://localhost:5173
  - service: http_status:404
```

然後啟動：
```bash
cloudflared tunnel --config cloudflare-tunnel.yml
```

### 多服務 Tunnel
如果你想同時暴露前端和後端，可以使用 Ingress 規則：
```yaml
ingress:
  - hostname: api.xxx.trycloudflare.com
    service: http://localhost:8080
  - hostname: "*"
    service: http://localhost:5173
  - service: http_status:404
```

### 固定 URL
要獲得固定的 URL，需要：
1. 註冊 Cloudflare Zero Trust 賬號（免費）
2. 創建命名的 Tunnel
3. 使用 `cloudflared login` 登錄
4. 創建配置文件並指定 Tunnel 名稱

詳見: https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/

## 故障排除

### Tunnel 啟動失敗
```bash
# 檢查日誌
cat tunnel.log

# 檢查 cloudflared 是否安裝
cloudflared --version

# 重新安裝
make install-cloudflared
```

### 無法訪問服務
```bash
# 檢查服務狀態
make status

# 檢查端口是否被占用
lsof -i :5173
lsof -i :8080

# 重啟所有服務
make stop
./start-tunnel.sh
```

### URL 未顯示
```bash
# 檢查 tunnel 日誌
cat tunnel.log

# 手動啟動 tunnel
cloudflared tunnel --url http://localhost:5173
```

## 參考資源

- [Cloudflare Tunnel 官方文檔](https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/)
- [TryCloudflare 快速開始](https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/do-more-with-tunnels/trycloudflare/)
- [Cloudflared GitHub](https://github.com/cloudflare/cloudflared)

## 支持

如有問題，請：
1. 查看日誌文件: `server.log`, `frontend.log`, `tunnel.log`
2. 運行 `make status` 檢查服務狀態
3. 查看故障排除部分
4. 訪問 Cloudflare 官方文檔
