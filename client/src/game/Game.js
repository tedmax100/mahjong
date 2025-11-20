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
  }

  async init() {
    // 加载素材
    await this.loadAssets();

    // 创建牌桌
    this.table = new Table(this.app.screen.width, this.app.screen.height);
    this.container.addChild(this.table.container);

    // 添加弃牌区域容器
    this.container.addChild(this.discardContainer);

    // 创建玩家区域
    this.createPlayers();

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

    console.log(`我的位置: ${this.myPosition}, 当前轮次: ${this.currentTurn}`);

    // 更新所有玩家的轮次状态
    this.updateTurnStatus();
  }

  dealTiles(data) {
    console.log('发牌:', data);

    const { tiles, position } = data;
    const player = this.players[position];

    if (player) {
      player.setTiles(tiles, this.tileAssets);
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

    // 设置弃牌大小（稍微大一些，方便阅读）
    const scale = 0.9;
    tileSprite.scale.set(scale);

    // 计算弃牌位置（在中央区域，根据玩家位置排列）
    const centerX = this.app.screen.width / 2;
    const centerY = this.app.screen.height / 2;
    const tileWidth = 60 * scale;
    const tileHeight = 80 * scale;
    const spacing = 5;

    // 计算该玩家已经打出的牌数
    const playerDiscards = this.discardedTiles.filter(d => d.playerPosition === playerPosition);
    const discardIndex = playerDiscards.length;

    // 根据玩家位置计算弃牌位置
    let x, y;
    const maxTilesPerRow = 6; // 每行最多6张牌
    const row = Math.floor(discardIndex / maxTilesPerRow);
    const col = discardIndex % maxTilesPerRow;

    // 所有玩家的弃牌都朝向下方玩家，方便阅读
    switch (playerPosition) {
      case 0: // 底部玩家 - 弃牌放在中央偏下，从左到右排列
        x = centerX - (maxTilesPerRow * (tileWidth + spacing)) / 2 + col * (tileWidth + spacing) + tileWidth / 2;
        y = centerY + 120 + row * (tileHeight + spacing) + tileHeight / 2;
        // 不旋转
        break;

      case 1: // 右侧玩家 - 弃牌放在中央偏右，从上到下排列
        x = centerX + 150 + row * (tileWidth + spacing) + tileWidth / 2;
        y = centerY - (maxTilesPerRow * (tileHeight + spacing)) / 2 + col * (tileHeight + spacing) + tileHeight / 2;
        // 不旋转，保持正向
        break;

      case 2: // 顶部玩家 - 弃牌放在中央偏上，从右到左排列（旋转180度）
        x = centerX + (maxTilesPerRow * (tileWidth + spacing)) / 2 - col * (tileWidth + spacing) - tileWidth / 2;
        y = centerY - 150 - row * (tileHeight + spacing) - tileHeight / 2;
        tileSprite.rotation = Math.PI; // 旋转180度朝向下方玩家
        break;

      case 3: // 左侧玩家 - 弃牌放在中央偏左，从下到上排列
        x = centerX - 180 - row * (tileWidth + spacing) - tileWidth / 2;
        y = centerY + (maxTilesPerRow * (tileHeight + spacing)) / 2 - col * (tileHeight + spacing) - tileHeight / 2;
        // 不旋转，保持正向
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

  resize(width, height) {
    if (this.table) {
      this.table.resize(width, height);
    }

    this.players.forEach(player => {
      player.resize(width, height);
    });
  }
}
