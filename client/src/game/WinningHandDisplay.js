import { Container, Graphics, Sprite, Assets } from 'pixi.js';
import { MahjongLogic } from './MahjongLogic.js';

/**
 * 負責顯示胡牌手牌的 UI 邏輯
 */
export class WinningHandDisplay {
  constructor(winningHandContainer, appScreen, tileAssets) {
    this.container = winningHandContainer;
    this.appScreen = appScreen; // { width, height }
    this.tileAssets = tileAssets;
    this.baseTexture = null;
  }

  /**
   * 創建單張牌的顯示容器
   * @param {string} tileType - 牌型
   * @param {number} scale - 縮放比例
   * @returns {Container} 牌的容器
   */
  createTileDisplay(tileType, scale) {
    const tileContainer = new Container();

    // 添加牌底
    if (this.baseTexture) {
      const baseSprite = new Sprite(this.baseTexture);
      baseSprite.anchor.set(0, 0);
      baseSprite.scale.set(scale);
      tileContainer.addChild(baseSprite);
    }

    // 添加牌面
    const texture = this.tileAssets[tileType] || this.tileAssets['back'];
    if (texture) {
      const tileSprite = new Sprite(texture);
      tileSprite.anchor.set(0, 0);
      tileSprite.scale.set(scale);
      // 筒子微調
      if (tileType.startsWith('tong-')) {
        tileSprite.y = 8 * scale;
      }
      tileContainer.addChild(tileSprite);
    }

    return tileContainer;
  }

  /**
   * 顯示胡牌手牌
   * @param {Array<string>} hand - 玩家胡牌時的手牌
   * @param {Array<Object>} melds - 吃碰槓牌組
   * @param {string} winTile - 胡的那張牌
   */
  async display(hand, melds, winTile) {
    this.container.removeChildren();

    // 預先載入牌底紋理
    try {
      this.baseTexture = await Assets.load('/assets/tiles/carddown/basefdown.png');
    } catch (error) {
      console.warn('無法載入牌底圖片', error);
    }

    const centerX = this.appScreen.width / 2;
    const centerY = this.appScreen.height / 2;

    const bg = new Graphics();
    bg.rect(0, 0, this.appScreen.width, this.appScreen.height);
    bg.fill({ color: 0x000000, alpha: 0.7 });
    this.container.addChild(bg);

    const allTilesContainer = new Container();
    this.container.addChild(allTilesContainer);

    const tileScale = 0.8;
    const tileWidth = 75 * tileScale;
    const spacing = 8;
    const meldSpacing = 20;

    let currentX = 0;

    // Display Melds
    if (melds) {
      for (const meld of melds) {
        const meldContainer = new Container();
        const tiles = meld.Tiles || meld.tiles;
        for (let i = 0; i < tiles.length; i++) {
          const tileContainer = this.createTileDisplay(tiles[i], tileScale);
          tileContainer.x = i * (tileWidth + spacing);
          meldContainer.addChild(tileContainer);
        }
        meldContainer.x = currentX;
        allTilesContainer.addChild(meldContainer);
        currentX += meldContainer.width + meldSpacing;
      }
    }

    // Display Hand, separating the winning tile
    const handToShow = [...hand];
    const winTileIndex = handToShow.lastIndexOf(winTile);
    if (winTileIndex > -1) {
      handToShow.splice(winTileIndex, 1);
    }
    handToShow.sort((a, b) => (MahjongLogic.tileValue(a) - MahjongLogic.tileValue(b)));

    for (const tileType of handToShow) {
      const tileContainer = this.createTileDisplay(tileType, tileScale);
      tileContainer.x = currentX;
      allTilesContainer.addChild(tileContainer);
      currentX += tileWidth + spacing;
    }

    // Display the winning tile at the end, highlighted
    if (winTile) {
      currentX += meldSpacing;
      const winningTileContainer = this.createTileDisplay(winTile, tileScale);
      winningTileContainer.x = currentX;

      const highlight = new Graphics();
      highlight.roundRect(-5, -5, 75 * tileScale + 10, 95 * tileScale + 10, 8);
      highlight.fill({color: 0xFFD700, alpha: 0.6});
      winningTileContainer.addChildAt(highlight, 0);

      allTilesContainer.addChild(winningTileContainer);
    }

    // 立即置中
    allTilesContainer.x = centerX - allTilesContainer.width / 2;
    allTilesContainer.y = centerY - allTilesContainer.height / 2;

    this.container.visible = true;
    setTimeout(() => {
      this.container.visible = false;
    }, 5000);
  }
}
