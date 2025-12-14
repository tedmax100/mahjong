import { Container, Graphics, Text, Sprite, Assets, Texture, Rectangle } from 'pixi.js';

/**
 * 麻將桌類
 */
export class Table {
  constructor(width, height) {
    this.width = width;
    this.height = height;
    this.container = new Container();
    this.centerSprite = null;
    this.centerTextures = [];
    this.textureMetadata = []; // 儲存每個紋理的原始尺寸資訊
    this.currentTextureIndex = 0;
    this.targetDisplaySize = 0; // 目標顯示尺寸

    this.create();
  }

  create() {
    // 繪製桌面
    const table = new Graphics();

    // 外邊框 - 深綠色
    table.rect(0, 0, this.width, this.height);
    table.fill(0x1a5f3c);

    // 中央牌桌區域 - 稍淺的綠色
    const centerWidth = Math.min(this.width * 0.92, 1600);
    const centerHeight = Math.min(this.height * 0.92, 950);
    const centerX = (this.width - centerWidth) / 2;
    const centerY = (this.height - centerHeight) / 2;

    table.rect(centerX, centerY, centerWidth, centerHeight);
    table.fill(0x2d7a4f);
    table.stroke({ width: 4, color: 0x8B7355 });

    this.container.addChild(table);

    // 非同步載入中央裝飾圖片
    this.loadCenterImage();

    // 玩家區域標籤已移除（由 Player 資訊條取代）
  }

  async loadCenterImage() {
    try {
      // 載入三個中央圖片，每個圖片有不同的網格尺寸
      const imageConfigs = [
        { file: '/assets/ui/bg_center.jpg', cols: 3, rows: 4 },    // 12個小圖
        { file: '/assets/ui/bg_center_2.jpg', cols: 4, rows: 4 },  // 16個小圖
        { file: '/assets/ui/bg_center_3.jpg', cols: 3, rows: 5 }   // 15個小圖
      ];

      // 遍歷載入每個圖片
      for (const config of imageConfigs) {
        const texture = await Assets.load(config.file);
        const cellWidth = texture.width / config.cols;
        const cellHeight = texture.height / config.rows;

        // 設定目標顯示尺寸為第一張圖片小圖尺寸的75%
        if (this.targetDisplaySize === 0) {
          this.targetDisplaySize = cellWidth * 0.75;
        }

        // 創建紋理並儲存元數據
        for (let row = 0; row < config.rows; row++) {
          for (let col = 0; col < config.cols; col++) {
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
            // 儲存原始尺寸資訊
            this.textureMetadata.push({ cellWidth, cellHeight });
          }
        }
      }

      // 創建中央 sprite
      this.centerSprite = new Sprite(this.centerTextures[0]);
      this.centerSprite.anchor.set(0.5);
      this.centerSprite.x = this.width / 2;
      this.centerSprite.y = this.height / 2;

      // 設定初始縮放
      this.updateSpriteScale();

      this.container.addChild(this.centerSprite);

      // 啟動輪播（每2秒切換一張圖）
      this.startCarousel();
    } catch (error) {
      console.warn('無法載入中央裝飾圖片:', error);
      // 如果載入失敗，顯示原來的文字
      this.showFallbackCenter();
    }
  }

  updateSpriteScale() {
    if (!this.centerSprite || this.textureMetadata.length === 0) return;

    // 獲取當前紋理的原始尺寸
    const metadata = this.textureMetadata[this.currentTextureIndex];
    // 計算縮放比例，使所有圖片都顯示為目標尺寸
    const scale = this.targetDisplaySize / metadata.cellWidth;
    this.centerSprite.scale.set(scale);
  }

  startCarousel() {
    setInterval(() => {
      this.currentTextureIndex = (this.currentTextureIndex + 1) % this.centerTextures.length;
      if (this.centerSprite) {
        this.centerSprite.texture = this.centerTextures[this.currentTextureIndex];
        // 更新縮放以保持統一大小
        this.updateSpriteScale();
      }
    }, 2000); // 每2秒切換一張
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

    // 重新創建桌面
    this.container.removeChildren();
    this.create();
  }
}