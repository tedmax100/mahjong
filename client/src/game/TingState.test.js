import { describe, it, expect, beforeEach, vi } from 'vitest';
import { Game } from './Game.js';
import { Player } from './Player.js';

// Mock PixiJS dependencies
vi.mock('pixi.js', () => ({
  Application: class {},
  Container: class {
    constructor() { this.children = []; this.addChild = () => {}; this.removeChildren = () => {}; this.destroy = () => {}; }
  },
  Graphics: class {
    constructor() { 
      this.clear = () => {}; 
      this.beginFill = () => {}; 
      this.drawRect = () => {}; 
      this.rect = () => {}; 
      this.endFill = () => {}; 
      this.fill = () => {}; 
      this.roundRect = () => {}; 
      this.stroke = () => {};
      this.addChild = () => {};
      this.removeChildren = () => {};
      this.destroy = () => {};
    }
  },
  Sprite: class {
    constructor() { this.anchor = { set: () => {} }; this.on = () => {}; }
    static from() { return new this(); }
  },
  Text: class {
    constructor() { this.anchor = { set: () => {} }; }
  },
  Assets: { load: () => Promise.resolve({}) },
  TextStyle: class {}
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

describe('聽牌狀態測試 (Real Game & Player)', () => {
  let player;
  let game;

  beforeEach(() => {
    const mockApp = {
      stage: { addChild: () => {} },
      screen: { width: 800, height: 600 }
    };
    game = new Game(mockApp, null);
    
    // We need to initialize players properly using Player class if we want to test Player methods
    // Game constructor initializes empty players array, then init/start methods fill it.
    // Let's manually add a Player instance.
    const mockUser = { id: 'p0', name: 'TestUser' };
    player = new Player(mockUser, 0, false);
    
    // Inject game reference if Player needs it (usually it doesn't, it emits events)
    // In Player.js: constructor(user, position, isMe)
    
    game.players[0] = player;
    game.myPosition = 0;
    
    // Mock actionButtons
    game.actionButtons = {
      show: vi.fn(),
      hide: vi.fn()
    };
  });

  describe('打牌限制 (Player.js)', () => {
    it('未聽牌時可以打任何牌', () => {
      player.isTing = false;
      player.lastDrawnTile = 'wan-1';
      player.isInteractive = true;

      const onDiscardSpy = vi.fn();
      player.onDiscard = onDiscardSpy;

      // Create a mock tile object as passed from UI
      const mockTile = { type: 'wan-2' };
      
      player.onTileClick(mockTile);
      
      expect(onDiscardSpy).toHaveBeenCalledWith('wan-2');
    });

    it('聽牌後只能打剛摸到的牌', () => {
      player.isTing = true;
      player.lastDrawnTile = 'wan-1';
      player.isInteractive = true;

      const onDiscardSpy = vi.fn();
      player.onDiscard = onDiscardSpy;

      // 嘗試打不是最後摸到的牌
      player.onTileClick({ type: 'wan-2' });
      expect(onDiscardSpy).not.toHaveBeenCalled();

      // 打剛摸到的牌
      player.onTileClick({ type: 'wan-1' });
      expect(onDiscardSpy).toHaveBeenCalledWith('wan-1');
    });
  });

  describe('吃碰槓限制 (Game.js checkPossibleActions)', () => {
    // Note: checkPossibleActions in Game.js (real) might be empty/server-driven.
    // In Game.js read earlier:
    // checkPossibleActions(tile, playerPosition) {
    //   console.log('checkPossibleActions 呼叫（由伺服器端處理）');
    // }
    // So the previous test was testing a Mock implementation that doesn't exist in Client Game.js anymore.
    // The Client Game.js relies on 'possible_actions' message from server.
    
    it('Client Game.js delegates action checking to server', () => {
        // verify calling it logs or does nothing
        const spy = vi.spyOn(console, 'log');
        game.checkPossibleActions('wan-1', 3);
        expect(spy).toHaveBeenCalledWith(expect.stringContaining('checkPossibleActions'));
    });
  });

  describe('自動打牌邏輯 (Game.js checkSelfActions)', () => {
      // Game.js checkSelfActions:
      // 1. checks if isTing -> returns
      // 2. sends check_ting
      // 3. checks kong
      
      // It does NOT implement 'auto_discard' return value locally.
      // So testing for 'auto_discard' is testing non-existent logic.
      
      it('聽牌後 checkSelfActions 應直接返回', () => {
          player.isTing = true;
          const spy = vi.spyOn(console, 'log');
          
          game.checkSelfActions();
          
          expect(spy).toHaveBeenCalledWith(expect.stringContaining('玩家已聽牌'));
      });
  });

  describe('聽牌狀態設定 (Player.js)', () => {
    it('初始狀態應該是未聽牌', () => {
      expect(player.isTing).toBe(false);
      expect(player.winningTiles).toEqual([]); 
    });

    it('設定聽牌後狀態應更新', () => {
      // Simulate server setting ting state
      player.isTing = true;
      player.winningTiles = ['wan-1', 'wan-4'];

      expect(player.isTing).toBe(true);
      expect(player.winningTiles).toContain('wan-1');
      expect(player.winningTiles).toContain('wan-4');
    });
  });
});
