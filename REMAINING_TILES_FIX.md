# 剩餘牌數計算錯誤修復

## 🔴 問題描述

遊戲剛開始就顯示「流局 海底剩餘不足8張」。

## 🔍 根本原因

客戶端的剩餘牌數計算邏輯錯誤，導致**重複扣除**：

### 錯誤邏輯（修復前）
```javascript
// handleDiscard() - 打牌時
if (this.remainingTiles > 0) {
  this.updateRemainingTiles(this.remainingTiles - 1);  // ❌ 錯誤：打牌時減1
}

// handleDraw() - 摸牌時
if (this.remainingTiles > 0) {
  this.updateRemainingTiles(this.remainingTiles - 1);  // ✅ 正確：摸牌時減1
}
```

**結果**：每個回合（打牌+摸牌）剩餘牌數減少**2次**，導致牌數快速歸零。

### 計算示例（錯誤邏輯）
- 初始：79張
- 玩家NN打牌：78張 ❌
- 機器人A摸牌：77張 ❌
- 機器人A打牌：76張 ❌
- 機器人B摸牌：75張 ❌
- ...（快速減少）

## 🛠️ 修復方案

**移除 `handleDiscard()` 中的剩餘牌數更新**

### 正確邏輯
```javascript
// handleDiscard() - 打牌時
// ❌ 移除錯誤的剩餘牌數計算
// 打牌不會減少牌山的牌數，只有摸牌才會減少
// 剩餘牌數的更新已經在 handleDraw() 中處理

// handleDraw() - 摸牌時
if (this.remainingTiles > 0) {
  this.updateRemainingTiles(this.remainingTiles - 1);  // ✅ 正確
}
```

### 計算示例（正確邏輯）
- 初始：79張
- 玩家NN打牌：79張 ✅（不變）
- 機器人A摸牌：78張 ✅（-1）
- 機器人A打牌：78張 ✅（不變）
- 機器人B摸牌：77張 ✅（-1）
- 機器人B打牌：77張 ✅（不變）
- ...（正常遊戲）

## 📝 修改文件

**文件**：`client/src/game/Game.js`
**位置**：第616-619行
**動作**：移除 `handleDiscard()` 中的 `updateRemainingTiles()` 調用

## ✅ 修復效果

- ✅ 剩餘牌數只在摸牌時減少（正確）
- ✅ 打牌不影響剩餘牌數（正確）
- ✅ 流局判斷正常（牌山真的剩餘 ≤ 8張時才流局）
- ✅ 遊戲可以正常進行到結束

## 🎯 預期行為

- 初始牌數：144張
- 發牌後：約79張（144 - 4×16 - 1 = 79）
- 每次摸牌：-1張
- 每次打牌：不變
- 流局條件：牌山 ≤ 8張時

---

**修復日期**: 2025-11-29
**修復文件**: client/src/game/Game.js:616-619
