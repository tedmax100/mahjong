import { Container, Text, Graphics, Sprite, Assets } from 'pixi.js';
import { Tile } from './Tile.js';
import { Table } from './Table.js';
import { Player } from './Player.js';

/**
 * 主游戏类
 */
export class Game {
  constructor(app, ws) {
    this.app = app;
    this.ws = ws;
    this.container = new Container();
    this.app.stage.addChild(this.container);

    this.table = null;
    this.players = [];
    this.currentPlayer = null;
    this.myPosition = 0; // 0=下, 1=右, 2=上, 3=左
    this.currentTurn = 0; // 当前轮到谁（0-3）
    this.tileAssets = {};
    this.discardedTiles = []; // 存储所有打出的牌的sprite
    this.discardContainer = new Container(); // 中央弃牌区域容器

    // 牌山相關
    this.wallContainer = new Container(); // 牌山容器
    this.wallTiles = []; // 牌山的 sprite
    this.remainingTiles = 144; // 剩餘可摸牌數（初始144張）
    this.wallText = null; // 顯示剩餘牌數的文字
    this.tilePool = []; // 可用牌池（用於摸牌）

    // 莊家相關
    this.dealerPosition = 0; // 莊家位置（0-3）
    this.dealerFirstDiscard = true; // 莊家是否還沒打過第一張牌
  }

  async init() {
    // 加载素材
    await this.loadAssets();

    // 创建牌桌
    this.table = new Table(this.app.screen.width, this.app.screen.height);
    this.container.addChild(this.table.container);

    // 添加牌山容器
    this.container.addChild(this.wallContainer);

    // 添加弃牌区域容器
    this.container.addChild(this.discardContainer);

    // 创建玩家区域
    this.createPlayers();

    // 創建牌山視覺
    this.createWalls();

    // 显示等待文字
    this.showWaitingText();
  }

  async loadAssets() {
    console.log('开始加载麻将牌素材...');

    const tileTypes = [
      // 万子
      ...Array.from({ length: 9 }, (_, i) => `wan-${i + 1}`),
      // 筒子
      ...Array.from({ length: 9 }, (_, i) => `tong-${i + 1}`),
      // 条子
      ...Array.from({ length: 9 }, (_, i) => `tiao-${i + 1}`),
      // 风牌
      'dong', 'nan', 'xi', 'bei',
      // 三元牌
      'zhong', 'fa', 'bai',
      // 花牌
      'flower-chun', 'flower-xia', 'flower-qiu', 'flower-dong',
      'flower-mei', 'flower-lan', 'flower-zhu', 'flower-ju',
      // 牌背
      'back'
    ];

    // 加载所有素材
    for (const type of tileTypes) {
      try {
        const texture = await Assets.load(`/assets/tiles/${type}.png`);
        this.tileAssets[type] = texture;
      } catch (error) {
        console.warn(`加载素材失败: ${type}.png`, error);
        // 创建占位纹理
        this.tileAssets[type] = this.createPlaceholderTexture(type);
      }
    }

    console.log('素材加载完成');
  }

  createPlaceholderTexture(type) {
    // 创建临时占位图
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

    return this.app.renderer.generateTexture(graphics);
  }

  createPlayers() {
    const positions = ['bottom', 'right', 'top', 'left'];

    for (let i = 0; i < 4; i++) {
      const player = new Player(i, positions[i], this.app.screen.width, this.app.screen.height);

      // 设置底部玩家（自己）的出牌回调
      if (i === 0) {
        player.onDiscard = (tileType) => this.handlePlayerDiscard(tileType);
      }

      this.players.push(player);
      this.container.addChild(player.container);
    }
  }

  handlePlayerDiscard(tileType) {
    console.log('玩家打出:', tileType);

    // 通过WebSocket发送出牌消息
    if (this.ws) {
      this.ws.sendAction('discard', { tile: tileType });
    }

    // 臨時方案：如果伺服器還沒實作摸牌邏輯，在本地自動模擬摸牌
    // TODO: 當伺服器實作後，移除這段程式碼
    setTimeout(() => {
      this.simulateDrawTile();
    }, 500); // 延遲 0.5 秒模擬摸牌
  }

  /**
   * 臨時方案：模擬摸牌（開發測試用）
   * 當伺服器實作摸牌邏輯後，應移除此方法
   */
  simulateDrawTile() {
    // 檢查是否是莊家的第一次打牌
    if (this.dealerFirstDiscard && this.myPosition === this.dealerPosition) {
      console.log('🎴 莊家第一次打牌，不摸牌（17張→16張）');
      this.dealerFirstDiscard = false; // 之後可以正常摸牌
      return; // 不摸牌，直接返回
    }

    // 檢查牌池是否還有牌
    if (this.tilePool.length === 0) {
      console.warn('⚠️ 牌池已空，無法摸牌');
      return;
    }

    // 從牌池中取出一張牌（從尾部取出，模擬從牌山摸牌）
    const drawnTile = this.tilePool.pop();

    console.log(`🎲 模擬摸牌: ${drawnTile}（牌池剩餘: ${this.tilePool.length}張）`);

    // 底部玩家（自己）摸牌
    const player = this.players[0];
    if (player && this.tileAssets[drawnTile]) {
      player.addTile(drawnTile, this.tileAssets);

      // 注意：剩餘牌數已經在 handleDiscard 中更新了，這裡不需要再更新

      console.log(`✅ 摸牌完成！手牌數量: ${player.tiles.length}，牌池剩餘: ${this.tilePool.length}張`);
    } else if (!this.tileAssets[drawnTile]) {
      console.warn(`⚠️ 找不到牌的素材: ${drawnTile}`);
      // 放回牌池
      this.tilePool.push(drawnTile);
    }
  }

  showWaitingText() {
    const text = new Text({
      text: '等待其他玩家加入...',
      style: {
        fontSize: 32,
        fill: 0xffffff,
        fontWeight: 'bold'
      }
    });

    text.x = this.app.screen.width / 2 - text.width / 2;
    text.y = this.app.screen.height / 2 - text.height / 2;

    this.waitingText = text;
    this.container.addChild(text);
  }

  updatePlayers(playersData) {
    console.log('更新玩家信息:', playersData);

    playersData.forEach((playerData, index) => {
      if (this.players[index]) {
        this.players[index].updateInfo(playerData);
      }
    });
  }

  /**
   * 初始化牌池（144 張牌）
   */
  initializeTilePool() {
    this.tilePool = [];

    // 萬子（1-9，各 4 張）
    for (let num = 1; num <= 9; num++) {
      for (let i = 0; i < 4; i++) {
        this.tilePool.push(`wan-${num}`);
      }
    }

    // 筒子（1-9，各 4 張）
    for (let num = 1; num <= 9; num++) {
      for (let i = 0; i < 4; i++) {
        this.tilePool.push(`tong-${num}`);
      }
    }

    // 條子（1-9，各 4 張）
    for (let num = 1; num <= 9; num++) {
      for (let i = 0; i < 4; i++) {
        this.tilePool.push(`tiao-${num}`);
      }
    }

    // 風牌（東南西北，各 4 張）
    ['dong', 'nan', 'xi', 'bei'].forEach(tile => {
      for (let i = 0; i < 4; i++) {
        this.tilePool.push(tile);
      }
    });

    // 三元牌（中發白，各 4 張）
    ['zhong', 'fa', 'bai'].forEach(tile => {
      for (let i = 0; i < 4; i++) {
        this.tilePool.push(tile);
      }
    });

    // 洗牌（隨機打亂）
    for (let i = this.tilePool.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1));
      [this.tilePool[i], this.tilePool[j]] = [this.tilePool[j], this.tilePool[i]];
    }

    console.log(`✅ 牌池初始化完成：共 ${this.tilePool.length} 張牌`);
  }

  /**
   * 從牌池中移除指定的牌（移除一張）
   * @param {string} tile - 牌的類型（如 'dong', 'wan-1'）
   * @returns {boolean} - 是否成功移除
   */
  removeTileFromPool(tile) {
    const index = this.tilePool.indexOf(tile);
    if (index !== -1) {
      this.tilePool.splice(index, 1);
      return true;
    }
    return false;
  }

  startGame(data) {
    console.log('游戏开始!', data);

    // 移除等待文字
    if (this.waitingText) {
      this.container.removeChild(this.waitingText);
      this.waitingText = null;
    }

    this.myPosition = data.myPosition || 0;
    this.currentPlayer = data.currentPlayer;
    this.currentTurn = data.currentTurn || 0;

    // 設定莊家位置（通常是第一個玩家，或由伺服器指定）
    this.dealerPosition = data.dealerPosition || 0;
    this.dealerFirstDiscard = true; // 重置莊家第一次打牌標記

    console.log(`我的位置: ${this.myPosition}, 莊家位置: ${this.dealerPosition}, 当前轮次: ${this.currentTurn}`);

    if (this.myPosition === this.dealerPosition) {
      console.log('🎴 你是莊家！起手 17 張，第一次打牌後不摸牌');
    }

    // 初始化牌池（臨時方案：用於本地模擬摸牌）
    this.initializeTilePool();

    // 計算已發出的牌數並更新剩餘牌數
    // 4個玩家 × 16張 + 莊家多1張 = 65張
    // 144張 - 65張 = 79張（不考慮花牌補牌）
    const dealtTiles = 65; // 簡化計算
    this.updateRemainingTiles(144 - dealtTiles);

    // 更新所有玩家的轮次状态
    this.updateTurnStatus();
  }

  dealTiles(data) {
    console.log('发牌:', data);

    const { tiles, position } = data;
    const player = this.players[position];

    if (player) {
      player.setTiles(tiles, this.tileAssets);

      // ✅ 從牌池中移除已發的牌
      tiles.forEach(tile => {
        const removed = this.removeTileFromPool(tile);
        if (!removed) {
          console.warn(`⚠️ 無法從牌池移除 ${tile}（可能已經被移除或不存在）`);
        }
      });

      console.log(`✅ 玩家 ${position} 發牌完成，已從牌池移除 ${tiles.length} 張牌。牌池剩餘: ${this.tilePool.length}張`);
    }
  }

  handlePlayerAction(data) {
    console.log('玩家动作:', data);

    const { playerId, action, tile, currentTurn } = data;

    // 更新当前轮次
    if (currentTurn !== undefined) {
      this.currentTurn = currentTurn;
      console.log(`更新轮次: ${this.currentTurn}`);
      this.updateTurnStatus();
    }

    switch (action) {
      case 'discard':
        this.handleDiscard(playerId, tile);
        break;
      case 'draw':
        this.handleDraw(playerId, tile);
        break;
      case 'pong':
        this.handlePong(playerId, tile);
        break;
      case 'kong':
        this.handleKong(playerId, tile);
        break;
      case 'hu':
        this.handleHu(playerId, tile);
        break;
    }
  }

  updateTurnStatus() {
    // 更新所有玩家的可交互状态
    this.players.forEach((player, index) => {
      const isMyTurn = (index === this.currentTurn && index === this.myPosition);
      player.setInteractive(isMyTurn);
    });
  }

  handleDiscard(playerId, tile) {
    console.log(`玩家 ${playerId} 打出了 ${tile}`);

    // 找到打出牌的玩家
    let playerPosition = -1;
    for (let i = 0; i < this.players.length; i++) {
      if (this.players[i].userId === playerId) {
        playerPosition = i;
        break;
      }
    }

    if (playerPosition === -1) {
      console.error('未找到玩家:', playerId);
      return;
    }

    // 创建弃牌sprite
    const texture = this.tileAssets[tile] || this.tileAssets['back'];
    const tileSprite = new Sprite(texture);

    // 设置弃牌大小（參考圖片樣式，縮小一點）
    const scale = 0.6;
    tileSprite.scale.set(scale);
    tileSprite.anchor.set(0.5); // 設置錨點在中心

    // 计算弃牌位置（在中央区域，根据玩家位置排列）
    const centerX = this.app.screen.width / 2;
    const centerY = this.app.screen.height / 2;
    const tileWidth = 60 * scale;
    const tileHeight = 80 * scale;
    const spacing = 3; // 更緊湊的間距

    // 计算该玩家已经打出的牌数
    const playerDiscards = this.discardedTiles.filter(d => d.playerPosition === playerPosition);
    const discardIndex = playerDiscards.length;

    // 根据玩家位置计算弃牌位置
    let x, y;
    const maxTilesPerRow = 10; // 每行最多10張牌（參考圖片）
    const row = Math.floor(discardIndex / maxTilesPerRow);
    const col = discardIndex % maxTilesPerRow;

    // 所有弃牌都保持正向（不旋轉），參考圖片的風格
    switch (playerPosition) {
      case 0: // 底部玩家 - 弃牌放在中央偏下，从左到右排列
        x = centerX - (maxTilesPerRow * (tileWidth + spacing)) / 2 + col * (tileWidth + spacing) + tileWidth / 2;
        y = centerY + 80 + row * (tileHeight + spacing);
        break;

      case 1: // 右侧玩家 - 弃牌放在中央偏右，从左到右排列
        x = centerX + 90 + col * (tileWidth + spacing);
        y = centerY - (maxTilesPerRow * (tileHeight + spacing)) / 2 + row * (tileHeight + spacing);
        break;

      case 2: // 顶部玩家 - 弃牌放在中央偏上，从左到右排列
        x = centerX - (maxTilesPerRow * (tileWidth + spacing)) / 2 + col * (tileWidth + spacing) + tileWidth / 2;
        y = centerY - 100 - row * (tileHeight + spacing);
        break;

      case 3: // 左侧玩家 - 弃牌放在中央偏左，从左到右排列
        x = centerX - 110 - (maxTilesPerRow - 1 - col) * (tileWidth + spacing);
        y = centerY - (maxTilesPerRow * (tileHeight + spacing)) / 2 + row * (tileHeight + spacing);
        break;
    }

    tileSprite.x = x;
    tileSprite.y = y;

    // 记录弃牌信息
    this.discardedTiles.push({
      sprite: tileSprite,
      playerPosition: playerPosition,
      tile: tile
    });

    // 添加到弃牌区域
    this.discardContainer.addChild(tileSprite);

    // 从玩家手牌中移除该牌（视觉上）
    const player = this.players[playerPosition];
    if (player) {
      player.removeTile(tile);
    }

    // 更新剩餘牌數（任何玩家打牌後，下家都會摸一張牌）
    if (this.remainingTiles > 0) {
      this.updateRemainingTiles(this.remainingTiles - 1);
      console.log(`📊 玩家 ${playerPosition} 打牌，下家摸牌，剩餘: ${this.remainingTiles}張`);
    }
  }

  handleDraw(playerId, tile) {
    console.log(`玩家 ${playerId} 摸牌: ${tile}`);

    // 找到摸牌的玩家
    let playerPosition = -1;
    for (let i = 0; i < this.players.length; i++) {
      if (this.players[i].userId === playerId) {
        playerPosition = i;
        break;
      }
    }

    if (playerPosition === -1) {
      console.error('未找到玩家:', playerId);
      return;
    }

    const player = this.players[playerPosition];

    // 如果是自己，顯示摸到的牌；如果是其他玩家，顯示牌背
    const tileToAdd = (playerPosition === this.myPosition) ? tile : 'back';

    // 加入新牌到手牌
    player.addTile(tileToAdd, this.tileAssets);

    // 更新剩餘牌數
    if (this.remainingTiles > 0) {
      this.updateRemainingTiles(this.remainingTiles - 1);
    }

    console.log(`玩家 ${playerPosition} 摸牌完成，剩餘 ${this.remainingTiles} 張`);
  }

  handlePong(playerId, tile) {
    // 处理碰牌
    console.log(`玩家 ${playerId} 碰了 ${tile}`);
  }

  handleKong(playerId, tile) {
    // 处理杠牌
    console.log(`玩家 ${playerId} 杠了 ${tile}`);
  }

  handleHu(playerId, tile) {
    // 处理胡牌
    console.log(`玩家 ${playerId} 胡了 ${tile}`);
  }

  gameOver(data) {
    console.log('游戏结束:', data);

    const { winner, winType, points } = data;

    // 显示游戏结果
    const resultText = new Text({
      text: `游戏结束!\n${winner} 胡牌 (${winType})\n得分: ${points}`,
      style: {
        fontSize: 36,
        fill: 0xffffff,
        fontWeight: 'bold',
        align: 'center'
      }
    });

    resultText.x = this.app.screen.width / 2 - resultText.width / 2;
    resultText.y = this.app.screen.height / 2 - resultText.height / 2;

    this.container.addChild(resultText);
  }

  /**
   * 創建牌山視覺呈現
   */
  createWalls() {
    const centerX = this.app.screen.width / 2;
    const centerY = this.app.screen.height / 2;
    const tileWidth = 60;
    const tileHeight = 80;

    // 根據螢幕大小計算牌山位置（更靠近邊緣）
    const wallDistanceVertical = Math.min(centerY - 100, 350);   // 上下方向
    const wallDistanceHorizontal = Math.min(centerX - 150, 400); // 左右方向
    const tilesPerSide = 18; // 每邊18張牌（144張 / 4邊 / 2層）

    // 獲取牌背紋理
    const backTexture = this.tileAssets['back'];

    if (!backTexture) {
      console.warn('牌背素材未載入，無法創建牌山');
      return;
    }

    // 清空舊的牌山
    this.wallContainer.removeChildren();
    this.wallTiles = [];

    // 四個方向的牌山（根據參考圖調整位置）
    const positions = [
      { name: 'bottom', x: centerX, y: centerY + wallDistanceVertical, rotation: 0 },
      { name: 'right', x: centerX + wallDistanceHorizontal, y: centerY, rotation: Math.PI / 2 },
      { name: 'top', x: centerX, y: centerY - wallDistanceVertical, rotation: Math.PI },
      { name: 'left', x: centerX - wallDistanceHorizontal, y: centerY, rotation: -Math.PI / 2 }
    ];

    positions.forEach((pos) => {
      // 創建每邊的牌山容器
      const sideContainer = new Container();
      sideContainer.x = pos.x;
      sideContainer.y = pos.y;
      sideContainer.rotation = pos.rotation;

      // 繪製兩層牌
      for (let layer = 0; layer < 2; layer++) {
        for (let i = 0; i < tilesPerSide; i++) {
          const tile = new Sprite(backTexture);
          tile.width = tileWidth * 0.5; // 縮小一點
          tile.height = tileHeight * 0.5;
          tile.anchor.set(0.5);

          // 水平排列
          tile.x = (i - tilesPerSide / 2) * (tileWidth * 0.5 + 2);
          tile.y = -layer * 5; // 第二層稍微往上偏移，製造堆疊效果

          sideContainer.addChild(tile);
        }
      }

      this.wallContainer.addChild(sideContainer);
    });

    // 創建剩餘牌數顯示
    this.createRemainingTilesText();

    console.log('牌山創建完成');
  }

  /**
   * 創建剩餘牌數文字顯示
   */
  createRemainingTilesText() {
    // 移除舊的顯示
    if (this.wallTextContainer) {
      this.container.removeChild(this.wallTextContainer);
    }

    // 創建新的容器（獨立於 wallContainer，確保在最上層）
    this.wallTextContainer = new Container();

    // 創建背景
    const bg = new Graphics();
    bg.roundRect(-80, -20, 160, 40, 10);
    bg.fill({ color: 0x000000, alpha: 0.7 });
    bg.stroke({ width: 2, color: 0xFFD700 }); // 金色邊框

    // 創建文字
    this.wallText = new Text({
      text: `海底: ${this.remainingTiles}張`,
      style: {
        fontSize: 22,
        fill: 0xFFFFFF,
        fontWeight: 'bold',
        stroke: 0x000000,
        strokeThickness: 2
      }
    });

    this.wallText.anchor.set(0.5);
    this.wallText.x = 0;
    this.wallText.y = 0;

    // 加入到容器
    this.wallTextContainer.addChild(bg);
    this.wallTextContainer.addChild(this.wallText);

    // 設置位置（畫面上方中央）
    this.wallTextContainer.x = this.app.screen.width / 2;
    this.wallTextContainer.y = 40;

    // 加入到主容器的最上層
    this.container.addChild(this.wallTextContainer);

    console.log(`✅ 剩餘牌數顯示已創建: ${this.remainingTiles}張`);
  }

  /**
   * 更新剩餘牌數
   */
  updateRemainingTiles(count) {
    this.remainingTiles = count;
    if (this.wallText) {
      this.wallText.text = `海底: ${this.remainingTiles}張`;
      console.log(`🎲 剩餘牌數更新: ${this.remainingTiles}張`);
    }
  }

  resize(width, height) {
    if (this.table) {
      this.table.resize(width, height);
    }

    this.players.forEach(player => {
      player.resize(width, height);
    });

    // 重新創建牌山以適應新的螢幕尺寸
    if (this.wallContainer) {
      this.createWalls();
    }
  }
}
