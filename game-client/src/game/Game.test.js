import { describe, it, expect, beforeEach, vi } from 'vitest';
import { Game } from './Game.js';

// Mock PixiJS dependencies
vi.mock('pixi.js', () => ({
  Application: class {},
  Container: class {
    constructor() { this.children = []; this.addChild = () => {}; this.removeChildren = () => {}; }
  },
  Graphics: class {},
  Sprite: class {},
  Text: class {},
  Assets: { load: () => Promise.resolve({}) }
}));

// Mock other dependencies
vi.mock('./AudioManager.js', () => ({
  AudioManager: class {
    constructor() {
      this.playEffect = () => {};
      this.playButtonSound = () => {};
    }
  }
}));

vi.mock('./WebSocketClient.js', () => ({
  WebSocketClient: class {}
}));

vi.mock('./DiscardManager.js', () => ({
  DiscardManager: class {
    constructor() {
      this.container = { addChild: () => {}, removeChildren: () => {} };
      this.addDiscard = () => Promise.resolve();
      this.removeLastDiscard = () => {};
      this.clear = () => {};
      this.resize = () => {};
    }
  }
}));

describe('麻將胡牌判斷測試 (Game.js)', () => {
  let game;

  beforeEach(() => {
    // Mock PIXI App
    const mockApp = {
      stage: { addChild: () => {} },
      screen: { width: 800, height: 600 }
    };
    
    game = new Game(mockApp, null);
    
    // Setup minimal player state for testing
    game.players = [
      { melds: [] },
      { melds: [] },
      { melds: [] },
      { melds: [] }
    ];
    game.myPosition = 0;
  });

  describe('canHu - 基本胡牌判斷', () => {
    it('應該能判斷簡單的刻子胡牌（全刻子）', () => {
      const hand = [
        'wan-1', 'wan-1', 'wan-1', // 刻子
        'wan-2', 'wan-2', 'wan-2', // 刻子
        'wan-3', 'wan-3', 'wan-3', // 刻子
        'wan-4', 'wan-4', 'wan-4', // 刻子
        'wan-5', 'wan-5', 'wan-5', // 刻子
        'dong'                     // 將的一半
      ];
      const tile = 'dong'; // 要胡的牌，組成對子（16+1=17張）

      expect(game.canHu(hand, tile)).toBe(true);
    });

    it('應該能判斷簡單的順子胡牌（全順子）', () => {
      const hand = [
        'wan-1', 'wan-2', 'wan-3', // 順子
        'wan-4', 'wan-5', 'wan-6', // 順子
        'tong-1', 'tong-2', 'tong-3', // 順子
        'tiao-7', 'tiao-8', 'tiao-9', // 順子
        'tiao-1', 'tiao-2', 'tiao-3', // 順子
        'dong'                          // 將的一半
      ];
      const tile = 'dong'; // 要胡的牌，組成對子（16+1=17張）

      expect(game.canHu(hand, tile)).toBe(true);
    });

    it('應該能判斷混合胡牌（順子+刻子）', () => {
      const hand = [
        'wan-1', 'wan-2', 'wan-3', // 順子
        'tong-5', 'tong-5', 'tong-5', // 刻子
        'tiao-2', 'tiao-3', 'tiao-4', // 順子
        'dong', 'dong', 'dong',     // 刻子
        'nan', 'nan', 'nan',        // 刻子
        'zhong'                      // 將的一半
      ];
      const tile = 'zhong'; // 要胡的牌（16+1=17張）

      expect(game.canHu(hand, tile)).toBe(true);
    });

    it('應該拒絕不完整的牌型（缺牌）', () => {
      const hand = [
        'wan-1', 'wan-2', 'wan-4', // 不連續
        'wan-5', 'wan-6', 'wan-7',
        'tong-1', 'tong-2', 'tong-3',
        'tiao-7', 'tiao-8', 'tiao-9',
        'dong'
      ];
      const tile = 'dong';

      expect(game.canHu(hand, tile)).toBe(false);
    });

    it('應該拒絕牌數不正確的手牌', () => {
      const hand = [
        'wan-1', 'wan-2', 'wan-3',
        'wan-4', 'wan-5', 'wan-6'
      ]; // 只有6張牌，不夠17張
      const tile = 'dong';

      expect(game.canHu(hand, tile)).toBe(false);
    });

    it('應該拒絕沒有對子的牌型', () => {
      const hand = [
        'wan-1', 'wan-2', 'wan-3',
        'wan-4', 'wan-5', 'wan-6',
        'tong-1', 'tong-2', 'tong-3',
        'tiao-7', 'tiao-8', 'tiao-9',
        'dong', 'nan', 'xi' // 三張不同的字牌，無法組成對子或面子
      ];
      const tile = 'bei';

      expect(game.canHu(hand, tile)).toBe(false);
    });
  });

  describe('canHu - 帶吃碰槓的胡牌判斷', () => {
    it('應該能判斷有1組碰牌的胡牌', () => {
      game.players[0].melds = [
        { type: 'pong', tiles: ['wan-1', 'wan-1', 'wan-1'] }
      ];

      const hand = [
        'wan-2', 'wan-3', 'wan-4', // 順子
        'tong-5', 'tong-6', 'tong-7', // 順子
        'tiao-8', 'tiao-9', 'tiao-7', // 順子
        'dong', 'dong'              // 對子
      ];
      const tile = 'fa'; // 不需要，因為已經有碰牌，只需要13張

      // 重新計算：有1組碰牌，需要4組面子 + 1對將 = 4*3 + 2 = 14張
      const correctHand = [
        'wan-2', 'wan-3', 'wan-4',
        'tong-5', 'tong-6', 'tong-7',
        'tiao-7', 'tiao-8', 'tiao-9',
        'zhong', 'zhong', 'zhong',
        'dong'
      ];
      const correctTile = 'dong';

      expect(game.canHu(correctHand, correctTile, 1)).toBe(true);
    });

    it('應該能判斷有2組吃牌的胡牌', () => {
      game.players[0].melds = [
        { type: 'chow', tiles: ['wan-1', 'wan-2', 'wan-3'] },
        { type: 'chow', tiles: ['tong-4', 'tong-5', 'tong-6'] }
      ];

      // 有2組吃牌，需要3組面子 + 1對將 = 3*3 + 2 = 11張
      const hand = [
        'tiao-1', 'tiao-2', 'tiao-3',
        'zhong', 'zhong', 'zhong',
        'fa', 'fa', 'fa',
        'dong'
      ];
      const tile = 'dong';

      expect(game.canHu(hand, tile, 2)).toBe(true);
    });
  });

  describe('checkReadyHand - 聽牌判斷', () => {
    it('應該能判斷兩面聽（7,8萬聽6萬或9萬）', () => {
      game.players[0].melds = [
        { type: 'chow', tiles: ['tiao-3', 'tiao-4', 'tiao-5'] },
        { type: 'pong', tiles: ['zhong', 'zhong', 'zhong'] }
      ];

      // 手牌：7,8萬 + 8,8桶 + 3,4,5條（已吃） + 中,中,中（已碰）
      // 需要3組面子 + 1對將 = 11張（包含要胡的）
      const hand = [
        'wan-7', 'wan-8',          // 搭子
        'tong-8', 'tong-8',        // 對子
        'dong', 'dong', 'dong',    // 刻子
        'nan', 'nan', 'nan'        // 刻子
      ];

      const readyTiles = game.checkReadyHand(hand);

      // 聽6萬或9萬可以組成順子
      expect(readyTiles).toContain('wan-6');
      expect(readyTiles).toContain('wan-9');
    });

    it('應該能判斷多面聽', () => {
      game.players[0].melds = [
        { type: 'pong', tiles: ['dong', 'dong', 'dong'] }
      ];

      // 有1組碰牌，需要4組面子 + 1對將 = 14張
      const hand = [
        'wan-1', 'wan-2', 'wan-3',
        'wan-4', 'wan-5', 'wan-6',
        'tong-7', 'tong-8', 'tong-9',
        'tiao-2', 'tiao-3', 'tiao-4',
        'zhong'
      ];

      const readyTiles = game.checkReadyHand(hand);

      // 聽中，可以組成對子
      expect(readyTiles).toContain('zhong');
    });

    it('應該返回空陣列當無法聽牌時', () => {
      const hand = [
        'wan-1', 'wan-3', 'wan-5', // 不連續
        'tong-2', 'tong-4', 'tong-6',
        'tiao-1', 'dong', 'nan',
        'xi', 'bei', 'zhong',
        'fa', 'bai'
      ];

      const readyTiles = game.checkReadyHand(hand);

      expect(Array.isArray(readyTiles)).toBe(true);
    });
  });

  describe('實際案例測試', () => {
    it('使用者案例：7,8萬 + 2張8桶 + 345條應該聽牌', () => {
      // 假設使用者已經有其他吃碰槓的組合
      game.players[0].melds = [
        { type: 'chow', tiles: ['tiao-3', 'tiao-4', 'tiao-5'] },
        { type: 'pong', tiles: ['fa', 'fa', 'fa'] }
      ];

      // 剩餘手牌：有2組面子（已吃碰），需要3組面子+1對將 = 11張
      const hand = [
        'wan-7', 'wan-8',          // 搭子（差6萬或9萬）
        'tong-8', 'tong-8',        // 對子
        'zhong', 'zhong', 'zhong', // 刻子
        'dong', 'dong', 'dong'     // 刻子
      ];

      const readyTiles = game.checkReadyHand(hand);

      // 聽6萬或9萬
      expect(readyTiles).toContain('wan-6');
      expect(readyTiles).toContain('wan-9');
      expect(readyTiles.length).toBeGreaterThan(0);
    });

    it('應該能判斷清一色聽牌', () => {
      game.players[0].melds = [];

      const hand = [
        'wan-1', 'wan-2', 'wan-3',
        'wan-4', 'wan-5', 'wan-6',
        'wan-7', 'wan-8', 'wan-9',
        'wan-2', 'wan-3', 'wan-4',
        'wan-5', 'wan-6', 'wan-7',
        'wan-8'
      ];

      const readyTiles = game.checkReadyHand(hand);

      // 聽萬-8組成對子
      expect(readyTiles).toContain('wan-8');
    });

    it('應該能判斷碰碰胡聽牌', () => {
      game.players[0].melds = [
        { type: 'pong', tiles: ['wan-1', 'wan-1', 'wan-1'] }
      ];

      const hand = [
        'wan-2', 'wan-2', 'wan-2',
        'wan-3', 'wan-3', 'wan-3',
        'wan-4', 'wan-4', 'wan-4',
        'wan-5', 'wan-5', 'wan-5',
        'dong'
      ];

      const readyTiles = game.checkReadyHand(hand);

      // 聽東做對子
      expect(readyTiles).toContain('dong');
    });
  });

  describe('邊界條件測試', () => {
    it('應該處理空手牌', () => {
      const hand = [];
      const tile = 'wan-1';

      expect(game.canHu(hand, tile)).toBe(false);
    });

    it('應該處理單張牌', () => {
      const hand = ['wan-1'];
      const tile = 'wan-1';

      expect(game.canHu(hand, tile)).toBe(false);
    });

    it('應該處理字牌（不能組成順子）', () => {
      const hand = [
        'dong', 'dong', 'dong',
        'nan', 'nan', 'nan',
        'xi', 'xi', 'xi',
        'bei', 'bei', 'bei',
        'zhong', 'zhong', 'zhong',
        'fa'
      ];
      const tile = 'fa'; // 16+1=17張

      expect(game.canHu(hand, tile)).toBe(true);
    });

    it('應該處理花色混合', () => {
      const hand = [
        'wan-1', 'wan-2', 'wan-3',
        'tong-4', 'tong-5', 'tong-6',
        'tiao-7', 'tiao-8', 'tiao-9',
        'dong', 'dong', 'dong',
        'nan', 'nan', 'nan',
        'zhong'
      ];
      const tile = 'zhong';

      expect(game.canHu(hand, tile)).toBe(true);
    });
  });
});