# Phase 1 Data Model: 字幕字體管理

## 後端（Go）

### Font

位置：`src/fonts/types.go`

| 欄位 | 型別 | 說明 |
|------|------|------|
| `ID` | `string` | 使用者字體為 UUID；系統字體為檔名（`noto-sans-tc`） |
| `FamilyName` | `string` | 字體 family name（例：`Noto Sans CJK TC`） |
| `Source` | `FontSource` | 列舉：`SourceSystem` / `SourceUser` |
| `Ext` | `string` | `ttf` / `otf` / `ttc` |
| `Size` | `int64` | bytes |
| `UploadedAt` | `time.Time` | 僅 user 字體有值；系統字體為零值 |

驗證規則：
- `Ext` ∈ {`ttf`, `otf`, `ttc`}
- `FamilyName` 符合 `^[A-Za-z0-9 \u4e00-\u9fff\-_]{1,100}$`
- user 字體的 `Size` ≤ 10 MB

### FontSource 列舉

```go
type FontSource string

const (
    SourceSystem FontSource = "system"
    SourceUser   FontSource = "user"
)
```

### FontManager

位置：`src/fonts/manager.go`

| 方法 | 說明 |
|------|------|
| `List() ([]Font, error)` | 合併系統與使用者字體 |
| `Get(id string) (*Font, error)` | 取單一字體 metadata |
| `Upload(file io.Reader, origName string) (*Font, error)` | 驗證檔頭 → parse family name → 存檔 → 寫 metadata JSON → 建立 symlink |
| `Delete(id string) error` | 僅能刪 user；移除檔、metadata、symlink |
| `DefaultFamily() string` | 讀 `/fonts/default.txt`，空值回傳空字串 |
| `SetDefaultFamily(name string) error` | 寫 `/fonts/default.txt`（驗證該 family 存在於列表） |

配置：
- `UserDir`：預設 `/fonts/user`
- `SystemDirs`：預設 `["/usr/share/fonts"]`
- `MaxFileSize`：預設 10 MB
- `MaxCount`：預設 50（僅計 user）
- `SymlinkDir`：預設 `/usr/share/fonts/user`（讓 fontconfig 看到使用者字體）

### 檔案系統佈局

```
/fonts/                            # Docker named volume
├── default.txt                    # 全站預設 family name（單行）
└── user/
    ├── <uuid>.ttf                 # 實體字體檔
    └── <uuid>.json                # { family, ext, size, uploaded_at }

/usr/share/fonts/                  # 映像內（系統）
├── noto/
│   ├── NotoSansCJK-Regular.ttc    # 預裝
│   └── ...
└── user/                          # symlink → /fonts/user/
    └── <uuid>.ttf
```

## MediaComposition（修改）

`src/media/composition.go` 新增欄位：

| 欄位 | 型別 | 備註 |
|------|------|------|
| `FontFamily` | `string` | 空字串 = 使用全站預設 |

`Loader.Load(id)`：讀取 `/media/{id}/font.txt`，若有內容則賦值到 `FontFamily`。

## 前端 WordPress（PHP）

### Post Meta（新增）

| Meta key | 型別 | 說明 |
|----------|------|------|
| `_streamixer_font` | string | 素材指定的字體 family name；空字串 = 使用全站預設 |

### Option（新增）

| Option key | 型別 | 說明 |
|------------|------|------|
| `streamixer_default_font` | string | WP 端對全站預設字體的鏡像（與 Streamixer `/fonts/default.txt` 同步） |

儲存時機：WP 設定頁送出 → 同時 `update_option` 與 `PUT /fonts/default`。

## 合成管線

FFmpeg `subtitles=` 濾鏡組裝邏輯（`src/composer/ffmpeg.go`）：

```
resolveFontFamily(comp *MediaComposition, manager *FontManager) string:
    if comp.FontFamily != "":
        return comp.FontFamily
    return manager.DefaultFamily()  // 空字串代表使用預設 fallback

buildForceStyle(family string) string:
    parts := ["MarginV=30", "Fontsize=28"]
    if family != "":
        parts = append(parts, "FontName="+family)
    return strings.Join(parts, ",")
```

## 無狀態性

新增端點皆為 CRUD，不影響現有合成／下載狀態機。字體刪除後既有快取分段保持不變，libass fallback 保證後續合成不因找不到字體而失敗。
