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

M0 脚手架：发现页与搜索已可用（`Discover` / `Search`），交叉编译产出 Windows exe。

后续里程碑见技术方案第 9 节。

## 环境要求

- Go 1.22+
- Node.js 20+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)：`go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- ffmpeg（播放加密分集时需要）

## 开发

```sh
# 热重载开发（需要本机 WebView2 / WebKitGTK）
wails dev

# 交叉编译 Windows amd64
wails build -platform windows/amd64
```

产物：`build/bin/win-red-apple.exe`。

## 目录结构

```
.
├── app.go                  # 绑定给前端的 App 方法
├── main.go                 # Wails 入口
├── internal/
│   ├── dto/                # 暴露给前端的传输结构
│   ├── ohg/                # openhongguo SDK 封装
│   └── paths/              # 应用数据目录
├── frontend/               # React + TS 前端
└── docs/技术方案.md         # 技术方案
```

## 免责声明

本项目仅供学习和研究使用，禁止用于商业用途。

本项目基于 [openhongguo](https://github.com/beilunyang/openhongguo) 提供的非官方 SDK，所访问内容的著作权归原权利人所有。使用者应遵守相关法律法规，自行承担使用风险。
