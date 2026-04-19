package media

// MediaComposition 代表一組待合成的媒體素材
type MediaComposition struct {
	ID         string
	Audio      Audio
	Background Background
	Subtitle   *Subtitle   // nil 表示無字幕
	Transcript *Transcript // nil 表示無逐字稿
	FontFamily string      // 空字串 = 使用全站預設
}
