import { describe, it, expect } from 'vitest';
import { MahjongLogic } from './MahjongLogic.js';

describe('MahjongLogic 純邏輯測試', () => {
  describe('canHu - 基本胡牌判斷', () => {
    it('應該能判斷簡單的刻子胡牌（全刻子）', () => {
      const hand = [
        'wan-1', 'wan-1', 'wan-1',
        'wan-2', 'wan-2', 'wan-2',
        'wan-3', 'wan-3', 'wan-3',
        'wan-4', 'wan-4', 'wan-4',
        'wan-5', 'wan-5', 'wan-5',
        'dong'
      ];
      const tile = 'dong';
      expect(MahjongLogic.canHu(hand, tile, 0)).toBe(true);
    });

    it('應該能判斷簡單的順子胡牌（全順子）', () => {
      const hand = [
        'wan-1', 'wan-2', 'wan-3',
        'wan-4', 'wan-5', 'wan-6',
        'tong-1', 'tong-2', 'tong-3',
        'tiao-7', 'tiao-8', 'tiao-9',
        'tiao-1', 'tiao-2', 'tiao-3',
        'dong'
      ];
      const tile = 'dong';
      expect(MahjongLogic.canHu(hand, tile, 0)).toBe(true);
    });

    it('應該能判斷混合胡牌（順子+刻子）', () => {
      const hand = [
        'wan-1', 'wan-2', 'wan-3',
        'tong-5', 'tong-5', 'tong-5',
        'tiao-2', 'tiao-3', 'tiao-4',
        'dong', 'dong', 'dong',
        'nan', 'nan', 'nan',
        'zhong'
      ];
      const tile = 'zhong';
      expect(MahjongLogic.canHu(hand, tile, 0)).toBe(true);
    });

    it('應該拒絕不完整的牌型', () => {
      const hand = [
        'wan-1', 'wan-2', 'wan-4',
        'wan-5', 'wan-6', 'wan-7',
        'tong-1', 'tong-2', 'tong-3',
        'tiao-7', 'tiao-8', 'tiao-9',
        'dong'
      ];
      const tile = 'dong';
      expect(MahjongLogic.canHu(hand, tile, 0)).toBe(false);
    });
  });

  describe('canHu - 帶吃碰槓的胡牌判斷', () => {
    it('應該能判斷有1組碰牌的胡牌', () => {
      const hand = [
        'wan-2', 'wan-3', 'wan-4',
        'tong-5', 'tong-6', 'tong-7',
        'tiao-7', 'tiao-8', 'tiao-9',
        'zhong', 'zhong', 'zhong',
        'dong'
      ];
      const tile = 'dong';
      // 1 meld, so need 4 sets + pair = 14 tiles
      expect(MahjongLogic.canHu(hand, tile, 1)).toBe(true);
    });

    it('應該能判斷有2組吃牌的胡牌', () => {
      const hand = [
        'tiao-1', 'tiao-2', 'tiao-3',
        'zhong', 'zhong', 'zhong',
        'fa', 'fa', 'fa',
        'dong'
      ];
      const tile = 'dong';
      // 2 melds, so need 3 sets + pair = 11 tiles
      expect(MahjongLogic.canHu(hand, tile, 2)).toBe(true);
    });
  });

  describe('checkReadyHand - 聽牌判斷', () => {
    it('應該能判斷兩面聽（7,8萬聽6萬或9萬）', () => {
      // 2 melds (simulated count)
      // hand: 7,8wan + 8,8tong + 3xDong + 3xNan (need 3 sets+pair = 11 tiles)
      const hand = [
        'wan-7', 'wan-8',
        'tong-8', 'tong-8',
        'dong', 'dong', 'dong',
        'nan', 'nan', 'nan'
      ];
      // meldCount = 2
      const readyTiles = MahjongLogic.checkReadyHand(hand, 2);
      expect(readyTiles).toContain('wan-6');
      expect(readyTiles).toContain('wan-9');
    });

    it('應該能判斷多面聽', () => {
      // 1 meld
      const hand = [
        'wan-1', 'wan-2', 'wan-3',
        'wan-4', 'wan-5', 'wan-6',
        'tong-7', 'tong-8', 'tong-9',
        'tiao-2', 'tiao-3', 'tiao-4',
        'zhong'
      ];
      const readyTiles = MahjongLogic.checkReadyHand(hand, 1);
      expect(readyTiles).toContain('zhong');
    });
  });

  describe('tileValue - 牌值排序', () => {
    it('應該正確計算牌值', () => {
      expect(MahjongLogic.tileValue('wan-1')).toBe(11);
      expect(MahjongLogic.tileValue('wan-9')).toBe(19);
      expect(MahjongLogic.tileValue('tong-1')).toBe(21);
      expect(MahjongLogic.tileValue('tiao-1')).toBe(31);
      expect(MahjongLogic.tileValue('dong')).toBe(41);
      expect(MahjongLogic.tileValue('zhong')).toBe(51);
    });
  });
});
