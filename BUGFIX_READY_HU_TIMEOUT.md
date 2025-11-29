# Bug修復：聽牌按鈕與胡牌超時問題

## 修復日期
2025-11-30

## 問題描述

### Bug 1: 聽牌按鈕不出現
**症狀**：
- 玩家有4組面子（吃/碰/槓）+ 4張手牌時應該進入聽牌狀態
- 但前端沒有顯示「聽牌」(Ready/Ting) 按鈕

**範例情境**：
```
手牌: fa fa tong-7 tong-7 tong-8 (5張，剛摸牌)
面子: bei bei bei (碰), zhong zhong zhong (碰),
      tong-7 tong-8 tong-9 (吃), tiao-5 tiao-6 tiao-7 (吃)
總牌數: 5 + 12 = 17張

出牌後：
手牌: fa fa tong-7 tong-8 (4張)
面子: 4組
總牌數: 4 + 12 = 16張
聽牌：聽 tong-6 或 tong-9（完成 tong-7-8-9 + fa fa 對子）
```

**根本原因**：
- `checkSelfActions()` 只在**摸牌後**被調用
- 聽牌狀態通常在**出牌後**才達成（4組面子 + 4張手牌）
- 出牌後沒有重新檢查聽牌狀態

### Bug 2: 胡牌按鈕5秒超時太快
**症狀**：
- 當Bot打出牌後，玩家可以胡牌
- 前端顯示胡牌按鈕
- 但在玩家還沒來得及點擊時（<5秒），按鈕就消失了

**日誌證據**：
```
01:32:38 广播玩家动作: bot_4a7021_D discard tong-9
01:32:38 检测到玩家 NN 有可执行动作，开始 5 秒等待...
01:32:40 收到消息: action from NN
```

**根本原因**：
- hub.go 在檢測到玩家有可執行動作時，啟動5秒倒計時
- 5秒後自動為玩家摸牌，即使玩家還在思考
- 這導致按鈕在玩家反應過來前就消失了

## 修復方案

### 修復 1: 出牌後檢查聽牌狀態

**檔案**: `client/src/game/Game.js:227-254`

**修改內容**：
在 `handlePlayerDiscard()` 函數中，出牌後延遲300ms檢查聽牌狀態：

```javascript
handlePlayerDiscard(tileType) {
  console.log('玩家打出:', tileType);

  // 通过WebSocket发送出牌消息
  if (this.ws) {
    this.ws.sendAction('discard', { tile: tileType });
  }

  // 出牌後延遲檢查聽牌狀態（給伺服器時間更新狀態）
  setTimeout(() => {
    const myPlayer = this.players[this.myPosition];
    if (myPlayer && myPlayer.tiles) {
      const myHand = myPlayer.tiles.map(t => t.type);
      const meldCount = myPlayer.melds ? myPlayer.melds.length : 0;

      console.log(`🎴 出牌後檢查聽牌 - 手牌數: ${myHand.length}, 面子數: ${meldCount}`, myHand);

      // 檢查是否聽牌
      const readyTiles = this.checkReadyHand(myHand);
      if (readyTiles.length > 0) {
        console.log(`🎯 聽牌！聽: ${readyTiles.join(', ')}（共${readyTiles.length}張）`);
        // 顯示聽牌按鈕
        this.actionButtons.show(['ready', 'cancel']);
        this.pendingActions = ['ready'];
      }
    }
  }, 300);
}
```

**邏輯說明**：
1. 玩家出牌後，延遲300ms等待伺服器更新狀態
2. 獲取當前手牌和面子數量
3. 調用 `checkReadyHand()` 檢查是否聽牌
4. 如果聽牌，顯示「聽牌」和「取消」按鈕

### 修復 2: 延長等待時間至15秒

**檔案**: `server/internal/websocket/hub.go:262`

**修改內容**：
將等待時間從5秒延長至15秒：

```go
if withDelay {
    go func() {
        log.Printf("检测到玩家 %s (位置 %d) 有可执行动作，开始 15 秒等待...", currentPlayer.Name, currentTurnAtCheck)
        time.Sleep(15 * time.Second)  // 從 5 秒延長至 15 秒
        h.mu.Lock()
        defer h.mu.Unlock()
        log.Printf("玩家 %s 的等待时间结束。当前轮次: %d, 期望轮次: %d", currentPlayer.Name, room.CurrentTurn, currentTurnAtCheck)
        if room.GameStarted && room.CurrentTurn == currentTurnAtCheck {
            log.Printf("轮次未变，为玩家 %s 执行摸牌", currentPlayer.Name)
            h.drawForRealPlayer_needsLock(room)
        } else {
            log.Printf("轮次已从 %d 变为 %d，取消为玩家 %s 的自动摸牌", currentTurnAtCheck, room.CurrentTurn, currentPlayer.Name)
        }
    }()
}
```

**邏輯說明**：
1. 檢測到玩家有可執行動作（胡/碰/吃/槓）時，等待15秒
2. 15秒後，如果玩家還沒執行動作，自動為玩家摸牌
3. 如果輪次已變（玩家已執行動作），取消自動摸牌

## 測試建議

### 測試場景 1: 聽牌檢測
1. 開始新遊戲
2. 通過吃/碰/槓獲得4組面子
3. 剩下4張手牌，形成聽牌狀態（例如：2張對子 + 2張順子缺1）
4. 出牌後，檢查是否顯示「聽牌」按鈕

**預期結果**：
- 出牌後延遲300ms顯示「聽牌」按鈕
- 控制台輸出：`🎯 聽牌！聽: ...`

### 測試場景 2: 胡牌等待時間
1. 開始新遊戲
2. 讓自己進入聽牌狀態
3. 等待Bot打出可以胡的牌
4. 觀察胡牌按鈕顯示時間

**預期結果**：
- 胡牌按鈕顯示後，至少有15秒的時間可以點擊
- 控制台輸出：`检测到玩家 ... 有可执行动作，开始 15 秒等待...`
- 點擊胡牌按鈕後，動作立即執行

### 測試場景 3: 綜合測試
1. 玩完整的一局遊戲
2. 測試各種聽牌情況（清一色、碰碰胡、七對等）
3. 測試胡牌、碰牌、吃牌、槓牌的優先權
4. 確認15秒等待時間足夠玩家反應

## 已知限制

1. **聽牌檢測邏輯**：目前依賴 `checkReadyHand()` 函數，該函數會測試所有可能的牌（34種）來判斷是否聽牌。對於複雜牌型可能有性能影響。

2. **出牌後延遲**：300ms的延遲是為了等待伺服器更新狀態，但在網路延遲較大時可能不足夠。

3. **等待時間固定**：15秒對大多數情況足夠，但對於需要仔細思考的複雜胡牌（如大三元、清一色）可能還是不夠。未來可考慮根據動作類型調整等待時間（胡牌30秒，碰/吃/槓10秒）。

## 未來改進建議

1. **動態等待時間**：
   - 胡牌：30秒
   - 槓牌：15秒
   - 碰牌：10秒
   - 吃牌：10秒

2. **聽牌提示優化**：
   - 顯示具體聽哪些牌
   - 顯示聽牌後可能的台數
   - 高亮聽牌的組合

3. **優先權系統整合**：
   - 目前優先權系統已在後端實現（`room.go`），但尚未完全整合到WebSocket流程
   - 應該收集所有玩家的動作，然後根據優先權決定執行順序

4. **用戶體驗改進**：
   - 當有可執行動作時，播放音效提示
   - 顯示倒計時，讓玩家知道還有多少時間
   - 按鈕閃爍或動畫效果吸引注意

## 相關檔案

- `client/src/game/Game.js` - 前端遊戲邏輯
- `server/internal/websocket/hub.go` - WebSocket通訊處理
- `server/internal/game/room.go` - 遊戲房間邏輯（優先權系統）
- `server/internal/game/priority_test.go` - 優先權系統單元測試
