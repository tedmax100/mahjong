import { Container, Graphics, Text, Sprite, Assets, Texture, Rectangle } from 'pixi.js';

/**
 * 麻将桌类
 */
export class Table {
  constructor(width, height) {
    this.width = width;
    this.height = height;
    this.container = new Container();
    this.centerSprite = null;
    this.centerTextures = [];
    this.currentTextureIndex = 0;

    this.create();
  }

  create() {
    // 绘制桌面
    const table = new Graphics();

    // 外边框 - 深绿色
    table.rect(0, 0, this.width, this.height);
    table.fill(0x1a5f3c);

    // 中央牌桌区域 - 稍浅的绿色
    const centerWidth = Math.min(this.width * 0.92, 1600);
    const centerHeight = Math.min(this.height * 0.92, 950);
    const centerX = (this.width - centerWidth) / 2;
    const centerY = (this.height - centerHeight) / 2;

    table.rect(centerX, centerY, centerWidth, centerHeight);
    table.fill(0x2d7a4f);
    table.stroke({ width: 4, color: 0x8B7355 });

    this.container.addChild(table);

    // 异步加载中央装饰图片
    this.loadCenterImage();

    // 绘制四个玩家区域标记
    this.drawPlayerMarkers(centerX, centerY, centerWidth, centerHeight);
  }

  drawPlayerMarkers(x, y, width, height) {
    const positions = [
      { x: this.width / 2, y: y + height + 20, text: '你' },        // 下
      { x: x + width + 20, y: this.height / 2, text: '右家' },      // 右
      { x: this.width / 2, y: y - 20, text: '对家' },               // 上
      { x: x - 20, y: this.height / 2, text: '左家' }               // 左
    ];

    positions.forEach((pos, index) => {
      const marker = new Graphics();
      marker.circle(0, 0, 30);
      marker.fill(0x4a7c59);
      marker.stroke({ width: 2, color: 0x8B7355 });

      const text = new Text({
        text: pos.text,
        style: {
          fontSize: 16,
          fill: 0xffffff,
          fontWeight: 'bold'
        }
      });
      text.anchor.set(0.5);

      marker.addChild(text);
      marker.x = pos.x;
      marker.y = pos.y;

      this.container.addChild(marker);
    });
  }

  async loadCenterImage() {
    try {
      // 加载中央图片
      const texture = await Assets.load('/assets/ui/bg_center.jpg');

      // 图片是 3列 x 4行 = 12个小图
      const cols = 3;
      const rows = 4;
      const cellWidth = texture.width / cols;
      const cellHeight = texture.height / rows;

      // 创建12个纹理
      for (let row = 0; row < rows; row++) {
        for (let col = 0; col < cols; col++) {
          const rect = new Rectangle(
            col * cellWidth,
            row * cellHeight,
            cellWidth,
            cellHeight
          );
          const cellTexture = new Texture({
            source: texture.source,
            frame: rect
          });
          this.centerTextures.push(cellTexture);
        }
      }

      // 创建中央 sprite
      this.centerSprite = new Sprite(this.centerTextures[0]);
      this.centerSprite.anchor.set(0.5);
      this.centerSprite.x = this.width / 2;
      this.centerSprite.y = this.height / 2;

      // 设置大小（调整为合适的显示尺寸）
      const displaySize = 200;
      const scale = displaySize / cellWidth;
      this.centerSprite.scale.set(scale);

      this.container.addChild(this.centerSprite);

      // 启动轮播（每2秒切换一张图）
      this.startCarousel();
    } catch (error) {
      console.warn('无法加载中央装饰图片:', error);
      // 如果加载失败，显示原来的文字
      this.showFallbackCenter();
    }
  }

  startCarousel() {
    setInterval(() => {
      this.currentTextureIndex = (this.currentTextureIndex + 1) % this.centerTextures.length;
      if (this.centerSprite) {
        this.centerSprite.texture = this.centerTextures[this.currentTextureIndex];
      }
    }, 2000); // 每2秒切换一张
  }

  showFallbackCenter() {
    const centerText = new Text({
      text: '🀄',
      style: {
        fontSize: 64,
        fill: 0x1a5f3c
      }
    });
    centerText.x = this.width / 2 - 32;
    centerText.y = this.height / 2 - 40;
    this.container.addChild(centerText);
  }

  resize(width, height) {
    this.width = width;
    this.height = height;

    // 重新创建桌面
    this.container.removeChildren();
    this.create();
  }
}
