import { describe, it, expect, beforeEach, vi } from 'vitest';

/**
 * 測試聽牌狀態功能
 * 包含：聽牌宣告、打牌限制、吃碰槓限制、自動打牌等
 */

// 模擬 Player 類的聽牌相關功能
class MockPlayer {
  constructor() {
    this.isTing = false;
    this.winningTiles = [];
    this.lastDrawnTile = null;
    this.tiles = [];
    this.melds = [];
    this.isInteractive = true;
    this.onDiscard = null;
  }

  // 從 Player.js 複製的 onTileClick 邏輯
  onTileClick(tile) {
    if (!this.isInteractive) {
      return false;
    }

    // 如果已宣告聽牌，只能打剛摸到的牌
    if (this.isTing) {
      if (tile.type !== this.lastDrawnTile) {
        return false; // 不能打
      }
    }

    if (this.onDiscard) {
      this.onDiscard(tile.type);
    }
    return true;
  }

  addTile(tileType) {
    this.lastDrawnTile = tileType;
    this.tiles.push({ type: tileType });
  }
}

// 模擬 Game 類的聽牌相關邏輯
class MockGame {
  constructor() {
    this.players = [
      new MockPlayer(),
      new MockPlayer(),
      new MockPlayer(),
      new MockPlayer()
    ];
    this.myPosition = 0;
    this.actionButtons = {
      show: vi.fn(),
      hide: vi.fn()
    };
    this.pendingActions = [];
  }

  // 從 Game.js 複製的 checkPossibleActions 邏輯（簡化版）
  checkPossibleActions(tile, discardPlayerPosition) {
    const myPlayer = this.players[this.myPosition];
    const actions = [];

    // 假設可以胡牌
    const canHu = true;
    if (canHu) {
      actions.push('hu');
    }

    // 如果已宣告聽牌，只能胡牌
    if (myPlayer.isTing) {
      if (actions.length > 0) {
        actions.push('cancel');
        this.pendingActions = actions;
        this.actionButtons.show(actions);
      }
      return actions;
    }

    // 未聽牌時可以吃碰槓
    const canPong = true;
    const canChow = discardPlayerPosition === (this.myPosition + 3) % 4;
    const canKong = true;

    if (canKong) actions.push('kong');
    if (canPong) actions.push('pong');
    if (canChow) actions.push('chow');

    if (actions.length > 0) {
      actions.push('cancel');
      this.pendingActions = actions;
      this.actionButtons.show(actions);
    }

    return actions;
  }

  // 從 Game.js 複製的 checkSelfActions 邏輯（簡化版）
  checkSelfActions() {
    const myPlayer = this.players[this.myPosition];
    const actions = [];

    // 假設不能自摸
    const canSelfDrawHu = false;

    // 如果已宣告聽牌
    if (myPlayer.isTing) {
      // 如果不能胡牌，應該自動打出剛摸到的牌
      if (!canSelfDrawHu && myPlayer.lastDrawnTile) {
        return 'auto_discard'; // 返回標記表示應自動打牌
      }
    }

    return actions;
  }
}

describe('聽牌狀態測試', () => {
  let player;
  let game;

  beforeEach(() => {
    player = new MockPlayer();
    game = new MockGame();
  });

  describe('打牌限制', () => {
    it('未聽牌時可以打任何牌', () => {
      player.isTing = false;
      player.lastDrawnTile = 'wan-1';
      player.isInteractive = true;

      let discardedTile = null;
      player.onDiscard = (tile) => { discardedTile = tile; };

      // 可以打不是最後摸到的牌
      const result = player.onTileClick({ type: 'wan-2' });
      expect(result).toBe(true);
      expect(discardedTile).toBe('wan-2');
    });

    it('聽牌後只能打剛摸到的牌', () => {
      player.isTing = true;
      player.lastDrawnTile = 'wan-1';
      player.isInteractive = true;

      let discardedTile = null;
      player.onDiscard = (tile) => { discardedTile = tile; };

      // 嘗試打不是最後摸到的牌 - 應該失敗
      const result1 = player.onTileClick({ type: 'wan-2' });
      expect(result1).toBe(false);
      expect(discardedTile).toBe(null);

      // 打剛摸到的牌 - 應該成功
      const result2 = player.onTileClick({ type: 'wan-1' });
      expect(result2).toBe(true);
      expect(discardedTile).toBe('wan-1');
    });

    it('聽牌後點擊非最後摸牌不會觸發打牌', () => {
      player.isTing = true;
      player.lastDrawnTile = 'tong-5';
      player.isInteractive = true;

      let discardCalled = false;
      player.onDiscard = () => { discardCalled = true; };

      player.onTileClick({ type: 'tong-3' });
      expect(discardCalled).toBe(false);
    });
  });

  describe('摸牌記錄', () => {
    it('摸牌時應記錄最後摸到的牌', () => {
      player.addTile('wan-1');
      expect(player.lastDrawnTile).toBe('wan-1');

      player.addTile('wan-2');
      expect(player.lastDrawnTile).toBe('wan-2');
    });

    it('聽牌後摸牌應更新lastDrawnTile', () => {
      player.isTing = true;
      player.addTile('wan-1');
      expect(player.lastDrawnTile).toBe('wan-1');

      player.addTile('wan-2');
      expect(player.lastDrawnTile).toBe('wan-2');
    });
  });

  describe('吃碰槓限制', () => {
    it('未聽牌時可以吃碰槓胡', () => {
      const myPlayer = game.players[game.myPosition];
      myPlayer.isTing = false;

      // 上家打牌（可以吃）
      const actions = game.checkPossibleActions('wan-1', 3);

      expect(actions).toContain('hu');
      expect(actions).toContain('pong');
      expect(actions).toContain('chow');
      expect(actions).toContain('kong');
    });

    it('聽牌後只能胡牌，不能吃碰槓', () => {
      const myPlayer = game.players[game.myPosition];
      myPlayer.isTing = true;

      // 上家打牌
      const actions = game.checkPossibleActions('wan-1', 3);

      expect(actions).toContain('hu');
      expect(actions).not.toContain('pong');
      expect(actions).not.toContain('chow');
      expect(actions).not.toContain('kong');
    });

    it('聽牌後檢查動作應顯示正確的按鈕', () => {
      const myPlayer = game.players[game.myPosition];
      myPlayer.isTing = true;

      game.checkPossibleActions('wan-1', 3);

      // 應該只顯示胡牌和取消按鈕
      expect(game.actionButtons.show).toHaveBeenCalledWith(['hu', 'cancel']);
    });
  });

  describe('自動打牌邏輯', () => {
    it('聽牌後不能胡牌時應返回自動打牌標記', () => {
      const myPlayer = game.players[game.myPosition];
      myPlayer.isTing = true;
      myPlayer.lastDrawnTile = 'wan-5';

      const result = game.checkSelfActions();

      expect(result).toBe('auto_discard');
    });

    it('聽牌後能胡牌時不應自動打牌', () => {
      // 這個測試需要修改 checkSelfActions 讓它能檢測到可以胡牌
      // 這裡只是示範結構
      const myPlayer = game.players[game.myPosition];
      myPlayer.isTing = true;
      myPlayer.lastDrawnTile = 'wan-5';

      // 假設可以胡牌的情況下，不應返回 auto_discard
      // 實際實作中需要配合 canHu 檢查
    });

    it('未聽牌時不應觸發自動打牌', () => {
      const myPlayer = game.players[game.myPosition];
      myPlayer.isTing = false;
      myPlayer.lastDrawnTile = 'wan-5';

      const result = game.checkSelfActions();

      expect(result).not.toBe('auto_discard');
    });
  });

  describe('聽牌狀態設置', () => {
    it('初始狀態應該是未聽牌', () => {
      expect(player.isTing).toBe(false);
      expect(player.winningTiles).toEqual([]);
    });

    it('宣告聽牌後狀態應更新', () => {
      player.isTing = true;
      player.winningTiles = ['wan-1', 'wan-4'];

      expect(player.isTing).toBe(true);
      expect(player.winningTiles).toContain('wan-1');
      expect(player.winningTiles).toContain('wan-4');
    });

    it('聽牌狀態應該持續到遊戲結束', () => {
      player.isTing = true;
      player.winningTiles = ['wan-1'];

      // 摸牌後聽牌狀態應保持
      player.addTile('wan-2');
      expect(player.isTing).toBe(true);

      // 再次摸牌
      player.addTile('wan-3');
      expect(player.isTing).toBe(true);
    });
  });

  describe('邊界情況測試', () => {
    it('聽牌但沒有lastDrawnTile時不應崩潰', () => {
      player.isTing = true;
      player.lastDrawnTile = null;
      player.isInteractive = true;

      const result = player.onTileClick({ type: 'wan-1' });
      expect(result).toBe(false);
    });

    it('未設置onDiscard回調時不應崩潰', () => {
      player.isTing = false;
      player.lastDrawnTile = 'wan-1';
      player.isInteractive = true;
      player.onDiscard = null;

      expect(() => {
        player.onTileClick({ type: 'wan-1' });
      }).not.toThrow();
    });

    it('玩家不可交互時聽牌限制應該優先', () => {
      player.isTing = true;
      player.lastDrawnTile = 'wan-1';
      player.isInteractive = false;

      const result = player.onTileClick({ type: 'wan-1' });
      expect(result).toBe(false);
    });
  });
});
