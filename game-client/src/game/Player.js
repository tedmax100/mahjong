import { Container, Text, Graphics } from 'pixi.js';
import { Tile } from './Tile.js';

/**
 * 玩家類
 */
export class Player {
  constructor(id, position, screenWidth, screenHeight) {
    this.id = id;
    this.userId = null; // 玩家的實際ID（從伺服器獲取）
    this.position = position; // 'bottom', 'right', 'top', 'left'
    this.screenWidth = screenWidth;
    this.screenHeight = screenHeight;
    this.container = new Container();
    this.tiles = [];
    this.discardedTiles = [];
    this.melds = []; // 吃/碰/槓的牌組 [{type: 'chow'|'pong'|'kong', tiles: [...]}]
    this.flowers = []; // 花牌 // 吃/碰/槓的牌組 [{type: 'chow'|'pong'|'kong', tiles: [...]}]
    this.meldsContainer = new Container(); // 用於顯示吃/碰/槓牌組的容器
    this.name = '';
    this.score = 1000;
    this.isInteractive = false; // 是否可以互動（輪到自己）
    this.isTing = false; // 是否已宣告聽牌
    this.winningTiles = []; // 聽的牌
    this.lastDrawnTile = null; // 最後摸到的牌（用於聽牌後限制打牌）

    this.infoText = null;
    this.tingStatusText = null; // 聽牌狀態文字
    this.seatWind = 'E'; // 門風: E=東, S=南, W=西, N=北
    this.seatWindText = null; // 門風顯示文字

    // 輪次指示器相關
    this.turnIndicator = null; // 輪次指示器容器
    this.turnIndicatorBg = null; // 高亮邊框
    this.turnIndicatorIcon = null; // 出牌中圖示
    this.isTurnActive = false; // 是否輪到該玩家
    this.pulseAnimation = null; // 脈動動畫計時器

    // 語音通話相關
    this.isTalking = false; // 是否正在說話
    this.talkingIndicator = null; // 說話指示器圖形
    this.talkingPulseTimer = null; // 說話動畫計時器

    // 語音控制按鈕相關
    this.voiceButton = null; // 麥克風按鈕（自己）或靜音按鈕（其他玩家）
    this.voiceButtonBg = null;
    this.voiceButtonIcon = null;
    this.isVoiceMuted = false; // 是否已靜音（自己）或已將對方靜音
    this.isVoiceConnected = false; // 語音是否已連線
    this.isVoiceConnecting = false; // 語音是否正在連線中
    this.onVoiceButtonClick = null; // 按鈕點擊回調（靜音/取消靜音）
    this.onVoiceConnectClick = null; // 連線/斷線點擊回調（僅限底部玩家）

    this.createInfoDisplay();
    this.container.addChild(this.meldsContainer);
  }

  createInfoDisplay() {
    // 輪次指示器容器（包含高亮邊框）
    this.turnIndicator = new Container();

    // 風位顏色對照表
    const windColors = {
      E: 0xE53935, // 東 - 紅色
      S: 0x43A047, // 南 - 綠色
      W: 0x1E88E5, // 西 - 藍色
      N: 0x8E24AA  // 北 - 紫色
    };

    // 根據位置決定資訊條的尺寸和方向
    const isHorizontal = (this.position === 'bottom' || this.position === 'top');
    const barWidth = isHorizontal ? 200 : 50;
    const barHeight = isHorizontal ? 36 : 180;

    // 高亮邊框（輪到該玩家時顯示）- 發光效果
    this.turnIndicatorBg = new Graphics();
    this.turnIndicatorBg.roundRect(-4, -4, barWidth + 8, barHeight + 8, 8);
    this.turnIndicatorBg.stroke({ width: 3, color: 0xFFD700 }); // 金色高亮邊框
    this.turnIndicatorBg.visible = false;
    this.turnIndicator.addChild(this.turnIndicatorBg);

    // 玩家資訊背景（半透明深色底）
    this.infoBg = new Graphics();
    this.infoBg.roundRect(0, 0, barWidth, barHeight, 6);
    this.infoBg.fill({ color: 0x000000, alpha: 0.6 });
    this.turnIndicator.addChild(this.infoBg);

    if (isHorizontal) {
      // 水平布局（上下玩家）：[風位圓標] 玩家名 · 分數

      // 風位圓形徽章
      this.windBadge = new Graphics();
      this.windBadge.circle(18, 18, 14);
      this.windBadge.fill({ color: windColors[this.seatWind] || 0xE53935 });
      this.turnIndicator.addChild(this.windBadge);

      // 風位文字（在圓形徽章內）
      this.seatWindText = new Text({
        text: this.getWindChar(this.seatWind),
        style: {
          fontSize: 16,
          fill: 0xFFFFFF,
          fontWeight: 'bold'
        }
      });
      this.seatWindText.anchor.set(0.5);
      this.seatWindText.x = 18;
      this.seatWindText.y = 18;
      this.turnIndicator.addChild(this.seatWindText);

      // 玩家名稱
      this.nameText = new Text({
        text: this.name || '等待中...',
        style: {
          fontSize: 14,
          fill: 0xFFFFFF,
          fontWeight: 'normal'
        }
      });
      this.nameText.x = 40;
      this.nameText.y = 6;
      this.turnIndicator.addChild(this.nameText);

      // 分數
      this.scoreText = new Text({
        text: `${this.score}`,
        style: {
          fontSize: 13,
          fill: 0xFFD700, // 金黃色
          fontWeight: 'bold'
        }
      });
      this.scoreText.x = 40;
      this.scoreText.y = 20;
      this.turnIndicator.addChild(this.scoreText);

      // 出牌中標籤
      this.turnLabel = new Text({
        text: '出牌中',
        style: {
          fontSize: 11,
          fill: 0x00FF00,
          fontWeight: 'bold'
        }
      });
      this.turnLabel.x = barWidth - 45;
      this.turnLabel.y = 12;
      this.turnLabel.visible = false;
      this.turnIndicator.addChild(this.turnLabel);

    } else {
      // 垂直布局（左右玩家）：風位圓標在上，名稱和分數在下

      // 風位圓形徽章
      this.windBadge = new Graphics();
      this.windBadge.circle(barWidth / 2, 20, 16);
      this.windBadge.fill({ color: windColors[this.seatWind] || 0xE53935 });
      this.turnIndicator.addChild(this.windBadge);

      // 風位文字（在圓形徽章內）
      this.seatWindText = new Text({
        text: this.getWindChar(this.seatWind),
        style: {
          fontSize: 18,
          fill: 0xFFFFFF,
          fontWeight: 'bold'
        }
      });
      this.seatWindText.anchor.set(0.5);
      this.seatWindText.x = barWidth / 2;
      this.seatWindText.y = 20;
      this.turnIndicator.addChild(this.seatWindText);

      // 玩家名稱（垂直排列）
      const displayName = this.truncateName(this.name || '等待中', 4);
      this.nameText = new Text({
        text: displayName.split('').join('\n'),
        style: {
          fontSize: 13,
          fill: 0xFFFFFF,
          fontWeight: 'normal',
          align: 'center',
          lineHeight: 16
        }
      });
      this.nameText.anchor.set(0.5, 0);
      this.nameText.x = barWidth / 2;
      this.nameText.y = 45;
      this.turnIndicator.addChild(this.nameText);

      // 分數
      this.scoreText = new Text({
        text: `${this.score}`,
        style: {
          fontSize: 12,
          fill: 0xFFD700,
          fontWeight: 'bold'
        }
      });
      this.scoreText.anchor.set(0.5);
      this.scoreText.x = barWidth / 2;
      this.scoreText.y = barHeight - 20;
      this.turnIndicator.addChild(this.scoreText);

      // 出牌中標籤（垂直顯示）
      this.turnLabel = new Text({
        text: '▶',
        style: {
          fontSize: 14,
          fill: 0x00FF00,
          fontWeight: 'bold'
        }
      });
      this.turnLabel.anchor.set(0.5);
      this.turnLabel.x = barWidth / 2;
      this.turnLabel.y = barHeight - 5;
      this.turnLabel.visible = false;
      this.turnIndicator.addChild(this.turnLabel);
    }

    // 保留舊的 infoText 引用以相容現有程式碼
    this.infoText = this.nameText;
    this.turnIndicatorIcon = this.turnLabel;

    // 建立說話指示器
    this.createTalkingIndicator(isHorizontal, barWidth, barHeight);

    // 建立語音控制按鈕
    this.createVoiceButton(isHorizontal, barWidth, barHeight);

    // 根據位置設定資訊框位置
    this.positionInfoDisplay(this.turnIndicator);

    this.container.addChild(this.turnIndicator);
  }

  /**
   * 建立說話中指示器（語音通話用）
   */
  createTalkingIndicator(isHorizontal, barWidth, barHeight) {
    this.talkingIndicator = new Graphics();
    this.talkingIndicator.visible = false;

    // 根據布局決定位置
    if (isHorizontal) {
      // 水平布局：在資訊條右側
      this.talkingIndicator.x = barWidth - 18;
      this.talkingIndicator.y = 18;
    } else {
      // 垂直布局：在風位徽章旁邊
      this.talkingIndicator.x = barWidth - 12;
      this.talkingIndicator.y = 20;
    }

    // 繪製初始狀態
    this.drawTalkingIcon(false);

    this.turnIndicator.addChild(this.talkingIndicator);
  }

  /**
   * 繪製說話圖示
   * @param {boolean} isActive - 是否正在說話
   */
  drawTalkingIcon(isActive) {
    if (!this.talkingIndicator) return;

    this.talkingIndicator.clear();

    const color = isActive ? 0x48BB78 : 0x718096; // 綠色 or 灰色
    const alpha = isActive ? 1.0 : 0.5;

    // 繪製小麥克風圓圈
    this.talkingIndicator.circle(0, 0, 6);
    this.talkingIndicator.fill({ color, alpha });

    if (isActive) {
      // 繪製音波效果
      this.talkingIndicator.circle(0, 0, 9);
      this.talkingIndicator.stroke({ width: 1.5, color, alpha: 0.7 });
      this.talkingIndicator.circle(0, 0, 12);
      this.talkingIndicator.stroke({ width: 1.5, color, alpha: 0.4 });
    }
  }

  /**
   * 設定說話狀態（語音通話用）
   * @param {boolean} isTalking - 是否正在說話
   */
  setTalking(isTalking) {
    if (this.isTalking === isTalking) return;

    this.isTalking = isTalking;

    if (!this.talkingIndicator) return;

    this.talkingIndicator.visible = isTalking;
    this.drawTalkingIcon(isTalking);

    // 清除之前的動畫
    if (this.talkingPulseTimer) {
      clearInterval(this.talkingPulseTimer);
      this.talkingPulseTimer = null;
    }

    // 如果正在說話，啟動脈動動畫
    if (isTalking) {
      let scale = 1.0;
      let growing = true;

      this.talkingPulseTimer = setInterval(() => {
        if (growing) {
          scale += 0.05;
          if (scale >= 1.15) growing = false;
        } else {
          scale -= 0.05;
          if (scale <= 1.0) growing = true;
        }
        this.talkingIndicator.scale.set(scale);
      }, 50);
    } else {
      this.talkingIndicator.scale.set(1.0);
    }
  }

  /**
   * 建立語音控制按鈕
   * - 底部玩家（自己）：連線/麥克風開關（始終顯示）
   * - 其他玩家：靜音按鈕（語音連線後才顯示）
   */
  createVoiceButton(isHorizontal, barWidth, barHeight) {
    this.voiceButton = new Container();
    this.voiceButton.eventMode = 'static';
    this.voiceButton.cursor = 'pointer';

    // 按鈕背景
    this.voiceButtonBg = new Graphics();

    // 按鈕圖示文字
    // 底部玩家預設顯示連線圖示，其他玩家顯示喇叭圖示
    this.voiceButtonIcon = new Text({
      text: this.position === 'bottom' ? '📞' : '🔊',
      style: {
        fontSize: isHorizontal ? 16 : 14,
        fill: 0xFFFFFF
      }
    });
    this.voiceButtonIcon.anchor.set(0.5);

    // 根據布局設定按鈕大小和位置
    const buttonSize = isHorizontal ? 28 : 24;

    // 繪製按鈕背景（底部玩家預設為「未連線」狀態）
    if (this.position === 'bottom') {
      this.drawVoiceButtonBgWithState(buttonSize, 'disconnected');
    } else {
      this.drawVoiceButtonBg(buttonSize, false);
    }

    this.voiceButton.addChild(this.voiceButtonBg);
    this.voiceButton.addChild(this.voiceButtonIcon);

    // 設定按鈕位置
    if (isHorizontal) {
      if (this.position === 'bottom') {
        // 底部玩家：麥克風按鈕在名牌右側
        this.voiceButton.x = barWidth + 8;
        this.voiceButton.y = barHeight / 2;
      } else {
        // 頂部玩家：靜音按鈕在名牌右側
        this.voiceButton.x = barWidth + 8;
        this.voiceButton.y = barHeight / 2;
      }
    } else {
      // 垂直布局（左右玩家）：按鈕在名牌下方
      this.voiceButton.x = barWidth / 2;
      this.voiceButton.y = barHeight + 8;
    }

    // 長按計時器（用於斷線功能）
    let longPressTimer = null;
    let isLongPress = false;

    // 按下事件
    this.voiceButton.on('pointerdown', () => {
      isLongPress = false;

      if (this.position === 'bottom' && this.isVoiceConnected) {
        // 底部玩家且已連線：啟動長按計時器（1秒後斷線）
        longPressTimer = setTimeout(() => {
          isLongPress = true;
          // 長按 → 斷線
          if (this.onVoiceConnectClick) {
            this.onVoiceConnectClick(false); // disconnect
          }
        }, 1000);
      }
    });

    // 放開事件
    this.voiceButton.on('pointerup', () => {
      // 清除長按計時器
      if (longPressTimer) {
        clearTimeout(longPressTimer);
        longPressTimer = null;
      }

      // 如果是長按，不執行短按行為
      if (isLongPress) {
        isLongPress = false;
        return;
      }

      if (this.position === 'bottom') {
        // 底部玩家：根據連線狀態決定行為
        if (!this.isVoiceConnected && !this.isVoiceConnecting) {
          // 未連線 → 開始連線
          if (this.onVoiceConnectClick) {
            this.onVoiceConnectClick(true); // connect
          }
        } else if (this.isVoiceConnected) {
          // 已連線 → 切換靜音
          if (this.onVoiceButtonClick) {
            this.onVoiceButtonClick(this.userId, true);
          }
        }
      } else {
        // 其他玩家：切換對方靜音
        if (this.onVoiceButtonClick) {
          this.onVoiceButtonClick(this.userId, false);
        }
      }
    });

    // 離開事件（取消長按）
    this.voiceButton.on('pointerout', () => {
      if (longPressTimer) {
        clearTimeout(longPressTimer);
        longPressTimer = null;
      }
      isLongPress = false;
    });

    // 底部玩家始終顯示按鈕，其他玩家預設隱藏（語音連線後才顯示）
    this.voiceButton.visible = (this.position === 'bottom');

    this.turnIndicator.addChild(this.voiceButton);
  }

  /**
   * 繪製語音按鈕背景
   */
  drawVoiceButtonBg(size, isMuted) {
    if (!this.voiceButtonBg) return;

    this.voiceButtonBg.clear();

    const bgColor = isMuted ? 0xE53E3E : 0x48BB78; // 紅色表示靜音，綠色表示正常
    const alpha = 0.9;

    this.voiceButtonBg.circle(0, 0, size / 2);
    this.voiceButtonBg.fill({ color: bgColor, alpha });
    this.voiceButtonBg.stroke({ width: 2, color: 0xFFFFFF, alpha: 0.5 });
  }

  /**
   * 繪製語音按鈕背景（含連線狀態）
   * @param {number} size - 按鈕大小
   * @param {'disconnected' | 'connecting' | 'connected' | 'muted'} state - 狀態
   */
  drawVoiceButtonBgWithState(size, state) {
    if (!this.voiceButtonBg) return;

    this.voiceButtonBg.clear();

    let bgColor;
    let alpha = 0.9;

    switch (state) {
      case 'disconnected':
        bgColor = 0x667EEA; // 藍紫色 - 未連線
        break;
      case 'connecting':
        bgColor = 0xED8936; // 橙色 - 連線中
        break;
      case 'connected':
        bgColor = 0x48BB78; // 綠色 - 已連線
        break;
      case 'muted':
        bgColor = 0xE53E3E; // 紅色 - 已靜音
        break;
      default:
        bgColor = 0x667EEA;
    }

    this.voiceButtonBg.circle(0, 0, size / 2);
    this.voiceButtonBg.fill({ color: bgColor, alpha });
    this.voiceButtonBg.stroke({ width: 2, color: 0xFFFFFF, alpha: 0.5 });
  }

  /**
   * 設定語音按鈕狀態
   * @param {boolean} isMuted - 是否靜音
   */
  setVoiceMuted(isMuted) {
    this.isVoiceMuted = isMuted;

    if (!this.voiceButtonIcon || !this.voiceButtonBg) return;

    const isHorizontal = (this.position === 'bottom' || this.position === 'top');
    const buttonSize = isHorizontal ? 28 : 24;

    // 更新圖示和背景
    if (this.position === 'bottom') {
      // 自己的麥克風
      if (this.isVoiceConnected) {
        this.voiceButtonIcon.text = isMuted ? '🔇' : '🎤';
        this.drawVoiceButtonBgWithState(buttonSize, isMuted ? 'muted' : 'connected');
      }
      // 未連線時不更新（保持連線圖示）
    } else {
      // 其他玩家的靜音狀態
      this.voiceButtonIcon.text = isMuted ? '🔇' : '🔊';
      this.drawVoiceButtonBg(buttonSize, isMuted);
    }
  }

  /**
   * 設定語音連線狀態（僅限底部玩家）
   * @param {'disconnected' | 'connecting' | 'connected'} state - 連線狀態
   */
  setVoiceConnectionState(state) {
    if (this.position !== 'bottom') return;

    const isHorizontal = true; // 底部玩家是水平布局
    const buttonSize = 28;

    this.isVoiceConnecting = (state === 'connecting');
    this.isVoiceConnected = (state === 'connected');

    if (!this.voiceButtonIcon || !this.voiceButtonBg) return;

    switch (state) {
      case 'disconnected':
        this.voiceButtonIcon.text = '📞';
        this.drawVoiceButtonBgWithState(buttonSize, 'disconnected');
        break;
      case 'connecting':
        this.voiceButtonIcon.text = '⏳';
        this.drawVoiceButtonBgWithState(buttonSize, 'connecting');
        break;
      case 'connected':
        this.voiceButtonIcon.text = this.isVoiceMuted ? '🔇' : '🎤';
        this.drawVoiceButtonBgWithState(buttonSize, this.isVoiceMuted ? 'muted' : 'connected');
        break;
    }
  }

  /**
   * 顯示/隱藏語音按鈕
   * @param {boolean} visible - 是否顯示
   */
  setVoiceButtonVisible(visible) {
    if (this.voiceButton) {
      // 底部玩家的按鈕始終顯示（狀態由 setVoiceConnectionState 控制）
      if (this.position === 'bottom') {
        this.voiceButton.visible = true;
        // 更新連線狀態
        this.setVoiceConnectionState(visible ? 'connected' : 'disconnected');
      } else {
        // 其他玩家：語音連線後才顯示靜音按鈕
        this.voiceButton.visible = visible;
        this.isVoiceConnected = visible;
      }
    }
  }

  /**
   * 取得風位的中文字
   */
  getWindChar(wind) {
    const windChars = { E: '東', S: '南', W: '西', N: '北' };
    return windChars[wind] || '東';
  }

  /**
   * 截斷名稱到指定長度
   */
  truncateName(name, maxLength) {
    if (!name) return '';
    if (name.length <= maxLength) return name;
    return name.substring(0, maxLength);
  }

  positionInfoDisplay(bg) {
    // 資訊條貼在手牌外側
    switch (this.position) {
      case 'bottom':
        // 下方玩家：資訊條在手牌下方（靠近螢幕底部）
        bg.x = this.screenWidth / 2 - 100; // 置中（寬度 200px）
        bg.y = this.screenHeight - 45;     // 靠近底部
        break;
      case 'right':
        // 右側玩家：資訊條在手牌右側（靠近螢幕右邊）
        bg.x = this.screenWidth - 60;      // 靠近右邊（寬度 50px）
        bg.y = this.screenHeight / 2 - 90; // 垂直置中（高度 180px）
        break;
      case 'top':
        // 上方玩家：資訊條貼齊頂端
        bg.x = this.screenWidth / 2 - 100; // 置中（寬度 200px）
        bg.y = 5;                          // 貼齊頂端
        break;
      case 'left':
        // 左側玩家：資訊條在手牌左側（靠近螢幕左邊）
        bg.x = 10;                         // 靠近左邊
        bg.y = this.screenHeight / 2 - 90; // 垂直置中（高度 180px）
        break;
    }
  }

  updateInfo(playerData) {
    this.userId = playerData.id; // 儲存玩家ID
    this.name = playerData.name || '玩家';
    this.score = playerData.score || 1000;

    this.updateNameDisplay();
  }

  /**
   * 更新玩家名稱顯示
   */
  updateNameDisplay() {
    const isHorizontal = (this.position === 'bottom' || this.position === 'top');

    if (this.nameText) {
      if (isHorizontal) {
        // 水平布局：直接顯示名稱
        const displayName = this.name.length > 10 ? this.name.substring(0, 10) + '...' : this.name;
        this.nameText.text = displayName;
      } else {
        // 垂直布局：名稱豎排
        const displayName = this.truncateName(this.name, 4);
        this.nameText.text = displayName.split('').join('\n');
      }
    }

    if (this.scoreText) {
      this.scoreText.text = `${this.score}`;
    }
  }

  /**
   * 設定並顯示門風
   * @param {string} wind - 門風代碼: 'E'=東, 'S'=南, 'W'=西, 'N'=北
   */
  setSeatWind(wind) {
    this.seatWind = wind;

    // 風位顏色對照表
    const windColors = {
      E: 0xE53935, // 東 - 紅色
      S: 0x43A047, // 南 - 綠色
      W: 0x1E88E5, // 西 - 藍色
      N: 0x8E24AA  // 北 - 紫色
    };

    // 更新風位文字
    if (this.seatWindText) {
      this.seatWindText.text = this.getWindChar(wind);
    }

    // 更新風位徽章顏色
    if (this.windBadge) {
      this.windBadge.clear();
      const isHorizontal = (this.position === 'bottom' || this.position === 'top');
      const barWidth = isHorizontal ? 200 : 50;

      if (isHorizontal) {
        this.windBadge.circle(18, 18, 14);
      } else {
        this.windBadge.circle(barWidth / 2, 20, 16);
      }
      this.windBadge.fill({ color: windColors[wind] || 0xE53935 });
    }
  }

  /**
   * 排序手牌（按照臺灣麻將規則：萬 -> 筒 -> 條 -> 字牌 -> 花牌，每組內由小到大）
   */
  sortTiles(tilesData) {
    return tilesData.sort((a, b) => {
      // 定義花色順序
      const getSuitOrder = (tile) => {
        if (tile.startsWith('wan-')) return 1;      // 萬子
        if (tile.startsWith('tong-')) return 2;     // 筒子
        if (tile.startsWith('tiao-')) return 3;     // 條子
        // 字牌
        if (['dong', 'nan', 'xi', 'bei'].includes(tile)) return 4; // 風牌
        if (['zhong', 'fa', 'bai'].includes(tile)) return 5;       // 三元牌
        if (tile.startsWith('flower-')) return 6;   // 花牌
        return 7; // 其他
      };

      // 取得數字（如果有）
      const getNumber = (tile) => {
        const match = tile.match(/-(\d+)$/);
        return match ? parseInt(match[1]) : 0;
      };

      const suitA = getSuitOrder(a);
      const suitB = getSuitOrder(b);

      // 先比較花色
      if (suitA !== suitB) {
        return suitA - suitB;
      }

      // 同花色，比較數字
      return getNumber(a) - getNumber(b);
    });
  }

  async setTiles(tilesData, tileAssets) {
    // 清除現有牌
    this.tiles.forEach(tile => tile.destroy());
    this.tiles = [];

    // 排序手牌
    const sortedTiles = this.sortTiles([...tilesData]);

    // 創建新牌（同步創建，因為牌底紋理已預載入）
    for (const tileType of sortedTiles) {
      const texture = tileAssets[tileType] || tileAssets['back'];
      const tile = new Tile(tileType, texture);

      // 只有底部玩家（自己）的牌可以點擊
      if (this.position === 'bottom') {
        tile.on('click', (clickedTile) => this.onTileClick(clickedTile));
      }

      this.tiles.push(tile);
      this.container.addChild(tile.container);
    }

    // 所有牌都加入後，統一設定位置
    this.tiles.forEach((tile, index) => {
      this.positionTile(tile, index);
    });
  }

  positionTile(tile, index) {
    // 手牌包含牌底，實際佔用空間更大
    const tileWidth = 42.1875 * 0.8;  // 縮小至原來的80%
    const tileHeight = 53.4375 * 0.8; // 縮小至原來的80%

    // 動態調整間距：根據牌的數量和螢幕寬度
    const totalTiles = this.tiles.length;
    let spacing = 5; // 預設間距

    switch (this.position) {
      case 'bottom': {
        // 底部 - 水平排列
        // We now treat hand + melds as a single block and center it.

        const meldsWidth = this.meldsContainer.width;
        const spacingBetweenHandAndMelds = meldsWidth > 0 ? 30 : 0; // Gap between hand and melds

        // Use a fixed spacing for hand tiles to ensure consistency
        const handSpacing = 20;
        const handWidth = totalTiles * tileWidth + (totalTiles > 0 ? (totalTiles - 1) * handSpacing : 0);

        const totalLayoutWidth = handWidth + spacingBetweenHandAndMelds + meldsWidth;
        const startX = (this.screenWidth / 2) - (totalLayoutWidth / 2);

        // Position hand tile - 往下移 30px (從 -180 改為 -150)
        tile.setPosition(
          startX + index * (tileWidth + handSpacing),
          this.screenHeight - 150
        );
        tile.setScale(0.75 * 0.8);

        // Position the entire melds container to the right of the hand
        // This is done for every tile, which is redundant but ensures it's updated correctly
        this.meldsContainer.x = startX + handWidth + spacingBetweenHandAndMelds;
        this.meldsContainer.y = this.screenHeight - 150;
        break;
      }

      case 'right': {
        // 右側 - 垂直排列，超過9張換到第二排（往左）
        spacing = 12;
        const maxPerColumnRight = 9;
        const colRight = Math.floor(index / maxPerColumnRight);
        const rowRight = index % maxPerColumnRight;
        const tilesInThisColRight = Math.min(totalTiles - colRight * maxPerColumnRight, maxPerColumnRight);

        tile.setPosition(
          this.screenWidth - 70 - colRight * (tileWidth + 15),  // 第二排往左移
          this.screenHeight / 2 - (tilesInThisColRight * (tileHeight + spacing)) / 2 + rowRight * (tileHeight + spacing)
        );
        tile.setRotation(Math.PI / 2);
        tile.setScale(0.6 * 0.8);
        break;
      }

      case 'top':
        // 頂部 - 水平排列（背面）- 往上移 (從 30 改為 10)
        spacing = 12; // 增加間距避免重疊
        tile.setPosition(
          this.screenWidth / 2 - (totalTiles * (tileWidth + spacing)) / 2 + index * (tileWidth + spacing),
          10  // 更靠近頂部
        );
        tile.setScale(0.6 * 0.8); // 縮小牌底和牌面至75% (0.8 * 0.75 = 0.6)，再縮小至90%
        break;

      case 'left': {
        // 左側 - 垂直排列，超過9張換到第二排（往右）
        spacing = 12;
        const maxPerColumnLeft = 9;
        const colLeft = Math.floor(index / maxPerColumnLeft);
        const rowLeft = index % maxPerColumnLeft;
        const tilesInThisColLeft = Math.min(totalTiles - colLeft * maxPerColumnLeft, maxPerColumnLeft);

        tile.setPosition(
          30 + colLeft * (tileWidth + 15),  // 第二排往右移
          this.screenHeight / 2 - (tilesInThisColLeft * (tileHeight + spacing)) / 2 + rowLeft * (tileHeight + spacing)
        );
        tile.setRotation(-Math.PI / 2);
        tile.setScale(0.6 * 0.8);
        break;
      }
    }
  }

  onTileClick(tile) {
    console.log('點擊了牌:', tile.type);

    // 只有輪到自己時才能出牌
    if (!this.isInteractive) {
      console.log('還沒輪到你，不能出牌！');
      return;
    }

    // 如果已宣告聽牌，只能打剛摸到的牌
    if (this.isTing) {
      if (tile.type !== this.lastDrawnTile) {
        console.log('已宣告聽牌，只能打剛摸到的牌！');
        // TODO: 顯示提示訊息給玩家
        return;
      }
    }

    // 觸發出牌事件（由Game類處理）
    if (this.onDiscard) {
      this.onDiscard(tile.type);
    }
  }

  setInteractive(interactive) {
    this.isInteractive = interactive;
    console.log(`玩家 ${this.name || this.id} 可互動狀態: ${interactive}`);

    // 可以在這裡新增視覺回饋，比如高亮手牌
    if (this.position === 'bottom') {
      this.tiles.forEach(tile => {
        // 實際禁用/啟用牌面的互動性
        if (tile.sprite) {
          tile.sprite.eventMode = interactive ? 'static' : 'none';
        }

        if (interactive) {
          tile.container.alpha = 1.0; // 完全不透明
        } else {
          tile.container.alpha = 0.7; // 半透明表示不可互動
        }
      });
    }
  }

  /**
   * 設定玩家是否為當前行動者
   * @param {boolean} active - 是否輪到該玩家
   * @param {boolean} isSelf - 是否為本機玩家（自己）
   */
  setTurnActive(active, isSelf = false) {
    this.isTurnActive = active;

    // 根據位置決定資訊條尺寸
    const isHorizontal = (this.position === 'bottom' || this.position === 'top');
    const barWidth = isHorizontal ? 200 : 50;
    const barHeight = isHorizontal ? 36 : 180;

    // 清除之前的脈動動畫
    if (this.pulseAnimation) {
      clearInterval(this.pulseAnimation);
      this.pulseAnimation = null;
    }

    if (active) {
      // 顯示高亮邊框（發光效果）
      if (this.turnIndicatorBg) {
        this.turnIndicatorBg.visible = true;
        // 如果是自己，使用金色；其他玩家使用綠色
        const highlightColor = isSelf ? 0xFFD700 : 0x00FF00;
        this.turnIndicatorBg.clear();
        this.turnIndicatorBg.roundRect(-4, -4, barWidth + 8, barHeight + 8, 8);
        this.turnIndicatorBg.stroke({ width: 3, color: highlightColor });
      }

      // 顯示出牌中標籤
      if (this.turnLabel) {
        this.turnLabel.visible = true;
        this.turnLabel.style.fill = isSelf ? 0xFFD700 : 0x00FF00;
      }

      // 如果是自己，啟動脈動動畫
      if (isSelf) {
        let alpha = 1.0;
        let direction = -1;
        this.pulseAnimation = setInterval(() => {
          alpha += direction * 0.05;
          if (alpha <= 0.5) {
            alpha = 0.5;
            direction = 1;
          } else if (alpha >= 1.0) {
            alpha = 1.0;
            direction = -1;
          }
          if (this.turnIndicatorBg) {
            this.turnIndicatorBg.alpha = alpha;
          }
        }, 50);
      }

      console.log(`🎯 玩家 ${this.name || this.id} 進入行動狀態 (isSelf: ${isSelf})`);
    } else {
      // 隱藏高亮邊框
      if (this.turnIndicatorBg) {
        this.turnIndicatorBg.visible = false;
        this.turnIndicatorBg.alpha = 1.0; // 重置透明度
      }

      // 隱藏出牌中標籤
      if (this.turnLabel) {
        this.turnLabel.visible = false;
      }
    }
  }

  addDiscardedTile(tile) {
    this.discardedTiles.push(tile);
    // TODO: 在中央區域顯示打出的牌
  }

  removeTile(tileType) {
    console.log(`🗑️ [removeTile] 嘗試移除: ${tileType}, 當前手牌數: ${this.tiles.length}`);
    console.log(`🗑️ [removeTile] 當前手牌:`, this.tiles.map(t => t.type));

    // 從手牌陣列中找到並移除該牌
    const index = this.tiles.findIndex(tile => tile.type === tileType);
    if (index !== -1) {
      const tile = this.tiles[index];
      console.log(`✅ [removeTile] 找到牌，索引: ${index}`);

      // 從顯示中移除
      this.container.removeChild(tile.container);
      // 從陣列中移除
      this.tiles.splice(index, 1);

      // 重新排列剩餘的牌
      this.tiles.forEach((tile, i) => {
        this.positionTile(tile, i);
      });

      console.log(`✅ [removeTile] 移除成功，剩餘手牌數: ${this.tiles.length}`);
    } else {
      console.error(`❌ [removeTile] 找不到要移除的牌: ${tileType}`);
      console.error(`❌ [removeTile] 可能的原因: addTile 還未完成或牌已被移除`);
    }
  }

  /**
   * 加入一張新牌到手牌（摸牌）
   */
  async addTile(tileType, tileAssets) {
    // 記錄最後摸到的牌（用於聽牌後限制打牌）
    this.lastDrawnTile = tileType;

    // 先排序：將新牌加入並排序
    const allTileTypes = [...this.tiles.map(t => t.type), tileType];
    const sortedTypes = this.sortTiles(allTileTypes);

    // 清除舊的牌（視覺上）
    this.tiles.forEach(tile => {
      this.container.removeChild(tile.container);
      tile.destroy();
    });
    this.tiles = [];

    // 按照排序後的順序重新創建所有牌（同步創建，因為牌底紋理已預載入）
    for (const type of sortedTypes) {
      const texture = tileAssets[type] || tileAssets['back'];
      const tile = new Tile(type, texture);

      // 只有底部玩家（自己）的牌可以點擊
      if (this.position === 'bottom') {
        tile.on('click', (clickedTile) => this.onTileClick(clickedTile));

        // 設定初始互動狀態
        if (tile.sprite) {
          tile.sprite.eventMode = this.isInteractive ? 'static' : 'none';
        }
      }

      // 先加入到陣列
      this.tiles.push(tile);

      // 加入到容器
      this.container.addChild(tile.container);
    }

    // 所有牌都加入後，統一設定位置
    this.tiles.forEach((tile, index) => {
      this.positionTile(tile, index);
    });

    // 恢復聽牌狀態指示器（如果已宣告聽牌）
    if (this.isTing) {
      this.showTingStatus();
    }

    console.log(`✅ 加入新牌完成，手牌數: ${this.tiles.length}, 最後摸牌: ${this.lastDrawnTile}`);
  }

  /**
   * 重新排列所有手牌（只重新計算位置，不重新創建）
   */
  rearrangeTiles() {
    // 只重新計算每張牌的位置
    this.tiles.forEach((tile, index) => {
      this.positionTile(tile, index);
    });
  }

  /**
   * 高亮顯示可以組合的牌組
   * @param {Array<Array<string>>} tileGroups - 牌組列表，例如 [['wan-1', 'wan-2', 'wan-3'], ['wan-4', 'wan-4']]
   */
  highlightTileGroups(tileGroups) {
    // 先重置所有牌的位置
    this.clearHighlight();

    if (!tileGroups || tileGroups.length === 0) {
      return;
    }

    // 將牌組扁平化，找出所有需要高亮的牌
    const tilesToHighlight = new Set();
    tileGroups.forEach(group => {
      group.forEach(tileType => tilesToHighlight.add(tileType));
    });

    // 對於底部玩家，將可組合的牌往上移動
    if (this.position === 'bottom') {
      this.tiles.forEach((tile, index) => {
        if (tilesToHighlight.has(tile.type)) {
          // 往上移動 20 像素
          tile.container.y -= 20;
          // 新增發光效果
          tile.container.alpha = 1.0;
        } else {
          // 其他牌變暗
          tile.container.alpha = 0.5;
        }
      });
    }

    console.log(`✨ 高亮顯示 ${tilesToHighlight.size} 張牌`);
  }

  /**
   * 高亮顯示指定的牌
   * @param {Array<string>} tilesToHighlight - 要高亮的牌的類型列表
   */
  highlightTiles(tilesToHighlight) {
    // 先重置所有牌的位置
    this.clearHighlight();

    if (!tilesToHighlight || tilesToHighlight.length === 0) {
      return;
    }

    const highlightSet = new Set(tilesToHighlight);

    // 對於底部玩家，將可高亮的牌往上移動
    if (this.position === 'bottom') {
      this.tiles.forEach((tile) => {
        if (highlightSet.has(tile.type)) {
          // 往上移動 20 像素
          tile.container.y -= 20;
          // 新增發光效果
          tile.container.alpha = 1.0;
        } else {
          // 其他牌變暗
          tile.container.alpha = 0.5;
        }
      });
    }

    console.log(`✨ 高亮顯示 ${highlightSet.size} 張牌`);
  }

  /**
   * 清除高亮效果
   */
  clearHighlight() {
    if (this.position === 'bottom') {
      this.tiles.forEach((tile, index) => {
        // 重新定位到正確位置
        this.positionTile(tile, index);
        // 恢復透明度
        tile.container.alpha = 1.0;
      });
    }
  }

  /**
   * 新增吃/碰/槓牌組
   * @param {Object} meldData - {type: 'chow'|'pong'|'kong', tiles: [...]} 
   * @param {Object} tileAssets - 牌面素材
   */
  addFlower(flowerType) {
    this.flowers.push(flowerType);
  }

  /**
   * 新增吃/碰/槓牌組 (只更新數據)
   * @param {Object} meldData - {type: 'chow'|'pong'|'kong', tiles: [...]} 
   */
  addMeld(meldData) {
    const meldType = meldData.Type || meldData.type;
    const meldTiles = meldData.Tiles || meldData.tiles;

    if (meldType === 'kong_promoted') {
      const tile = meldTiles[0];
      const pongIndex = this.melds.findIndex(m => (m.Type || m.type) === 'pong' && (m.Tiles || m.tiles)[0] === tile);
      if (pongIndex !== -1) {
        this.melds[pongIndex] = meldData;
      } else {
        console.error(`Could not find a pong to promote for tile ${tile}. Adding kong as a new meld.`);
        this.melds.push(meldData);
      }
    } else {
      this.melds.push(meldData);
    }
  }

  /**
   * 統一更新所有明牌（吃碰槓+花牌）的顯示
   */
  async updateOpenLayout(tileAssets) {
    this.meldsContainer.removeChildren();

    if (!tileAssets || Object.keys(tileAssets).length === 0) {
      console.error('❌ updateOpenLayout: tileAssets is empty or undefined');
      return;
    }
    console.log(`📋 updateOpenLayout: position=${this.position}, melds=${this.melds.length}, flowers=${this.flowers.length}, tileAssets keys=${Object.keys(tileAssets).length}`);

        const tileWidth = 60;
        const tileHeight = 80;
        const groupSpacing = 25; // Increased from 15
        const gapFromHand = tileWidth * 0.8;
        let currentOffset = 0;
        let currentRow = 0; // 用於追蹤當前在第幾排

        const { Sprite, Container, Assets } = await import('pixi.js');
        const baseTexture = await Assets.load('/assets/tiles/carddown/basefdown.png');

        const handTileCount = this.tiles.length || 13;
        const handTileWidth = tileWidth * 0.75;
        const handSpacing = 5;
        const handWidth = handTileCount * handTileWidth + (handTileCount - 1) * handSpacing;

        // 計算可用的垂直空間（左右側玩家）
        // 明牌從 y=200 開始，可用空間 = 螢幕高度 - 起始位置
        // 當 currentOffset + 牌組高度 > maxVerticalOffset 時換排
        const maxVerticalOffset = this.screenHeight - 200;

        // 1. 渲染吃碰槓牌組
        for (const meld of this.melds) {
            const meldGroup = new Container();
            const meldType = meld.Type || meld.type;
            let meldTiles = meld.Tiles || meld.tiles;

            if (meldType === 'chow') {
                meldTiles = this.sortTiles([...meldTiles]);
            }

            for (let i = 0; i < meldTiles.length; i++) {
                const tileType = meldTiles[i];
                const tileContainer = new Container();
                const baseSprite = new Sprite(baseTexture);
                tileContainer.addChild(baseSprite);

                let texture = (meldType === 'kong_concealed') ? tileAssets['back'] : tileAssets[tileType];
                if (!texture) {
                    console.warn(`⚠️ updateOpenLayout: Missing texture for "${tileType}" (meldType: ${meldType}), available keys:`, Object.keys(tileAssets).slice(0, 10));
                    texture = tileAssets['back'];
                }
                if (!texture) {
                    console.error(`❌ updateOpenLayout: Even 'back' texture is missing! tileType: ${tileType}`);
                    continue; // 跳過這張牌
                }

                const tileSprite = new Sprite(texture);
                if (tileType && tileType.startsWith('tong-')) tileSprite.y = 5;
                tileContainer.addChild(tileSprite);
                // console.log(`🎴 Meld tile: ${tileType}, meldType: ${meldType}`);

                const isKong = meldType && meldType.includes('kong');
                if (isKong && i === 3) {
                    tileContainer.x = 1 * (tileWidth + 35); // Special positioning for 4th tile in a kong
                    tileContainer.y = -tileHeight * 0.1;
                } else {
                    tileContainer.x = i * (tileWidth + 35); // Horizontal spacing *within* a group
                }
                meldGroup.addChild(tileContainer);
            }

            const groupWidth = (meldTiles.length === 4 ? 3 : meldTiles.length) * (tileWidth + 35);
            const scale = (this.position === 'bottom' ? 0.75 : 0.6) * 0.8;
            const scaledGroupHeight = groupWidth * scale; // 旋轉後寬度變高度

            // 檢查是否需要換排（左右側玩家）
            if ((this.position === 'left' || this.position === 'right') &&
                currentOffset + scaledGroupHeight > maxVerticalOffset && currentOffset > 0) {
                currentRow++;
                currentOffset = 0;
            }

            this.positionOpenGroup(meldGroup, scale, groupWidth, currentOffset, handWidth, gapFromHand, currentRow);
            currentOffset += groupWidth * scale + groupSpacing;
            this.meldsContainer.addChild(meldGroup);
        }

        // 2. 渲染花牌
        if (this.flowers.length > 0) {
            const flowerGroup = new Container();
            for (let i = 0; i < this.flowers.length; i++) {
                const tileType = this.flowers[i];
                const tileContainer = new Container();
                const baseSprite = new Sprite(baseTexture);
                tileContainer.addChild(baseSprite);

                let texture = tileAssets[tileType] || tileAssets['back'];
                if (!texture) {
                    console.warn(`⚠️ updateOpenLayout: Missing flower texture for "${tileType}"`);
                    continue;
                }
                const tileSprite = new Sprite(texture);
                if (tileType && tileType.startsWith('tong-')) tileSprite.y = 5;
                tileContainer.addChild(tileSprite);
                tileContainer.x = i * (tileWidth + 35);
                flowerGroup.addChild(tileContainer);
            }

            const groupWidth = this.flowers.length * (tileWidth + 35);
            const scale = (this.position === 'bottom' ? 0.75 : 0.6) * 0.8;
            const scaledGroupHeight = groupWidth * scale;

            // 檢查是否需要換排（左右側玩家）
            if ((this.position === 'left' || this.position === 'right') &&
                currentOffset + scaledGroupHeight > maxVerticalOffset && currentOffset > 0) {
                currentRow++;
                currentOffset = 0;
            }

            this.positionOpenGroup(flowerGroup, scale, groupWidth, currentOffset, handWidth, gapFromHand, currentRow);
            this.meldsContainer.addChild(flowerGroup);
        }
  }

  positionOpenGroup(group, scale, groupWidth, offset, handWidth, gapFromHand, row = 0) {
    group.scale.set(scale);

    // 牌組旋轉 90 度後，原本的高度變成寬度
    // 一組牌（3張）總寬度 = 3 * (60+35) = 285，縮放後約 171
    // 但旋轉後這個寬度變成高度，而原本的高度 80 變成寬度 = 80 * 0.6 = 48
    // 排與排之間需要足夠的間距避免重疊，使用 90px
    const rowSpacing = 90;

    switch (this.position) {
      case 'bottom': {
        // For the bottom player, groups are positioned horizontally within the meldsContainer.
        // The container itself is positioned by the positionTile function.
        group.x = offset;
        group.y = 0;
        break;
      }
      case 'right': {
        // 右側玩家：第二排往右側開啟（往螢幕外側）
        // 旋轉 90 度後，牌組從 group.x 向左延伸
        // 第二排要往右移，所以 group.x 要增加
        group.x = this.screenWidth - 120 + row * rowSpacing;
        group.y = 100 + offset; // 往上移 50px（從 150 改為 100）
        group.rotation = Math.PI / 2;
        break;
      }
      case 'top': {
        group.x = 500 + offset;
        group.y = 150;
        group.rotation = Math.PI;
        break;
      }
      case 'left': {
        // 左側玩家：第二排往左側開啟（往螢幕外側）
        // 旋轉 90 度後，牌組從 group.x 向左延伸
        // 第二排要往左移，所以 group.x 要減少
        group.x = 200 - row * rowSpacing;
        group.y = 100 + offset; // 往上移 50px（從 150 改為 100）
        group.rotation = Math.PI / 2;
        break;
      }
    }
  }

  /**
   * 顯示聽牌狀態
   */
  showTingStatus() {
    // 如果已經有聽牌狀態文字，先移除
    if (this.tingStatusText) {
      this.container.removeChild(this.tingStatusText);
    }

    // 創建聽牌狀態文字
    this.tingStatusText = new Text({
      text: '聽',
      style: {
        fontSize: 32,
        fill: 0xFF0000, // 紅色
        fontWeight: 'bold',
        stroke: 0xFFFFFF,
        strokeThickness: 3
      }
    });
    this.tingStatusText.anchor.set(0.5);

    // 根據位置設定聽牌圖示位置
    switch (this.position) {
      case 'bottom':
        this.tingStatusText.x = this.screenWidth / 2;
        this.tingStatusText.y = this.screenHeight - 120;
        break;
      case 'right':
        this.tingStatusText.x = this.screenWidth - 60;
        this.tingStatusText.y = this.screenHeight / 2;
        break;
      case 'top':
        this.tingStatusText.x = this.screenWidth / 2;
        this.tingStatusText.y = 80;
        break;
      case 'left':
        this.tingStatusText.x = 60;
        this.tingStatusText.y = this.screenHeight / 2;
        break;
    }

    this.container.addChild(this.tingStatusText);
    console.log(`✅ 顯示聽牌狀態: ${this.name || this.id}`);
  }

  /**
   * 隱藏聽牌狀態
   */
  hideTingStatus() {
    if (this.tingStatusText) {
      this.container.removeChild(this.tingStatusText);
      this.tingStatusText = null;
    }
  }

  resize(width, height) {
    this.screenWidth = width;
    this.screenHeight = height;

    // 重新定位元素
    this.container.removeChildren();
    this.createInfoDisplay();

    // 重新定位手牌
    this.tiles.forEach((tile, index) => {
      this.positionTile(tile, index);
    });

    // 重新顯示聽牌狀態（如果有）
    if (this.isTing) {
      this.showTingStatus();
    }
  }

  reset() {
    // 清除手牌
    this.tiles.forEach(tile => tile.destroy());
    this.tiles = [];

    // 清除吃碰槓
    this.meldsContainer.removeChildren();
    this.melds = []; // 吃/碰/槓的牌組 [{type: 'chow'|'pong'|'kong', tiles: [...]}]
    this.flowers = []; // 花牌

    // 重置狀態
    this.isTing = false;
    this.winningTiles = [];
    this.lastDrawnTile = null;
    this.hideTingStatus();

    // 清除輪次指示器
    this.setTurnActive(false);
  }
}