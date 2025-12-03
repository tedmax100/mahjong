#!/bin/bash

# 顏色定義
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 找到專案根目錄
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR" || exit 1

# 驗證目錄結構
if [ ! -d "server" ] || [ ! -d "client" ]; then
    echo -e "${RED}✗ 錯誤: 找不到 server 或 client 目錄${NC}"
    echo -e "${YELLOW}請確保在專案根目錄執行此腳本${NC}"
    echo -e "${YELLOW}當前目錄: $(pwd)${NC}"
    exit 1
fi

# 配置
BACKEND_PORT=8080
FRONTEND_PORT=5173
TUNNEL_URL_FILE=".tunnel-url"
LOG_FILE="tunnel.log"

echo -e "${BLUE}======================================${NC}"
echo -e "${BLUE}  麻將遊戲 - Cloudflare Tunnel 啟動${NC}"
echo -e "${BLUE}======================================${NC}"
echo -e "${YELLOW}專案目錄: $SCRIPT_DIR${NC}"
echo ""

# 清理舊的 URL 文件
rm -f "$TUNNEL_URL_FILE"

# 清理函數 - 當腳本退出時執行
cleanup() {
    echo ""
    echo -e "${YELLOW}正在停止所有服務...${NC}"
    kill $BACKEND_PID 2>/dev/null
    kill $FRONTEND_PID 2>/dev/null
    kill $TUNNEL_PID 2>/dev/null
    pkill -P $$ 2>/dev/null
    rm -f "$TUNNEL_URL_FILE"
    echo -e "${GREEN}✓ 所有服務已停止${NC}"
    exit 0
}

# 設置 trap 來捕獲 Ctrl+C
trap cleanup SIGINT SIGTERM

# 檢查 cloudflared 是否安裝
if ! command -v cloudflared &> /dev/null; then
    echo -e "${RED}✗ cloudflared 未安裝${NC}"
    echo -e "${YELLOW}請運行: make install-cloudflared${NC}"
    exit 1
fi

# 啟動後端服務器
echo -e "${YELLOW}🚀 啟動後端服務器 (Port $BACKEND_PORT)...${NC}"
(cd "$SCRIPT_DIR/server" && go run cmd/main.go > "$SCRIPT_DIR/server.log" 2>&1) &
BACKEND_PID=$!

# 等待後端服務器啟動
sleep 2

if ! kill -0 $BACKEND_PID 2>/dev/null; then
    echo -e "${RED}✗ 後端服務器啟動失敗${NC}"
    exit 1
fi
echo -e "${GREEN}✓ 後端服務器已啟動 (PID: $BACKEND_PID)${NC}"

# 啟動前端開發服務器
echo -e "${YELLOW}🚀 啟動前端開發服務器 (Port $FRONTEND_PORT)...${NC}"
(cd "$SCRIPT_DIR/client" && npm run dev > "$SCRIPT_DIR/frontend.log" 2>&1) &
FRONTEND_PID=$!

# 等待前端服務器啟動
sleep 3

if ! kill -0 $FRONTEND_PID 2>/dev/null; then
    echo -e "${RED}✗ 前端服務器啟動失敗${NC}"
    kill $BACKEND_PID 2>/dev/null
    exit 1
fi
echo -e "${GREEN}✓ 前端開發服務器已啟動 (PID: $FRONTEND_PID)${NC}"

# 啟動 Cloudflare Tunnel
echo ""
echo -e "${YELLOW}🌐 啟動 Cloudflare Tunnel...${NC}"
echo -e "${BLUE}正在建立安全連接，請稍候...${NC}"

# 啟動 tunnel 並捕獲輸出
cloudflared tunnel --url http://localhost:$FRONTEND_PORT 2>&1 | tee "$LOG_FILE" &
TUNNEL_PID=$!

# 等待並提取 tunnel URL
echo -e "${YELLOW}等待 Cloudflare Tunnel URL...${NC}"
TIMEOUT=30
COUNT=0

while [ $COUNT -lt $TIMEOUT ]; do
    # 從日誌中提取 URL（cloudflared 會輸出類似 "Your quick Tunnel has been created! Visit it at: https://xxx.trycloudflare.com"）
    TUNNEL_URL=$(grep -oP 'https://[a-z0-9-]+\.trycloudflare\.com' "$LOG_FILE" | head -1)

    if [ ! -z "$TUNNEL_URL" ]; then
        echo "$TUNNEL_URL" > "$TUNNEL_URL_FILE"
        break
    fi

    sleep 1
    COUNT=$((COUNT + 1))

    # 檢查 tunnel 進程是否還在運行
    if ! kill -0 $TUNNEL_PID 2>/dev/null; then
        echo -e "${RED}✗ Cloudflare Tunnel 啟動失敗${NC}"
        echo -e "${YELLOW}請查看日誌: $LOG_FILE${NC}"
        kill $BACKEND_PID $FRONTEND_PID 2>/dev/null
        exit 1
    fi
done

if [ -z "$TUNNEL_URL" ]; then
    echo -e "${RED}✗ 無法獲取 Tunnel URL（超時）${NC}"
    echo -e "${YELLOW}請查看日誌: $LOG_FILE${NC}"
    cleanup
    exit 1
fi

# 顯示成功訊息
echo ""
echo -e "${GREEN}======================================${NC}"
echo -e "${GREEN}✓ 所有服務已成功啟動！${NC}"
echo -e "${GREEN}======================================${NC}"
echo ""
echo -e "${BLUE}📍 本地訪問:${NC}"
echo -e "   前端: ${YELLOW}http://localhost:$FRONTEND_PORT${NC}"
echo -e "   後端: ${YELLOW}http://localhost:$BACKEND_PORT${NC}"
echo ""
echo -e "${BLUE}🌐 公網訪問 (Cloudflare Tunnel):${NC}"
echo -e "   ${GREEN}$TUNNEL_URL${NC}"
echo ""
echo -e "${BLUE}📝 進程信息:${NC}"
echo -e "   後端 PID: $BACKEND_PID"
echo -e "   前端 PID: $FRONTEND_PID"
echo -e "   Tunnel PID: $TUNNEL_PID"
echo ""
echo -e "${YELLOW}提示:${NC}"
echo -e "  • 按 ${RED}Ctrl+C${NC} 停止所有服務"
echo -e "  • Tunnel URL 已保存到: $TUNNEL_URL_FILE"
echo -e "  • 日誌文件: server.log, frontend.log, $LOG_FILE"
echo ""

# 嘗試在瀏覽器中打開 URL
echo -e "${YELLOW}正在瀏覽器中打開 $TUNNEL_URL ...${NC}"
if command -v xdg-open &> /dev/null; then
    xdg-open "$TUNNEL_URL" &
elif command -v open &> /dev/null; then
    open "$TUNNEL_URL" &
elif command -v start &> /dev/null; then
    start "$TUNNEL_URL" &
else
    echo -e "${YELLOW}無法自動打開瀏覽器，請手動訪問: $TUNNEL_URL${NC}"
fi

# 實時顯示日誌
echo ""
echo -e "${BLUE}實時日誌 (按 Ctrl+C 停止):${NC}"
echo -e "${BLUE}======================================${NC}"

# 持續運行並顯示日誌
tail -f "$LOG_FILE" &
wait $TUNNEL_PID
