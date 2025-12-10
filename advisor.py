import json
from openai import OpenAI

client = OpenAI(
    base_url='http://localhost:11434/v1',
    api_key='ollama',
)

def get_discard_json(hand_tiles, discarded_tiles):
    hand_str = "、".join(hand_tiles)
    discard_str = "、".join(discarded_tiles)

    # 修改點 1：System Prompt 加入極度嚴格的限制
    system_prompt = """
    你是一個負責打麻將的 JSON API。
    
    【嚴格規則】
    1. target 欄位只能回傳「一張」牌的名稱。
    2. 嚴禁回傳多張牌（例如 "東風、發財" 是禁止的）。
    3. 你建議打出的牌，必須存在於「手牌」列表中，不能無中生有。
    4. 回傳格式範例：{"action": "discard", "target": "1萬", "reason": "因為..."}
    """

    user_prompt = f"""
    【手牌列表】：{hand_str}
    【場上棄牌】：{discard_str}
    
    請從【手牌列表】中選出唯一一張最不需要的牌打出。
    """

    try:
        response = client.chat.completions.create(
            model="qwen2.5:1.5b",
            messages=[
                {"role": "system", "content": system_prompt},
                {"role": "user", "content": user_prompt},
            ],
            response_format={"type": "json_object"}, 
            temperature=0.1, # 修改點 2：降低隨機性，讓它更死板聽話
            stream=False 
        )

        content = response.choices[0].message.content
        result = json.loads(content)
        
        # 修改點 3：Python 端強制防呆處理
        target_card = result.get("target", "")
        
        # 如果它還是回傳了 "東風、發財"，我們手動切分只拿第一個
        if "、" in target_card:
            target_card = target_card.split("、")[0]
        elif "," in target_card:
            target_card = target_card.split(",")[0]
        elif " " in target_card: # 如果是用空白分隔
            target_card = target_card.split(" ")[0]
            
        # 更新回 result
        result["target"] = target_card
        
        return result

    except Exception as e:
        print(f"發生錯誤: {e}")
        return None

# ==========================================
# 測試資料
# ==========================================
my_hand = [
    "1萬", "2萬", "5萬", "8萬", "9萬",
    "1筒", "3筒", "5筒", "6筒", "7筒",
    "2條", "8條", "9條",
    "東風", "紅中", "白板", "發財"
]

table_discards = [
    "1萬", "9條", "西風", "北風", "1筒", "8萬", "紅中", "2條", "5萬"
]

if __name__ == "__main__":
    print("正在分析手牌 (嚴格模式)...")
    result = get_discard_json(my_hand, table_discards)
    
    if result:
        print("\n=== 最終結果 ===")
        print(json.dumps(result, ensure_ascii=False, indent=4))
