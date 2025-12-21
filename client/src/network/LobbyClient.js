/**
 * Lobby WebSocket Client
 * 處理大廳的即時通信（房間列表更新、聊天）
 */
export class LobbyClient {
  constructor(user, lobbyUrl = null) {
    this.user = user;
    this.lobbyUrl = lobbyUrl || this.getLobbyUrl();
    this.ws = null;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectDelay = 2000;
    this.isConnecting = false;

    // 回調函數
    this.onRoomListUpdate = null;
    this.onChatMessage = null;
    this.onOnlineCountUpdate = null;
    this.onError = null;
    this.onConnect = null;
    this.onDisconnect = null;
  }

  /**
   * 獲取 Lobby WebSocket URL
   */
  getLobbyUrl() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.hostname;
    const port = '3001'; // Lobby Service 端口
    return `${protocol}//${host}:${port}`;
  }

  /**
   * 連接到大廳
   */
  connect() {
    if (this.isConnecting || (this.ws && this.ws.readyState === WebSocket.OPEN)) {
      return;
    }

    this.isConnecting = true;

    const wsUrl = `${this.lobbyUrl}/ws/lobby?userId=${encodeURIComponent(this.user.id)}&userName=${encodeURIComponent(this.user.name)}`;

    console.log('[LobbyClient] 連接到:', wsUrl);

    this.ws = new WebSocket(wsUrl);

    this.ws.onopen = () => {
      console.log('[LobbyClient] 已連接');
      this.isConnecting = false;
      this.reconnectAttempts = 0;

      if (this.onConnect) {
        this.onConnect();
      }
    };

    this.ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        this.handleMessage(message);
      } catch (error) {
        console.error('[LobbyClient] 解析訊息失敗:', error);
      }
    };

    this.ws.onclose = (event) => {
      console.log('[LobbyClient] 連接關閉:', event.code, event.reason);
      this.isConnecting = false;

      if (this.onDisconnect) {
        this.onDisconnect();
      }

      // 自動重連
      if (this.reconnectAttempts < this.maxReconnectAttempts) {
        this.scheduleReconnect();
      }
    };

    this.ws.onerror = (error) => {
      console.error('[LobbyClient] WebSocket 錯誤:', error);
      this.isConnecting = false;

      if (this.onError) {
        this.onError('連接錯誤');
      }
    };
  }

  /**
   * 處理收到的訊息
   */
  handleMessage(message) {
    switch (message.type) {
      case 'room_list':
        if (this.onRoomListUpdate && message.data && message.data.rooms) {
          this.onRoomListUpdate(message.data.rooms);
        }
        break;

      case 'room_update':
        if (this.onRoomListUpdate && message.data && message.data.rooms) {
          this.onRoomListUpdate(message.data.rooms);
        }
        break;

      case 'chat_message':
        if (this.onChatMessage) {
          this.onChatMessage(message.data);
        }
        break;

      case 'online_count':
        if (this.onOnlineCountUpdate && message.data) {
          this.onOnlineCountUpdate(message.data.count);
        }
        break;

      case 'error':
        if (this.onError && message.data) {
          this.onError(message.data.message);
        }
        break;

      default:
        console.log('[LobbyClient] 未知訊息類型:', message.type);
    }
  }

  /**
   * 發送聊天訊息
   */
  sendChat(content) {
    if (!content || content.trim() === '') {
      return;
    }

    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({
        type: 'chat',
        data: { content: content.trim() }
      }));
    } else {
      console.warn('[LobbyClient] 無法發送訊息：未連接');
      if (this.onError) {
        this.onError('未連接到大廳');
      }
    }
  }

  /**
   * 請求刷新房間列表
   */
  refreshRoomList() {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({
        type: 'get_rooms',
        data: {}
      }));
    }
  }

  /**
   * 排程重連
   */
  scheduleReconnect() {
    this.reconnectAttempts++;
    const delay = this.reconnectDelay * this.reconnectAttempts;

    console.log(`[LobbyClient] ${delay}ms 後嘗試重連 (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);

    setTimeout(() => {
      if (!this.ws || this.ws.readyState === WebSocket.CLOSED) {
        this.connect();
      }
    }, delay);
  }

  /**
   * 關閉連接
   */
  close() {
    this.reconnectAttempts = this.maxReconnectAttempts; // 防止重連

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  /**
   * 檢查是否已連接
   */
  isConnected() {
    return this.ws && this.ws.readyState === WebSocket.OPEN;
  }
}
