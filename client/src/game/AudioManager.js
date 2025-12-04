/**
 * 音效管理器
 * 負責管理所有遊戲音效、語音和背景音樂的播放
 */
export class AudioManager {
  constructor() {
    this.sounds = new Map(); // 緩存已加載的音效
    this.bgm = null; // 當前背景音樂
    this.enabled = true; // 是否啟用音效

    // 音量設定
    this.volume = {
      master: 0.7,    // 主音量
      voice: 1.0,     // 語音音量
      effect: 0.8,    // 音效音量
      music: 0.4      // 背景音樂音量
    };

    // 隨機選擇語音性別（可以之後改成讓玩家選擇）
    this.genderMap = new Map(); // playerId -> gender
    this.defaultGender = 'boy'; // 預設性別

    console.log('🔊 AudioManager 初始化完成');
  }

  /**
   * 設定玩家語音性別
   * @param {string} playerId - 玩家 ID
   * @param {string} gender - 'boy' 或 'girl'
   */
  setPlayerGender(playerId, gender) {
    this.genderMap.set(playerId, gender);
  }

  /**
   * 獲取玩家語音性別
   * @param {string} playerId - 玩家 ID
   * @returns {string} 'boy' 或 'girl'
   */
  getPlayerGender(playerId) {
    return this.genderMap.get(playerId) || this.defaultGender;
  }

  /**
   * 播放牌的語音
   * @param {string} playerId - 玩家 ID
   * @param {string} tile - 牌的類型（如 'wan-1', 'dong'）
   */
  playTileVoice(playerId, tile) {
    if (!this.enabled) return;

    const gender = this.getPlayerGender(playerId);
    const path = `/assets/music/${gender}_${tile}.ogg`;
    this.playSound(path, this.volume.voice);
    console.log(`🎵 播放牌語音: ${path}`);
  }

  /**
   * 播放動作語音
   * @param {string} playerId - 玩家 ID
   * @param {string} action - 動作名稱（如 'chi', 'peng', 'gang', 'hu', 'ting'）
   */
  playActionVoice(playerId, action) {
    if (!this.enabled) return;

    const gender = this.getPlayerGender(playerId);
    const path = `/assets/music/${gender}_action_${action}.ogg`;
    this.playSound(path, this.volume.voice);
    console.log(`🎵 播放動作語音: ${path}`);
  }

  /**
   * 播放音效
   * @param {string} effect - 音效名稱（如 'clock', 'win', 'deal', 'dice'）
   */
  playEffect(effect) {
    if (!this.enabled) return;

    const path = `/assets/music/effect_${effect}.ogg`;
    this.playSound(path, this.volume.effect);
    console.log(`🔔 播放音效: ${path}`);
  }

  /**
   * 播放按鈕音效
   */
  playButtonSound() {
    if (!this.enabled) return;
    this.playSound('/assets/music/btn.ogg', this.volume.effect);
  }

  /**
   * 播放玩家加入音效
   */
  playPlayerJoin() {
    if (!this.enabled) return;
    this.playSound('/assets/music/effect_join.ogg', this.volume.effect);
    console.log('🔔 播放玩家加入音效');
  }

  /**
   * 播放玩家離開/斷線音效
   */
  playPlayerLeft() {
    if (!this.enabled) return;
    this.playSound('/assets/music/effect_player_lost.mp3', this.volume.effect);
    console.log('🔔 播放玩家離開音效');
  }

  /**
   * 播放背景音樂
   * @param {string} scene - 場景名稱（如 'game', 'menu'）
   * @param {boolean} loop - 是否循環播放，預設為 true
   */
  playBGM(scene, loop = true) {
    if (!this.enabled) return;

    const path = `/assets/music/bg_${scene}.mp3`;

    // 如果正在播放相同的背景音樂，不需要重新播放
    if (this.bgm && this.bgm.src.includes(path)) {
      return;
    }

    // 停止目前的背景音樂
    this.stopBGM();

    // 播放新的背景音樂
    try {
      this.bgm = new Audio(path);
      this.bgm.loop = loop;
      this.bgm.volume = this.volume.music * this.volume.master;

      this.bgm.play().catch(err => {
        console.warn('背景音樂播放失敗:', path, err);
      });

      console.log(`🎵 播放背景音樂: ${path}`);
    } catch (error) {
      console.warn('背景音樂載入失敗:', path, error);
    }
  }

  /**
   * 停止背景音樂
   */
  stopBGM() {
    if (this.bgm) {
      this.bgm.pause();
      this.bgm.currentTime = 0;
      this.bgm = null;
    }
  }

  /**
   * 播放音效（內部方法）
   * @param {string} path - 音效檔案路徑
   * @param {number} volumeType - 音量類型（voice 或 effect）
   */
  playSound(path, volumeType) {
    try {
      const audio = new Audio(path);
      audio.volume = volumeType * this.volume.master;

      audio.play().catch(err => {
        // 忽略自動播放被阻止的錯誤（瀏覽器安全政策）
        if (err.name !== 'NotAllowedError') {
          console.warn('音效播放失敗:', path, err);
        }
      });

      // 播放完成後清理
      audio.addEventListener('ended', () => {
        audio.remove();
      });
    } catch (error) {
      console.warn('音效載入失敗:', path, error);
    }
  }

  /**
   * 設定音量
   * @param {string} type - 音量類型（'master', 'voice', 'effect', 'music'）
   * @param {number} value - 音量值（0.0 ~ 1.0）
   */
  setVolume(type, value) {
    if (this.volume.hasOwnProperty(type)) {
      this.volume[type] = Math.max(0, Math.min(1, value));
      console.log(`🔊 設定 ${type} 音量: ${this.volume[type]}`);

      // 更新背景音樂音量
      if ((type === 'music' || type === 'master') && this.bgm) {
        this.bgm.volume = this.volume.music * this.volume.master;
      }
    }
  }

  /**
   * 切換音效啟用狀態
   */
  toggle() {
    this.enabled = !this.enabled;
    console.log(`🔊 音效${this.enabled ? '啟用' : '停用'}`);

    if (!this.enabled) {
      this.stopBGM();
    }

    return this.enabled;
  }

  /**
   * 啟用音效
   */
  enable() {
    this.enabled = true;
    console.log('🔊 音效已啟用');
  }

  /**
   * 停用音效
   */
  disable() {
    this.enabled = false;
    this.stopBGM();
    console.log('🔊 音效已停用');
  }

  /**
   * 停止所有音效
   */
  stopAll() {
    this.stopBGM();
    this.sounds.clear();
    console.log('🔊 所有音效已停止');
  }

  /**
   * 預載音效（可選，用於減少播放延遲）
   * @param {Array<string>} paths - 音效檔案路徑陣列
   */
  async preload(paths) {
    console.log('🔊 預載音效...', paths.length, '個檔案');

    const promises = paths.map(path => {
      return new Promise((resolve) => {
        const audio = new Audio(path);
        audio.addEventListener('canplaythrough', () => {
          this.sounds.set(path, audio);
          resolve();
        }, { once: true });
        audio.addEventListener('error', () => {
          console.warn('預載失敗:', path);
          resolve();
        }, { once: true });
      });
    });

    await Promise.all(promises);
    console.log('🔊 預載完成');
  }
}
