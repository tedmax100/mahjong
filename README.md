# 🀄 台湾16张麻将 - 在线对战游戏

一个支持四人在线对战的台湾16张麻将网页游戏，使用Pixi.js渲染和Go语言后端。

## ✨ 功能特性

- ✅ 四人实时在线对战
- ✅ Google OAuth 登录认证
- ✅ 房间创建和加入系统
- ✅ 房间分享功能（房号 + URL链接）
- ✅ WebSocket实时通信
- ✅ Pixi.js 2D游戏渲染
- ✅ 台湾16张麻将规则
- ✅ 自动生成麻将牌素材

## 🏗️ 项目结构

```
mahjong/
├── client/                 # 前端项目 (Vite + Pixi.js)
│   ├── src/
│   │   ├── game/          # 游戏核心逻辑
│   │   │   ├── Game.js    # 主游戏类
│   │   │   ├── Table.js   # 牌桌
│   │   │   ├── Player.js  # 玩家
│   │   │   └── Tile.js    # 麻将牌
│   │   ├── network/
│   │   │   └── WebSocketClient.js
│   │   ├── auth/
│   │   │   └── GoogleAuth.js
│   │   └── main.js
│   ├── public/
│   │   └── assets/tiles/  # 麻将牌素材
│   ├── index.html
│   └── package.json
│
├── server/                 # Go后端项目
│   ├── cmd/
│   │   └── main.go        # 主程序入口
│   ├── internal/
│   │   ├── websocket/     # WebSocket处理
│   │   │   ├── hub.go     # 连接管理
│   │   │   └── client.go  # 客户端
│   │   ├── game/          # 游戏逻辑
│   │   │   ├── room.go    # 房间管理
│   │   │   └── mahjong.go # 麻将规则
│   │   └── api/           # HTTP API
│   │       └── room.go
│   └── go.mod
│
└── tools/                  # 工具
    ├── generate-tiles.html # 素材生成器（浏览器）
    └── package.json
```

## 🚀 快速开始

### 1. 生成麻将牌素材

```bash
# 用浏览器打开素材生成器
open tools/generate-tiles.html

# 点击"生成所有素材"，然后"打包下载全部"
# 解压 mahjong-tiles.zip 到 client/public/assets/tiles/
```

### 2. 启动后端服务器

```bash
cd server

# 下载依赖
go mod download

# 运行服务器
go run cmd/main.go
```

后端将在 `http://localhost:8080` 启动

### 3. 启动前端开发服务器

```bash
cd client

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

前端将在 `http://localhost:3000` 启动

### 4. 访问游戏

打开浏览器访问 `http://localhost:3000`

## 🎮 游戏流程

1. **登录** - 使用Google账号登录（或测试模式）
2. **创建/加入房间**
   - 点击"创建新房间"获得房间号
   - 或输入房间号"加入房间"
3. **分享房间** - 点击"分享房间"复制链接给朋友
4. **等待玩家** - 等待4名玩家齐聚
5. **开始游戏** - 系统自动发牌，游戏开始！

## 🔧 配置

### Google OAuth 设置

1. 前往 [Google Cloud Console](https://console.cloud.google.com/)
2. 创建新项目或选择现有项目
3. 启用 Google+ API
4. 创建 OAuth 2.0 客户端ID
5. 将客户端ID填入：
   - `client/index.html` 第7行
   - `client/src/auth/GoogleAuth.js` 第12行

```html
<!-- client/index.html -->
<meta name="google-signin-client_id" content="YOUR_CLIENT_ID.apps.googleusercontent.com">
```

```javascript
// client/src/auth/GoogleAuth.js
client_id: 'YOUR_CLIENT_ID.apps.googleusercontent.com',
```

## 📦 生产环境部署

### 1. 构建前端

```bash
cd client
npm run build
```

构建文件将输出到 `dist/public/`

### 2. 部署后端

```bash
cd server

# 构建二进制文件
go build -o mahjong cmd/main.go

# 运行
./mahjong
```

### 3. 使用Docker（可选）

```bash
# TODO: 添加Dockerfile
```

## 🎯 台湾16张麻将规则

### 基本规则

- **牌数**: 每人16张牌（庄家17张）
- **牌型**: 万、筒、条、风牌、三元牌、花牌
- **花牌**: 抽到花牌立即明示并补牌
- **胡牌**: 5组顺子/刻子 + 1对眼

### 台数计算（简化版）

- 自摸: +1台
- 门清: +1台
- 碰碰胡: +2台
- 清一色: +5台
- 字一色: +8台
- ... (更多台型待实现)

## 🛠️ 技术栈

**前端:**
- Pixi.js 8.x - 2D WebGL渲染
- Vite - 构建工具
- JavaScript (ES6+)

**后端:**
- Go 1.24
- Gin - Web框架
- Gorilla WebSocket
- Google UUID

**素材生成:**
- HTML5 Canvas API
- JSZip - 打包下载

## 📝 待实现功能

- [ ] 完整的台湾16张麻将胡牌判断
- [ ] 台数计算系统
- [ ] AI玩家
- [ ] 游戏历史记录
- [ ] 排行榜系统
- [ ] 牌局回放
- [ ] 音效和动画
- [ ] 移动端适配
- [ ] Google OAuth后端验证
- [ ] Redis缓存房间状态

## 🤝 贡献

欢迎提交Issue和Pull Request！

## 📄 许可证

MIT License

## 👨‍💻 作者

Generated with Claude Code

---

**Enjoy the game! 🀄**
