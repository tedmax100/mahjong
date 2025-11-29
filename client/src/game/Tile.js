import { Sprite, Container, Assets } from 'pixi.js';

/**
 * 麻将牌类
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
    // 載入牌底圖片
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

    // 創建牌面 sprite
    this.sprite = new Sprite(this.texture);
    this.sprite.anchor.set(0, 0); // 左上角對齊
    this.sprite.interactive = true;
    this.sprite.buttonMode = true;

    // 设置点击事件
    this.sprite.on('pointerdown', () => this.onClick());
    this.sprite.on('pointerover', () => this.onHover());
    this.sprite.on('pointerout', () => this.onHoverOut());

    // 針對筒子微調
    if (this.type.startsWith('tong-')) {
      this.sprite.y = 8; // 往下移8個像素
    }

    this.container.addChild(this.sprite);
  }

  onClick() {
    // 牌被点击
    this.emit('click', this);
  }

  onHover() {
    // 鼠标悬停 - 整張牌上移
    if (this.faceUp && !this.isHovered) {
      this.container.y -= this.hoverOffset;
      this.isHovered = true;
    }
  }

  onHoverOut() {
    // 鼠标离开 - 恢复位置
    if (this.faceUp && this.isHovered) {
      this.container.y += this.hoverOffset;
      this.isHovered = false;
    }
  }

  emit(event, data) {
    // 简单的事件系统
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
    // 可以在这里切换纹理到牌背
  }

  destroy() {
    if (this.sprite) {
      this.sprite.destroy();
    }
    this.container.destroy();
  }
}
