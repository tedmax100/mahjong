# 麻將遊戲測試指南

## 快速開始

### 安裝依賴

```bash
npm install
```

### 運行測試

```bash
# 監聽模式（開發時推薦）
npm test

# 運行一次（CI模式）
npm run test:run

# UI界面模式
npm run test:ui
```

## 測試結構

```
client/
├── src/
│   └── game/
│       ├── Game.js           # 源代碼
│       └── Game.test.js      # 測試文件
├── vite.config.js            # 測試配置
└── package.json              # 測試腳本
```

## 主要測試內容

### 1. 胡牌判斷 (canHu)
測試各種胡牌牌型的識別：
- 刻子胡（全刻子）
- 順子胡（全順子）
- 混合胡（順子+刻子）
- 帶吃碰槓的胡牌

### 2. 聽牌檢測 (checkReadyHand)
測試聽牌判斷邏輯：
- 兩面聽
- 單吊聽
- 多面聽
- 實際遊戲案例

### 3. 邊界條件
測試特殊情況處理：
- 空手牌
- 字牌處理
- 花色混合

## 測試結果

```
✓ 19個測試全部通過
✓ 測試時間: ~325ms
✓ 覆蓋率: 核心邏輯100%
```

## 添加新測試

在 `Game.test.js` 中添加新的測試：

```javascript
it('應該能判斷[新場景]', () => {
  const hand = [
    // 16張手牌
  ];
  const tile = 'wan-1'; // 要胡的牌

  expect(logic.canHu(hand, tile)).toBe(true);
});
```

**注意**: 確保 `hand + tile` 總共17張牌（台灣麻將標準）

## 調試測試

使用 `console.log` 在測試中輸出調試信息：

```javascript
it('調試測試', () => {
  const hand = [...];
  console.log('手牌數量:', hand.length);
  console.log('聽牌結果:', logic.checkReadyHand(hand));

  expect(...).toBe(...);
});
```

## 相關文檔

- [完整測試覆蓋文檔](../TEST_COVERAGE.md)
- [Vitest 官方文檔](https://vitest.dev/)

## 問題排查

### 測試失敗

1. 檢查手牌數量（應為16張）
2. 檢查加上要胡的牌後是否為17張
3. 檢查牌名格式（如 'wan-1', 'dong' 等）

### 測試超時

增加超時時間：

```javascript
it('長時間測試', () => {
  // 測試代碼
}, { timeout: 10000 }); // 10秒
```

## 持續集成

在 CI/CD 中運行測試：

```yaml
# .github/workflows/test.yml
- name: Run tests
  run: |
    cd client
    npm install
    npm run test:run
```
