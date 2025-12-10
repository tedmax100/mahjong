/**
 * 純麻將規則邏輯
 * 負責胡牌判斷、聽牌偵測、牌值計算等，不涉及 UI
 */
export class MahjongLogic {
  /**
   * 檢查是否可以胡牌
   * @param {Array<string>} hand - 手牌列表
   * @param {string} tile - 進張（要胡的牌）
   * @param {number} meldCount - 已有的面子數量（吃碰槓）
   * @returns {boolean}
   */
  static canHu(hand, tile, meldCount) {
    const fullHand = [...hand, tile];
    const requiredMelds = 5 - meldCount;
    // Standard hand size check: 16 tiles + 1 winning tile = 17 total. 
    // Each meld reduces hand size by 3.
    const requiredTiles = requiredMelds * 3 + 2;

    if (fullHand.length !== requiredTiles) {
      return false;
    }

    const tileCount = {};
    fullHand.forEach(t => {
      tileCount[t] = (tileCount[t] || 0) + 1;
    });

    for (const [t, count] of Object.entries(tileCount)) {
      if (count >= 2) {
        const remainingTiles = { ...tileCount };
        remainingTiles[t] -= 2;

        if (this.canFormMelds(remainingTiles, requiredMelds)) {
          return true;
        }
      }
    }

    return false;
  }

  /**
   * 檢查剩餘的牌是否能組成指定數量的面子
   * @param {Object} tileCount - 牌的計數物件
   * @param {number} meldsNeeded - 需要組成的面子數量
   * @returns {boolean}
   */
  static canFormMelds(tileCount, meldsNeeded) {
    const tiles = { ...tileCount };

    Object.keys(tiles).forEach(key => {
      if (tiles[key] === 0) delete tiles[key];
    });

    if (Object.keys(tiles).length === 0 && meldsNeeded === 0) {
      return true;
    }

    if (Object.keys(tiles).length === 0 || meldsNeeded === 0) {
      return false;
    }

    const firstTile = Object.keys(tiles).sort()[0];
    const count = tiles[firstTile];

    // 嘗試組成刻子
    if (count >= 3) {
      const newTiles = { ...tiles };
      newTiles[firstTile] -= 3;
      if (newTiles[firstTile] === 0) delete newTiles[firstTile];

      if (this.canFormMelds(newTiles, meldsNeeded - 1)) {
        return true;
      }
    }

    // 嘗試組成順子
    const match = firstTile.match(/^(wan|tong|tiao)-(\d)$/);
    if (match) {
      const suit = match[1];
      const num = parseInt(match[2]);

      if (num <= 7) {
        const tile2 = `${suit}-${num + 1}`;
        const tile3 = `${suit}-${num + 2}`;

        if (tiles[tile2] >= 1 && tiles[tile3] >= 1) {
          const newTiles = { ...tiles };
          newTiles[firstTile] -= 1;
          newTiles[tile2] -= 1;
          newTiles[tile3] -= 1;

          if (newTiles[firstTile] === 0) delete newTiles[firstTile];
          if (newTiles[tile2] === 0) delete newTiles[tile2];
          if (newTiles[tile3] === 0) delete newTiles[tile3];

          if (this.canFormMelds(newTiles, meldsNeeded - 1)) {
            return true;
          }
        }
      }
    }

    return false;
  }

  /**
   * 檢查是否聽牌
   * @param {Array<string>} hand - 手牌
   * @param {number} meldCount - 面子數量
   * @returns {Array<string>} - 聽的牌列表
   */
  static checkReadyHand(hand, meldCount) {
    const allPossibleTiles = [];

    ['wan', 'tong', 'tiao'].forEach(suit => {
      for (let num = 1; num <= 9; num++) {
        allPossibleTiles.push(`${suit}-${num}`);
      }
    });

    ['dong', 'nan', 'xi', 'bei', 'zhong', 'fa', 'bai'].forEach(tile => {
      allPossibleTiles.push(tile);
    });

    const readyTiles = [];
    for (const tile of allPossibleTiles) {
      if (this.canHu(hand, tile, meldCount)) {
        readyTiles.push(tile);
      }
    }

    return readyTiles;
  }

  /**
   * 計算牌的排序值（用於手牌排序）
   * @param {string} tile - 牌型
   * @returns {number} - 排序值
   */
  static tileValue(tile) {
    const parts = tile.split('-');
    const suit = parts[0];
    let num = 0;

    if (parts.length > 1) {
        num = parseInt(parts[1], 10);
    }

    let suitOrder = 0;

    switch (suit) {
    case "wan":
        suitOrder = 1;
        break;
    case "tong":
        suitOrder = 2;
        break;
    case "tiao":
        suitOrder = 3;
        break;
    case "dong":
        suitOrder = 4;
        num = 1;
        break;
    case "nan":
        suitOrder = 4;
        num = 2;
        break;
    case "xi":
        suitOrder = 4;
        num = 3;
        break;
    case "bei":
        suitOrder = 4;
        num = 4;
        break;
    case "zhong":
        suitOrder = 5;
        num = 1;
        break;
    case "fa":
        suitOrder = 5;
        num = 2;
        break;
    case "bai":
        suitOrder = 5;
        num = 3;
        break;
    }

    return suitOrder * 10 + num;
  }
}
