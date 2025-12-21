import { Container, Text, Graphics, Sprite, Assets } from 'pixi.js';
import { Tile } from './Tile.js';
import { Table } from './Table.js';
import { Player } from './Player.js';
import { ActionButtons } from './ActionButtons.js';
import { AudioManager } from './AudioManager.js';
import { MahjongLogic } from './MahjongLogic.js';
import { AssetLoader } from './AssetLoader.js';
import { WinningHandDisplay } from './WinningHandDisplay.js';
import { DiscardManager } from './DiscardManager.js';
import { ChowSelectionUI } from './ChowSelectionUI.js';
import { DiceRollUI } from './DiceRollUI.js';

/**
 * 主遊戲類
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
    this.currentTurn = 0; // 當前輪到誰（0-3）
    this.tileAssets = {};
    this.discardManager = new DiscardManager(this.app.screen.width, this.app.screen.height);

    // 牌山相關
    this.wallContainer = new Container(); // 牌山容器
    this.wallTiles = []; // 牌山的 sprite
    this.remainingTiles = 144; // 剩餘可摸牌數（初始144張）
    this.wallText = null; // 顯示剩餘牌數的文字
    this.tilePool = []; // 可用牌池（用於摸牌）

    // 莊家相關
    this.dealerPosition = 0; // 莊家位置（0-3）
    this.dealerFirstDiscard = true; // 莊家是否還沒打過第一張牌

    // 風位相關
    this.roundWind = 'E'; // 場風: E=東, S=南, W=西, N=北
    this.mySeatWind = 'E'; // 我的門風
    this.allSeatWinds = ['E', 'S', 'W', 'N']; // 所有玩家的門風

    // 動作按鈕
    this.actionButtons = null;
    this.lastDiscardedTile = null; // 最後被打出的牌
    this.pendingActions = []; // 可執行的動作列表
    this.possibleTingDiscards = {};
    this.isDeclaringTing = false;
    this.selfKongOptions = [];
    this.chowSelectionUI = null;

    // 遊戲公告
    this.announcementText = null;
    this.winningHandContainer = null;
    this.endScreenContainer = null;

    // 音效管理器
    this.audioManager = new AudioManager();

    // 素材載入器
    this.assetLoader = new AssetLoader(this.app.renderer);

    // 擲骰 UI
    this.diceRollUI = null;
  }

  /**
   * 檢查是否可以胡牌 (Delegated to MahjongLogic)
   */
  canHu(hand, tile, meldCount = null) {
    if (meldCount === null) {
        // Fallback if meldCount not provided, check current player
        const player = this.players[this.myPosition];
        meldCount = player ? player.melds.length : 0;
    }
    return MahjongLogic.canHu(hand, tile, meldCount);
  }

  /**
   * 檢查是否聽牌 (Delegated to MahjongLogic)
   */
  checkReadyHand(hand) {
    const player = this.players[this.myPosition];
    const meldCount = player ? player.melds.length : 0;
    return MahjongLogic.checkReadyHand(hand, meldCount);
  }

  /**
   * 設定 WebSocket 實體（用於延後連接）
   */
  setWebSocket(ws) {
    this.ws = ws;
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
    console.log(`
${'='.repeat(60)}`);
    console.log(`📊 ${action || '當前遊戲狀態'}`);
    console.log(`${'='.repeat(60)}
`);
    this.players.forEach((player, index) => {
      this.logPlayerHand(index, '');
    });
    console.log(`${'='.repeat(60)}
`);
  }

  async init() {
    // 啟用容器的 zIndex 排序，確保層級正確
    this.container.sortableChildren = true;

    // 預載入牌底紋理（所有 Tile 共享）
    await Tile.preloadBaseTexture();

    // 載入素材 (Using AssetLoader)
    this.tileAssets = await this.assetLoader.load();

    // 初始化吃牌選擇 UI
    this.chowSelectionUI = new ChowSelectionUI(this.container, this.app.screen, this.tileAssets);


    // 創建牌桌
    this.table = new Table(this.app.screen.width, this.app.screen.height);
    this.container.addChild(this.table.container);

    // 新增棄牌區域容器
    this.container.addChild(this.discardManager.container);

    // 創建玩家區域
    this.createPlayers();

    // 創建動作按鈕（非同步初始化）
    this.actionButtons = new ActionButtons(this.app.screen.width, this.app.screen.height);
    this.container.addChild(this.actionButtons.container);
    // 設定最高 zIndex 確保按鈕不被遮擋
    this.actionButtons.container.zIndex = 3000;

    // 設定按鈕回呼
    this.actionButtons.on('pong', () => this.handlePongAction());
    this.actionButtons.on('chow', () => this.handleChowAction());
    this.actionButtons.on('kong', () => this.handleKongAction());
    this.actionButtons.on('ready', () => this.handleReadyAction());
    this.actionButtons.on('hu', () => this.handleHuAction());
    this.actionButtons.on('cancel', () => this.handleCancelAction());

    // 創建遊戲公告區域
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

    // Winning hand display container
    this.winningHandContainer = new Container();
    this.winningHandContainer.visible = false;
    this.winningHandContainer.zIndex = 11000;
    this.container.addChild(this.winningHandContainer);
    // Initialize with loaded tile assets
    this.winningHandDisplayHandler = new WinningHandDisplay(this.winningHandContainer, this.app.screen, this.tileAssets);

    // 初始化擲骰 UI
    this.diceRollUI = new DiceRollUI(
      this.app.screen.width,
      this.app.screen.height,
      this.audioManager
    );
    await this.diceRollUI.load();
    this.container.addChild(this.diceRollUI.container);

    // 顯示等待文字
    this.showWaitingText();

    // 播放選單背景音樂
    this.audioManager.playBGM('menu');

    console.log('✅ Game initialized successfully');
  }

  createPlayers() {
    const positions = ['bottom', 'right', 'top', 'left'];

    for (let i = 0; i < 4; i++) {
      const player = new Player(i, positions[i], this.app.screen.width, this.app.screen.height);

      // 設定底部玩家（自己）的出牌回呼
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
            const myPlayer = this.players[0]; // 視覺位置 0 是自己
            if (myPlayer) {
              myPlayer.clearHighlight();
            }
        } else {
            this.showAnnouncement('這張牌不能打出以聽牌！', 2000);
            return;
        }
    } else {
        // 透過WebSocket發送出牌訊息
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

      // 摸牌後，重新設定為可互動狀態
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

      // 摸牌後，重新設定為可互動狀態（可以繼續出牌）
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
    console.log('更新玩家資訊:', playersData);

    // 在遊戲開始前，直接按伺服器位置順序更新
    // 遊戲開始後，startGame 會根據 myPosition 重新排列
    playersData.forEach((playerData, serverPosition) => {
      if (this.players[serverPosition]) {
        this.players[serverPosition].updateInfo(playerData);
        // 儲存玩家的原始伺服器位置
        this.players[serverPosition].serverPosition = serverPosition;
      }
    });
  }

  /**
   * 設定玩家說話狀態（語音通話用）
   * @param {string} peerId - 玩家 ID
   * @param {boolean} isTalking - 是否正在說話
   */
  setPlayerTalking(peerId, isTalking) {
    // 尋找對應玩家並更新說話狀態
    for (const player of this.players) {
      if (player && player.userId === peerId) {
        player.setTalking(isTalking);
        break;
      }
    }
  }

  /**
   * 設定所有玩家的語音按鈕顯示狀態
   * @param {boolean} visible - 是否顯示
   */
  setVoiceButtonsVisible(visible) {
    for (const player of this.players) {
      if (player) {
        player.setVoiceButtonVisible(visible);
      }
    }
    console.log(`[Game] 語音按鈕顯示狀態: ${visible}`);
  }

  /**
   * 設定玩家的語音靜音狀態
   * @param {string} userId - 玩家 ID
   * @param {boolean} isMuted - 是否靜音
   */
  setPlayerVoiceMuted(userId, isMuted) {
    for (const player of this.players) {
      if (player && player.userId === userId) {
        player.setVoiceMuted(isMuted);
        break;
      }
    }
  }

  /**
   * 設定語音按鈕的回調函數
   * @param {Function} callback - 回調函數 (userId, isSelf) => void
   */
  setupVoiceButtonCallbacks(callback) {
    for (const player of this.players) {
      if (player) {
        player.onVoiceButtonClick = callback;
      }
    }
    console.log('[Game] 語音按鈕回調已設定');
  }

  /**
   * 設定語音連線按鈕的回調函數（僅限底部玩家）
   * @param {Function} callback - 回調函數 (connect: boolean) => void
   */
  setupVoiceConnectCallback(callback) {
    // 只設定底部玩家（自己）的連線回調
    const bottomPlayer = this.players.find(p => p && p.position === 'bottom');
    if (bottomPlayer) {
      bottomPlayer.onVoiceConnectClick = callback;
    }
    console.log('[Game] 語音連線按鈕回調已設定');
  }

  /**
   * 設定底部玩家的語音連線狀態
   * @param {'disconnected' | 'connecting' | 'connected'} state - 連線狀態
   */
  setBottomPlayerVoiceState(state) {
    const bottomPlayer = this.players.find(p => p && p.position === 'bottom');
    if (bottomPlayer) {
      bottomPlayer.setVoiceConnectionState(state);
    }
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

  /**
   * 處理擲骰事件
   * @param {Object} data - 擲骰結果資料
   */
  async handleDiceRoll(data) {
    const { diceResults, totalSum, dealerPlayerId, dealerSeatIndex } = data;

    console.log('收到擲骰結果:', diceResults, '總和:', totalSum, '莊家位置:', dealerSeatIndex);

    // 找出莊家名稱
    let dealerName = '玩家';
    for (const player of this.players) {
      if (player.userId === dealerPlayerId) {
        dealerName = player.name || '玩家';
        break;
      }
    }

    // 移除等待文字
    if (this.waitingText) {
      this.container.removeChild(this.waitingText);
      this.waitingText = null;
    }

    // 播放擲骰動畫
    await this.diceRollUI.play(diceResults, totalSum, dealerName, dealerSeatIndex);

    // 隱藏擲骰 UI
    this.diceRollUI.hide();

    console.log('擲骰動畫完成，等待 game_start');
  }

  startGame(data) {
    console.log('遊戲開始!', data);

    // At the start of a new round, clean up the UI from the previous round.
    if (this.endScreenContainer) {
      this.container.removeChild(this.endScreenContainer);
      this.endScreenContainer = null;
    }
    this.resetForNewRound();

    // 播放遊戲背景音樂
    this.audioManager.playBGM('game');

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

    // 設定風位資訊
    this.roundWind = data.roundWind || 'E';
    this.mySeatWind = data.seatWind || 'E';
    this.allSeatWinds = data.allSeatWinds || this.calculateAllSeatWinds();

    const windLabel = this.getWindLabel(this.roundWind);
    const mySeatWindLabel = this.getWindLabel(this.mySeatWind);
    console.log(`我的位置: ${this.myPosition}, 莊家位置: ${this.dealerPosition}, 當前輪次: ${this.currentTurn}`);
    console.log(`場風: ${windLabel}風局, 我的門風: ${mySeatWindLabel}家`);

    if (this.myPosition === this.dealerPosition) {
      console.log('🎴 你是莊家！起手 17 張，第一次打牌後不摸牌');
    }

    // 🔄 根據 myPosition 重新排列玩家資訊
    // 儲存原始玩家資訊（按伺服器位置）
    const originalPlayerInfo = this.players.map(p => ({
      userId: p.userId,
      name: p.name,
      picture: p.picture,
      score: p.score,
      serverPosition: p.serverPosition
    }));

    // 重新分配到視覺位置
    for (let serverPos = 0; serverPos < 4; serverPos++) {
      const visualPos = this.serverToVisualPosition(serverPos);
      const info = originalPlayerInfo[serverPos];
      if (info && this.players[visualPos]) {
        this.players[visualPos].userId = info.userId;
        this.players[visualPos].name = info.name;
        this.players[visualPos].picture = info.picture;
        this.players[visualPos].score = info.score;
        this.players[visualPos].serverPosition = serverPos;
        // 更新顯示
        this.players[visualPos].updateNameDisplay();
        console.log(`玩家 ${info.name} (伺服器位置 ${serverPos}) -> 視覺位置 ${visualPos}`);
      }
    }

    // 初始化牌池（臨時方案：用於本機模擬摸牌）
    this.initializeTilePool();

    // 計算已發出的牌數並更新剩餘牌數
    // 4個玩家 × 16張 + 莊家多1張 = 65張
    // 144張 - 65張 = 79張（不考慮花牌補牌）
    const dealtTiles = 65; // 簡化計算
    this.remainingTiles = 144 - dealtTiles;

    // 創建場風/剩餘牌數顯示
    this.createRemainingTilesText();

    // 更新所有玩家的門風顯示
    this.updatePlayerSeatWinds();

    // 更新所有玩家的輪次狀態
    this.updateTurnStatus();
  }

  /**
   * 將伺服器的絕對位置轉換為相對於自己的視覺位置
   * 自己永遠在下方（視覺位置 0），下家在右邊（1），對家在上方（2），上家在左邊（3）
   * @param {number} serverPosition - 伺服器給的絕對位置
   * @returns {number} - 相對於自己的視覺位置
   */
  serverToVisualPosition(serverPosition) {
    // 例如：如果我是位置 2，伺服器位置 2 應該顯示在下方（視覺位置 0）
    // 伺服器位置 3 應該顯示在右邊（視覺位置 1）
    // 伺服器位置 0 應該顯示在上方（視覺位置 2）
    // 伺服器位置 1 應該顯示在左邊（視覺位置 3）
    return (serverPosition - this.myPosition + 4) % 4;
  }

  /**
   * 將視覺位置轉換回伺服器的絕對位置
   * @param {number} visualPosition - 視覺位置（0=下, 1=右, 2=上, 3=左）
   * @returns {number} - 伺服器的絕對位置
   */
  visualToServerPosition(visualPosition) {
    return (visualPosition + this.myPosition) % 4;
  }

  dealTiles(data) {
    console.log('發牌:', data);

    const { tiles, position } = data;
    // 將伺服器位置轉換為視覺位置
    const visualPosition = this.serverToVisualPosition(position);
    console.log(`發牌: 伺服器位置 ${position} -> 視覺位置 ${visualPosition}`);
    const player = this.players[visualPosition];

    // 播放發牌音效
    this.audioManager.playEffect('deal');

    if (player) {
      // 檢查發牌數據是否有重複牌（每張牌最多4張）
      const tileCount = {};
      tiles.forEach(tile => {
        tileCount[tile] = (tileCount[tile] || 0) + 1;
        if (tileCount[tile] > 4) {
          console.error(`❌ 錯誤：${tile} 在手牌中出現 ${tileCount[tile]} 次！這是伺服器端的BUG`);
        }
      });

      player.setTiles(tiles, this.tileAssets);

      // ✅ 從牌池中移除已發的牌
      tiles.forEach(tile => {
        const removed = this.removeTileFromPool(tile);
        if (!removed) {
          console.warn(`⚠️ 無法從牌池移除 ${tile}（可能已經被移除或不存在）`);
          console.warn(`⚠️ 這可能是伺服器端發送了重複的牌`);
        }
      });

      console.log(`✅ 玩家 ${position}(視覺${visualPosition}) 發牌完成，已從牌池移除 ${tiles.length} 張牌。牌池剩餘: ${this.tilePool.length}張`);

      // 📋 記錄初始手牌
      setTimeout(() => {
        this.logPlayerHand(visualPosition, '初始手牌');

        // 如果所有玩家都發牌完成，記錄所有玩家的手牌
        const allDealt = this.players.every(p => p.tiles.length > 0);
        if (allDealt) {
          this.logAllPlayersHands('遊戲開始 - 所有玩家初始手牌');
        }
      }, 100);
    }
  }

  handlePlayerAction(data) {
    console.log('玩家動作:', data);

    const {playerId, action, tile, currentTurn, remainingTiles} = data;

    // 更新剩餘牌數（如果伺服器發送了這個資訊）
    if (remainingTiles !== undefined) {
      this.updateRemainingTiles(remainingTiles);
    }

    // 更新當前輪次
    if (currentTurn !== undefined) {
      this.currentTurn = currentTurn;
      console.log(`更新輪次: ${this.currentTurn}`);
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
      case 'flower':
        this.handleFlower(playerId, data.flowers);
        break;
    }
  }

  updateTurnStatus() {
    // currentTurn 是伺服器的絕對位置，轉換為視覺位置
    const visualTurn = this.serverToVisualPosition(this.currentTurn);
    console.log(`更新回合狀態: 伺服器回合 ${this.currentTurn} -> 視覺回合 ${visualTurn}`);

    // 更新所有玩家的可互動狀態和輪次指示器
    this.players.forEach((player, index) => {
      // index 是視覺位置，視覺位置 0 是自己
      const isCurrentTurn = (index === visualTurn);
      const isSelf = (index === 0);
      const isMyTurn = isCurrentTurn && isSelf;

      // 更新輪次指示器
      player.setTurnActive(isCurrentTurn, isSelf);

      // 如果玩家已聽牌，禁用互動（伺服器會自動打牌）
      const canInteract = isMyTurn && !player.isTing;
      player.setInteractive(canInteract);

      // 如果輪到自己（視覺位置 0），檢查是否能聽牌或自摸，並顯示提示
      if (isMyTurn) {
        // 顯示「輪到你出牌」提示
        this.showYourTurnHint();

        setTimeout(() => {
          this.checkSelfActions();
        }, 200);
      }
    });
  }

  /**
   * 顯示「輪到你出牌」提示
   */
  showYourTurnHint() {
    // 避免重複顯示（如果公告已經在顯示其他重要訊息）
    if (this.announcementText && this.announcementText.visible) {
      return;
    }

    // 創建或更新輪次提示
    if (!this.yourTurnHint) {
      this.yourTurnHint = new Text({
        text: '輪到你出牌',
        style: {
          fontSize: 36,
          fill: 0xFFD700, // 金色
          fontWeight: 'bold',
          stroke: { color: 0x000000, width: 4 },
          dropShadow: {
            color: 0x000000,
            blur: 4,
            distance: 2
          }
        }
      });
      this.yourTurnHint.anchor.set(0.5);
      this.yourTurnHint.x = this.app.screen.width / 2;
      this.yourTurnHint.y = this.app.screen.height / 2 - 100;
      this.yourTurnHint.zIndex = 1500;
      this.container.addChild(this.yourTurnHint);
    }

    // 顯示提示
    this.yourTurnHint.visible = true;
    this.yourTurnHint.alpha = 1.0;

    // 淡出動畫
    let fadeAlpha = 1.0;
    const fadeInterval = setInterval(() => {
      fadeAlpha -= 0.02;
      if (this.yourTurnHint) {
        this.yourTurnHint.alpha = fadeAlpha;
      }
      if (fadeAlpha <= 0) {
        clearInterval(fadeInterval);
        if (this.yourTurnHint) {
          this.yourTurnHint.visible = false;
        }
      }
    }, 50);

    // 2秒後強制隱藏（以防萬一）
    setTimeout(() => {
      if (this.yourTurnHint) {
        this.yourTurnHint.visible = false;
      }
    }, 2500);
  }

  async handleDiscard(playerId, tile) {
    console.log(`玩家 ${playerId} 打出了 ${tile}`);

    // 找到打出牌的玩家（透過 userId 查找視覺位置）
    let visualPosition = -1;
    for (let i = 0; i < this.players.length; i++) {
      if (this.players[i].userId === playerId) {
        visualPosition = i;
        break;
      }
    }

    if (visualPosition === -1) {
      console.error('未找到玩家:', playerId);
      return;
    }

    const player = this.players[visualPosition];
    // 用於棄牌位置計算（視覺位置）
    const playerPosition = visualPosition;

    // 播放打牌語音
    this.audioManager.playTileVoice(playerId, tile);

    // 🎯 檢查是否為聽牌後的自動打牌
    if (player.isTing) {
      console.log(`🎯 玩家 ${player.name} 聽牌中，自動打出 ${tile}`);

      // 如果是自己聽牌自動打牌（視覺位置 0），顯示提示
      if (visualPosition === 0) {
        this.showAnnouncement(`自動打出 ${this.getTileName(tile)}`, 1500);
      }
    }

    // 創建棄牌容器（包含牌底和牌面）
    // 使用 DiscardManager 處理視覺呈現
    await this.discardManager.addDiscard(tile, playerPosition, this.tileAssets);

    // 從玩家手牌中移除該牌（視覺上）- 只處理自己的手牌
    if (player && visualPosition === 0) {
      // 🎯 如果玩家已聽牌，等待一小段時間確保摸牌動畫完成
      if (player.isTing) {
        console.log(`⏳ [handleDiscard] 玩家已聽牌，等待摸牌完成後再移除 ${tile}`);
        await new Promise(resolve => setTimeout(resolve, 200));
      }

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
    // 只有當不是自己打的牌時才檢查（視覺位置 0 是自己）
    if (visualPosition !== 0) {
      this.checkPossibleActions(tile, visualPosition);
    } else {
      // 如果是自己打的牌，清除動作按鈕
      this.actionButtons.hide();
      this.pendingActions = [];
      this.possibleTingDiscards = {};
      const myPlayer = this.players[0]; // 視覺位置 0 是自己
      if (myPlayer) {
        myPlayer.clearHighlight();
      }
    }
  }

  async handlePlayerDraw(playerId, tile) {
    console.log(`玩家 ${playerId} 摸牌: ${tile}`);

    // 🌸 檢查是否是花牌，如果是則交給 handleFlower 處理
    if (tile && tile.startsWith('flower-')) {
      console.log(`🌸 摸到花牌 ${tile}，轉交給 handleFlower 處理`);
      await this.handleFlower(playerId, [tile]);
      return;
    }

    // 找到摸牌的玩家（視覺位置）
    let visualPosition = -1;
    for (let i = 0; i < this.players.length; i++) {
      if (this.players[i].userId === playerId) {
        visualPosition = i;
        break;
      }
    }

    if (visualPosition === -1) {
      console.error('未找到玩家:', playerId);
      return;
    }

    const player = this.players[visualPosition];
    const playerPosition = visualPosition;

    // 🎯 檢查玩家是否已聽牌
    if (player.isTing) {
      console.log(`🎯 玩家 ${player.name} 已聽牌，摸到 ${tile}，即將自動打出`);

      // 如果是自己聽牌（視覺位置 0），顯示提示
      if (visualPosition === 0) {
        this.showAnnouncement(`聽牌中，摸到 ${this.getTileName(tile)}`, 1500);
      }
    }

    // 只為自己（視覺位置 0）加入手牌，其他玩家不顯示手牌
    if (visualPosition === 0) {
      await player.addTile(tile, this.tileAssets);

      // 📋 記錄摸牌後的手牌狀態
      setTimeout(() => {
        this.logPlayerHand(playerPosition, `摸牌: ${tile}`);
      }, 100);
    }

    // 更新剩餘牌數
    if (this.remainingTiles > 0) {
      this.updateRemainingTiles(this.remainingTiles - 1);
    }

    console.log(`玩家 ${playerPosition} 摸牌完成，剩餘 ${this.remainingTiles} 張`);

    // 如果是自己摸牌（視覺位置 0），檢查是否可以自摸或聽牌
    if (visualPosition === 0) {
      // 等待一小段時間，確保牌已經加入並顯示
      setTimeout(() => {
        this.checkSelfActions();
      }, 100);
    }
  }

  async handleChow(playerId, data) {
    // 處理吃牌
    const chowTiles = data.chowTiles || [];
    const tile = data.tile;
    console.log(`🍜 handleChow - 玩家 ${playerId} 吃了 ${tile}，牌組: ${chowTiles.join(', ')}`);

    // 播放吃牌語音
    this.audioManager.playActionVoice(playerId, 'chi');

    // 找到執行吃牌的玩家（視覺位置）
    let visualPosition = -1;
    for (let i = 0; i < this.players.length; i++) {
      if (this.players[i].userId === playerId) {
        visualPosition = i;
        break;
      }
    }
    const playerPosition = visualPosition;

    if (visualPosition !== -1 && chowTiles.length === 3) {
      const player = this.players[visualPosition];

      console.log(`🍜 吃牌前手牌 (${player.tiles.length}張):`, player.tiles.map(t => t.type));

      // 從手牌中移除用於吃牌的2張牌（不包括被吃的那張）
      // 如果是自己（visualPosition === 0），按牌型移除
      // 如果是其他玩家，他們的手牌都是 'back'，直接移除2張
      if (visualPosition === 0) {
        // 自己的牌，按牌型移除
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
      } else {
        // 其他玩家的牌都是 'back'，直接移除2張
        for (let i = 0; i < 2 && player.tiles.length > 0; i++) {
          const tileToRemove = player.tiles.pop();
          tileToRemove.destroy();
        }
        console.log(`🍜 已移除其他玩家的2張牌`);
      }

      console.log(`🍜 吃牌後手牌 (${player.tiles.length}張):`, player.tiles.map(t => t.type));

      // 重新排列剩餘手牌
      player.rearrangeTiles();

      // 新增吃牌組到顯示區域
      player.addMeld({
        type: 'chow',
        tiles: chowTiles
      });
      await player.updateOpenLayout(this.tileAssets);

      // 從棄牌堆中移除最後一張（被吃的牌）
      this.discardManager.removeLastDiscard();

      console.log(`✅ 玩家 ${playerPosition} 吃牌完成`);

      // 📋 記錄吃牌後的手牌狀態
      setTimeout(() => {
        this.logPlayerHand(playerPosition, `吃牌: ${chowTiles.join(',')}`);
      }, 100);

      // 如果是自己吃牌（視覺位置 0），檢查是否能聽牌或自摸
      if (visualPosition === 0) {
        setTimeout(() => {
          this.checkSelfActions();
        }, 300);
      }
    }

    // 清除動作按鈕
    this.actionButtons.hide();
    this.pendingActions = [];
    this.possibleTingDiscards = {};
    const myPlayer = this.players[0]; // 視覺位置 0 是自己
    if (myPlayer) {
      myPlayer.clearHighlight();
    }
  }

  async handlePong(playerId, tile) {
    // 處理碰牌
    console.log(`玩家 ${playerId} 碰了 ${tile}`);

    // 播放碰牌語音
    this.audioManager.playActionVoice(playerId, 'peng');

    // 找到執行碰牌的玩家（視覺位置）
    let visualPosition = -1;
    for (let i = 0; i < this.players.length; i++) {
      if (this.players[i].userId === playerId) {
        visualPosition = i;
        break;
      }
    }
    const playerPosition = visualPosition;

    if (visualPosition !== -1) {
      const player = this.players[visualPosition];

      // 從手牌中移除2張牌
      // 如果是自己（visualPosition === 0），按牌型移除
      // 如果是其他玩家，他們的手牌都是 'back'，直接移除2張
      let removed = 0;
      if (visualPosition === 0) {
        // 自己的牌，按牌型移除
        for (let i = player.tiles.length - 1; i >= 0 && removed < 2; i--) {
          if (player.tiles[i].type === tile) {
            player.tiles[i].destroy();
            player.tiles.splice(i, 1);
            removed++;
          }
        }
      } else {
        // 其他玩家的牌都是 'back'，直接移除2張
        for (let i = 0; i < 2 && player.tiles.length > 0; i++) {
          const tileToRemove = player.tiles.pop();
          tileToRemove.destroy();
          removed++;
        }
      }

      // 重新排列剩餘手牌
      player.rearrangeTiles();

      // 新增碰牌組到顯示區域
      player.addMeld({
        type: 'pong',
        tiles: [tile, tile, tile]
      });
      await player.updateOpenLayout(this.tileAssets);

      // 從棄牌堆中移除最後一張（被碰的牌）
      this.discardManager.removeLastDiscard();

      console.log(`✅ 玩家 ${playerPosition} 碰牌完成`);

      // 📋 記錄碰牌後的手牌狀態
      setTimeout(() => {
        this.logPlayerHand(playerPosition, `碰牌: ${tile}`);
      }, 100);

      // 如果是自己碰牌（視覺位置 0），檢查是否能聽牌或自摸
      if (visualPosition === 0) {
        setTimeout(() => {
          this.checkSelfActions();
        }, 300);
      }
    }

    // 清除動作按鈕（因為這張牌已經被碰了）
    this.actionButtons.hide();
    this.pendingActions = [];
    this.possibleTingDiscards = {};
    const myPlayer = this.players[0]; // 視覺位置 0 是自己
    if (myPlayer) {
      myPlayer.clearHighlight();
    }
  }

  async handleKong(playerId, data) {
    // 處理槓牌
    const meld = data.meld; // Note: Keys will be capitalized: Type, Tiles
    if (!meld) {
      console.error('無效的槓牌動作: 缺少牌組資訊', data);
      return;
    }
    const tile = meld.Tiles[0];
    console.log(`玩家 ${playerId} 槓了 ${tile} (類型: ${meld.Type})`);

    // 播放槓牌語音
    this.audioManager.playActionVoice(playerId, 'gang');

    // 找到執行槓牌的玩家（視覺位置）
    let visualPosition = -1;
    for (let i = 0; i < this.players.length; i++) {
      if (this.players[i].userId === playerId) {
        visualPosition = i;
        break;
      }
    }
    const playerPosition = visualPosition;

    if (visualPosition !== -1) {
      const player = this.players[visualPosition];

      // 根據槓的類型從手牌中移除對應數量的牌
      // 如果是自己（visualPosition === 0），按牌型移除
      // 如果是其他玩家，他們的手牌都是 'back'，直接移除對應數量
      if (meld.Type === 'kong_exposed') {
        // 明槓，從手牌移除3張
        if (visualPosition === 0) {
          let removed = 0;
          for (let i = player.tiles.length - 1; i >= 0 && removed < 3; i--) {
            if (player.tiles[i].type === tile) {
              player.tiles[i].destroy();
              player.tiles.splice(i, 1);
              removed++;
            }
          }
        } else {
          for (let i = 0; i < 3 && player.tiles.length > 0; i++) {
            const tileToRemove = player.tiles.pop();
            tileToRemove.destroy();
          }
        }
      } else if (meld.Type === 'kong_concealed') {
        // 暗槓，從手牌移除4張
        if (visualPosition === 0) {
          let removed = 0;
          for (let i = player.tiles.length - 1; i >= 0 && removed < 4; i--) {
            if (player.tiles[i].type === tile) {
              player.tiles[i].destroy();
              player.tiles.splice(i, 1);
              removed++;
            }
          }
        } else {
          for (let i = 0; i < 4 && player.tiles.length > 0; i++) {
            const tileToRemove = player.tiles.pop();
            tileToRemove.destroy();
          }
        }
      } else if (meld.Type === 'kong_promoted') {
        // 加槓（從碰升級為槓），從手牌移除1張
        if (visualPosition === 0) {
          for (let i = player.tiles.length - 1; i >= 0; i--) {
            if (player.tiles[i].type === tile) {
              player.tiles[i].destroy();
              player.tiles.splice(i, 1);
              break; // 只移除1張
            }
          }
        } else {
          if (player.tiles.length > 0) {
            const tileToRemove = player.tiles.pop();
            tileToRemove.destroy();
          }
        }
      }

      // 重新排列剩餘手牌
      player.rearrangeTiles();

      // 新增槓牌組到顯示區域
      player.addMeld(meld);
      await player.updateOpenLayout(this.tileAssets);

      // 如果是明槓，從棄牌堆中移除最後一張（被槓的牌）
      if (meld.Type === 'kong_exposed') {
        this.discardManager.removeLastDiscard();
      }

      console.log(`✅ 玩家 ${playerPosition} 槓牌完成`);

      // 📋 記錄槓牌後的手牌狀態
      setTimeout(() => {
        this.logPlayerHand(playerPosition, `槓牌: ${tile} (${meld.Type})`);
      }, 100);

      // 如果是自己槓牌（視覺位置 0），檢查是否能聽牌或自摸（槓牌後會補牌，所以可能自摸）
      if (visualPosition === 0) {
        setTimeout(() => {
          this.checkSelfActions();
        }, 300);
      }
    }

    // 清除動作按鈕
    this.actionButtons.hide();
    this.pendingActions = [];
    this.possibleTingDiscards = {};
    const myPlayer = this.players[0]; // 視覺位置 0 是自己
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

    // 播放聽牌語音
    this.audioManager.playActionVoice(playerId, 'ting');

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

    // 如果是自己（視覺位置 0），顯示提示
    if (player === this.players[0]) {
      this.showAnnouncement('聽牌！', 2000);
    } else {
      // 其他玩家聽牌，也顯示提示
      this.showAnnouncement(`${player.name} 聽牌！`, 2000);
    }
  }

  /**
   * 處理花牌廣播
   * @param {string} playerId - 玩家ID
   * @param {Array<string>} flowers - 花牌列表
   */
  async handleFlower(playerId, flowers) {
    console.log(`玩家 ${playerId} 摸到花牌:`, flowers);

    const player = this.players.find(p => p.userId === playerId);
    if (!player) {
      console.error('未找到玩家:', playerId);
      return;
    }
    
    const visualPosition = this.serverToVisualPosition(player.serverPosition);

    for (const flower of flowers) {
      player.addFlower(flower);
    }

    // Call the unified layout update
    await player.updateOpenLayout(this.tileAssets);

    // For other players, remove a corresponding number of placeholder tiles from their hand
    if (visualPosition !== 0) {
      for (let i = 0; i < flowers.length && player.tiles.length > 0; i++) {
        const tileToRemove = player.tiles.pop();
        tileToRemove.destroy();
      }
      player.rearrangeTiles();
    }

    console.log(`✅ 花牌處理完成，玩家 ${player.name} 現有 ${player.flowers.length} 張花牌`);
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

  /**
   * 顯示玩家加入通知
   * @param {string} playerName - 玩家名稱
   */
  showPlayerJoinNotification(playerName) {
    // 播放加入音效
    this.audioManager.playPlayerJoin();

    // 顯示大字通知
    this.showAnnouncement(`${playerName} 加入遊戲`, 2000);

    console.log(`🎮 玩家 ${playerName} 加入遊戲`);
  }

  /**
   * 顯示玩家離開/斷線通知
   * @param {string} playerName - 玩家名稱
   */
  showPlayerLeftNotification(playerName) {
    // 播放離開音效
    this.audioManager.playPlayerLeft();

    // 顯示大字通知
    this.showAnnouncement(`${playerName} 斷線，由電腦代打`, 2000);

    console.log(`🎮 玩家 ${playerName} 斷線，由電腦代打`);
  }

  /**
   * 獲取牌的中文名稱（用於顯示）
   */
  getTileName(tile) {
    const tileNames = {
      // 萬子
      'wan-1': '一萬', 'wan-2': '二萬', 'wan-3': '三萬', 'wan-4': '四萬',
      'wan-5': '五萬', 'wan-6': '六萬', 'wan-7': '七萬', 'wan-8': '八萬', 'wan-9': '九萬',
      // 條子
      'tiao-1': '一條', 'tiao-2': '二條', 'tiao-3': '三條', 'tiao-4': '四條',
      'tiao-5': '五條', 'tiao-6': '六條', 'tiao-7': '七條', 'tiao-8': '八條', 'tiao-9': '九條',
      // 筒子
      'tong-1': '一筒', 'tong-2': '二筒', 'tong-3': '三筒', 'tong-4': '四筒',
      'tong-5': '五筒', 'tong-6': '六筒', 'tong-7': '七筒', 'tong-8': '八筒', 'tong-9': '九筒',
      // 風牌
      'dong': '東', 'nan': '南', 'xi': '西', 'bei': '北',
      // 箭牌
      'zhong': '中', 'fa': '發', 'bai': '白'
    };
    return tileNames[tile] || tile;
  }

  gameOver(data) {
    console.log('遊戲結束:', data);

    const { winner, winType, points } = data;

    // 顯示遊戲結果
    const resultText = new Text({
      text: `遊戲結束!
${winner} 胡牌 (${winType})
得分: ${points}`,
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
    // 牌桌是 92% 大小，所以邊緣在 46% 的位置，牌山應該在稍微內側
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
   * 取得場風中文名稱
   */
  getWindLabel(wind) {
    const windLabels = { E: '東', S: '南', W: '西', N: '北' };
    return windLabels[wind] || '東';
  }

  /**
   * 創建剩餘牌數文字顯示（含場風）
   */
  createRemainingTilesText() {
    // 移除舊的顯示
    if (this.wallTextContainer) {
      this.container.removeChild(this.wallTextContainer);
    }

    // 創建新的容器（獨立於 wallContainer，確保在最上層）
    this.wallTextContainer = new Container();

    // 創建背景（加寬以容納場風資訊）
    const bg = new Graphics();
    bg.roundRect(-120, -20, 240, 40, 10);
    bg.fill({ color: 0x000000, alpha: 0.7 });
    bg.stroke({ width: 2, color: 0xFFD700 }); // 金色邊框

    // 創建文字（場風 | 海底）
    const windLabel = this.getWindLabel(this.roundWind);
    this.wallText = new Text({
      text: `${windLabel}風局 | 海底: ${this.remainingTiles}張`,
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

    // 設定位置（左上角）
    this.wallTextContainer.x = 140;
    this.wallTextContainer.y = 30;

    // 設定 zIndex 確保顯示在最上層
    this.wallTextContainer.zIndex = 1000;

    // 加入到主容器的最上層
    this.container.addChild(this.wallTextContainer);

    console.log(`✅ 場風/剩餘牌數顯示已創建: ${windLabel}風局, ${this.remainingTiles}張`);
  }

  /**
   * 更新剩餘牌數
   */
  updateRemainingTiles(count) {
    this.remainingTiles = count;
    if (this.wallText) {
      const windLabel = this.getWindLabel(this.roundWind);
      this.wallText.text = `${windLabel}風局 | 海底: ${this.remainingTiles}張`;
      console.log(`🎲 剩餘牌數更新: ${this.remainingTiles}張`);
    }

    // The server will handle the draw condition and send a message
  }

  /**
   * 更新場風顯示
   */
  updateRoundWindDisplay() {
    if (this.wallText) {
      const windLabel = this.getWindLabel(this.roundWind);
      this.wallText.text = `${windLabel}風局 | 海底: ${this.remainingTiles}張`;
    }
  }

  /**
   * 根據莊家位置計算所有玩家的門風（前端備用）
   */
  calculateAllSeatWinds() {
    const winds = ['E', 'S', 'W', 'N'];
    const seatWinds = [];
    for (let seat = 0; seat < 4; seat++) {
      const offset = (seat - this.dealerPosition + 4) % 4;
      seatWinds.push(winds[offset]);
    }
    return seatWinds;
  }

  /**
   * 更新所有玩家的門風顯示
   */
  updatePlayerSeatWinds() {
    for (let serverPos = 0; serverPos < 4; serverPos++) {
      const visualPos = this.serverToVisualPosition(serverPos);
      const seatWind = this.allSeatWinds[serverPos];

      if (this.players[visualPos]) {
        this.players[visualPos].setSeatWind(seatWind);
      }
    }
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

    // 播放按鈕音效
    this.audioManager.playButtonSound();

    const myPlayer = this.players[0]; // 視覺位置 0 是自己
    if (!myPlayer) return;

    this.isDeclaringTing = true;

    const validDiscards = Object.keys(this.possibleTingDiscards);
    myPlayer.highlightTiles(validDiscards);
    myPlayer.setInteractive(true); // 確保玩家在選擇聽牌時可互動
    
    this.showAnnouncement('請選擇一張牌打出以聽牌', 3000);
    this.actionButtons.hide();
  }

  /**
   * 處理碰牌動作
   */
  handlePongAction() {
    console.log('執行碰牌');

    // 播放按鈕音效
    this.audioManager.playButtonSound();

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

    // 播放按鈕音效
    this.audioManager.playButtonSound();

    if (this.lastDiscardedTile && this.pendingActions.chow) {
      if (this.pendingActions.chow.length > 1) {
        // 如果有多個吃牌組合，提示玩家選擇
        this.chowSelectionUI.promptSelection(this.pendingActions.chow, this.lastDiscardedTile, (selectedCombination) => {
          this.sendAction('chow', this.lastDiscardedTile, selectedCombination);
        });
      } else if (this.pendingActions.chow.length === 1) {
        // 只有一個組合，直接執行
        const combination = this.pendingActions.chow[0];
        this.sendAction('chow', this.lastDiscardedTile, combination);
      }
    }
  }



  /**
   * 處理槓牌動作
   */
  handleKongAction() {
    console.log('執行槓牌');

    // 播放按鈕音效
    this.audioManager.playButtonSound();

    let tileToKong = null;
    let isConcealed = false;

    // Case 1: Self-drawn kong (promoted or concealed)
    if (this.selfKongOptions && this.selfKongOptions.length > 0) {
        // UI should let user choose if multiple options. For now, take the first.
        tileToKong = this.selfKongOptions[0];

        const myPlayer = this.players[0]; // 視覺位置 0 是自己
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
      
      // 清理所有待處理動作狀態
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
    console.log('✅ 執行胡牌動作');

    // 播放按鈕音效
    this.audioManager.playButtonSound();

    const myPlayer = this.players[0]; // 視覺位置 0 是自己
    const myHand = myPlayer ? myPlayer.tiles.map(t => t.type) : [];
    console.log('當前手牌:', myHand);
    console.log('lastDiscardedTile:', this.lastDiscardedTile);
    console.log('currentTurn:', this.currentTurn, 'myPosition:', this.myPosition);

    // 判斷是自摸還是別人放炮
    let isSelfDrawn = false;
    let winTile = this.lastDiscardedTile;

    // 如果沒有 lastDiscardedTile，說明是自摸
    // currentTurn 是伺服器位置，myPosition 也是伺服器位置
    if (!this.lastDiscardedTile || this.currentTurn === this.myPosition) {
      isSelfDrawn = true;
      console.log('判定為自摸，尋找胡牌牌型...');
      // Find the winning tile from hand (server will verify)
      // This is a simplified client-side check
      for (const tile of myHand) {
        const handWithoutTile = myHand.filter((t, i) => {
          const firstIndex = myHand.indexOf(tile);
          return i !== firstIndex;
        });
        if (this.canHu(handWithoutTile, tile)) {
          winTile = tile;
          console.log('找到胡牌牌型，胡牌:', winTile);
          break;
        }
      }
    } else {
      console.log('判定為放炮，胡牌:', winTile);
    }

    console.log('準備發送胡牌動作:', { winTile, isSelfDrawn });

    if (this.ws && winTile) {
      console.log('✅ 發送胡牌動作到伺服器');
      this.ws.sendAction('hu', {
        tile: winTile,
        isSelfDrawn: isSelfDrawn
      });
    } else {
      console.error('❌ 無法發送胡牌動作:', { hasWs: !!this.ws, winTile });
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

    // 播放按鈕音效
    this.audioManager.playButtonSound();

    // 清理吃牌選擇介面（如果存在）
    this.chowSelectionUI.clear();

    // 立即隱藏按鈕，避免重複點擊
    this.actionButtons.hide();
    this.pendingActions = [];

    // 清除動作超時計時器，避免重複發送 pass
    if (this.actionTimeout) {
      clearTimeout(this.actionTimeout);
      this.actionTimeout = null;
    }

    // 發送過的動作到伺服器
    if (this.lastDiscardedTile) {
      this.sendAction('pass', this.lastDiscardedTile);
    }

    this.possibleTingDiscards = {};
    this.isDeclaringTing = false;

    // 清除牌組高亮
    const myPlayer = this.players[0]; // 視覺位置 0 是自己
    if (myPlayer) {
      myPlayer.clearHighlight();
    }
  }

  resize(width, height) {
    if (this.table) {
      this.table.resize(width, height);
    }
    
    if (this.discardManager) {
      this.discardManager.resize(width, height);
    }

    this.players.forEach(player => {
      player.resize(width, height);
    });

    // 調整動作按鈕位置
    if (this.actionButtons) {
      this.actionButtons.resize(width, height);
    }
  }

    

    

    /**

     * 處理遊戲勝利

     */
  handleGameWin(data) {
    console.log('遊戲勝利', data);
    const { winnerId, winnerName, winResult, countdown } = data;
    const { HandTypes, TotalTai, BaseScore, WinningHand, Melds, WinTile } = winResult;

    // 播放胡牌語音和勝利音效
    if (winnerId) {
      this.audioManager.playActionVoice(winnerId, 'hu');
    }
    this.audioManager.playEffect('win');

    // Display the winning hand
    if (WinningHand && Melds && WinTile) {
        this.winningHandDisplayHandler.display(WinningHand, Melds, WinTile);
    }

    // 建構牌型描述
    const handTypesStr = HandTypes.map(ht => `${ht.Name} (${ht.Tai}臺)`).join(' ');
    const title = `恭喜 ${winnerName} 胡牌！`;
    const details = `${handTypesStr}\n總計: ${TotalTai}臺, 得分: ${BaseScore}`;

    setTimeout(() => {
        this.showEndRoundScreen(title, details, countdown);
    }, 5000);
  }

  /**
   * 處理流局（荒牌）
   */
  handleGameDraw(data) {
    console.log('遊戲流局，無人勝出', data);
    const { countdown = 5 } = data; // 預設5秒倒數計時

    // 播放流局音效
    this.audioManager.playEffect('lose');

    this.showEndRoundScreen('流局', '無人勝出', countdown);
  }

  /**
   * 檢查可執行的對手動作（吃/碰/槓/胡）
   * 注：現在由伺服器端偵測並透過 handlePossibleActions 通知
   */
  checkPossibleActions(tile, playerPosition) {
    // 伺服器端已經偵測並廣播，這裡不需要客戶端檢查
    console.log('checkPossibleActions 呼叫（由伺服器端處理）');
  }

  /**
   * 檢查自己可以執行的動作（暗槓/聽牌/自摸）
   */
  checkSelfActions() {
    const myPlayer = this.players[0]; // 視覺位置 0 是自己
    if (!myPlayer) return;

    // 如果已經聽牌，不再檢查其他動作
    if (myPlayer.isTing) {
      console.log('玩家已聽牌，跳過動作檢查');
      return;
    }

    // 1. 發送聽牌檢查請求
    if (this.ws) {
      console.log('向伺服器發送聽牌檢查請求');
      this.ws.sendAction('check_ting');
    }

    // 2. 本地檢查暗槓/加槓

    this.selfKongOptions = [];
    const hand = myPlayer.tiles.map(t => t.type);

    // 檢查加槓
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

    // 檢查暗槓
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
    
    // 如果有可用的槓牌選項，顯示按鈕
    // 注意：這可能會覆蓋由聽牌檢查顯示的按鈕，這是一個已知的UI限制
    if (this.selfKongOptions.length > 0) {
        console.log("發現可執行的個人槓牌:", this.selfKongOptions);
        this.actionButtons.show(['kong', 'cancel']);
    }

    // TODO: 在此處加入檢查自摸等其他動作
  }

  /**
   * 處理可執行動作通知
   */
  handlePossibleActions(data) {
    const {playerId, tile, actions, timeout} = data;
    console.log('收到可執行動作通知:', {playerId, tile, actions, timeout});

    // 檢查是否是當前玩家（使用 userId 而不是 id）
    // 視覺位置 0 是自己
    const myPlayer = this.players[0];
    console.log('DEBUG: myPlayer.userId =', myPlayer?.userId, 'playerId =', playerId, '是否相等?', myPlayer?.userId === playerId);

    if (!myPlayer || myPlayer.userId !== playerId) {
      console.log('不是當前玩家的動作，忽略 - myPlayer.userId:', myPlayer?.userId, 'expected:', playerId);
      return;
    }

    // 儲存最後打出的牌和可執行動作
    this.lastDiscardedTile = tile;
    this.pendingActions = actions;
    console.log('✅ 儲存 lastDiscardedTile:', this.lastDiscardedTile);
    console.log('✅ 儲存 pendingActions:', this.pendingActions);

    // 確定要顯示的按鈕
    const buttonsToShow = [];
    if (actions.pong) buttonsToShow.push('pong');
    if (actions.chow) buttonsToShow.push('chow');
    if (actions.kong) buttonsToShow.push('kong');
    if (actions.hu) {
      console.log('✅ 偵測到可以胡牌，新增 hu 按鈕');
      buttonsToShow.push('hu');
    }
    buttonsToShow.push('cancel'); // 總是顯示"過"按鈕

    console.log('✅ 顯示動作按鈕:', buttonsToShow);

    // 顯示按鈕
    if (this.actionButtons) {
      console.log('✅ actionButtons 存在，準備顯示按鈕');
      console.log('✅ actionButtons.container.zIndex:', this.actionButtons.container.zIndex);
      this.actionButtons.show(buttonsToShow);
      console.log('✅ 按鈕已顯示');
    } else {
      console.error('❌ actionButtons 不存在！');
    }

    // 清除之前的超時
    if (this.actionTimeout) {
      clearTimeout(this.actionTimeout);
    }

    // 設定超時自動過
    this.actionTimeout = setTimeout(() => {
      console.log('動作選擇超時，自動過');
      this.sendAction('pass', tile);
      if (this.actionButtons) {
        this.actionButtons.hide();
      }
      // 清理吃牌選擇介面（如果存在）
      this.chowSelectionUI.clear();
    }, timeout * 1000);
  }

  /**
   * 發送動作選擇到伺服器
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

    console.log('發送動作:', message);
    this.ws.send(message);

    // 清除超時
    if (this.actionTimeout) {
      clearTimeout(this.actionTimeout);
      this.actionTimeout = null;
    }

    // 隱藏按鈕
    if (this.actionButtons) {
      this.actionButtons.hide();
    }
  }

  /**
   * 顯示回合結束畫面（胡牌或流局）
   */
  async showEndRoundScreen(title, details, countdown) {
    // 禁用所有玩家互動
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

    // 標題（例如 "恭喜..." 或 "流局"）
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

    // 詳情（例如牌型和分數）
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

    // 等待下一局開始的提示
    const waitingText = new Text({
      text: '等待伺服器開始新的一局...',
      style: {
        fontSize: 28,
        fill: 0xCCCCCC,
      }
    });
    waitingText.anchor.set(0.5);
    waitingText.x = this.app.screen.width / 2;
    waitingText.y = this.app.screen.height / 2 + 100;
    endScreenContainer.addChild(waitingText);
    
    this.endScreenContainer = endScreenContainer;
    this.container.addChild(this.endScreenContainer);
  }

  /**
   * 為新的一局重置UI
   */
  resetForNewRound() {
    // 清空棄牌
    this.discardManager.clear();

    // 重置所有玩家
    this.players.forEach(p => p.reset());

    console.log("UI已重置，準備新的一局");
  }
}
