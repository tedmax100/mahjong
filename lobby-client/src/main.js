/**
 * Lobby Client - 大廳前端入口
 * 負責登入、大廳顯示、房間列表和聊天
 */
import { GoogleAuth } from '@shared/auth/GoogleAuth.js';
import { LobbyClient } from './network/LobbyClient.js';

class LobbyApp {
  constructor() {
    this.user = null;
    this.auth = new GoogleAuth();
    this.lobbyClient = null;
    this.authFetch = null;

    // 預設遊戲伺服器 URL（開發環境使用 5175，生產環境使用 8080）
    const isDev = import.meta.env.DEV;
    this.defaultGameServerUrl = import.meta.env.VITE_GAME_SERVER_URL ||
      (isDev ? 'http://localhost:5175' : 'http://localhost:8080');

    this.init();
  }

  init() {
    this.showSplashScreen();
    this.initAuth();
    this.bindEvents();
  }

  /**
   * 顯示啟動畫面
   */
  showSplashScreen() {
    const splash = document.getElementById('splash-screen');
    if (splash) {
      setTimeout(() => {
        splash.classList.add('hidden');
      }, 2000);
    }
  }

  /**
   * 初始化認證
   */
  initAuth() {
    this.auth.onSignIn = (user) => {
      this.user = user;
      this.authFetch = this.auth.createAuthFetch();
      console.log('用戶已登入:', user.name);
      this.showLobbyScreen();
    };

    this.auth.onSignOut = () => {
      this.user = null;
      this.authFetch = null;
      this.showLoginScreen();
    };

    this.auth.init();
  }

  /**
   * 綁定事件
   */
  bindEvents() {
    // 快速開始（訪客模式）
    const quickStartBtn = document.getElementById('quick-start-btn');
    if (quickStartBtn) {
      quickStartBtn.addEventListener('click', () => this.quickStart());
    }

    // 登出
    const lobbyBackBtn = document.getElementById('lobby-back-btn');
    if (lobbyBackBtn) {
      lobbyBackBtn.addEventListener('click', () => this.logout());
    }

    // 創建房間
    const createPublicBtn = document.getElementById('create-public-room-btn');
    if (createPublicBtn) {
      createPublicBtn.addEventListener('click', () => this.createRoom(true));
    }

    const createPrivateBtn = document.getElementById('create-private-room-btn');
    if (createPrivateBtn) {
      createPrivateBtn.addEventListener('click', () => this.createRoom(false));
    }

    // 加入私人房間
    const joinPrivateBtn = document.getElementById('join-private-btn');
    if (joinPrivateBtn) {
      joinPrivateBtn.addEventListener('click', () => {
        const roomId = document.getElementById('private-room-id')?.value.trim();
        if (roomId) {
          this.joinRoom(roomId, this.defaultGameServerUrl);
        }
      });
    }

    // 私人房間輸入框 Enter 鍵
    const privateRoomInput = document.getElementById('private-room-id');
    if (privateRoomInput) {
      privateRoomInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
          const roomId = privateRoomInput.value.trim();
          if (roomId) {
            this.joinRoom(roomId, this.defaultGameServerUrl);
          }
        }
      });
    }

    // 發送聊天
    const sendChatBtn = document.getElementById('send-chat-btn');
    if (sendChatBtn) {
      sendChatBtn.addEventListener('click', () => this.sendChatMessage());
    }

    // 聊天輸入框 Enter 鍵
    const chatInput = document.getElementById('chat-input');
    if (chatInput) {
      chatInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
          this.sendChatMessage();
        }
      });
    }
  }

  /**
   * 訪客模式快速開始
   */
  quickStart() {
    const nameInput = document.getElementById('player-name-input');
    const name = nameInput?.value.trim() || null;

    this.auth.createTestUser(name);
  }

  /**
   * 顯示登入畫面
   */
  showLoginScreen() {
    document.getElementById('login-screen')?.classList.remove('hidden');
    document.getElementById('lobby-screen')?.classList.add('hidden');

    // 關閉大廳連線
    if (this.lobbyClient) {
      this.lobbyClient.close();
      this.lobbyClient = null;
    }
  }

  /**
   * 顯示大廳畫面
   */
  showLobbyScreen() {
    document.getElementById('login-screen')?.classList.add('hidden');
    document.getElementById('lobby-screen')?.classList.remove('hidden');

    // 連接大廳
    this.connectToLobby();
  }

  /**
   * 連接到大廳
   */
  connectToLobby() {
    if (this.lobbyClient) {
      this.lobbyClient.close();
    }

    this.lobbyClient = new LobbyClient(this.user);

    this.lobbyClient.onRoomListUpdate = (rooms) => {
      console.log('[LobbyApp] 收到房間列表更新:', rooms);
      this.updateRoomList(rooms);
    };

    this.lobbyClient.onChatMessage = (message) => {
      this.addChatMessage(message);
    };

    this.lobbyClient.onOnlineCountUpdate = (count) => {
      const countEl = document.getElementById('online-count');
      if (countEl) {
        countEl.textContent = count;
      }
    };

    this.lobbyClient.onError = (error) => {
      console.error('大廳錯誤:', error);
    };

    this.lobbyClient.connect();
  }

  /**
   * 更新房間列表
   */
  updateRoomList(rooms) {
    const roomList = document.getElementById('room-list');
    if (!roomList) return;

    if (!rooms || rooms.length === 0) {
      roomList.innerHTML = `
        <div class="empty-room-list">
          <div class="empty-room-list-icon">🀄</div>
          <div>目前沒有公開房間</div>
          <div style="font-size: 14px; margin-top: 5px;">創建一個新房間開始遊戲吧！</div>
        </div>
      `;
      return;
    }

    roomList.innerHTML = rooms.map(room => {
      const isFull = room.playerCount >= room.maxPlayers;
      const isPlaying = room.status === 'playing';
      const canJoin = !isFull && !isPlaying;
      const serverAddr = room.serverAddr || this.defaultGameServerUrl;

      return `
        <div class="room-item" data-room-id="${this.escapeHtml(room.id)}" data-server-addr="${this.escapeHtml(serverAddr)}">
          <div class="room-item-info">
            <div class="room-name">${this.escapeHtml(room.name || '房間 ' + room.id)}</div>
            <div class="room-host">房主: ${this.escapeHtml(room.hostName || '未知')}</div>
            ${room.isExternal ? `<div class="room-server">🌐 外部伺服器</div>` : ''}
          </div>
          <div class="room-item-status">
            ${room.isExternal ? '<span class="external-badge">外部</span>' : ''}
            <span class="player-count-badge">${room.playerCount}/${room.maxPlayers}</span>
            <button class="join-btn" ${canJoin ? '' : 'disabled'}>
              ${isPlaying ? '遊戲中' : (isFull ? '已滿' : '加入')}
            </button>
          </div>
        </div>
      `;
    }).join('');

    // 綁定加入按鈕事件
    roomList.querySelectorAll('.room-item').forEach(item => {
      const joinBtn = item.querySelector('.join-btn');
      if (joinBtn && !joinBtn.disabled) {
        joinBtn.addEventListener('click', () => {
          const roomId = item.dataset.roomId;
          const serverAddr = item.dataset.serverAddr;
          this.joinRoom(roomId, serverAddr);
        });
      }
    });
  }

  /**
   * 添加聊天訊息
   */
  addChatMessage(message) {
    const chatMessages = document.getElementById('chat-messages');
    if (!chatMessages) return;

    const time = new Date(message.timestamp).toLocaleTimeString('zh-TW', {
      hour: '2-digit',
      minute: '2-digit',
    });

    const msgEl = document.createElement('div');
    msgEl.className = `chat-message ${message.type === 'system' ? 'system' : ''}`;
    msgEl.innerHTML = `
      <span class="chat-time">${time}</span>
      ${message.type !== 'system' ? `<span class="chat-user">${this.escapeHtml(message.userName)}:</span>` : ''}
      <span class="chat-content">${this.escapeHtml(message.content)}</span>
    `;

    chatMessages.appendChild(msgEl);
    chatMessages.scrollTop = chatMessages.scrollHeight;
  }

  /**
   * 發送聊天訊息
   */
  sendChatMessage() {
    const input = document.getElementById('chat-input');
    if (!input) return;

    const content = input.value.trim();
    if (!content) return;

    if (this.lobbyClient && this.lobbyClient.isConnected()) {
      this.lobbyClient.sendChat(content);
      input.value = '';
    }
  }

  /**
   * 創建房間
   */
  async createRoom(isPublic) {
    if (!this.user) {
      alert('請先登入');
      return;
    }

    try {
      const response = await fetch('/api/lobby/rooms', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          userId: this.user.id,
          userName: this.user.name,
          isPublic: isPublic,
        }),
      });

      const data = await response.json();

      if (data.success && data.roomId) {
        const serverAddr = data.serverAddr || this.defaultGameServerUrl;
        this.joinRoom(data.roomId, serverAddr);
      } else {
        alert('創建房間失敗: ' + (data.error || '未知錯誤'));
      }
    } catch (error) {
      console.error('創建房間失敗:', error);
      alert('創建房間失敗: ' + error.message);
    }
  }

  /**
   * 加入房間（跳轉到遊戲頁面）
   */
  joinRoom(roomId, serverAddr) {
    if (!this.user) {
      alert('請先登入');
      return;
    }

    console.log(`加入房間 ${roomId} @ ${serverAddr}`);

    // 關閉大廳連線
    if (this.lobbyClient) {
      this.lobbyClient.close();
    }

    // 構建遊戲 URL
    const gameUrl = new URL('/game', serverAddr);
    gameUrl.searchParams.set('room', roomId);
    gameUrl.searchParams.set('token', this.auth.getToken() || '');
    gameUrl.searchParams.set('userId', this.user.id);
    gameUrl.searchParams.set('userName', this.user.name);

    // 跳轉到遊戲頁面
    window.location.href = gameUrl.toString();
  }

  /**
   * 登出
   */
  logout() {
    this.auth.signOut();
  }

  /**
   * HTML 轉義
   */
  escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
  }
}

// 啟動應用
new LobbyApp();
