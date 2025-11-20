# 台灣麻將 Claude Skills

這個目錄包含了三個專為台灣16張麻將遊戲開發而設計的 Claude Skills。

## 📚 可用的 Skills

### 1. 🎴 台灣麻將規則專家 (`taiwan-mahjong-expert`)

**用途**：回答關於台灣麻將規則、術語、玩法的問題

**適用情境**：
- 不確定某個規則或術語的意思
- 想了解某個台型的計算方式
- 需要解釋麻將規則給新手
- 確認特殊情況的處理方式

**使用範例**：
```
@taiwan-mahjong-expert 請解釋什麼是門清自摸？

@taiwan-mahjong-expert 清一色和混一色有什麼差別？

@taiwan-mahjong-expert 補槓後胡牌算槓上開花嗎？
```

---

### 2. 💻 麻將遊戲開發助手 (`mahjong-dev-helper`)

**用途**：協助麻將遊戲的程式開發工作

**適用情境**：
- 需要了解專案架構和程式碼結構
- 實作新功能（如聽牌提示、台數計算）
- 修復 Bug 或優化效能
- 不確定如何實作某個遊戲機制

**使用範例**：
```
@mahjong-dev-helper 如何在 Game.js 中實作聽牌判斷？

@mahjong-dev-helper Player.js 的 setTiles 方法是怎麼運作的？

@mahjong-dev-helper 我想加入音效，應該怎麼做？

@mahjong-dev-helper 如何優化麻將牌素材的載入速度？
```

---

### 3. 🧮 台數計算器 (`tai-calculator`)

**用途**：分析牌型並精確計算台數和得分

**適用情境**：
- 計算某副牌的台數
- 驗證台數計算是否正確
- 學習不同台型的組合
- 實作台數計算邏輯時參考

**使用範例**：
```
@tai-calculator 我有 🀙🀙🀙 🀚🀚🀚 🀛🀛🀛 🀜🀜🀜 🀝🀝，自摸胡牌，請計算台數

@tai-calculator 大三元是幾台？需要什麼條件？

@tai-calculator 如果我清一色加碰碰胡加自摸，總共幾台？
```

---

## 🚀 如何使用

**重要**：Skills 是**自動激活**的，不需要手動呼叫！

當你在 Claude Code 中提問時，Claude 會根據你的問題內容，自動判斷並啟用最適合的 Skill。

### 使用範例

```bash
# 當你問規則問題時，會自動啟用 taiwan-mahjong-expert
"什麼是門清自摸？"
"清一色和混一色有什麼差別？"

# 當你問開發問題時，會自動啟用 mahjong-dev-helper
"如何在 Game.js 中實作聽牌判斷？"
"Player.js 的 setTiles 方法是怎麼運作的？"

# 當你要計算台數時，會自動啟用 tai-calculator
"我有全筒子的碰碰胡自摸，總共幾台？"
"清一色是幾台？"
```

**關鍵**：只要你的問題包含相關的關鍵字（規則、台型、程式碼、Game.js、計算台數等），Claude 就會自動選擇正確的 Skill 來回答！

## 💡 使用技巧

### 自動組合使用
Claude 會根據你的問題自動選擇適合的 Skill，甚至可以在對話中切換 Skills：

```
1. 先問規則問題 → 自動啟用 taiwan-mahjong-expert
2. 再問台數計算 → 自動切換到 tai-calculator
3. 最後問如何實作 → 自動切換到 mahjong-dev-helper
```

### 提供清楚的描述
使用 Skills 時，提供越詳細的資訊，回答會越精確：

**❌ 不好的提問**：
```
這樣幾台？
```

**✅ 好的提問**：
```
我的手牌是 🀇🀈🀉 🀊🀋🀌 🀍🀎🀏 🀐🀑🀒 🀓🀓，
沒有吃碰槓，自己摸到胡牌，請計算總台數
```

### 持續對話
Skills 會記住對話脈絡，你可以持續追問：

```
User: 什麼是清一色？
Assistant: [taiwan-mahjong-expert 自動啟用，解釋清一色]

User: 那混一色呢？有什麼不同？
Assistant: [繼續使用 taiwan-mahjong-expert，解釋混一色並比較]
```

## 📖 相關文件

- [麻將Rules.md](../../麻將Rules.md) - 完整的台灣麻將規則說明
- [README.md](../../README.md) - 專案整體說明
- [client/src/game/](../../client/src/game/) - 遊戲核心程式碼

## 🤝 貢獻

如果你發現 Skills 有需要改進的地方，歡迎：
1. 直接編輯對應的 `skill.md` 檔案
2. 新增更多實用的範例
3. 補充更詳細的說明

## 📝 版本資訊

- **版本**: 1.0.0
- **建立日期**: 2025-11-20
- **相容性**: Claude Code, Anthropic Claude 3.5+

---

## 🎯 快速開始

嘗試以下問題來體驗 Skills（會自動啟用）：

```bash
# 學習規則（自動啟用 taiwan-mahjong-expert）
"請用簡單的方式解釋台灣麻將的基本規則"
"什麼是門清？什麼是自摸？"

# 開始開發（自動啟用 mahjong-dev-helper）
"專案的整體架構是什麼？"
"如何在 Game.js 中新增聽牌提示功能？"

# 計算台數（自動啟用 tai-calculator）
"清一色是幾台？有沒有範例？"
"碰碰胡加自摸總共幾台？"
```

## 🔍 如何確認 Skill 已啟用？

當 Skill 被啟用時，Claude 的回答會展現該 Skill 的專業知識。你可以觀察到：
- 使用專業術語和詳細說明
- 提供該領域特定的資訊
- 回答品質明顯針對該主題優化

祝你開發順利！🀄️
