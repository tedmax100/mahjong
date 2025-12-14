---
name: 功能實做 / Design Spec
about: 功能實做跟範圍
title: '[Implement] '
labels: feat
assignees: ''
---

### GitHub Issue 通用規格模板 (Markdown)

````markdown
# [類型] <填寫標題: 簡明扼要說明此 Issue 的目標>

## 摘要 / Summary
<在此填寫摘要>

## 背景與動機 / Context & Motivation
<在此填寫背景與動機>

## 範圍 / Scope
* **包含 (In-Scope):**
    * <條列出必須完成的功能點>
    * <條列出必須完成的功能點>
* **不含 (Out-of-Scope):**
    * <明確指出本次不做的事項>
    * <明確指出本次不做的事項>

## 核心需求與規格 / Core Requirements & Specifications
### <子需求標題 1>
<在此描述詳細規格>

### <子需求標題 2>
<在此描述詳細規格>

### <子需求標題 3 (例如: 資料模型 / Payload 定義)>
```typescript
// 若有需要定義資料結構，可使用 interface 示意
interface ExamplePayload {
  field1: string; // 說明
  field2: number; // 說明
}
````

## 流程圖 / Workflow Diagram (Optional)

```mermaid
graph TD;
    A-->B;
```

## 驗收標準 / Acceptance Criteria (AC)

  - [ ] **\<AC 標題 1\>**: \<具體的驗收條件描述，例如：當使用者輸入錯誤時，應顯示紅色的錯誤提示\>
  - [ ] **\<AC 標題 2\>**: \<具體的驗收條件描述\>
  - [ ] **\<AC 標題 3\>**: \<具體的驗收條件描述\>

## 技術筆記與限制 / Technical Notes & Constraints (Optional)

  * **技術棧建議**: \<例如: Node.js 使用 jose 套件\>
  * **安全性**: \<例如: Header 需設定 no-store\>
  * **其他限制**: \<任何需要注意的技術細節\>

<!-- end list -->

---

### 如何使用此模板

下次請給我類似這樣的 Prompt：
```
> 請使用以下模板，幫我撰寫一個新功能的規格。
>
> **模板：**
> (貼上方的 Markdown 模板內容)
>
> **需求描述：**
> 我們需要一個「玩家舉報功能」。
> 1. 玩家可以在遊戲結算畫面點擊別人的頭像進行舉報。
> 2. 舉報理由要有選單（作弊、言語辱罵、掛機）。
> 3. 後端要開一個 API 接收舉報資訊，存到 DB。
> 4. 不需要做後台審核介面，只要先存下來就好。
> 5. 技術上，DB 裡要有一張新的 `reports` 表。
```
