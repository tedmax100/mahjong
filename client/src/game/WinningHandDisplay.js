import { Container, Text, Graphics, Sprite, Assets } from 'pixi.js';
import { Tile } from './Tile.js';
import { MahjongLogic } from './MahjongLogic.js';

/**
 * 負責顯示胡牌手牌的 UI 邏輯
 */
export class WinningHandDisplay {
  constructor(winningHandContainer, appScreen, tileAssets) {
    this.container = winningHandContainer;
    this.appScreen = appScreen; // { width, height }
    this.tileAssets = tileAssets;
  }

  /**
   * 顯示胡牌手牌
   * @param {Array<string>} hand - 玩家胡牌時的手牌
   * @param {Array<Object>} melds - 吃碰槓牌組
   * @param {string} winTile - 胡的那張牌
   */
  async display(hand, melds, winTile) {
    this.container.removeChildren();

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
          const texture = this.tileAssets[tiles[i]] || this.tileAssets['back'];
          const tile = new Tile(tiles[i], texture);
          tile.setScale(tileScale);
          tile.container.x = i * (tileWidth + spacing);
          meldContainer.addChild(tile.container);
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
      const texture = this.tileAssets[tileType] || this.tileAssets['back'];
      const tile = new Tile(tileType, texture);
      tile.setScale(tileScale);
      tile.container.x = currentX;
      allTilesContainer.addChild(tile.container);
      currentX += tileWidth + spacing;
    }
    
    // Display the winning tile at the end, highlighted
    if (winTile) {
      currentX += meldSpacing;
      const texture = this.tileAssets[winTile] || this.tileAssets['back'];
      const winningTile = new Tile(winTile, texture);
      winningTile.setScale(tileScale);
      winningTile.container.x = currentX;

      const highlight = new Graphics();
      highlight.roundRect(-5, -5, 75 + 10, 95 + 10, 8); // A bit larger than the tile
      highlight.fill({color: 0xFFD700, alpha: 0.6});
      winningTile.container.addChildAt(highlight, 0); // Add behind tile sprites

      allTilesContainer.addChild(winningTile.container);
    }

    setTimeout(() => {
        allTilesContainer.x = centerX - allTilesContainer.width / 2;
        allTilesContainer.y = centerY - allTilesContainer.height / 2;
    }, 50);

    this.container.visible = true;
    setTimeout(() => {
      this.container.visible = false;
    }, 5000);
  }
}
