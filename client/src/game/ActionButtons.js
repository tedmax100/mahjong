import { Container, Sprite, Assets } from 'pixi.js';

/**
 * 遊戲動作按鈕（碰、吃、槓、胡）
 */
export class ActionButtons {
  constructor(screenWidth, screenHeight) {
    this.screenWidth = screenWidth;
    this.screenHeight = screenHeight;
    this.container = new Container();
    this.buttons = {}; // 存儲各個按鈕
    this.visible = false;
    this.callbacks = {}; // 存儲回調函數

    // 按鈕配置
    this.buttonConfigs = [
      { action: 'pong', image: '/assets/ui/Pong.png', x: -180 },
      { action: 'chow', image: '/assets/ui/Chow.png', x: -60 },
      { action: 'kong', image: '/assets/ui/Kong.png', x: 60 },
      { action: 'hu', image: '/assets/ui/Hu.png', x: 180 },
      { action: 'cancel', image: '/assets/ui/playcancel.png', x: 300 }
    ];

    this.init();
  }

  async init() {
    console.log('開始載入動作按鈕...');

    // 載入所有按鈕圖片
    for (const config of this.buttonConfigs) {
      try {
        const texture = await Assets.load(config.image);
        const button = new Sprite(texture);

        button.anchor.set(0.5);
        button.scale.set(0.8); // 縮小按鈕
        button.x = this.screenWidth / 2 + config.x;
        button.y = this.screenHeight - 250; // 按鈕位置（在手牌上方）

        // 設置為可交互
        button.eventMode = 'static';
        button.cursor = 'pointer';

        // 添加點擊事件
        button.on('pointerdown', () => this.onButtonClick(config.action));

        // 添加懸停效果
        button.on('pointerover', () => {
          button.scale.set(0.9);
        });
        button.on('pointerout', () => {
          button.scale.set(0.8);
        });

        this.buttons[config.action] = button;
        this.container.addChild(button);

        console.log(`✅ 按鈕載入成功: ${config.action}`);
      } catch (error) {
        console.warn(`⚠️ 無法載入按鈕圖片: ${config.image}`, error);
        // 即使圖片載入失敗，也不要中斷初始化
      }
    }

    // 初始隱藏所有按鈕
    this.hide();

    console.log('✅ 動作按鈕初始化完成');
  }

  /**
   * 顯示指定的按鈕
   * @param {Array<string>} actions - 要顯示的動作列表，例如 ['pong', 'chow', 'cancel']
   */
  show(actions) {
    // 先隱藏所有按鈕
    this.hide();

    // 顯示指定的按鈕
    actions.forEach(action => {
      if (this.buttons[action]) {
        this.buttons[action].visible = true;
      }
    });

    this.visible = true;
    this.container.visible = true;
  }

  /**
   * 隱藏所有按鈕
   */
  hide() {
    Object.values(this.buttons).forEach(button => {
      button.visible = false;
    });
    this.visible = false;
    this.container.visible = false;
  }

  /**
   * 按鈕點擊處理
   */
  onButtonClick(action) {
    console.log(`按鈕點擊: ${action}`);

    // 隱藏按鈕
    this.hide();

    // 觸發回調
    if (this.callbacks[action]) {
      this.callbacks[action]();
    }
  }

  /**
   * 設置按鈕回調函數
   * @param {string} action - 動作名稱
   * @param {Function} callback - 回調函數
   */
  on(action, callback) {
    this.callbacks[action] = callback;
  }

  /**
   * 調整按鈕位置（當螢幕大小改變時）
   */
  resize(width, height) {
    this.screenWidth = width;
    this.screenHeight = height;

    // 重新定位所有按鈕
    this.buttonConfigs.forEach(config => {
      if (this.buttons[config.action]) {
        this.buttons[config.action].x = width / 2 + config.x;
        this.buttons[config.action].y = height - 250;
      }
    });
  }
}
