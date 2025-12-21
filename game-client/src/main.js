/**
 * Game Client - 遊戲前端入口
 * 負責麻將遊戲、語音通話
 * 可從大廳跳轉或直接訪問
 */
import { Application } from 'pixi.js';
import { Game } from './game/Game.js';
import { WebSocketClient } from './network/WebSocketClient.js';
import { VoiceChat } from './voice/VoiceChat.js';
import { VoiceChatUI } from './voice/VoiceChatUI.js';
import { GoogleAuth } from '@shared/auth/GoogleAuth.js';

class GameApp {
  constructor() {
    this.app = null;
    this.game = null;
    this.ws = null;
    this.user = null;
    this.roomId = null;

    // Voice chat
    this.voiceChat = null;
    this.voiceChatUI = null;

    // Auth
    this.auth = null;
    this.authFetch = null;

    // 追蹤玩家數量
    this.lastPlayerCount = 0;

    this.init();
  }

  init() {
    // 檢查 URL 參數（從大廳跳轉）
    const params = new URLSearchParams(window.location.search);
    this.roomId = params.get('room');
    const token = params.get('token');
    const userId = params.get('userId');
    const userName = params.get('userName');

    if (token && userId && userName) {
      // 從大廳跳轉，使用 URL 參數
      console.log('從大廳跳轉，使用 URL 參數認證');

      // 立即隱藏啟動畫面
      const splashScreen = document.getElementById('splash-screen');
      if (splashScreen) {
        splashScreen.remove();
      }

      // 清除 URL 參數（避免洩漏 token）
      window.history.replaceState({}, document.title, window.location.pathname);

      this.user = {
        id: userId,
        name: userName,
        token: token,
      };

      // 初始化認證系統並設定用戶
      this.auth = new GoogleAuth();
      this.auth.setUserFromExternal(this.user);
      this.authFetch = this.auth.createAuthFetch();

      // 綁定按鈕事件
      this.bindEvents();

      if (this.roomId) {
        this.startGame();
      } else {
        this.showError('缺少房間 ID');
      }
    } else {
      // 直接訪問，初始化認證系統
      this.initAuth();

      // 顯示啟動畫面
      this.showSplashScreen();

      // 綁定事件
      this.bindEvents();
    }
  }

  /**
   * 初始化認證系統
   */
  initAuth() {
    this.auth = new GoogleAuth();

    this.auth.onSignIn = (user) => {
      console.log('使用者已登入:', user.name);
      this.user = user;
      this.authFetch = this.auth.createAuthFetch();

      // 如果有房間參數，直接進入遊戲
      const params = new URLSearchParams(window.location.search);
      const roomId = params.get('room');
      if (roomId) {
        this.roomId = roomId;
        window.history.replaceState({}, document.title, window.location.pathname);
        this.startGame();
      }
    };

    this.auth.onSignOut = () => {
      console.log('使用者已登出');
      this.user = null;
      if (this.roomId) {
        this.leaveRoom();
      }
    };

    this.auth.init();
  }

  /**
   * 顯示啟動畫面
   */
  showSplashScreen() {
    setTimeout(() => {
      const splashScreen = document.getElementById('splash-screen');
      if (splashScreen) {
        splashScreen.classList.add('hidden');
        setTimeout(() => {
          if (splashScreen && splashScreen.parentNode) {
            splashScreen.remove();
          }
        }, 500);
      }
    }, 2000);
  }

  /**
   * 綁定事件
   */
  bindEvents() {
    // 快速開始按鈕
    const quickStartBtn = document.getElementById('quick-start-btn');
    if (quickStartBtn) {
      quickStartBtn.addEventListener('click', () => this.quickStart());
    }

    // 輸入框回車
    const playerNameInput = document.getElementById('player-name-input');
    if (playerNameInput) {
      playerNameInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
          this.quickStart();
        }
      });
    }

    // 創建房間
    const createRoomBtn = document.getElementById('create-room-btn');
    if (createRoomBtn) {
      createRoomBtn.addEventListener('click', () => this.createRoom());
    }

    // 加入房間
    const joinRoomBtn = document.getElementById('join-room-btn');
    if (joinRoomBtn) {
      joinRoomBtn.addEventListener('click', () => {
        const roomId = document.getElementById('room-id-input')?.value.trim();
        if (roomId) {
          this.roomId = roomId;
          this.startGame();
        } else {
          alert('請輸入房間號');
        }
      });
    }

    // 房間輸入框回車
    const roomIdInput = document.getElementById('room-id-input');
    if (roomIdInput) {
      roomIdInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
          const roomId = roomIdInput.value.trim();
          if (roomId) {
            this.roomId = roomId;
            this.startGame();
          }
        }
      });
    }

    // 分享房間
    const shareRoomBtn = document.getElementById('share-room-btn');
    if (shareRoomBtn) {
      shareRoomBtn.addEventListener('click', () => this.shareRoom());
    }

    // 新增Bot
    const addBotBtn = document.getElementById('add-bot-btn');
    if (addBotBtn) {
      addBotBtn.addEventListener('click', () => this.addBot());
    }

    // 切換聲音
    const toggleSoundBtn = document.getElementById('toggle-sound-btn');
    if (toggleSoundBtn) {
      toggleSoundBtn.addEventListener('click', () => this.toggleSound());
    }
  }

  /**
   * 訪客模式快速開始
   */
  quickStart() {
    if (this.user && this.auth && this.auth.isAuthenticated()) {
      console.log('使用已登入帳號:', this.user.name);
    } else {
      let playerName = document.getElementById('player-name-input')?.value.trim();

      if (!playerName) {
        const adjectives = ['快樂的', '勇敢的', '聰明的', '幸運的', '神秘的', '強大的', '可愛的', '酷炫的'];
        const nouns = ['麻將王', '雀神', '元肥', '東協', '西卡', '北麥', '南沾', '中周'];
        const randomAdj = adjectives[Math.floor(Math.random() * adjectives.length)];
        const randomNoun = nouns[Math.floor(Math.random() * nouns.length)];
        playerName = randomAdj + randomNoun;
      }

      this.user = {
        id: 'guest_' + Date.now() + '_' + Math.random().toString(36).substring(2, 11),
        name: playerName,
        picture: `https://ui-avatars.com/api/?name=${encodeURIComponent(playerName)}&background=random&size=40`,
        isGuest: true,
      };

      console.log('創建訪客玩家:', this.user);
    }

    // 檢查 URL 是否有房間參數
    const urlParams = new URLSearchParams(window.location.search);
    const roomIdFromUrl = urlParams.get('room');
    if (roomIdFromUrl) {
      console.log('從 URL 參數自動加入房間:', roomIdFromUrl);
      window.history.replaceState({}, document.title, window.location.pathname);
      this.roomId = roomIdFromUrl;
      this.startGame();
    } else {
      // 顯示房間選擇介面
      this.showRoomScreen();
    }
  }

  /**
   * 顯示房間選擇畫面
   */
  showRoomScreen() {
    document.getElementById('login-screen')?.classList.add('hidden');
    document.getElementById('room-screen')?.classList.remove('hidden');
  }

  /**
   * 顯示錯誤訊息
   */
  showError(message) {
    console.error(message);
    alert(message);
  }

  /**
   * 創建房間
   */
  async createRoom() {
    if (!this.user) {
      alert('請先登入');
      return;
    }

    try {
      const response = await fetch('/api/rooms/create', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          userId: this.user.id,
          userName: this.user.name,
          isPublic: false, // 獨立模式默認私人房間
        }),
      });

      const data = await response.json();
      if (data.success) {
        this.roomId = data.roomId;
        this.startGame();
      } else {
        alert('創建房間失敗：' + data.error);
      }
    } catch (error) {
      console.error('創建房間失敗:', error);
      alert('創建房間失敗，請檢查網路連線');
    }
  }

  /**
   * 開始遊戲
   */
  async startGame() {
    if (!this.user) {
      this.showError('請先登入');
      return;
    }

    if (!this.roomId) {
      this.showError('請先選擇或創建房間');
      return;
    }

    console.log(`開始遊戲：房間 ${this.roomId}`);

    // 顯示房間資訊
    const currentRoomId = document.getElementById('current-room-id');
    if (currentRoomId) {
      currentRoomId.textContent = this.roomId;
    }
    document.getElementById('room-info')?.classList.remove('hidden');

    // 隱藏其他介面
    document.getElementById('login-screen')?.classList.add('hidden');
    document.getElementById('room-screen')?.classList.add('hidden');
    document.getElementById('game-container')?.classList.remove('hidden');

    // 初始化遊戲
    await this.initGame();

    // 連接 WebSocket
    this.ws = new WebSocketClient(this.roomId, this.user);
    this.ws.onMessage = this.handleServerMessage.bind(this);

    if (this.game) {
      this.game.setWebSocket(this.ws);
    }

    // 初始化語音通話
    this.initVoiceChat();
  }

  /**
   * 初始化語音通話功能
   */
  initVoiceChat() {
    this.voiceChat = new VoiceChat();
    this.voiceChat.setWebSocket(this.ws, this.user.id);

    this.voiceChatUI = new VoiceChatUI(this.voiceChat);

    this.voiceChatUI.onPlayerTalkingChange = (peerId, isTalking) => {
      if (this.game) {
        this.game.setPlayerTalking(peerId, isTalking);
      }
    };

    this.voiceChatUI.onVoiceConnectionChange = (isConnected) => {
      if (this.game) {
        this.game.setVoiceButtonsVisible(isConnected);
        this.game.setBottomPlayerVoiceState(isConnected ? 'connected' : 'disconnected');
      }
    };

    this.voiceChatUI.onMuteStateChange = (peerId, isMuted, isSelf) => {
      if (this.game) {
        this.game.setPlayerVoiceMuted(peerId, isMuted);
      }
    };

    if (this.game) {
      this.game.setupVoiceButtonCallbacks((userId, isSelf) => {
        this.handleGameVoiceButtonClick(userId, isSelf);
      });

      this.game.setupVoiceConnectCallback((connect) => {
        this.handleVoiceConnectClick(connect);
      });
    }

    console.log('[VoiceChat] 語音通話功能已初始化');
  }

  async handleVoiceConnectClick(connect) {
    if (!this.voiceChat || !this.voiceChatUI || !this.game) return;

    if (connect) {
      this.game.setBottomPlayerVoiceState('connecting');

      try {
        await this.voiceChatUI.connect();
      } catch (error) {
        console.error('[VoiceChat] 連線失敗:', error);
        this.game.setBottomPlayerVoiceState('disconnected');

        if (this.game.showAnnouncement) {
          this.game.showAnnouncement('語音連線失敗，請檢查麥克風權限', 3000);
        }
      }
    } else {
      this.voiceChatUI.disconnect();
    }
  }

  handleGameVoiceButtonClick(userId, isSelf) {
    if (!this.voiceChat || !this.voiceChatUI) return;

    if (isSelf) {
      const isMuted = this.voiceChat.toggleMute();
      this.voiceChatUI.updateMuteUI(isMuted);

      if (this.game) {
        this.game.setPlayerVoiceMuted(this.user.id, isMuted);
      }
    } else {
      const isMuted = this.voiceChat.togglePeerMute(userId);
      this.voiceChatUI.updatePeerMuteUI(userId, isMuted);

      if (this.game) {
        this.game.setPlayerVoiceMuted(userId, isMuted);
      }
    }
  }

  async initGame() {
    try {
      console.log('🎮 開始初始化遊戲...');

      const container = document.getElementById('game-container');

      const MIN_WIDTH = 1280;
      const MIN_HEIGHT = 720;
      const CANVAS_WIDTH = Math.max(window.innerWidth, MIN_WIDTH);
      const CANVAS_HEIGHT = Math.max(window.innerHeight, MIN_HEIGHT);

      console.log(`📐 畫布尺寸: ${CANVAS_WIDTH}x${CANVAS_HEIGHT}`);

      console.log('📱 創建 PixiJS 應用...');
      this.app = new Application();
      await this.app.init({
        width: CANVAS_WIDTH,
        height: CANVAS_HEIGHT,
        backgroundColor: 0x1a5f3c,
        antialias: true,
        resolution: window.devicePixelRatio || 1,
        autoDensity: true,
      });

      container.appendChild(this.app.canvas);
      console.log('✅ PixiJS 應用創建成功');

      console.log('🎲 創建遊戲實體...');
      this.game = new Game(this.app, null);
      await this.game.init();
      console.log('✅ 遊戲初始化完成');

      window.addEventListener('resize', () => {
        const newWidth = Math.max(window.innerWidth, MIN_WIDTH);
        const newHeight = Math.max(window.innerHeight, MIN_HEIGHT);
        console.log(`🔄 視窗調整: ${newWidth}x${newHeight}`);
        this.app.renderer.resize(newWidth, newHeight);
        if (this.game) {
          this.game.resize(newWidth, newHeight);
        }
      });
    } catch (error) {
      console.error('❌ 遊戲初始化失敗:', error);
      alert('遊戲初始化失敗，請刷新頁面重試。\n錯誤: ' + error.message);
    }
  }

  handleServerMessage(message) {
    console.log('收到伺服器訊息:', message);

    switch (message.type) {
      case 'room_update':
        this.updateRoomInfo(message.data);
        break;
      case 'dice_roll':
        if (this.game) {
          this.game.handleDiceRoll(message.data);
        }
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
      case 'ting_result':
        this.game.handleTingResult(message.data);
        break;
      case 'game_win':
        this.game.handleGameWin(message.data);
        break;
      case 'game_draw':
        this.game.handleGameDraw(message.data);
        break;
      case 'game_over':
        this.game.gameOver(message.data);
        break;
      case 'possible_actions':
        this.game.handlePossibleActions(message.data);
        break;
      case 'player_left':
        if (this.game) {
          this.game.showPlayerLeftNotification(message.data.playerName);
        }
        break;
      case 'error':
        console.error('伺服器錯誤:', message.message);
        alert(message.message || '發生錯誤');
        if (message.message === '房間已滿' || message.message === '遊戲已開始') {
          this.leaveRoom();
        }
        break;
      case 'webrtc_signal':
        if (this.voiceChat && message.data) {
          const { fromId, signalType, payload } = message.data;
          this.voiceChat.handleSignal(fromId, signalType, payload);
        }
        break;
      default:
        console.warn('未知訊息類型:', message.type);
    }
  }

  updateRoomInfo(data) {
    console.log('更新房間資訊:', data);

    const oldPlayerCount = this.lastPlayerCount || 0;
    const newPlayerCount = data.playerCount;

    if (newPlayerCount > oldPlayerCount && oldPlayerCount > 0) {
      const newPlayers = data.players.slice(oldPlayerCount);
      for (const player of newPlayers) {
        if (player && player.name) {
          console.log(`新玩家加入: ${player.name}`);
          if (this.game) {
            this.game.showPlayerJoinNotification(player.name);
          }
        }
      }
    }

    this.lastPlayerCount = newPlayerCount;
    const playerCount = document.getElementById('player-count');
    if (playerCount) {
      playerCount.textContent = `${data.playerCount}/4`;
    }

    if (this.game) {
      this.game.updatePlayers(data.players);
    }

    if (this.voiceChat && this.voiceChatUI && data.players) {
      const otherPlayers = data.players
        .filter((p) => p && p.id !== this.user.id && !p.id.startsWith('bot_'))
        .map((p) => ({ id: p.id, name: p.name }));
      this.voiceChat.setRoomPlayers(otherPlayers);
      this.voiceChatUI.updatePlayerList(otherPlayers);
    }
  }

  toggleSound() {
    if (!this.game || !this.game.audioManager) {
      console.warn('遊戲尚未初始化');
      return;
    }

    const enabled = this.game.audioManager.toggle();
    const btn = document.getElementById('toggle-sound-btn');

    if (enabled) {
      btn.textContent = '🔊 聲音開';
      btn.style.background = '#667eea';
      this.game.audioManager.playBGM('game');
    } else {
      btn.textContent = '🔇 聲音關';
      btn.style.background = '#999';
    }
  }

  addBot() {
    if (!this.ws) {
      alert('請先創建或加入房間');
      return;
    }

    this.ws.sendAction('add_bot', {
      roomId: this.roomId,
    });

    console.log('請求新增Bot');
  }

  shareRoom() {
    const shareUrl = `${window.location.origin}${window.location.pathname}?room=${this.roomId}`;
    const shareText = `加入我的麻將遊戲！\n房間號：${this.roomId}\n連結：${shareUrl}`;

    if (navigator.share) {
      navigator.share({
        title: '臺灣16張麻將',
        text: shareText,
        url: shareUrl,
      }).catch((err) => console.log('分享失敗:', err));
    } else {
      navigator.clipboard
        .writeText(shareUrl)
        .then(() => {
          alert(`房間連結已複製到剪貼簿！\n${shareUrl}`);
        })
        .catch(() => {
          prompt('複製此連結分享給朋友:', shareUrl);
        });
    }
  }

  leaveRoom() {
    if (this.voiceChat) {
      this.voiceChat.disconnect();
      this.voiceChat = null;
    }
    if (this.voiceChatUI) {
      this.voiceChatUI.destroy();
      this.voiceChatUI = null;
    }

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }

    this.roomId = null;
    this.lastPlayerCount = 0;

    if (this.game) {
      this.game.destroy();
      this.game = null;
    }

    if (this.app) {
      this.app.destroy(true);
      this.app = null;
    }

    document.getElementById('game-container')?.classList.add('hidden');
    document.getElementById('room-info')?.classList.add('hidden');
    document.getElementById('room-screen')?.classList.remove('hidden');

    console.log('已離開房間');
  }
}

// 啟動應用
new GameApp();
