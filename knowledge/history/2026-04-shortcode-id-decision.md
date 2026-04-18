# 2026-04 shortcode 嵌入碼改用 post ID

## 背景

`post_name` 在資料庫中儲存為 URL 編碼形式（中文 slug 會變 `%e6%b8%ac%e8%a9%a6...`）。
早期 shortcode 使用 `[streamixer id="<post_name>"]` 顯示嵌入碼，碰到中文標題時嵌入碼長達 200+ 字元，
複製貼上體驗極差（image #15 可見）。

## 決定

素材編輯頁與「插入播放器」modal 的嵌入碼改為顯示 `[streamixer id="<post->ID>"]`，永遠是短數字。

`Streamixer_Shortcode::render` 原本就同時支援 slug 與 numeric ID：
- slug 路徑用 `get_page_by_path`
- 若 slug 查不到且是數字，改以 `get_post(intval(...))` 查
所以改顯示 numeric ID 不影響既有已發布文章的 shortcode 行為。

## 為什麼不列入 experience.md

這是純 UI 折衷決策，沒有通用可遷移的教訓。純粹是「post_name 已編碼」那條 experience（仍保留）在 UI 層的另一個具體應用。
