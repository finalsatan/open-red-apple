# open-red-apple

基于 [openhongguo](https://github.com/beilunyang/openhongguo) SDK 的 Windows 11 桌面端短剧应用。

技术方案见 [`docs/技术方案.md`](docs/技术方案.md)。

## 技术栈

| 层 | 选型 |
|---|---|
| 外壳 | Wails v2（WebView2） |
| 后端 | Go + openhongguo SDK |
| 前端 | React + TypeScript + Vite |
| 播放 | hls.js + 本地 ffmpeg 解密网关 |
| 输入 | 键盘 / 鼠标 / 手柄（XInput） |

## 当前状态

M1 浏览：

- 发现：频道切换 + 无限滚动
- 搜索：联想词、热搜、历史记录
- 短剧详情：分集、多季、相关推荐
- 榜单、演员主页、上新日历
- 收藏与观看记录（本地 SQLite）

M2 清晰播放：支持清晰 MP4、HLS 和带独立音轨的清晰资源、拖动进度、观看进度记忆和自动下一集；观看记录可直接续播。独立音轨需要 FFmpeg，将 `ffmpeg.exe` 放在程序目录或加入系统 PATH；视频会逐步合流，已生成的部分可拖动。设备池、加密播放与手柄操作见后续里程碑。

## 环境要求

- Go 1.26+
- Node.js 20+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)：`go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- FFmpeg（播放独立音轨的清晰资源、运行浏览器测试时需要）

## 开发

```sh
# 热重载开发（需要本机 WebView2 / WebKitGTK）
wails dev

# 交叉编译 Windows amd64
wails build -platform windows/amd64
```

产物：`build/bin/win-red-apple.exe`。

播放器回归测试使用本地生成的视频，并覆盖 Go 播放接口、媒体网关和 SQLite；播放模型使用测试数据，不依赖线上资源：

```sh
cd frontend
npm ci
npx playwright install chromium
npm run test:e2e
```

在已登录的 Windows 桌面上，可将 `build/bin/wra-webview-smoke-windows.zip` 解压到单独文件夹，双击其中的 `wra-webview-check.exe`。自检使用本地样本检查 WebView2 播放、拖动、续播和自动切集；Windows 端只需要 WebView2 Runtime。开发者可在已安装 Go、Node.js、Wails CLI 和 FFmpeg 的 Linux 环境运行 `tools/build-windows-webview-smoke.sh` 生成该压缩包；为避免覆盖现有文件，生成前请先移走旧压缩包。

也可以在安装了 Go、Node.js、Wails CLI、FFmpeg 和 WebView2 Runtime 的 Windows 开发环境中构建并运行自检：

```powershell
cd frontend
npm ci
npm run build
node tests/prepare-webview2-smoke.mjs
cd ..
wails build -platform windows/amd64 -s -skipbindings -o win-red-apple-webview-test.exe
cd frontend
node tests/run-webview2-smoke.mjs
npm run build
```

自检构建会向临时前端产物插入测试脚本，输出到单独的 `win-red-apple-webview-test.exe`；最后的 `npm run build` 会恢复正常前端产物。自检通过不代表线上清晰分集已验证。

开发环境可运行 `node tests/run-webview2-smoke.mjs --simulate`，用 Chromium 验证自检驱动和本地测试服务；该模式不验证 WebView2。

## 目录结构

```
.
├── app.go                  # 绑定给前端的 App 方法
├── main.go                 # Wails 入口
├── internal/
│   ├── dto/                # 暴露给前端的传输结构
│   ├── ohg/                # openhongguo SDK 封装
│   ├── paths/              # 应用数据目录
│   └── store/              # SQLite：搜索记录 / 收藏 / 观看记录
├── frontend/               # React + TS + Tailwind 前端
│   └── src/
│       ├── components/     # 卡片、网格、状态
│       ├── hooks/          # 分页等
│       ├── lib/            # Wails 绑定出口、错误与格式化
│       └── pages/          # 发现 / 搜索 / 详情 / 榜单 / 演员 / 收藏 / 记录 / 设置
└── docs/技术方案.md         # 技术方案
```

## 免责声明

本项目仅供学习和研究使用，禁止用于商业用途。

本项目基于 [openhongguo](https://github.com/beilunyang/openhongguo) 提供的非官方 SDK，所访问内容的著作权归原权利人所有。使用者应遵守相关法律法规，自行承担使用风险。
