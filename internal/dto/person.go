package dto

import "github.com/beilunyang/openhongguo/sdk/model"

// Person 是演员主页。
type Person struct {
	UserID      string   `json:"userId"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	AvatarURL   string   `json:"avatarUrl"`
	Tags        []string `json:"tags"`
	Labels      []string `json:"labels"`
	Followers   int64    `json:"followers"`
	Likes       int64    `json:"likes"`
}

// Work 是演员参演的一部短剧。
type Work struct {
	SeriesID     string   `json:"seriesId"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	CoverURL     string   `json:"coverUrl"`
	EpisodeCount int      `json:"episodeCount"`
	Tags         []string `json:"tags"`
	Heat         int64    `json:"heat"`
	Plays        int64    `json:"plays"`
}

// WorkPage 是一页作品。
type WorkPage struct {
	Items   []Work `json:"items"`
	HasMore bool   `json:"hasMore"`
	Cursor  string `json:"cursor"`
}

// FromPerson 转换演员主页。
func FromPerson(p model.Person) Person {
	return Person{
		UserID:      p.UserID,
		Name:        p.Name,
		Description: p.Description,
		AvatarURL:   p.AvatarURL,
		Tags:        p.Tags,
		Labels:      p.Labels,
		Followers:   p.Followers,
		Likes:       p.Likes,
	}
}

// FromWorkPage 转换一页作品。
func FromWorkPage(p model.WorkPage) WorkPage {
	out := WorkPage{HasMore: p.HasMore, Cursor: p.Cursor}
	out.Items = make([]Work, 0, len(p.Items))
	for _, w := range p.Items {
		out.Items = append(out.Items, Work{
			SeriesID:     w.SeriesID,
			Title:        w.Title,
			Description:  w.Description,
			CoverURL:     w.CoverURL,
			EpisodeCount: w.EpisodeCount,
			Tags:         w.Tags,
			Heat:         w.Heat,
			Plays:        w.Plays,
		})
	}
	return out
}
