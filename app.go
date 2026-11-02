package main

import (
	"context"
	"fmt"

	"github.com/beilunyang/openhongguo/sdk"

	"github.com/finalsatan/open-red-apple/internal/dto"
	"github.com/finalsatan/open-red-apple/internal/ohg"
	"github.com/finalsatan/open-red-apple/internal/paths"
)

// App 持有 Wails 上下文和 SDK 门面。
type App struct {
	ctx     context.Context
	svc     *ohg.Service
	initErr error
}

// NewApp 创建 App，并初始化 SDK 客户端。
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
	return a
}

// startup 由 Wails 在启动时调用。
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
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

// Discover 加载书城的一个频道（空频道即推荐）。
func (a *App) Discover(channel string) (*dto.SeriesPage, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	page, err := a.svc.Discover(a.context(), sdk.DiscoverOptions{Channel: channel})
	if err != nil {
		return nil, err
	}
	out := dto.FromSeriesPage(page.Items, page.Cursor, page.HasMore)
	return &out, nil
}

// Search 搜索短剧。
func (a *App) Search(query string) (*dto.SeriesPage, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	page, err := a.svc.Search(a.context(), query, sdk.SearchOptions{})
	if err != nil {
		return nil, err
	}
	out := dto.FromSeriesPage(page.Items, page.Cursor, page.HasMore)
	return &out, nil
}

// ResetDevice 更换匿名设备。
func (a *App) ResetDevice() error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.svc.ResetDevice(a.context())
}
