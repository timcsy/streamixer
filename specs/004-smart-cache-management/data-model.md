# 資料模型：智慧快取管理

## 新增實體

### CacheEntry（快取條目）

一個素材組合在暫存中的記錄。

| 欄位 | 說明 |
|------|------|
| compositionID | 素材組合 ID |
| lastAccessed | 最後被存取的時間 |
| size | 該素材佔用的暫存空間（bytes） |
| active | 是否正在預生成中（活躍狀態不可清除） |

**狀態規則**：
- `active = true`：正在預生成中，不可清除
- `lastAccessed + TTL < now`：已過期，可被清掃清除
- 容量超限時：按 lastAccessed 排序，最久未存取的優先淘汰

### Sweeper（清掃排程）

背景執行的定期清掃程序。

| 屬性 | 說明 |
|------|------|
| interval | 清掃頻率（預設 5 分鐘） |
| ttl | 快取過期時間（預設 30 分鐘） |
| maxSize | 容量上限（bytes） |

**清掃邏輯**：
1. 掃描所有 CacheEntry
2. 清除 `active = false` 且 `lastAccessed + TTL < now` 的條目
3. 計算總使用量，若超過 maxSize 的 90%，按 lastAccessed 升序淘汰最舊的條目直到低於 90%

## 與既有實體的關係

```
CacheManager 1──N CacheEntry
CacheEntry 1──1 PregenTask（透過 compositionID 關聯）
Sweeper 1──1 CacheManager
```

CacheManager 持有所有 CacheEntry。Sweeper 定期呼叫 CacheManager 的清掃方法。PregenTask 的狀態影響 CacheEntry 的 active 欄位。
