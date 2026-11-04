import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import {
    GetSeries,
    IsFavorite,
    ListEpisodes,
    Seasons,
    SeriesRecommend,
    ToggleFavorite,
    dto,
} from '../lib/wails';
import { Cover } from '../components/Cover';
import { SeriesGrid } from '../components/SeriesGrid';
import { ErrorBox, Loading } from '../components/StateBox';
import { errorText } from '../lib/errors';
import { formatHeat } from '../lib/format';

export default function SeriesDetailPage() {
    const { id = '' } = useParams();
    const [series, setSeries] = useState<dto.Series | null>(null);
    const [episodes, setEpisodes] = useState<dto.Episode[]>([]);
    const [album, setAlbum] = useState<dto.Album | null>(null);
    const [related, setRelated] = useState<dto.Recommend | null>(null);
    const [favorite, setFavorite] = useState(false);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');

    useEffect(() => {
        let alive = true;
        setLoading(true);
        setError('');
        setSeries(null);
        setEpisodes([]);
        setAlbum(null);
        setRelated(null);
        (async () => {
            try {
                const s = await GetSeries(id);
                if (!alive) return;
                setSeries(s);
                const [eps, rec, fav] = await Promise.all([
                    ListEpisodes(id, ''),
                    SeriesRecommend(id).catch(() => null),
                    IsFavorite(id).catch(() => false),
                ]);
                if (!alive) return;
                setEpisodes(eps.items);
                setRelated(rec);
                setFavorite(fav);
                if (s.isMultiSeason) {
                    Seasons(id)
                        .then((a) => {
                            if (alive) setAlbum(a);
                        })
                        .catch(() => {});
                }
            } catch (e) {
                if (alive) setError(errorText(e));
            } finally {
                if (alive) setLoading(false);
            }
        })();
        return () => {
            alive = false;
        };
    }, [id]);

    async function toggleFavorite() {
        if (!series) return;
        try {
            setFavorite(await ToggleFavorite(series.id, series.title, series.coverUrl));
        } catch {
            /* 忽略收藏失败 */
        }
    }

    if (loading) return <Loading />;
    if (error) return <ErrorBox message={error} />;
    if (!series) return null;

    return (
        <div className="space-y-8">
            <div className="flex flex-col gap-6 sm:flex-row">
                <Cover
                    src={series.coverUrl}
                    alt={series.title}
                    className="aspect-[3/4] w-40 shrink-0 rounded-lg object-cover"
                />
                <div className="min-w-0 flex-1">
                    <h1 className="text-2xl font-bold">{series.title}</h1>
                    <div className="mt-2 flex flex-wrap items-center gap-3 text-sm text-zinc-400">
                        {series.rightText && (
                            <span className="rounded bg-orange-500/15 px-2 py-0.5 text-orange-400">
                                {series.rightText}
                            </span>
                        )}
                        {series.heat > 0 && <span>热度 {formatHeat(series.heat)}</span>}
                        {series.playCount > 0 && <span>播放 {formatHeat(series.playCount)}</span>}
                    </div>
                    {series.rows.length > 0 && (
                        <div className="mt-3 space-y-1 text-sm text-zinc-400">
                            {series.rows.map((r, i) => (
                                <div key={i}>{r.text}</div>
                            ))}
                        </div>
                    )}
                    {series.description && (
                        <p className="mt-3 line-clamp-3 text-sm text-zinc-300">
                            {series.description}
                        </p>
                    )}
                    <button
                        onClick={toggleFavorite}
                        className={`mt-4 rounded-lg px-4 py-2 text-sm font-medium transition ${
                            favorite
                                ? 'bg-orange-500 text-white'
                                : 'bg-zinc-800 text-zinc-200 hover:bg-zinc-700'
                        }`}
                    >
                        {favorite ? '已收藏' : '收藏'}
                    </button>
                </div>
            </div>

            {album && album.seasons.length > 1 && (
                <section>
                    <h2 className="mb-3 text-lg font-semibold">季</h2>
                    <div className="flex flex-wrap gap-3">
                        {album.seasons.map((s) => (
                            <Link
                                key={s.seriesId}
                                to={`/series/${s.seriesId}`}
                                className={`rounded-lg border px-3 py-2 text-sm transition ${
                                    s.seriesId === id
                                        ? 'border-orange-500 text-orange-400'
                                        : 'border-zinc-700 hover:border-zinc-500'
                                }`}
                            >
                                {s.title || `第${s.index}季`} · {s.episodeCount}集
                            </Link>
                        ))}
                    </div>
                </section>
            )}

            <section>
                <h2 className="mb-3 text-lg font-semibold">
                    选集 <span className="text-sm text-zinc-500">共 {episodes.length} 集</span>
                </h2>
                <div className="grid grid-cols-5 gap-2 sm:grid-cols-8 md:grid-cols-10 lg:grid-cols-12">
                    {episodes.map((ep) => (
                        ep.playable ? <Link
                            key={ep.id}
                            to={`/play/${series.id}/${ep.id}`}
                            className={`rounded-md py-2 text-center text-sm transition ${
                                ep.playable
                                    ? 'bg-zinc-800 hover:bg-zinc-700'
                                    : 'cursor-not-allowed bg-zinc-900 text-zinc-600'
                            }`}
                        >
                            {ep.number}
                        </Link> : <span key={ep.id} className="cursor-not-allowed rounded-md bg-zinc-900 py-2 text-center text-sm text-zinc-600">{ep.number}</span>
                    ))}
                </div>
            </section>

            {related && related.items.length > 0 && (
                <section>
                    <h2 className="mb-3 text-lg font-semibold">
                        {related.cellTitle || '相关推荐'}
                    </h2>
                    <SeriesGrid items={related.items} />
                </section>
            )}
        </div>
    );
}
