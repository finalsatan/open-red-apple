package dto

import "github.com/beilunyang/openhongguo/sdk/model"

// InfoRow 是详情页上带标签的一行。
type InfoRow struct {
	Kind int    `json:"kind"`
	Text string `json:"text"`
}

// Series 是列表页或详情页里的一部短剧。
type Series struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	CoverURL      string    `json:"coverUrl"`
	EpisodeCount  int       `json:"episodeCount"`
	Status        int       `json:"status"`
	PayType       int       `json:"payType"`
	ContentType   int       `json:"contentType"`
	Platform      int       `json:"platform"`
	Heat          int64     `json:"heat"`
	PlayCount     int64     `json:"playCount"`
	FollowCount   int64     `json:"followCount"`
	DiggCount     int64     `json:"diggCount"`
	RightText     string    `json:"rightText"`
	Subtitles     []string  `json:"subtitles"`
	IsMultiSeason bool      `json:"isMultiSeason"`
	SeasonIndex   int       `json:"seasonIndex"`
	Rows          []InfoRow `json:"rows"`
}

// SeriesPage 是一页短剧列表。
type SeriesPage struct {
	Items   []Series `json:"items"`
	Cursor  string   `json:"cursor"`
	HasMore bool     `json:"hasMore"`
}

// Episode 是一集。
type Episode struct {
	ID       string `json:"id"`
	Number   int    `json:"number"`
	Title    string `json:"title"`
	Playable bool   `json:"playable"`
	Vertical bool   `json:"vertical"`
	IsNew    bool   `json:"isNew"`
	CoverURL string `json:"coverUrl"`
}

// EpisodePage 是一页分集。
type EpisodePage struct {
	Items   []Episode `json:"items"`
	Cursor  string    `json:"cursor"`
	HasMore bool      `json:"hasMore"`
}

// Season 是合集里的一季。
type Season struct {
	SeriesID     string `json:"seriesId"`
	Title        string `json:"title"`
	Index        int    `json:"index"`
	CoverURL     string `json:"coverUrl"`
	EpisodeCount int    `json:"episodeCount"`
	Status       int    `json:"status"`
}

// Album 是多季合集。
type Album struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	CoverURL    string   `json:"coverUrl"`
	SeasonCount int      `json:"seasonCount"`
	Seasons     []Season `json:"seasons"`
}

// Recommend 是一部短剧的相关推荐。
type Recommend struct {
	CellTitle string   `json:"cellTitle"`
	Items     []Series `json:"items"`
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

// FromSeries 把 SDK 的短剧详情转成前端结构。
func FromSeries(s model.Series) Series {
	out := FromSeriesSummary(s.SeriesSummary)
	out.Status = int(s.Status)
	out.PayType = int(s.PayType)
	out.ContentType = int(s.ContentType)
	out.Platform = int(s.Platform)
	out.Heat = s.Heat
	out.PlayCount = s.PlayCount
	out.FollowCount = s.FollowCount
	out.DiggCount = s.DiggCount
	out.RightText = s.RightText
	out.Subtitles = s.Subtitles
	out.IsMultiSeason = s.IsMultiSeason
	out.SeasonIndex = s.SeasonIndex
	out.Rows = make([]InfoRow, 0, len(s.Rows))
	for _, r := range s.Rows {
		out.Rows = append(out.Rows, InfoRow{Kind: int(r.Kind), Text: r.Text})
	}
	return out
}

// FromSeriesPage 把 SDK 的一页短剧转成前端结构。
func FromSeriesPage(items []model.SeriesSummary, cursor string, hasMore bool) SeriesPage {
	out := make([]Series, 0, len(items))
	for _, it := range items {
		out = append(out, FromSeriesSummary(it))
	}
	return SeriesPage{Items: out, Cursor: cursor, HasMore: hasMore}
}

// FromEpisode 把 SDK 的分集转成前端结构。
func FromEpisode(e model.Episode) Episode {
	return Episode{
		ID:       e.ID,
		Number:   e.Number,
		Title:    e.Title,
		Playable: e.Playable,
		Vertical: e.Vertical,
		IsNew:    e.IsNew,
		CoverURL: e.CoverURL,
	}
}

// FromEpisodePage 把 SDK 的一页分集转成前端结构。
func FromEpisodePage(items []model.Episode, cursor string, hasMore bool) EpisodePage {
	out := make([]Episode, 0, len(items))
	for _, it := range items {
		out = append(out, FromEpisode(it))
	}
	return EpisodePage{Items: out, Cursor: cursor, HasMore: hasMore}
}

// FromAlbum 把 SDK 的合集转成前端结构。
func FromAlbum(a model.Album) Album {
	out := Album{
		ID:          a.ID,
		Title:       a.Title,
		Description: a.Description,
		CoverURL:    a.CoverURL,
		SeasonCount: a.SeasonCount,
	}
	out.Seasons = make([]Season, 0, len(a.Seasons))
	for _, s := range a.Seasons {
		out.Seasons = append(out.Seasons, Season{
			SeriesID:     s.SeriesID,
			Title:        s.Title,
			Index:        s.Index,
			CoverURL:     s.CoverURL,
			EpisodeCount: s.EpisodeCount,
			Status:       int(s.Status),
		})
	}
	return out
}

// FromRecommend 把 SDK 的相关推荐转成前端结构。
func FromRecommend(r model.Recommend) Recommend {
	out := Recommend{CellTitle: r.CellTitle}
	out.Items = make([]Series, 0, len(r.Items))
	for _, it := range r.Items {
		out.Items = append(out.Items, FromSeriesSummary(it))
	}
	return out
}
