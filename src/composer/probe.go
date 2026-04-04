package composer

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// ProbeDuration 使用 ffprobe 取得音檔長度（秒）
func ProbeDuration(audioPath string) (float64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "csv=p=0",
		audioPath,
	)

	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe 執行失敗：%w", err)
	}

	durStr := strings.TrimSpace(string(out))
	duration, err := strconv.ParseFloat(durStr, 64)
	if err != nil {
		return 0, fmt.Errorf("無法解析音檔長度 %q：%w", durStr, err)
	}

	return duration, nil
}
