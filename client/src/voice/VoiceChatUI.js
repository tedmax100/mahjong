/**
 * Voice Chat UI Controls
 * HTML-based floating panel for voice chat controls
 */
export class VoiceChatUI {
  /**
   * @param {import('./VoiceChat.js').VoiceChat} voiceChat - VoiceChat instance
   */
  constructor(voiceChat) {
    this.voiceChat = voiceChat;
    this.container = null;
    this.playerControls = new Map(); // peerId -> {indicator, muteBtn, nameSpan}

    // State
    this.isExpanded = false;
    this.isConnected = false;

    // Callback for talking indicator in game
    this.onPlayerTalkingChange = null; // (peerId, isTalking) => void

    this.createUI();
    this.bindEvents();
  }

  /**
   * Create the UI elements
   */
  createUI() {
    // Create main container
    this.container = document.createElement('div');
    this.container.id = 'voice-chat-panel';
    this.container.innerHTML = `
      <style>
        #voice-chat-panel {
          position: absolute;
          top: 70px;
          right: 20px;
          z-index: 200;
          font-family: 'Microsoft JhengHei', 'PingFang TC', Arial, sans-serif;
        }

        .voice-chat-toggle {
          width: 48px;
          height: 48px;
          background: white;
          border-radius: 8px;
          box-shadow: 0 4px 12px rgba(0,0,0,0.15);
          display: flex;
          align-items: center;
          justify-content: center;
          cursor: pointer;
          font-size: 24px;
          transition: all 0.3s ease;
          border: 2px solid transparent;
        }

        .voice-chat-toggle:hover {
          background: #f0f0f0;
          transform: scale(1.05);
        }

        .voice-chat-toggle.active {
          background: #48bb78;
          color: white;
          border-color: #38a169;
        }

        .voice-chat-toggle.muted {
          background: #e53e3e;
          color: white;
          border-color: #c53030;
        }

        .voice-chat-toggle.talking {
          animation: voice-pulse 0.5s infinite alternate;
        }

        @keyframes voice-pulse {
          from { box-shadow: 0 4px 12px rgba(0,0,0,0.15); }
          to { box-shadow: 0 4px 20px rgba(72, 187, 120, 0.6); }
        }

        .voice-chat-content {
          position: absolute;
          top: 0;
          right: 0;
          background: white;
          padding: 16px;
          border-radius: 12px;
          box-shadow: 0 4px 20px rgba(0,0,0,0.15);
          min-width: 240px;
          display: none;
        }

        #voice-chat-panel.expanded .voice-chat-content {
          display: block;
        }

        #voice-chat-panel.expanded .voice-chat-toggle {
          display: none;
        }

        .voice-chat-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 16px;
          padding-bottom: 12px;
          border-bottom: 1px solid #e2e8f0;
        }

        .voice-chat-title {
          font-weight: bold;
          font-size: 16px;
          color: #2d3748;
        }

        .voice-chat-close {
          cursor: pointer;
          font-size: 20px;
          color: #718096;
          padding: 4px;
          line-height: 1;
          transition: color 0.2s;
        }

        .voice-chat-close:hover {
          color: #2d3748;
        }

        .voice-chat-main-controls {
          display: flex;
          gap: 10px;
          margin-bottom: 16px;
        }

        .voice-btn {
          flex: 1;
          padding: 12px 8px;
          border: none;
          border-radius: 8px;
          cursor: pointer;
          font-size: 14px;
          font-weight: 500;
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 6px;
          transition: all 0.2s;
        }

        .voice-btn:disabled {
          opacity: 0.6;
          cursor: not-allowed;
        }

        .voice-btn-mute {
          background: #48bb78;
          color: white;
        }

        .voice-btn-mute:hover:not(:disabled) {
          background: #38a169;
        }

        .voice-btn-mute.muted {
          background: #e53e3e;
        }

        .voice-btn-mute.muted:hover:not(:disabled) {
          background: #c53030;
        }

        .voice-btn-connect {
          background: #667eea;
          color: white;
        }

        .voice-btn-connect:hover:not(:disabled) {
          background: #5a67d8;
        }

        .voice-btn-connect.connected {
          background: #e53e3e;
        }

        .voice-btn-connect.connected:hover:not(:disabled) {
          background: #c53030;
        }

        .voice-player-list {
          max-height: 200px;
          overflow-y: auto;
        }

        .voice-player-item {
          display: flex;
          align-items: center;
          justify-content: space-between;
          padding: 10px 0;
          border-bottom: 1px solid #edf2f7;
        }

        .voice-player-item:last-child {
          border-bottom: none;
        }

        .voice-player-info {
          display: flex;
          align-items: center;
          gap: 10px;
        }

        .voice-indicator {
          width: 12px;
          height: 12px;
          border-radius: 50%;
          background: #cbd5e0;
          transition: all 0.2s;
          flex-shrink: 0;
        }

        .voice-indicator.talking {
          background: #48bb78;
          box-shadow: 0 0 10px rgba(72, 187, 120, 0.8);
          animation: indicator-pulse 0.3s infinite alternate;
        }

        .voice-indicator.connected {
          background: #667eea;
        }

        .voice-indicator.disconnected {
          background: #e53e3e;
        }

        .voice-indicator.connecting {
          background: #ed8936;
          animation: indicator-blink 1s infinite;
        }

        @keyframes indicator-pulse {
          from { transform: scale(1); }
          to { transform: scale(1.2); }
        }

        @keyframes indicator-blink {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.3; }
        }

        .voice-player-name {
          font-size: 14px;
          color: #4a5568;
          max-width: 120px;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }

        .voice-player-mute {
          width: 32px;
          height: 32px;
          border: none;
          border-radius: 6px;
          cursor: pointer;
          font-size: 16px;
          background: #edf2f7;
          transition: all 0.2s;
          display: flex;
          align-items: center;
          justify-content: center;
        }

        .voice-player-mute:hover {
          background: #e2e8f0;
        }

        .voice-player-mute.muted {
          background: #fed7d7;
          color: #c53030;
        }

        .voice-status {
          font-size: 12px;
          color: #718096;
          text-align: center;
          padding: 8px;
          background: #f7fafc;
          border-radius: 6px;
          margin-top: 12px;
        }

        .voice-error {
          background: #fed7d7;
          color: #c53030;
          padding: 12px;
          border-radius: 8px;
          font-size: 13px;
          margin-bottom: 12px;
          display: none;
        }

        .voice-error.visible {
          display: block;
        }

        .voice-permission-hint {
          font-size: 11px;
          color: #a0aec0;
          text-align: center;
          margin-top: 8px;
        }
      </style>

      <div class="voice-chat-toggle" title="語音通話">
        <span class="toggle-icon">🎙️</span>
      </div>

      <div class="voice-chat-content">
        <div class="voice-chat-header">
          <span class="voice-chat-title">語音通話</span>
          <span class="voice-chat-close">✕</span>
        </div>

        <div class="voice-error"></div>

        <div class="voice-chat-main-controls">
          <button class="voice-btn voice-btn-mute" disabled>
            <span class="mute-icon">🎤</span>
            <span class="mute-text">靜音</span>
          </button>
          <button class="voice-btn voice-btn-connect">
            <span class="connect-icon">📞</span>
            <span class="connect-text">連線</span>
          </button>
        </div>

        <div class="voice-player-list"></div>

        <div class="voice-status">點擊「連線」開始語音通話</div>
        <div class="voice-permission-hint">需要麥克風權限</div>
      </div>
    `;

    document.body.appendChild(this.container);

    // Cache DOM references
    this.toggleBtn = this.container.querySelector('.voice-chat-toggle');
    this.toggleIcon = this.container.querySelector('.toggle-icon');
    this.content = this.container.querySelector('.voice-chat-content');
    this.closeBtn = this.container.querySelector('.voice-chat-close');
    this.muteBtn = this.container.querySelector('.voice-btn-mute');
    this.muteIcon = this.container.querySelector('.mute-icon');
    this.muteText = this.container.querySelector('.mute-text');
    this.connectBtn = this.container.querySelector('.voice-btn-connect');
    this.connectIcon = this.container.querySelector('.connect-icon');
    this.connectText = this.container.querySelector('.connect-text');
    this.playerList = this.container.querySelector('.voice-player-list');
    this.statusText = this.container.querySelector('.voice-status');
    this.errorDiv = this.container.querySelector('.voice-error');
    this.permissionHint = this.container.querySelector('.voice-permission-hint');
  }

  /**
   * Bind event handlers
   */
  bindEvents() {
    // Toggle panel
    this.toggleBtn.addEventListener('click', () => this.togglePanel());
    this.closeBtn.addEventListener('click', () => this.togglePanel());

    // Mute button
    this.muteBtn.addEventListener('click', () => this.handleMuteClick());

    // Connect button
    this.connectBtn.addEventListener('click', () => this.handleConnectClick());

    // Voice chat callbacks
    this.voiceChat.onTalkingStateChange = (peerId, isTalking) => {
      this.updateTalkingIndicator(peerId, isTalking);

      // Forward to game for player avatar indicator
      if (this.onPlayerTalkingChange) {
        this.onPlayerTalkingChange(peerId, isTalking);
      }
    };

    this.voiceChat.onConnectionStateChange = (peerId, state) => {
      this.updateConnectionState(peerId, state);
    };

    this.voiceChat.onError = (error) => {
      this.showError(error.message);
    };
  }

  /**
   * Toggle panel expanded state
   */
  togglePanel() {
    this.isExpanded = !this.isExpanded;
    this.container.classList.toggle('expanded', this.isExpanded);

    // Resume audio when user interacts (handle autoplay policy)
    if (this.isConnected) {
      this.voiceChat.resumeAudio();
    }
  }

  /**
   * Handle mute button click
   */
  handleMuteClick() {
    const isMuted = this.voiceChat.toggleMute();
    this.updateMuteUI(isMuted);
  }

  /**
   * Update mute button UI
   * @param {boolean} isMuted
   */
  updateMuteUI(isMuted) {
    this.muteBtn.classList.toggle('muted', isMuted);
    this.muteIcon.textContent = isMuted ? '🔇' : '🎤';
    this.muteText.textContent = isMuted ? '解除靜音' : '靜音';

    // Update toggle button
    this.toggleBtn.classList.toggle('muted', isMuted);
    this.toggleIcon.textContent = isMuted ? '🔇' : '🎙️';
  }

  /**
   * Handle connect button click
   */
  async handleConnectClick() {
    if (!this.isConnected) {
      await this.connect();
    } else {
      this.disconnect();
    }
  }

  /**
   * Connect to voice chat
   */
  async connect() {
    this.connectBtn.disabled = true;
    this.connectText.textContent = '連線中...';
    this.hideError();

    const success = await this.voiceChat.initialize();

    if (success) {
      this.isConnected = true;
      await this.voiceChat.connectToRoom();

      this.muteBtn.disabled = false;
      this.connectBtn.classList.add('connected');
      this.connectIcon.textContent = '📵';
      this.connectText.textContent = '斷線';
      this.statusText.textContent = '語音已連線';
      this.permissionHint.style.display = 'none';

      this.toggleBtn.classList.add('active');

      // Resume audio (handle autoplay policy)
      await this.voiceChat.resumeAudio();
    } else {
      this.connectText.textContent = '連線';
    }

    this.connectBtn.disabled = false;
  }

  /**
   * Disconnect from voice chat
   */
  disconnect() {
    this.voiceChat.disconnect();
    this.isConnected = false;

    this.muteBtn.disabled = true;
    this.updateMuteUI(false);

    this.connectBtn.classList.remove('connected');
    this.connectIcon.textContent = '📞';
    this.connectText.textContent = '連線';
    this.statusText.textContent = '點擊「連線」開始語音通話';
    this.permissionHint.style.display = 'block';

    this.toggleBtn.classList.remove('active', 'talking');

    // Reset all player indicators
    for (const [peerId, controls] of this.playerControls) {
      controls.indicator.classList.remove('talking', 'connected', 'disconnected', 'connecting');
    }
  }

  /**
   * Update player list
   * @param {Array<{id: string, name: string}>} players
   */
  updatePlayerList(players) {
    this.playerList.innerHTML = '';
    this.playerControls.clear();

    for (const player of players) {
      const item = document.createElement('div');
      item.className = 'voice-player-item';
      item.dataset.peerId = player.id;

      item.innerHTML = `
        <div class="voice-player-info">
          <span class="voice-indicator"></span>
          <span class="voice-player-name" title="${player.name}">${player.name}</span>
        </div>
        <button class="voice-player-mute" title="靜音 ${player.name}">🔊</button>
      `;

      const indicator = item.querySelector('.voice-indicator');
      const muteBtn = item.querySelector('.voice-player-mute');
      const nameSpan = item.querySelector('.voice-player-name');

      muteBtn.addEventListener('click', () => {
        const isMuted = this.voiceChat.togglePeerMute(player.id);
        muteBtn.classList.toggle('muted', isMuted);
        muteBtn.textContent = isMuted ? '🔇' : '🔊';
        muteBtn.title = isMuted ? `解除靜音 ${player.name}` : `靜音 ${player.name}`;
      });

      this.playerList.appendChild(item);
      this.playerControls.set(player.id, { indicator, muteBtn, nameSpan });

      // Update connection state if already connected
      if (this.isConnected) {
        const state = this.voiceChat.getConnectionState(player.id);
        if (state) {
          this.updateConnectionState(player.id, state);
        }
      }
    }
  }

  /**
   * Update talking indicator for a player
   * @param {string} peerId
   * @param {boolean} isTalking
   */
  updateTalkingIndicator(peerId, isTalking) {
    // Update in player list
    const controls = this.playerControls.get(peerId);
    if (controls) {
      if (isTalking) {
        controls.indicator.classList.add('talking');
      } else {
        controls.indicator.classList.remove('talking');
      }
    }

    // Update toggle button for self
    if (peerId === this.voiceChat.myUserId) {
      this.toggleBtn.classList.toggle('talking', isTalking && !this.voiceChat.isMuted);
    }
  }

  /**
   * Update connection state for a player
   * @param {string} peerId
   * @param {string} state
   */
  updateConnectionState(peerId, state) {
    const controls = this.playerControls.get(peerId);
    if (!controls) return;

    const { indicator } = controls;

    // Remove all state classes
    indicator.classList.remove('connected', 'disconnected', 'connecting', 'talking');

    switch (state) {
      case 'connected':
        indicator.classList.add('connected');
        break;
      case 'connecting':
      case 'new':
        indicator.classList.add('connecting');
        break;
      case 'disconnected':
      case 'failed':
      case 'closed':
        indicator.classList.add('disconnected');
        break;
    }

    // Update overall status
    this.updateStatusText();
  }

  /**
   * Update status text based on connection states
   */
  updateStatusText() {
    if (!this.isConnected) return;

    const connectedCount = this.voiceChat.getConnectedPeers().length;
    const totalPeers = this.voiceChat.roomPlayers.length;

    if (connectedCount === totalPeers && totalPeers > 0) {
      this.statusText.textContent = `已與 ${connectedCount} 位玩家連線`;
    } else if (connectedCount > 0) {
      this.statusText.textContent = `已連線 ${connectedCount}/${totalPeers} 位玩家`;
    } else {
      this.statusText.textContent = '等待其他玩家連線...';
    }
  }

  /**
   * Show error message
   * @param {string} message
   */
  showError(message) {
    this.errorDiv.textContent = message;
    this.errorDiv.classList.add('visible');

    // Auto-hide after 5 seconds
    setTimeout(() => {
      this.hideError();
    }, 5000);
  }

  /**
   * Hide error message
   */
  hideError() {
    this.errorDiv.classList.remove('visible');
  }

  /**
   * Hide the UI panel
   */
  hide() {
    this.container.style.display = 'none';
  }

  /**
   * Show the UI panel
   */
  show() {
    this.container.style.display = 'block';
  }

  /**
   * Cleanup and remove UI
   */
  destroy() {
    if (this.container && this.container.parentNode) {
      this.container.remove();
    }
    this.playerControls.clear();
  }
}
