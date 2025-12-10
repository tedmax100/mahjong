import { describe, it, expect, beforeEach, vi } from 'vitest';
import { Game } from './Game.js';

/**
 * 剩餘牌數功能測試 (Refactored to test Game.js)
 * 
 * Note: The client-side Game.js relies on the server to provide the absolute
 * number of remaining tiles via 'remainingTiles' field in actions.
 * It does NOT calculate the decrement locally for sync reasons, nor does it
 * trigger 'handleGameDraw' locally (that is a server event).
 */

// Mock PixiJS dependencies
vi.mock('pixi.js', () => ({
  Application: class {},
  Container: class {
    constructor() { this.children = []; this.addChild = () => {}; this.removeChildren = () => {}; }
  },
  Graphics: class {},
  Sprite: class {},
  Text: class {
    constructor() { this.text = ''; }
  },
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

describe('剩餘牌數功能測試 (Game.js)', () => {
  let game;

  beforeEach(() => {
    // Mock PIXI App
    const mockApp = {
      stage: { addChild: () => {} },
      screen: { width: 800, height: 600 }
    };

    game = new Game(mockApp, null);
    
    // Mock UI elements that are created in init() or createTable()
    // Since we are unit testing methods, we might need to manually setup wallText
    game.wallText = { text: '' };
    game.remainingTiles = 144;
  });

  describe('updateRemainingTiles', () => {
    it('應該更新 remainingTiles 屬性', () => {
      game.updateRemainingTiles(100);
      expect(game.remainingTiles).toBe(100);
    });

    it('應該更新 UI 文字', () => {
      game.updateRemainingTiles(50);
      expect(game.wallText.text).toContain('50');
    });

    it('應該正確處理 0 張', () => {
      game.updateRemainingTiles(0);
      expect(game.remainingTiles).toBe(0);
      expect(game.wallText.text).toContain('0');
    });
  });

  describe('handlePlayerAction 更新剩餘牌數', () => {
    it('應該根據 action data 中的 remainingTiles 更新狀態', () => {
      const data = {
        playerId: 'player1',
        action: 'draw',
        tile: 'wan-1',
        currentTurn: 0,
        remainingTiles: 70
      };
      
      // Mock players array to avoid errors in handlePlayerAction if it accesses players
      game.players = [
        { tiles: [], melds: [], container: { addChild: () => {} }, setPosition: () => {} },
        { tiles: [], melds: [], container: { addChild: () => {} }, setPosition: () => {} },
        { tiles: [], melds: [], container: { addChild: () => {} }, setPosition: () => {} },
        { tiles: [], melds: [], container: { addChild: () => {} }, setPosition: () => {} }
      ];
      // Mock getPlayerByServerId if needed, or just ensure players mapping works
      // handlePlayerAction might rely on index or id.
      // Let's assume we don't need full player setup just to check remainingTiles update 
      // IF the method updates it early.
      
      // We need to spy on updateRemainingTiles
      const spy = vi.spyOn(game, 'updateRemainingTiles');
      
      // Mock some internal methods called by handlePlayerAction to avoid crash
      game.getPlayerByServerId = () => game.players[0];
      game.performDrawAnimation = () => Promise.resolve();
      game.audioManager.playEffect = () => {};
      
      try {
        game.handlePlayerAction(data);
      } catch (e) {
        // Ignore errors related to animation/UI, just check if updateRemainingTiles was called
      }
      
      expect(spy).toHaveBeenCalledWith(70);
    });

    it('如果 data 中沒有 remainingTiles 則不應更新', () => {
      const initial = game.remainingTiles;
      const data = {
        playerId: 'player1',
        action: 'discard', // discard usually doesn't change count
        tile: 'wan-1',
        currentTurn: 1
        // no remainingTiles
      };

      // Mock players
      game.players = [
        { tiles: [], melds: [], container: { addChild: () => {} }, setPosition: () => {}, removeTile: () => {}, sortHand: () => {} },
        { tiles: [], melds: [], container: { addChild: () => {} }, setPosition: () => {} },
        { tiles: [], melds: [], container: { addChild: () => {} }, setPosition: () => {} },
        { tiles: [], melds: [], container: { addChild: () => {} }, setPosition: () => {} }
      ];

      // Mock internals
      game.getPlayerByServerId = () => game.players[0];
      game.performDiscardAnimation = () => Promise.resolve();

      try {
        game.handlePlayerAction(data);
      } catch (e) {
         // ignore
      }

      expect(game.remainingTiles).toBe(initial);
    });
  });

  describe('流局邏輯 (Client-side passive)', () => {
    it('客戶端不應主動觸發流局，而是等待 handleGameDraw', () => {
      // Mock handleGameDraw
      game.handleGameDraw = vi.fn();
      
      // Update to 8 (limit)
      game.updateRemainingTiles(8);
      
      // Should NOT call handleGameDraw automatically
      expect(game.handleGameDraw).not.toHaveBeenCalled();
    });
  });
});
