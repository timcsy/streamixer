# Feature Specification: fMP4 分段格式（消除破音）

**Feature Branch**: `003-fmp4-segment-format`
**Created**: 2026-04-04
**Status**: Draft
**Input**: 將 HLS 分段格式從 MPEG-TS 改為 fMP4（CMAF），消除分段切換時的音訊破音

## User Scenarios & Testing *(mandatory)*

### User Story 1 - 順序播放無破音 (Priority: P1)

訪客播放一則講道影片，從頭到尾聆聽，分段切換時不會聽到任何破音、雜音或靜音間隙，音訊體驗與播放單一完整音檔無異。

**Why this priority**: 這是本功能的核心目標。目前的 .ts 格式在每次分段切換時都會出現可感知的破音，直接影響聆聽體驗。

**Independent Test**: 播放一段 30 秒以上的合成串流，仔細聆聽每個分段切換點，確認無破音。

**Acceptance Scenarios**:

1. **Given** 一個合成的影片串流，**When** 播放器從分段 N 切換到分段 N+1，**Then** 音訊銜接平順，無可感知的破音或間隙
2. **Given** 一個 10 分鐘的合成串流，**When** 從頭播放到尾，**Then** 所有分段切換點的音訊品質一致，無任何異常聲響

---

### User Story 2 - 跳轉後音訊平順 (Priority: P2)

訪客在播放過程中跳轉到任意位置，跳轉後音訊從正確的時間點開始，無破音或異常聲響。

**Why this priority**: 跳轉是次於順序播放的常見操作。新的分段格式需要確保跳轉體驗同樣平順。

**Independent Test**: 跳轉到多個不同時間點，確認每次跳轉後音訊立即正常播放，無破音。

**Acceptance Scenarios**:

1. **Given** 一段正在播放的串流，**When** 使用者跳轉到任意位置，**Then** 跳轉後音訊立即正常播放，無破音
2. **Given** 使用者連續快速跳轉多次，**When** 最終停在某個位置，**Then** 播放的音訊正常，無殘留的異常聲響

---

### User Story 3 - 主流播放器相容 (Priority: P3)

使用不同瀏覽器與裝置的訪客都能正常播放新格式的串流，不因格式變更而失去相容性。

**Why this priority**: 格式變更不能犧牲相容性，否則部分使用者將無法觀看內容。

**Independent Test**: 在 Chrome（hls.js）、Safari（原生 HLS）、手機瀏覽器上播放同一串流，確認全部正常。

**Acceptance Scenarios**:

1. **Given** 新格式的串流，**When** 在 Chrome 瀏覽器（透過 hls.js）中播放，**Then** 串流正常播放，包含音訊與畫面
2. **Given** 新格式的串流，**When** 在 Safari 瀏覽器（原生 HLS）中播放，**Then** 串流正常播放
3. **Given** 新格式的串流，**When** 在手機瀏覽器中播放，**Then** 串流正常播放

---

### Edge Cases

- 極短音檔（< 6 秒，只有一個分段）時，系統 MUST 仍能正確產生初始化段與媒體分段
- 預生成進行中使用者請求初始化段時，系統 MUST 能提供有效的初始化段
- 按需生成的分段 MUST 與預生成的分段格式完全相容，可混合播放
- 含字幕的合成 MUST 在新格式下仍正常顯示字幕

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 系統 MUST 產生共享的初始化段，供所有媒體分段使用
- **FR-002**: 系統 MUST 產生新格式的媒體分段，取代現有的分段格式
- **FR-003**: 系統 MUST 在 playlist 中包含初始化段的引用資訊
- **FR-004**: 系統 MUST 能回傳初始化段給播放器
- **FR-005**: 系統 MUST 確保預生成的分段與按需生成的分段格式一致，可混合播放
- **FR-006**: 系統 MUST 在含字幕的情況下仍能正常產生新格式的分段
- **FR-007**: 系統 MUST 在背景預生成時同時產生初始化段

### Key Entities

- **初始化段（Init Segment）**: 包含解碼器配置資訊的檔案，所有媒體分段共享此配置
- **媒體分段（Media Segment）**: 包含音訊與影片資料的分段檔案，格式與初始化段配對

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 順序播放 10 分鐘以上的合成串流，所有分段切換點無可感知的破音或間隙
- **SC-002**: 跳轉到任意位置後 2 秒內開始播放，無破音
- **SC-003**: 在 Chrome（hls.js）與 Safari 上均能正常播放
- **SC-004**: 合成效能（首個分段回應時間）不因格式變更而顯著劣化

## Assumptions

- 使用者的瀏覽器為近 3 年內的版本，支援現代 HLS 播放
- 現有的測試用前端（hls.js）不需更換，只需確認相容
- 初始化段的檔案大小很小（< 1KB），不影響首次載入速度
- 按需生成的分段仍需要初始化段存在，若不存在則由系統自動產生
