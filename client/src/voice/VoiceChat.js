/**
 * WebRTC Voice Chat Manager
 * Manages P2P voice connections between 4 mahjong players using Mesh topology
 */
export class VoiceChat {
  constructor() {
    // Peer connections map: peerId -> RTCPeerConnection
    this.peerConnections = new Map();

    // Remote audio streams map: peerId -> MediaStream
    this.remoteStreams = new Map();

    // Local microphone stream
    this.localStream = null;

    // Audio elements map: peerId -> HTMLAudioElement
    this.audioElements = new Map();

    // Volume analyzers for talking detection
    this.analyzers = new Map(); // peerId -> AnalyserNode
    this.localAnalyzer = null;
    this.audioContext = null;

    // State
    this.isMuted = false;
    this.mutedPeers = new Set(); // Locally muted peers
    this.isEnabled = false;

    // WebSocket reference (set via setWebSocket)
    this.ws = null;
    this.myUserId = null;
    this.roomPlayers = []; // Array of {id, name}

    // Callbacks
    this.onTalkingStateChange = null; // (peerId, isTalking) => void
    this.onConnectionStateChange = null; // (peerId, state) => void
    this.onError = null; // (error) => void

    // Talking state tracking
    this.talkingStates = new Map(); // peerId -> boolean

    // STUN server configuration (Google public servers)
    this.rtcConfig = {
      iceServers: [
        { urls: 'stun:stun.l.google.com:19302' },
        { urls: 'stun:stun1.l.google.com:19302' },
        { urls: 'stun:stun2.l.google.com:19302' }
      ]
    };

    // Audio constraints (audio only, no video)
    this.audioConstraints = {
      audio: {
        echoCancellation: true,
        noiseSuppression: true,
        autoGainControl: true,
        sampleRate: 48000
      },
      video: false
    };

    // Volume detection settings
    this.volumeThreshold = 0.02; // Threshold for "talking" detection
    this.volumeCheckInterval = 100; // ms
    this.volumeCheckTimer = null;

    // Reconnection settings
    this.maxReconnectAttempts = 3;
    this.reconnectAttempts = new Map(); // peerId -> attempt count

    // Autoplay unlock handler
    this.autoplayUnlocked = false;
    this.autoplayUnlockHandler = null;
  }

  /**
   * Set WebSocket reference for signaling
   * @param {WebSocketClient} ws - WebSocket client instance
   * @param {string} myUserId - Current user's ID
   */
  setWebSocket(ws, myUserId) {
    this.ws = ws;
    this.myUserId = myUserId;
  }

  /**
   * Update room players list
   * @param {Array<{id: string, name: string}>} players - Other players in the room
   */
  setRoomPlayers(players) {
    this.roomPlayers = players.filter(p => p.id !== this.myUserId);
  }

  /**
   * Initialize voice chat - request microphone permission
   * @returns {Promise<boolean>} Success status
   */
  async initialize() {
    try {
      console.log('[VoiceChat] Requesting microphone permission...');

      this.localStream = await navigator.mediaDevices.getUserMedia(
        this.audioConstraints
      );

      console.log('[VoiceChat] Microphone access granted');

      // Setup AudioContext for volume analysis
      this.audioContext = new (window.AudioContext || window.webkitAudioContext)();

      // Setup local volume analyzer
      this.setupLocalAnalyzer();

      // Setup autoplay unlock handler
      this.setupAutoplayUnlock();

      this.isEnabled = true;
      return true;
    } catch (error) {
      console.error('[VoiceChat] Failed to access microphone:', error);

      let errorType = 'unknown';
      let errorMessage = error.message;

      if (error.name === 'NotAllowedError') {
        errorType = 'permission_denied';
        errorMessage = '麥克風權限被拒絕';
      } else if (error.name === 'NotFoundError') {
        errorType = 'no_device';
        errorMessage = '找不到麥克風設備';
      } else if (error.name === 'NotReadableError') {
        errorType = 'device_in_use';
        errorMessage = '麥克風正在被其他應用程式使用';
      }

      if (this.onError) {
        this.onError({ type: errorType, message: errorMessage });
      }

      return false;
    }
  }

  /**
   * Setup AudioContext analyzer for local stream volume detection
   */
  setupLocalAnalyzer() {
    if (!this.localStream || !this.audioContext) return;

    const source = this.audioContext.createMediaStreamSource(this.localStream);
    this.localAnalyzer = this.audioContext.createAnalyser();
    this.localAnalyzer.fftSize = 256;
    source.connect(this.localAnalyzer);

    // Start volume monitoring
    this.startVolumeMonitoring();
  }

  /**
   * Setup handler to unlock autoplay on user interaction
   */
  setupAutoplayUnlock() {
    if (this.autoplayUnlocked) return;

    this.autoplayUnlockHandler = async () => {
      if (this.autoplayUnlocked) return;

      console.log('[VoiceChat] User interaction detected, unlocking autoplay...');

      // Resume AudioContext
      if (this.audioContext && this.audioContext.state === 'suspended') {
        try {
          await this.audioContext.resume();
          console.log('[VoiceChat] AudioContext resumed via user interaction');
        } catch (e) {
          console.warn('[VoiceChat] Failed to resume AudioContext:', e);
        }
      }

      // Try to play all audio elements
      for (const [peerId, audio] of this.audioElements) {
        if (audio.paused) {
          try {
            await audio.play();
            console.log(`[VoiceChat] Audio resumed for ${peerId} via user interaction`);
          } catch (e) {
            console.warn(`[VoiceChat] Still cannot play audio for ${peerId}:`, e);
          }
        }
      }

      this.autoplayUnlocked = true;

      // Remove listeners after first successful interaction
      document.removeEventListener('click', this.autoplayUnlockHandler);
      document.removeEventListener('touchstart', this.autoplayUnlockHandler);
      document.removeEventListener('keydown', this.autoplayUnlockHandler);
    };

    // Add listeners for user interaction
    document.addEventListener('click', this.autoplayUnlockHandler, { once: false });
    document.addEventListener('touchstart', this.autoplayUnlockHandler, { once: false });
    document.addEventListener('keydown', this.autoplayUnlockHandler, { once: false });

    console.log('[VoiceChat] Autoplay unlock handler registered');
  }

  /**
   * Connect to all peers in the room
   */
  async connectToRoom() {
    if (!this.localStream) {
      console.warn('[VoiceChat] Cannot connect: no local stream');
      return;
    }

    console.log('[VoiceChat] Connecting to room, peers:', this.roomPlayers.map(p => p.id));

    for (const player of this.roomPlayers) {
      // Use ID comparison to determine who initiates
      // The peer with the "greater" ID sends the offer to avoid collision
      if (this.myUserId > player.id) {
        console.log(`[VoiceChat] I will initiate connection to ${player.id}`);
        await this.createOffer(player.id);
      } else {
        console.log(`[VoiceChat] Waiting for ${player.id} to initiate connection`);
      }
    }
  }

  /**
   * Create and send an offer to a peer
   * @param {string} peerId - Target peer ID
   */
  async createOffer(peerId) {
    const pc = this.createPeerConnection(peerId);

    try {
      const offer = await pc.createOffer();
      await pc.setLocalDescription(offer);

      this.sendSignal(peerId, 'offer', {
        sdp: offer.sdp,
        type: offer.type
      });

      console.log(`[VoiceChat] Sent offer to ${peerId}`);
    } catch (error) {
      console.error(`[VoiceChat] Failed to create offer for ${peerId}:`, error);
    }
  }

  /**
   * Create a new RTCPeerConnection for a peer
   * @param {string} peerId - Peer ID
   * @returns {RTCPeerConnection}
   */
  createPeerConnection(peerId) {
    // Close existing connection if any
    if (this.peerConnections.has(peerId)) {
      console.log(`[VoiceChat] Closing existing connection to ${peerId}`);
      this.peerConnections.get(peerId).close();
    }

    const pc = new RTCPeerConnection(this.rtcConfig);
    this.peerConnections.set(peerId, pc);

    // Add local stream tracks to connection
    if (this.localStream) {
      this.localStream.getTracks().forEach(track => {
        pc.addTrack(track, this.localStream);
      });
    }

    // Handle ICE candidates
    pc.onicecandidate = (event) => {
      if (event.candidate) {
        this.sendSignal(peerId, 'ice-candidate', event.candidate.toJSON());
      }
    };

    // Handle ICE connection state changes
    pc.oniceconnectionstatechange = () => {
      console.log(`[VoiceChat] ICE state with ${peerId}: ${pc.iceConnectionState}`);
    };

    // Handle connection state changes
    pc.onconnectionstatechange = () => {
      const state = pc.connectionState;
      console.log(`[VoiceChat] Connection state with ${peerId}: ${state}`);

      if (this.onConnectionStateChange) {
        this.onConnectionStateChange(peerId, state);
      }

      if (state === 'failed' || state === 'disconnected') {
        this.handleConnectionFailure(peerId);
      } else if (state === 'connected') {
        // Reset reconnect attempts on successful connection
        this.reconnectAttempts.set(peerId, 0);
      }
    };

    // Handle incoming remote stream
    pc.ontrack = (event) => {
      console.log(`[VoiceChat] Received track from ${peerId}`);

      const remoteStream = event.streams[0];
      if (remoteStream) {
        this.remoteStreams.set(peerId, remoteStream);

        // Create audio element for playback
        this.createAudioElement(peerId, remoteStream);

        // Setup volume analyzer for remote stream
        this.setupRemoteAnalyzer(peerId, remoteStream);
      }
    };

    return pc;
  }

  /**
   * Handle incoming WebRTC signal from server
   * @param {string} fromId - Sender's user ID
   * @param {string} signalType - Signal type: 'offer', 'answer', 'ice-candidate'
   * @param {object} payload - Signal data
   */
  async handleSignal(fromId, signalType, payload) {
    console.log(`[VoiceChat] Received ${signalType} from ${fromId}`);

    // Ignore signals if voice chat is not enabled
    if (!this.isEnabled && signalType === 'offer') {
      console.log(`[VoiceChat] Voice chat not enabled, ignoring offer from ${fromId}`);
      return;
    }

    switch (signalType) {
      case 'offer':
        await this.handleOffer(fromId, payload);
        break;
      case 'answer':
        await this.handleAnswer(fromId, payload);
        break;
      case 'ice-candidate':
        await this.handleIceCandidate(fromId, payload);
        break;
      default:
        console.warn(`[VoiceChat] Unknown signal type: ${signalType}`);
    }
  }

  /**
   * Handle incoming offer
   * @param {string} fromId - Sender's user ID
   * @param {RTCSessionDescriptionInit} offer - SDP offer
   */
  async handleOffer(fromId, offer) {
    // Initialize if not already done (auto-accept incoming call)
    if (!this.localStream) {
      const success = await this.initialize();
      if (!success) {
        console.error('[VoiceChat] Cannot accept offer: failed to initialize');
        return;
      }
    }

    const pc = this.createPeerConnection(fromId);

    try {
      await pc.setRemoteDescription(new RTCSessionDescription(offer));

      const answer = await pc.createAnswer();
      await pc.setLocalDescription(answer);

      this.sendSignal(fromId, 'answer', {
        sdp: answer.sdp,
        type: answer.type
      });

      console.log(`[VoiceChat] Sent answer to ${fromId}`);
    } catch (error) {
      console.error(`[VoiceChat] Failed to handle offer from ${fromId}:`, error);
    }
  }

  /**
   * Handle incoming answer
   * @param {string} fromId - Sender's user ID
   * @param {RTCSessionDescriptionInit} answer - SDP answer
   */
  async handleAnswer(fromId, answer) {
    const pc = this.peerConnections.get(fromId);
    if (!pc) {
      console.warn(`[VoiceChat] No peer connection for ${fromId}`);
      return;
    }

    try {
      await pc.setRemoteDescription(new RTCSessionDescription(answer));
      console.log(`[VoiceChat] Set remote description for ${fromId}`);
    } catch (error) {
      console.error(`[VoiceChat] Failed to handle answer from ${fromId}:`, error);
    }
  }

  /**
   * Handle incoming ICE candidate
   * @param {string} fromId - Sender's user ID
   * @param {RTCIceCandidateInit} candidate - ICE candidate
   */
  async handleIceCandidate(fromId, candidate) {
    const pc = this.peerConnections.get(fromId);
    if (!pc) {
      console.warn(`[VoiceChat] No peer connection for ${fromId}`);
      return;
    }

    try {
      await pc.addIceCandidate(new RTCIceCandidate(candidate));
    } catch (error) {
      // Ignore errors for candidates that arrive before remote description
      if (pc.remoteDescription) {
        console.error(`[VoiceChat] Failed to add ICE candidate from ${fromId}:`, error);
      }
    }
  }

  /**
   * Send signaling message through WebSocket
   * @param {string} targetId - Target peer ID
   * @param {string} signalType - Signal type
   * @param {object} payload - Signal data
   */
  sendSignal(targetId, signalType, payload) {
    if (!this.ws) {
      console.error('[VoiceChat] WebSocket not available for signaling');
      return;
    }

    const message = {
      type: 'webrtc_signal',
      data: {
        targetId: targetId,
        signalType: signalType,
        payload: payload
      }
    };

    this.ws.send(message);
  }

  /**
   * Create audio element for remote stream playback
   * @param {string} peerId - Peer ID
   * @param {MediaStream} stream - Remote audio stream
   */
  createAudioElement(peerId, stream) {
    // Check if stream has audio tracks
    const audioTracks = stream.getAudioTracks();
    console.log(`[VoiceChat] Creating audio element for ${peerId}, audio tracks: ${audioTracks.length}`);

    if (audioTracks.length === 0) {
      console.warn(`[VoiceChat] No audio tracks in stream from ${peerId}`);
      return;
    }

    // Log track info
    audioTracks.forEach((track, i) => {
      console.log(`[VoiceChat] Audio track ${i}: enabled=${track.enabled}, muted=${track.muted}, readyState=${track.readyState}`);
    });

    // Remove existing element
    if (this.audioElements.has(peerId)) {
      const oldElement = this.audioElements.get(peerId);
      oldElement.srcObject = null;
      oldElement.remove();
      console.log(`[VoiceChat] Removed old audio element for ${peerId}`);
    }

    const audio = document.createElement('audio');
    audio.id = `voice-audio-${peerId}`;
    audio.autoplay = true;
    audio.playsInline = true;
    audio.volume = 1.0;
    audio.srcObject = stream;
    audio.muted = this.mutedPeers.has(peerId);

    // Debug: log audio element state
    audio.onloadedmetadata = () => {
      console.log(`[VoiceChat] Audio metadata loaded for ${peerId}, duration: ${audio.duration}`);
    };

    audio.onplay = () => {
      console.log(`[VoiceChat] Audio playing for ${peerId}`);
    };

    audio.onerror = (e) => {
      console.error(`[VoiceChat] Audio error for ${peerId}:`, e);
    };

    this.audioElements.set(peerId, audio);
    document.body.appendChild(audio);

    // Handle autoplay policy - try to play with user gesture context
    this.tryPlayAudio(audio, peerId);
  }

  /**
   * Try to play audio element with retry
   * @param {HTMLAudioElement} audio
   * @param {string} peerId
   */
  async tryPlayAudio(audio, peerId) {
    // First, ensure AudioContext is resumed
    if (this.audioContext && this.audioContext.state === 'suspended') {
      try {
        await this.audioContext.resume();
        console.log('[VoiceChat] AudioContext resumed');
      } catch (e) {
        console.warn('[VoiceChat] Failed to resume AudioContext:', e);
      }
    }

    // Try to play with retry
    const maxRetries = 3;
    for (let i = 0; i < maxRetries; i++) {
      try {
        await audio.play();
        console.log(`[VoiceChat] Audio playback started for ${peerId}`);
        return;
      } catch (error) {
        console.warn(`[VoiceChat] Autoplay attempt ${i + 1} failed for ${peerId}:`, error.name);
        if (i < maxRetries - 1) {
          await new Promise(resolve => setTimeout(resolve, 500));
        }
      }
    }

    // If all retries failed, notify user
    console.warn(`[VoiceChat] Audio autoplay blocked for ${peerId}, user interaction required`);
    if (this.onError) {
      this.onError({
        type: 'autoplay_blocked',
        message: '請點擊頁面以啟用語音播放'
      });
    }
  }

  /**
   * Setup volume analyzer for remote stream
   * @param {string} peerId - Peer ID
   * @param {MediaStream} stream - Remote audio stream
   */
  setupRemoteAnalyzer(peerId, stream) {
    if (!this.audioContext) return;

    try {
      const source = this.audioContext.createMediaStreamSource(stream);
      const analyzer = this.audioContext.createAnalyser();
      analyzer.fftSize = 256;
      source.connect(analyzer);

      this.analyzers.set(peerId, analyzer);
    } catch (error) {
      console.error(`[VoiceChat] Failed to setup analyzer for ${peerId}:`, error);
    }
  }

  /**
   * Start monitoring volume levels for talking detection
   */
  startVolumeMonitoring() {
    if (this.volumeCheckTimer) {
      clearInterval(this.volumeCheckTimer);
    }

    const dataArray = new Uint8Array(128);

    this.volumeCheckTimer = setInterval(() => {
      // Check local volume (self)
      if (this.localAnalyzer && !this.isMuted) {
        this.localAnalyzer.getByteFrequencyData(dataArray);
        const volume = this.calculateVolume(dataArray);
        const isTalking = volume > this.volumeThreshold;

        this.updateTalkingState(this.myUserId, isTalking);
      } else if (this.isMuted) {
        this.updateTalkingState(this.myUserId, false);
      }

      // Check remote volumes
      for (const [peerId, analyzer] of this.analyzers) {
        analyzer.getByteFrequencyData(dataArray);
        const volume = this.calculateVolume(dataArray);
        const isTalking = volume > this.volumeThreshold;

        this.updateTalkingState(peerId, isTalking);
      }
    }, this.volumeCheckInterval);
  }

  /**
   * Update talking state and trigger callback if changed
   * @param {string} peerId - Peer ID (or self ID)
   * @param {boolean} isTalking - Current talking state
   */
  updateTalkingState(peerId, isTalking) {
    const previousState = this.talkingStates.get(peerId) || false;

    if (previousState !== isTalking) {
      this.talkingStates.set(peerId, isTalking);

      if (this.onTalkingStateChange) {
        this.onTalkingStateChange(peerId, isTalking);
      }
    }
  }

  /**
   * Calculate volume from frequency data
   * @param {Uint8Array} dataArray - Frequency data
   * @returns {number} Normalized volume (0-1)
   */
  calculateVolume(dataArray) {
    let sum = 0;
    for (let i = 0; i < dataArray.length; i++) {
      sum += dataArray[i];
    }
    return sum / (dataArray.length * 255);
  }

  /**
   * Toggle self mute
   * @returns {boolean} New muted state
   */
  toggleMute() {
    this.isMuted = !this.isMuted;

    if (this.localStream) {
      this.localStream.getAudioTracks().forEach(track => {
        track.enabled = !this.isMuted;
      });
    }

    console.log(`[VoiceChat] Mute toggled: ${this.isMuted}`);
    return this.isMuted;
  }

  /**
   * Set mute state directly
   * @param {boolean} muted - Muted state
   */
  setMute(muted) {
    this.isMuted = muted;

    if (this.localStream) {
      this.localStream.getAudioTracks().forEach(track => {
        track.enabled = !this.isMuted;
      });
    }
  }

  /**
   * Toggle mute for a specific peer (local only)
   * @param {string} peerId - Peer ID to mute/unmute
   * @returns {boolean} New muted state for the peer
   */
  togglePeerMute(peerId) {
    const isMuted = this.mutedPeers.has(peerId);

    if (isMuted) {
      this.mutedPeers.delete(peerId);
      if (this.audioElements.has(peerId)) {
        this.audioElements.get(peerId).muted = false;
      }
      console.log(`[VoiceChat] Unmuted peer ${peerId}`);
      return false;
    } else {
      this.mutedPeers.add(peerId);
      if (this.audioElements.has(peerId)) {
        this.audioElements.get(peerId).muted = true;
      }
      console.log(`[VoiceChat] Muted peer ${peerId}`);
      return true;
    }
  }

  /**
   * Check if a peer is locally muted
   * @param {string} peerId - Peer ID
   * @returns {boolean}
   */
  isPeerMuted(peerId) {
    return this.mutedPeers.has(peerId);
  }

  /**
   * Handle connection failure - attempt reconnection
   * @param {string} peerId - Peer ID that failed
   */
  handleConnectionFailure(peerId) {
    const attempts = this.reconnectAttempts.get(peerId) || 0;

    if (attempts >= this.maxReconnectAttempts) {
      console.log(`[VoiceChat] Max reconnect attempts reached for ${peerId}`);
      return;
    }

    this.reconnectAttempts.set(peerId, attempts + 1);
    console.log(`[VoiceChat] Attempting reconnection to ${peerId} (attempt ${attempts + 1})`);

    // Wait before reconnecting
    setTimeout(() => {
      // Only initiator should retry
      if (this.myUserId > peerId) {
        this.createOffer(peerId);
      }
    }, 2000 * (attempts + 1)); // Exponential backoff
  }

  /**
   * Resume audio playback (handle autoplay policy)
   * Call this after user interaction
   */
  async resumeAudio() {
    // Resume AudioContext if suspended
    if (this.audioContext && this.audioContext.state === 'suspended') {
      await this.audioContext.resume();
    }

    // Try to play all audio elements
    for (const [peerId, audio] of this.audioElements) {
      try {
        await audio.play();
      } catch (error) {
        console.warn(`[VoiceChat] Failed to resume audio for ${peerId}:`, error);
      }
    }
  }

  /**
   * Get connection state for a peer
   * @param {string} peerId - Peer ID
   * @returns {string|null} Connection state or null
   */
  getConnectionState(peerId) {
    const pc = this.peerConnections.get(peerId);
    return pc ? pc.connectionState : null;
  }

  /**
   * Check if connected to a peer
   * @param {string} peerId - Peer ID
   * @returns {boolean}
   */
  isConnectedToPeer(peerId) {
    return this.getConnectionState(peerId) === 'connected';
  }

  /**
   * Get all connected peers
   * @returns {string[]} Array of connected peer IDs
   */
  getConnectedPeers() {
    const connected = [];
    for (const [peerId, pc] of this.peerConnections) {
      if (pc.connectionState === 'connected') {
        connected.push(peerId);
      }
    }
    return connected;
  }

  /**
   * Disconnect from a specific peer
   * @param {string} peerId - Peer ID
   */
  disconnectPeer(peerId) {
    // Close peer connection
    if (this.peerConnections.has(peerId)) {
      this.peerConnections.get(peerId).close();
      this.peerConnections.delete(peerId);
    }

    // Remove audio element
    if (this.audioElements.has(peerId)) {
      const audio = this.audioElements.get(peerId);
      audio.srcObject = null;
      audio.remove();
      this.audioElements.delete(peerId);
    }

    // Clear remote stream
    this.remoteStreams.delete(peerId);

    // Clear analyzer
    this.analyzers.delete(peerId);

    // Clear states
    this.talkingStates.delete(peerId);
    this.reconnectAttempts.delete(peerId);

    console.log(`[VoiceChat] Disconnected from peer ${peerId}`);
  }

  /**
   * Disconnect and cleanup all resources
   */
  disconnect() {
    console.log('[VoiceChat] Disconnecting...');

    // Stop volume monitoring
    if (this.volumeCheckTimer) {
      clearInterval(this.volumeCheckTimer);
      this.volumeCheckTimer = null;
    }

    // Close all peer connections
    for (const [peerId, pc] of this.peerConnections) {
      pc.close();
    }
    this.peerConnections.clear();

    // Remove all audio elements
    for (const [peerId, audio] of this.audioElements) {
      audio.srcObject = null;
      audio.remove();
    }
    this.audioElements.clear();

    // Clear remote streams
    this.remoteStreams.clear();

    // Stop local stream
    if (this.localStream) {
      this.localStream.getTracks().forEach(track => track.stop());
      this.localStream = null;
    }

    // Close AudioContext
    if (this.audioContext) {
      this.audioContext.close().catch(() => {});
      this.audioContext = null;
    }

    // Clear analyzers
    this.analyzers.clear();
    this.localAnalyzer = null;

    // Clear states
    this.talkingStates.clear();
    this.reconnectAttempts.clear();
    this.mutedPeers.clear();

    // Remove autoplay unlock handlers
    if (this.autoplayUnlockHandler) {
      document.removeEventListener('click', this.autoplayUnlockHandler);
      document.removeEventListener('touchstart', this.autoplayUnlockHandler);
      document.removeEventListener('keydown', this.autoplayUnlockHandler);
      this.autoplayUnlockHandler = null;
    }
    this.autoplayUnlocked = false;

    this.isEnabled = false;
    this.isMuted = false;

    console.log('[VoiceChat] Disconnected');
  }
}
