# API 契約：影片下載功能

## 新增端點

### GET /download/{id}

下載合成後的完整 MP4 影片。

**路徑參數**：
- `id`：素材組合 ID

**成功回應（200）**：
- Content-Type: `video/mp4`
- Content-Disposition: `attachment; filename="素材標題.mp4"`
- Content-Length: 所有分段檔案大小的加總
- Body: init.mp4 + seg_000.m4s + seg_001.m4s + ... 的串接二進位

**錯誤回應**：
- 404：素材組合不存在
- 503：預生成尚未完成（等待逾時）

## WordPress 外掛變更

### 前台播放器

播放器 HTML 新增下載按鈕：
```html
<a href="http://streamixer:8080/download/{id}" class="streamixer-download-btn">下載影片</a>
```

### 後台素材編輯頁

meta box 中新增「匯出影片」按鈕（連結到 `/download/{id}`）。

### 後台素材列表頁

批次操作新增「匯出影片」選項，前端 JS 逐個開啟下載 URL。
