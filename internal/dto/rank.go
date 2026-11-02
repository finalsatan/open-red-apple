package dto

import "github.com/beilunyang/openhongguo/sdk/model"

// RankBoard 是某个内容下的一个榜。
type RankBoard struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Selected bool   `json:"selected"`
}

// RankContent 是短剧榜单页的一个内容（全部、真人剧、漫剧……）。
type RankContent struct {
	ID       string      `json:"id"`
	Name     string      `json:"name"`
	Selected bool        `json:"selected"`
	Boards   []RankBoard `json:"boards"`
}

// RankChange 是榜单行的名次变动。
type RankChange struct {
	Kind   int `json:"kind"`
	Places int `json:"places"`
}

// RankItem 是榜单的一行。
type RankItem struct {
	Rank         int         `json:"rank"`
	Position     int         `json:"position"`
	SeriesID     string      `json:"seriesId"`
	Title        string      `json:"title"`
	CoverURL     string      `json:"coverUrl"`
	Score        string      `json:"score"`
	EpisodeCount int         `json:"episodeCount"`
	Promoted     bool        `json:"promoted"`
	Actor        string      `json:"actor"`
	UserID       string      `json:"userId"`
	AvatarURL    string      `json:"avatarUrl"`
	Intro        string      `json:"intro"`
	Details      []string    `json:"details"`
	Change       *RankChange `json:"change"`
}

// ShortRanking 是短剧榜单的一页。
type ShortRanking struct {
	Contents []RankContent `json:"contents"`
	Content  string        `json:"content"`
	Board    string        `json:"board"`
	Items    []RankItem    `json:"items"`
	HasMore  bool          `json:"hasMore"`
	Cursor   string        `json:"cursor"`
}

// FromShortRanking 转换短剧榜单。
func FromShortRanking(r model.ShortRanking) ShortRanking {
	out := ShortRanking{Content: r.Content, Board: r.Board, HasMore: r.HasMore, Cursor: r.Cursor}
	out.Contents = make([]RankContent, 0, len(r.Contents))
	for _, c := range r.Contents {
		content := RankContent{ID: c.ID, Name: c.Name, Selected: c.Selected}
		content.Boards = make([]RankBoard, 0, len(c.Boards))
		for _, b := range c.Boards {
			content.Boards = append(content.Boards, RankBoard{ID: b.ID, Name: b.Name, Selected: b.Selected})
		}
		out.Contents = append(out.Contents, content)
	}
	out.Items = make([]RankItem, 0, len(r.Items))
	for _, it := range r.Items {
		item := RankItem{
			Rank:         it.Rank,
			Position:     it.Position,
			SeriesID:     it.SeriesID,
			Title:        it.Title,
			CoverURL:     it.CoverURL,
			Score:        it.Score,
			EpisodeCount: it.EpisodeCount,
			Promoted:     it.Promoted,
			Actor:        it.Actor,
			UserID:       it.UserID,
			AvatarURL:    it.AvatarURL,
			Intro:        it.Intro,
			Details:      it.Details,
		}
		if it.Change != nil {
			item.Change = &RankChange{Kind: int(it.Change.Kind), Places: it.Change.Places}
		}
		out.Items = append(out.Items, item)
	}
	return out
}
