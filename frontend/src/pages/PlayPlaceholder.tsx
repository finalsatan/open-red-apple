import { Link, useParams } from 'react-router-dom';

// PlayPlaceholder 是播放页的占位，M2 会替换成真正的播放器。
export default function PlayPlaceholder() {
    const { seriesId = '', episodeId = '' } = useParams();
    return (
        <div className="flex h-full flex-col items-center justify-center gap-4 text-center">
            <div className="text-5xl">🎬</div>
            <div className="text-lg font-medium">播放功能即将上线</div>
            <div className="text-sm text-zinc-500">
                剧集 {seriesId} · 分集 {episodeId}
            </div>
            <Link
                to={`/series/${seriesId}`}
                className="rounded-lg bg-zinc-800 px-4 py-2 text-sm hover:bg-zinc-700"
            >
                返回详情
            </Link>
        </div>
    );
}
