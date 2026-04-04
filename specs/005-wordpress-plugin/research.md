# 研究筆記：WordPress 外掛

## 決策 1：素材檔案處理方式

**選擇**：使用 WordPress Media Library + `wp_enqueue_media()` + attachment ID 存為 post meta

**理由**：
- WordPress Media Library 是標準的檔案管理介面，使用者熟悉
- 檔案以 attachment 形式存在，可重複使用（例如多組素材共用同一張背景圖）
- 以 attachment ID 存入 post meta，不直接存路徑（路徑可能在搬站後改變）
- 同步至 Streamixer 時從 attachment 取得實際檔案路徑

**考慮過的替代方案**：
- 自建上傳表單：需自行處理安全性、檔案類型驗證，且 UX 不如 Media Library
- 直接存檔案路徑：搬站後路徑會失效

## 決策 2：Gutenberg Block 建置方式

**選擇**：`@wordpress/scripts` + `block.json` + 動態渲染（`render.php`）

**理由**：
- `block.json` 是 WordPress 6.x 的標準做法
- 動態渲染（server-side render）確保播放器 HTML 永遠來自 PHP，不存入區塊內容
- 避免區塊更新時的「此區塊包含非預期內容」錯誤
- `edit.js` 提供編輯器預覽，`render.php` 提供前台渲染

**考慮過的替代方案**：
- 靜態 save.js：播放器 HTML 存入區塊內容，日後修改結構需要遷移
- ACF Block：依賴第三方外掛

## 決策 3：素材同步至 Streamixer 的時機

**選擇**：在 `save_post_streamixer` hook 中以 `wp_remote_post()` 同步

**理由**：
- 儲存文章時自動觸發，不需額外操作
- 使用 WordPress 內建的 HTTP API，處理 timeout 和錯誤回傳
- 可在 meta box 中顯示同步狀態（成功/失敗）

**考慮過的替代方案**：
- 背景排程（WP Cron）：延遲同步，UX 不即時
- 手動按鈕觸發：多一步操作

## 決策 4：前台模板載入方式

**選擇**：外掛內建模板，透過 `single_template` 和 `archive_template` filter 載入

**理由**：
- 外掛提供預設模板（`templates/single-streamixer.php`、`templates/archive-streamixer.php`）
- 主題可在自己的目錄中放同名檔案覆蓋（WordPress 標準行為）
- 不強制綁定特定主題

**考慮過的替代方案**：
- 只用 Shortcode，不提供獨立頁面：不方便瀏覽
- 使用主題 template parts：依賴主題支援
