import { Container, Graphics, Text, Sprite, Texture, Assets, Rectangle } from 'pixi.js';

/**
 * 擲骰動畫 UI 元件
 * 使用 2D 精靈圖模擬 3D 骰子滾動效果
 */
export class DiceRollUI {
  constructor(screenWidth, screenHeight, audioManager) {
    this.screenWidth = screenWidth;
    this.screenHeight = screenHeight;
    this.audioManager = audioManager;

    this.container = new Container();
    this.container.zIndex = 15000; // 高於所有其他 UI
    this.container.visible = false;

    // 精靈圖配置（10x10 格子，每格約 74x74 像素）
    this.frameWidth = 74;
    this.frameHeight = 74;
    this.framesPerRow = 10;
    this.totalRows = 10;
    this.totalFrames = 100;

    // 動畫狀態
    this.diceSprites = [];
    this.diceShadows = [];
    this.diceFrameTextures = [];
    this.isAnimating = false;
    this.targetValues = [1, 1, 1]; // 來自伺服器的骰子結果

    // 物理模擬參數
    this.dicePhysics = [];

    // 最終值對應的幀索引（需要根據實際精靈圖調整）
    // 這些索引對應到精靈圖中顯示特定點數朝上的幀
    this.finalFrameMap = {
      1: 40,  // 面朝上顯示 1 點
      2: 20,  // 面朝上顯示 2 點
      3: 30,  // 面朝上顯示 3 點
      4: 0,   // 面朝上顯示 4 點
      5: 50,  // 面朝上顯示 5 點
      6: 60   // 面朝上顯示 6 點
    };

    // 子元件
    this.overlay = null;
    this.titleText = null;
    this.diceContainer = null;
    this.resultText = null;
    this.dealerText = null;
    this.shadowTexture = null;
  }

  /**
   * 載入骰子精靈圖並創建幀材質
   */
  async load() {
    try {
      // 載入精靈圖
      const spriteSheet = await Assets.load('/assets/ui/dice.png');

      // 從精靈圖創建幀材質
      this.diceFrameTextures = [];
      for (let row = 0; row < this.totalRows; row++) {
        for (let col = 0; col < this.framesPerRow; col++) {
          const frame = new Rectangle(
            col * this.frameWidth,
            row * this.frameHeight,
            this.frameWidth,
            this.frameHeight
          );
          const frameTexture = new Texture({
            source: spriteSheet.source,
            frame: frame
          });
          this.diceFrameTextures.push(frameTexture);
        }
      }

      // 載入陰影材質
      try {
        this.shadowTexture = await Assets.load('/assets/ui/dice_shadow.png');
      } catch (e) {
        console.warn('無法載入骰子陰影材質，將不顯示陰影');
        this.shadowTexture = null;
      }

      console.log('骰子素材載入完成，共 ' + this.diceFrameTextures.length + ' 幀');
    } catch (error) {
      console.error('載入骰子素材失敗:', error);
    }
  }

  /**
   * 創建遮罩背景和 UI 元素
   */
  createOverlay() {
    // 清除現有內容
    this.container.removeChildren();

    // 半透明暗色背景 - 深綠色麻將桌風格
    this.overlay = new Graphics();
    this.overlay.rect(0, 0, this.screenWidth, this.screenHeight);
    this.overlay.fill({ color: 0x0a2a1a, alpha: 0.92 });
    this.container.addChild(this.overlay);

    // 中央面板背景
    const panelWidth = 450;
    const panelHeight = 350;
    const panelX = (this.screenWidth - panelWidth) / 2;
    const panelY = (this.screenHeight - panelHeight) / 2 - 30;

    const panel = new Graphics();
    // 外層陰影
    panel.roundRect(panelX - 6, panelY - 6, panelWidth + 12, panelHeight + 12, 20);
    panel.fill({ color: 0x000000, alpha: 0.4 });
    // 主面板 - 深綠色
    panel.roundRect(panelX, panelY, panelWidth, panelHeight, 16);
    panel.fill({ color: 0x1a4a3a, alpha: 0.95 });
    // 內層高光
    panel.roundRect(panelX + 4, panelY + 4, panelWidth - 8, panelHeight - 8, 12);
    panel.fill({ color: 0x2a6a5a, alpha: 0.2 });
    // 金色邊框
    panel.roundRect(panelX, panelY, panelWidth, panelHeight, 16);
    panel.stroke({ width: 3, color: 0xFFD700 });
    this.container.addChild(panel);

    // 標題文字
    this.titleText = new Text({
      text: '擲骰決定莊家',
      style: {
        fontSize: 42,
        fill: 0xFFD700,
        fontWeight: 'bold',
        fontFamily: 'Arial, sans-serif',
        stroke: { color: 0x000000, width: 4 }
      }
    });
    this.titleText.anchor.set(0.5);
    this.titleText.x = this.screenWidth / 2;
    this.titleText.y = 80;
    this.container.addChild(this.titleText);

    // 骰子容器（置中）
    this.diceContainer = new Container();
    this.diceContainer.x = this.screenWidth / 2;
    this.diceContainer.y = this.screenHeight / 2 - 30;
    this.container.addChild(this.diceContainer);

    // 結果文字（初始隱藏）
    this.resultText = new Text({
      text: '',
      style: {
        fontSize: 72,
        fill: 0xFFFFFF,
        fontWeight: 'bold',
        fontFamily: 'Arial, sans-serif',
        stroke: { color: 0x000000, width: 6 }
      }
    });
    this.resultText.anchor.set(0.5);
    this.resultText.x = this.screenWidth / 2;
    this.resultText.y = this.screenHeight / 2 + 130;
    this.resultText.visible = false;
    this.container.addChild(this.resultText);

    // 莊家公告文字
    this.dealerText = new Text({
      text: '',
      style: {
        fontSize: 36,
        fill: 0x00FF88,
        fontWeight: 'bold',
        fontFamily: 'Arial, sans-serif',
        stroke: { color: 0x000000, width: 3 }
      }
    });
    this.dealerText.anchor.set(0.5);
    this.dealerText.x = this.screenWidth / 2;
    this.dealerText.y = this.screenHeight / 2 + 200;
    this.dealerText.visible = false;
    this.container.addChild(this.dealerText);
  }

  /**
   * 初始化骰子精靈和物理狀態
   */
  createDice() {
    this.diceSprites = [];
    this.diceShadows = [];
    this.dicePhysics = [];

    const spacing = 130;
    const startX = -spacing;

    for (let i = 0; i < 3; i++) {
      // 陰影精靈
      if (this.shadowTexture) {
        const shadow = new Sprite(this.shadowTexture);
        shadow.anchor.set(0.5);
        shadow.x = startX + i * spacing;
        shadow.y = 50;
        shadow.alpha = 0.4;
        shadow.scale.set(0.9);
        this.diceContainer.addChild(shadow);
        this.diceShadows.push(shadow);
      }

      // 骰子精靈
      const dice = new Sprite(this.diceFrameTextures[0]);
      dice.anchor.set(0.5);
      dice.x = startX + i * spacing;
      dice.y = -100; // 從上方開始
      dice.scale.set(1.3);
      this.diceContainer.addChild(dice);
      this.diceSprites.push(dice);

      // 此骰子的物理狀態
      this.dicePhysics.push({
        x: startX + i * spacing,
        y: -100 - Math.random() * 50, // 初始高度隨機偏移
        velocityX: (Math.random() - 0.5) * 8,
        velocityY: 0,
        angularVelocity: (Math.random() - 0.5) * 0.8,
        rotation: Math.random() * Math.PI * 2,
        frameIndex: Math.floor(Math.random() * this.totalFrames),
        bounces: 0,
        settled: false,
        groundY: 0 // 地面 Y 座標
      });
    }
  }

  /**
   * 播放擲骰動畫
   * @param {Array<number>} targetValues - [骰子1, 骰子2, 骰子3] 來自伺服器的結果
   * @param {number} totalSum - 骰子總和
   * @param {string} dealerName - 莊家名稱
   * @param {number} dealerPosition - 莊家位置 (0-3)
   * @returns {Promise} 動畫完成時解析
   */
  async play(targetValues, totalSum, dealerName, dealerPosition) {
    this.targetValues = targetValues;
    this.totalSum = totalSum;
    this.dealerName = dealerName;
    this.dealerPosition = dealerPosition;

    // 初始化
    this.createOverlay();
    this.createDice();
    this.container.visible = true;
    this.isAnimating = true;

    // 播放音效
    if (this.audioManager) {
      this.audioManager.playEffect('dice');
    }

    // 執行動畫循環
    return new Promise((resolve) => {
      const startTime = Date.now();
      const maxDuration = 3000; // 最長動畫時間 3 秒

      const animate = () => {
        if (!this.isAnimating) {
          resolve();
          return;
        }

        const elapsed = Date.now() - startTime;
        let allSettled = true;

        for (let i = 0; i < 3; i++) {
          const physics = this.dicePhysics[i];
          const dice = this.diceSprites[i];
          const shadow = this.diceShadows[i];

          if (!physics.settled) {
            allSettled = false;

            // 重力
            physics.velocityY += 0.6;

            // 更新位置
            physics.y += physics.velocityY;
            physics.x += physics.velocityX;

            // 更新旋轉
            physics.rotation += physics.angularVelocity;

            // 根據旋轉計算幀索引（模擬 3D 旋轉）
            const rotationFactor = Math.abs(physics.rotation) % (Math.PI * 2);
            const rowOffset = Math.floor((rotationFactor / (Math.PI * 2)) * this.totalRows);
            const colOffset = Math.floor((physics.x / 50 + 10) % this.framesPerRow);
            physics.frameIndex = (rowOffset * this.framesPerRow + colOffset) % this.totalFrames;

            // 反彈（碰到地面）
            if (physics.y > physics.groundY && physics.velocityY > 0) {
              physics.y = physics.groundY;
              physics.velocityY = -physics.velocityY * 0.4;
              physics.velocityX *= 0.8;
              physics.angularVelocity *= 0.6;
              physics.bounces++;

              // 經過足夠次數的反彈後停止
              if (physics.bounces >= 3 || Math.abs(physics.velocityY) < 1.5 || elapsed > maxDuration - 500) {
                physics.settled = true;
                physics.y = physics.groundY;
                // 對齊到伺服器指定的結果
                physics.frameIndex = this.finalFrameMap[this.targetValues[i]] || 0;
              }
            }

            // 水平邊界碰撞
            const maxX = 150;
            if (Math.abs(physics.x) > maxX) {
              physics.x = Math.sign(physics.x) * maxX;
              physics.velocityX = -physics.velocityX * 0.6;
            }

            // 更新精靈
            dice.x = physics.x;
            dice.y = physics.y - 40; // 視覺高度偏移

            if (this.diceFrameTextures[physics.frameIndex]) {
              dice.texture = this.diceFrameTextures[physics.frameIndex];
            }

            // 更新陰影
            if (shadow) {
              shadow.x = physics.x;
              const heightRatio = Math.max(0, 1 - Math.abs(physics.y) / 100);
              shadow.scale.set(0.6 + heightRatio * 0.4);
              shadow.alpha = 0.2 + heightRatio * 0.3;
            }
          } else {
            // 已停止，確保顯示正確的幀
            dice.texture = this.diceFrameTextures[this.finalFrameMap[this.targetValues[i]] || 0];
          }
        }

        // 強制超時後停止
        if (elapsed > maxDuration) {
          for (let i = 0; i < 3; i++) {
            this.dicePhysics[i].settled = true;
            this.diceSprites[i].texture = this.diceFrameTextures[this.finalFrameMap[this.targetValues[i]] || 0];
            this.diceSprites[i].y = -40;
          }
          allSettled = true;
        }

        if (allSettled) {
          // 所有骰子已停止，顯示結果
          this.showResult();
          setTimeout(() => {
            this.isAnimating = false;
            resolve();
          }, 1800);
        } else {
          requestAnimationFrame(animate);
        }
      };

      requestAnimationFrame(animate);
    });
  }

  /**
   * 骰子停止後顯示結果
   */
  showResult() {
    // 顯示總點數
    this.resultText.text = `${this.totalSum} 點`;
    this.resultText.visible = true;

    // 顯示莊家
    const positionNames = ['東家', '南家', '西家', '北家'];
    this.dealerText.text = `莊家: ${this.dealerName} (${positionNames[this.dealerPosition]})`;
    this.dealerText.visible = true;

    // 添加閃爍動畫
    let alpha = 1;
    let direction = -1;
    const pulseInterval = setInterval(() => {
      if (!this.container.visible) {
        clearInterval(pulseInterval);
        return;
      }
      alpha += direction * 0.05;
      if (alpha <= 0.6 || alpha >= 1) {
        direction = -direction;
      }
      if (this.dealerText) {
        this.dealerText.alpha = alpha;
      }
    }, 50);

    // 2 秒後停止閃爍
    setTimeout(() => {
      clearInterval(pulseInterval);
      if (this.dealerText) {
        this.dealerText.alpha = 1;
      }
    }, 1800);
  }

  /**
   * 隱藏並清理
   */
  hide() {
    this.container.visible = false;
    this.isAnimating = false;

    // 清理
    if (this.diceContainer) {
      this.diceContainer.removeChildren();
    }
    this.diceSprites = [];
    this.diceShadows = [];
    this.dicePhysics = [];
  }

  /**
   * 畫面尺寸變更處理
   */
  resize(width, height) {
    this.screenWidth = width;
    this.screenHeight = height;

    // 如果正在顯示，重新創建 UI
    if (this.container.visible && !this.isAnimating) {
      this.createOverlay();
    }
  }

  /**
   * 銷毀資源
   */
  destroy() {
    this.hide();
    this.container.removeChildren();
    this.diceFrameTextures = [];
  }
}
