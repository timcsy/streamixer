# 快速驗證：WordPress 外掛

## 前置需求

- WordPress 6.0+ 網站
- Streamixer Docker 服務已運行
- Node.js 18+（建置 Gutenberg Block）

## 安裝外掛

```bash
# 複製外掛到 WordPress 外掛目錄
cp -r wordpress/streamixer/ /path/to/wordpress/wp-content/plugins/streamixer/

# 建置 Gutenberg Block
cd /path/to/wordpress/wp-content/plugins/streamixer/
npm install && npm run build
```

在 WordPress 後台「外掛」頁面啟用「Streamixer」。

## 設定

1. 前往「設定 > Streamixer」
2. 輸入 Streamixer 服務 URL（例如 `http://localhost:8080`）
3. 選擇性設定預設背景圖片
4. 儲存

## 新增素材

1. 前往「素材組合 > 新增」
2. 輸入標題
3. 在 meta box 中上傳音檔、背景圖片、字幕（選填）
4. 設定分類和標籤
5. 發布 → 自動同步至 Streamixer

## 嵌入播放器

### Shortcode
在任何文章或頁面中輸入：
```
[streamixer id="my-audio-slug"]
```

### Gutenberg Block
1. 在編輯器中點擊「+」新增區塊
2. 搜尋「Streamixer」
3. 從下拉選單選取素材組合

## 瀏覽前台

- 列表頁：`https://your-site.com/streamixer/`
- 單則頁：`https://your-site.com/streamixer/my-audio-slug/`
