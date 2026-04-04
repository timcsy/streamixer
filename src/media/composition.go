package media

// MediaComposition 代表一組待合成的媒體素材
type MediaComposition struct {
	ID         string
	Audio      Audio
	Background Background
	Subtitle   *Subtitle // nil 表示無字幕
}
