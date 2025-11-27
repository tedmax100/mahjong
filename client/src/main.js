import { Application } from 'pixi.js';
import { Game } from './game/Game.js';
import { WebSocketClient } from './network/WebSocketClient.js';

class MahjongApp {
  constructor() {
    this.app = null;
    this.game = null;
    this.ws = null;
    this.user = null;
    this.roomId = null;

    this.init();
  }

  init() {
    // 显示启动画面动画
    this.showSplashScreen();

    // 绑定UI事件
    this.bindEvents();
  }

  showSplashScreen() {
    // 3秒后隐藏启动画面
    setTimeout(() => {
      const splashScreen = document.getElementById('splash-screen');
      splashScreen.classList.add('hidden');

      // 再等0.5秒后完全移除元素（等待淡出动画完成）
      setTimeout(() => {
        splashScreen.remove();
      }, 500);
    }, 3000);
  }

  bindEvents() {
    // 快速开始按钮
    document.getElementById('quick-start-btn').addEventListener('click', () => {
      this.quickStart();
    });

    // 输入框回车
    document.getElementById('player-name-input').addEventListener('keypress', (e) => {
      if (e.key === 'Enter') {
        this.quickStart();
      }
    });

    // 创建房间
    document.getElementById('create-room-btn').addEventListener('click', () => {
      this.createRoom();
    });

    // 加入房间
    document.getElementById('join-room-btn').addEventListener('click', () => {
      const roomId = document.getElementById('room-id-input').value.trim();
      if (roomId) {
        this.joinRoom(roomId);
      } else {
        alert('请输入房间号');
      }
    });

    // 分享房间
    document.getElementById('share-room-btn').addEventListener('click', () => {
      this.shareRoom();
    });

    // 添加Bot
    document.getElementById('add-bot-btn').addEventListener('click', () => {
      this.addBot();
    });
  }

  quickStart() {
    // 获取输入的名字或生成随机名字
    let playerName = document.getElementById('player-name-input').value.trim();

    if (!playerName) {
      // 随机生成名字
      const adjectives = ['快乐的', '勇敢的', '聪明的', '幸运的', '神秘的', '强大的', '可爱的', '酷炫的'];
      const nouns = ['麻将王', '牌神', '高手', '大师', '玩家', '战士', '冠军', '传说'];
      const randomAdj = adjectives[Math.floor(Math.random() * adjectives.length)];
      const randomNoun = nouns[Math.floor(Math.random() * nouns.length)];
      playerName = randomAdj + randomNoun;
    }

    // 创建用户对象
    this.user = {
      id: 'player_' + Date.now() + '_' + Math.random().toString(36).substring(2, 11),
      name: playerName,
      picture: `https://ui-avatars.com/api/?name=${encodeURIComponent(playerName)}&background=random&size=40`
    };

    console.log('创建玩家:', this.user);

    // 显示房间选择界面
    this.showRoomScreen();
  }

  showRoomScreen() {
    document.getElementById('login-screen').classList.add('hidden');
    document.getElementById('room-screen').classList.remove('hidden');

    // 显示用户信息
    const userInfo = document.getElementById('user-info');
    document.getElementById('user-avatar').src = this.user.picture;
    document.getElementById('user-name').textContent = this.user.name;
    userInfo.classList.remove('hidden');
  }

  async createRoom() {
    try {
      const response = await fetch('/api/rooms/create', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${this.user.token}`
        },
        body: JSON.stringify({
          userId: this.user.id,
          userName: this.user.name
        })
      });

      const data = await response.json();
      if (data.success) {
        this.joinRoom(data.roomId);
      } else {
        alert('创建房间失败：' + data.error);
      }
    } catch (error) {
      console.error('创建房间失败:', error);
      alert('创建房间失败，请检查网络连接');
    }
  }

  async joinRoom(roomId) {
    this.roomId = roomId;

    // 显示房间信息
    document.getElementById('current-room-id').textContent = roomId;
    document.getElementById('room-info').classList.remove('hidden');

    // 连接WebSocket
    this.ws = new WebSocketClient(roomId, this.user);
    this.ws.onMessage = this.handleServerMessage.bind(this);

    // 隐藏房间选择界面
    document.getElementById('room-screen').classList.add('hidden');
    document.getElementById('game-container').classList.remove('hidden');

    // 初始化Pixi游戏
    await this.initGame();
  }

  async initGame() {
    const container = document.getElementById('game-container');

    // 创建Pixi应用
    this.app = new Application();
    await this.app.init({
      width: window.innerWidth,
      height: window.innerHeight,
      backgroundColor: 0x1a5f3c,
      antialias: true,
      resolution: window.devicePixelRatio || 1,
      autoDensity: true
    });

    container.appendChild(this.app.canvas);

    // 创建游戏实例
    this.game = new Game(this.app, this.ws);
    await this.game.init();

    // 窗口大小调整
    window.addEventListener('resize', () => {
      this.app.renderer.resize(window.innerWidth, window.innerHeight);
      this.game.resize(window.innerWidth, window.innerHeight);
    });
  }

  handleServerMessage(message) {
    console.log('收到服务器消息:', message);

    switch (message.type) {
      case 'room_update':
        this.updateRoomInfo(message.data);
        break;
      case 'game_start':
        this.game.startGame(message.data);
        break;
      case 'deal_tiles':
        this.game.dealTiles(message.data);
        break;
      case 'player_action':
        this.game.handlePlayerAction(message.data);
        break;
      case 'game_over':
        this.game.gameOver(message.data);
        break;
      default:
        console.warn('未知消息类型:', message.type);
    }
  }

  updateRoomInfo(data) {
    console.log('更新房间信息:', data);
    document.getElementById('player-count').textContent = `${data.playerCount}/4`;

    // 即使游戏还没开始，也更新玩家列表
    if (this.game) {
      this.game.updatePlayers(data.players);
    }
  }

  addBot() {
    if (!this.ws) {
      alert('请先创建或加入房间');
      return;
    }

    // 发送添加Bot消息到服务器
    this.ws.sendAction('add_bot', {
      roomId: this.roomId
    });

    console.log('请求添加Bot');
  }

  shareRoom() {
    const shareUrl = `${window.location.origin}?room=${this.roomId}`;
    const shareText = `加入我的麻将游戏！\n房间号：${this.roomId}\n链接：${shareUrl}`;

    if (navigator.share) {
      navigator.share({
        title: '台湾16张麻将',
        text: shareText,
        url: shareUrl
      }).catch(err => console.log('分享失败:', err));
    } else {
      // 复制到剪贴板
      navigator.clipboard.writeText(shareUrl).then(() => {
        alert(`房间链接已复制到剪贴板！\n${shareUrl}`);
      }).catch(() => {
        prompt('复制此链接分享给朋友:', shareUrl);
      });
    }
  }
}

// 启动应用
new MahjongApp();
