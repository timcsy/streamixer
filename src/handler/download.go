package handler

import (
	"fmt"
	"io"
	"net/http"
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

// Download 將 fMP4 分段串接為完整 MP4 回傳
func (h *DownloadHandler) Download(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	comp, err := h.loader.Load(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	// 確保預生成已啟動
	duration, err := composer.ProbeDuration(comp.Audio.Path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("無法讀取音檔資訊：%v", err))
		return
	}
	h.pregen.StartPregen(comp, duration)

	// 等待預生成完成
	if err := h.pregen.WaitForPlaylist(id, 300); err != nil {
		writeError(w, http.StatusServiceUnavailable, fmt.Sprintf("預生成尚未完成：%v", err))
		return
	}

	outDir := filepath.Join(h.cfg.TmpDir, id)

	// 收集所有分段檔案
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

	// 計算總大小
	var totalSize int64
	initInfo, _ := os.Stat(initPath)
	totalSize += initInfo.Size()
	for _, seg := range segments {
		info, _ := os.Stat(seg)
		totalSize += info.Size()
	}

	// 設定回應標頭
	title := id // 預設用 ID 作為檔名
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.mp4"`, sanitizeFilename(title)))
	w.Header().Set("Content-Length", strconv.FormatInt(totalSize, 10))

	// 串流寫入：init.mp4 + 所有 .m4s
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
