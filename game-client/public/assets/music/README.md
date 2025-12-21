# 音樂檔案命名規則

## 檔案命名格式

### 1. 牌的語音檔案
格式：`{角色}_{牌類型}.ogg`

#### 萬子
- `boy_wan-1.ogg` ~ `boy_wan-9.ogg`（男生）
- `girl_wan-1.ogg` ~ `girl_wan-9.ogg`（女生）

#### 筒子
- `boy_tong-1.ogg` ~ `boy_tong-9.ogg`（男生）
- `girl_tong-1.ogg` ~ `girl_tong-9.ogg`（女生）

#### 條子
- `boy_tiao-1.ogg` ~ `boy_tiao-9.ogg`（男生）
- `girl_tiao-1.ogg` ~ `girl_tiao-9.ogg`（女生）

#### 風牌
- `boy_dong.ogg`, `boy_nan.ogg`, `boy_xi.ogg`, `boy_bei.ogg`（男生）
- `girl_dong.ogg`, `girl_nan.ogg`, `girl_xi.ogg`, `girl_bei.ogg`（女生）

#### 三元牌
- `boy_zhong.ogg`, `boy_fa.ogg`, `boy_bai.ogg`（男生）
- `girl_zhong.ogg`, `girl_fa.ogg`, `girl_bai.ogg`（女生）

### 2. 動作語音檔案
格式：`{角色}_action_{動作}.ogg`

- `boy_action_chi.ogg`, `boy_action_peng.ogg`, `boy_action_gang.ogg`, `boy_action_hu.ogg`, `boy_action_ting.ogg`（男生）
- `girl_action_chi.ogg`, `girl_action_peng.ogg`, `girl_action_gang.ogg`, `girl_action_hu.ogg`, `girl_action_ting.ogg`（女生）

### 3. 音效檔案
格式：`effect_{效果名稱}.ogg`

- `effect_clock.ogg` - 倒計時音效
- `effect_coins.ogg` - 得分音效
- `effect_dice.ogg` - 擲骰子音效
- `effect_lose.ogg` - 失敗音效
- `effect_win.ogg` - 勝利音效
- `effect_buhua.ogg` - 補花音效
- `effect_deal.ogg` - 發牌音效
- `effect_buhua_alt.ogg` - 補花替代音效

### 4. 背景音樂
格式：`bg_{場景}.mp3`

- `bg_game.mp3` - 遊戲中背景音樂
- `bg_menu.mp3` - 菜單背景音樂

### 5. 按鈕音效
- `btn.mp3`, `btn.ogg` - 按鈕點擊音效

---

## 程式使用範例

### AudioManager 類別範例

```javascript
/**
 * 音效管理器
 */
export class AudioManager {
  constructor() {
    this.sounds = {};
    this.bgm = null;
    this.volume = {
      master: 1.0,
      voice: 1.0,
      effect: 1.0,
      music: 0.5
    };
  }

  /**
   * 播放牌的語音
   * @param {string} gender - 'boy' 或 'girl'
   * @param {string} tile - 牌的類型（如 'wan-1', 'dong'）
   */
  playTileVoice(gender, tile) {
    const path = `/assets/music/${gender}_${tile}.ogg`;
    this.playSound(path, this.volume.voice);
  }

  /**
   * 播放動作語音
   * @param {string} gender - 'boy' 或 'girl'
   * @param {string} action - 動作名稱（如 'chi', 'peng', 'hu'）
   */
  playActionVoice(gender, action) {
    const path = `/assets/music/${gender}_action_${action}.ogg`;
    this.playSound(path, this.volume.voice);
  }

  /**
   * 播放音效
   * @param {string} effect - 音效名稱（如 'clock', 'win', 'deal'）
   */
  playEffect(effect) {
    const path = `/assets/music/effect_${effect}.ogg`;
    this.playSound(path, this.volume.effect);
  }

  /**
   * 播放背景音樂
   * @param {string} scene - 場景名稱（如 'game', 'menu'）
   */
  playBGM(scene) {
    const path = `/assets/music/bg_${scene}.mp3`;

    // 停止目前的背景音樂
    if (this.bgm) {
      this.bgm.pause();
      this.bgm = null;
    }

    // 播放新的背景音樂
    this.bgm = new Audio(path);
    this.bgm.loop = true;
    this.bgm.volume = this.volume.music * this.volume.master;
    this.bgm.play();
  }

  /**
   * 播放音效（內部方法）
   */
  playSound(path, volume) {
    const audio = new Audio(path);
    audio.volume = volume * this.volume.master;
    audio.play().catch(err => {
      console.warn('音效播放失敗:', path, err);
    });
  }

  /**
   * 設定音量
   */
  setVolume(type, value) {
    if (this.volume.hasOwnProperty(type)) {
      this.volume[type] = Math.max(0, Math.min(1, value));

      // 更新背景音樂音量
      if (type === 'music' || type === 'master') {
        if (this.bgm) {
          this.bgm.volume = this.volume.music * this.volume.master;
        }
      }
    }
  }

  /**
   * 停止所有音效
   */
  stopAll() {
    if (this.bgm) {
      this.bgm.pause();
      this.bgm = null;
    }
  }
}
```

### 在 Game.js 中使用

```javascript
import { AudioManager } from './AudioManager.js';

export class Game {
  constructor(app, ws) {
    // ... 其他初始化代碼 ...

    // 初始化音效管理器
    this.audioManager = new AudioManager();

    // 播放菜單音樂
    this.audioManager.playBGM('menu');
  }

  startGame(data) {
    // 遊戲開始時播放遊戲音樂
    this.audioManager.playBGM('game');

    // 播放發牌音效
    this.audioManager.playEffect('deal');

    // ... 其他遊戲開始邏輯 ...
  }

  handleDiscard(playerId, tile) {
    // 播放打牌語音（可以根據玩家性別選擇）
    const gender = this.getPlayerGender(playerId); // 'boy' 或 'girl'
    this.audioManager.playTileVoice(gender, tile);

    // ... 其他打牌邏輯 ...
  }

  handlePong(playerId, tile) {
    // 播放碰牌語音
    const gender = this.getPlayerGender(playerId);
    this.audioManager.playActionVoice(gender, 'peng');

    // ... 其他碰牌邏輯 ...
  }

  handleGameWin(data) {
    // 播放勝利音效
    this.audioManager.playEffect('win');

    // 播放胡牌語音
    const gender = this.getPlayerGender(data.winnerId);
    this.audioManager.playActionVoice(gender, 'hu');

    // ... 其他胡牌邏輯 ...
  }

  getPlayerGender(playerId) {
    // 根據玩家 ID 或設定返回性別
    // 這裡可以根據玩家的個人設定或隨機選擇
    return 'boy'; // 或 'girl'
  }
}
```

---

## 備份資料夾

`backup_old_audio/` 資料夾包含舊版的音樂檔案，這些檔案已經不再使用。如果確認新命名的檔案正常運作，可以刪除此資料夾。
