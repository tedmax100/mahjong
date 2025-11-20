import { Container, Text, Graphics } from 'pixi.js';
import { Tile } from './Tile.js';

/**
 * 玩家类
 */
export class Player {
  constructor(id, position, screenWidth, screenHeight) {
    this.id = id;
    this.userId = null; // 玩家的实际ID（从服务器获取）
    this.position = position; // 'bottom', 'right', 'top', 'left'
    this.screenWidth = screenWidth;
    this.screenHeight = screenHeight;
    this.container = new Container();
    this.tiles = [];
    this.discardedTiles = [];
    this.name = '';
    this.score = 1000;
    this.isInteractive = false; // 是否可以交互（轮到自己）

    this.infoText = null;

    this.createInfoDisplay();
  }

  createInfoDisplay() {
    // 玩家信息显示
    const bg = new Graphics();
    bg.rect(0, 0, 120, 60);
    bg.fill(0x333333, 0.8);
    bg.stroke({ width: 2, color: 0x666666 });

    this.infoText = new Text({
      text: '等待玩家...',
      style: {
        fontSize: 14,
        fill: 0xffffff,
        wordWrap: true,
        wordWrapWidth: 110
      }
    });
    this.infoText.x = 5;
    this.infoText.y = 5;

    bg.addChild(this.infoText);

    // 根据位置设置信息框位置
    this.positionInfoDisplay(bg);

    this.container.addChild(bg);
  }

  positionInfoDisplay(bg) {
    switch (this.position) {
      case 'bottom':
        bg.x = this.screenWidth / 2 - 60;
        bg.y = this.screenHeight - 200;
        break;
      case 'right':
        bg.x = this.screenWidth - 140;
        bg.y = this.screenHeight / 2 - 30;
        break;
      case 'top':
        bg.x = this.screenWidth / 2 - 60;
        bg.y = 20;
        break;
      case 'left':
        bg.x = 20;
        bg.y = this.screenHeight / 2 - 30;
        break;
    }
  }

  updateInfo(playerData) {
    this.userId = playerData.id; // 存储玩家ID
    this.name = playerData.name || '玩家';
    this.score = playerData.score || 1000;

    if (this.infoText) {
      this.infoText.text = `${this.name}\n分数: ${this.score}`;
    }
  }

  /**
   * 排序手牌（按照台灣麻將規則：萬 -> 筒 -> 條 -> 字牌 -> 花牌，每組內由小到大）
   */
  sortTiles(tilesData) {
    return tilesData.sort((a, b) => {
      // 定義花色順序
      const getSuitOrder = (tile) => {
        if (tile.startsWith('wan-')) return 1;      // 萬子
        if (tile.startsWith('tong-')) return 2;     // 筒子
        if (tile.startsWith('tiao-')) return 3;     // 條子
        // 字牌
        if (['dong', 'nan', 'xi', 'bei'].includes(tile)) return 4; // 風牌
        if (['zhong', 'fa', 'bai'].includes(tile)) return 5;       // 三元牌
        if (tile.startsWith('flower-')) return 6;   // 花牌
        return 7; // 其他
      };

      // 取得數字（如果有）
      const getNumber = (tile) => {
        const match = tile.match(/-(\d+)$/);
        return match ? parseInt(match[1]) : 0;
      };

      const suitA = getSuitOrder(a);
      const suitB = getSuitOrder(b);

      // 先比較花色
      if (suitA !== suitB) {
        return suitA - suitB;
      }

      // 同花色，比較數字
      return getNumber(a) - getNumber(b);
    });
  }

  setTiles(tilesData, tileAssets) {
    // 清除现有牌
    this.tiles.forEach(tile => tile.destroy());
    this.tiles = [];

    // 排序手牌
    const sortedTiles = this.sortTiles([...tilesData]);

    // 创建新牌
    sortedTiles.forEach((tileType) => {
      const texture = tileAssets[tileType] || tileAssets['back'];
      const tile = new Tile(tileType, texture);

      // 只有底部玩家（自己）的牌可以点击
      if (this.position === 'bottom') {
        tile.on('click', (clickedTile) => this.onTileClick(clickedTile));
      }

      this.tiles.push(tile);
      this.container.addChild(tile.container);
    });

    // 所有牌都加入後，統一設置位置
    this.tiles.forEach((tile, index) => {
      this.positionTile(tile, index);
    });
  }

  positionTile(tile, index) {
    const tileWidth = 60;
    const tileHeight = 80;

    // 動態調整間距：根據牌的數量和螢幕寬度
    const totalTiles = this.tiles.length;
    let spacing = 5; // 預設間距

    switch (this.position) {
      case 'bottom':
        // 底部 - 水平排列
        // 計算最佳間距，確保牌不重疊
        const availableWidth = this.screenWidth - 200; // 左右各留 100px 邊距
        const totalTileWidth = totalTiles * tileWidth;
        const totalSpacingWidth = availableWidth - totalTileWidth;

        if (totalSpacingWidth > 0) {
          // 有足夠空間，平均分配間距
          spacing = Math.min(totalSpacingWidth / (totalTiles - 1 || 1), 10); // 最大 10px
        } else {
          // 空間不足，縮小間距甚至重疊
          spacing = (availableWidth / totalTiles) - tileWidth;
        }

        const startX = this.screenWidth / 2 - (totalTiles * tileWidth + (totalTiles - 1) * spacing) / 2;

        tile.setPosition(
          startX + index * (tileWidth + spacing),
          this.screenHeight - 150
        );
        break;

      case 'right':
        // 右侧 - 垂直排列
        spacing = 8; // 右側固定間距
        tile.setPosition(
          this.screenWidth - 100,
          this.screenHeight / 2 - (totalTiles * (tileHeight * 0.8 + spacing)) / 2 + index * (tileHeight * 0.8 + spacing)
        );
        tile.setRotation(Math.PI / 2);
        tile.setScale(0.8);
        break;

      case 'top':
        // 顶部 - 水平排列（背面）
        spacing = 8; // 上方固定間距
        tile.setPosition(
          this.screenWidth / 2 - (totalTiles * (tileWidth * 0.8 + spacing)) / 2 + index * (tileWidth * 0.8 + spacing),
          50
        );
        tile.setScale(0.8);
        break;

      case 'left':
        // 左侧 - 垂直排列
        spacing = 8; // 左側固定間距
        tile.setPosition(
          50,
          this.screenHeight / 2 - (totalTiles * (tileHeight * 0.8 + spacing)) / 2 + index * (tileHeight * 0.8 + spacing)
        );
        tile.setRotation(-Math.PI / 2);
        tile.setScale(0.8);
        break;
    }
  }

  onTileClick(tile) {
    console.log('点击了牌:', tile.type);

    // 只有轮到自己时才能出牌
    if (!this.isInteractive) {
      console.log('还没轮到你，不能出牌！');
      return;
    }

    // 触发出牌事件（由Game类处理）
    if (this.onDiscard) {
      this.onDiscard(tile.type);
    }
  }

  setInteractive(interactive) {
    this.isInteractive = interactive;
    console.log(`玩家 ${this.name || this.id} 可交互状态: ${interactive}`);

    // 可以在这里添加视觉反馈，比如高亮手牌
    if (this.position === 'bottom') {
      this.tiles.forEach(tile => {
        if (interactive) {
          tile.container.alpha = 1.0; // 完全不透明
        } else {
          tile.container.alpha = 0.7; // 半透明表示不可交互
        }
      });
    }
  }

  addDiscardedTile(tile) {
    this.discardedTiles.push(tile);
    // TODO: 在中央区域显示打出的牌
  }

  removeTile(tileType) {
    // 从手牌数组中找到并移除该牌
    const index = this.tiles.findIndex(tile => tile.type === tileType);
    if (index !== -1) {
      const tile = this.tiles[index];
      // 从显示中移除
      this.container.removeChild(tile.container);
      // 从数组中移除
      this.tiles.splice(index, 1);

      // 重新排列剩余的牌
      this.tiles.forEach((tile, i) => {
        this.positionTile(tile, i);
      });
    }
  }

  /**
   * 加入一張新牌到手牌（摸牌）
   */
  addTile(tileType, tileAssets) {
    // 先排序：將新牌加入並排序
    const allTileTypes = [...this.tiles.map(t => t.type), tileType];
    const sortedTypes = this.sortTiles(allTileTypes);

    // 清除舊的牌（視覺上）
    this.tiles.forEach(tile => {
      this.container.removeChild(tile.container);
      tile.destroy();
    });
    this.tiles = [];

    // 按照排序後的順序重新創建所有牌
    sortedTypes.forEach((type) => {
      const texture = tileAssets[type] || tileAssets['back'];
      const tile = new Tile(type, texture);

      // 只有底部玩家（自己）的牌可以点击
      if (this.position === 'bottom') {
        tile.on('click', (clickedTile) => this.onTileClick(clickedTile));
      }

      // 先加入到陣列
      this.tiles.push(tile);

      // 加入到容器
      this.container.addChild(tile.container);
    });

    // 所有牌都加入後，統一設置位置
    this.tiles.forEach((tile, index) => {
      this.positionTile(tile, index);
    });

    console.log(`✅ 加入新牌完成，手牌數: ${this.tiles.length}`);
  }

  /**
   * 重新排列所有手牌（只重新計算位置，不重新創建）
   */
  rearrangeTiles() {
    // 只重新計算每張牌的位置
    this.tiles.forEach((tile, index) => {
      this.positionTile(tile, index);
    });
  }

  resize(width, height) {
    this.screenWidth = width;
    this.screenHeight = height;

    // 重新定位元素
    this.container.removeChildren();
    this.createInfoDisplay();

    // 重新定位手牌
    this.tiles.forEach((tile, index) => {
      this.positionTile(tile, index);
    });
  }
}
