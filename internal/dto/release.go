package dto

import (
	"time"

	"github.com/beilunyang/openhongguo/sdk/model"
)

// NewRelease 是上新日历中的一部作品。
type NewRelease struct {
	SeriesID     string   `json:"seriesId"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	CoverURL     string   `json:"coverUrl"`
	EpisodeCount int      `json:"episodeCount"`
	Online       bool     `json:"online"`
	PublishAt    string   `json:"publishAt"`
	Categories   []string `json:"categories"`
	Heat         string   `json:"heat"`
	Note         string   `json:"note"`
}

// NewReleasePage 是上新日历中的一天。
type NewReleasePage struct {
	Date    string       `json:"date"`
	Dates   []string     `json:"dates"`
	Items   []NewRelease `json:"items"`
	HasMore bool         `json:"hasMore"`
	Cursor  string       `json:"cursor"`
}

// FromNewReleasePage 转换上新日历的一页。
func FromNewReleasePage(p model.NewReleasePage) NewReleasePage {
	out := NewReleasePage{
		Date:    p.Date,
		Dates:   p.Dates,
		HasMore: p.HasMore,
		Cursor:  p.Cursor,
	}
	out.Items = make([]NewRelease, 0, len(p.Items))
	for _, r := range p.Items {
		publishAt := ""
		if !r.PublishAt.IsZero() {
			publishAt = r.PublishAt.Format(time.RFC3339)
		}
		out.Items = append(out.Items, NewRelease{
			SeriesID:     r.SeriesID,
			Title:        r.Title,
			Description:  r.Description,
			CoverURL:     r.CoverURL,
			EpisodeCount: r.EpisodeCount,
			Online:       r.Online,
			PublishAt:    publishAt,
			Categories:   r.Categories,
			Heat:         r.Heat,
			Note:         r.Note,
		})
	}
	return out
}
