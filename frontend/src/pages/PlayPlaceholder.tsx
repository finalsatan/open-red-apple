import { useEffect, useRef, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import Hls from 'hls.js';
import { BeginPlayback, CancelPlayback, GetSeries, GetProgress, ListEpisodes, SaveProgress, PlayURLForRequest, ReleasePlayback, dto } from '../lib/wails';
import { errorText } from '../lib/errors';

// Serialise writes across episode changes, so an old episode cannot overwrite a newer one.
let progressWrites: Promise<unknown> = Promise.resolve();
function writeProgress(item: dto.HistoryItem): Promise<unknown> {
    progressWrites = progressWrites.catch(() => {}).then(() => SaveProgress(item));
    return progressWrites.catch(() => {});
}

export default function PlayPage() {
    const { seriesId = '', episodeId = '' } = useParams();
    const navigate = useNavigate();
    const video = useRef<HTMLVideoElement>(null);
    const attachedVideo = useRef<HTMLVideoElement | null>(null);
    const saveCurrent = useRef<(v: HTMLVideoElement | null) => Promise<unknown>>(() => Promise.resolve());
    const [url, setUrl] = useState('');
    const [error, setError] = useState('');
    const [attempt, setAttempt] = useState(0);
    const [series, setSeries] = useState<dto.Series | null>(null);
    const [episodes, setEpisodes] = useState<dto.Episode[]>([]);
    const [waitingResume, setWaitingResume] = useState(false);
    const [preparingMedia, setPreparingMedia] = useState(false);
    const resume = useRef(0);
    const lastSave = useRef(0);
    const finished = useRef(false);
    const mounted = useRef(false);
    const playbackGeneration = useRef(0);
    const attachedGeneration = useRef(0);

    useEffect(() => {
        mounted.current = true;
        return () => { mounted.current = false; };
    }, []);

    useEffect(() => {
        let alive = true;
        const generation = ++playbackGeneration.current;
        let sourceForRequest = '';
        let pendingRequest = '';
        setUrl(''); setError(''); setSeries(null); setEpisodes([]); resume.current = 0;
        setWaitingResume(false);
        setPreparingMedia(false);
        finished.current = false;
        (async () => {
            try {
                await progressWrites.catch(() => {});
                if (!alive) return;
                const [s, progress] = await Promise.all([GetSeries(seriesId), GetProgress(seriesId)]);
                if (!alive) return;
                const all: dto.Episode[] = [];
                const seen = new Set<string>();
                let cursor = '';
                do {
                    seen.add(cursor);
                    const page = await ListEpisodes(seriesId, cursor);
                    if (!alive) return;
                    all.push(...page.items);
                    cursor = page.hasMore && !seen.has(page.cursor) ? page.cursor : '';
                } while (cursor);
                const request = await BeginPlayback();
                if (!alive) { void CancelPlayback(request).catch(() => {}); return; }
                pendingRequest = request;
                const source = await PlayURLForRequest(episodeId, request);
                pendingRequest = '';
                if (!alive) { void ReleasePlayback(source).catch(() => {}); return; }
                sourceForRequest = source;
                setSeries(s); setEpisodes(all);
                if (progress && progress.episodeId === episodeId && !progress.finished) resume.current = progress.position;
                setPreparingMedia(source.includes('?mux=1'));
                setUrl(source);
            } catch (e) { if (alive) setError(errorText(e)); }
        })();
        return () => {
            alive = false;
            if (playbackGeneration.current === generation) playbackGeneration.current++;
            if (pendingRequest) void CancelPlayback(pendingRequest).catch(() => {});
            if (sourceForRequest) {
                // Queue this snapshot before the next attempt reads history.
                if (!finished.current) void saveCurrent.current(attachedVideo.current);
                void ReleasePlayback(sourceForRequest).catch(() => {});
            }
        };
    }, [seriesId, episodeId, attempt]);

    const current = episodes.findIndex((e) => e.id === episodeId);
    const next = current >= 0 ? episodes.slice(current + 1).find((e) => e.playable) : undefined;
    function save(v: HTMLVideoElement | null, isFinished = false): Promise<unknown> {
        if (!v || !series || resume.current > 0 || !Number.isFinite(v.duration) || v.duration <= 0 || !Number.isFinite(v.currentTime)) return Promise.resolve();
        const declaredDuration = Number(new URL(url).searchParams.get('duration'));
        const duration = Number.isFinite(declaredDuration) ? Math.max(v.duration, declaredDuration) : v.duration;
        return writeProgress(new dto.HistoryItem({ seriesId, episodeId, episodeNumber: episodes[current]?.number || 0, title: series.title, coverUrl: series.coverUrl, position: v.currentTime, duration, finished: isFinished, updatedAt: 0 }));
    }
    saveCurrent.current = save;

    function restorePosition(v: HTMLVideoElement) {
        if (resume.current <= 0 || !Number.isFinite(v.duration)) return;
        if (resume.current >= v.duration) {
            if (!url.includes('?mux=1')) { resume.current = 0; return; }
            v.pause();
            setWaitingResume(true);
            return;
        }
        v.currentTime = resume.current;
        resume.current = 0;
        setWaitingResume(false);
        void v.play().catch(() => {});
    }

    useEffect(() => {
        const v = video.current;
        if (!url || !v) return;
        attachedVideo.current = v;
        attachedGeneration.current = playbackGeneration.current;
        let hls: Hls | undefined;
        if (!new URL(url).pathname.endsWith('/index.m3u8')) v.src = url;
        else if (Hls.isSupported()) {
            hls = new Hls({ startPosition: 0 });
            hls.on(Hls.Events.ERROR, (_event, data) => {
                if (data.fatal) setError('视频加载失败，请重试。');
            });
            hls.on(Hls.Events.LEVEL_LOADED, (_event, data) => {
                if (!data.details.live) setPreparingMedia(false);
                if (!data.details.live && resume.current >= data.details.totalduration && data.details.totalduration > 0) {
                    resume.current = 0;
                    setWaitingResume(false);
                    void v.play().catch(() => {});
                }
            });
            hls.loadSource(url);
            hls.attachMedia(v);
        } else if (v.canPlayType('application/vnd.apple.mpegurl')) v.src = url;
        else setError('当前环境不支持 HLS 播放。');
        return () => {
            hls?.destroy();
            v.removeAttribute('src');
            v.load();
            if (attachedVideo.current === v) attachedVideo.current = null;
        };
        // The metadata is set together with url; this effect owns the corresponding video element.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [url]);

    return <div className="space-y-4">
        <Link className="text-sm text-orange-400" to={`/series/${seriesId}`}>返回详情</Link>
        <h1 className="text-xl font-bold">{series?.title || '播放'} {current >= 0 && `· 第 ${episodes[current].number} 集`}</h1>
        {error && <div role="alert" className="flex items-center gap-3 text-red-400"><span>{error}</span><button className="rounded bg-zinc-800 px-3 py-1 text-zinc-100" onClick={() => { playbackGeneration.current++; setAttempt((n) => n + 1); }}>重试</button></div>}
        {waitingResume && <p className="text-sm text-zinc-400">正在准备续播位置，请稍候…</p>}
        {preparingMedia && !waitingResume && <p className="text-sm text-zinc-400">视频正在准备，可拖动已生成的部分。</p>}
        {!error && !url && <p>正在加载播放资源…</p>}
        {url && <video ref={video} className="max-h-[70vh] w-full bg-black" controls autoPlay
            onPlay={() => { if (finished.current && !next) finished.current = false; }}
            onLoadedMetadata={() => { if (video.current) restorePosition(video.current); }}
            onDurationChange={() => { if (video.current) restorePosition(video.current); }}
            onTimeUpdate={() => { if (!finished.current && Date.now() - lastSave.current > 5000) { lastSave.current = Date.now(); void save(video.current); } }}
            onSeeked={() => { if (!finished.current) void save(video.current); }}
            onPause={() => { if (!finished.current && !video.current?.ended) void save(video.current); }}
            onEnded={() => { const generation = attachedGeneration.current; if (generation !== playbackGeneration.current) return; finished.current = true; void save(video.current, true).then(() => { if (next && mounted.current && playbackGeneration.current === generation) navigate(`/play/${seriesId}/${next.id}`); }); }}
            onError={() => setError('视频加载失败，请重试。')} />}
        <div className="flex flex-wrap gap-2">{episodes.map((ep) => ep.playable ? <Link key={ep.id} to={`/play/${seriesId}/${ep.id}`} className={`rounded px-3 py-2 ${ep.id === episodeId ? 'bg-orange-600' : 'bg-zinc-800'}`}>{ep.number}</Link> : <span key={ep.id} className="rounded bg-zinc-900 px-3 py-2 text-zinc-600">{ep.number}</span>)}</div>
    </div>;
}
