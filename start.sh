#!/bin/bash

echo "🀄 臺灣16張麻將 - 啟動腳本"
echo "================================"

# 檢查素材是否存在
if [ ! -d "client/public/assets/tiles" ] || [ -z "$(ls -A client/public/assets/tiles)" ]; then
    echo "⚠️  素材未生成！"
    echo "請先運行: open tools/generate-tiles.html"
    echo "然後點擊'生成所有素材'和'打包下載全部'"
    echo "將tiles解壓縮到 client/public/assets/ 目錄"
    echo ""
    read -p "按Enter鍵繼續（將使用佔位符）或Ctrl+C退出..."
    mkdir -p client/public/assets/tiles
fi

# 啟動後端
echo ""
echo "📡 啟動Go後端伺服器..."
cd server
go mod download 2>/dev/null
go run cmd/main.go &
BACKEND_PID=$!
cd ..

# 等待後端啟動
sleep 2

# 啟動前端
echo ""
echo "🎨 啟動前端開發伺服器..."
cd client
if [ ! -d "node_modules" ]; then
    echo "安裝前端依賴..."
    npm install
fi
npm run dev &
FRONTEND_PID=$!
cd ..

echo ""
echo "================================"
echo "✅ 服務啟動成功！"
echo "後端: http://localhost:8080"
echo "前端: http://localhost:3000"
echo ""
echo "按 Ctrl+C 停止所有服務"
echo "================================"

# 捕獲Ctrl+C信號
trap "echo ''; echo '停止服務...'; kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; exit" INT

# 等待
wait