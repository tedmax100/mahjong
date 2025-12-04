import { describe, it, expect, beforeEach, vi } from 'vitest';

/**
 * 測試聽牌偵測功能
 * 這個測試檔案專門測試吃碰槓後的聽牌偵測邏輯
 */

// 模擬的 Game 邏輯類（只包含聽牌偵測相關的方法）
class MockGameLogic {
  constructor() {
    this.players = [
      { melds: [], tiles: [] },
      { melds: [], tiles: [] },
      { melds: [], tiles: [] },
      { melds: [], tiles: [] }
    ];
    this.myPosition = 0;
    this.currentTurn = 0;
    this.actionButtonsShown = false;
    this.lastShownActions = [];
  }

  // 從 Game.js 複製的 canHu 方法
  canHu(hand, tile, meldCount = null) {
    if (meldCount === null) {
      meldCount = this.players[this.myPosition].melds.length;
    }

    const fullHand = [...hand, tile];
    const requiredMelds = 5 - meldCount;
    const requiredTiles = requiredMelds * 3 + 2;

    if (fullHand.length !== requiredTiles) {
      return false;
    }

    const tileCount = {};
    fullHand.forEach(t => {
      tileCount[t] = (tileCount[t] || 0) + 1;
    });

    for (const [tile, count] of Object.entries(tileCount)) {
      if (count >= 2) {
        const remainingTiles = { ...tileCount };
        remainingTiles[tile] -= 2;

        if (this.canFormMelds(remainingTiles, requiredMelds)) {
          return true;
        }
      }
    }

    return false;
  }

  // 從 Game.js 複製的 canFormMelds 方法
  canFormMelds(tileCount, meldsNeeded) {
    const tiles = { ...tileCount };

    Object.keys(tiles).forEach(key => {
      if (tiles[key] === 0) delete tiles[key];
    });

    if (Object.keys(tiles).length === 0 && meldsNeeded === 0) {
      return true;
    }

    if (Object.keys(tiles).length === 0 || meldsNeeded === 0) {
      return false;
    }

    const firstTile = Object.keys(tiles).sort()[0];
    const count = tiles[firstTile];

    if (count >= 3) {
      const newTiles = { ...tiles };
      newTiles[firstTile] -= 3;
      if (newTiles[firstTile] === 0) delete newTiles[firstTile];

      if (this.canFormMelds(newTiles, meldsNeeded - 1)) {
        return true;
      }
    }

    const match = firstTile.match(/^(wan|tong|tiao)-(\d)$/);
    if (match) {
      const suit = match[1];
      const num = parseInt(match[2]);

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

    return false;
  }

  // 從 Game.js 複製的 checkReadyHand 方法
  checkReadyHand(hand) {
    const meldCount = this.players[this.myPosition].melds.length;
    const allPossibleTiles = [];

    ['wan', 'tong', 'tiao'].forEach(suit => {
      for (let num = 1; num <= 9; num++) {
        allPossibleTiles.push(`${suit}-${num}`);
      }
    });

    ['dong', 'nan', 'xi', 'bei', 'zhong', 'fa', 'bai'].forEach(tile => {
      allPossibleTiles.push(tile);
    });

    const readyTiles = [];
    for (const tile of allPossibleTiles) {
      if (this.canHu(hand, tile, meldCount)) {
        readyTiles.push(tile);
      }
    }

    return readyTiles;
  }

  // 從 Game.js 複製的 checkSelfActions 方法
  checkSelfActions() {
    const myPlayer = this.players[this.myPosition];
    if (!myPlayer || !myPlayer.tiles || this.myPosition !== this.currentTurn) {
      return;
    }

    const myHand = myPlayer.tiles;
    const actions = [];

    // 檢查是否可以自摸
    for (const tile of myHand) {
      const handWithoutTile = myHand.filter((t, i) => {
        const firstIndex = myHand.indexOf(tile);
        return i !== firstIndex;
      });

      if (this.canHu(handWithoutTile, tile)) {
        actions.push('hu');
        break;
      }
    }

    // 如果不能自摸，檢查是否聽牌
    if (!actions.includes('hu')) {
      const readyTiles = this.checkReadyHand(myHand);
      if (readyTiles.length > 0) {
        actions.push('ready');
      }
    }

    // 模擬顯示動作按鈕
    if (actions.length > 0) {
      this.actionButtonsShown = true;
      this.lastShownActions = [...actions, 'cancel'];
    }

    return actions;
  }

  // 設定玩家手牌
  setPlayerTiles(tiles) {
    this.players[this.myPosition].tiles = tiles;
  }

  // 新增吃碰槓牌組
  addMeld(meld) {
    this.players[this.myPosition].melds.push(meld);
  }

  // 重置狀態
  reset() {
    this.players[this.myPosition].tiles = [];
    this.players[this.myPosition].melds = [];
    this.actionButtonsShown = false;
    this.lastShownActions = [];
  }
}

describe('聽牌偵測功能測試', () => {
  let game;

  beforeEach(() => {
    game = new MockGameLogic();
  });

  describe('checkSelfActions - 基本功能', () => {
    it('應該在自摸時顯示胡牌按鈕', () => {
      // 設定一個已經胡牌的手牌
      game.setPlayerTiles([
        'wan-1', 'wan-2', 'wan-3',
        'wan-4', 'wan-5', 'wan-6',
        'tong-1', 'tong-2', 'tong-3',
        'tiao-7', 'tiao-8', 'tiao-9',
        'tiao-1', 'tiao-2', 'tiao-3',
        'dong', 'dong'
      ]);

      const actions = game.checkSelfActions();

      expect(actions).toContain('hu');
      expect(actions).not.toContain('ready');
      expect(game.actionButtonsShown).toBe(true);
      expect(game.lastShownActions).toContain('hu');
      expect(game.lastShownActions).toContain('cancel');
    });

    it('應該在聽牌時顯示聽牌按鈕', () => {
      // 設定一個聽牌的手牌（差一張就能胡）
      game.setPlayerTiles([
        'wan-1', 'wan-2', 'wan-3',
        'wan-4', 'wan-5', 'wan-6',
        'tong-1', 'tong-2', 'tong-3',
        'tiao-7', 'tiao-8', 'tiao-9',
        'tiao-1', 'tiao-2', 'tiao-3',
        'dong' // 只有一張東，聽東
      ]);

      const actions = game.checkSelfActions();

      expect(actions).toContain('ready');
      expect(actions).not.toContain('hu');
      expect(game.actionButtonsShown).toBe(true);
      expect(game.lastShownActions).toContain('ready');
      expect(game.lastShownActions).toContain('cancel');
    });

    it('應該在既不能自摸也不能聽牌時不顯示按鈕', () => {
      // 設定一個雜亂的手牌
      game.setPlayerTiles([
        'wan-1', 'wan-3', 'wan-5',
        'tong-2', 'tong-4', 'tong-6',
        'tiao-1', 'dong', 'nan',
        'xi', 'bei'
      ]);

      const actions = game.checkSelfActions();

      expect(actions).not.toContain('hu');
      expect(actions).not.toContain('ready');
      expect(game.actionButtonsShown).toBe(false);
    });
  });

  describe('吃牌後的聽牌偵測', () => {
    it('應該在吃牌後偵測到聽牌', () => {
      // 模擬吃了一組牌（wan-7, wan-8, wan-9）
      game.addMeld({
        type: 'chow',
        tiles: ['wan-7', 'wan-8', 'wan-9']
      });

      // 剩餘手牌：還需要4組面子 + 1對將 = 14張（包含要胡的）
      game.setPlayerTiles([
        'wan-1', 'wan-2', 'wan-3',
        'tong-5', 'tong-5', 'tong-5',
        'tiao-2', 'tiao-3', 'tiao-4',
        'dong', 'dong', 'dong',
        'zhong' // 聽中
      ]);

      const actions = game.checkSelfActions();

      expect(actions).toContain('ready');

      // 驗證聽的牌
      const readyTiles = game.checkReadyHand(game.players[0].tiles);
      expect(readyTiles).toContain('zhong');
    });

    it('應該在吃了兩組牌後偵測到聽牌', () => {
      // 吃了兩組牌
      game.addMeld({ type: 'chow', tiles: ['wan-1', 'wan-2', 'wan-3'] });
      game.addMeld({ type: 'chow', tiles: ['tong-4', 'tong-5', 'tong-6'] });

      // 剩餘手牌：還需要3組面子 + 1對將 = 11張
      game.setPlayerTiles([
        'tiao-1', 'tiao-2', 'tiao-3',
        'dong', 'dong', 'dong',
        'nan', 'nan', 'nan',
        'zhong' // 聽中
      ]);

      const actions = game.checkSelfActions();

      expect(actions).toContain('ready');

      const readyTiles = game.checkReadyHand(game.players[0].tiles);
      expect(readyTiles).toContain('zhong');
    });

    it('應該在吃牌後偵測到兩面聽', () => {
      game.addMeld({ type: 'chow', tiles: ['tiao-3', 'tiao-4', 'tiao-5'] });
      game.addMeld({ type: 'pong', tiles: ['zhong', 'zhong', 'zhong'] });

      // 手牌：7,8萬 + 8,8桶 + 東,東,東 + 南,南,南
      game.setPlayerTiles([
        'wan-7', 'wan-8',
        'tong-8', 'tong-8',
        'dong', 'dong', 'dong',
        'nan', 'nan', 'nan'
      ]);

      const actions = game.checkSelfActions();

      expect(actions).toContain('ready');

      const readyTiles = game.checkReadyHand(game.players[0].tiles);
      expect(readyTiles).toContain('wan-6');
      expect(readyTiles).toContain('wan-9');
    });
  });

  describe('碰牌後的聽牌偵測', () => {
    it('應該在碰牌後偵測到聽牌', () => {
      // 碰了一組牌
      game.addMeld({
        type: 'pong',
        tiles: ['wan-1', 'wan-1', 'wan-1']
      });

      // 剩餘手牌
      game.setPlayerTiles([
        'wan-2', 'wan-3', 'wan-4',
        'tong-5', 'tong-6', 'tong-7',
        'tiao-7', 'tiao-8', 'tiao-9',
        'zhong', 'zhong', 'zhong',
        'dong' // 聽東
      ]);

      const actions = game.checkSelfActions();

      expect(actions).toContain('ready');

      const readyTiles = game.checkReadyHand(game.players[0].tiles);
      expect(readyTiles).toContain('dong');
    });

    it('應該在碰了多組牌後偵測到聽牌', () => {
      // 碰了三組牌
      game.addMeld({ type: 'pong', tiles: ['wan-1', 'wan-1', 'wan-1'] });
      game.addMeld({ type: 'pong', tiles: ['wan-2', 'wan-2', 'wan-2'] });
      game.addMeld({ type: 'pong', tiles: ['wan-3', 'wan-3', 'wan-3'] });

      // 剩餘手牌：只需要2組面子 + 1對將 = 8張
      game.setPlayerTiles([
        'tong-1', 'tong-2', 'tong-3',
        'tiao-7', 'tiao-8', 'tiao-9',
        'dong' // 聽東
      ]);

      const actions = game.checkSelfActions();

      expect(actions).toContain('ready');

      const readyTiles = game.checkReadyHand(game.players[0].tiles);
      expect(readyTiles).toContain('dong');
    });
  });

  describe('槓牌後的聽牌偵測', () => {
    it('應該在明槓後偵測到聽牌', () => {
      // 明槓一組
      game.addMeld({
        type: 'kong_exposed',
        tiles: ['wan-1', 'wan-1', 'wan-1', 'wan-1']
      });

      // 剩餘手牌
      game.setPlayerTiles([
        'wan-2', 'wan-3', 'wan-4',
        'tong-5', 'tong-6', 'tong-7',
        'tiao-7', 'tiao-8', 'tiao-9',
        'zhong', 'zhong', 'zhong',
        'dong' // 聽東
      ]);

      const actions = game.checkSelfActions();

      expect(actions).toContain('ready');
    });

    it('應該在暗槓後偵測到聽牌', () => {
      // 暗槓一組
      game.addMeld({
        type: 'kong_concealed',
        tiles: ['dong', 'dong', 'dong', 'dong']
      });

      // 剩餘手牌
      game.setPlayerTiles([
        'wan-1', 'wan-2', 'wan-3',
        'wan-4', 'wan-5', 'wan-6',
        'tong-1', 'tong-2', 'tong-3',
        'tiao-7', 'tiao-8', 'tiao-9',
        'zhong' // 聽中
      ]);

      const actions = game.checkSelfActions();

      expect(actions).toContain('ready');

      const readyTiles = game.checkReadyHand(game.players[0].tiles);
      expect(readyTiles).toContain('zhong');
    });

    it('應該在加槓後偵測到聽牌', () => {
      // 加槓（從碰升級為槓）
      game.addMeld({
        type: 'kong_promoted',
        tiles: ['wan-5', 'wan-5', 'wan-5', 'wan-5']
      });

      // 剩餘手牌
      game.setPlayerTiles([
        'wan-1', 'wan-2', 'wan-3',
        'tong-4', 'tong-5', 'tong-6',
        'tiao-7', 'tiao-8', 'tiao-9',
        'dong', 'dong', 'dong',
        'nan' // 聽南
      ]);

      const actions = game.checkSelfActions();

      expect(actions).toContain('ready');
    });
  });

  describe('混合吃碰槓後的聽牌偵測', () => {
    it('應該在吃+碰+槓的組合後偵測到聽牌', () => {
      // 各種組合
      game.addMeld({ type: 'chow', tiles: ['wan-1', 'wan-2', 'wan-3'] });
      game.addMeld({ type: 'pong', tiles: ['tong-5', 'tong-5', 'tong-5'] });
      game.addMeld({ type: 'kong_exposed', tiles: ['dong', 'dong', 'dong', 'dong'] });

      // 剩餘手牌：只需要2組面子 + 1對將 = 8張
      game.setPlayerTiles([
        'tiao-1', 'tiao-2', 'tiao-3',
        'wan-7', 'wan-8', 'wan-9',
        'zhong' // 聽中
      ]);

      const actions = game.checkSelfActions();

      expect(actions).toContain('ready');

      const readyTiles = game.checkReadyHand(game.players[0].tiles);
      expect(readyTiles).toContain('zhong');
    });

    it('應該在4組吃碰槓後偵測到聽牌（只差一對將）', () => {
      // 4組面子
      game.addMeld({ type: 'chow', tiles: ['wan-1', 'wan-2', 'wan-3'] });
      game.addMeld({ type: 'pong', tiles: ['tong-5', 'tong-5', 'tong-5'] });
      game.addMeld({ type: 'chow', tiles: ['tiao-7', 'tiao-8', 'tiao-9'] });
      game.addMeld({ type: 'pong', tiles: ['dong', 'dong', 'dong'] });

      // 剩餘手牌：還需要1組面子 + 1對將 = 5張（包含要胡的）
      // 所以手牌應該有4張
      game.setPlayerTiles([
        'nan', 'nan', 'nan', // 刻子
        'zhong' // 聽中做對子
      ]);

      const actions = game.checkSelfActions();

      expect(actions).toContain('ready');

      const readyTiles = game.checkReadyHand(game.players[0].tiles);
      expect(readyTiles).toContain('zhong');
    });
  });

  describe('使用者報告的實際案例', () => {
    it('案例：3張8條 + 1張9條，已有2組吃牌應該聽牌', () => {
      // 使用者的吃牌組合
      game.addMeld({ type: 'chow', tiles: ['wan-7', 'wan-8', 'wan-9'] });
      game.addMeld({ type: 'chow', tiles: ['wan-5', 'wan-6', 'wan-7'] });

      // 假設還有其他碰牌或吃牌（使用者手牌只剩4張說明有很多吃碰槓）
      game.addMeld({ type: 'pong', tiles: ['dong', 'dong', 'dong'] });

      // 剩餘手牌：還需要2組面子 + 1對將 = 8張（包含要胡的）
      // 所以手牌應該有7張
      game.setPlayerTiles([
        'tiao-8', 'tiao-8', 'tiao-8', // 刻子
        'tong-1', 'tong-2', 'tong-3', // 順子
        'zhong' // 聽中做對子
      ]);

      const actions = game.checkSelfActions();

      expect(actions).toContain('ready');

      const readyTiles = game.checkReadyHand(game.players[0].tiles);
      expect(readyTiles.length).toBeGreaterThan(0);
      expect(readyTiles).toContain('zhong');
    });
  });

  describe('邊界條件測試', () => {
    it('應該處理不是自己回合的情況', () => {
      game.currentTurn = 1; // 不是自己的回合
      game.setPlayerTiles([
        'wan-1', 'wan-2', 'wan-3',
        'dong'
      ]);

      const actions = game.checkSelfActions();

      expect(actions).toBeUndefined();
      expect(game.actionButtonsShown).toBe(false);
    });

    it('應該處理手牌為空的情況', () => {
      game.setPlayerTiles([]);

      const actions = game.checkSelfActions();

      expect(actions).not.toContain('hu');
      expect(actions).not.toContain('ready');
    });

    it('應該處理手牌不足的情況', () => {
      game.setPlayerTiles(['wan-1', 'wan-2']);

      const actions = game.checkSelfActions();

      expect(actions).not.toContain('hu');
      expect(actions).not.toContain('ready');
    });
  });

  describe('優先級測試', () => {
    it('自摸應該優先於聽牌', () => {
      // 設定一個已經胡牌的手牌（也可以說是聽牌）
      game.setPlayerTiles([
        'wan-1', 'wan-2', 'wan-3',
        'wan-4', 'wan-5', 'wan-6',
        'tong-1', 'tong-2', 'tong-3',
        'tiao-7', 'tiao-8', 'tiao-9',
        'tiao-1', 'tiao-2', 'tiao-3',
        'dong', 'dong'
      ]);

      const actions = game.checkSelfActions();

      // 應該顯示胡牌，而不是聽牌
      expect(actions).toContain('hu');
      expect(actions).not.toContain('ready');
    });
  });
});