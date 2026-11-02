package main

import (
	"context"
	"fmt"

	"github.com/beilunyang/openhongguo/sdk"

	"github.com/finalsatan/open-red-apple/internal/dto"
	"github.com/finalsatan/open-red-apple/internal/ohg"
	"github.com/finalsatan/open-red-apple/internal/paths"
	"github.com/finalsatan/open-red-apple/internal/store"
)

// App 持有 Wails 上下文、SDK 门面和本地存储。
type App struct {
	ctx     context.Context
	svc     *ohg.Service
	store   *store.Store
	initErr error
}

// NewApp 创建 App，并初始化 SDK 客户端与数据库。
func NewApp() *App {
	a := &App{}
	profile, err := paths.ProfilePath()
	if err != nil {
		a.initErr = fmt.Errorf("resolve profile path: %w", err)
		return a
	}
	svc, err := ohg.NewService(profile)
	if err != nil {
		a.initErr = fmt.Errorf("create client: %w", err)
		return a
	}
	a.svc = svc

	dbPath, err := paths.DBPath()
	if err != nil {
		a.initErr = fmt.Errorf("resolve db path: %w", err)
		return a
	}
	st, err := store.Open(dbPath)
	if err != nil {
		a.initErr = fmt.Errorf("open store: %w", err)
		return a
	}
	a.store = st
	return a
}

// startup 由 Wails 在启动时调用。
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// shutdown 由 Wails 在退出时调用。
func (a *App) shutdown(ctx context.Context) {
	if a.store != nil {
		_ = a.store.Close()
	}
}

func (a *App) context() context.Context {
	if a.ctx != nil {
		return a.ctx
	}
	return context.Background()
}

func (a *App) ready() error {
	if a.initErr != nil {
		return a.initErr
	}
	if a.svc == nil {
		return fmt.Errorf("service not initialised")
	}
	return nil
}

// --- 发现与搜索 ---

// Discover 加载书城的一个频道（空频道即推荐）。
func (a *App) Discover(channel, cursor string) (*dto.SeriesPage, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	page, err := a.svc.Discover(a.context(), sdk.DiscoverOptions{Channel: channel, Cursor: cursor})
	if err != nil {
		return nil, err
	}
	out := dto.FromSeriesPage(page.Items, page.Cursor, page.HasMore)
	return &out, nil
}

// Search 搜索短剧。
func (a *App) Search(query, cursor, tab string, filters []string) (*dto.SeriesPage, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	page, err := a.svc.Search(a.context(), query, sdk.SearchOptions{
		Cursor:  cursor,
		Tab:     tab,
		Filters: filters,
	})
	if err != nil {
		return nil, err
	}
	out := dto.FromSeriesPage(page.Items, page.Cursor, page.HasMore)
	return &out, nil
}

// Suggest 返回某个前缀的搜索联想词。
func (a *App) Suggest(prefix string) ([]dto.Suggestion, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	items, err := a.svc.Suggest(a.context(), prefix)
	if err != nil {
		return nil, err
	}
	out := make([]dto.Suggestion, 0, len(items))
	for _, s := range items {
		out = append(out, dto.FromSuggestion(s))
	}
	return out, nil
}

// HotSearch 加载搜索页的热搜。
func (a *App) HotSearch() (*dto.HotSearch, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	h, err := a.svc.HotSearch(a.context())
	if err != nil {
		return nil, err
	}
	out := dto.FromHotSearch(*h)
	return &out, nil
}

// SearchCues 加载搜索框里轮播的提示词。
func (a *App) SearchCues() ([]string, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.svc.SearchCues(a.context())
}

// --- 短剧与分集 ---

// GetSeries 加载一部短剧的详情。
func (a *App) GetSeries(seriesID string) (*dto.Series, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	s, err := a.svc.GetSeries(a.context(), seriesID)
	if err != nil {
		return nil, err
	}
	out := dto.FromSeries(*s)
	return &out, nil
}

// ListEpisodes 加载一部短剧的分集。
func (a *App) ListEpisodes(seriesID, cursor string) (*dto.EpisodePage, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	page, err := a.svc.ListEpisodes(a.context(), seriesID, sdk.EpisodeOptions{Cursor: cursor})
	if err != nil {
		return nil, err
	}
	out := dto.FromEpisodePage(page.Items, page.Cursor, page.HasMore)
	return &out, nil
}

// Seasons 列出多季短剧的全部季。
func (a *App) Seasons(seriesID string) (*dto.Album, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	album, err := a.svc.Seasons(a.context(), seriesID)
	if err != nil {
		return nil, err
	}
	out := dto.FromAlbum(*album)
	return &out, nil
}

// SeriesRecommend 加载一部短剧的相关推荐。
func (a *App) SeriesRecommend(seriesID string) (*dto.Recommend, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	r, err := a.svc.SeriesRecommend(a.context(), seriesID)
	if err != nil {
		return nil, err
	}
	out := dto.FromRecommend(*r)
	return &out, nil
}

// --- 榜单与筛选 ---

// ShortRank 加载短剧排行榜的一页。
func (a *App) ShortRank(content, board, cursor string) (*dto.ShortRanking, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	r, err := a.svc.ShortRank(a.context(), sdk.ShortRankQuery{
		Content: content,
		Board:   board,
		Cursor:  cursor,
	})
	if err != nil {
		return nil, err
	}
	out := dto.FromShortRanking(*r)
	return &out, nil
}

// ShortFilterPanel 加载短剧筛选面板。
func (a *App) ShortFilterPanel(genre string) ([]dto.FilterRow, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	rows, err := a.svc.ShortFilterPanel(a.context(), genre)
	if err != nil {
		return nil, err
	}
	return dto.FromFilterRows(rows), nil
}

// ShortFilter 按条件筛选短剧。
func (a *App) ShortFilter(selections map[string][]string, cursor string) (*dto.ShortFilterPage, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	page, err := a.svc.ShortFilter(a.context(), sdk.ShortFilterQuery{
		Selections: selections,
		Cursor:     cursor,
	})
	if err != nil {
		return nil, err
	}
	out := dto.FromShortFilterPage(*page)
	return &out, nil
}

// --- 演员与上新 ---

// Person 加载一位演员的主页。
func (a *App) Person(userID string) (*dto.Person, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	p, err := a.svc.Person(a.context(), userID)
	if err != nil {
		return nil, err
	}
	out := dto.FromPerson(*p)
	return &out, nil
}

// PersonWorks 加载一位演员参演的短剧。
func (a *App) PersonWorks(userID, cursor string) (*dto.WorkPage, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	page, err := a.svc.PersonWorks(a.context(), userID, cursor)
	if err != nil {
		return nil, err
	}
	out := dto.FromWorkPage(*page)
	return &out, nil
}

// NewReleases 加载上新日历的一页。
func (a *App) NewReleases(date, cursor string) (*dto.NewReleasePage, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	page, err := a.svc.NewReleases(a.context(), sdk.NewReleaseQuery{Date: date, Cursor: cursor})
	if err != nil {
		return nil, err
	}
	out := dto.FromNewReleasePage(*page)
	return &out, nil
}

// ResetDevice 更换匿名设备。
func (a *App) ResetDevice() error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.svc.ResetDevice(a.context())
}

// --- 本地库：搜索记录 / 收藏 / 观看记录 ---

// AddSearchKeyword 记录一个搜索词。
func (a *App) AddSearchKeyword(keyword string) error {
	if a.store == nil {
		return a.ready()
	}
	return a.store.AddSearch(keyword)
}

// SearchHistory 返回最近的搜索词。
func (a *App) SearchHistory() ([]string, error) {
	if a.store == nil {
		return nil, a.ready()
	}
	return a.store.SearchHistory(20)
}

// ClearSearchHistory 清空搜索记录。
func (a *App) ClearSearchHistory() error {
	if a.store == nil {
		return a.ready()
	}
	return a.store.ClearSearchHistory()
}

// ToggleFavorite 切换收藏状态，返回切换后是否已收藏。
func (a *App) ToggleFavorite(seriesID, title, coverURL string) (bool, error) {
	if a.store == nil {
		return false, a.ready()
	}
	return a.store.ToggleFavorite(dto.Favorite{SeriesID: seriesID, Title: title, CoverURL: coverURL})
}

// IsFavorite 报告一部短剧是否已收藏。
func (a *App) IsFavorite(seriesID string) (bool, error) {
	if a.store == nil {
		return false, a.ready()
	}
	return a.store.IsFavorite(seriesID)
}

// ListFavorites 返回收藏。
func (a *App) ListFavorites() ([]dto.Favorite, error) {
	if a.store == nil {
		return nil, a.ready()
	}
	return a.store.ListFavorites()
}

// SaveProgress 写入观看进度。
func (a *App) SaveProgress(item dto.HistoryItem) error {
	if a.store == nil {
		return a.ready()
	}
	return a.store.SaveProgress(item)
}

// GetProgress 返回一部短剧的观看进度；没有记录时返回 nil。
func (a *App) GetProgress(seriesID string) (*dto.HistoryItem, error) {
	if a.store == nil {
		return nil, a.ready()
	}
	return a.store.GetProgress(seriesID)
}

// ListHistory 返回观看记录。
func (a *App) ListHistory() ([]dto.HistoryItem, error) {
	if a.store == nil {
		return nil, a.ready()
	}
	return a.store.ListHistory(50)
}

// DeleteHistory 删除一部短剧的观看记录。
func (a *App) DeleteHistory(seriesID string) error {
	if a.store == nil {
		return a.ready()
	}
	return a.store.DeleteHistory(seriesID)
}

// ClearHistory 清空观看记录。
func (a *App) ClearHistory() error {
	if a.store == nil {
		return a.ready()
	}
	return a.store.ClearHistory()
}
