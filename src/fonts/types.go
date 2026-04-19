package fonts

import "time"

// FontSource 字體來源
type FontSource string

const (
	SourceSystem FontSource = "system"
	SourceUser   FontSource = "user"
)

// Font 代表一個可用字體
type Font struct {
	ID         string     `json:"id"`
	FamilyName string     `json:"family_name"`
	Source     FontSource `json:"source"`
	Ext        string     `json:"ext"`
	Size       int64      `json:"size"`
	UploadedAt *time.Time `json:"uploaded_at,omitempty"`
	// Path 非 JSON 輸出：實體檔案路徑
	Path string `json:"-"`
}

// ListResponse GET /fonts 回應
type ListResponse struct {
	DefaultFamily string `json:"default_family"`
	Fonts         []Font `json:"fonts"`
}
