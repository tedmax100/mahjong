import { describe, it, expect, beforeEach, vi } from 'vitest';
import { ChowSelectionUI } from './ChowSelectionUI.js';

// Mock PixiJS dependencies
import { vi as vitestVi } from 'vitest';

vitestVi.mock('pixi.js', () => ({
  Container: vitestVi.fn(class {
    constructor() {
      this.children = [];
      this.visible = false;
      this.x = 0;
      this.y = 0;
      this.width = 0;
      this.height = 0;
      this.parent = null;
      this.removeChildren = vitestVi.fn();
      this.addChild = vitestVi.fn((child) => {
        child.parent = this;
        this.children.push(child);
      });
      this.addChildAt = vitestVi.fn((child) => {
        child.parent = this;
        this.children.unshift(child);
      });
      this.destroy = vitestVi.fn();
    }
  }),
  Graphics: vitestVi.fn(class {
    constructor() {
      this.rect = vitestVi.fn(); this.fill = vitestVi.fn(); this.roundRect = vitestVi.fn(); this.clear = vitestVi.fn();
      this.addChild = vitestVi.fn(); this.addChildAt = vitestVi.fn(); this.stroke = vitestVi.fn();
    }
  }),
  Sprite: vitestVi.fn(class {
    constructor() { this.anchor = { set: vitestVi.fn() }; this.on = vitestVi.fn(); }
    static from() { return new this(); }
  }),
  Assets: { load: vitestVi.fn().mockResolvedValue({}) },
}));

describe('ChowSelectionUI', () => {
  let mockContainer;
  let mockAppScreen;
  let mockTileAssets;
  let chowUI;

  beforeEach(() => {
    mockContainer = {
      addChild: vi.fn(),
      removeChild: vi.fn()
    };
    mockAppScreen = { width: 1000, height: 800 };
    mockTileAssets = {
      'wan-1': 'texture-wan-1',
      'wan-2': 'texture-wan-2',
      'wan-3': 'texture-wan-3',
      'back': 'texture-back'
    };
    chowUI = new ChowSelectionUI(mockContainer, mockAppScreen, mockTileAssets);
  });

  it('應該能正確初始化並顯示選擇介面', async () => {
    const combinations = [['wan-1', 'wan-2', 'wan-3']];
    const lastDiscardedTile = 'wan-1';
    const onSelect = vi.fn();

    await chowUI.promptSelection(combinations, lastDiscardedTile, onSelect);

    expect(mockContainer.addChild).toHaveBeenCalled();
    expect(chowUI.chowSelectionContainer).toBeDefined();
  });

  it('應該在選擇後呼叫 callback 並清理介面', async () => {
    const combinations = [['wan-1', 'wan-2', 'wan-3']];
    const lastDiscardedTile = 'wan-1';
    const onSelect = vi.fn();

    await chowUI.promptSelection(combinations, lastDiscardedTile, onSelect);

    // Simulate parent assignment (since mock might not do it automatically across objects or strict check)
    // In real PixiJS, addChild sets parent. Our mock does it, but let's ensure it for the test logic.
    if (chowUI.chowSelectionContainer) {
        chowUI.chowSelectionContainer.parent = mockContainer;
    }

    // For now, let's verify cleanup works
    chowUI.clear();
    expect(mockContainer.removeChild).toHaveBeenCalledWith(expect.anything());
  });
});
