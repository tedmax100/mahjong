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

    // 游戏公告
    this.announcementText = null;
  }

  /**
   * 記錄玩家手牌狀態（用於 debug）
   */
  logPlayerHand(playerPosition, action = '') {
    const player = this.players[playerPosition];
    if (!player) return;

    const handTiles = player.tiles.map(t => t.type).sort();
    const melds = player.melds.map(m => {
      const type = m.Type || m.type;
      const tiles = m.Tiles || m.tiles;
      return `[${type}: ${tiles.join(',')}]`;
    });

    const actionPrefix = action ? `[${action}] ` : '';
    console.log(`📋 ${actionPrefix}玩家 ${player.name || playerPosition} (位置${playerPosition})`);
    console.log(`   手牌 (${handTiles.length}張): ${handTiles.join(' ')}`);
    if (melds.length > 0) {
      console.log(`   吃碰槓: ${melds.join(' ')}`);
    }
    console.log(`   總牌數: ${handTiles.length + melds.length * 3}張`);
  }

  /**
   * 記錄所有玩家的手牌狀態
   */
  logAllPlayersHands(action = '') {
    console.log(`\n${'='.repeat(60)}`);
    console.log(`📊 ${action || '當前遊戲狀態'}`);
    console.log(`${'='.repeat(60)}`);
    this.players.forEach((player, index) => {
      this.logPlayerHand(index, '');
    });
    console.log(`${'='.repeat(60)}\n`);
  }

  async init() {
    // 啟用容器的 zIndex 排序，確保層級正確
    this.container.sortableChildren = true;

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
    this.actionButtons.on('ready', () => this.handleReadyAction());
    this.actionButtons.on('hu', () => this.handleHuAction());
    this.actionButtons.on('cancel', () => this.handleCancelAction());

    // 创建游戏公告区域
    this.announcementText = new Text({ text: '', style: {
        fontSize: 80,
        fill: 0xFFD700,
        fontWeight: 'bold',
        stroke: 0x000000,
        strokeThickness: 6,
        align: 'center',
    }});
    this.announcementText.anchor.set(0.5);
    this.announcementText.x = this.app.screen.width / 2;
    this.announcementText.y = 120;
    this.announcementText.visible = false;
    this.announcementText.zIndex = 2000;
    this.container.addChild(this.announcementText);

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

    // 出牌後延遲檢查聽牌狀態（給伺服器時間更新狀態）
    setTimeout(() => {
      const myPlayer = this.players[this.myPosition];
      if (myPlayer && myPlayer.tiles) {
        const myHand = myPlayer.tiles.map(t => t.type);
        const meldCount = myPlayer.melds ? myPlayer.melds.length : 0;

        console.log(`🎴 出牌後檢查聽牌 - 手牌數: ${myHand.length}, 面子數: ${meldCount}`, myHand);

        // 檢查是否聽牌
        const readyTiles = this.checkReadyHand(myHand);
        if (readyTiles.length > 0) {
          console.log(`🎯 聽牌！聽: ${readyTiles.join(', ')}（共${readyTiles.length}張）`);
          // 顯示聽牌按鈕
          this.actionButtons.show(['ready', 'cancel']);
          this.pendingActions = ['ready'];
        }
      }
    }, 300);
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

      // 📋 記錄初始手牌
      setTimeout(() => {
        this.logPlayerHand(position, '初始手牌');

        // 如果所有玩家都發牌完成，記錄所有玩家的手牌
        const allDealt = this.players.every(p => p.tiles.length > 0);
        if (allDealt) {
          this.logAllPlayersHands('遊戲開始 - 所有玩家初始手牌');
        }
      }, 100);
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
        this.handlePlayerDraw(playerId, tile);
        break;
      case 'chow':
        this.handleChow(playerId, data);
        break;
      case 'pong':
        this.handlePong(playerId, tile);
        break;
      case 'kong':
        this.handleKong(playerId, data);
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

      // 如果輪到自己，檢查是否能聽牌或自摸
      if (isMyTurn) {
        setTimeout(() => {
          this.checkSelfActions();
        }, 200);
      }
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
      baseTexture = await Assets.load('/assets/tiles/carddown/basefdown.png');
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
    tileSprite.y = 5; // 調整牌面位置，讓它貼齊牌底下緣

    // 針對筒子微調
    if (tile.startsWith('tong-')) {
      tileSprite.y += 8; // 往下移8個像素
    }

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

      case 2: // 顶部玩家 - 弃牌放在顶部中央区域，從右邊開始，第二排往下
        x = centerX + (maxTilesPerRow * (tileWidth + spacing)) / 2 - col * (tileWidth + spacing) - tileWidth / 2;
        y = centerY - 280 + row * (tileHeight + spacing); // 改為 + row，讓第二排往下移
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

      // 📋 記錄打牌後的手牌狀態
      setTimeout(() => {
        this.logPlayerHand(playerPosition, `打牌: ${tile}`);
      }, 100);
    }

    // ❌ 移除錯誤的剩餘牌數計算
    // 打牌不會減少牌山的牌數，只有摸牌才會減少
    // 剩餘牌數的更新已經在 handleDraw() 中處理

    // 記錄最後打出的牌
    this.lastDiscardedTile = tile;

    // 檢查是否可以執行動作（碰、吃、槓、胡）
    // 只有當不是自己打的牌時才檢查
    if (playerPosition !== this.myPosition) {
      this.checkPossibleActions(tile, playerPosition);
    } else {
      // 如果是自己打的牌，清除動作按鈕
      this.actionButtons.hide();
      this.pendingActions = [];
      const myPlayer = this.players[this.myPosition];
      if (myPlayer) {
        myPlayer.clearHighlight();
      }
    }
  }

  async handlePlayerDraw(playerId, tile) {
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

    // 加入新牌到手牌（等待完成）
    await player.addTile(tileToAdd, this.tileAssets);

    // 更新剩餘牌數
    if (this.remainingTiles > 0) {
      this.updateRemainingTiles(this.remainingTiles - 1);
    }

    console.log(`玩家 ${playerPosition} 摸牌完成，剩餘 ${this.remainingTiles} 張`);

    // 📋 記錄摸牌後的手牌狀態
    setTimeout(() => {
      this.logPlayerHand(playerPosition, `摸牌: ${tile}`);
    }, 100);

    // 如果是自己摸牌，檢查是否可以自摸或聽牌
    if (playerPosition === this.myPosition) {
      // 等待一小段時間，確保牌已經加入並顯示
      setTimeout(() => {
        this.checkSelfActions();
      }, 100);
    }
  }

  async handleChow(playerId, data) {
    // 处理吃牌
    const chowTiles = data.chowTiles || [];
    const tile = data.tile;
    console.log(`🍜 handleChow - 玩家 ${playerId} 吃了 ${tile}，牌組: ${chowTiles.join(', ')}`);

    // 找到執行吃牌的玩家
    let playerPosition = -1;
    for (let i = 0; i < this.players.length; i++) {
      if (this.players[i].userId === playerId) {
        playerPosition = i;
        break;
      }
    }

    if (playerPosition !== -1 && chowTiles.length === 3) {
      const player = this.players[playerPosition];

      console.log(`🍜 吃牌前手牌 (${player.tiles.length}張):`, player.tiles.map(t => t.type));

      // 從手牌中移除用於吃牌的2張牌（不包括被吃的那張）
      const tilesToRemove = chowTiles.filter(t => t !== tile);
      console.log(`🍜 要移除的牌:`, tilesToRemove);

      for (const tileToRemove of tilesToRemove) {
        for (let i = player.tiles.length - 1; i >= 0; i--) {
          if (player.tiles[i].type === tileToRemove) {
            player.tiles[i].destroy();
            player.tiles.splice(i, 1);
            console.log(`🍜 已移除: ${tileToRemove}`);
            break;
          }
        }
      }

      console.log(`🍜 吃牌後手牌 (${player.tiles.length}張):`, player.tiles.map(t => t.type));

      // 重新排列剩餘手牌
      player.rearrangeTiles();

      // 添加吃牌組到顯示區域
      await player.addMeld({
        type: 'chow',
        tiles: chowTiles
      }, this.tileAssets);

      // 從棄牌堆中移除最後一張（被吃的牌）
      if (this.discardedTiles.length > 0) {
        const lastDiscard = this.discardedTiles.pop();
        this.discardContainer.removeChild(lastDiscard.sprite);
      }

      console.log(`✅ 玩家 ${playerPosition} 吃牌完成`);

      // 📋 記錄吃牌後的手牌狀態
      setTimeout(() => {
        this.logPlayerHand(playerPosition, `吃牌: ${chowTiles.join(',')}`);
      }, 100);

      // 如果是自己吃牌，檢查是否能聽牌或自摸
      if (playerPosition === this.myPosition) {
        setTimeout(() => {
          this.checkSelfActions();
        }, 300);
      }
    }

    // 清除動作按鈕
    this.actionButtons.hide();
    this.pendingActions = [];
    const myPlayer = this.players[this.myPosition];
    if (myPlayer) {
      myPlayer.clearHighlight();
    }
  }

  async handlePong(playerId, tile) {
    // 处理碰牌
    console.log(`玩家 ${playerId} 碰了 ${tile}`);

    // 找到執行碰牌的玩家
    let playerPosition = -1;
    for (let i = 0; i < this.players.length; i++) {
      if (this.players[i].userId === playerId) {
        playerPosition = i;
        break;
      }
    }

    if (playerPosition !== -1) {
      const player = this.players[playerPosition];

      // 從手牌中移除2張相同的牌
      let removed = 0;
      for (let i = player.tiles.length - 1; i >= 0 && removed < 2; i--) {
        if (player.tiles[i].type === tile) {
          player.tiles[i].destroy();
          player.tiles.splice(i, 1);
          removed++;
        }
      }

      // 重新排列剩餘手牌
      player.rearrangeTiles();

      // 添加碰牌組到顯示區域
      await player.addMeld({
        type: 'pong',
        tiles: [tile, tile, tile]
      }, this.tileAssets);

      // 從棄牌堆中移除最後一張（被碰的牌）
      if (this.discardedTiles.length > 0) {
        const lastDiscard = this.discardedTiles.pop();
        this.discardContainer.removeChild(lastDiscard.sprite);
      }

      console.log(`✅ 玩家 ${playerPosition} 碰牌完成`);

      // 📋 記錄碰牌後的手牌狀態
      setTimeout(() => {
        this.logPlayerHand(playerPosition, `碰牌: ${tile}`);
      }, 100);

      // 如果是自己碰牌，檢查是否能聽牌或自摸
      if (playerPosition === this.myPosition) {
        setTimeout(() => {
          this.checkSelfActions();
        }, 300);
      }
    }

    // 清除動作按鈕（因為這張牌已經被碰了）
    this.actionButtons.hide();
    this.pendingActions = [];
    const myPlayer = this.players[this.myPosition];
    if (myPlayer) {
      myPlayer.clearHighlight();
    }
  }

  async handleKong(playerId, data) {
    // 处理杠牌
    const meld = data.meld; // Note: Keys will be capitalized: Type, Tiles
    if (!meld) {
      console.error('无效的杠牌动作: 缺少牌组信息', data);
      return;
    }
    const tile = meld.Tiles[0];
    console.log(`玩家 ${playerId} 杠了 ${tile} (类型: ${meld.Type})`);

    // 找到執行槓牌的玩家
    let playerPosition = -1;
    for (let i = 0; i < this.players.length; i++) {
      if (this.players[i].userId === playerId) {
        playerPosition = i;
        break;
      }
    }

    if (playerPosition !== -1) {
      const player = this.players[playerPosition];

      // 只为新杠从手牌中移除牌，而不是为加杠
      if (meld.Type === 'kong_exposed') {
        // 明杠，从手牌移除3张
        let removed = 0;
        for (let i = player.tiles.length - 1; i >= 0 && removed < 3; i--) {
          if (player.tiles[i].type === tile) {
            player.tiles[i].destroy();
            player.tiles.splice(i, 1);
            removed++;
          }
        }
      } else if (meld.Type === 'kong_concealed') {
        // 暗杠，从手牌移除4张
        let removed = 0;
        for (let i = player.tiles.length - 1; i >= 0 && removed < 4; i--) {
          if (player.tiles[i].type === tile) {
            player.tiles[i].destroy();
            player.tiles.splice(i, 1);
            removed++;
          }
        }
      }
      // 对于 'kong_promoted'，不从手牌中移除牌

      // 重新排列剩餘手牌
      player.rearrangeTiles();

      // 添加槓牌組到顯示區域
      await player.addMeld(meld, this.tileAssets);

      // 如果是明杠，从弃牌堆中移除最后一张（被槓的牌）
      if (meld.Type === 'kong_exposed' && this.discardedTiles.length > 0) {
        const lastDiscard = this.discardedTiles.pop();
        this.discardContainer.removeChild(lastDiscard.sprite);
      }

      console.log(`✅ 玩家 ${playerPosition} 杠牌完成`);

      // 📋 記錄槓牌後的手牌狀態
      setTimeout(() => {
        this.logPlayerHand(playerPosition, `槓牌: ${tile} (${meld.Type})`);
      }, 100);

      // 如果是自己槓牌，檢查是否能聽牌或自摸（槓牌後會補牌，所以可能自摸）
      if (playerPosition === this.myPosition) {
        setTimeout(() => {
          this.checkSelfActions();
        }, 300);
      }
    }

    // 清除動作按鈕
    this.actionButtons.hide();
    this.pendingActions = [];
    const myPlayer = this.players[this.myPosition];
    if (myPlayer) {
      myPlayer.clearHighlight();
    }
  }

  handleHu(playerId, tile) {
    // 处理胡牌
    console.log(`玩家 ${playerId} 胡了 ${tile}`);
    const player = this.players.find(p => p.userId === playerId);
    const winnerName = player ? player.name : '一个玩家';
    this.showAnnouncement(`${winnerName} 胡!`, 5000);

    // 清除動作按鈕（因為有人胡牌了）
    this.actionButtons.hide();
    this.pendingActions = [];
    const myPlayer = this.players[this.myPosition];
    if (myPlayer) {
      myPlayer.clearHighlight();
    }
  }

  showAnnouncement(text, duration = 3000) {
    this.announcementText.text = text;
    this.announcementText.visible = true;

    if (this.announcementTimeout) {
      clearTimeout(this.announcementTimeout);
    }
    this.announcementTimeout = setTimeout(() => {
      this.hideAnnouncement();
    }, duration);
  }

  hideAnnouncement() {
    this.announcementText.visible = false;
    if (this.announcementTimeout) {
      clearTimeout(this.announcementTimeout);
      this.announcementTimeout = null;
    }
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

    // 設置位置（置中，避免被遮擋）
    this.wallTextContainer.x = this.app.screen.width / 2;
    this.wallTextContainer.y = 30;

    // 設置 zIndex 確保顯示在最上層
    this.wallTextContainer.zIndex = 1000;

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
      this.handleGameDraw(); // 触发流局
    }
  }

  /**
   * 处理流局（荒牌）
   */
  handleGameDraw() {
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
    console.log(`🎴 檢查動作 - 棄牌: ${tile}, 我的手牌:`, myHand);
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
    if (discardPlayerPosition === previousPlayer) {
      const canChowResult = this.canChow(myHand, tile);
      console.log(`🍜 吃牌檢查 - 上家: ${previousPlayer}, 棄牌者: ${discardPlayerPosition}, 可以吃: ${canChowResult}`);
      if (canChowResult) {
        actions.push('chow');
        // 吃牌：可能有多種組合
        const chowCombinations = this.getChowCombinations(myHand, tile);
        console.log(`🍜 吃牌組合:`, chowCombinations);
        chowCombinations.forEach(combo => {
          // 只高亮手牌中的牌（不包含要吃的牌）
          const handTiles = combo.filter(t => t !== tile);
          if (handTiles.length > 0) {
            highlightGroups.push(handTiles);
          }
        });
      }
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
    } else {
      // 沒有可執行的動作時，隱藏按鈕並清除高亮
      this.actionButtons.hide();
      this.pendingActions = [];
      myPlayer.clearHighlight();
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
  /**
   * 檢查是否可以胡牌
   * @param {Array<string>} hand - 手牌陣列
   * @param {string} tile - 要檢查的牌（可能是摸的或別人打的）
   * @param {number} meldCount - 已經吃碰槓的組數（默認為當前玩家的melds數量）
   * @returns {boolean}
   */
  canHu(hand, tile, meldCount = null) {
    // 獲取已有的吃碰槓組數
    if (meldCount === null) {
      const myPlayer = this.players[this.myPosition];
      meldCount = myPlayer ? myPlayer.melds.length : 0;
    }

    // 將要胡的牌加入手牌（創建副本，不修改原手牌）
    const fullHand = [...hand, tile];

    // 台灣16張麻將：總共需要 5組面子 + 1對將 = 17張（包含胡的那張）
    // 如果有吃碰槓，每組算一個面子，手牌相應減少
    const requiredMelds = 5 - meldCount;
    const requiredTiles = requiredMelds * 3 + 2; // N組面子(每組3張) + 1對將(2張)

    if (fullHand.length !== requiredTiles) {
      return false;
    }

    // 統計手牌
    const tileCount = {};
    fullHand.forEach(t => {
      tileCount[t] = (tileCount[t] || 0) + 1;
    });

    // 嘗試找出所有可能的將（對子）
    for (const [tile, count] of Object.entries(tileCount)) {
      if (count >= 2) {
        // 嘗試將這張牌作為將
        const remainingTiles = { ...tileCount };
        remainingTiles[tile] -= 2;

        // 檢查剩餘的牌是否能組成所需數量的面子
        if (this.canFormMelds(remainingTiles, requiredMelds)) {
          return true;
        }
      }
    }

    return false;
  }

  /**
   * 檢查剩餘的牌是否能組成指定數量的面子（順子或刻子）
   * @param {Object} tileCount - 牌的統計 {牌名: 數量}
   * @param {number} meldsNeeded - 需要的面子數量
   * @returns {boolean}
   */
  canFormMelds(tileCount, meldsNeeded) {
    // 深度複製，避免修改原始數據
    const tiles = { ...tileCount };

    // 移除數量為0的牌
    Object.keys(tiles).forEach(key => {
      if (tiles[key] === 0) delete tiles[key];
    });

    // 如果沒有剩餘牌且不需要更多面子，則成功
    if (Object.keys(tiles).length === 0 && meldsNeeded === 0) {
      return true;
    }

    // 如果還需要面子但沒有牌了，或不需要面子但還有牌，則失敗
    if (Object.keys(tiles).length === 0 || meldsNeeded === 0) {
      return false;
    }

    // 獲取第一張牌
    const firstTile = Object.keys(tiles).sort()[0];
    const count = tiles[firstTile];

    // 嘗試用這張牌組成刻子
    if (count >= 3) {
      const newTiles = { ...tiles };
      newTiles[firstTile] -= 3;
      if (newTiles[firstTile] === 0) delete newTiles[firstTile];

      if (this.canFormMelds(newTiles, meldsNeeded - 1)) {
        return true;
      }
    }

    // 嘗試用這張牌組成順子（只有數字牌可以）
    const match = firstTile.match(/^(wan|tong|tiao)-(\d)$/);
    if (match) {
      const suit = match[1];
      const num = parseInt(match[2]);

      // 檢查是否能組成順子 [n, n+1, n+2]
      if (num <= 7) {
        const tile2 = `${suit}-${num + 1}`;
        const tile3 = `${suit}-${num + 2}`;

        if (tiles[tile2] >= 1 && tiles[tile3] >= 1) {
          const newTiles = { ...tiles };
          newTiles[firstTile] -= 1;
          newTiles[tile2] -= 1;
          newTiles[tile3] -= 1;

          if (newTiles[firstTile] === 0) delete newTiles[firstTile];
          if (newTiles[tile2] === 0) delete newTiles[tile2];
          if (newTiles[tile3] === 0) delete newTiles[tile3];

          if (this.canFormMelds(newTiles, meldsNeeded - 1)) {
            return true;
          }
        }
      }
    }

    // 如果都不行，則無法組成
    return false;
  }

  /**
   * 檢查是否聽牌（差一張就能胡）
   * @param {Array<string>} hand - 手牌陣列
   * @returns {Array<string>} 聽的牌列表
   */
  checkReadyHand(hand) {
    const myPlayer = this.players[this.myPosition];
    const meldCount = myPlayer ? myPlayer.melds.length : 0;

    // 檢查所有可能的牌（萬筒條 1-9，東南西北中發白）
    const allPossibleTiles = [];

    // 萬筒條 1-9
    ['wan', 'tong', 'tiao'].forEach(suit => {
      for (let num = 1; num <= 9; num++) {
        allPossibleTiles.push(`${suit}-${num}`);
      }
    });

    // 字牌
    ['dong', 'nan', 'xi', 'bei', 'zhong', 'fa', 'bai'].forEach(tile => {
      allPossibleTiles.push(tile);
    });

    // 測試每張牌，看加入後是否能胡
    const readyTiles = [];
    for (const tile of allPossibleTiles) {
      if (this.canHu(hand, tile, meldCount)) {
        readyTiles.push(tile);
      }
    }

    return readyTiles;
  }

  /**
   * 檢查自己摸牌後可以執行的動作（自摸、聽牌、暗槓）
   */
  checkSelfActions() {
    const myPlayer = this.players[this.myPosition];
    if (!myPlayer || !myPlayer.tiles || this.myPosition !== this.currentTurn) {
      return;
    }

    const myHand = myPlayer.tiles.map(t => t.type);
    const actions = [];

    console.log(`🎴 檢查自己的動作 - 手牌數: ${myHand.length}`, myHand);

    // 檢查是否可以自摸（手牌中任意一張都可能是剛摸的）
    // 由於我們不知道哪張是剛摸的，檢查是否已經胡牌
    for (const tile of myHand) {
      const handWithoutTile = myHand.filter((t, i) => {
        // 只移除一張
        const firstIndex = myHand.indexOf(tile);
        return i !== firstIndex;
      });

      if (this.canHu(handWithoutTile, tile)) {
        actions.push('hu');
        console.log(`✅ 可以自摸！胡牌: ${tile}`);
        break; // 找到一個就夠了
      }
    }

    // 如果不能自摸，檢查是否聽牌
    if (!actions.includes('hu')) {
      const readyTiles = this.checkReadyHand(myHand);
      if (readyTiles.length > 0) {
        actions.push('ready');
        console.log(`🎯 聽牌！聽: ${readyTiles.join(', ')}（共${readyTiles.length}張）`);
      }
    }

    // 顯示動作按鈕
    if (actions.length > 0) {
      this.actionButtons.show([...actions, 'cancel']);
      this.pendingActions = actions;
    }
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
   * 處理聽牌動作
   */
  handleReadyAction() {
    console.log('宣告聽牌');

    const myPlayer = this.players[this.myPosition];
    if (!myPlayer) return;

    // 檢查聽哪些牌
    const myHand = myPlayer.tiles.map(t => t.type);
    const readyTiles = this.checkReadyHand(myHand);

    if (readyTiles.length > 0) {
      console.log(`🎯 聽牌宣告成功！聽: ${readyTiles.join(', ')}`);

      // 發送聽牌通知到服務器（如果需要）
      if (this.ws) {
        this.ws.sendAction('ting', {
          readyTiles: readyTiles
        });
      }

      // 顯示提示訊息
      this.showAnnouncement('聽牌！', 2000);
    }

    // 隱藏按鈕
    this.actionButtons.hide();
    this.pendingActions = [];
  }

  /**
   * 處理胡牌動作
   */
  handleHuAction() {
    console.log('執行胡牌');

    const myPlayer = this.players[this.myPosition];
    const myHand = myPlayer ? myPlayer.tiles.map(t => t.type) : [];

    // 判斷是自摸還是別人放炮
    let isSelfDrawn = false;
    let winTile = this.lastDiscardedTile;

    // 如果沒有 lastDiscardedTile，說明是自摸
    if (!this.lastDiscardedTile || this.currentTurn === this.myPosition) {
      isSelfDrawn = true;
      // 找出胡的是哪張牌（檢查手牌中的每張）
      for (const tile of myHand) {
        const handWithoutTile = myHand.filter((t, i) => {
          const firstIndex = myHand.indexOf(tile);
          return i !== firstIndex;
        });
        if (this.canHu(handWithoutTile, tile)) {
          winTile = tile;
          break;
        }
      }
    }

    if (this.ws && winTile) {
      this.ws.sendAction('hu', {
        tile: winTile,
        isSelfDrawn: isSelfDrawn
      });
    }

    // 清除牌組高亮
    if (myPlayer) {
      myPlayer.clearHighlight();
    }

    // 隱藏按鈕
    this.actionButtons.hide();
    this.pendingActions = [];
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
