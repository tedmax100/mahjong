#!/bin/bash

# 音樂檔案重命名腳本
# 將音樂檔案命名調整為與程式命名一致

cd /home/nathan/Project/mahjong/client/public/assets/music

echo "開始重命名音樂檔案..."

# 1. 重命名牌的語音檔案（男生）- 萬子
for i in {1..9}; do
    [ -f "boy_${i}wan.ogg" ] && mv "boy_${i}wan.ogg" "boy_wan-${i}.ogg" && echo "✓ boy_${i}wan.ogg → boy_wan-${i}.ogg"
done

# 重命名牌的語音檔案（男生）- 筒子
for i in {1..9}; do
    [ -f "boy_${i}tong.ogg" ] && mv "boy_${i}tong.ogg" "boy_tong-${i}.ogg" && echo "✓ boy_${i}tong.ogg → boy_tong-${i}.ogg"
done

# 重命名牌的語音檔案（男生）- 條子
for i in {1..9}; do
    [ -f "boy_${i}tiao.ogg" ] && mv "boy_${i}tiao.ogg" "boy_tiao-${i}.ogg" && echo "✓ boy_${i}tiao.ogg → boy_tiao-${i}.ogg"
done

# 風牌和三元牌（男生）- 已經符合命名規則，不需要改
# boy_dong.ogg, boy_nan.ogg, boy_xi.ogg, boy_bei.ogg
# boy_zhong.ogg, boy_fa.ogg, boy_bai.ogg

# 2. 重命名牌的語音檔案（女生）- 萬子
for i in {1..9}; do
    [ -f "girl_${i}wan.ogg" ] && mv "girl_${i}wan.ogg" "girl_wan-${i}.ogg" && echo "✓ girl_${i}wan.ogg → girl_wan-${i}.ogg"
done

# 重命名牌的語音檔案（女生）- 筒子
for i in {1..9}; do
    [ -f "girl_${i}tong.ogg" ] && mv "girl_${i}tong.ogg" "girl_tong-${i}.ogg" && echo "✓ girl_${i}tong.ogg → girl_tong-${i}.ogg"
done

# 重命名牌的語音檔案（女生）- 條子
for i in {1..9}; do
    [ -f "girl_${i}tiao.ogg" ] && mv "girl_${i}tiao.ogg" "girl_tiao-${i}.ogg" && echo "✓ girl_${i}tiao.ogg → girl_tiao-${i}.ogg"
done

# 風牌和三元牌（女生）- 已經符合命名規則，不需要改
# girl_dong.ogg, girl_nan.ogg, girl_xi.ogg, girl_bei.ogg
# girl_zhong.ogg, girl_fa.ogg, girl_bai.ogg

# 3. 重命名動作語音檔案（男生）
[ -f "boy_ac_chi.ogg" ] && mv "boy_ac_chi.ogg" "boy_action_chi.ogg" && echo "✓ boy_ac_chi.ogg → boy_action_chi.ogg"
[ -f "boy_ac_peng.ogg" ] && mv "boy_ac_peng.ogg" "boy_action_peng.ogg" && echo "✓ boy_ac_peng.ogg → boy_action_peng.ogg"
[ -f "boy_ac_gang.ogg" ] && mv "boy_ac_gang.ogg" "boy_action_gang.ogg" && echo "✓ boy_ac_gang.ogg → boy_action_gang.ogg"
[ -f "boy_ac_hu.ogg" ] && mv "boy_ac_hu.ogg" "boy_action_hu.ogg" && echo "✓ boy_ac_hu.ogg → boy_action_hu.ogg"
[ -f "boy_ac_ting.ogg" ] && mv "boy_ac_ting.ogg" "boy_action_ting.ogg" && echo "✓ boy_ac_ting.ogg → boy_action_ting.ogg"

# 重命名動作語音檔案（女生）
[ -f "girl_ac_chi.ogg" ] && mv "girl_ac_chi.ogg" "girl_action_chi.ogg" && echo "✓ girl_ac_chi.ogg → girl_action_chi.ogg"
[ -f "girl_ac_peng.ogg" ] && mv "girl_ac_peng.ogg" "girl_action_peng.ogg" && echo "✓ girl_ac_peng.ogg → girl_action_peng.ogg"
[ -f "girl_ac_gang.ogg" ] && mv "girl_ac_gang.ogg" "girl_action_gang.ogg" && echo "✓ girl_ac_gang.ogg → girl_action_gang.ogg"
[ -f "girl_ac_hu.ogg" ] && mv "girl_ac_hu.ogg" "girl_action_hu.ogg" && echo "✓ girl_ac_hu.ogg → girl_action_hu.ogg"
[ -f "girl_ac_ting.ogg" ] && mv "girl_ac_ting.ogg" "girl_action_ting.ogg" && echo "✓ girl_ac_ting.ogg → girl_action_ting.ogg"

# 4. 重命名音效檔案
[ -f "ef_clock.ogg" ] && mv "ef_clock.ogg" "effect_clock.ogg" && echo "✓ ef_clock.ogg → effect_clock.ogg"
[ -f "ef_coins.ogg" ] && mv "ef_coins.ogg" "effect_coins.ogg" && echo "✓ ef_coins.ogg → effect_coins.ogg"
[ -f "ef_dice.ogg" ] && mv "ef_dice.ogg" "effect_dice.ogg" && echo "✓ ef_dice.ogg → effect_dice.ogg"
[ -f "ef_lose.ogg" ] && mv "ef_lose.ogg" "effect_lose.ogg" && echo "✓ ef_lose.ogg → effect_lose.ogg"
[ -f "ef_win.ogg" ] && mv "ef_win.ogg" "effect_win.ogg" && echo "✓ ef_win.ogg → effect_win.ogg"

# 5. 重命名其他音效
[ -f "buhua.ogg" ] && mv "buhua.ogg" "effect_buhua.ogg" && echo "✓ buhua.ogg → effect_buhua.ogg"
[ -f "deal.ogg" ] && mv "deal.ogg" "effect_deal.ogg" && echo "✓ deal.ogg → effect_deal.ogg"
[ -f "g_buhua.ogg" ] && mv "g_buhua.ogg" "effect_buhua_alt.ogg" && echo "✓ g_buhua.ogg → effect_buhua_alt.ogg"

# 6. 處理舊版語音檔案（沒有前綴的）- 移到 backup 資料夾
if ls [0-9]*.ogg 1> /dev/null 2>&1; then
    mkdir -p backup_old_audio
    echo ""
    echo "移動舊版語音檔案到 backup_old_audio/..."
    for file in [0-9]*.ogg g_[0-9]*.ogg; do
        [ -f "$file" ] && mv "$file" "backup_old_audio/" && echo "  → $file 移至 backup_old_audio/"
    done
fi

echo ""
echo "✅ 音樂檔案重命名完成！"
echo ""
echo "新的命名規則："
echo "  - 牌的語音：{角色}_{牌類型}.ogg (例如: boy_wan-1.ogg, girl_tong-5.ogg)"
echo "  - 動作語音：{角色}_action_{動作}.ogg (例如: boy_action_chi.ogg)"
echo "  - 音效：effect_{效果}.ogg (例如: effect_clock.ogg)"
echo "  - 背景音樂：bg_{場景}.mp3 (保持不變)"
echo ""
echo "舊版檔案已移至 backup_old_audio/ 資料夾"
