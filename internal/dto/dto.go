// Package dto 定义暴露给前端的传输结构，避免前端直接依赖 SDK 的 JSON 形态。
package dto

import "github.com/beilunyang/openhongguo/sdk/model"

// Series 是列表页里的一部短剧。
type Series struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	CoverURL     string `json:"coverUrl"`
	EpisodeCount int    `json:"episodeCount"`
}

// SeriesPage 是一页短剧列表。
type SeriesPage struct {
	Items   []Series `json:"items"`
	Cursor  string   `json:"cursor"`
	HasMore bool     `json:"hasMore"`
}

// FromSeriesSummary 把 SDK 的短剧摘要转成前端结构。
func FromSeriesSummary(s model.SeriesSummary) Series {
	return Series{
		ID:           s.ID,
		Title:        s.Title,
		Description:  s.Description,
		CoverURL:     s.CoverURL,
		EpisodeCount: s.EpisodeCount,
	}
}

// FromSeriesPage 把 SDK 的一页短剧转成前端结构。
func FromSeriesPage(items []model.SeriesSummary, cursor string, hasMore bool) SeriesPage {
	out := make([]Series, 0, len(items))
	for _, it := range items {
		out = append(out, FromSeriesSummary(it))
	}
	return SeriesPage{Items: out, Cursor: cursor, HasMore: hasMore}
}
