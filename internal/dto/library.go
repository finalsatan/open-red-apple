package dto

// Favorite 是收藏的一部短剧。
type Favorite struct {
	SeriesID string `json:"seriesId"`
	Title    string `json:"title"`
	CoverURL string `json:"coverUrl"`
	AddedAt  int64  `json:"addedAt"`
}

// HistoryItem 是一条观看记录（每部剧保留最近一集）。
type HistoryItem struct {
	SeriesID      string  `json:"seriesId"`
	EpisodeID     string  `json:"episodeId"`
	EpisodeNumber int     `json:"episodeNumber"`
	Title         string  `json:"title"`
	CoverURL      string  `json:"coverUrl"`
	Position      float64 `json:"position"`
	Duration      float64 `json:"duration"`
	Finished      bool    `json:"finished"`
	UpdatedAt     int64   `json:"updatedAt"`
}
