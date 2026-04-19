package fonts

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"golang.org/x/image/font/sfnt"
)

// ErrInvalidFontHeader 檔頭不是 ttf/otf/ttc
var ErrInvalidFontHeader = errors.New("檔案不是合法的字體格式（ttf/otf/ttc）")

// ErrUnparsableFont 無法解析 name table
var ErrUnparsableFont = errors.New("無法解析字體 name table")

// 允許的 family name 字元：ASCII 字母數字 + 空白 + CJK + - _
var familyNameRe = regexp.MustCompile(`^[A-Za-z0-9 .\-_()'\x{4e00}-\x{9fff}]{1,100}$`)

// SniffFontExt 從前 4 bytes 判斷字體格式，回傳副檔名（不含點）
// ttf: 00 01 00 00
// otf: OTTO (0x4F 54 54 4F)
// ttc: ttcf (0x74 74 63 66)
func SniffFontExt(r io.Reader) (string, []byte, error) {
	head := make([]byte, 4)
	n, err := io.ReadFull(r, head)
	if err != nil {
		return "", head[:n], ErrInvalidFontHeader
	}

	switch {
	case bytes.Equal(head, []byte{0x00, 0x01, 0x00, 0x00}):
		return "ttf", head, nil
	case bytes.Equal(head, []byte("OTTO")):
		return "otf", head, nil
	case bytes.Equal(head, []byte("ttcf")):
		return "ttc", head, nil
	default:
		return "", head, ErrInvalidFontHeader
	}
}

// ParseFamilyName 解析字體檔的 family name
// 對 .ttc 回傳第一個 face 的 family name
func ParseFamilyName(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return parseFamilyNameFromBytes(data)
}

func parseFamilyNameFromBytes(data []byte) (string, error) {
	// 嘗試以 ttc 解析
	if coll, err := sfnt.ParseCollection(data); err == nil && coll.NumFonts() > 0 {
		font, err := coll.Font(0)
		if err == nil {
			return fontFamilyName(font)
		}
	}
	// 退回單一字體
	font, err := sfnt.Parse(data)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrUnparsableFont, err)
	}
	return fontFamilyName(font)
}

func fontFamilyName(font *sfnt.Font) (string, error) {
	var buf sfnt.Buffer
	name, err := font.Name(&buf, sfnt.NameIDFamily)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrUnparsableFont, err)
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return "", ErrUnparsableFont
	}
	return name, nil
}

// ValidateFamilyName 檢查 family name 是否為可安全傳給 ASS force_style 的字元集
func ValidateFamilyName(name string) error {
	if !familyNameRe.MatchString(name) {
		return fmt.Errorf("family name 含不允許字元或長度不符：%q", name)
	}
	return nil
}
