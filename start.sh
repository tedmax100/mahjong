#!/bin/bash

echo "🀄 台湾16张麻将 - 启动脚本"
echo "================================"

# 检查素材是否存在
if [ ! -d "client/public/assets/tiles" ] || [ -z "$(ls -A client/public/assets/tiles)" ]; then
    echo "⚠️  素材未生成！"
    echo "请先运行: open tools/generate-tiles.html"
    echo "然后点击'生成所有素材'和'打包下载全部'"
    echo "将tiles解压到 client/public/assets/ 目录"
    echo ""
    read -p "按Enter键继续（将使用占位符）或Ctrl+C退出..."
    mkdir -p client/public/assets/tiles
fi

# 启动后端
echo ""
echo "📡 启动Go后端服务器..."
cd server
go mod download 2>/dev/null
go run cmd/main.go &
BACKEND_PID=$!
cd ..

# 等待后端启动
sleep 2

# 启动前端
echo ""
echo "🎨 启动前端开发服务器..."
cd client
if [ ! -d "node_modules" ]; then
    echo "安装前端依赖..."
    npm install
fi
npm run dev &
FRONTEND_PID=$!
cd ..

echo ""
echo "================================"
echo "✅ 服务启动成功！"
echo "后端: http://localhost:8080"
echo "前端: http://localhost:3000"
echo ""
echo "按 Ctrl+C 停止所有服务"
echo "================================"

# 捕获Ctrl+C信号
trap "echo ''; echo '停止服务...'; kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; exit" INT

# 等待
wait
