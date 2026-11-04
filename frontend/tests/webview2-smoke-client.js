// Injected only into a temporary smoke-test build, never the normal application.
(async () => {
    const control = window.__WRA_SMOKE_ORIGIN__;
    async function report(result) {
        await fetch(`${control}/report`, { method: 'POST', headers: { 'Content-Type': 'text/plain' }, body: JSON.stringify(result) });
    }
    async function until(description, predicate, timeout = 20000) {
        const end = Date.now() + timeout;
        while (Date.now() < end) {
            const value = await predicate();
            if (value) return value;
            await new Promise((resolve) => setTimeout(resolve, 100));
        }
        throw new Error(`等待超时：${description}`);
    }
    try {
        if (!control) throw new Error('缺少测试控制地址');
        const api = await until('Wails 绑定', () => window.go?.main?.App);
        await until('应用界面', () => document.querySelector('#root > *'));
        const configResponse = await fetch(`${control}/config`);
        if (!configResponse.ok) throw new Error('无法读取测试配置');
        const { rpc: rpcOrigin } = await configResponse.json();
        const rpc = async (method, fields = {}) => {
            const response = await fetch(`${rpcOrigin}/rpc`, {
                method: 'POST', headers: { 'Content-Type': 'text/plain' }, body: JSON.stringify({ method, ...fields }),
            });
            if (!response.ok) throw new Error(await response.text());
            return response.json();
        };
        // Catalog and remote playback resolution use local fixtures. History
        // keeps the actual Wails bindings and the app's SQLite database.
        api.GetSeries = async () => ({ id: 'series', title: '测试剧集', coverUrl: '' });
        api.ListEpisodes = async () => ({ items: [{ id: 'ep1', number: 1, playable: true }, { id: 'ep2', number: 2, playable: true }], hasMore: false, cursor: '' });
        api.BeginPlayback = () => rpc('BeginPlayback');
        api.CancelPlayback = (requestId) => rpc('CancelPlayback', { requestId });
        api.PlayURLForRequest = (id, requestId) => rpc('PlayURLForRequest', { id, requestId });
        api.ReleasePlayback = (url) => rpc('ReleasePlayback', { url });
        await api.SaveProgress({ seriesId: 'series', episodeId: 'ep1', episodeNumber: 1, title: '测试剧集', coverUrl: '', position: 4, duration: 12, finished: false, updatedAt: 0 });
        window.location.hash = '#/play/series/ep1';
        let video = await until('MP4 元数据', () => { const v = document.querySelector('video'); return v?.readyState >= 1 ? v : null; });
        await until('续播位置', () => video.currentTime >= 3.9);
        video.muted = true;
        await video.play();
        video.currentTime = 8;
        await until('进度保存', async () => (await api.GetProgress('series'))?.position >= 7.9);
        video.currentTime = video.duration - 0.2;
        await video.play();
        const previousVideo = video;
        await until('自动下一集', () => window.location.hash === '#/play/series/ep2');
        video = await until('HLS 元数据', () => { const v = document.querySelector('video'); return v !== previousVideo && v?.readyState >= 1 ? v : null; });
        video.muted = true;
        await video.play();
        video.currentTime = 6;
        await until('HLS 拖动', () => video.currentTime >= 5.9);
        await until('下一集记录', async () => (await api.GetProgress('series'))?.episodeId === 'ep2');
        await report({ ok: true, checks: ['mp4-resume', 'mp4-seek', 'sqlite-progress', 'automatic-next', 'hls-seek'] });
    } catch (err) {
        try { await report({ ok: false, error: String(err?.stack || err) }); } catch (reportErr) { console.error('WebView smoke report failed', err, reportErr); }
    }
})();
