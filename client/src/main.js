import { Application } from 'pixi.js';
import { Game } from './game/Game.js';
import { WebSocketClient } from './network/WebSocketClient.js';
import { LobbyClient } from './network/LobbyClient.js';
import { VoiceChat } from './voice/VoiceChat.js';
import { VoiceChatUI } from './voice/VoiceChatUI.js';
import { GoogleAuth } from './auth/GoogleAuth.js';

class MahjongApp {
  constructor() {
    this.app = null;
    this.game = null;
    this.ws = null;
    this.user = null;
    this.roomId = null;

    // Voice chat
    this.voiceChat = null;
    this.voiceChatUI = null;

    // Lobby
    this.lobbyClient = null;

    // Auth
    this.auth = null;
    this.authFetch = null;

    this.init();
  }

  init() {
    // 初始化認證系統
    this.initAuth();

    // 顯示啟動畫面動畫
    this.showSplashScreen();

    // 綁定UI事件
    this.bindEvents();
  }

  /**
   * 初始化認證系統
   */
  initAuth() {
    this.auth = new GoogleAuth();

    // 設定登入回調
    this.auth.onSignIn = (user) => {
      console.log('使用者已登入:', user.name);
      this.user = user;
      // 如果已在登入畫面，自動跳轉到大廳
      const loginScreen = document.getElementById('login-screen');
      if (loginScreen && !loginScreen.classList.contains('hidden')) {
        this.showLobbyScreen();
      }
    };

    // 設定登出回調
    this.auth.onSignOut = () => {
      console.log('使用者已登出');
      this.user = null;
      // 如果在遊戲中，離開房間
      if (this.roomId) {
        this.leaveRoom();
      }
    };

    // 初始化（檢查 URL token 和 localStorage）
    this.auth.init();

    // 建立帶認證的 fetch（自動處理 401）
    this.authFetch = this.auth.createAuthFetch();
  }

  showSplashScreen() {
    // 3秒後隱藏啟動畫面
    setTimeout(() => {
      try {
        const splashScreen = document.getElementById('splash-screen');
        if (splashScreen) {
          splashScreen.classList.add('hidden');

          // 再等0.5秒後完全移除元素（等待淡出動畫完成）
          setTimeout(() => {
            if (splashScreen && splashScreen.parentNode) {
              splashScreen.remove();
            }
          }, 500);
        }
      } catch (error) {
        console.error('隱藏啟動畫面失敗:', error);
      }
    }, 3000);
  }

  bindEvents() {
    // 快速開始按鈕
    document.getElementById('quick-start-btn').addEventListener('click', () => {
      this.quickStart();
    });

    // 輸入框回車
    document.getElementById('player-name-input').addEventListener('keypress', (e) => {
      if (e.key === 'Enter') {
        this.quickStart();
      }
    });

    // 創建房間
    document.getElementById('create-room-btn').addEventListener('click', () => {
      this.createRoom();
    });

    // 加入房間
    document.getElementById('join-room-btn').addEventListener('click', () => {
      const roomId = document.getElementById('room-id-input').value.trim();
      if (roomId) {
        this.joinRoom(roomId);
      } else {
        alert('請輸入房間號');
      }
    });

    // 分享房間
    document.getElementById('share-room-btn').addEventListener('click', () => {
      this.shareRoom();
    });

    // 新增Bot
    document.getElementById('add-bot-btn').addEventListener('click', () => {
      this.addBot();
    });

    // 切換聲音
    document.getElementById('toggle-sound-btn').addEventListener('click', () => {
      this.toggleSound();
    });

    // === 大廳相關事件 ===

    // 返回登入畫面
    document.getElementById('lobby-back-btn').addEventListener('click', () => {
      this.leaveLobby();
    });

    // 創建公開房間
    document.getElementById('create-public-room-btn').addEventListener('click', () => {
      this.createRoom(true);
    });

    // 創建私人房間
    document.getElementById('create-private-room-btn').addEventListener('click', () => {
      this.createRoom(false);
    });

    // 加入私人房間
    document.getElementById('join-private-btn').addEventListener('click', () => {
      const roomId = document.getElementById('private-room-id').value.trim();
      if (roomId) {
        this.joinRoom(roomId);
      } else {
        alert('請輸入房間號');
      }
    });

    // 私人房間輸入框回車
    document.getElementById('private-room-id').addEventListener('keypress', (e) => {
      if (e.key === 'Enter') {
        const roomId = document.getElementById('private-room-id').value.trim();
        if (roomId) {
          this.joinRoom(roomId);
        }
      }
    });

    // 發送聊天訊息
    document.getElementById('send-chat-btn').addEventListener('click', () => {
      this.sendChatMessage();
    });

    // 聊天輸入框回車
    document.getElementById('chat-input').addEventListener('keypress', (e) => {
      if (e.key === 'Enter') {
        this.sendChatMessage();
      }
    });
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
      // 重新播放背景音樂
      this.game.audioManager.playBGM('game');
    } else {
      btn.textContent = '🔇 聲音關';
      btn.style.background = '#999';
    }
  }

  quickStart() {
    // 如果已經透過 Google 登入，直接使用該帳號
    if (this.user && this.auth.isAuthenticated()) {
      console.log('使用 Google 帳號:', this.user.name);
    } else {
      // 訪客模式：獲取輸入的名字或生成隨機名字
      let playerName = document.getElementById('player-name-input').value.trim();

      if (!playerName) {
        // 隨機生成名字
        const adjectives = ['快樂的', '勇敢的', '聰明的', '幸運的', '神秘的', '強大的', '可愛的', '酷炫的'];
        const nouns = ['麻將王', '雀神', '元肥', '東協', '西卡', '北麥', '南沾', '中周'];
        const randomAdj = adjectives[Math.floor(Math.random() * adjectives.length)];
        const randomNoun = nouns[Math.floor(Math.random() * nouns.length)];
        playerName = randomAdj + randomNoun;
      }

      // 創建訪客使用者物件
      this.user = {
        id: 'guest_' + Date.now() + '_' + Math.random().toString(36).substring(2, 11),
        name: playerName,
        picture: `https://ui-avatars.com/api/?name=${encodeURIComponent(playerName)}&background=random&size=40`
      };

      console.log('創建訪客玩家:', this.user);
    }

    // 檢查 URL 是否有房間參數，自動加入房間
    const urlParams = new URLSearchParams(window.location.search);
    const roomIdFromUrl = urlParams.get('room');
    if (roomIdFromUrl) {
      console.log('從 URL 參數自動加入房間:', roomIdFromUrl);
      // 清除 URL 參數，避免重複加入
      window.history.replaceState({}, document.title, window.location.pathname);
      this.joinRoom(roomIdFromUrl);
    } else {
      // 顯示大廳介面
      this.showLobbyScreen();
    }
  }

  /**
   * 顯示大廳介面
   */
  showLobbyScreen() {
    document.getElementById('login-screen').classList.add('hidden');
    document.getElementById('room-screen').classList.add('hidden');
    document.getElementById('lobby-screen').classList.remove('hidden');

    // 連接大廳 WebSocket
    this.lobbyClient = new LobbyClient(this.user);

    this.lobbyClient.onRoomListUpdate = (rooms) => {
      this.updateRoomList(rooms);
    };

    this.lobbyClient.onChatMessage = (msg) => {
      this.addChatMessage(msg);
    };

    this.lobbyClient.onOnlineCountUpdate = (count) => {
      document.getElementById('online-count').textContent = count;
    };

    this.lobbyClient.onError = (message) => {
      console.error('[Lobby] 錯誤:', message);
    };

    this.lobbyClient.connect();
  }

  /**
   * 離開大廳，返回登入畫面
   */
  leaveLobby() {
    if (this.lobbyClient) {
      this.lobbyClient.close();
      this.lobbyClient = null;
    }

    document.getElementById('lobby-screen').classList.add('hidden');
    document.getElementById('login-screen').classList.remove('hidden');

    // 清空聊天訊息
    document.getElementById('chat-messages').innerHTML = '';
  }

  /**
   * 更新房間列表
   */
  updateRoomList(rooms) {
    const container = document.getElementById('room-list');

    if (!rooms || rooms.length === 0) {
      container.innerHTML = `
        <div class="empty-room-list">
          <div class="empty-room-list-icon">🀄</div>
          <div>目前沒有公開房間</div>
          <div style="font-size: 14px; margin-top: 5px;">創建一個新房間開始遊戲吧！</div>
        </div>
      `;
      return;
    }

    container.innerHTML = rooms.map(room => `
      <div class="room-item" data-room-id="${room.id}">
        <div class="room-item-info">
          <span class="room-name">${this.escapeHtml(room.name || '房間 ' + room.id)}</span>
          <span class="room-host">房主: ${this.escapeHtml(room.hostName)}</span>
        </div>
        <div class="room-item-status">
          <span class="player-count-badge">${room.playerCount}/4</span>
          <button class="join-btn" ${room.playerCount >= 4 ? 'disabled' : ''}>
            ${room.playerCount >= 4 ? '已滿' : '加入'}
          </button>
        </div>
      </div>
    `).join('');

    // 綁定加入按鈕事件
    container.querySelectorAll('.join-btn:not([disabled])').forEach(btn => {
      btn.addEventListener('click', (e) => {
        const roomId = e.target.closest('.room-item').dataset.roomId;
        this.joinRoom(roomId);
      });
    });
  }

  /**
   * 添加聊天訊息
   */
  addChatMessage(msg) {
    const container = document.getElementById('chat-messages');
    const msgElement = document.createElement('div');
    msgElement.className = `chat-message ${msg.type === 'system' ? 'system' : ''}`;

    const time = new Date(msg.timestamp).toLocaleTimeString('zh-TW', {
      hour: '2-digit',
      minute: '2-digit'
    });

    msgElement.innerHTML = `
      <span class="chat-time">${time}</span>
      <span class="chat-user">${this.escapeHtml(msg.userName)}:</span>
      <span class="chat-content">${this.escapeHtml(msg.content)}</span>
    `;

    container.appendChild(msgElement);
    container.scrollTop = container.scrollHeight;
  }

  /**
   * 發送聊天訊息
   */
  sendChatMessage() {
    const input = document.getElementById('chat-input');
    const content = input.value.trim();

    if (content && this.lobbyClient) {
      this.lobbyClient.sendChat(content);
      input.value = '';
    }
  }

  /**
   * HTML 轉義
   */
  escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
  }

  // 舊的房間選擇介面（保留向後兼容）
  showRoomScreen() {
    document.getElementById('login-screen').classList.add('hidden');
    document.getElementById('room-screen').classList.remove('hidden');
  }

  async createRoom(isPublic = false) {
    try {
      // 使用 authFetch 自動附加 token 並處理 401
      const response = await this.authFetch('/api/rooms/create', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          userId: this.user.id,
          userName: this.user.name,
          isPublic: isPublic
        })
      });

      // 如果是 401，authFetch 已經處理登出，這裡直接返回
      if (response.status === 401) {
        return;
      }

      const data = await response.json();
      if (data.success) {
        this.joinRoom(data.roomId);
      } else {
        alert('創建房間失敗：' + data.error);
      }
    } catch (error) {
      console.error('創建房間失敗:', error);
      alert('創建房間失敗，請檢查網路連線');
    }
  }

  async joinRoom(roomId) {
    this.roomId = roomId;

    // 關閉大廳連接
    if (this.lobbyClient) {
      this.lobbyClient.close();
      this.lobbyClient = null;
    }

    // 顯示房間資訊
    document.getElementById('current-room-id').textContent = roomId;
    document.getElementById('room-info').classList.remove('hidden');

    // 隱藏大廳和房間選擇介面
    document.getElementById('lobby-screen').classList.add('hidden');
    document.getElementById('room-screen').classList.add('hidden');
    document.getElementById('game-container').classList.remove('hidden');

    // 先初始化Pixi遊戲，確保 this.game 準備好
    await this.initGame();

    // 遊戲初始化完成後才連接 WebSocket，避免錯過訊息
    this.ws = new WebSocketClient(roomId, this.user);
    this.ws.onMessage = this.handleServerMessage.bind(this);

    // 將 ws 實體傳給 game，以便發送訊息
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

    // 設定說話指示器回調 - 當玩家說話時更新遊戲中的指示器
    this.voiceChatUI.onPlayerTalkingChange = (peerId, isTalking) => {
      if (this.game) {
        this.game.setPlayerTalking(peerId, isTalking);
      }
    };

    // 設定語音連線狀態變更回調 - 控制玩家名牌上的語音按鈕顯示
    this.voiceChatUI.onVoiceConnectionChange = (isConnected) => {
      if (this.game) {
        // 更新其他玩家的靜音按鈕顯示狀態
        this.game.setVoiceButtonsVisible(isConnected);
        // 更新底部玩家的連線狀態顯示
        this.game.setBottomPlayerVoiceState(isConnected ? 'connected' : 'disconnected');
      }
    };

    // 設定靜音狀態同步回調 - 當 VoiceChatUI 面板中的靜音狀態改變時同步到遊戲
    this.voiceChatUI.onMuteStateChange = (peerId, isMuted, isSelf) => {
      if (this.game) {
        this.game.setPlayerVoiceMuted(peerId, isMuted);
      }
    };

    // 設定遊戲中玩家語音按鈕的回調（靜音/取消靜音）
    if (this.game) {
      this.game.setupVoiceButtonCallbacks((userId, isSelf) => {
        this.handleGameVoiceButtonClick(userId, isSelf);
      });

      // 設定底部玩家語音連線按鈕的回調
      this.game.setupVoiceConnectCallback((connect) => {
        this.handleVoiceConnectClick(connect);
      });
    }

    console.log('[VoiceChat] 語音通話功能已初始化');
  }

  /**
   * 處理語音連線/斷線按鈕點擊（底部玩家）
   * @param {boolean} connect - true 為連線，false 為斷線
   */
  async handleVoiceConnectClick(connect) {
    if (!this.voiceChat || !this.voiceChatUI || !this.game) return;

    if (connect) {
      // 顯示連線中狀態
      this.game.setBottomPlayerVoiceState('connecting');

      try {
        // 連線
        await this.voiceChatUI.connect();

        // 連線成功後會透過 onVoiceConnectionChange 回調更新其他玩家的按鈕
      } catch (error) {
        console.error('[VoiceChat] 連線失敗:', error);
        // 恢復為未連線狀態
        this.game.setBottomPlayerVoiceState('disconnected');

        // 在遊戲界面中顯示錯誤訊息
        if (this.game.showAnnouncement) {
          this.game.showAnnouncement('語音連線失敗，請檢查麥克風權限', 3000);
        }
      }
    } else {
      // 斷線
      this.voiceChatUI.disconnect();
    }
  }

  /**
   * 處理遊戲中玩家語音按鈕點擊
   * @param {string} userId - 玩家 ID
   * @param {boolean} isSelf - 是否是自己的麥克風按鈕
   */
  handleGameVoiceButtonClick(userId, isSelf) {
    if (!this.voiceChat || !this.voiceChatUI) return;

    if (isSelf) {
      // 切換自己的麥克風
      const isMuted = this.voiceChat.toggleMute();
      this.voiceChatUI.updateMuteUI(isMuted);

      // 同步到遊戲中的按鈕
      if (this.game) {
        this.game.setPlayerVoiceMuted(this.user.id, isMuted);
      }
    } else {
      // 切換對方的靜音狀態
      const isMuted = this.voiceChat.togglePeerMute(userId);

      // 同步到 VoiceChatUI
      this.voiceChatUI.updatePeerMuteUI(userId, isMuted);

      // 同步到遊戲中的按鈕
      if (this.game) {
        this.game.setPlayerVoiceMuted(userId, isMuted);
      }
    }
  }

  async initGame() {
    try {
      console.log('🎮 開始初始化遊戲...');

      const container = document.getElementById('game-container');

      // 固定最小尺寸，如果螢幕更大則使用螢幕尺寸
      const MIN_WIDTH = 1280;
      const MIN_HEIGHT = 720;
      const CANVAS_WIDTH = Math.max(window.innerWidth, MIN_WIDTH);
      const CANVAS_HEIGHT = Math.max(window.innerHeight, MIN_HEIGHT);

      console.log(`📐 畫布尺寸: ${CANVAS_WIDTH}x${CANVAS_HEIGHT}`);

      // 創建Pixi應用
      console.log('📱 創建 PixiJS 應用...');
      this.app = new Application();
      await this.app.init({
        width: CANVAS_WIDTH,
        height: CANVAS_HEIGHT,
        backgroundColor: 0x1a5f3c,
        antialias: true,
        resolution: window.devicePixelRatio || 1,
        autoDensity: true
      });

      container.appendChild(this.app.canvas);
      console.log('✅ PixiJS 應用創建成功');

      // 創建遊戲實體（ws 稍後透過 setWebSocket 設定）
      console.log('🎲 創建遊戲實體...');
      this.game = new Game(this.app, null);
      await this.game.init();
      console.log('✅ 遊戲初始化完成');

      // 新增視窗大小調整（響應式，但有最小尺寸限制）
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
        // 擲骰決定莊家動畫
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
        // 如果是房間已滿或遊戲已開始，返回房間選擇畫面
        if (message.message === '房間已滿' || message.message === '遊戲已開始') {
          this.leaveRoom();
        }
        break;
      case 'webrtc_signal':
        // 處理 WebRTC 信令訊息
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

    // 檢查是否有新玩家加入
    const oldPlayerCount = this.lastPlayerCount || 0;
    const newPlayerCount = data.playerCount;

    // 如果有新玩家加入（且不是自己剛加入）
    if (newPlayerCount > oldPlayerCount && oldPlayerCount > 0) {
      // 找出新加入的玩家
      const newPlayers = data.players.slice(oldPlayerCount);
      for (const player of newPlayers) {
        if (player && player.name) {
          console.log(`新玩家加入: ${player.name}`);
          // 顯示加入通知
          if (this.game) {
            this.game.showPlayerJoinNotification(player.name);
          }
        }
      }
    }

    this.lastPlayerCount = newPlayerCount;
    document.getElementById('player-count').textContent = `${data.playerCount}/4`;

    // 即使遊戲還沒開始，也更新玩家列表
    if (this.game) {
      this.game.updatePlayers(data.players);
    }

    // 更新語音通話玩家列表（過濾掉 Bot 玩家，只保留真人玩家）
    if (this.voiceChat && this.voiceChatUI && data.players) {
      const otherPlayers = data.players
        .filter(p => p && p.id !== this.user.id && !p.id.startsWith('bot_'))
        .map(p => ({ id: p.id, name: p.name }));
      this.voiceChat.setRoomPlayers(otherPlayers);
      this.voiceChatUI.updatePlayerList(otherPlayers);
    }
  }

  addBot() {
    if (!this.ws) {
      alert('請先創建或加入房間');
      return;
    }

    // 發送新增Bot訊息到伺服器
    this.ws.sendAction('add_bot', {
      roomId: this.roomId
    });

    console.log('請求新增Bot');
  }

  shareRoom() {
    const shareUrl = `${window.location.origin}?room=${this.roomId}`;
    const shareText = `加入我的麻將遊戲！\n房間號：${this.roomId}\n連結：${shareUrl}`;

    if (navigator.share) {
      navigator.share({
        title: '臺灣16張麻將',
        text: shareText,
        url: shareUrl
      }).catch(err => console.log('分享失敗:', err));
    } else {
      // 複製到剪貼簿
      navigator.clipboard.writeText(shareUrl).then(() => {
        alert(`房間連結已複製到剪貼簿！\n${shareUrl}`);
      }).catch(() => {
        prompt('複製此連結分享給朋友:', shareUrl);
      });
    }
  }

  leaveRoom() {
    // 關閉語音通話
    if (this.voiceChat) {
      this.voiceChat.disconnect();
      this.voiceChat = null;
    }
    if (this.voiceChatUI) {
      this.voiceChatUI.destroy();
      this.voiceChatUI = null;
    }

    // 關閉 WebSocket 連線
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }

    // 清除遊戲狀態
    this.roomId = null;
    this.lastPlayerCount = 0;

    // 銷毀遊戲實體
    if (this.game) {
      this.game.destroy();
      this.game = null;
    }

    // 銷毀 Pixi 應用
    if (this.app) {
      this.app.destroy(true);
      this.app = null;
    }

    // 隱藏遊戲畫面，顯示房間選擇畫面
    document.getElementById('game-container').classList.add('hidden');
    document.getElementById('room-info').classList.add('hidden');
    document.getElementById('room-screen').classList.remove('hidden');

    console.log('已離開房間');
  }
}

// 啟動應用
new MahjongApp();