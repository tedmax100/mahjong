import { Sprite, Container } from 'pixi.js';

/**
 * 麻将牌类
 */
export class Tile {
  constructor(type, texture, faceUp = true) {
    this.type = type; // 例如: 'wan-1', 'tong-5', 'dong'
    this.texture = texture;
    this.faceUp = faceUp;
    this.container = new Container();
    this.sprite = null;

    this.create();
  }

  create() {
    this.sprite = new Sprite(this.texture);
    this.sprite.interactive = true;
    this.sprite.buttonMode = true;

    // 设置点击事件
    this.sprite.on('pointerdown', () => this.onClick());
    this.sprite.on('pointerover', () => this.onHover());
    this.sprite.on('pointerout', () => this.onHoverOut());

    this.container.addChild(this.sprite);
  }

  onClick() {
    // 牌被点击
    this.emit('click', this);
  }

  onHover() {
    // 鼠标悬停 - 牌上移
    if (this.faceUp) {
      this.sprite.y = -10;
    }
  }

  onHoverOut() {
    // 鼠标离开 - 恢复位置
    this.sprite.y = 0;
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
    this.container.x = x;
    this.container.y = y;
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
