package dto

import "github.com/beilunyang/openhongguo/sdk/model"

// FilterOption 是筛选页上的一个可选值。
type FilterOption struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// FilterRow 是短剧筛选页的一个选择行。
type FilterRow struct {
	Key     string         `json:"key"`
	Name    string         `json:"name"`
	Options []FilterOption `json:"options"`
}

// ShortFilterPage 是符合筛选条件的短剧的一页。
type ShortFilterPage struct {
	Items   []Series `json:"items"`
	HasMore bool     `json:"hasMore"`
	Cursor  string   `json:"cursor"`
}

// FromFilterRows 转换筛选面板。
func FromFilterRows(rows []model.FilterRow) []FilterRow {
	out := make([]FilterRow, 0, len(rows))
	for _, r := range rows {
		row := FilterRow{Key: r.Key, Name: r.Name}
		row.Options = make([]FilterOption, 0, len(r.Options))
		for _, o := range r.Options {
			row.Options = append(row.Options, FilterOption{Name: o.Name, Value: o.Value})
		}
		out = append(out, row)
	}
	return out
}

// FromShortFilterPage 转换筛选结果。
func FromShortFilterPage(p model.ShortFilterPage) ShortFilterPage {
	return ShortFilterPage{
		Items:   FromSeriesPage(p.Items, p.Cursor, p.HasMore).Items,
		HasMore: p.HasMore,
		Cursor:  p.Cursor,
	}
}
