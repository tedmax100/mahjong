# 音效系統實作說明

## 概述
已完成音效系統的實作，包含語音、音效和背景音樂的播放功能。

## 已實作功能

### 1. AudioManager 音效管理器
**檔案位置**: `client/src/game/AudioManager.js`

**主要功能**:
- 音效播放管理
- 音量控制（主音量、語音、音效、背景音樂）
- 音效啟用/停用切換
- 玩家語音性別設定

**主要方法**:
- `playTileVoice(playerId, tile)` - 播放牌的語音
- `playActionVoice(playerId, action)` - 播放動作語音
- `playEffect(effect)` - 播放音效
- `playBGM(scene, loop)` - 播放背景音樂
- `playButtonSound()` - 播放按鈕音效
- `setVolume(type, value)` - 設定音量
- `toggle()` - 切換音效啟用狀態

### 2. 已整合的音效播放時機

#### 背景音樂
- **菜單音樂** (`bg_menu.mp3`): 遊戲初始化時播放
- **遊戲音樂** (`bg_game.mp3`): 遊戲開始時播放

#### 遊戲音效
- **骰子音效** (`effect_dice.ogg`): 遊戲開始時播放
- **發牌音效** (`effect_deal.ogg`): 發牌時播放
- **勝利音效** (`effect_win.ogg`): 胡牌時播放
- **失敗音效** (`effect_lose.ogg`): 流局時播放
- **按鈕音效** (`btn.ogg`): 點擊動作按鈕時播放

#### 牌的語音
**播放時機**: 玩家打牌時
**檔案命名**: `{性別}_{牌類型}.ogg`

**範例**:
- 打出一萬: `boy_wan-1.ogg` 或 `girl_wan-1.ogg`
- 打出五筒: `boy_tong-5.ogg` 或 `girl_tong-5.ogg`
- 打出東風: `boy_dong.ogg` 或 `girl_dong.ogg`

**實作位置**: `Game.js` 的 `handleDiscard()` 方法

#### 動作語音
**播放時機**: 玩家執行動作時
**檔案命名**: `{性別}_action_{動作}.ogg`

**已實作的動作**:
1. **吃牌** (`chi`) - `handleChow()` 方法
2. **碰牌** (`peng`) - `handlePong()` 方法
3. **槓牌** (`gang`) - `handleKong()` 方法
4. **聽牌** (`ting`) - `handleTing()` 方法
5. **胡牌** (`hu`) - `handleGameWin()` 方法

### 3. 音量設定
預設音量配置:
```javascript
{
  master: 0.7,    // 主音量 70%
  voice: 1.0,     // 語音音量 100%
  effect: 0.8,    // 音效音量 80%
  music: 0.4      // 背景音樂音量 40%
}
```

### 4. 玩家語音性別設定
- 預設使用 `boy` 語音
- 可透過 `audioManager.setPlayerGender(playerId, gender)` 設定玩家語音性別
- 支援 `'boy'` 和 `'girl'` 兩種語音

## 音效檔案對應表

### 牌的語音 (54 個檔案 × 2 性別)
| 牌類型 | 男生語音 | 女生語音 |
|--------|---------|---------|
| 萬子 1-9 | `boy_wan-1.ogg` ~ `boy_wan-9.ogg` | `girl_wan-1.ogg` ~ `girl_wan-9.ogg` |
| 筒子 1-9 | `boy_tong-1.ogg` ~ `boy_tong-9.ogg` | `girl_tong-1.ogg` ~ `girl_tong-9.ogg` |
| 條子 1-9 | `boy_tiao-1.ogg` ~ `boy_tiao-9.ogg` | `girl_tiao-1.ogg` ~ `girl_tiao-9.ogg` |
| 東南西北 | `boy_dong.ogg`, `boy_nan.ogg`, `boy_xi.ogg`, `boy_bei.ogg` | `girl_dong.ogg`, `girl_nan.ogg`, `girl_xi.ogg`, `girl_bei.ogg` |
| 中發白 | `boy_zhong.ogg`, `boy_fa.ogg`, `boy_bai.ogg` | `girl_zhong.ogg`, `girl_fa.ogg`, `girl_bai.ogg` |

### 動作語音 (5 個動作 × 2 性別)
| 動作 | 男生語音 | 女生語音 |
|------|---------|---------|
| 吃 | `boy_action_chi.ogg` | `girl_action_chi.ogg` |
| 碰 | `boy_action_peng.ogg` | `girl_action_peng.ogg` |
| 槓 | `boy_action_gang.ogg` | `girl_action_gang.ogg` |
| 胡 | `boy_action_hu.ogg` | `girl_action_hu.ogg` |
| 聽 | `boy_action_ting.ogg` | `girl_action_ting.ogg` |

### 音效 (7 個檔案)
| 音效 | 檔案名稱 | 說明 |
|------|---------|------|
| 時鐘 | `effect_clock.ogg` | 倒計時音效 |
| 金幣 | `effect_coins.ogg` | 得分音效 |
| 骰子 | `effect_dice.ogg` | 遊戲開始音效 |
| 失敗 | `effect_lose.ogg` | 流局音效 |
| 勝利 | `effect_win.ogg` | 胡牌音效 |
| 補花 | `effect_buhua.ogg` | 補花音效 |
| 發牌 | `effect_deal.ogg` | 發牌音效 |

### 背景音樂 (2 個檔案)
| 場景 | 檔案名稱 |
|------|---------|
| 菜單 | `bg_menu.mp3` |
| 遊戲 | `bg_game.mp3` |

### 按鈕音效 (1 個檔案)
| 音效 | 檔案名稱 |
|------|---------|
| 按鈕點擊 | `btn.ogg` |

## 使用方式

### 在 Game.js 中使用音效
```javascript
// AudioManager 已在 Game 類別中初始化
// this.audioManager = new AudioManager();

// 播放牌的語音
this.audioManager.playTileVoice(playerId, 'wan-1');

// 播放動作語音
this.audioManager.playActionVoice(playerId, 'peng');

// 播放音效
this.audioManager.playEffect('win');

// 播放背景音樂
this.audioManager.playBGM('game');

// 播放按鈕音效
this.audioManager.playButtonSound();

// 設定玩家語音性別
this.audioManager.setPlayerGender(playerId, 'girl');

// 調整音量
this.audioManager.setVolume('music', 0.5); // 50%

// 停用/啟用音效
this.audioManager.toggle();
```

### 擴展功能

#### 添加音量控制 UI
可以在設定界面添加音量控制滑桿：
```javascript
// 主音量控制
volumeSlider.addEventListener('input', (e) => {
  game.audioManager.setVolume('master', e.target.value / 100);
});

// 背景音樂音量控制
musicSlider.addEventListener('input', (e) => {
  game.audioManager.setVolume('music', e.target.value / 100);
});
```

#### 添加音效開關按鈕
```javascript
toggleButton.addEventListener('click', () => {
  const enabled = game.audioManager.toggle();
  toggleButton.textContent = enabled ? '🔊 音效開啟' : '🔇 音效關閉';
});
```

#### 設定玩家語音偏好
```javascript
// 在玩家設定中添加語音性別選項
genderSelect.addEventListener('change', (e) => {
  const myPlayerId = game.players[game.myPosition].userId;
  game.audioManager.setPlayerGender(myPlayerId, e.target.value);
});
```

## 注意事項

1. **瀏覽器自動播放政策**: 部分瀏覽器需要用戶互動後才能播放音效，因此背景音樂可能需要用戶點擊後才能播放。

2. **音效檔案路徑**: 所有音效檔案位於 `client/public/assets/music/` 目錄。

3. **性別語音**: 目前預設使用 `boy` 語音，可以根據需求添加性別選擇功能。

4. **音效緩存**: AudioManager 會自動管理音效播放，不需要手動緩存。

5. **音量調整**: 可以在遊戲中動態調整各類型音量，不會影響已播放的背景音樂。

## 測試建議

1. 開始遊戲時確認背景音樂切換
2. 打牌時確認語音播放
3. 執行吃碰槓胡聽動作時確認語音播放
4. 點擊按鈕時確認按鈕音效
5. 胡牌/流局時確認音效播放
6. 測試音量調整功能
7. 測試音效啟用/停用功能

## 未來擴展建議

1. **音效預載**: 實作音效預載功能，減少首次播放延遲
2. **語音性別選擇**: 添加玩家語音性別選擇 UI
3. **音量控制 UI**: 添加音量控制設定界面
4. **更多音效**: 補花、摸牌等其他動作的音效
5. **語音隨機化**: 每個玩家隨機分配不同語音
6. **自定義語音包**: 支援用戶自定義語音包

## 檔案清單

- `client/src/game/AudioManager.js` - 音效管理器
- `client/src/game/Game.js` - 已整合音效播放
- `client/public/assets/music/` - 音效檔案目錄
- `client/public/assets/music/README.md` - 音效檔案命名說明
