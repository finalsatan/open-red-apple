// Runs a temporary, test-instrumented Wails executable in an interactive Windows session.
import { execFileSync, spawn, spawnSync } from 'node:child_process';
import { createReadStream, existsSync, mkdirSync, mkdtempSync, rmSync, statSync } from 'node:fs';
import { createServer } from 'node:http';
import { tmpdir } from 'node:os';
import { join, resolve } from 'node:path';

const simulate = process.argv.includes('--simulate');
if (process.platform !== 'win32' && !simulate) throw new Error('此自检需要交互式 Windows 环境');
const port = Number(process.env.WRA_WEBVIEW_SMOKE_PORT || 34827);
if (!Number.isInteger(port) || port < 1024 || port > 65535) throw new Error('无效的自检端口');
const executable = simulate ? '' : resolve(process.argv[2] || '../build/bin/win-red-apple-webview-test.exe');
if (!simulate && !existsSync(executable)) throw new Error(`未找到测试程序：${executable}`);
const directory = mkdtempSync(join(tmpdir(), 'win-red-apple-webview-'));
let server, harness, app, browser, vite, harnessOrigin = '';
let resolveReport;
const reported = new Promise((resolve) => { resolveReport = resolve; });

async function waitForHarness(child) {
    return new Promise((resolve, reject) => {
        let output = '';
        const timer = setTimeout(() => reject(new Error(`测试后端启动超时：${output}`)), 20000);
        const collect = (data) => {
            output += data.toString();
            const match = /WRA_HARNESS_URL=(http:\/\/[^\s]+)/.exec(output);
            if (match) { clearTimeout(timer); resolve(match[1]); }
        };
        child.stdout.on('data', collect);
        child.stderr.on('data', collect);
        child.once('error', (err) => { clearTimeout(timer); reject(err); });
        child.once('exit', (code) => { clearTimeout(timer); reject(new Error(`测试后端退出 ${code}：${output}`)); });
    });
}

try {
    execFileSync('ffmpeg', ['-hide_banner', '-loglevel', 'error', '-y', '-f', 'lavfi', '-i', 'testsrc2=size=320x180:rate=24', '-t', '12', '-c:v', 'libx264', '-threads', '1', '-pix_fmt', 'yuv420p', '-g', '48', '-movflags', '+faststart', join(directory, 'video.mp4')]);
    execFileSync('ffmpeg', ['-hide_banner', '-loglevel', 'error', '-y', '-i', join(directory, 'video.mp4'), '-c', 'copy', '-hls_time', '2', '-hls_playlist_type', 'vod', '-hls_segment_filename', join(directory, 'seg%02d.ts'), join(directory, 'index.m3u8')]);
    server = createServer((req, res) => {
        const name = new URL(req.url, 'http://localhost').pathname.slice(1);
        res.setHeader('Access-Control-Allow-Origin', '*');
        res.setHeader('Access-Control-Allow-Private-Network', 'true');
        if (req.method === 'OPTIONS') { res.setHeader('Access-Control-Allow-Methods', 'GET, POST'); res.setHeader('Access-Control-Allow-Headers', 'Content-Type'); res.writeHead(204).end(); return; }
        if (name === 'config') { res.setHeader('Content-Type', 'application/json'); res.end(JSON.stringify({ rpc: harnessOrigin })); return; }
        if (name === 'report' && req.method === 'POST') {
            let body = '';
            req.setEncoding('utf8');
            req.on('data', (chunk) => { body += chunk; });
            req.on('end', () => { try { resolveReport(JSON.parse(body)); res.end('ok'); } catch { res.writeHead(400).end(); } });
            return;
        }
        if (!/^(video\.mp4|index\.m3u8|seg\d+\.ts)$/.test(name)) { res.writeHead(404).end(); return; }
        const file = join(directory, name);
        const size = statSync(file).size;
        res.setHeader('Accept-Ranges', 'bytes');
        res.setHeader('Content-Type', name.endsWith('.m3u8') ? 'application/vnd.apple.mpegurl' : name.endsWith('.ts') ? 'video/mp2t' : 'video/mp4');
        let start = 0, end = size - 1;
        const range = /^bytes=(\d+)-(\d*)$/.exec(req.headers.range || '');
        if (range) { start = Number(range[1]); if (range[2]) end = Math.min(Number(range[2]), end); res.statusCode = 206; res.setHeader('Content-Range', `bytes ${start}-${end}/${size}`); }
        res.setHeader('Content-Length', end - start + 1);
        if (req.method === 'HEAD') { res.end(); return; }
        createReadStream(file, { start, end }).pipe(res);
    });
    await new Promise((resolve, reject) => { server.once('error', reject); server.listen(port, '127.0.0.1', resolve); });
    const harnessBinary = join(directory, process.platform === 'win32' ? 'player-harness.exe' : 'player-harness');
    execFileSync('go', ['test', '-c', '-o', harnessBinary, '.'], { cwd: join(process.cwd(), '..'), timeout: 120000, stdio: 'inherit' });
    harness = spawn(harnessBinary, ['-test.run=^TestPlayerHarness$', '-test.v'], { env: { ...process.env, WRA_E2E_MEDIA_ORIGIN: `http://127.0.0.1:${port}` }, stdio: ['ignore', 'pipe', 'pipe'] });
    harnessOrigin = await waitForHarness(harness);
    let exited = new Promise(() => {});
    if (simulate) {
        // Exercises the driver and fixtures without claiming WebView2 coverage.
        vite = spawn(process.execPath, [join(process.cwd(), 'node_modules', 'vite', 'bin', 'vite.js'), '--host', 'localhost', '--port', '34115', '--strictPort'], { stdio: ['ignore', 'pipe', 'pipe'] });
        await new Promise((resolve, reject) => {
            let output = '';
            const timeout = setTimeout(() => reject(new Error(`Vite 启动超时：${output}`)), 15000);
            const collect = (data) => { output += data.toString(); if (output.replace(/\x1b\[[0-9;]*m/g, '').includes('http://localhost:34115/')) { clearTimeout(timeout); resolve(); } };
            vite.stdout.on('data', collect);
            vite.stderr.on('data', collect);
            vite.once('error', (err) => { clearTimeout(timeout); reject(err); });
            vite.once('exit', (code) => { clearTimeout(timeout); reject(new Error(`Vite 退出 ${code}：${output}`)); });
        });
        const { chromium } = await import('@playwright/test');
        browser = await chromium.launch({ channel: 'chromium', headless: true, args: ['--autoplay-policy=no-user-gesture-required'] });
        const page = await browser.newPage();
        await page.addInitScript(({ origin, control }) => {
            const rpc = async (method, fields = {}) => {
                const response = await fetch(`${origin}/rpc`, { method: 'POST', headers: { 'Content-Type': 'text/plain' }, body: JSON.stringify({ method, ...fields }) });
                if (!response.ok) throw new Error(await response.text());
                return response.json();
            };
            window.__WRA_SMOKE_ORIGIN__ = control;
            window.go = { main: { App: { GetProgress: (id) => rpc('GetProgress', { id }), SaveProgress: (item) => rpc('SaveProgress', { item }) } } };
        }, { origin: harnessOrigin, control: `http://127.0.0.1:${port}` });
        await page.goto('http://localhost:34115/#/settings');
        await page.addScriptTag({ path: join(process.cwd(), 'tests', 'webview2-smoke-client.js') });
    } else {
        const appData = join(directory, 'appdata');
        const localAppData = join(directory, 'localappdata');
        mkdirSync(appData); mkdirSync(localAppData);
        // Windows environment names are case-insensitive; avoid duplicate variants.
        const appEnv = Object.fromEntries(Object.entries(process.env).filter(([key]) => !['appdata', 'localappdata'].includes(key.toLowerCase())));
        app = spawn(executable, [], { env: { ...appEnv, AppData: appData, LOCALAPPDATA: localAppData }, stdio: 'ignore' });
        exited = new Promise((_, reject) => {
            app.once('error', reject);
            app.once('exit', (code) => reject(new Error(`测试程序在报告前退出：${code}`)));
        });
    }
    let timer;
    const timeout = new Promise((_, reject) => { timer = setTimeout(() => reject(new Error('WebView2 自检超时；请确认在已登录的 Windows 桌面运行，并检查 WebView2 Runtime。')), 90000); });
    const result = await Promise.race([reported, exited, timeout]).finally(() => clearTimeout(timer));
    if (!result.ok) throw new Error(result.error || 'WebView2 自检失败');
    console.log(`${simulate ? 'Chromium 测试模拟' : 'WebView2 自检'}通过：${result.checks.join(', ')}`);
} finally {
    if (browser) await browser.close();
    if (vite && vite.exitCode === null) vite.kill();
    if (app?.pid && app.exitCode === null) spawnSync('taskkill', ['/PID', String(app.pid), '/T', '/F'], { stdio: 'ignore' });
    if (harness && harness.exitCode === null) {
        const exited = new Promise((resolve) => harness.once('exit', resolve));
        if (harnessOrigin) await fetch(`${harnessOrigin}/shutdown`, { method: 'POST' }).catch(() => {});
        await Promise.race([exited, new Promise((resolve) => setTimeout(resolve, 3000))]);
        if (harness.exitCode === null) { harness.kill(); await Promise.race([exited, new Promise((resolve) => setTimeout(resolve, 3000))]); }
    }
    if (server) { server.closeAllConnections(); await new Promise((resolve) => server.close(resolve)); }
    rmSync(directory, { recursive: true, force: true, maxRetries: 8, retryDelay: 250 });
}
