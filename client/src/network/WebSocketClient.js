/**
 * WebSocket 客户端
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

    console.log('连接WebSocket:', wsUrl);

    this.ws = new WebSocket(wsUrl);

    this.ws.onopen = () => {
      console.log('WebSocket已连接');
      this.reconnectAttempts = 0;

      // 发送加入房间消息
      this.send({
        type: 'join',
        roomId: this.roomId,
        userId: this.user.id,
        userName: this.user.name,
        userPicture: this.user.picture
      });
    };

    this.ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        console.log('收到消息:', message);

        if (this.onMessage) {
          this.onMessage(message);
        }
      } catch (error) {
        console.error('解析消息失败:', error);
      }
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket错误:', error);
      if (this.onError) {
        this.onError(error);
      }
    };

    this.ws.onclose = () => {
      console.log('WebSocket已断开');

      // 尝试重连
      if (this.reconnectAttempts < this.maxReconnectAttempts) {
        this.reconnectAttempts++;
        console.log(`尝试重连 (${this.reconnectAttempts}/${this.maxReconnectAttempts})...`);
        setTimeout(() => this.connect(), 2000 * this.reconnectAttempts);
      } else {
        console.error('重连失败，已达到最大尝试次数');
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
      console.error('WebSocket未连接，无法发送消息');
    }
  }

  /**
   * 发送游戏动作
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
   * 关闭连接
   */
  close() {
    this.reconnectAttempts = this.maxReconnectAttempts; // 防止自动重连
    if (this.ws) {
      this.ws.close();
    }
  }
}
