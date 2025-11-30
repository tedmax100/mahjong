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
    this.possibleTingDiscards = {};
    this.isDeclaringTing = false;
    this.selfKongOptions = [];

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
    console.log(`${'='.repeat(60)}\n`);
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

    if (this.isDeclaringTing) {
        if (this.possibleTingDiscards[tileType]) {
            console.log(`聽牌，打出: ${tileType}`);
            this.ws.sendAction('ting', { tile: tileType });
            this.isDeclaringTing = false;
            const myPlayer = this.players[this.myPosition];
            if (myPlayer) {
              myPlayer.clearHighlight();
            }
        } else {
            this.showAnnouncement('這張牌不能打出以聽牌！', 2000);
            return;
        }
    } else {
        // 通过WebSocket发送出牌消息
        if (this.ws) {
            this.ws.sendAction('discard', { tile: tileType });
        }
    }
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

    const {playerId, action, tile, currentTurn, remainingTiles} = data;

    // 更新剩余牌数（如果服务器发送了这个信息）
    if (remainingTiles !== undefined) {
      this.updateRemainingTiles(remainingTiles);
    }

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
      case 'ting':
        this.handleTing(playerId, data);
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
    const scale = 0.45; // 縮小牌底和牌面至75% (0.6 * 0.75 = 0.45)
    discardContainer.scale.set(scale);

    // 计算弃牌位置（在中央区域，根据玩家位置排列）
    const centerX = this.app.screen.width / 2;
    const centerY = this.app.screen.height / 2;
    // 使用实际的牌底尺寸来计算间距（牌底比牌面稍大）
    // 弃牌包含牌底和牌面，实际占用空间更大
    const tileWidth = 53.4375 * scale;  // 縮小至56.25%（原95）
    const tileHeight = 64.6875 * scale; // 縮小至56.25%（原115）
    const spacing = 18; // 增加間距避免重疊

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
      this.possibleTingDiscards = {};
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
    this.possibleTingDiscards = {};
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
    this.possibleTingDiscards = {};
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
    this.possibleTingDiscards = {};
    const myPlayer = this.players[this.myPosition];
    if (myPlayer) {
      myPlayer.clearHighlight();
    }
  }

  handleHu(playerId, tile) {
    // This is now handled by handleGameWin
  }

  /**
   * 處理聽牌廣播
   */
  handleTing(playerId, data) {
    console.log(`玩家 ${playerId} 宣告聽牌`, data);

    // 找到該玩家
    const player = this.players.find(p => p.userId === playerId);
    if (!player) {
      console.error('未找到玩家:', playerId);
      return;
    }

    // 標記玩家已聽牌
    player.isTing = true;
    if (data.winningTiles) {
      player.winningTiles = data.winningTiles;
    }

    // 顯示聽牌UI
    player.showTingStatus();

    // 如果是自己，顯示提示
    if (player === this.players[this.myPosition]) {
      this.showAnnouncement('聽牌！', 2000);
    } else {
      // 其他玩家聽牌，也顯示提示
      this.showAnnouncement(`${player.name} 聽牌！`, 2000);
    }
  }

  showAnnouncement(text, duration = 3000) {
    this.announcementText.text = text;
    this.announcementText.visible = true;
    this.announcementText.eventMode = 'none'; // Prevent text from blocking clicks

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
    const tileWidth = 33.75;  // 縮小至56.25%（原60）
    const tileHeight = 45;    // 縮小至56.25%（原80）

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
          tile.x = (i - tilesPerSide / 2) * (tileWidth * 0.5 + 5); // 增加間距避免重疊
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

    // The server will handle the draw condition and send a message
  }

  handleTingResult(data) {
    this.possibleTingDiscards = data;
    const actions = [];

    if (Object.keys(this.possibleTingDiscards).length > 0) {
        actions.push('ready');
        console.log('🎯 聽牌！可打出的牌:', Object.keys(this.possibleTingDiscards));
    }

    if (actions.length > 0) {
      this.actionButtons.show([...actions, 'cancel']);
      this.pendingActions = actions;
    }
  }

  /**
   * 處理聽牌動作
   */
  handleReadyAction() {
    console.log('宣告聽牌');

    const myPlayer = this.players[this.myPosition];
    if (!myPlayer) return;

    this.isDeclaringTing = true;

    const validDiscards = Object.keys(this.possibleTingDiscards);
    myPlayer.highlightTiles(validDiscards);
    myPlayer.setInteractive(true); // 确保玩家在选择听牌时可交互
    
    this.showAnnouncement('請選擇一張牌打出以聽牌', 3000);
    this.actionButtons.hide();
  }

  /**
   * 處理碰牌動作
   */
  handlePongAction() {
    console.log('執行碰牌');
    if (this.lastDiscardedTile) {
      this.sendAction('pong', this.lastDiscardedTile);
    }
  }

  /**
   * 處理吃牌動作
   */
  /**
   * 處理吃牌動作
   */
  handleChowAction() {
    console.log('執行吃牌');
    if (this.lastDiscardedTile && this.pendingActions.chow) {
      if (this.pendingActions.chow.length > 1) {
        // 如果有多個吃牌組合，提示玩家選擇
        this.promptChowSelection(this.pendingActions.chow);
      } else if (this.pendingActions.chow.length === 1) {
        // 只有一個組合，直接執行
        const combination = this.pendingActions.chow[0];
        this.sendAction('chow', this.lastDiscardedTile, combination);
      }
    }
  }

  /**
   * 提示玩家選擇吃牌組合
   * @param {Array<Array<string>>} combinations - 可用的吃牌組合
   */
  promptChowSelection(combinations) {
    this.actionButtons.hide(); // 隱藏主操作按鈕

    const selectionContainer = new Container();
    selectionContainer.position.set(this.app.screen.width / 2, this.app.screen.height / 2);
    selectionContainer.zIndex = 2000; // 確保在最上層
    this.container.addChild(selectionContainer);

    const bg = new Graphics();
    bg.roundRect(-100, -20, 200, combinations.length * 40 + 40, 10);
    bg.fill({ color: 0x000000, alpha: 0.8 });
    selectionContainer.addChild(bg);

    combinations.forEach((combo, index) => {
        const comboText = new Text({
          text: combo.join(' '),
          style: {
            fontSize: 24,
            fill: 0xFFFFFF,
            align: 'center',
          }
        });
        comboText.anchor.set(0.5);
        comboText.y = (index * 40);
        comboText.eventMode = 'static';
        comboText.cursor = 'pointer';

        comboText.on('pointerdown', () => {
            this.sendAction('chow', this.lastDiscardedTile, combo);
            this.container.removeChild(selectionContainer);
        });
        selectionContainer.addChild(comboText);
    });
  }

  /**
   * 處理槓牌動作
   */
  handleKongAction() {
    console.log('執行槓牌');
    
    let tileToKong = null;
    let isConcealed = false;

    // Case 1: Self-drawn kong (promoted or concealed)
    if (this.selfKongOptions && this.selfKongOptions.length > 0) {
        // UI should let user choose if multiple options. For now, take the first.
        tileToKong = this.selfKongOptions[0];
        
        const myPlayer = this.players[this.myPosition];
        const hand = myPlayer.tiles.map(t => t.type);
        const handCounts = {};
        for (const tile of hand) {
            handCounts[tile] = (handCounts[tile] || 0) + 1;
        }

        if (handCounts[tileToKong] === 4) {
            isConcealed = true;
        }
    } 
    // Case 2: Kong on a discard
    else if (this.pendingActions && this.pendingActions.kong) {
        tileToKong = this.lastDiscardedTile;
        isConcealed = false; // Cannot be concealed if it's from a discard
    }

    if (tileToKong) {
      console.log(`準備發送槓牌動作: ${tileToKong}, 暗槓: ${isConcealed}`);
      this.ws.sendAction('kong', { tile: tileToKong, concealed: isConcealed });
      
      // 清理所有待处理动作状态
      this.selfKongOptions = [];
      this.pendingActions = {};
      this.actionButtons.hide();
      if (this.actionTimeout) {
          clearTimeout(this.actionTimeout);
          this.actionTimeout = null;
      }
    } else {
        console.error("無法確定要槓哪張牌");
    }
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
      // Find the winning tile from hand (server will verify)
      // This is a simplified client-side check
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
   * 處理取消動作（過）
   */
  handleCancelAction() {
    console.log('取消動作（過）');

    // 发送过的动作到服务器
    if (this.lastDiscardedTile) {
      this.sendAction('pass', this.lastDiscardedTile);
    } else {
      // 如果没有 lastDiscardedTile，只是隐藏按钮
      this.actionButtons.hide();
      this.pendingActions = [];
    }

    this.possibleTingDiscards = {};
    this.isDeclaringTing = false;

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
  
  /**
   * 处理游戏胜利
   */
  handleGameWin(data) {
    console.log('游戏胜利', data);
    const { winnerName, winResult, countdown } = data;
    const { HandTypes, TotalTai, BaseScore } = winResult;

    // 构建牌型描述
    const handTypesStr = HandTypes.map(ht => `${ht.Name} (${ht.Tai}台)`).join(' ');
    const title = `恭喜 ${winnerName} 胡牌！`;
    const details = `${handTypesStr}\n总计: ${TotalTai}台, 得分: ${BaseScore}`;

    this.showEndRoundScreen(title, details, countdown);
  }

  /**
   * 处理流局（荒牌）
   */
  handleGameDraw(data) {
    console.log('游戏流局，无人胜出', data);
    const { countdown } = data;
    this.showEndRoundScreen('流局', '无人胜出', countdown);
  }

  /**
   * 检查可执行的对手动作（吃/碰/槓/胡）
   * 注：现在由服务器端检测并通过 handlePossibleActions 通知
   */
  checkPossibleActions(tile, playerPosition) {
    // 服务器端已经检测并广播，这里不需要客户端检查
    console.log('checkPossibleActions 调用（由服务器端处理）');
  }

  /**
   * 检查自己可以执行的动作（暗槓/听牌/自摸）
   */
  checkSelfActions() {
    // 1. 发送听牌检查请求
    if (this.ws) {
      console.log('向伺服器發送聽牌檢查請求');
      this.ws.sendAction('check_ting');
    }

    // 2. 本地检查暗槓/加槓
    const myPlayer = this.players[this.myPosition];
    if (!myPlayer) return;

    this.selfKongOptions = [];
    const hand = myPlayer.tiles.map(t => t.type);

    // 检查加槓
    const pongMelds = myPlayer.melds.filter(m => (m.type === 'pong' || m.Type === 'pong'));
    for (const tileInHand of new Set(hand)) {
        for (const meld of pongMelds) {
            const meldTiles = meld.tiles || meld.Tiles;
            if (meldTiles[0] === tileInHand) {
                if (!this.selfKongOptions.includes(tileInHand)) {
                    this.selfKongOptions.push(tileInHand);
                }
            }
        }
    }

    // 检查暗槓
    const handCounts = {};
    for (const tile of hand) {
        handCounts[tile] = (handCounts[tile] || 0) + 1;
    }
    for (const tile in handCounts) {
        if (handCounts[tile] === 4) {
            if (!this.selfKongOptions.includes(tile)) {
                this.selfKongOptions.push(tile);
            }
        }
    }
    
    // 如果有可用的槓牌選項，显示按钮
    // 注意：这可能会覆盖由听牌检查显示的按钮，这是一个已知的UI限制
    if (this.selfKongOptions.length > 0) {
        console.log("发现可执行的个人槓牌:", this.selfKongOptions);
        this.actionButtons.show(['kong', 'cancel']);
    }

    // TODO: 在此處加入檢查自摸等其他動作
  }

  /**
   * 处理可执行动作通知
   */
  handlePossibleActions(data) {
    const { playerId, tile, actions, timeout } = data;
    console.log('收到可执行动作通知:', { playerId, tile, actions, timeout });

    // 检查是否是当前玩家（使用 userId 而不是 id）
    const myPlayer = this.players[this.myPosition];
    console.log('DEBUG: myPlayer.userId =', myPlayer?.userId, 'playerId =', playerId, '是否相等?', myPlayer?.userId === playerId);

    if (!myPlayer || myPlayer.userId !== playerId) {
      console.log('不是当前玩家的动作，忽略 - myPlayer.userId:', myPlayer?.userId, 'expected:', playerId);
      return;
    }

    // 保存最后打出的牌和可执行动作
    this.lastDiscardedTile = tile;
    this.pendingActions = actions;

    // 确定要显示的按钮
    const buttonsToShow = [];
    if (actions.pong) buttonsToShow.push('pong');
    if (actions.chow) buttonsToShow.push('chow');
    if (actions.kong) buttonsToShow.push('kong');
    if (actions.hu) buttonsToShow.push('hu');
    buttonsToShow.push('cancel'); // 总是显示"过"按钮

    console.log('显示动作按钮:', buttonsToShow);

    // 显示按钮
    if (this.actionButtons) {
      this.actionButtons.show(buttonsToShow);
    }

    // 清除之前的超时
    if (this.actionTimeout) {
      clearTimeout(this.actionTimeout);
    }

    // 设置超时自动过
    this.actionTimeout = setTimeout(() => {
      console.log('动作选择超时，自动过');
      this.sendAction('pass', tile);
      if (this.actionButtons) {
        this.actionButtons.hide();
      }
    }, timeout * 1000);
  }

  /**
   * 发送动作选择到服务器
   */
  sendAction(action, tile, combination = null) {
    const data = {
      tile: tile
    };

    if (combination) {
      if (action === 'chow') {
        data.chowTiles = combination;
      } else {
        data.combination = combination;
      }
    }

    const message = {
      type: 'action',
      action: action,
      data: data
    };

    console.log('发送动作:', message);
    this.ws.send(message);

    // 清除超时
    if (this.actionTimeout) {
      clearTimeout(this.actionTimeout);
      this.actionTimeout = null;
    }

    // 隐藏按钮
    if (this.actionButtons) {
      this.actionButtons.hide();
    }
  }

  /**
   * 显示回合结束画面（胡牌或流局）
   */
  async showEndRoundScreen(title, details, countdown) {
    // 禁用所有玩家交互
    this.players.forEach(player => {
      player.setInteractive(false);
    });
    this.actionButtons.hide();

    const endScreenContainer = new Container();
    endScreenContainer.zIndex = 10000;

    // 半透明背景
    const bg = new Graphics();
    bg.rect(0, 0, this.app.screen.width, this.app.screen.height);
    bg.fill({ color: 0x000000, alpha: 0.7 });
    endScreenContainer.addChild(bg);

    // 标题（例如 "恭喜..." 或 "流局"）
    const titleText = new Text({ text: title, style: {
      fontSize: 60,
      fill: 0xFFD700,
      fontWeight: 'bold',
      align: 'center',
    }});
    titleText.anchor.set(0.5);
    titleText.x = this.app.screen.width / 2;
    titleText.y = this.app.screen.height / 2 - 100;
    endScreenContainer.addChild(titleText);

    // 详情（例如牌型和分数）
    const detailsText = new Text({ text: details, style: {
      fontSize: 32,
      fill: 0xFFFFFF,
      align: 'center',
      wordWrap: true,
      wordWrapWidth: this.app.screen.width * 0.8,
    }});
    detailsText.anchor.set(0.5);
    detailsText.x = this.app.screen.width / 2;
    detailsText.y = this.app.screen.height / 2;
    endScreenContainer.addChild(detailsText);

    // 倒计时文本
    const countdownText = new Text({
      text: `${countdown}秒后开始新局`,
      style: {
        fontSize: 28,
        fill: 0xCCCCCC,
      }
    });
    countdownText.anchor.set(0.5);
    countdownText.x = this.app.screen.width / 2;
    countdownText.y = this.app.screen.height / 2 + 100;
    endScreenContainer.addChild(countdownText);

    this.container.addChild(endScreenContainer);

    // 倒计时逻辑
    let remaining = countdown;
    const countdownInterval = setInterval(() => {
      remaining--;
      if (remaining > 0) {
        countdownText.text = `${remaining}秒后开始新局`;
      } else {
        clearInterval(countdownInterval);
        // 服务器会自动开始新的一局，客户端等待 game_start 消息
        // 清理界面
        this.container.removeChild(endScreenContainer);
        this.resetForNewRound();
      }
    }, 1000);
  }

  /**
   * 为新的一局重置UI
   */
  resetForNewRound() {
    // 清空弃牌
    this.discardContainer.removeChildren();
    this.discardedTiles = [];

    // 重置所有玩家
    this.players.forEach(p => p.reset());

    console.log("UI已重置，准备新的一局");
  }
}