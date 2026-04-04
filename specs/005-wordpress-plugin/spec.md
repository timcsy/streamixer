# Feature Specification: WordPress 外掛

**Feature Branch**: `005-wordpress-plugin`
**Created**: 2026-04-05
**Status**: Draft
**Input**: 通用 WordPress 外掛：後台管理素材組合，前台播放合成影片，支援 Shortcode、Gutenberg Block、響應式播放頁面

## User Scenarios & Testing *(mandatory)*

### User Story 1 - 後台上傳與管理素材 (Priority: P1)

網站管理員在 WordPress 後台新增一組素材：輸入標題，上傳一個音檔和一張背景圖片，選擇性上傳字幕檔，設定分類和標籤。儲存後，素材自動同步至 Streamixer 服務。管理員可在列表中看到所有已建立的素材組合，並可編輯或刪除。

**Why this priority**: 沒有素材管理，其他功能都無法運作。這是整個外掛的基礎。

**Independent Test**: 在 WordPress 後台新增一組素材，確認可儲存、編輯、刪除，且素材已同步至 Streamixer。

**Acceptance Scenarios**:

1. **Given** 管理員在 WordPress 後台，**When** 點擊「新增素材組合」並上傳音檔與背景圖片，**Then** 素材組合建立成功，列表中可見
2. **Given** 一組已建立的素材組合，**When** 管理員編輯標題或替換音檔，**Then** 變更儲存成功，Streamixer 服務同步更新
3. **Given** 一組已建立的素材組合，**When** 管理員刪除該素材，**Then** 素材從列表中移除，Streamixer 服務中的對應內容也被清除
4. **Given** 管理員上傳音檔但未上傳背景圖片，**When** 嘗試儲存，**Then** 系統提示背景圖片為必填，或使用預設背景圖片

---

### User Story 2 - Shortcode 嵌入播放器 (Priority: P2)

編輯者在任何文章或頁面中插入一段簡短的標記，前台訪客就能看到一個播放器，可以播放該素材組合的合成影片。

**Why this priority**: Shortcode 是最基本且最廣泛支援的嵌入方式，不依賴特定編輯器。

**Independent Test**: 在文章中插入 Shortcode，前台瀏覽該文章，確認播放器正常顯示並可播放。

**Acceptance Scenarios**:

1. **Given** 一組已建立的素材組合 ID 為 "my-audio"，**When** 編輯者在文章中插入 `[streamixer id="my-audio"]`，**Then** 前台顯示播放器，可正常播放合成影片
2. **Given** Shortcode 中指定的 ID 不存在，**When** 前台渲染該文章，**Then** 顯示友善的錯誤訊息（例如「找不到指定的素材」），而非空白或程式錯誤
3. **Given** 同一頁面中有多個 Shortcode，**When** 前台載入頁面，**Then** 每個播放器獨立運作，互不干擾

---

### User Story 3 - Gutenberg Block 嵌入 (Priority: P3)

編輯者在 Gutenberg 編輯器中透過視覺化區塊選取素材組合並嵌入播放器，不需手動輸入 ID 或記住語法。

**Why this priority**: Gutenberg 是 WordPress 的預設編輯器，視覺化嵌入比 Shortcode 更直覺，但功能上可由 Shortcode 替代。

**Independent Test**: 在 Gutenberg 編輯器中新增 Streamixer 區塊，從下拉選單選取素材，確認前台正常播放。

**Acceptance Scenarios**:

1. **Given** 編輯者在 Gutenberg 編輯器中，**When** 搜尋並新增「Streamixer」區塊，**Then** 顯示素材選擇介面，可從已建立的素材組合中選取
2. **Given** 編輯者選取了一組素材，**When** 在編輯器中預覽，**Then** 顯示播放器的預覽畫面（含標題和縮圖）
3. **Given** 已嵌入的區塊，**When** 前台瀏覽該頁面，**Then** 播放器正常顯示並可播放

---

### User Story 4 - 素材組合播放頁面 (Priority: P4)

訪客可在網站上瀏覽所有素材組合的列表，依分類或標籤篩選，點擊進入單則播放頁面觀看。播放頁面在各種裝置上都有良好的體驗。

**Why this priority**: 列表頁和播放頁提供獨立的瀏覽入口，但核心播放功能已由 Shortcode/Block 覆蓋。

**Independent Test**: 前台瀏覽素材列表頁，點擊進入播放頁，在手機和桌面上分別測試播放。

**Acceptance Scenarios**:

1. **Given** 多組已建立的素材組合，**When** 訪客瀏覽列表頁，**Then** 看到依日期排列的素材清單，含標題、日期、分類
2. **Given** 列表頁上有分類篩選，**When** 訪客選取某分類，**Then** 只顯示該分類的素材
3. **Given** 訪客點擊某素材，**When** 進入播放頁，**Then** 顯示播放器、標題、描述、分類、標籤
4. **Given** 訪客用手機瀏覽播放頁，**When** 播放影片，**Then** 播放器自動適應螢幕寬度，操作順暢

---

### User Story 5 - 外掛設定 (Priority: P5)

網站管理員可在後台設定外掛的基本參數，例如 Streamixer 服務的連線位址和預設背景圖片。

**Why this priority**: 設定是讓外掛能在不同環境運作的基礎，但可用合理的預設值延後。

**Independent Test**: 在設定頁變更 Streamixer URL，確認後續操作使用新的位址。

**Acceptance Scenarios**:

1. **Given** 管理員進入外掛設定頁，**When** 修改 Streamixer 服務 URL 並儲存，**Then** 後續所有素材同步和播放使用新的 URL
2. **Given** 管理員設定了預設背景圖片，**When** 新增素材組合但未上傳背景圖片，**Then** 自動使用預設背景圖片
3. **Given** Streamixer 服務 URL 設定錯誤，**When** 管理員嘗試同步素材，**Then** 顯示清楚的連線錯誤訊息

---

### Edge Cases

- Streamixer 服務不可達時，播放器 MUST 顯示友善的錯誤訊息而非空白頁面
- 上傳超大音檔（> 500MB）時，系統 MUST 顯示進度指示或限制提示
- 外掛停用後，已發布的文章中的 Shortcode MUST 不造成前台錯誤（顯示為純文字或提示訊息）
- 多個管理員同時編輯同一素材組合時，WordPress 內建的文章鎖定機制 MUST 正常運作
- 素材組合未完成（缺少音檔）時，前台播放器 MUST 顯示提示而非嘗試播放

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 外掛 MUST 註冊自訂文章類型（Custom Post Type）用於管理素材組合
- **FR-002**: 外掛 MUST 註冊自訂分類法（Taxonomy）用於分類和標籤素材
- **FR-003**: 素材組合的編輯頁 MUST 提供音檔、背景圖片、字幕檔的上傳欄位
- **FR-004**: 素材儲存時 MUST 自動將檔案同步至 Streamixer 服務
- **FR-005**: 外掛 MUST 提供 Shortcode 用於在任何文章或頁面中嵌入播放器
- **FR-006**: 外掛 MUST 提供 Gutenberg Block 用於視覺化嵌入播放器
- **FR-007**: 播放器 MUST 支援播放/暫停、進度條跳轉、音量控制、全螢幕
- **FR-008**: 素材組合 MUST 有獨立的前台播放頁面
- **FR-009**: 素材列表頁 MUST 支援依分類和標籤篩選
- **FR-010**: 外掛 MUST 提供設定頁面用於配置 Streamixer 服務 URL 和預設背景圖片
- **FR-011**: 播放器 MUST 在 Streamixer 服務不可達時顯示友善的錯誤訊息
- **FR-012**: 外掛 MUST 支援響應式設計，在手機、平板、桌面上正常顯示

### Key Entities

- **素材組合（Composition）**: 自訂文章類型，包含標題、描述、音檔、背景圖片、字幕檔、Streamixer ID
- **素材分類（Category）**: 自訂分類法，用於組織素材（例如：講道、課程、音樂）
- **素材標籤（Tag）**: 自訂分類法，用於標記素材（例如：講員名稱、主題）
- **外掛設定（Settings）**: Streamixer 服務 URL、預設背景圖片、其他全域設定

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 管理員可在 3 分鐘內完成一組素材的新增（上傳音檔+圖片+儲存）
- **SC-002**: 訪客在前台點擊播放後 5 秒內開始觀看合成影片
- **SC-003**: Shortcode 和 Gutenberg Block 皆可在前台正確渲染播放器
- **SC-004**: 播放頁面在手機（寬度 375px）和桌面（寬度 1920px）上均可正常操作
- **SC-005**: 素材列表頁可依分類篩選，結果在 1 秒內顯示

## Assumptions

- WordPress 版本為 6.0 以上（支援 Gutenberg Block API）
- Streamixer 服務已部署且可透過 HTTP 存取
- 素材檔案透過 WordPress 的媒體庫上傳，再由外掛同步至 Streamixer
- 權限管理使用 WordPress 內建的角色與能力（Roles & Capabilities），不另外實作
- 外掛只負責前端嵌入與後台管理，不修改 Streamixer 服務本身
- 外掛以獨立的 WordPress 外掛目錄存在，與 Streamixer Go 服務分開維護
