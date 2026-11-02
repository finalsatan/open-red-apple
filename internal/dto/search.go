package dto

import "github.com/beilunyang/openhongguo/sdk/model"

// Suggestion 是一条搜索联想词。
type Suggestion struct {
	Text    string `json:"text"`
	Keyword string `json:"keyword"`
	Kind    int    `json:"kind"`
}

// HotItem 是热榜上的一部作品。
type HotItem struct {
	Position     int    `json:"position"`
	SeriesID     string `json:"seriesId"`
	Title        string `json:"title"`
	CoverURL     string `json:"coverUrl"`
	EpisodeCount int    `json:"episodeCount"`
	SearchScore  string `json:"searchScore"`
	Heat         string `json:"heat"`
	Tag          string `json:"tag"`
}

// HotTopic 是热点话题榜的一项。
type HotTopic struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Query string `json:"query"`
	Heat  string `json:"heat"`
}

// HotBoard 是一个热榜。
type HotBoard struct {
	Name   string     `json:"name"`
	Items  []HotItem  `json:"items"`
	Topics []HotTopic `json:"topics"`
}

// HotSearch 是搜索页的热搜内容。
type HotSearch struct {
	Guesses []HotItem  `json:"guesses"`
	Boards  []HotBoard `json:"boards"`
}

// FromSuggestion 转换搜索联想词。
func FromSuggestion(s model.Suggestion) Suggestion {
	return Suggestion{Text: s.Text, Keyword: s.Keyword, Kind: s.Kind}
}

func fromHotItem(h model.HotItem) HotItem {
	return HotItem{
		Position:     h.Position,
		SeriesID:     h.SeriesID,
		Title:        h.Title,
		CoverURL:     h.CoverURL,
		EpisodeCount: h.EpisodeCount,
		SearchScore:  h.SearchScore,
		Heat:         h.Heat,
		Tag:          h.Tag,
	}
}

// FromHotSearch 转换搜索页的热搜内容。
func FromHotSearch(h model.HotSearch) HotSearch {
	out := HotSearch{}
	out.Guesses = make([]HotItem, 0, len(h.Guesses))
	for _, g := range h.Guesses {
		out.Guesses = append(out.Guesses, fromHotItem(g))
	}
	out.Boards = make([]HotBoard, 0, len(h.Boards))
	for _, b := range h.Boards {
		board := HotBoard{Name: b.Name}
		board.Items = make([]HotItem, 0, len(b.Items))
		for _, it := range b.Items {
			board.Items = append(board.Items, fromHotItem(it))
		}
		board.Topics = make([]HotTopic, 0, len(b.Topics))
		for _, t := range b.Topics {
			board.Topics = append(board.Topics, HotTopic{ID: t.ID, Title: t.Title, Query: t.Query, Heat: t.Heat})
		}
		out.Boards = append(out.Boards, board)
	}
	return out
}
