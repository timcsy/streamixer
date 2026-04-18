package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/timcsy/streamixer/src/composer"
	"github.com/timcsy/streamixer/src/config"
	"github.com/timcsy/streamixer/src/media"
)

// contentDisposition 產生符合 RFC 5987 的 attachment 標頭，保留原檔名（支援非 ASCII）
func contentDisposition(name string) string {
	ascii := sanitizeFilename(name)
	// 非 ASCII 字元以 ? 取代作為 fallback
	asciiSafe := strings.Map(func(r rune) rune {
		if r > 127 {
			return '_'
		}
		return r
	}, ascii)
	return fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, asciiSafe, url.PathEscape(name))
}

// DownloadHandler 處理影片下載請求
type DownloadHandler struct {
	cfg    config.Config
	loader *media.Loader
	pregen *composer.PregenManager
}

// NewDownloadHandler 建立下載 handler
func NewDownloadHandler(cfg config.Config, pregen *composer.PregenManager) *DownloadHandler {
	return &DownloadHandler{
		cfg:    cfg,
		loader: media.NewLoader(cfg.MediaDir),
		pregen: pregen,
	}
}

// Progress 回傳預生成進度（JSON）
func (h *DownloadHandler) Progress(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	comp, err := h.loader.Load(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	duration, err := composer.ProbeDuration(comp.Audio.Path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("無法讀取音檔資訊：%v", err))
		return
	}

	total := composer.SegmentCount(duration, h.cfg.SegmentDuration)
	// 若尚未啟動預生成，啟動之
	h.pregen.StartPregen(comp, duration)

	done := h.pregen.CountReadySegments(id)
	ready := h.pregen.IsPlaylistComplete(id)

	status := "running"
	task := h.pregen.GetStatus(id)
	if task != nil {
		switch task.Status {
		case composer.PregenCompleted:
			if ready {
				status = "completed"
			}
		case composer.PregenFailed:
			status = "failed"
		}
	}

	percent := 0
	if total > 0 {
		percent = int(float64(done) / float64(total) * 100)
		if percent > 100 {
			percent = 100
		}
	}

	resp := map[string]interface{}{
		"status":  status,
		"done":    done,
		"total":   total,
		"percent": percent,
		"ready":   ready,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Download 將 fMP4 分段串接為完整 MP4 回傳
func (h *DownloadHandler) Download(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	comp, err := h.loader.Load(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	duration, err := composer.ProbeDuration(comp.Audio.Path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("無法讀取音檔資訊：%v", err))
		return
	}
	h.pregen.StartPregen(comp, duration)

	// 等待整個預生成完成（包含 EXT-X-ENDLIST），確保所有分段都已寫出
	// 依每小時音檔約需數分鐘合成，設 30 分鐘上限
	if err := h.pregen.WaitForComplete(id, 1800); err != nil {
		writeError(w, http.StatusServiceUnavailable, fmt.Sprintf("預生成尚未完成：%v", err))
		return
	}

	outDir := filepath.Join(h.cfg.TmpDir, id)

	initPath := filepath.Join(outDir, "init.mp4")
	if _, err := os.Stat(initPath); err != nil {
		writeError(w, http.StatusServiceUnavailable, "init.mp4 尚未就緒")
		return
	}

	segments, err := collectSegments(outDir)
	if err != nil || len(segments) == 0 {
		writeError(w, http.StatusServiceUnavailable, "分段尚未就緒")
		return
	}

	// 驗證分段數與預期一致
	expected := composer.SegmentCount(duration, h.cfg.SegmentDuration)
	if len(segments) < expected {
		writeError(w, http.StatusServiceUnavailable, fmt.Sprintf("分段不完整（%d/%d）", len(segments), expected))
		return
	}

	// 計算總大小
	var totalSize int64
	initInfo, _ := os.Stat(initPath)
	totalSize += initInfo.Size()
	for _, seg := range segments {
		info, _ := os.Stat(seg)
		totalSize += info.Size()
	}

	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Content-Disposition", contentDisposition(id+".mp4"))
	w.Header().Set("Content-Length", strconv.FormatInt(totalSize, 10))

	initFile, err := os.Open(initPath)
	if err != nil {
		return
	}
	io.Copy(w, initFile)
	initFile.Close()

	for _, seg := range segments {
		f, err := os.Open(seg)
		if err != nil {
			return
		}
		io.Copy(w, f)
		f.Close()
	}
}

// collectSegments 收集並排序所有 .m4s 分段
func collectSegments(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var segments []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".m4s") {
			segments = append(segments, filepath.Join(dir, e.Name()))
		}
	}

	sort.Strings(segments)
	return segments, nil
}

// sanitizeFilename 清理檔名中的特殊字元
func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "_", "\\", "_", "\"", "_", "<", "_", ">", "_",
		"|", "_", ":", "_", "*", "_", "?", "_",
	)
	return replacer.Replace(name)
}
