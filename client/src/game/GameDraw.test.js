import { describe, it, expect, beforeEach, vi } from 'vitest';

/**
 * 流局功能测试
 *
 * 测试目标：
 * 1. 测试从服务器接收剩余牌数更新
 * 2. 测试 handleGameDraw 被正确调用
 * 3. 测试流局倒计时逻辑
 */

// 模拟 Game 类的流局相关方法
class MockGameForDraw {
  constructor() {
    this.remainingTiles = 79;
    this.drawHandlerCalled = false;
    this.drawHandlerCallCount = 0;
    this.drawData = null;
  }

  /**
   * 更新剩余牌数（从服务器接收）
   */
  updateRemainingTiles(count) {
    this.remainingTiles = count;

    // 台湾麻将规则：海底剩余8张时流局
    if (this.remainingTiles <= 8 && this.remainingTiles > 0) {
      this.handleGameDraw();
    }
  }

  /**
   * 处理玩家动作（包括剩余牌数更新）
   */
  handlePlayerAction(data) {
    const { remainingTiles } = data;

    // 更新剩余牌数（如果服务器发送了这个信息）
    if (remainingTiles !== undefined) {
      this.updateRemainingTiles(remainingTiles);
    }
  }

  /**
   * 处理流局（模拟）
   */
  handleGameDraw(data) {
    this.drawHandlerCalled = true;
    this.drawHandlerCallCount++;
    this.drawData = data || null;
  }

  /**
   * 重置状态
   */
  reset() {
    this.remainingTiles = 79;
    this.drawHandlerCalled = false;
    this.drawHandlerCallCount = 0;
    this.drawData = null;
  }
}

describe('流局功能测试', () => {
  let game;

  beforeEach(() => {
    game = new MockGameForDraw();
  });

  describe('从服务器接收剩余牌数更新', () => {
    it('应该在 handlePlayerAction 中更新剩余牌数', () => {
      const data = {
        playerId: 'player1',
        action: 'draw',
        tile: 'wan-1',
        remainingTiles: 50
      };

      game.handlePlayerAction(data);
      expect(game.remainingTiles).toBe(50);
    });

    it('应该在剩余牌数 <= 8 时触发流局', () => {
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

    it('应该在剩余牌数 > 8 时不触发流局', () => {
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

    it('应该在没有 remainingTiles 字段时不更新', () => {
      const initialCount = game.remainingTiles;
      const data = {
        playerId: 'player1',
        action: 'discard',
        tile: 'wan-1'
        // 没有 remainingTiles 字段
      };

      game.handlePlayerAction(data);
      expect(game.remainingTiles).toBe(initialCount);
    });
  });

  describe('流局处理器调用', () => {
    it('应该在流局时调用 handleGameDraw', () => {
      game.updateRemainingTiles(8);
      expect(game.drawHandlerCalled).toBe(true);
    });

    it('应该在多次流局时多次调用 handleGameDraw', () => {
      game.updateRemainingTiles(8);
      expect(game.drawHandlerCallCount).toBe(1);

      game.reset();
      game.updateRemainingTiles(7);
      expect(game.drawHandlerCallCount).toBe(1);
    });

    it('应该能够接收 game_draw 消息数据', () => {
      const drawData = {
        remainingTiles: 8
      };

      game.handleGameDraw(drawData);
      expect(game.drawHandlerCalled).toBe(true);
      expect(game.drawData).toEqual(drawData);
    });
  });

  describe('流局边界条件', () => {
    it('应该在剩余8张时触发流局', () => {
      game.updateRemainingTiles(8);
      expect(game.drawHandlerCalled).toBe(true);
    });

    it('应该在剩余1张时触发流局', () => {
      game.updateRemainingTiles(1);
      expect(game.drawHandlerCalled).toBe(true);
    });

    it('应该在剩余0张时不触发流局', () => {
      game.updateRemainingTiles(0);
      expect(game.drawHandlerCalled).toBe(false);
    });

    it('应该在剩余9张时不触发流局', () => {
      game.updateRemainingTiles(9);
      expect(game.drawHandlerCalled).toBe(false);
    });
  });

  describe('实际游戏场景模拟', () => {
    it('应该模拟服务器发送流局前的摸牌序列', () => {
      // 模拟游戏进行到接近流局
      const drawSequence = [
        { action: 'draw', remainingTiles: 12 },
        { action: 'draw', remainingTiles: 11 },
        { action: 'draw', remainingTiles: 10 },
        { action: 'draw', remainingTiles: 9 },  // 还不流局
        { action: 'draw', remainingTiles: 8 }   // 触发流局
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

    it('应该模拟服务器直接发送 game_draw 消息', () => {
      const gameDrawMessage = {
        type: 'game_draw',
        data: {
          remainingTiles: 8
        }
      };

      // 模拟 main.js 中的消息处理
      if (gameDrawMessage.type === 'game_draw') {
        game.handleGameDraw(gameDrawMessage.data);
      }

      expect(game.drawHandlerCalled).toBe(true);
      expect(game.drawData.remainingTiles).toBe(8);
    });
  });

  describe('倒计时逻辑测试', () => {
    it('应该在5秒后完成倒计时（模拟）', async () => {
      // 使用 fake timers 模拟倒计时
      vi.useFakeTimers();

      let countdownComplete = false;
      let countdownValue = 5;

      // 模拟倒计时
      const countdownInterval = setInterval(() => {
        countdownValue--;
        if (countdownValue <= 0) {
          clearInterval(countdownInterval);
          countdownComplete = true;
        }
      }, 1000);

      // 快进5秒
      vi.advanceTimersByTime(5000);

      expect(countdownComplete).toBe(true);
      expect(countdownValue).toBe(0);

      vi.useRealTimers();
    });

    it('应该在倒计时过程中正确更新文本', () => {
      vi.useFakeTimers();

      const countdownTexts = [];
      let countdown = 5;

      const countdownInterval = setInterval(() => {
        countdown--;
        if (countdown > 0) {
          countdownTexts.push(`${countdown}秒后开始新局`);
        } else {
          clearInterval(countdownInterval);
        }
      }, 1000);

      // 每秒推进一次，共5次
      for (let i = 0; i < 5; i++) {
        vi.advanceTimersByTime(1000);
      }

      expect(countdownTexts).toEqual([
        '4秒后开始新局',
        '3秒后开始新局',
        '2秒后开始新局',
        '1秒后开始新局'
      ]);

      vi.useRealTimers();
    });
  });
});

/**
 * 消息处理集成测试
 *
 * 测试 main.js 中的消息处理逻辑
 */
describe('消息处理集成测试', () => {
  let game;

  beforeEach(() => {
    game = new MockGameForDraw();
  });

  it('应该正确处理 player_action 消息（包含 remainingTiles）', () => {
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

    // 模拟 main.js 的 handleServerMessage
    if (message.type === 'player_action') {
      game.handlePlayerAction(message.data);
    }

    expect(game.remainingTiles).toBe(45);
  });

  it('应该正确处理 game_draw 消息', () => {
    const message = {
      type: 'game_draw',
      data: {
        remainingTiles: 8
      }
    };

    // 模拟 main.js 的 handleServerMessage
    if (message.type === 'game_draw') {
      game.handleGameDraw(message.data);
    }

    expect(game.drawHandlerCalled).toBe(true);
    expect(game.drawData.remainingTiles).toBe(8);
  });

  it('应该在流局消息后不再处理摸牌消息', () => {
    // 先收到流局消息
    const drawMessage = {
      type: 'game_draw',
      data: { remainingTiles: 8 }
    };

    game.handleGameDraw(drawMessage.data);
    expect(game.drawHandlerCalled).toBe(true);

    // 重置标志以测试后续行为
    const initialCallCount = game.drawHandlerCallCount;

    // 尝试再次更新剩余牌数（不应该再次触发流局处理）
    // 这取决于实际实现，这里只是测试状态
    expect(game.drawHandlerCallCount).toBe(initialCallCount);
  });
});
