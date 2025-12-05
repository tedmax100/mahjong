import { Sprite, Container, Assets } from 'pixi.js';

/**
 * 麻將牌類
 */
export class Tile {
  constructor(type, texture, faceUp = true) {
    this.type = type; // 例如: 'wan-1', 'tong-5', 'dong'
    this.texture = texture;
    this.faceUp = faceUp;
    this.container = new Container();
    this.baseSprite = null; // 牌底
    this.sprite = null; // 牌面
    this.isHovered = false; // 追蹤懸停狀態
    this.hoverOffset = 15; // 懸停時上移的像素

    this.create();
  }

  async create() {
    // 對於牌背類型（對手的手牌），不需要顯示牌底，直接顯示牌背圖片
    const isBackTile = this.type === 'back';

    // 載入牌底圖片（僅對正面牌需要）
    if (!isBackTile) {
      let baseTexture;
      try {
        baseTexture = await Assets.load('/assets/tiles/carddown/basefdown.png');
      } catch (error) {
        console.warn('無法載入牌底圖片，使用預設', error);
      }

      // 創建牌底 sprite（如果有載入成功）
      if (baseTexture) {
        this.baseSprite = new Sprite(baseTexture);
        this.baseSprite.anchor.set(0, 0); // 左上角對齊
        this.container.addChild(this.baseSprite);
      }
    }

    // 創建牌面 sprite（對於牌背類型，這就是牌背圖片）
    this.sprite = new Sprite(this.texture);
    this.sprite.anchor.set(0, 0); // 左上角對齊
    this.sprite.interactive = true;
    this.sprite.buttonMode = true;

    // 設定點擊事件
    this.sprite.on('pointerdown', () => this.onClick());
    this.sprite.on('pointerover', () => this.onHover());
    this.sprite.on('pointerout', () => this.onHoverOut());

    // 針對筒子微調（牌背不需要）
    if (!isBackTile && this.type.startsWith('tong-')) {
      this.sprite.y = 8; // 往下移8個像素
    }

    this.container.addChild(this.sprite);
  }

  onClick() {
    // 牌被點擊
    this.emit('click', this);
  }

  onHover() {
    // 滑鼠懸停 - 整張牌上移
    if (this.faceUp && !this.isHovered) {
      this.container.y -= this.hoverOffset;
      this.isHovered = true;
    }
  }

  onHoverOut() {
    // 滑鼠離開 - 恢復位置
    if (this.faceUp && this.isHovered) {
      this.container.y += this.hoverOffset;
      this.isHovered = false;
    }
  }

  emit(event, data) {
    // 簡單的事件系統
    if (this.container.eventListeners && this.container.eventListeners[event]) {
      this.container.eventListeners[event].forEach(callback => callback(data));
    }
  }

  on(event, callback) {
    if (!this.container.eventListeners) {
      this.container.eventListeners = {};
    }
    if (!this.container.eventListeners[event]) {
      this.container.eventListeners[event] = [];
    }
    this.container.eventListeners[event].push(callback);
  }

  setPosition(x, y) {
    // 如果正在懸停，需要考慮偏移量
    if (this.isHovered) {
      this.container.x = x;
      this.container.y = y - this.hoverOffset;
    } else {
      this.container.x = x;
      this.container.y = y;
    }
  }

  setRotation(angle) {
    this.container.rotation = angle;
  }

  setScale(scale) {
    this.container.scale.set(scale);
  }

  flip(faceUp) {
    this.faceUp = faceUp;
    // 可以在這裡切換紋理到牌背
  }

  destroy() {
    if (this.sprite) {
      this.sprite.destroy();
    }
    this.container.destroy();
  }
}