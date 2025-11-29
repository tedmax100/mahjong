# 真實玩家自動摸牌功能修復

## 📋 問題描述

### Bug 發現
在實際遊戲測試中發現，真實玩家在輪到自己回合時**沒有自動摸牌**，導致手牌數量持續減少。

**症狀**：
- 機器人每次回合都會自動摸牌（日誌顯示："Bot XX 摸到了 YY"）
- 真實玩家沒有摸牌記錄（日誌中缺少："玩家 XX 摸到了 YY"）
- 玩家手牌數量異常：應該是16張，實際只有13張

**手牌數量變化追蹤**：
| 動作 | 手牌數 | 應有數量 | 狀態 |
|------|--------|----------|------|
| 初始（莊家） | 17張 | 17張 | ✅ |
| 打出 wan-8 | 16張 | 16張 | ✅ |
| 打出 tiao-2 | 15張 | 16張（摸牌17→打牌16） | ❌ |
| 打出 tong-6 | 14張 | 16張 | ❌ |
| 打出 wan-4 | 13張 | 16張 | ❌ |

## 🔍 根本原因

### 機器人的流程（正常）：
1. `CheckAndPlayBotTurn()` 被調用
2. 檢查手牌是否有16張
3. 如果有16張，自動從牌山摸一張牌
4. 自動打出一張牌
5. 廣播 discard 動作

**代碼位置**：`server/internal/websocket/hub.go:270-277`
```go
// 只有在正常轮次（手牌16张）才摸牌
if len(currentPlayer.Hand) == 16 {
    drawnTile := room.Game.DrawTile()
    if drawnTile != "" {
        log.Printf("Bot %s 摸到了 %s", currentPlayer.Name, drawnTile)
        currentPlayer.Hand = append(currentPlayer.Hand, drawnTile)
    }
}
```

### 真實玩家的流程（有問題）：
1. `HandleDiscard()` 被調用
2. 切換到下一個玩家（`NextTurn()`）
3. 廣播 discard 動作
4. **沒有摸牌邏輯！** ❌

## 🛠️ 修復方案

### 1. 修改 `CheckAndPlayBotTurn` 函數
在檢測到當前玩家是真實玩家時，調用新的 `DrawForRealPlayer` 函數。

**修改位置**：`server/internal/websocket/hub.go:257-261`
```go
// 如果是真实玩家，自动发牌给他们
if len(currentPlayer.ID) <= 4 || currentPlayer.ID[:4] != "bot_" {
    h.DrawForRealPlayer(room)
    return
}
```

### 2. 新增 `DrawForRealPlayer` 函數
專門處理真實玩家的自動摸牌。

**代碼位置**：`server/internal/websocket/hub.go:307-341`
```go
// DrawForRealPlayer 为真实玩家自动发牌
func (h *Hub) DrawForRealPlayer(room *game.Room) {
    if room == nil || room.Game == nil || !room.GameStarted {
        return
    }

    h.mu.Lock()
    defer h.mu.Unlock()

    if room.CurrentTurn < 0 || room.CurrentTurn >= len(room.Players) {
        return
    }

    currentPlayer := room.Players[room.CurrentTurn]

    // 确认是真实玩家
    if len(currentPlayer.ID) > 4 && currentPlayer.ID[:4] == "bot_" {
        return
    }

    // 只有在正常轮次（手牌16张）才摸牌
    if len(currentPlayer.Hand) == 16 {
        drawnTile := room.Game.DrawTile()
        if drawnTile != "" {
            log.Printf("玩家 %s 摸到了 %s", currentPlayer.Name, drawnTile)
            currentPlayer.Hand = append(currentPlayer.Hand, drawnTile)

            // 广播摸牌事件
            h.BroadcastDrawTile(room, currentPlayer.ID, drawnTile)

            // 記錄摸牌後的手牌狀態
            game.LogPlayerHand(currentPlayer, "摸牌: "+drawnTile)
        }
    }
}
```

### 3. 新增 `BroadcastDrawTile` 函數
廣播摸牌事件給所有客戶端。

**代碼位置**：`server/internal/websocket/hub.go:343-371`
```go
// BroadcastDrawTile 广播摸牌事件
func (h *Hub) BroadcastDrawTile(room *game.Room, playerID, tile string) {
    message := map[string]interface{}{
        "type": "player_action",
        "data": map[string]interface{}{
            "playerId":    playerID,
            "action":      "draw",
            "tile":        tile,
            "currentTurn": room.CurrentTurn,
        },
    }

    msgBytes, _ := json.Marshal(message)

    log.Printf("广播玩家摸牌: %s 摸 %s (当前轮次: %d)", playerID, tile, room.CurrentTurn)

    // 向所有玩家发送
    for _, clientInterface := range room.Clients {
        client, ok := clientInterface.(*Client)
        if !ok {
            continue
        }
        select {
        case client.Send <- msgBytes:
        default:
            log.Printf("警告：客户端 %s 的发送缓冲区已满，摸牌消息被丢弃", client.UserName)
        }
    }
}
```

## ✅ 測試覆蓋

### 新增測試文件
**文件**：`server/internal/websocket/hub_test.go`

### 測試用例

#### 1. `TestDrawForRealPlayer` - 真實玩家自動摸牌功能
- **測試內容**：真實玩家在正常輪次（16張手牌）時能自動摸牌
- **驗證點**：
  - 手牌數量從16張增加到17張 ✅
  - 牌山數量減少1 ✅
  - 摸牌事件被廣播 ✅

#### 2. `TestDrawForRealPlayer_BotShouldNotDraw` - 機器人不應摸牌
- **測試內容**：機器人不應該通過 `DrawForRealPlayer` 摸牌
- **驗證點**：
  - 機器人手牌數量保持不變 ✅

#### 3. `TestDrawForRealPlayer_WrongHandSize` - 錯誤手牌數量
- **測試內容**：手牌不是16張時不應該摸牌
- **驗證點**：
  - 手牌數量保持不變 ✅

#### 4. `TestDrawForRealPlayer_GameNotStarted` - 遊戲未開始
- **測試內容**：遊戲未開始時不應該摸牌
- **驗證點**：
  - 手牌數量保持不變 ✅

#### 5. `TestCheckAndPlayBotTurn_CallsDrawForRealPlayer` - 整合測試
- **測試內容**：`CheckAndPlayBotTurn` 會為真實玩家調用 `DrawForRealPlayer`
- **驗證點**：
  - 真實玩家成功摸到牌 ✅
  - 手牌數量正確增加 ✅

## 📊 測試結果

### WebSocket 層測試
```bash
=== RUN   TestDrawForRealPlayer
--- PASS: TestDrawForRealPlayer (0.00s)
=== RUN   TestDrawForRealPlayer_BotShouldNotDraw
--- PASS: TestDrawForRealPlayer_BotShouldNotDraw (0.00s)
=== RUN   TestDrawForRealPlayer_WrongHandSize
--- PASS: TestDrawForRealPlayer_WrongHandSize (0.00s)
=== RUN   TestDrawForRealPlayer_GameNotStarted
--- PASS: TestDrawForRealPlayer_GameNotStarted (0.00s)
=== RUN   TestCheckAndPlayBotTurn_CallsDrawForRealPlayer
--- PASS: TestCheckAndPlayBotTurn_CallsDrawForRealPlayer (0.00s)
PASS
ok  	mahjong/internal/websocket	0.003s
```

**通過率**: 100% (5/5)

### 遊戲邏輯層測試
```bash
=== RUN   TestCanHu
--- PASS: TestCanHu (0.00s)
=== RUN   TestCanPong
--- PASS: TestCanPong (0.00s)
=== RUN   TestCanKong
--- PASS: TestCanKong (0.00s)
=== RUN   TestCanChow
--- PASS: TestCanChow (0.00s)
=== RUN   TestIsFlowerTile
--- PASS: TestIsFlowerTile (0.00s)
=== RUN   TestCheckDraw
--- PASS: TestCheckDraw (0.00s)
=== RUN   TestCanExposedKong
--- PASS: TestCanExposedKong (0.00s)
=== RUN   TestDrawTileFromEnd
--- PASS: TestDrawTileFromEnd (0.00s)
=== RUN   TestHandleKong_PromotedKongFromDiscard
--- PASS: TestHandleKong_PromotedKongFromDiscard (0.00s)
PASS
ok  	mahjong/internal/game	0.002s
```

**通過率**: 100% (9/9)

### 總計
- **總測試數**: 14
- **通過測試數**: 14
- **失敗測試數**: 0
- **通過率**: 100%

## 🎯 修復效果

修復後，真實玩家的流程變為：

1. 玩家打出牌 → `HandleDiscard()`
2. 切換到下一個玩家 → `NextTurn()`
3. 廣播打牌動作 → `BroadcastPlayerAction()`
4. 檢查並執行下一個玩家回合 → `CheckAndPlayBotTurn()`
5. **如果是真實玩家** → `DrawForRealPlayer()` ✅
   - 自動從牌山摸一張牌
   - 添加到玩家手牌
   - 廣播摸牌事件
   - 記錄日誌
6. 等待玩家打牌

### 預期日誌輸出
```
2025/11/29 17:47:12 轮到下一位玩家（位置: 0）
2025/11/29 17:47:12 玩家 NN 摸到了 tiao-3  ← 新增！
2025/11/29 17:47:12 广播玩家摸牌: player_xxx 摸 tiao-3 (当前轮次: 0)  ← 新增！
2025/11/29 17:47:12 📋 [摸牌: tiao-3] 玩家 NN (位置0)  ← 新增！
2025/11/29 17:47:12    手牌 (17張): tiao-3 tiao-6 tiao-6 ...
2025/11/29 17:47:12    總牌數: 17張
```

## 🔄 相關文件

### 修改的文件
- `server/internal/websocket/hub.go` - 添加真實玩家摸牌邏輯

### 新增的文件
- `server/internal/websocket/hub_test.go` - 摸牌功能測試

### 相關文檔
- `DEBUG_LOGGING.md` - 遊戲日誌記錄功能說明
- `TEST_SUMMARY_v2.md` - 測試總結（吃碰槓後聽牌檢測）
- `TEST_COVERAGE.md` - 測試覆蓋文檔

## 📝 後續建議

### 可能的改進
1. **客戶端優化**：在客戶端顯示摸牌動畫
2. **性能優化**：考慮批量處理摸牌事件以減少網絡開銷
3. **錯誤處理**：添加更完善的錯誤處理和回滾機制
4. **日誌級別**：添加可配置的日誌級別控制

### 已知限制
- 當前實現假設玩家ID不以"bot_"開頭就是真實玩家
- 摸牌邏輯只在手牌正好16張時觸發（正常輪次）

## 🎉 總結

✅ **成功修復真實玩家自動摸牌bug**
✅ **添加完整的單元測試覆蓋**
✅ **所有測試100%通過**
✅ **日誌記錄正確顯示摸牌動作**
✅ **不影響現有功能**

---

**修復版本**: v1.0
**修復日期**: 2025-11-29
**測試通過率**: 100% (14/14)
**修復作者**: Claude Code
