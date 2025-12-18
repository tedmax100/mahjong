/**
 * WebSocket 客戶端
 */
export class WebSocketClient {
  constructor(roomId, user) {
    this.roomId = roomId;
    this.user = user;
    this.ws = null;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.onMessage = null;
    this.onClose = null;
    this.onError = null;

    this.connect();
  }

  connect() {
    const wsUrl = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws?room=${this.roomId}&userId=${this.user.id}&userName=${encodeURIComponent(this.user.name)}`;

    console.log('連接WebSocket:', wsUrl);

    this.ws = new WebSocket(wsUrl);

    this.ws.onopen = () => {
      console.log('WebSocket已連接');
      this.reconnectAttempts = 0;

      // 發送加入房間訊息
      this.send({
        type: 'join',
        roomId: this.roomId,
        userId: this.user.id,
        userName: this.user.name,
        userPicture: this.user.picture
      });

      // [TEMP] 偵測並發送 IP - 移除時刪除這段
      this._sendClientIP();
    };

    this.ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        console.log('收到訊息:', message);

        if (this.onMessage) {
          this.onMessage(message);
        }
      } catch (error) {
        console.error('解析訊息失敗:', error);
      }
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket錯誤:', error);
      if (this.onError) {
        this.onError(error);
      }
    };

    this.ws.onclose = () => {
      console.log('WebSocket已斷開');

      // 嘗試重連
      if (this.reconnectAttempts < this.maxReconnectAttempts) {
        this.reconnectAttempts++;
        console.log(`嘗試重連 (${this.reconnectAttempts}/${this.maxReconnectAttempts})...`);
        setTimeout(() => this.connect(), 2000 * this.reconnectAttempts);
      } else {
        console.error('重連失敗，已達到最大嘗試次數');
        if (this.onClose) {
          this.onClose();
        }
      }
    };
  }

  send(message) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    } else {
      console.error('WebSocket未連接，無法發送訊息');
    }
  }

  /**
   * 發送遊戲動作
   */
  sendAction(action, data) {
    this.send({
      type: 'action',
      action,
      data,
      userId: this.user.id,
      roomId: this.roomId
    });
  }

  /**
   * 關閉連線
   */
  close() {
    this.reconnectAttempts = this.maxReconnectAttempts; // 防止自動重連
    if (this.ws) {
      this.ws.close();
    }
  }

  // ============================================================
  // [TEMP] IP 偵測功能 - 之後要移除時，刪除這整段到檔案結尾
  // ============================================================

  /**
   * [TEMP] 偵測並發送客戶端 IP 到後端
   */
  async _sendClientIP() {
    try {
      // 使用公開 API 取得 IP
      const response = await fetch('https://api.ipify.org?format=json');
      const result = await response.json();

      if (result.ip) {
        this.send({
          type: 'client_ip',
          data: {
            userId: this.user.id,
            ip: result.ip
          }
        });
        console.log('[TEMP] 已發送客戶端 IP:', result.ip);
      }
    } catch (error) {
      console.warn('[TEMP] 無法偵測 IP:', error.message);
    }
  }
  // [TEMP] END - IP 偵測功能
}