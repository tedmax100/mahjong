import { Container, Text, Graphics } from 'pixi.js';
import { Tile } from './Tile.js';

/**
 * 玩家類
 */
export class Player {
  constructor(id, position, screenWidth, screenHeight) {
    this.id = id;
    this.userId = null; // 玩家的實際ID（從伺服器獲取）
    this.position = position; // 'bottom', 'right', 'top', 'left'
    this.screenWidth = screenWidth;
    this.screenHeight = screenHeight;
    this.container = new Container();
    this.tiles = [];
    this.discardedTiles = [];
    this.melds = []; // 吃/碰/槓的牌組 [{type: 'chow'|'pong'|'kong', tiles: [...]}]
    this.meldsContainer = new Container(); // 用於顯示吃/碰/槓牌組的容器
    this.name = '';
    this.score = 1000;
    this.isInteractive = false; // 是否可以互動（輪到自己）
    this.isTing = false; // 是否已宣告聽牌
    this.winningTiles = []; // 聽的牌
    this.lastDrawnTile = null; // 最後摸到的牌（用於聽牌後限制打牌）

    this.infoText = null;
    this.tingStatusText = null; // 聽牌狀態文字

    this.createInfoDisplay();
    this.container.addChild(this.meldsContainer);
  }

  createInfoDisplay() {
    // 玩家資訊顯示
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

    // 根據位置設定資訊框位置
    this.positionInfoDisplay(bg);

    this.container.addChild(bg);
  }

  positionInfoDisplay(bg) {
    switch (this.position) {
      case 'bottom':
        bg.x = this.screenWidth / 2 - 60;
        bg.y = this.screenHeight - 250; // Move up
        break;
      case 'right':
        bg.x = this.screenWidth - 140;
        bg.y = this.screenHeight / 2 - 30;
        break;
      case 'top':
        bg.x = this.screenWidth / 2 - 60;
        bg.y = 10;
        break;
      case 'left':
        bg.x = 10;
        bg.y = this.screenHeight / 2 - 30;
        break;
    }
  }

  updateInfo(playerData) {
    this.userId = playerData.id; // 儲存玩家ID
    this.name = playerData.name || '玩家';
    this.score = playerData.score || 1000;

    this.updateNameDisplay();
  }

  /**
   * 更新玩家名稱顯示
   */
  updateNameDisplay() {
    if (this.infoText) {
      this.infoText.text = `${this.name}\n分數: ${this.score}`;
    }
  }

  /**
   * 排序手牌（按照臺灣麻將規則：萬 -> 筒 -> 條 -> 字牌 -> 花牌，每組內由小到大）
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

  async setTiles(tilesData, tileAssets) {
    // 清除現有牌
    this.tiles.forEach(tile => tile.destroy());
    this.tiles = [];

    // 排序手牌
    const sortedTiles = this.sortTiles([...tilesData]);

    // 創建新牌（等待所有牌創建完成）
    for (const tileType of sortedTiles) {
      const texture = tileAssets[tileType] || tileAssets['back'];
      const tile = new Tile(tileType, texture);

      // 等待牌創建完成（包括牌底載入）
      await new Promise(resolve => setTimeout(resolve, 10));

      // 只有底部玩家（自己）的牌可以點擊
      if (this.position === 'bottom') {
        tile.on('click', (clickedTile) => this.onTileClick(clickedTile));
      }

      this.tiles.push(tile);
      this.container.addChild(tile.container);
    }

    // 所有牌都加入後，統一設定位置
    this.tiles.forEach((tile, index) => {
      this.positionTile(tile, index);
    });
  }

  positionTile(tile, index) {
    // 手牌包含牌底，實際佔用空間更大
    const tileWidth = 42.1875;  // 縮小至56.25%（原75）
    const tileHeight = 53.4375; // 縮小至56.25%（原95）

    // 動態調整間距：根據牌的數量和螢幕寬度
    const totalTiles = this.tiles.length;
    let spacing = 5; // 預設間距

    switch (this.position) {
      case 'bottom':
        // 底部 - 水平排列
        // 計算最佳間距，確保牌不重疊
        const availableWidth = this.screenWidth - 120; // 左右各留 60px 邊距（更靠近邊緣）
        const totalTileWidth = totalTiles * tileWidth;
        const totalSpacingWidth = availableWidth - totalTileWidth;

        if (totalSpacingWidth > 0) {
          // 有足夠空間，平均分配間距
          spacing = Math.min(totalSpacingWidth / (totalTiles - 1 || 1), 15); // 增加最大間距至15px
        } else {
          // 空間不足，縮小間距但不重疊
          spacing = Math.max((availableWidth / totalTiles) - tileWidth, 2); // 最小2px間距
        }

        const startX = this.screenWidth / 2 - (totalTiles * tileWidth + (totalTiles - 1) * spacing) / 2;

        tile.setPosition(
          startX + index * (tileWidth + spacing),
          this.screenHeight - 180  // 更靠近底部
        );
        tile.setScale(0.75); // 縮小牌底和牌面至75%
        break;

      case 'right':
        // 右側 - 垂直排列
        spacing = 12; // 增加間距避免重疊
        tile.setPosition(
          this.screenWidth - 70,  // 更靠近右邊（從 100 改為 70）
          this.screenHeight / 2 - (totalTiles * (tileHeight + spacing)) / 2 + index * (tileHeight + spacing)
        );
        tile.setRotation(Math.PI / 2);
        tile.setScale(0.6); // 縮小牌底和牌面至75% (0.8 * 0.75 = 0.6)
        break;

      case 'top':
        // 頂部 - 水平排列（背面）
        spacing = 12; // 增加間距避免重疊
        tile.setPosition(
          this.screenWidth / 2 - (totalTiles * (tileWidth + spacing)) / 2 + index * (tileWidth + spacing),
          30  // 更靠近頂部（從 50 改為 30）
        );
        tile.setScale(0.6); // 縮小牌底和牌面至75% (0.8 * 0.75 = 0.6)
        break;

      case 'left':
        // 左側 - 垂直排列
        spacing = 12; // 增加間距避免重疊
        tile.setPosition(
          30,  // 更靠近左邊（從 50 改為 30）
          this.screenHeight / 2 - (totalTiles * (tileHeight + spacing)) / 2 + index * (tileHeight + spacing)
        );
        tile.setRotation(-Math.PI / 2);
        tile.setScale(0.6); // 縮小牌底和牌面至75% (0.8 * 0.75 = 0.6)
        break;
    }
  }

  onTileClick(tile) {
    console.log('點擊了牌:', tile.type);

    // 只有輪到自己時才能出牌
    if (!this.isInteractive) {
      console.log('還沒輪到你，不能出牌！');
      return;
    }

    // 如果已宣告聽牌，只能打剛摸到的牌
    if (this.isTing) {
      if (tile.type !== this.lastDrawnTile) {
        console.log('已宣告聽牌，只能打剛摸到的牌！');
        // TODO: 顯示提示訊息給玩家
        return;
      }
    }

    // 觸發出牌事件（由Game類處理）
    if (this.onDiscard) {
      this.onDiscard(tile.type);
    }
  }

  setInteractive(interactive) {
    this.isInteractive = interactive;
    console.log(`玩家 ${this.name || this.id} 可互動狀態: ${interactive}`);

    // 可以在這裡新增視覺回饋，比如高亮手牌
    if (this.position === 'bottom') {
      this.tiles.forEach(tile => {
        // 實際禁用/啟用牌面的互動性
        if (tile.sprite) {
          tile.sprite.eventMode = interactive ? 'static' : 'none';
        }

        if (interactive) {
          tile.container.alpha = 1.0; // 完全不透明
        } else {
          tile.container.alpha = 0.7; // 半透明表示不可互動
        }
      });
    }
  }

  addDiscardedTile(tile) {
    this.discardedTiles.push(tile);
    // TODO: 在中央區域顯示打出的牌
  }

  removeTile(tileType) {
    console.log(`🗑️ [removeTile] 嘗試移除: ${tileType}, 當前手牌數: ${this.tiles.length}`);
    console.log(`🗑️ [removeTile] 當前手牌:`, this.tiles.map(t => t.type));

    // 從手牌陣列中找到並移除該牌
    const index = this.tiles.findIndex(tile => tile.type === tileType);
    if (index !== -1) {
      const tile = this.tiles[index];
      console.log(`✅ [removeTile] 找到牌，索引: ${index}`);

      // 從顯示中移除
      this.container.removeChild(tile.container);
      // 從陣列中移除
      this.tiles.splice(index, 1);

      // 重新排列剩餘的牌
      this.tiles.forEach((tile, i) => {
        this.positionTile(tile, i);
      });

      console.log(`✅ [removeTile] 移除成功，剩餘手牌數: ${this.tiles.length}`);
    } else {
      console.error(`❌ [removeTile] 找不到要移除的牌: ${tileType}`);
      console.error(`❌ [removeTile] 可能的原因: addTile 還未完成或牌已被移除`);
    }
  }

  /**
   * 加入一張新牌到手牌（摸牌）
   */
  async addTile(tileType, tileAssets) {
    // 記錄最後摸到的牌（用於聽牌後限制打牌）
    this.lastDrawnTile = tileType;

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
    for (const type of sortedTypes) {
      const texture = tileAssets[type] || tileAssets['back'];
      const tile = new Tile(type, texture);

      // 等待牌創建完成（包括牌底載入）
      await new Promise(resolve => setTimeout(resolve, 10));

      // 只有底部玩家（自己）的牌可以點擊
      if (this.position === 'bottom') {
        tile.on('click', (clickedTile) => this.onTileClick(clickedTile));

        // 設定初始互動狀態
        if (tile.sprite) {
          tile.sprite.eventMode = this.isInteractive ? 'static' : 'none';
        }
      }

      // 先加入到陣列
      this.tiles.push(tile);

      // 加入到容器
      this.container.addChild(tile.container);
    }

    // 所有牌都加入後，統一設定位置
    this.tiles.forEach((tile, index) => {
      this.positionTile(tile, index);
    });

    // 恢復聽牌狀態指示器（如果已宣告聽牌）
    if (this.isTing) {
      this.showTingStatus();
    }

    console.log(`✅ 加入新牌完成，手牌數: ${this.tiles.length}, 最後摸牌: ${this.lastDrawnTile}`);
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

  /**
   * 高亮顯示可以組合的牌組
   * @param {Array<Array<string>>} tileGroups - 牌組列表，例如 [['wan-1', 'wan-2', 'wan-3'], ['wan-4', 'wan-4']]
   */
  highlightTileGroups(tileGroups) {
    // 先重置所有牌的位置
    this.clearHighlight();

    if (!tileGroups || tileGroups.length === 0) {
      return;
    }

    // 將牌組扁平化，找出所有需要高亮的牌
    const tilesToHighlight = new Set();
    tileGroups.forEach(group => {
      group.forEach(tileType => tilesToHighlight.add(tileType));
    });

    // 對於底部玩家，將可組合的牌往上移動
    if (this.position === 'bottom') {
      this.tiles.forEach((tile, index) => {
        if (tilesToHighlight.has(tile.type)) {
          // 往上移動 20 像素
          tile.container.y -= 20;
          // 新增發光效果
          tile.container.alpha = 1.0;
        } else {
          // 其他牌變暗
          tile.container.alpha = 0.5;
        }
      });
    }

    console.log(`✨ 高亮顯示 ${tilesToHighlight.size} 張牌`);
  }

  /**
   * 高亮顯示指定的牌
   * @param {Array<string>} tilesToHighlight - 要高亮的牌的類型列表
   */
  highlightTiles(tilesToHighlight) {
    // 先重置所有牌的位置
    this.clearHighlight();

    if (!tilesToHighlight || tilesToHighlight.length === 0) {
      return;
    }

    const highlightSet = new Set(tilesToHighlight);

    // 對於底部玩家，將可高亮的牌往上移動
    if (this.position === 'bottom') {
      this.tiles.forEach((tile) => {
        if (highlightSet.has(tile.type)) {
          // 往上移動 20 像素
          tile.container.y -= 20;
          // 新增發光效果
          tile.container.alpha = 1.0;
        } else {
          // 其他牌變暗
          tile.container.alpha = 0.5;
        }
      });
    }

    console.log(`✨ 高亮顯示 ${highlightSet.size} 張牌`);
  }

  /**
   * 清除高亮效果
   */
  clearHighlight() {
    if (this.position === 'bottom') {
      this.tiles.forEach((tile, index) => {
        // 重新定位到正確位置
        this.positionTile(tile, index);
        // 恢復透明度
        tile.container.alpha = 1.0;
      });
    }
  }

  /**
   * 新增吃/碰/槓牌組
   * @param {Object} meldData - {type: 'chow'|'pong'|'kong', tiles: [...]} 
   * @param {Object} tileAssets - 牌面素材
   */
  async addMeld(meldData, tileAssets) {
    // Note: meldData keys from the server are capitalized (Type, Tiles), but local ones might be lowercase.
    const meldType = meldData.Type || meldData.type;
    const meldTiles = meldData.Tiles || meldData.tiles;

    // Handle promotion from pong to kong to prevent duplicate melds
    if (meldType === 'kong_promoted') {
      const tile = meldTiles[0];
      const pongIndex = this.melds.findIndex(m => {
        const mType = m.Type || m.type;
        const mTiles = m.Tiles || m.tiles;
        return mType === 'pong' && mTiles[0] === tile;
      });

      if (pongIndex !== -1) {
        // Replace the existing pong with the new promoted kong data
        this.melds[pongIndex] = meldData;
        console.log(`✅ Player ${this.id} promoted pong to kong for tile: ${tile}`);
      } else {
        // This case indicates a logic error upstream, but we add the meld to avoid data loss.
        console.error(`Could not find a pong to promote for tile ${tile}. Adding kong as a new meld.`);
        this.melds.push(meldData);
      }
    } else {
      // For chow, pong, or other new kong types, just add the new meld
      this.melds.push(meldData);
    }

    await this.displayMelds(tileAssets);
  }

  /**
   * 顯示所有吃/碰/槓牌組
   */
  async displayMelds(tileAssets) {
    this.meldsContainer.removeChildren();

    const tileWidth = 60;
    const tileHeight = 80;
    const groupSpacing = 20; // 增加牌組間距
    let currentOffset = 0;

    const { Sprite, Container, Assets } = await import('pixi.js');
    const baseTexture = await Assets.load('/assets/tiles/carddown/basefdown.png');

    for (const meld of this.melds) {
      const meldGroup = new Container();
      const meldType = meld.Type || meld.type;
      const meldTiles = meld.Tiles || meld.tiles;

      // Create tiles with bases
      for (let i = 0; i < meldTiles.length; i++) {
        const tileType = meldTiles[i];
        const tileContainer = new Container();
        const baseSprite = new Sprite(baseTexture);
        tileContainer.addChild(baseSprite);

        // For concealed kong, two tiles are face down.
        let texture;
        if (meldType === 'kong_concealed' && (i === 0 || i === 3)) {
            texture = tileAssets['back'];
        } else {
            texture = tileAssets[tileType];
            if (!texture) {
                console.warn(`⚠️ displayMelds: 找不到牌面貼圖 "${tileType}"，使用牌背。可用的 key:`, Object.keys(tileAssets || {}).slice(0, 10));
                texture = tileAssets['back'];
            }
        }
        
        const tileSprite = new Sprite(texture);
        if (tileType.startsWith('tong-')) {
            tileSprite.y = 5;
        }
        tileContainer.addChild(tileSprite);
        
        const isKong = meldType && meldType.includes('kong');
        const isFourthTile = i === 3;

        if (isKong && isFourthTile) {
            // For promoted kong, place on top of one of the existing 3 tiles.
            // For concealed kong, place on top of the middle ones.
            const baseTileIndex = 1;
            tileContainer.x = baseTileIndex * (tileWidth + 5);
            tileContainer.y = -tileHeight * 0.1;
        } else {
            tileContainer.x = i * (tileWidth + 5); // Add spacing
        }

        meldGroup.addChild(tileContainer);
      }
      
      const groupWidth = (meldTiles.length === 4 ? 3 : meldTiles.length) * (tileWidth + 5);

      switch (this.position) {
        case 'bottom':
          meldGroup.scale.set(0.8);
          meldGroup.x = this.screenWidth - 180 - currentOffset;
          meldGroup.y = this.screenHeight - 120;
          currentOffset += groupWidth * 0.8 + groupSpacing;
          break;
        case 'right':
          meldGroup.scale.set(0.7);
          meldGroup.x = this.screenWidth - 80;
          meldGroup.y = 150 + currentOffset;
          meldGroup.rotation = Math.PI / 2;
          currentOffset += groupWidth * 0.7 + groupSpacing;
          break;
        case 'top':
           meldGroup.scale.set(0.8);
          meldGroup.x = 150 + currentOffset;
          meldGroup.y = 100;
          meldGroup.rotation = Math.PI;
          currentOffset += groupWidth * 0.8 + groupSpacing;
          break;
        case 'left':
          meldGroup.scale.set(0.7);
          meldGroup.x = 80;
          meldGroup.y = this.screenHeight - 150 - currentOffset;
          meldGroup.rotation = -Math.PI / 2;
          currentOffset += groupWidth * 0.7 + groupSpacing;
          break;
      }
      this.meldsContainer.addChild(meldGroup);
    }
  }

  /**
   * 顯示聽牌狀態
   */
  showTingStatus() {
    // 如果已經有聽牌狀態文字，先移除
    if (this.tingStatusText) {
      this.container.removeChild(this.tingStatusText);
    }

    // 創建聽牌狀態文字
    this.tingStatusText = new Text({
      text: '聽',
      style: {
        fontSize: 32,
        fill: 0xFF0000, // 紅色
        fontWeight: 'bold',
        stroke: 0xFFFFFF,
        strokeThickness: 3
      }
    });
    this.tingStatusText.anchor.set(0.5);

    // 根據位置設定聽牌圖示位置
    switch (this.position) {
      case 'bottom':
        this.tingStatusText.x = this.screenWidth / 2;
        this.tingStatusText.y = this.screenHeight - 120;
        break;
      case 'right':
        this.tingStatusText.x = this.screenWidth - 60;
        this.tingStatusText.y = this.screenHeight / 2;
        break;
      case 'top':
        this.tingStatusText.x = this.screenWidth / 2;
        this.tingStatusText.y = 80;
        break;
      case 'left':
        this.tingStatusText.x = 60;
        this.tingStatusText.y = this.screenHeight / 2;
        break;
    }

    this.container.addChild(this.tingStatusText);
    console.log(`✅ 顯示聽牌狀態: ${this.name || this.id}`);
  }

  /**
   * 隱藏聽牌狀態
   */
  hideTingStatus() {
    if (this.tingStatusText) {
      this.container.removeChild(this.tingStatusText);
      this.tingStatusText = null;
    }
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

    // 重新顯示聽牌狀態（如果有）
    if (this.isTing) {
      this.showTingStatus();
    }
  }

  reset() {
    // 清除手牌
    this.tiles.forEach(tile => tile.destroy());
    this.tiles = [];

    // 清除吃碰槓
    this.meldsContainer.removeChildren();
    this.melds = [];

    // 重置狀態
    this.isTing = false;
    this.winningTiles = [];
    this.lastDrawnTile = null;
    this.hideTingStatus();
  }
}