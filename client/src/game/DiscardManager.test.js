import { describe, it, expect, beforeEach, vi } from 'vitest';
import { DiscardManager } from './DiscardManager.js';

vi.mock('pixi.js', () => {
    const Container = class {
        constructor() {
            this.children = [];
            this.x = 0;
            this.y = 0;
            this.scale = { set: vi.fn() };
            this.destroy = vi.fn();
        }
        addChild(child) { this.children.push(child); }
        removeChild(child) { 
             const index = this.children.indexOf(child);
             if(index > -1) this.children.splice(index, 1);
        }
        removeChildren() { this.children = []; }
    };

    const Sprite = class {
        constructor() {
            this.anchor = { set: vi.fn() };
            this.x = 0;
            this.y = 0;
            this.width = 0;
            this.height = 0;
            this.destroy = vi.fn();
        }
    };
    
    return {
        Container,
        Sprite,
        Assets: { load: vi.fn().mockResolvedValue({}) },
        Graphics: class {}
    };
});

describe('DiscardManager', () => {
    let discardManager;

    beforeEach(() => {
        discardManager = new DiscardManager(800, 600);
    });

    it('should initialize correctly', () => {
        expect(discardManager.discardedTiles).toEqual([]);
        expect(discardManager.container).toBeDefined();
        expect(discardManager.screenWidth).toBe(800);
        expect(discardManager.screenHeight).toBe(600);
    });

    it('should add a discard', async () => {
        const tileAssets = { 'wan-1': {} };
        await discardManager.addDiscard('wan-1', 0, tileAssets);
        
        expect(discardManager.discardedTiles.length).toBe(1);
        expect(discardManager.discardedTiles[0].tile).toBe('wan-1');
        expect(discardManager.discardedTiles[0].playerPosition).toBe(0);
        expect(discardManager.container.children.length).toBe(1);
    });

    it('should remove last discard', async () => {
        const tileAssets = { 'wan-1': {} };
        await discardManager.addDiscard('wan-1', 0, tileAssets);
        expect(discardManager.discardedTiles.length).toBe(1);

        const removed = discardManager.removeLastDiscard();
        expect(removed.tile).toBe('wan-1');
        expect(discardManager.discardedTiles.length).toBe(0);
        expect(discardManager.container.children.length).toBe(0);
        // Sprite destruction is called on the sprite container (which is a Container)
        // Wait, addDiscard creates a Container, adds Sprite to it.
        // removeLastDiscard calls destroy() on the container.
        // My mock Container doesn't have destroy().
    });
    
    it('should clear all discards', async () => {
        const tileAssets = { 'wan-1': {} };
        await discardManager.addDiscard('wan-1', 0, tileAssets);
        await discardManager.addDiscard('wan-2', 1, tileAssets);
        
        discardManager.clear();
        
        expect(discardManager.discardedTiles.length).toBe(0);
        expect(discardManager.container.children.length).toBe(0);
    });
});
