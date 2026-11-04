import { test, expect } from '@playwright/test';
import { execFileSync, spawn } from 'node:child_process';
import { createReadStream, mkdtempSync, readFileSync, rmSync, statSync } from 'node:fs';
import { createServer } from 'node:http';
import { tmpdir } from 'node:os';
import { join } from 'node:path';

let mediaDir;
let server;
let mediaOrigin;
let publishAll = false;
let harness;
let harnessOrigin;
const smokeReports = [];

test.beforeAll(async ({}, testInfo) => {
    testInfo.setTimeout(120_000);
    mediaDir = mkdtempSync(join(tmpdir(), 'win-red-apple-e2e-'));
    // These generated files are deterministic local fixtures, not external media.
    execFileSync('ffmpeg', ['-hide_banner', '-loglevel', 'error', '-y', '-f', 'lavfi', '-i', 'testsrc2=size=320x180:rate=24', '-t', '12', '-c:v', 'libx264', '-threads', '1', '-pix_fmt', 'yuv420p', '-g', '48', '-movflags', '+faststart', join(mediaDir, 'video.mp4')]);
    execFileSync('ffmpeg', ['-hide_banner', '-loglevel', 'error', '-y', '-i', join(mediaDir, 'video.mp4'), '-c', 'copy', '-hls_time', '2', '-hls_playlist_type', 'vod', '-hls_segment_filename', join(mediaDir, 'seg%02d.ts'), join(mediaDir, 'index.m3u8')]);
    server = createServer((req, res) => {
        const parsed = new URL(req.url, 'http://localhost');
        const name = parsed.pathname.slice(1);
        if (name === 'config') {
            res.setHeader('Access-Control-Allow-Origin', '*');
            res.setHeader('Content-Type', 'application/json');
            res.end(JSON.stringify({ rpc: harnessOrigin }));
            return;
        }
        if (name === 'report' && req.method === 'POST') {
            res.setHeader('Access-Control-Allow-Origin', '*');
            let body = '';
            req.setEncoding('utf8');
            req.on('data', (chunk) => { body += chunk; });
            req.on('end', () => { smokeReports.push(JSON.parse(body)); res.end('ok'); });
            return;
        }
        if (!/^(video\.mp4|index\.m3u8|seg\d+\.ts)$/.test(name)) { res.writeHead(404).end(); return; }
        const file = join(mediaDir, name);
        const size = statSync(file).size;
        res.setHeader('Access-Control-Allow-Origin', '*');
        res.setHeader('Access-Control-Expose-Headers', 'Content-Range, Accept-Ranges, Content-Length');
        res.setHeader('Accept-Ranges', 'bytes');
        res.setHeader('Content-Type', name.endsWith('.m3u8') ? 'application/vnd.apple.mpegurl' : name.endsWith('.ts') ? 'video/mp2t' : 'video/mp4');
        if (name === 'index.m3u8' && parsed.searchParams.has('mux')) {
            const playlist = readFileSync(file, 'utf8').replace('PLAYLIST-TYPE:VOD', 'PLAYLIST-TYPE:EVENT');
            const body = publishAll ? playlist : playlist.split('#EXTINF:')[0] + playlist.split('#EXTINF:').slice(1, 3).map((line) => '#EXTINF:' + line).join('');
            res.setHeader('Cache-Control', 'no-store');
            res.end(body);
            return;
        }
        let start = 0, end = size - 1;
        const range = /^bytes=(\d+)-(\d*)$/.exec(req.headers.range || '');
        if (range) {
            start = Number(range[1]);
            if (range[2]) end = Math.min(Number(range[2]), end);
            res.statusCode = 206;
            res.setHeader('Content-Range', `bytes ${start}-${end}/${size}`);
        }
        res.setHeader('Content-Length', end - start + 1);
        createReadStream(file, { start, end }).pipe(res);
    });
    await new Promise((resolve) => server.listen(0, '127.0.0.1', resolve));
    mediaOrigin = `http://127.0.0.1:${server.address().port}`;
    const binary = join(mediaDir, process.platform === 'win32' ? 'player-harness.exe' : 'player-harness');
    execFileSync('go', ['test', '-c', '-o', binary, '.'], { cwd: join(process.cwd(), '..'), stdio: 'pipe', timeout: 100_000 });
    harness = spawn(binary, ['-test.run=^TestPlayerHarness$', '-test.v'], {
        env: { ...process.env, WRA_E2E_MEDIA_ORIGIN: mediaOrigin },
        stdio: ['ignore', 'pipe', 'pipe'],
    });
    harnessOrigin = await new Promise((resolve, reject) => {
        let output = '';
        const timer = setTimeout(() => reject(new Error(`Go harness did not start: ${output}`)), 15_000);
        const collect = (data) => {
            output += data.toString();
            const match = /WRA_HARNESS_URL=(http:\/\/[^\s]+)/.exec(output);
            if (match) { clearTimeout(timer); resolve(match[1]); }
        };
        harness.stdout.on('data', collect);
        harness.stderr.on('data', collect);
        harness.once('error', (err) => { clearTimeout(timer); reject(err); });
        harness.once('exit', (code) => { clearTimeout(timer); reject(new Error(`Go harness exited ${code}: ${output}`)); });
    });
});

test.afterAll(async () => {
    if (harness && harness.exitCode === null) {
        const exited = new Promise((resolve) => harness.once('exit', resolve));
        if (harnessOrigin) await fetch(`${harnessOrigin}/shutdown`, { method: 'POST' }).catch(() => {});
        await Promise.race([exited, new Promise((resolve) => setTimeout(resolve, 3000))]);
        if (harness.exitCode === null) {
            harness.kill();
            await Promise.race([exited, new Promise((resolve) => setTimeout(resolve, 3000))]);
        }
    }
    if (server) await new Promise((resolve) => server.close(resolve));
    if (mediaDir) rmSync(mediaDir, { recursive: true, force: true });
});

async function harnessRPC(method, fields = {}) {
    const response = await fetch(`${harnessOrigin}/rpc`, {
        method: 'POST', headers: { 'Content-Type': 'text/plain' }, body: JSON.stringify({ method, ...fields }),
    });
    if (!response.ok) throw new Error(await response.text());
    return response.json();
}

async function mockBackend(page, { failFirst = false, mux = false, delay = 0, savedPosition, finishedDelay = 0 } = {}) {
    await page.addInitScript(({ origin, shouldFail, isMux, delayMs, position, finishedDelayMs }) => {
        const initial = { seriesId: 'series', episodeId: 'ep1', episodeNumber: 1, title: '测试剧集', coverUrl: '', position: position ?? (isMux ? 9 : 4), duration: 12, finished: false, updatedAt: 0 };
        if (!localStorage.getItem('progress')) localStorage.setItem('progress', JSON.stringify(initial));
        window.savedItems = [];
        window.releasedURLs = [];
        window.canceledRequests = [];
        window.playStarted = false;
        window.playCalls = 0;
        window.progressReads = [];
        window.finishedSaveStarted = false;
        let attempts = 0;
        window.go = { main: { App: {
            GetSeries: async () => ({ id: 'series', title: '测试剧集', coverUrl: '' }),
            GetProgress: async () => { const value = JSON.parse(localStorage.getItem('progress')); window.progressReads.push(value); return value; },
            ListEpisodes: async () => ({ items: [{ id: 'ep1', number: 1, playable: true }, { id: 'ep2', number: 2, playable: true }], hasMore: false, cursor: '' }),
            BeginPlayback: async () => `request-${window.playCalls + 1}`,
            CancelPlayback: async (request) => { window.canceledRequests.push(request); },
            PlayURLForRequest: async (id) => {
                window.playStarted = true;
                window.playCalls++;
                if (delayMs) await new Promise((resolve) => setTimeout(resolve, delayMs));
                if (shouldFail && attempts++ === 0) throw new Error('测试播放失败');
                return `${origin}/${isMux ? 'index.m3u8?mux=1&duration=12' : id === 'ep1' ? 'video.mp4' : 'index.m3u8'}`;
            },
            SaveProgress: async (item) => {
                if (item.finished) window.finishedSaveStarted = true;
                await new Promise((resolve) => setTimeout(resolve, item.finished ? Math.max(25, finishedDelayMs) : 25));
                localStorage.setItem('progress', JSON.stringify(item));
                window.savedItems.push(item);
            },
            ReleasePlayback: async (url) => { window.releasedURLs.push(url); },
        } } };
    }, { origin: mediaOrigin, shouldFail: failFirst, isMux: mux, delayMs: delay, position: savedPosition, finishedDelayMs: finishedDelay });
}

test('resumes MP4, saves seeks, plays the next HLS episode and releases the source', async ({ page }) => {
    await mockBackend(page);
    await page.goto('/#/play/series/ep1');
    await page.waitForFunction(() => document.querySelector('video')?.readyState >= 1);
    await expect.poll(() => page.locator('video').evaluate((v) => v.currentTime)).toBeGreaterThanOrEqual(3.9);
    await page.locator('video').evaluate((v) => { v.currentTime = 8; });
    await expect.poll(() => page.evaluate(() => JSON.parse(localStorage.getItem('progress')).position)).toBeGreaterThanOrEqual(7.9);
    await page.reload();
    await page.waitForFunction(() => document.querySelector('video')?.readyState >= 1);
    await expect.poll(() => page.locator('video').evaluate((v) => v.currentTime)).toBeGreaterThanOrEqual(7.9);
    await page.locator('video').evaluate((v) => { v.currentTime = v.duration - 0.2; return v.play(); });
    await expect(page).toHaveURL(/#\/play\/series\/ep2$/);
    await page.waitForFunction(() => document.querySelector('video')?.readyState >= 1);
    expect(await page.evaluate(() => window.savedItems.some((item) => item.episodeId === 'ep1' && item.finished))).toBe(true);
    expect(await page.evaluate(() => window.releasedURLs.some((url) => url.endsWith('/video.mp4')))).toBe(true);
    await page.locator('video').evaluate((v) => { v.currentTime = 6; });
    await expect.poll(() => page.locator('video').evaluate((v) => v.currentTime)).toBeGreaterThanOrEqual(5.9);
});

test('shows a playback error and can retry', async ({ page }) => {
    await mockBackend(page, { failFirst: true });
    await page.goto('/#/play/series/ep1');
    await expect(page.getByRole('alert')).toContainText('测试播放失败');
    await page.getByRole('button', { name: '重试' }).click();
    await page.waitForFunction(() => document.querySelector('video')?.readyState >= 1);
    await expect(page.getByRole('alert')).toHaveCount(0);
});

test('retry resumes from the most recent playback position', async ({ page }) => {
    await mockBackend(page);
    await page.goto('/#/play/series/ep1');
    await page.waitForFunction(() => document.querySelector('video')?.readyState >= 1);
    await expect.poll(() => page.locator('video').evaluate((v) => v.currentTime)).toBeGreaterThan(5.5);
    await page.locator('video').evaluate((v) => v.dispatchEvent(new Event('error')));
    await page.getByRole('button', { name: '重试' }).click();
    await page.waitForFunction(() => window.playCalls >= 2 && document.querySelector('video')?.readyState >= 1);
    expect(await page.evaluate(() => window.progressReads.at(-1).position)).toBeGreaterThan(5.4);
});

test('leaving the player saves the latest position before releasing it', async ({ page }) => {
    await mockBackend(page);
    await page.goto('/#/play/series/ep1');
    await page.waitForFunction(() => document.querySelector('video')?.readyState >= 1);
    await expect.poll(() => page.locator('video').evaluate((v) => v.currentTime)).toBeGreaterThan(5.5);
    await page.getByRole('link', { name: '⚙️ 设置' }).click();
    await expect.poll(() => page.evaluate(() => JSON.parse(localStorage.getItem('progress')).position)).toBeGreaterThan(5.4);
    await expect.poll(() => page.evaluate(() => window.releasedURLs.length)).toBe(1);
});

test('retry cancels a pending automatic episode change', async ({ page }) => {
    await mockBackend(page, { finishedDelay: 1200 });
    await page.goto('/#/play/series/ep1');
    await page.waitForFunction(() => document.querySelector('video')?.readyState >= 1);
    await page.locator('video').evaluate((v) => { v.currentTime = v.duration - 0.2; return v.play(); });
    await page.waitForFunction(() => window.finishedSaveStarted);
    await page.locator('video').evaluate((v) => v.dispatchEvent(new Event('error')));
    await page.getByRole('button', { name: '重试' }).click();
    await page.waitForFunction(() => window.playCalls >= 2 && document.querySelector('video')?.readyState >= 1);
    await expect(page).toHaveURL(/#\/play\/series\/ep1$/);
});

test('waits for a growing playlist before restoring the saved position', async ({ page }) => {
    publishAll = false;
    await mockBackend(page, { mux: true });
    await page.goto('/#/play/series/ep1');
    await expect(page.getByText('正在准备续播位置，请稍候…')).toBeVisible();
    expect(await page.locator('video').evaluate((v) => v.paused && v.duration < 9)).toBe(true);
    expect(await page.evaluate(() => JSON.parse(localStorage.getItem('progress')).position)).toBe(9);
    publishAll = true;
    await expect.poll(() => page.locator('video').evaluate((v) => v.currentTime), { timeout: 15_000 }).toBeGreaterThanOrEqual(8.9);
    await expect(page.getByText('正在准备续播位置，请稍候…')).toHaveCount(0);
});

test('releases a playback source that resolves after leaving the player', async ({ page }) => {
    await mockBackend(page, { delay: 1000 });
    await page.goto('/#/play/series/ep1');
    await page.waitForFunction(() => window.playStarted);
    await page.getByRole('link', { name: '⚙️ 设置' }).click();
    await expect(page).toHaveURL(/#\/settings$/);
    await expect.poll(() => page.evaluate(() => window.releasedURLs.length)).toBe(1);
    expect(await page.evaluate(() => window.canceledRequests.length)).toBe(1);
    expect(await page.evaluate(() => window.releasedURLs[0])).toContain('video.mp4');
});

test('keeps the full duration in history while a mux playlist is growing', async ({ page }) => {
    publishAll = false;
    await mockBackend(page, { mux: true, savedPosition: 0 });
    await page.goto('/#/play/series/ep1');
    await page.waitForFunction(() => document.querySelector('video')?.readyState >= 1);
    await expect.poll(() => page.evaluate(() => window.savedItems.length)).toBeGreaterThan(0);
    const state = await page.evaluate(() => ({
        visibleDuration: document.querySelector('video').duration,
        savedDuration: JSON.parse(localStorage.getItem('progress')).duration,
    }));
    expect(state.visibleDuration).toBeLessThan(9);
    expect(state.savedDuration).toBe(12);
});

async function attachHarnessBackend(page) {
    await page.addInitScript(({ origin }) => {
        const rpc = async (method, fields = {}) => {
            const response = await fetch(`${origin}/rpc`, {
                method: 'POST', headers: { 'Content-Type': 'text/plain' }, body: JSON.stringify({ method, ...fields }),
            });
            if (!response.ok) throw new Error(await response.text());
            return response.json();
        };
        window.playbackSources = [];
        window.go = { main: { App: {
            GetSeries: async () => ({ id: 'series', title: '测试剧集', coverUrl: '' }),
            ListEpisodes: async () => ({ items: [{ id: 'slow', number: 0, playable: true }, { id: 'ep1', number: 1, playable: true }, { id: 'ep2', number: 2, playable: true }], hasMore: false, cursor: '' }),
            GetProgress: (id) => rpc('GetProgress', { id }),
            BeginPlayback: () => rpc('BeginPlayback'),
            CancelPlayback: (requestId) => rpc('CancelPlayback', { requestId }),
            PlayURLForRequest: async (id, requestId) => { const source = await rpc('PlayURLForRequest', { id, requestId }); window.playbackSources.push(source); return source; },
            SaveProgress: (item) => rpc('SaveProgress', { item }),
            ReleasePlayback: (url) => rpc('ReleasePlayback', { url }),
        } } };
    }, { origin: harnessOrigin });
}

test('uses the Go playback gateway and SQLite to resume and change episodes', async ({ page }) => {
    await harnessRPC('SaveProgress', { item: { seriesId: 'series', episodeId: 'ep1', episodeNumber: 1, title: '测试剧集', coverUrl: '', position: 4, duration: 12, finished: false } });
    await attachHarnessBackend(page);
    await page.goto('/#/play/series/ep1');
    await page.waitForFunction(() => document.querySelector('video')?.readyState >= 1);
    await expect.poll(() => page.locator('video').evaluate((v) => v.currentTime)).toBeGreaterThanOrEqual(3.9);
    await page.locator('video').evaluate((v) => { v.currentTime = 8; });
    await expect.poll(async () => (await harnessRPC('GetProgress', { id: 'series' }))?.position).toBeGreaterThanOrEqual(7.9);
    await page.reload();
    await page.waitForFunction(() => document.querySelector('video')?.readyState >= 1);
    await expect.poll(() => page.locator('video').evaluate((v) => v.currentTime)).toBeGreaterThanOrEqual(7.9);
    const previousSource = await page.evaluate(() => window.playbackSources[0]);
    await page.locator('video').evaluate((v) => { v.currentTime = v.duration - 0.2; return v.play(); });
    await expect(page).toHaveURL(/#\/play\/series\/ep2$/);
    await page.waitForFunction(() => document.querySelector('video')?.readyState >= 1);
    await expect.poll(async () => (await fetch(previousSource)).status).toBe(404);
    await page.locator('video').evaluate((v) => { v.currentTime = 6; });
    await expect.poll(async () => {
        const progress = await harnessRPC('GetProgress', { id: 'series' });
        return progress?.episodeId === 'ep2' ? progress.position : 0;
    }).toBeGreaterThanOrEqual(5.9);
});

test('switching episodes cancels a blocked Go playback resolution', async ({ page }) => {
    await attachHarnessBackend(page);
    await page.goto('/#/play/series/slow');
    await expect.poll(() => harnessRPC('IsSlowStarted')).toBe(true);
    await page.evaluate(() => { window.location.hash = '#/play/series/ep2'; });
    await expect(page).toHaveURL(/#\/play\/series\/ep2$/);
    await page.waitForFunction(() => document.querySelector('video')?.readyState >= 1, null, { timeout: 10_000 });
    await expect(page.getByRole('alert')).toHaveCount(0);
});

test('native smoke script exercises playback and reports its result', async ({ page }) => {
    smokeReports.length = 0;
    await attachHarnessBackend(page);
    await page.goto('/#/settings');
    await page.evaluate((origin) => { window.__WRA_SMOKE_ORIGIN__ = origin; }, mediaOrigin);
    await page.addScriptTag({ path: join(process.cwd(), 'tests', 'webview2-smoke-client.js') });
    await expect.poll(() => smokeReports.length, { timeout: 30_000 }).toBeGreaterThan(0);
    expect(smokeReports[0]).toEqual({ ok: true, checks: ['mp4-resume', 'mp4-seek', 'sqlite-progress', 'automatic-next', 'hls-seek'] });
});
