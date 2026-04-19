package composer

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// writeWrapperASS 將 SRT/VTT 字幕內容包成一個帶有半透明背景框樣式的 ASS 檔，
// 回傳產生的 ASS 檔案路徑。若原始字幕已是 .ass/.ssa 則直接回傳原路徑。
// 動機：libass 對 SRT→ASS 的 force_style 不完整套用 BackColour alpha，
// 必須直接在 ASS Style 欄位指定才會生效。
func writeWrapperASS(subtitlePath, outDir, fontFamily string, width, height int) (string, error) {
	ext := strings.ToLower(filepath.Ext(subtitlePath))
	if ext == ".ass" || ext == ".ssa" {
		return subtitlePath, nil
	}

	events, err := parseSubtitleEvents(subtitlePath, ext)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return "", err
	}
	outPath := filepath.Join(outDir, "subtitle.ass")

	font := fontFamily
	if font == "" {
		font = "Noto Sans CJK TC"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "[Script Info]\n")
	fmt.Fprintf(&b, "ScriptType: v4.00+\n")
	fmt.Fprintf(&b, "PlayResX: %d\n", width)
	fmt.Fprintf(&b, "PlayResY: %d\n", height)
	fmt.Fprintf(&b, "ScaledBorderAndShadow: yes\n\n")

	fmt.Fprintf(&b, "[V4+ Styles]\n")
	fmt.Fprintf(&b, "Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding\n")
	// Fontsize=60：PlayResY=720 下 60px ≈ 原 SRT 預設 Fontsize=28 在 PlayResY=288 的視覺大小
	// BorderStyle=1 文字外框（非 box）；Outline=2 黑色描邊增可讀性；半透明底由 ffmpeg drawbox filter 繪製
	// libass 不正確套用 BackColour alpha，故改用 drawbox
	fmt.Fprintf(&b, "Style: Default,%s,60,&H00FFFFFF,&H00FFFFFF,&H00000000,&H00000000,0,0,0,0,100,100,0,0,1,0,0,2,10,10,60,1\n\n", font)

	fmt.Fprintf(&b, "[Events]\n")
	fmt.Fprintf(&b, "Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text\n")
	for _, e := range events {
		fmt.Fprintf(&b, "Dialogue: 0,%s,%s,Default,,0,0,0,,%s\n", e.Start, e.End, e.Text)
	}

	if err := os.WriteFile(outPath, []byte(b.String()), 0644); err != nil {
		return "", err
	}
	return outPath, nil
}

type subEvent struct {
	Start, End, Text string
}

// parseSubtitleEvents 解析 SRT / VTT 事件列（時間 + 文字）
func parseSubtitleEvents(path, ext string) ([]subEvent, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var events []subEvent
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)

	state := "idle"
	var cur subEvent
	var textLines []string

	flush := func() {
		if cur.Start != "" && cur.End != "" && len(textLines) > 0 {
			cur.Text = strings.Join(textLines, "\\N")
			events = append(events, cur)
		}
		cur = subEvent{}
		textLines = nil
	}

	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), "\r")
		if strings.HasPrefix(line, "WEBVTT") {
			continue
		}
		if strings.Contains(line, "-->") {
			flush()
			parts := strings.Split(line, "-->")
			if len(parts) == 2 {
				cur.Start = toASSTime(strings.TrimSpace(parts[0]))
				cur.End = toASSTime(strings.TrimSpace(parts[1]))
				state = "text"
			}
			continue
		}
		if line == "" {
			flush()
			state = "idle"
			continue
		}
		if state == "text" {
			// 跳過純數字的 index 行（SRT 序號）
			if !isAllDigits(line) || len(textLines) > 0 {
				textLines = append(textLines, escapeASSText(line))
			}
		}
	}
	flush()
	return events, sc.Err()
}

// toASSTime 將 HH:MM:SS,mmm 或 HH:MM:SS.mmm 轉成 ASS 的 H:MM:SS.cc
func toASSTime(s string) string {
	s = strings.Replace(s, ",", ".", 1)
	// VTT 可能是 MM:SS.mmm，補成 00:MM:SS.mmm
	if strings.Count(s, ":") == 1 {
		s = "00:" + s
	}
	// 保留到百分之一秒：SS.mmm → SS.cc
	if i := strings.Index(s, "."); i >= 0 && len(s) > i+3 {
		s = s[:i+3]
	}
	// 去掉小時前導零，確保至少 1 位（ASS 格式 H:MM:SS.cc）
	if len(s) >= 1 && s[0] == '0' && len(s) > 2 && s[1] != ':' {
		// 保留；ASS 接受 0:MM:SS.cc
	}
	return s
}

func escapeASSText(s string) string {
	// ASS 中 \ 要轉義、{ } 會被當作 override tag
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "{", "\\{")
	s = strings.ReplaceAll(s, "}", "\\}")
	return s
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
