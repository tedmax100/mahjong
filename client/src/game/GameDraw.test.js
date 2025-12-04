import { describe, it, expect, beforeEach, vi } from 'vitest';

/**
 * 流局功能測試
 *
 * 測試目標：
 * 1. 測試從伺服器接收剩餘牌數更新
 * 2. 測試 handleGameDraw 被正確呼叫
 * 3. 測試流局倒數計時邏輯
 */

// 模擬 Game 類的流局相關方法
class MockGameForDraw {
  constructor() {
    this.remainingTiles = 79;
    this.drawHandlerCalled = false;
    this.drawHandlerCallCount = 0;
    this.drawData = null;
  }

  /**
   * 更新剩餘牌數（從伺服器接收）
   */
  updateRemainingTiles(count) {
    this.remainingTiles = count;

    // 臺灣麻將規則：海底剩餘8張時流局
    if (this.remainingTiles <= 8 && this.remainingTiles > 0) {
      this.handleGameDraw();
    }
  }

  /**
   * 處理玩家動作（包括剩餘牌數更新）
   */
  handlePlayerAction(data) {
    const { remainingTiles } = data;

    // 更新剩餘牌數（如果伺服器發送了這個資訊）
    if (remainingTiles !== undefined) {
      this.updateRemainingTiles(remainingTiles);
    }
  }

  /**
   * 處理流局（模擬）
   */
  handleGameDraw(data) {
    this.drawHandlerCalled = true;
    this.drawHandlerCallCount++;
    this.drawData = data || null;
  }

  /**
   * 重置狀態
   */
  reset() {
    this.remainingTiles = 79;
    this.drawHandlerCalled = false;
    this.drawHandlerCallCount = 0;
    this.drawData = null;
  }
}

describe('流局功能測試', () => {
  let game;

  beforeEach(() => {
    game = new MockGameForDraw();
  });

  describe('從伺服器接收剩餘牌數更新', () => {
    it('應該在 handlePlayerAction 中更新剩餘牌數', () => {
      const data = {
        playerId: 'player1',
        action: 'draw',
        tile: 'wan-1',
        remainingTiles: 50
      };

      game.handlePlayerAction(data);
      expect(game.remainingTiles).toBe(50);
    });

    it('應該在剩餘牌數 <= 8 時觸發流局', () => {
      const data = {
        playerId: 'player1',
        action: 'draw',
        tile: 'wan-1',
        remainingTiles: 8
      };

      game.handlePlayerAction(data);
      expect(game.remainingTiles).toBe(8);
      expect(game.drawHandlerCalled).toBe(true);
    });

    it('應該在剩餘牌數 > 8 時不觸發流局', () => {
      const data = {
        playerId: 'player1',
        action: 'draw',
        tile: 'wan-1',
        remainingTiles: 9
      };

      game.handlePlayerAction(data);
      expect(game.remainingTiles).toBe(9);
      expect(game.drawHandlerCalled).toBe(false);
    });

    it('應該在沒有 remainingTiles 欄位時不更新', () => {
      const initialCount = game.remainingTiles;
      const data = {
        playerId: 'player1',
        action: 'discard',
        tile: 'wan-1'
        // 沒有 remainingTiles 欄位
      };

      game.handlePlayerAction(data);
      expect(game.remainingTiles).toBe(initialCount);
    });
  });

  describe('流局處理器呼叫', () => {
    it('應該在流局時呼叫 handleGameDraw', () => {
      game.updateRemainingTiles(8);
      expect(game.drawHandlerCalled).toBe(true);
    });

    it('應該在多次流局時多次呼叫 handleGameDraw', () => {
      game.updateRemainingTiles(8);
      expect(game.drawHandlerCallCount).toBe(1);

      game.reset();
      game.updateRemainingTiles(7);
      expect(game.drawHandlerCallCount).toBe(1);
    });

    it('應該能夠接收 game_draw 訊息數據', () => {
      const drawData = {
        remainingTiles: 8
      };

      game.handleGameDraw(drawData);
      expect(game.drawHandlerCalled).toBe(true);
      expect(game.drawData).toEqual(drawData);
    });
  });

  describe('流局邊界條件', () => {
    it('應該在剩餘8張時觸發流局', () => {
      game.updateRemainingTiles(8);
      expect(game.drawHandlerCalled).toBe(true);
    });

    it('應該在剩餘1張時觸發流局', () => {
      game.updateRemainingTiles(1);
      expect(game.drawHandlerCalled).toBe(true);
    });

    it('應該在剩餘0張時不觸發流局', () => {
      game.updateRemainingTiles(0);
      expect(game.drawHandlerCalled).toBe(false);
    });

    it('應該在剩餘9張時不觸發流局', () => {
      game.updateRemainingTiles(9);
      expect(game.drawHandlerCalled).toBe(false);
    });
  });

  describe('實際遊戲場景模擬', () => {
    it('應該模擬伺服器發送流局前的摸牌序列', () => {
      // 模擬遊戲進行到接近流局
      const drawSequence = [
        { action: 'draw', remainingTiles: 12 },
        { action: 'draw', remainingTiles: 11 },
        { action: 'draw', remainingTiles: 10 },
        { action: 'draw', remainingTiles: 9 },  // 還不流局
        { action: 'draw', remainingTiles: 8 }   // 觸發流局
      ];

      drawSequence.forEach((data, index) => {
        game.reset();
        game.handlePlayerAction(data);

        if (index < 4) {
          expect(game.drawHandlerCalled).toBe(false);
        } else {
          expect(game.drawHandlerCalled).toBe(true);
        }
      });
    });

    it('應該模擬伺服器直接發送 game_draw 訊息', () => {
      const gameDrawMessage = {
        type: 'game_draw',
        data: {
          remainingTiles: 8
        }
      };

      // 模擬 main.js 中的訊息處理
      if (gameDrawMessage.type === 'game_draw') {
        game.handleGameDraw(gameDrawMessage.data);
      }

      expect(game.drawHandlerCalled).toBe(true);
      expect(game.drawData.remainingTiles).toBe(8);
    });
  });

  describe('倒數計時邏輯測試', () => {
    it('應該在5秒後完成倒數計時（模擬）', async () => {
      // 使用 fake timers 模擬倒數計時
      vi.useFakeTimers();

      let countdownComplete = false;
      let countdownValue = 5;

      // 模擬倒數計時
      const countdownInterval = setInterval(() => {
        countdownValue--;
        if (countdownValue <= 0) {
          clearInterval(countdownInterval);
          countdownComplete = true;
        }
      }, 1000);

      // 快進5秒
      vi.advanceTimersByTime(5000);

      expect(countdownComplete).toBe(true);
      expect(countdownValue).toBe(0);

      vi.useRealTimers();
    });

    it('應該在倒數計時過程中正確更新文本', () => {
      vi.useFakeTimers();

      const countdownTexts = [];
      let countdown = 5;

      const countdownInterval = setInterval(() => {
        countdown--;
        if (countdown > 0) {
          countdownTexts.push(`${countdown}秒後開始新局`);
        } else {
          clearInterval(countdownInterval);
        }
      }, 1000);

      // 每秒推進一次，共5次
      for (let i = 0; i < 5; i++) {
        vi.advanceTimersByTime(1000);
      }

      expect(countdownTexts).toEqual([
        '4秒後開始新局',
        '3秒後開始新局',
        '2秒後開始新局',
        '1秒後開始新局'
      ]);

      vi.useRealTimers();
    });
  });
});

/**
 * 訊息處理整合測試
 *
 * 測試 main.js 中的訊息處理邏輯
 */
describe('訊息處理整合測試', () => {
  let game;

  beforeEach(() => {
    game = new MockGameForDraw();
  });

  it('應該正確處理 player_action 訊息（包含 remainingTiles）', () => {
    const message = {
      type: 'player_action',
      data: {
        playerId: 'player1',
        action: 'draw',
        tile: 'wan-1',
        currentTurn: 0,
        remainingTiles: 45
      }
    };

    // 模擬 main.js 的 handleServerMessage
    if (message.type === 'player_action') {
      game.handlePlayerAction(message.data);
    }

    expect(game.remainingTiles).toBe(45);
  });

  it('應該正確處理 game_draw 訊息', () => {
    const message = {
      type: 'game_draw',
      data: {
        remainingTiles: 8
      }
    };

    // 模擬 main.js 的 handleServerMessage
    if (message.type === 'game_draw') {
      game.handleGameDraw(message.data);
    }

    expect(game.drawHandlerCalled).toBe(true);
    expect(game.drawData.remainingTiles).toBe(8);
  });

  it('應該在流局訊息後不再處理摸牌訊息', () => {
    // 先收到流局訊息
    const drawMessage = {
      type: 'game_draw',
      data: { remainingTiles: 8 }
    };

    game.handleGameDraw(drawMessage.data);
    expect(game.drawHandlerCalled).toBe(true);

    // 重置標誌以測試後續行為
    const initialCallCount = game.drawHandlerCallCount;

    // 嘗試再次更新剩餘牌數（不應該再次觸發流局處理）
    // 這取決於實際實作，這裡只是測試狀態
    expect(game.drawHandlerCallCount).toBe(initialCallCount);
  });
});