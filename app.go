package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/beilunyang/openhongguo/adapters"
	"github.com/beilunyang/openhongguo/adapters/memstore"
	"github.com/beilunyang/openhongguo/sdk"
	"github.com/beilunyang/openhongguo/sdk/model"

	"github.com/finalsatan/open-red-apple/internal/dto"
	"github.com/finalsatan/open-red-apple/internal/gateway"
	"github.com/finalsatan/open-red-apple/internal/ohg"
	"github.com/finalsatan/open-red-apple/internal/paths"
	"github.com/finalsatan/open-red-apple/internal/store"
)

// App 持有 Wails 上下文、SDK 门面和本地存储。
type App struct {
	ctx          context.Context
	svc          *ohg.Service
	store        *store.Store
	initErr      error
	media        *gateway.Gateway
	playClient   playbackClient
	playGateOnce sync.Once
	playGate     chan struct{}
	requestMu    sync.Mutex
	playRequests map[string]*playRequest
	playClosing  bool
}

type playRequest struct {
	ctx     context.Context
	cancel  context.CancelFunc
	claimed bool
}

// playbackClient is kept separate from the browsing client because anonymous
// devices can be consumed by a playback lookup.
type playbackClient interface {
	ResolvePlayback(context.Context, string, sdk.PlaybackOptions) (*model.Playback, error)
	ResetDevice(context.Context) error
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
	a.playClient, err = adapters.NewClient("", sdk.Options{ProfileStore: memstore.New(), RetryAttempts: 3})
	if err != nil {
		a.initErr = err
		return a
	}
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
	a.media, err = gateway.New()
	if err != nil {
		_ = st.Close()
		a.store = nil
		a.initErr = fmt.Errorf("start media gateway: %w", err)
		return a
	}
	return a
}

// startup 由 Wails 在启动时调用。
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// shutdown 由 Wails 在退出时调用。
func (a *App) shutdown(ctx context.Context) {
	a.requestMu.Lock()
	a.playClosing = true
	for _, request := range a.playRequests {
		request.cancel()
	}
	a.playRequests = nil
	a.requestMu.Unlock()
	if a.media != nil {
		_ = a.media.Close()
	}
	if a.store != nil {
		_ = a.store.Close()
	}
}

// PlayURL resolves a clear stream and returns a local URL.
func (a *App) PlayURL(episodeID string) (string, error) {
	if err := a.ready(); err != nil {
		return "", err
	}
	return a.playURL(a.context(), episodeID)
}

// BeginPlayback creates a cancellable preparation request for the player.
func (a *App) BeginPlayback() (string, error) {
	if err := a.ready(); err != nil {
		return "", err
	}
	var raw [24]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	id := hex.EncodeToString(raw[:])
	ctx, cancel := context.WithTimeout(a.context(), 2*time.Minute)
	request := &playRequest{ctx: ctx, cancel: cancel}
	a.requestMu.Lock()
	if a.playClosing {
		a.requestMu.Unlock()
		cancel()
		return "", fmt.Errorf("应用正在关闭")
	}
	if a.playRequests == nil {
		a.playRequests = make(map[string]*playRequest)
	}
	a.playRequests[id] = request
	a.requestMu.Unlock()
	go func() { <-ctx.Done(); a.removePlaybackRequest(id, request) }()
	return id, nil
}

// CancelPlayback stops a pending request, including one waiting for the device.
func (a *App) CancelPlayback(id string) {
	a.requestMu.Lock()
	request := a.playRequests[id]
	if request != nil {
		delete(a.playRequests, id)
	}
	a.requestMu.Unlock()
	if request != nil {
		request.cancel()
	}
}

func (a *App) removePlaybackRequest(id string, request *playRequest) {
	a.requestMu.Lock()
	if a.playRequests[id] == request {
		delete(a.playRequests, id)
	}
	a.requestMu.Unlock()
}

// PlayURLForRequest resolves a stream while the preparation request remains active.
func (a *App) PlayURLForRequest(episodeID, requestID string) (string, error) {
	if err := a.ready(); err != nil {
		return "", err
	}
	a.requestMu.Lock()
	request := a.playRequests[requestID]
	if request != nil {
		if request.claimed {
			request = nil
		} else {
			request.claimed = true
		}
	}
	a.requestMu.Unlock()
	if request == nil {
		return "", context.Canceled
	}
	defer func() { a.removePlaybackRequest(requestID, request); request.cancel() }()
	return a.playURL(request.ctx, episodeID)
}

func (a *App) playURL(ctx context.Context, episodeID string) (string, error) {
	p, err := a.resolvePlayback(ctx, episodeID)
	if err != nil {
		return "", err
	}
	if p.Protected {
		return "", fmt.Errorf("此分集为加密资源，暂不支持播放")
	}
	stream, err := selectClearStream(p)
	if err != nil {
		return "", fmt.Errorf("此分集没有可用的清晰播放资源")
	}
	if p.AudioURL != "" {
		return a.media.RegisterMux(ctx, stream.URL, stream.BackupURL, p.AudioURL, stream.HLS, p.DurationSeconds)
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return a.media.RegisterStream(stream.URL, stream.BackupURL, stream.HLS)
}

func (a *App) acquirePlayback(ctx context.Context) error {
	a.playGateOnce.Do(func() { a.playGate = make(chan struct{}, 1) })
	select {
	case a.playGate <- struct{}{}:
		if err := ctx.Err(); err != nil {
			<-a.playGate
			return err
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (a *App) resolvePlayback(ctx context.Context, episodeID string) (*model.Playback, error) {
	if err := a.acquirePlayback(ctx); err != nil {
		return nil, err
	}
	defer func() { <-a.playGate }()
	var last error
	for i := 0; i < 3; i++ {
		p, err := a.playClient.ResolvePlayback(ctx, episodeID, sdk.PlaybackOptions{})
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if errors.Is(err, sdk.ErrDRMProtected) {
			return nil, fmt.Errorf("此分集为加密资源，暂不支持播放")
		}
		if errors.Is(err, sdk.ErrAnonymousLimited) {
			last = err
			if i < 2 {
				if err = a.playClient.ResetDevice(ctx); err != nil {
					if ctx.Err() != nil {
						return nil, ctx.Err()
					}
					return nil, err
				}
			}
			continue
		}
		if err != nil {
			return nil, err
		}
		if p == nil {
			return nil, fmt.Errorf("未找到播放资源")
		}
		return p, nil
	}
	return nil, fmt.Errorf("匿名设备暂时无法获取播放资源，请稍后重试：%w", last)
}

// ReleasePlayback stops resources owned by a playback URL when leaving the player.
func (a *App) ReleasePlayback(url string) {
	if a.media != nil {
		a.media.Release(url)
	}
}

// Prefer H.264 where offered: WebView2 support for HEVC depends on the system.
func selectClearStream(p *model.Playback) (model.Stream, error) {
	filtered := *p
	filtered.Streams = nil
	var h264 []model.Stream
	for _, stream := range p.Streams {
		if stream.Protected || (stream.Dash && p.AudioURL == "") || strings.TrimSpace(stream.URL) == "" {
			continue
		}
		filtered.Streams = append(filtered.Streams, stream)
		codec := strings.ToLower(stream.Codec)
		if codec == "h264" || codec == "h.264" || codec == "avc" || strings.HasPrefix(codec, "avc1") {
			h264 = append(h264, stream)
		}
	}
	if len(h264) > 0 {
		filtered.Streams = h264
	}
	return filtered.Select("")
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
	if err := a.svc.ResetDevice(a.context()); err != nil {
		return err
	}
	ctx := a.context()
	if err := a.acquirePlayback(ctx); err != nil {
		return err
	}
	defer func() { <-a.playGate }()
	return a.playClient.ResetDevice(ctx)
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
