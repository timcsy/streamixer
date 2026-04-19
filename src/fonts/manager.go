package fonts

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

// 錯誤型別
var (
	ErrFontNotFound  = errors.New("字體不存在")
	ErrSystemFont    = errors.New("系統內建字體不可刪除")
	ErrFontLimit     = errors.New("使用者字體總數已達上限")
	ErrFamilyTaken   = errors.New("已有同名使用者字體")
	ErrUnknownFamily = errors.New("字體 family 不存在於列表")
)

// Manager 負責字體管理
type Manager struct {
	userDir     string // /fonts/user
	defaultFile string // /fonts/default.txt
	symlinkDir  string // /usr/share/fonts/user
	systemDirs  []string
	maxSize     int64
	maxCount    int
}

// Config 初始化選項
type Config struct {
	FontDir     string
	SymlinkDir  string
	SystemDirs  []string
	MaxSize     int64
	MaxCount    int
}

// NewManager 建立字體管理器，同時確保目錄存在
func NewManager(cfg Config) (*Manager, error) {
	userDir := filepath.Join(cfg.FontDir, "user")
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return nil, err
	}
	if cfg.SymlinkDir != "" {
		_ = os.MkdirAll(cfg.SymlinkDir, 0755)
	}
	return &Manager{
		userDir:     userDir,
		defaultFile: filepath.Join(cfg.FontDir, "default.txt"),
		symlinkDir:  cfg.SymlinkDir,
		systemDirs:  cfg.SystemDirs,
		maxSize:     cfg.MaxSize,
		maxCount:    cfg.MaxCount,
	}, nil
}

// List 合併系統與使用者字體；系統在前，各組內按 family name 升序
func (m *Manager) List() ([]Font, error) {
	sys := m.scanSystemFonts()
	users, err := m.scanUserFonts()
	if err != nil {
		return nil, err
	}
	sort.Slice(sys, func(i, j int) bool { return sys[i].FamilyName < sys[j].FamilyName })
	sort.Slice(users, func(i, j int) bool { return users[i].FamilyName < users[j].FamilyName })
	return append(sys, users...), nil
}

// Get 取得單一字體 metadata（包含系統與使用者）
func (m *Manager) Get(id string) (*Font, error) {
	fonts, err := m.List()
	if err != nil {
		return nil, err
	}
	for i := range fonts {
		if fonts[i].ID == id {
			return &fonts[i], nil
		}
	}
	return nil, ErrFontNotFound
}

// DefaultFamily 讀全站預設 family name，未設定回傳空字串
func (m *Manager) DefaultFamily() string {
	data, err := os.ReadFile(m.defaultFile)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// SetDefaultFamily 設定全站預設；空字串代表清除
func (m *Manager) SetDefaultFamily(name string) error {
	if name != "" {
		if err := ValidateFamilyName(name); err != nil {
			return err
		}
		fonts, err := m.List()
		if err != nil {
			return err
		}
		exists := false
		for _, f := range fonts {
			if f.FamilyName == name {
				exists = true
				break
			}
		}
		if !exists {
			return ErrUnknownFamily
		}
	}
	return os.WriteFile(m.defaultFile, []byte(name), 0644)
}

// Upload 寫入新字體；先驗 header、parse family、檢查限制、寫檔與 metadata
func (m *Manager) Upload(src io.Reader, size int64) (*Font, error) {
	if m.maxSize > 0 && size > m.maxSize {
		return nil, fmt.Errorf("檔案大小超過上限 %d bytes", m.maxSize)
	}

	// 先讀前 4 bytes sniff，再串接剩餘寫入
	ext, head, err := SniffFontExt(src)
	if err != nil {
		return nil, err
	}

	// 組裝完整內容（head + 剩餘）；字體檔 <= 10MB，一次讀完可接受
	rest, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}
	data := append(head, rest...)
	if m.maxSize > 0 && int64(len(data)) > m.maxSize {
		return nil, fmt.Errorf("檔案大小超過上限 %d bytes", m.maxSize)
	}

	family, err := parseFamilyNameFromBytes(data)
	if err != nil {
		return nil, err
	}
	if err := ValidateFamilyName(family); err != nil {
		return nil, err
	}

	// 檢查使用者字體數量上限
	users, err := m.scanUserFonts()
	if err != nil {
		return nil, err
	}
	if m.maxCount > 0 && len(users) >= m.maxCount {
		return nil, ErrFontLimit
	}
	for _, u := range users {
		if u.FamilyName == family {
			return nil, ErrFamilyTaken
		}
	}

	id := uuid.New().String()
	fontPath := filepath.Join(m.userDir, id+"."+ext)
	metaPath := filepath.Join(m.userDir, id+".json")

	if err := os.WriteFile(fontPath, data, 0644); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	meta := Font{
		ID:         id,
		FamilyName: family,
		Source:     SourceUser,
		Ext:        ext,
		Size:       int64(len(data)),
		UploadedAt: &now,
		Path:       fontPath,
	}
	metaBytes, _ := json.Marshal(meta)
	if err := os.WriteFile(metaPath, metaBytes, 0644); err != nil {
		os.Remove(fontPath)
		return nil, err
	}

	// 建立 symlink 讓 fontconfig 可見
	if m.symlinkDir != "" {
		link := filepath.Join(m.symlinkDir, id+"."+ext)
		os.Remove(link)
		if err := os.Symlink(fontPath, link); err != nil {
			// symlink 失敗不致命，但記錄下來
			meta.Path = fontPath
		}
	}

	return &meta, nil
}

// Delete 刪除使用者字體；系統字體不可刪
func (m *Manager) Delete(id string) error {
	// 系統字體：ID 不帶 uuid 格式且存在於 scanSystemFonts 結果 → 禁止
	sys := m.scanSystemFonts()
	for _, f := range sys {
		if f.ID == id {
			return ErrSystemFont
		}
	}

	// 找 metadata
	metaPath := filepath.Join(m.userDir, id+".json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return ErrFontNotFound
	}
	var meta Font
	if err := json.Unmarshal(data, &meta); err != nil {
		return err
	}

	// 移除實體檔、metadata、symlink
	fontPath := filepath.Join(m.userDir, id+"."+meta.Ext)
	os.Remove(fontPath)
	os.Remove(metaPath)
	if m.symlinkDir != "" {
		os.Remove(filepath.Join(m.symlinkDir, id+"."+meta.Ext))
	}

	// 若 default 指向此 family，清空
	if m.DefaultFamily() == meta.FamilyName {
		os.WriteFile(m.defaultFile, []byte(""), 0644)
	}
	return nil
}

// Exists 檢查某 family 是否存在於目前列表中
func (m *Manager) Exists(family string) bool {
	if family == "" {
		return false
	}
	fonts, err := m.List()
	if err != nil {
		return false
	}
	for _, f := range fonts {
		if f.FamilyName == family {
			return true
		}
	}
	return false
}

// ------- internal -------

var fontExts = map[string]bool{".ttf": true, ".otf": true, ".ttc": true}

func (m *Manager) scanUserFonts() ([]Font, error) {
	entries, err := os.ReadDir(m.userDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var fonts []Font
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(m.userDir, e.Name()))
		if err != nil {
			continue
		}
		var f Font
		if err := json.Unmarshal(data, &f); err != nil {
			continue
		}
		f.Path = filepath.Join(m.userDir, f.ID+"."+f.Ext)
		fonts = append(fonts, f)
	}
	return fonts, nil
}

func (m *Manager) scanSystemFonts() []Font {
	var fonts []Font
	seenFamilies := make(map[string]bool)
	for _, dir := range m.systemDirs {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info == nil || info.IsDir() {
				return nil
			}
			// 跳過使用者字體子目錄避免重複
			if strings.Contains(path, "/user/") {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if !fontExts[ext] {
				return nil
			}
			family, err := ParseFamilyName(path)
			if err != nil || family == "" {
				return nil
			}
			if seenFamilies[family] {
				return nil
			}
			seenFamilies[family] = true
			fonts = append(fonts, Font{
				ID:         slugify(family),
				FamilyName: family,
				Source:     SourceSystem,
				Ext:        ext[1:],
				Size:       info.Size(),
				Path:       path,
			})
			return nil
		})
	}
	return fonts
}

// slugify 產生穩定 ID（小寫 + 連字號）
func slugify(s string) string {
	var b strings.Builder
	prevDash := false
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		case r >= 0x4e00 && r <= 0x9fff:
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash && b.Len() > 0 {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	return strings.TrimRight(b.String(), "-")
}
