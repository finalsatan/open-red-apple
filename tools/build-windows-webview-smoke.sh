#!/usr/bin/env bash
set -euo pipefail

project_dir="$(cd "$(dirname "$0")/.." && pwd)"
fixture_dir="$project_dir/cmd/webview-smoke/fixtures/generated"
output="$project_dir/build/bin/wra-webview-smoke-windows.zip"
test_name="wra-webview-test-$$.exe"
test_binary="$project_dir/build/bin/$test_name"

for tool in ffmpeg go node npm python3 wails; do
    command -v "$tool" >/dev/null || { printf '缺少构建工具：%s\n' "$tool" >&2; exit 1; }
done
if [ -e "$fixture_dir" ] || [ -e "$output" ] || [ -e "$test_binary" ]; then
    printf '测试素材目录或输出文件已存在，请先检查：%s、%s、%s\n' "$fixture_dir" "$output" "$test_binary" >&2
    exit 1
fi
mkdir -p "$project_dir/build/bin"
staging_dir="$(mktemp -d "${TMPDIR:-/tmp}/wra-webview-package-XXXXXX")"
fixture_created=0
restore_frontend=0
package_complete=0
cleanup() {
    status=$?
    trap - EXIT
    if [ "$restore_frontend" -eq 1 ]; then
        npm run build --prefix "$project_dir/frontend" || { status=1; printf '前端产物恢复失败，请手动运行 npm run build --prefix frontend\n' >&2; }
    fi
    if [ "$fixture_created" -eq 1 ]; then rm -rf -- "$fixture_dir"; fi
    rm -rf -- "$staging_dir"
    rm -f -- "$test_binary"
    if [ "$package_complete" -eq 0 ]; then rm -f -- "$output"; fi
    exit "$status"
}
trap cleanup EXIT
mkdir "$fixture_dir"
fixture_created=1

ffmpeg -hide_banner -loglevel error -y -f lavfi -i 'testsrc2=size=320x180:rate=24' -t 12 -c:v libx264 -threads 1 -pix_fmt yuv420p -g 48 -movflags +faststart "$fixture_dir/video.mp4"
ffmpeg -hide_banner -loglevel error -y -i "$fixture_dir/video.mp4" -c copy -hls_time 2 -hls_playlist_type vod -hls_segment_filename "$fixture_dir/seg%02d.ts" "$fixture_dir/index.m3u8"

npm run build --prefix "$project_dir/frontend"
restore_frontend=1
(cd "$project_dir/frontend" && WRA_WEBVIEW_SMOKE_PORT=34827 node tests/prepare-webview2-smoke.mjs)
(cd "$project_dir" && wails build -platform windows/amd64 -s -skipbindings -o "$test_name")
cp "$test_binary" "$staging_dir/win-red-apple-webview-test.exe"
(cd "$project_dir" && GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test -c -o "$staging_dir/wra-player-harness.exe" .)
(cd "$project_dir" && GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o "$staging_dir/wra-webview-check.exe" ./cmd/webview-smoke)
cat > "$staging_dir/使用说明.txt" <<'README'
Windows WebView2 本地自检

1. 将压缩包内的所有文件解压到同一个文件夹，不要直接在压缩包内运行。
2. 在已登录的 Windows 11 桌面双击 wra-webview-check.exe。
3. 程序会打开临时测试窗口，使用内置测试视频检查播放、拖动、续播、观看记录和自动下一集。
4. 等待控制台显示“WebView2 自检通过”或具体错误，按 Enter 退出。

需要系统已安装 WebView2 Runtime。本测试不需要安装 Go、Node.js 或 FFmpeg；不使用真实线上分集。测试使用临时用户数据，结束后清理。请保留同文件夹里的三个 exe。
README
(cd "$staging_dir" && python3 -m zipfile -c "$output" wra-webview-check.exe wra-player-harness.exe win-red-apple-webview-test.exe 使用说明.txt)
python3 -m zipfile -t "$output"
package_complete=1
printf '已生成自检包：%s\n' "$output"
