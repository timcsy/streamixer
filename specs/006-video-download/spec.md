# Feature Specification: 影片下載功能

**Feature Branch**: `006-video-download`
**Created**: 2026-04-05
**Status**: Draft
**Input**: 影片下載功能：訪客下載合成 MP4、管理員後台批次匯出影片

## User Scenarios & Testing *(mandatory)*

### User Story 1 - 訪客下載合成影片 (Priority: P1)

訪客在網站上觀看一則講道或課程影片後，想要下載到手機或電腦離線觀看。點擊播放器旁的「下載影片」按鈕，系統將已合成的 fMP4 分段合併為一個完整的 MP4 檔案，瀏覽器開始下載。

**Why this priority**: 下載是最基本的需求，讓使用者在無網路的環境下也能觀看內容。單一素材的下載是批次匯出的基礎。

**Independent Test**: 在播放頁面點擊「下載影片」，確認瀏覽器下載一個可播放的 MP4 檔案。

**Acceptance Scenarios**:

1. **Given** 一組已同步的素材組合，**When** 訪客點擊「下載影片」，**Then** 瀏覽器開始下載一個 MP4 檔案，檔名為素材標題
2. **Given** 素材組合的預生成尚未完成，**When** 訪客點擊「下載影片」，**Then** 系統等待預生成完成後開始下載，不報錯
3. **Given** 下載的 MP4 檔案，**When** 用本地播放器開啟，**Then** 影片正常播放，包含音訊、背景畫面與字幕

---

### User Story 2 - 管理員逐個匯出影片 (Priority: P2)

管理員在 WordPress 後台的素材組合編輯頁面，可以看到「匯出影片」按鈕，點擊後下載該素材的完整 MP4 影片。

**Why this priority**: 管理員需要能在後台直接匯出影片，用於分享到其他平台或備份。這是批次匯出的前置功能。

**Independent Test**: 在素材編輯頁面點擊「匯出影片」，確認下載成功。

**Acceptance Scenarios**:

1. **Given** 管理員在素材編輯頁面，**When** 點擊「匯出影片」，**Then** 瀏覽器開始下載 MP4 檔案
2. **Given** 素材尚未同步至 Streamixer，**When** 管理員點擊「匯出影片」，**Then** 顯示「請先發佈並同步素材」的提示

---

### User Story 3 - 管理員批次匯出影片 (Priority: P3)

管理員在素材組合列表頁面，勾選多個素材，從批次操作選單中選擇「匯出影片」，系統逐個觸發下載。

**Why this priority**: 需要匯出大量影片時，逐個操作太費時。但核心下載功能（US1/US2）需先完成。

**Independent Test**: 在列表頁勾選 3 個素材，選擇「匯出影片」，確認 3 個 MP4 檔案開始下載。

**Acceptance Scenarios**:

1. **Given** 管理員在素材列表頁勾選多個素材，**When** 選擇批次操作「匯出影片」，**Then** 每個素材的 MP4 依序開始下載
2. **Given** 選取的素材中有部分尚未同步，**When** 執行批次匯出，**Then** 已同步的素材正常下載，未同步的跳過並顯示提示

---

### Edge Cases

- 素材尚未預生成完成時請求下載，系統 MUST 等待預生成完成後再開始傳輸，不回傳空檔案
- 下載過程中網路中斷，瀏覽器 MUST 能顯示下載失敗（標準 HTTP 行為）
- 極長音檔（2 小時以上）的下載 MUST 在合理時間內完成，不超時
- 多人同時下載同一素材時，系統 MUST 能正常服務每個請求

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 系統 MUST 提供下載端點，將已合成的 fMP4 分段合併為完整 MP4 檔案回傳
- **FR-002**: 下載的回應 MUST 設定正確的 Content-Type（video/mp4）和 Content-Disposition（attachment，含檔名）
- **FR-003**: 下載 MUST 等待預生成完成後才開始傳輸
- **FR-004**: WordPress 外掛的播放器旁 MUST 顯示「下載影片」按鈕
- **FR-005**: WordPress 外掛的素材編輯頁 MUST 提供「匯出影片」按鈕
- **FR-006**: WordPress 外掛的素材列表頁 MUST 支援「匯出影片」批次操作
- **FR-007**: 下載端點的存取控制 MUST 與串流端點一致（公開或需 API Key，可設定）

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 訪客可在播放頁面一鍵下載完整 MP4 影片
- **SC-002**: 下載的 MP4 可在主流播放器（VLC、系統內建）正常播放
- **SC-003**: 管理員可在後台批次匯出多個素材的影片
- **SC-004**: 10 分鐘音檔的下載在 30 秒內開始傳輸

## Assumptions

- 下載使用已預生成的 fMP4 分段合併，不重新跑 FFmpeg
- 下載端點的存取控制預設為公開（與串流端點一致），可透過設定限制
- 批次匯出在瀏覽器端逐個觸發下載（非伺服器端打包 zip）
- 下載的檔名使用素材標題，不含特殊字元
