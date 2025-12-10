import { Graphics, Text, Assets } from 'pixi.js';

/**
 * 負責載入和管理遊戲素材
 */
export class AssetLoader {
  constructor(renderer) {
    this.renderer = renderer;
    this.tileMapping = {
      // 萬子 (w = 萬)
      'wan-1': '1wf', 'wan-2': '2wf', 'wan-3': '3wf',
      'wan-4': '4wf', 'wan-5': '5wf', 'wan-6': '6wf',
      'wan-7': '7wf', 'wan-8': '8wf', 'wan-9': '9wf',

      // 筒子 (t = 筒)
      'tong-1': '1tf', 'tong-2': '2tf', 'tong-3': '3tf',
      'tong-4': '4tf', 'tong-5': '5tf', 'tong-6': '6tf',
      'tong-7': '7tf', 'tong-8': '8tf', 'tong-9': '9tf',

      // 條子 (tt = 條)
      'tiao-1': '1ttf', 'tiao-2': '2ttf', 'tiao-3': '3ttf',
      'tiao-4': '4ttf', 'tiao-5': '5ttf', 'tiao-6': '6ttf',
      'tiao-7': '7ttf', 'tiao-8': '8ttf', 'tiao-9': '9ttf',

      // 風牌 (z1-z4)
      'dong': 'z1f', 'nan': 'z2f', 'xi': 'z3f', 'bei': 'z4f',

      // 三元牌 (z5-z7)
      'zhong': 'z5f', 'fa': 'z6f', 'bai': 'z7f',

      // 花牌
      'flower-chun': 'chun', 'flower-xia': 'xia', 'flower-qiu': 'qiu', 'flower-dong': 'dong',
      'flower-mei': 'mei', 'flower-lan': 'lan', 'flower-zhu': 'zhu', 'flower-ju': 'ju',

      // 牌背 (使用大尺寸的牌背圖片，與其他牌面尺寸一致)
      'back': 'pbaseBig'
    };
  }

  /**
   * 載入所有麻將牌素材
   * @returns {Promise<Object>} 包含所有紋理的物件 { tileType: Texture }
   */
  async load() {
    console.log('開始載入麻將牌素材...');
    const tileAssets = {};

    for (const [tileType, fileName] of Object.entries(this.tileMapping)) {
      try {
        const texture = await Assets.load(`/assets/tiles/carddown/${fileName}.png`);
        tileAssets[tileType] = texture;
      } catch (error) {
        console.warn(`載入素材失敗: ${fileName}.png`, error);
        // 創建佔位紋理
        tileAssets[tileType] = this.createPlaceholderTexture(tileType);
      }
    }

    // 預先載入牌底素材（如果需要）
    try {
        await Assets.load('/assets/tiles/carddown/basefdown.png');
    } catch (e) {
        console.warn('無法預載入牌底素材');
    }

    console.log('素材載入完成');
    return tileAssets;
  }

  /**
   * 創建臨時佔位圖
   * @param {string} type - 牌的類型
   * @returns {Texture}
   */
  createPlaceholderTexture(type) {
    const graphics = new Graphics();
    graphics.rect(0, 0, 60, 80);
    graphics.fill(0xF5F0E8);
    graphics.stroke({ width: 2, color: 0x8B7355 });

    const text = new Text({
      text: type,
      style: {
        fontSize: 10,
        fill: 0x000000,
        wordWrap: true,
        wordWrapWidth: 55
      }
    });
    text.x = 30 - text.width / 2;
    text.y = 40 - text.height / 2;
    graphics.addChild(text);

    return this.renderer.generateTexture(graphics);
  }
}
