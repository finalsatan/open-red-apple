// Creates a temporary test-only frontend; run `npm run build` to restore it.
import { copyFileSync, readFileSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';

const dist = join(process.cwd(), 'dist');
const index = join(dist, 'index.html');
const html = readFileSync(index, 'utf8');
if (!html.includes('</head>') || html.includes('webview2-smoke-client.js')) {
    throw new Error('请先运行 npm run build，再准备 WebView2 自检');
}
const port = Number(process.env.WRA_WEBVIEW_SMOKE_PORT || 34827);
if (!Number.isInteger(port) || port < 1024 || port > 65535) throw new Error('无效的自检端口');
copyFileSync(join(process.cwd(), 'tests', 'webview2-smoke-client.js'), join(dist, 'webview2-smoke-client.js'));
const script = `<script>window.__WRA_SMOKE_ORIGIN__="http://127.0.0.1:${port}";</script><script defer src="/webview2-smoke-client.js"></script>`;
writeFileSync(index, html.replace('</head>', `${script}</head>`));
console.log('已准备临时自检前端；请用 wails build -s 构建测试程序。');
