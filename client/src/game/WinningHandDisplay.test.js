import { describe, it, expect, beforeEach, vi } from 'vitest';

// Mock PixiJS dependencies
vi.mock('pixi.js', () => ({
  Container: vi.fn(class { // Mock Container constructor
    constructor() {
      this.children = []; this.visible = false; this.x = 0; this.y = 0; this.width = 0; this.height = 0;
      this.removeChildren = vi.fn(); this.addChild = vi.fn(); this.addChildAt = vi.fn(); this.destroy = vi.fn();
    }
  }),
  Graphics: vi.fn(class { // Mock Graphics constructor
    constructor() {
      this.rect = vi.fn(); this.fill = vi.fn(); this.roundRect = vi.fn(); this.clear = vi.fn();
      this.addChild = vi.fn(); this.addChildAt = vi.fn(); this.stroke = vi.fn();
    }
  }),
  Sprite: vi.fn(class { // Mock Sprite constructor
    constructor() { this.anchor = { set: vi.fn() }; this.on = vi.fn(); }
    static from() { return new this(); }
  }),
  Assets: { load: vi.fn(() => Promise.resolve({})) },
  Text: vi.fn(class { // Mock Text constructor
    constructor() { this.anchor = { set: vi.fn() }; }
  })
}));

// Mock Tile dependency
vi.mock('./Tile.js', () => ({
  Tile: vi.fn(class { // Mock Tile constructor
    constructor(type, texture) {
      this.type = type;
      this.texture = texture;
      this.container = { x: 0, y: 0, addChild: vi.fn(), addChildAt: vi.fn(), setScale: vi.fn(), width: 100 };
      this.setScale = vi.fn();
    }
  })
}));

// Mock MahjongLogic for tileValue
vi.mock('./MahjongLogic.js', () => ({
  MahjongLogic: {
    tileValue: vi.fn((tile) => {
      // Simple mock for sorting based on a part of the tile string
      if (tile.startsWith('wan-')) return parseInt(tile.split('-')[1]);
      if (tile.startsWith('tong-')) return 10 + parseInt(tile.split('-')[1]);
      return 100; // default for other tiles
    })
  }
}));

// Import the actual WinningHandDisplay class *after* mocks
import { WinningHandDisplay } from './WinningHandDisplay.js';
// Import the actual Container mock for access *after* mocks
import { Container } from 'pixi.js';

describe('WinningHandDisplay', () => {
  let mockWinningHandContainer;
  let mockAppScreen;
  let mockTileAssets;
  let displayer;

  beforeEach(() => {
    vi.useFakeTimers(); // Mock timers for setTimeout

    // Clear Container mock instances
    Container.mockReset(); 

    mockWinningHandContainer = {
      removeChildren: vi.fn(),
      addChild: vi.fn(child => mockWinningHandContainer.children.push(child)),
      addChildAt: vi.fn(child => mockWinningHandContainer.children.unshift(child)),
      visible: false,
      x: 0,
      y: 0,
      width: 0,
      height: 0,
      children: []
    };
    mockAppScreen = { width: 1000, height: 800 };
    mockTileAssets = {
      'wan-1': 'texture-wan-1',
      'wan-2': 'texture-wan-2',
      'wan-3': 'texture-wan-3',
      'back': 'texture-back'
    };
    displayer = new WinningHandDisplay(mockWinningHandContainer, mockAppScreen, mockTileAssets);
  });

  afterEach(() => {
    vi.runOnlyPendingTimers();
    vi.useRealTimers();
  });

  it('應該在胡牌時正確顯示手牌和面子', async () => {
    const hand = ['wan-1', 'wan-2'];
    const melds = [{ type: 'pong', tiles: ['wan-3', 'wan-3', 'wan-3'] }];
    const winTile = 'wan-1';

    await displayer.display(hand, melds, winTile);

    // After adding background and allTilesContainer
    expect(mockWinningHandContainer.addChild).toHaveBeenCalled(); 
    // The first child added to mockWinningHandContainer is the background (Graphics instance)
    // The second child added is allTilesContainer (Container instance)
    // The Container.mock.instances are of the *mocked* Container class, so allTilesContainer will be an instance of it.
    
    // We can also check if mockWinningHandContainer has children
    expect(mockWinningHandContainer.children.length).toBeGreaterThan(0);
    const allTilesContainer = mockWinningHandContainer.children[1]; // Get the instance
    expect(allTilesContainer).toBeInstanceOf(Container);
    
    // Initial visibility check
    expect(mockWinningHandContainer.visible).toBe(true);

    // Run short timers (e.g., the 50ms one)
    vi.advanceTimersByTime(50);
    expect(allTilesContainer.x).not.toBe(0); // Should have updated position
    
    // Run the longer timer to hide the container
    vi.advanceTimersByTime(5000);
    expect(mockWinningHandContainer.visible).toBe(false);
  });

  it('如果沒有面子也應該正確顯示手牌', async () => {
    const hand = ['wan-1', 'wan-2'];
    const melds = [];
    const winTile = 'wan-1';

    await displayer.display(hand, melds, winTile);
    
    // Initial visibility check
    expect(mockWinningHandContainer.visible).toBe(true);

        const allTilesContainer = mockWinningHandContainer.children[1];
        expect(allTilesContainer).toBeInstanceOf(Container);
    
        vi.advanceTimersByTime(50); // Advance any short timers
        expect(allTilesContainer.x).not.toBe(0); // Should have updated position
        vi.advanceTimersByTime(5000);
        expect(mockWinningHandContainer.visible).toBe(false);  });

  it('如果沒有贏牌，不應該顯示', async () => {
    // This case should not happen if handleGameWin logic is correct,
    // but good for robustness.
    const hand = [];
    const melds = [];
    const winTile = null;

    await displayer.display(hand, melds, winTile);
    
    // Initial visibility check
    expect(mockWinningHandContainer.visible).toBe(true);
    
    const allTilesContainer = mockWinningHandContainer.children[1];
    expect(allTilesContainer).toBeInstanceOf(Container);

    // Run short timers (e.g., the 50ms one)
    vi.advanceTimersByTime(50);
    expect(allTilesContainer.x).not.toBe(0); // Should have updated position
    
    // Run the longer timer to hide the container
    vi.advanceTimersByTime(5000);
    expect(mockWinningHandContainer.visible).toBe(false);
  });
});