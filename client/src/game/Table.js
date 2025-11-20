import { Container, Graphics, Text } from 'pixi.js';

/**
 * 麻将桌类
 */
export class Table {
  constructor(width, height) {
    this.width = width;
    this.height = height;
    this.container = new Container();

    this.create();
  }

  create() {
    // 绘制桌面
    const table = new Graphics();

    // 外边框 - 深绿色
    table.rect(0, 0, this.width, this.height);
    table.fill(0x1a5f3c);

    // 中央牌桌区域 - 稍浅的绿色
    const centerWidth = Math.min(this.width * 0.7, 900);
    const centerHeight = Math.min(this.height * 0.7, 600);
    const centerX = (this.width - centerWidth) / 2;
    const centerY = (this.height - centerHeight) / 2;

    table.rect(centerX, centerY, centerWidth, centerHeight);
    table.fill(0x2d7a4f);
    table.stroke({ width: 4, color: 0x8B7355 });

    // 中央装饰 - 麻将图案
    const centerCircle = new Graphics();
    centerCircle.circle(this.width / 2, this.height / 2, 80);
    centerCircle.stroke({ width: 3, color: 0x1a5f3c });

    const centerText = new Text({
      text: '🀄',
      style: {
        fontSize: 64,
        fill: 0x1a5f3c
      }
    });
    centerText.x = this.width / 2 - 32;
    centerText.y = this.height / 2 - 40;

    this.container.addChild(table);
    this.container.addChild(centerCircle);
    this.container.addChild(centerText);

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

  resize(width, height) {
    this.width = width;
    this.height = height;

    // 重新创建桌面
    this.container.removeChildren();
    this.create();
  }
}
