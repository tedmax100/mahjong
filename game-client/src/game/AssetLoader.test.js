import { describe, it, expect, beforeEach, vi } from 'vitest';
import { AssetLoader } from './AssetLoader.js';

// Mock PixiJS
vi.mock('pixi.js', () => ({
  Assets: {
    load: vi.fn()
  },
  Graphics: class {
    constructor() {
      this.rect = vi.fn();
      this.fill = vi.fn();
      this.stroke = vi.fn();
      this.addChild = vi.fn();
    }
  },
  Text: class {
    constructor() {
      this.width = 20;
      this.height = 10;
    }
  }
}));

describe('AssetLoader', () => {
  let renderer;
  let loader;
  let mockTexture;

  beforeEach(() => {
    mockTexture = { id: 'mock-texture' };
    renderer = {
      generateTexture: vi.fn().mockReturnValue(mockTexture)
    };
    loader = new AssetLoader(renderer);
  });

  it('應該能正確初始化', () => {
    expect(loader.tileMapping).toBeDefined();
    expect(loader.tileMapping['wan-1']).toBe('1wf');
  });

  it('應該能載入所有定義的素材', async () => {
    const { Assets } = await import('pixi.js');
    Assets.load.mockResolvedValue(mockTexture);

    const assets = await loader.load();

    // 檢查是否所有 key 都存在於結果中
    const keys = Object.keys(loader.tileMapping);
    keys.forEach(key => {
      expect(assets[key]).toBe(mockTexture);
    });

    // 檢查調用次數 (Mapping size + 1 base texture)
    expect(Assets.load).toHaveBeenCalledTimes(keys.length + 1);
  });

  it('當載入失敗時應該使用佔位圖', async () => {
    const { Assets } = await import('pixi.js');
    // 模擬 wan-1 載入失敗，其他成功
    Assets.load.mockImplementation((path) => {
      if (path.includes('1wf')) {
        return Promise.reject(new Error('File not found'));
      }
      return Promise.resolve(mockTexture);
    });

    const assets = await loader.load();

    expect(assets['wan-1']).toBe(mockTexture); // 來自 createPlaceholderTexture
    expect(renderer.generateTexture).toHaveBeenCalled();
  });
});
