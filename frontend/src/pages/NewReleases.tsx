import { useEffect, useMemo, useState } from 'react';
import { NewReleases, dto } from '../lib/wails';
import { useInfiniteList } from '../hooks/useInfiniteList';
import { SeriesGrid } from '../components/SeriesGrid';
import { LoadMore } from '../components/LoadMore';
import { Empty, ErrorBox, Loading } from '../components/StateBox';

function formatDate(d: string): string {
    if (d.length !== 8) return d;
    return `${d.slice(4, 6)}-${d.slice(6, 8)}`;
}

export default function NewReleasesPage() {
    const [date, setDate] = useState('');
    const [dates, setDates] = useState<string[]>([]);

    const load = useMemo(() => (cursor: string) => NewReleases(date, cursor), [date]);
    const list = useInfiniteList<dto.NewRelease>(load, [date]);

    useEffect(() => {
        if (dates.length > 0) return;
        NewReleases(date, '')
            .then((p) => {
                setDates(p.dates);
                if (!date && p.date) setDate(p.date);
            })
            .catch(() => {});
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    const items = list.items.map((r) => ({
        id: r.seriesId,
        title: r.title,
        coverUrl: r.coverUrl,
        episodeCount: r.episodeCount,
        rightText: r.online ? r.heat : r.note || '即将上线',
    }));

    return (
        <div>
            {dates.length > 0 && (
                <div className="mb-5 flex flex-wrap gap-2">
                    {dates.map((d) => (
                        <button
                            key={d}
                            onClick={() => setDate(d)}
                            className={`rounded-full px-3 py-1.5 text-sm transition ${
                                date === d
                                    ? 'bg-orange-500 font-medium text-white'
                                    : 'bg-zinc-800 text-zinc-300 hover:bg-zinc-700'
                            }`}
                        >
                            {formatDate(d)}
                        </button>
                    ))}
                </div>
            )}

            {items.length > 0 && <SeriesGrid items={items} />}
            {items.length === 0 && list.loading && <Loading />}
            {items.length === 0 && list.error && (
                <ErrorBox message={list.error} onRetry={list.reload} />
            )}
            {items.length === 0 && !list.loading && !list.error && <Empty text="这一天没有上新" />}
            {items.length > 0 && (
                <LoadMore onLoad={list.loadMore} hasMore={list.hasMore} loading={list.loading} />
            )}
        </div>
    );
}
