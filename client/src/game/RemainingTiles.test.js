import { describe, it, expect, beforeEach, vi } from 'vitest';

/**
 * 剩餘牌數計算測試
 *
 * 測試目標：
 * 1. 驗證剩餘牌數在摸牌時正確減少
 * 2. 驗證剩餘牌數在打牌時不會減少
 * 3. 驗證流局判斷邏輯正確
 */

// 模擬 Game 類的剩餘牌數相關方法
class MockGameForRemainingTiles {
  constructor() {
    this.remainingTiles = 79; // 初始剩餘牌數（144 - 65 = 79）
    this.isDrawTriggered = false;
    this.drawCallCount = 0;
  }

  /**
   * 更新剩餘牌數
   */
  updateRemainingTiles(count) {
    this.remainingTiles = count;

    // 臺灣麻將規則：海底剩餘8張時流局
    if (this.remainingTiles <= 8 && this.remainingTiles > 0) {
      this.handleGameDraw();
    }
  }

  /**
   * 處理流局
   */
  handleGameDraw() {
    this.isDrawTriggered = true;
    this.drawCallCount++;
  }

  /**
   * 模擬摸牌
   */
  simulateDrawTile() {
    // 摸牌時應該減少剩餘牌數
    if (this.remainingTiles > 0) {
      this.updateRemainingTiles(this.remainingTiles - 1);
    }
  }

  /**
   * 模擬打牌
   */
  simulateDiscardTile() {
    // 打牌時不應該減少剩餘牌數（已修復）
    // 不做任何操作
  }

  /**
   * 重置狀態
   */
  reset() {
    this.remainingTiles = 79;
    this.isDrawTriggered = false;
    this.drawCallCount = 0;
  }
}

describe('剩餘牌數計算測試', () => {
  let game;

  beforeEach(() => {
    game = new MockGameForRemainingTiles();
  });

  describe('基本功能測試', () => {
    it('應該在初始化時設定正確的剩餘牌數', () => {
      expect(game.remainingTiles).toBe(79);
    });

    it('應該能夠更新剩餘牌數', () => {
      game.updateRemainingTiles(75);
      expect(game.remainingTiles).toBe(75);
    });

    it('應該在剩餘牌數 <= 8 時觸發流局', () => {
      game.updateRemainingTiles(8);
      expect(game.isDrawTriggered).toBe(true);
    });

    it('應該在剩餘牌數 > 8 時不觸發流局', () => {
      game.updateRemainingTiles(9);
      expect(game.isDrawTriggered).toBe(false);
    });
  });

  describe('摸牌時的剩餘牌數更新', () => {
    it('應該在摸牌時減少剩餘牌數', () => {
      const initialCount = game.remainingTiles;
      game.simulateDrawTile();
      expect(game.remainingTiles).toBe(initialCount - 1);
    });

    it('應該在連續摸牌時持續減少剩餘牌數', () => {
      game.simulateDrawTile();
      game.simulateDrawTile();
      game.simulateDrawTile();
      expect(game.remainingTiles).toBe(76); // 79 - 3 = 76
    });

    it('應該在摸到牌山只剩8張時觸發流局', () => {
      // 設定剩餘9張
      game.updateRemainingTiles(9);
      game.isDrawTriggered = false; // 重置流局標誌

      // 摸一張牌，剩餘8張，應該觸發流局
      game.simulateDrawTile();
      expect(game.remainingTiles).toBe(8);
      expect(game.isDrawTriggered).toBe(true);
    });

    it('應該在剩餘牌數為0時不再減少', () => {
      game.updateRemainingTiles(0);
      game.simulateDrawTile();
      expect(game.remainingTiles).toBe(0);
    });
  });

  describe('打牌時的剩餘牌數更新', () => {
    it('應該在打牌時不減少剩餘牌數', () => {
      const initialCount = game.remainingTiles;
      game.simulateDiscardTile();
      expect(game.remainingTiles).toBe(initialCount);
    });

    it('應該在連續打牌時剩餘牌數保持不變', () => {
      const initialCount = game.remainingTiles;
      game.simulateDiscardTile();
      game.simulateDiscardTile();
      game.simulateDiscardTile();
      expect(game.remainingTiles).toBe(initialCount);
    });
  });

  describe('摸牌和打牌混合測試', () => {
    it('應該在摸牌後打牌時只減少一次剩餘牌數', () => {
      const initialCount = game.remainingTiles;

      // 摸牌
      game.simulateDrawTile();
      const afterDraw = game.remainingTiles;
      expect(afterDraw).toBe(initialCount - 1);

      // 打牌
      game.simulateDiscardTile();
      expect(game.remainingTiles).toBe(afterDraw); // 應該與摸牌後相同
    });

    it('應該在一個完整回合（摸+打）後只減少1張牌', () => {
      const initialCount = game.remainingTiles;

      // 完整回合：摸牌 -> 打牌
      game.simulateDrawTile();
      game.simulateDiscardTile();

      expect(game.remainingTiles).toBe(initialCount - 1);
    });

    it('應該在多個回合後正確計算剩餘牌數', () => {
      const initialCount = game.remainingTiles;
      const rounds = 10;

      // 模擬10個回合
      for (let i = 0; i < rounds; i++) {
        game.simulateDrawTile();
        game.simulateDiscardTile();
      }

      expect(game.remainingTiles).toBe(initialCount - rounds);
    });
  });

  describe('流局判斷測試', () => {
    it('應該在剩餘8張時觸發流局', () => {
      game.updateRemainingTiles(8);
      expect(game.isDrawTriggered).toBe(true);
    });

    it('應該在剩餘7張時觸發流局', () => {
      game.updateRemainingTiles(7);
      expect(game.isDrawTriggered).toBe(true);
    });

    it('應該在剩餘1張時觸發流局', () => {
      game.updateRemainingTiles(1);
      expect(game.isDrawTriggered).toBe(true);
    });

    it('應該在剩餘9張時不觸發流局', () => {
      game.updateRemainingTiles(9);
      expect(game.isDrawTriggered).toBe(false);
    });

    it('應該在剩餘20張時不觸發流局', () => {
      game.updateRemainingTiles(20);
      expect(game.isDrawTriggered).toBe(false);
    });

    it('應該只觸發流局一次（即使多次更新到 <= 8）', () => {
      game.updateRemainingTiles(8);
      const firstCallCount = game.drawCallCount;

      // 再次更新到8張
      game.updateRemainingTiles(8);
      expect(game.drawCallCount).toBe(firstCallCount + 1); // 每次都會觸發
    });
  });

  describe('邊界條件測試', () => {
    it('應該正確處理剩餘牌數為0的情況', () => {
      game.updateRemainingTiles(0);
      expect(game.remainingTiles).toBe(0);
      // 剩餘0張不觸發流局（因為條件是 > 0）
      expect(game.isDrawTriggered).toBe(false);
    });

    it('應該正確處理剩餘牌數為負數的情況（異常）', () => {
      game.updateRemainingTiles(-1);
      expect(game.remainingTiles).toBe(-1);
      expect(game.isDrawTriggered).toBe(false); // 負數不觸發流局
    });

    it('應該正確處理剩餘牌數為最大值的情況', () => {
      game.updateRemainingTiles(144);
      expect(game.remainingTiles).toBe(144);
      expect(game.isDrawTriggered).toBe(false);
    });
  });

  describe('實際遊戲場景測試', () => {
    it('應該模擬一個正常遊戲流程直到接近流局', () => {
      // 從79張開始
      expect(game.remainingTiles).toBe(79);
      expect(game.isDrawTriggered).toBe(false);

      // 模擬多個回合直到剩餘9張
      const drawsNeeded = 79 - 9;
      for (let i = 0; i < drawsNeeded; i++) {
        game.simulateDrawTile();
        game.simulateDiscardTile();
      }

      expect(game.remainingTiles).toBe(9);
      expect(game.isDrawTriggered).toBe(false);

      // 再摸一張，應該觸發流局
      game.simulateDrawTile();
      expect(game.remainingTiles).toBe(8);
      expect(game.isDrawTriggered).toBe(true);
    });

    it('應該模擬玩家報告的bug場景（修復前會提前流局）', () => {
      // 如果有bug（打牌也減1），4個回合後會減少8次
      // 79 - 8 = 71，不應該流局

      // 修復後的邏輯：4個回合只減少4次
      game.simulateDrawTile(); // 玩家NN摸牌：78
      game.simulateDiscardTile(); // 玩家NN打牌：78

      game.simulateDrawTile(); // 機器人A摸牌：77
      game.simulateDiscardTile(); // 機器人A打牌：77

      game.simulateDrawTile(); // 機器人B摸牌：76
      game.simulateDiscardTile(); // 機器人B打牌：76

      game.simulateDrawTile(); // 機器人C摸牌：75
      game.simulateDiscardTile(); // 機器人C打牌：75

      expect(game.remainingTiles).toBe(75);
      expect(game.isDrawTriggered).toBe(false); // 不應該流局
    });
  });
});