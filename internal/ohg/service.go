// Package ohg 封装 openhongguo SDK，供 Wails 后端调用。
//
// 读取类接口使用一个主 Client；播放解析在后续里程碑接入设备池。
package ohg

import (
	"context"

	"github.com/beilunyang/openhongguo/adapters"
	"github.com/beilunyang/openhongguo/sdk"
	"github.com/beilunyang/openhongguo/sdk/model"
)

// Service 是 SDK 的门面。
type Service struct {
	client *sdk.Client
}

// NewService 用 profilePath 保存匿名设备档案，创建客户端。
func NewService(profilePath string) (*Service, error) {
	client, err := adapters.NewClient(profilePath, sdk.Options{
		RetryAttempts: 3,
	})
	if err != nil {
		return nil, err
	}
	return &Service{client: client}, nil
}

// Discover 加载书城的一个频道。
func (s *Service) Discover(ctx context.Context, opts sdk.DiscoverOptions) (*sdk.Page[model.SeriesSummary], error) {
	return s.client.Discover(ctx, opts)
}

// Search 搜索短剧。
func (s *Service) Search(ctx context.Context, query string, opts sdk.SearchOptions) (*sdk.Page[model.SeriesSummary], error) {
	return s.client.Search(ctx, query, opts)
}

// Suggest 返回某个前缀的搜索联想词。
func (s *Service) Suggest(ctx context.Context, prefix string) ([]model.Suggestion, error) {
	return s.client.Suggest(ctx, prefix)
}

// HotSearch 加载搜索页的热搜。
func (s *Service) HotSearch(ctx context.Context) (*model.HotSearch, error) {
	return s.client.HotSearch(ctx)
}

// GetSeries 加载一部短剧的详情。
func (s *Service) GetSeries(ctx context.Context, seriesID string) (*model.Series, error) {
	return s.client.GetSeries(ctx, seriesID)
}

// ListEpisodes 加载一部短剧的分集列表。
func (s *Service) ListEpisodes(ctx context.Context, seriesID string, opts sdk.EpisodeOptions) (*sdk.Page[model.Episode], error) {
	return s.client.ListEpisodes(ctx, seriesID, opts)
}

// SeriesRecommend 加载一部短剧的相关推荐。
func (s *Service) SeriesRecommend(ctx context.Context, seriesID string) (*model.Recommend, error) {
	return s.client.SeriesRecommend(ctx, seriesID)
}

// Seasons 列出多季短剧的全部季。
func (s *Service) Seasons(ctx context.Context, seriesID string) (*model.Album, error) {
	return s.client.Seasons(ctx, seriesID)
}

// ShortRank 加载短剧排行榜的一页。
func (s *Service) ShortRank(ctx context.Context, q sdk.ShortRankQuery) (*model.ShortRanking, error) {
	return s.client.ShortRank(ctx, q)
}

// Person 加载一位演员的主页。
func (s *Service) Person(ctx context.Context, userID string) (*model.Person, error) {
	return s.client.Person(ctx, userID)
}

// ResolvePlayback 加载一集的播放模型。
func (s *Service) ResolvePlayback(ctx context.Context, episodeID string) (*model.Playback, error) {
	return s.client.ResolvePlayback(ctx, episodeID, sdk.PlaybackOptions{})
}

// ResetDevice 丢弃当前匿名设备并注册一台新设备。
func (s *Service) ResetDevice(ctx context.Context) error {
	return s.client.ResetDevice(ctx)
}
