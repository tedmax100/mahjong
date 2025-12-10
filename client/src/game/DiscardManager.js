import { Container, Sprite, Assets } from 'pixi.js';

export class DiscardManager {
    constructor(screenWidth, screenHeight) {
        this.container = new Container();
        this.screenWidth = screenWidth;
        this.screenHeight = screenHeight;
        this.discardedTiles = [];
    }

    resize(width, height) {
        this.screenWidth = width;
        this.screenHeight = height;
        this.repositionAll();
    }

    async addDiscard(tile, playerPosition, tileAssets) {
        // Create discard container (base + tile)
        const discardSpriteContainer = new Container();

        // Load base texture
        let baseTexture;
        try {
            baseTexture = await Assets.load('/assets/tiles/carddown/basefdown.png');
        } catch (error) {
            console.warn('Unable to load discard base texture', error);
        }

        if (baseTexture) {
            const baseSprite = new Sprite(baseTexture);
            baseSprite.anchor.set(0.5);
            discardSpriteContainer.addChild(baseSprite);
        }

        // Create tile face sprite
        const texture = tileAssets[tile] || tileAssets['back'];
        const tileSprite = new Sprite(texture);
        tileSprite.anchor.set(0.5);
        tileSprite.y = 5; // Adjust position to align with base

        // Adjust for 'tong' tiles
        if (tile.startsWith('tong-')) {
            tileSprite.y += 8;
        }

        discardSpriteContainer.addChild(tileSprite);

        // Scale down
        const scale = 0.45;
        discardSpriteContainer.scale.set(scale);

        // Calculate position
        const position = this.calculatePosition(playerPosition, this.discardedTiles.length);
        discardSpriteContainer.x = position.x;
        discardSpriteContainer.y = position.y;

        // Add to list and container
        this.discardedTiles.push({
            sprite: discardSpriteContainer,
            playerPosition: playerPosition,
            tile: tile
        });
        this.container.addChild(discardSpriteContainer);
    }

    removeLastDiscard() {
        if (this.discardedTiles.length > 0) {
            const lastDiscard = this.discardedTiles.pop();
            this.container.removeChild(lastDiscard.sprite);
            // lastDiscard.sprite.destroy({ children: true }); // PixiJS handles child destruction often, but good to be explicit if needed. 
            // However, Game.js didn't destroy it explicitly in all cases (sometimes it just removed child). 
            // Destroying is safer for memory.
            lastDiscard.sprite.destroy({ children: true });
            return lastDiscard;
        }
        return null;
    }

    calculatePosition(playerPosition, discardIndex) {
        const centerX = this.screenWidth / 2;
        const centerY = this.screenHeight / 2;
        const scale = 0.45;
        const tileWidth = 53.4375 * scale;
        const tileHeight = 64.6875 * scale;
        const spacing = 25;

        // Calculate discard count for *this player* to determine layout
        // NOTE: The original logic in Game.js calculated layout based on "playerDiscards.length" 
        // which implies it filters the GLOBAL list for THIS player.
        // My current addDiscard logic above passes "this.discardedTiles.length" which is GLOBAL index.
        // This is a bug in my thought process. I need to filter first.
        
        const playerDiscards = this.discardedTiles.filter(d => d.playerPosition === playerPosition);
        const playerDiscardIndex = playerDiscards.length; 
        // Note: In addDiscard, we haven't pushed the new one yet, so the length is correct index.

        const maxTilesPerRow = (playerPosition === 1 || playerPosition === 3) ? 8 : 10;
        const row = Math.floor(playerDiscardIndex / maxTilesPerRow);
        const col = playerDiscardIndex % maxTilesPerRow;

        let x, y;
        switch (playerPosition) {
            case 0: // Bottom
                x = centerX - (maxTilesPerRow * (tileWidth + spacing)) / 2 + col * (tileWidth + spacing) + tileWidth / 2;
                y = centerY + 200 + row * (tileHeight + spacing);
                break;
            case 1: // Right
                x = centerX + 400 + row * (tileWidth + spacing);
                y = centerY - (maxTilesPerRow * (tileHeight + spacing)) / 2 + col * (tileHeight + spacing) + tileHeight / 2;
                break;
            case 2: // Top
                x = centerX + (maxTilesPerRow * (tileWidth + spacing)) / 2 - col * (tileWidth + spacing) - tileWidth / 2;
                y = centerY - 280 + row * (tileHeight + spacing);
                break;
            case 3: // Left
                x = centerX - 400 - row * (tileWidth + spacing);
                y = centerY - (maxTilesPerRow * (tileHeight + spacing)) / 2 + col * (tileHeight + spacing) + tileHeight / 2;
                break;
        }

        return { x, y };
    }

    repositionAll() {
        // Re-calculate positions for all tiles (e.g. on resize)
        // We need to re-calculate indices per player
        const playerCounts = { 0: 0, 1: 0, 2: 0, 3: 0 };

        this.discardedTiles.forEach(discard => {
            const { playerPosition, sprite } = discard;
            
            // We need to simulate the index logic from calculatePosition
            // But calculatePosition uses filter().length which is O(N^2) if we loop.
            // Better to track counts locally in this loop.
            
            // Wait, calculatePosition logic:
            // const playerDiscards = this.discardedTiles.filter(...) 
            // It gets the index of the *current* tile among that player's discards.
            
            const index = playerCounts[playerPosition]++;
            
            // We can reuse the math part of calculatePosition if we extract it or just copy it.
            // Let's modify calculatePosition to take 'index' as argument instead of calculating it.
            
            const pos = this.calculatePositionFromIndex(playerPosition, index);
            sprite.x = pos.x;
            sprite.y = pos.y;
        });
    }

    // Refactored helper to be pure calculation
    calculatePositionFromIndex(playerPosition, discardIndex) {
        const centerX = this.screenWidth / 2;
        const centerY = this.screenHeight / 2;
        const scale = 0.45;
        const tileWidth = 53.4375 * scale;
        const tileHeight = 64.6875 * scale;
        const spacing = 25;

        const maxTilesPerRow = (playerPosition === 1 || playerPosition === 3) ? 8 : 10;
        const row = Math.floor(discardIndex / maxTilesPerRow);
        const col = discardIndex % maxTilesPerRow;

        let x, y;
        switch (playerPosition) {
            case 0:
                x = centerX - (maxTilesPerRow * (tileWidth + spacing)) / 2 + col * (tileWidth + spacing) + tileWidth / 2;
                y = centerY + 200 + row * (tileHeight + spacing);
                break;
            case 1:
                x = centerX + 400 + row * (tileWidth + spacing);
                y = centerY - (maxTilesPerRow * (tileHeight + spacing)) / 2 + col * (tileHeight + spacing) + tileHeight / 2;
                break;
            case 2:
                x = centerX + (maxTilesPerRow * (tileWidth + spacing)) / 2 - col * (tileWidth + spacing) - tileWidth / 2;
                y = centerY - 280 + row * (tileHeight + spacing);
                break;
            case 3:
                x = centerX - 400 - row * (tileWidth + spacing);
                y = centerY - (maxTilesPerRow * (tileHeight + spacing)) / 2 + col * (tileHeight + spacing) + tileHeight / 2;
                break;
        }
        return { x, y };
    }

    // Wrap calculatePosition to use the internal state for addDiscard
    calculatePosition(playerPosition) {
        const playerDiscardsCount = this.discardedTiles.filter(d => d.playerPosition === playerPosition).length;
        return this.calculatePositionFromIndex(playerPosition, playerDiscardsCount);
    }
    
    getContainer() {
        return this.container;
    }
    
    clear() {
        this.container.removeChildren();
        this.discardedTiles = [];
    }
}
