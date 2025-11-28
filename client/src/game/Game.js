import { Container, Text, Graphics, Sprite, Assets } from 'pixi.js';
import { Tile } from './Tile.js';
import { Table } from './Table.js';
import { Player } from './Player.js';
import { ActionButtons } from './ActionButtons.js';

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

    // 動作按鈕
    this.actionButtons = null;
    this.lastDiscardedTile = null; // 最後被打出的牌
    this.pendingActions = []; // 可執行的動作列表
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

    // 創建動作按鈕（異步初始化）
    this.actionButtons = new ActionButtons(this.app.screen.width, this.app.screen.height);
    this.container.addChild(this.actionButtons.container);

    // 設置按鈕回調
    this.actionButtons.on('pong', () => this.handlePongAction());
    this.actionButtons.on('chow', () => this.handleChowAction());
    this.actionButtons.on('kong', () => this.handleKongAction());
    this.actionButtons.on('hu', () => this.handleHuAction());
    this.actionButtons.on('cancel', () => this.handleCancelAction());

    // 显示等待文字
    this.showWaitingText();

    console.log('✅ Game initialized successfully');
  }

  async loadAssets() {
    console.log('开始加载麻将牌素材...');

    // 映射內部牌名到新圖片檔名
    const tileMapping = {
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

      // 牌背
      'back': 'pback1'
    };

    // 加载所有素材
    for (const [tileType, fileName] of Object.entries(tileMapping)) {
      try {
        const texture = await Assets.load(`/assets/tiles/carddown/${fileName}.png`);
        this.tileAssets[tileType] = texture;
      } catch (error) {
        console.warn(`加载素材失败: ${fileName}.png`, error);
        // 创建占位纹理
        this.tileAssets[tileType] = this.createPlaceholderTexture(tileType);
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

      // 摸牌后，重新设置为可交互状态
      this.players[0].setInteractive(true);
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

      // 摸牌后，重新设置为可交互状态（可以继续出牌）
      player.setInteractive(true);
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
      // 检查发牌数据是否有重复牌（每张牌最多4张）
      const tileCount = {};
      tiles.forEach(tile => {
        tileCount[tile] = (tileCount[tile] || 0) + 1;
        if (tileCount[tile] > 4) {
          console.error(`❌ 错误：${tile} 在手牌中出现 ${tileCount[tile]} 次！这是服务器端的BUG`);
        }
      });

      player.setTiles(tiles, this.tileAssets);

      // ✅ 從牌池中移除已發的牌
      tiles.forEach(tile => {
        const removed = this.removeTileFromPool(tile);
        if (!removed) {
          console.warn(`⚠️ 無法從牌池移除 ${tile}（可能已經被移除或不存在）`);
          console.warn(`⚠️ 这可能是服务器端发送了重复的牌`);
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

  async handleDiscard(playerId, tile) {
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

    // 創建棄牌容器（包含牌底和牌面）
    const discardContainer = new Container();

    // 載入並創建牌底 sprite
    let baseTexture;
    try {
      baseTexture = await Assets.load('/assets/tiles/carddown/pbaseBig.png');
    } catch (error) {
      console.warn('無法載入棄牌牌底圖片', error);
    }

    if (baseTexture) {
      const baseSprite = new Sprite(baseTexture);
      baseSprite.anchor.set(0.5); // 設置錨點在中心
      discardContainer.addChild(baseSprite);
    }

    // 创建牌面 sprite
    const texture = this.tileAssets[tile] || this.tileAssets['back'];
    const tileSprite = new Sprite(texture);
    tileSprite.anchor.set(0.5); // 設置錨點在中心
    tileSprite.y = -25; // 將牌面往上移一點點
    discardContainer.addChild(tileSprite);

    // 设置弃牌大小（參考圖片樣式，縮小一點）
    const scale = 0.6;
    discardContainer.scale.set(scale);

    // 计算弃牌位置（在中央区域，根据玩家位置排列）
    const centerX = this.app.screen.width / 2;
    const centerY = this.app.screen.height / 2;
    // 使用实际的牌底尺寸来计算间距（牌底比牌面稍大）
    // 弃牌包含牌底和牌面，实际占用空间更大
    const tileWidth = 95 * scale;  // 再次增加宽度避免重叠（从80改为95）
    const tileHeight = 115 * scale; // 再次增加高度避免重叠（从100改为115）
    const spacing = 10; // 进一步增加间距（从8改为10）

    // 计算该玩家已经打出的牌数
    const playerDiscards = this.discardedTiles.filter(d => d.playerPosition === playerPosition);
    const discardIndex = playerDiscards.length;

    // 根据玩家位置计算弃牌位置（按照红框标注的区域）
    let x, y;
    // 上下方向每行10张，左右方向每列8张
    const maxTilesPerRow = (playerPosition === 1 || playerPosition === 3) ? 8 : 10;
    const row = Math.floor(discardIndex / maxTilesPerRow);
    const col = discardIndex % maxTilesPerRow;

    // 所有弃牌都保持正向（不旋轉）
    switch (playerPosition) {
      case 0: // 底部玩家 - 弃牌放在底部中央区域
        x = centerX - (maxTilesPerRow * (tileWidth + spacing)) / 2 + col * (tileWidth + spacing) + tileWidth / 2;
        y = centerY + 200 + row * (tileHeight + spacing);
        break;

      case 1: // 右侧玩家 - 弃牌放在右侧区域，垂直排列（每列8张）
        x = centerX + 400 + row * (tileWidth + spacing);
        y = centerY - (maxTilesPerRow * (tileHeight + spacing)) / 2 + col * (tileHeight + spacing) + tileHeight / 2;
        break;

      case 2: // 顶部玩家 - 弃牌放在顶部中央区域
        x = centerX - (maxTilesPerRow * (tileWidth + spacing)) / 2 + col * (tileWidth + spacing) + tileWidth / 2;
        y = centerY - 200 - row * (tileHeight + spacing);
        break;

      case 3: // 左侧玩家 - 弃牌放在左侧区域，垂直排列（每列8张）
        x = centerX - 400 - row * (tileWidth + spacing);
        y = centerY - (maxTilesPerRow * (tileHeight + spacing)) / 2 + col * (tileHeight + spacing) + tileHeight / 2;
        break;
    }

    discardContainer.x = x;
    discardContainer.y = y;

    // 记录弃牌信息
    this.discardedTiles.push({
      sprite: discardContainer,
      playerPosition: playerPosition,
      tile: tile
    });

    // 添加到弃牌区域
    this.discardContainer.addChild(discardContainer);

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

    // 記錄最後打出的牌
    this.lastDiscardedTile = tile;

    // 檢查是否可以執行動作（碰、吃、槓、胡）
    // 只有當不是自己打的牌時才檢查
    if (playerPosition !== this.myPosition) {
      this.checkPossibleActions(tile, playerPosition);
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

    // 根據螢幕大小計算牌山位置（緊貼牌桌邊緣）
    // 牌桌是 92% 大小，所以边缘在 46% 的位置，牌山应该在稍微内侧
    const wallDistanceVertical = Math.min(centerY * 0.85, 460);   // 上下方向（更靠近牌桌）
    const wallDistanceHorizontal = Math.min(centerX * 0.86, 830); // 左右方向（更靠近牌桌）
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

    // 設置位置（左上角，緊挨著房間號框）
    this.wallTextContainer.x = 240;
    this.wallTextContainer.y = 30;

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

    // 台湾麻将规则：海底剩余8张时流局
    if (this.remainingTiles <= 8 && this.remainingTiles > 0) {
      console.log('🚫 海底剩余8张，流局！');
      this.handleDraw(); // 触发流局
    }
  }

  /**
   * 处理流局（荒牌）
   */
  handleDraw() {
    // 禁用所有玩家交互
    this.players.forEach(player => {
      player.setInteractive(false);
    });

    // 显示流局消息
    const drawText = new Text({
      text: '流局\n海底剩余不足8张',
      style: {
        fontSize: 48,
        fill: 0xFFFFFF,
        fontWeight: 'bold',
        align: 'center',
        stroke: 0x000000,
        strokeThickness: 4
      }
    });

    drawText.anchor.set(0.5);
    drawText.x = this.app.screen.width / 2;
    drawText.y = this.app.screen.height / 2;

    this.container.addChild(drawText);

    console.log('游戏流局，无人胜出');
  }

  /**
   * 檢查可執行的動作（碰、吃、槓、胡）
   */
  checkPossibleActions(tile, discardPlayerPosition) {
    const myPlayer = this.players[this.myPosition];
    if (!myPlayer || !myPlayer.tiles) {
      return;
    }

    const myHand = myPlayer.tiles.map(t => t.type);
    const actions = [];
    const highlightGroups = []; // 存儲所有可以高亮的牌組

    // 檢查是否可以胡牌（優先級最高）
    if (this.canHu(myHand, tile)) {
      actions.push('hu');
      // TODO: 添加胡牌的牌組高亮
    }

    // 檢查是否可以槓
    if (this.canKong(myHand, tile)) {
      actions.push('kong');
      // 槓牌：需要3張相同的牌
      const kongGroup = myHand.filter(t => t === tile);
      if (kongGroup.length >= 3) {
        highlightGroups.push(kongGroup.slice(0, 3));
      }
    }

    // 檢查是否可以碰
    if (this.canPong(myHand, tile)) {
      actions.push('pong');
      // 碰牌：需要2張相同的牌
      const pongGroup = myHand.filter(t => t === tile);
      if (pongGroup.length >= 2) {
        highlightGroups.push(pongGroup.slice(0, 2));
      }
    }

    // 檢查是否可以吃（只能吃上家的牌）
    const previousPlayer = (this.myPosition + 3) % 4; // 上家位置
    if (discardPlayerPosition === previousPlayer && this.canChow(myHand, tile)) {
      actions.push('chow');
      // 吃牌：可能有多種組合
      const chowCombinations = this.getChowCombinations(myHand, tile);
      chowCombinations.forEach(combo => {
        // 只高亮手牌中的牌（不包含要吃的牌）
        const handTiles = combo.filter(t => t !== tile);
        if (handTiles.length > 0) {
          highlightGroups.push(handTiles);
        }
      });
    }

    // 如果有可執行的動作，顯示按鈕並高亮牌組
    if (actions.length > 0) {
      actions.push('cancel'); // 總是添加取消按鈕
      this.pendingActions = actions;
      this.actionButtons.show(actions);

      // 高亮顯示可組合的牌
      myPlayer.highlightTileGroups(highlightGroups);

      console.log(`可執行的動作: ${actions.join(', ')}`);
      console.log(`高亮牌組數量: ${highlightGroups.length}`);
    }
  }

  /**
   * 檢查是否可以碰
   */
  canPong(hand, tile) {
    let count = 0;
    for (const t of hand) {
      if (t === tile) {
        count++;
      }
    }
    return count >= 2;
  }

  /**
   * 檢查是否可以槓
   */
  canKong(hand, tile) {
    let count = 0;
    for (const t of hand) {
      if (t === tile) {
        count++;
      }
    }
    return count >= 3;
  }

  /**
   * 檢查是否可以吃
   */
  canChow(hand, tile) {
    // 字牌不能吃
    const honors = ['dong', 'nan', 'xi', 'bei', 'zhong', 'fa', 'bai'];
    if (honors.includes(tile)) {
      return false;
    }

    // 解析牌的類型和數字
    const parseTile = (t) => {
      const match = t.match(/^(wan|tong|tiao)-(\d)$/);
      if (!match) return null;
      return { suit: match[1], num: parseInt(match[2]) };
    };

    const parsed = parseTile(tile);
    if (!parsed) return false;

    const { suit, num } = parsed;

    // 檢查三種可能的順子組合
    // 1. tile + tile+1 + tile+2
    if (num <= 7) {
      const tile2 = `${suit}-${num + 1}`;
      const tile3 = `${suit}-${num + 2}`;
      if (hand.includes(tile2) && hand.includes(tile3)) {
        return true;
      }
    }

    // 2. tile-1 + tile + tile+1
    if (num >= 2 && num <= 8) {
      const tile1 = `${suit}-${num - 1}`;
      const tile3 = `${suit}-${num + 1}`;
      if (hand.includes(tile1) && hand.includes(tile3)) {
        return true;
      }
    }

    // 3. tile-2 + tile-1 + tile
    if (num >= 3) {
      const tile1 = `${suit}-${num - 2}`;
      const tile2 = `${suit}-${num - 1}`;
      if (hand.includes(tile1) && hand.includes(tile2)) {
        return true;
      }
    }

    return false;
  }

  /**
   * 檢查是否可以胡牌（簡化版）
   */
  canHu(hand, tile) {
    // TODO: 實作完整的胡牌判斷邏輯
    // 這裡先返回 false，可以後續實作
    return false;
  }

  /**
   * 處理碰牌動作
   */
  handlePongAction() {
    console.log('執行碰牌');
    if (this.ws && this.lastDiscardedTile) {
      this.ws.sendAction('pong', { tile: this.lastDiscardedTile });
    }

    // 清除牌組高亮
    const myPlayer = this.players[this.myPosition];
    if (myPlayer) {
      myPlayer.clearHighlight();
    }
  }

  /**
   * 處理吃牌動作
   */
  handleChowAction() {
    console.log('執行吃牌');

    // 需要讓玩家選擇吃牌組合
    // 這裡先使用第一個可能的組合
    const myHand = this.players[this.myPosition].tiles.map(t => t.type);
    const combinations = this.getChowCombinations(myHand, this.lastDiscardedTile);

    if (combinations.length > 0 && this.ws) {
      // 如果有多個組合，這裡簡化處理，使用第一個
      // TODO: 可以添加UI讓玩家選擇
      this.ws.sendAction('chow', {
        tile: this.lastDiscardedTile,
        chowTiles: combinations[0]
      });
    }

    // 清除牌組高亮
    const myPlayer = this.players[this.myPosition];
    if (myPlayer) {
      myPlayer.clearHighlight();
    }
  }

  /**
   * 獲取所有可能的吃牌組合
   */
  getChowCombinations(hand, tile) {
    const combinations = [];

    const parseTile = (t) => {
      const match = t.match(/^(wan|tong|tiao)-(\d)$/);
      if (!match) return null;
      return { suit: match[1], num: parseInt(match[2]) };
    };

    const parsed = parseTile(tile);
    if (!parsed) return combinations;

    const { suit, num } = parsed;

    // 檢查三種可能的順子組合
    if (num <= 7) {
      const tile2 = `${suit}-${num + 1}`;
      const tile3 = `${suit}-${num + 2}`;
      if (hand.includes(tile2) && hand.includes(tile3)) {
        combinations.push([tile, tile2, tile3]);
      }
    }

    if (num >= 2 && num <= 8) {
      const tile1 = `${suit}-${num - 1}`;
      const tile3 = `${suit}-${num + 1}`;
      if (hand.includes(tile1) && hand.includes(tile3)) {
        combinations.push([tile1, tile, tile3]);
      }
    }

    if (num >= 3) {
      const tile1 = `${suit}-${num - 2}`;
      const tile2 = `${suit}-${num - 1}`;
      if (hand.includes(tile1) && hand.includes(tile2)) {
        combinations.push([tile1, tile2, tile]);
      }
    }

    return combinations;
  }

  /**
   * 處理槓牌動作
   */
  handleKongAction() {
    console.log('執行槓牌');
    if (this.ws && this.lastDiscardedTile) {
      this.ws.sendAction('kong', {
        tile: this.lastDiscardedTile,
        isConcealed: false // 明槓
      });
    }

    // 清除牌組高亮
    const myPlayer = this.players[this.myPosition];
    if (myPlayer) {
      myPlayer.clearHighlight();
    }
  }

  /**
   * 處理胡牌動作
   */
  handleHuAction() {
    console.log('執行胡牌');
    if (this.ws && this.lastDiscardedTile) {
      this.ws.sendAction('hu', {
        tile: this.lastDiscardedTile,
        isSelfDrawn: false // 放炮
      });
    }

    // 清除牌組高亮
    const myPlayer = this.players[this.myPosition];
    if (myPlayer) {
      myPlayer.clearHighlight();
    }
  }

  /**
   * 處理取消動作
   */
  handleCancelAction() {
    console.log('取消動作');
    this.actionButtons.hide();
    this.pendingActions = [];

    // 清除牌組高亮
    const myPlayer = this.players[this.myPosition];
    if (myPlayer) {
      myPlayer.clearHighlight();
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

    // 調整動作按鈕位置
    if (this.actionButtons) {
      this.actionButtons.resize(width, height);
    }
  }
}
