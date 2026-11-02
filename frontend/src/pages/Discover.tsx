import { useMemo, useState } from 'react';
import { Discover, dto } from '../lib/wails';
import { useInfiniteList } from '../hooks/useInfiniteList';
import { SeriesGrid } from '../components/SeriesGrid';
import { LoadMore } from '../components/LoadMore';
import { ErrorBox, Loading } from '../components/StateBox';

const CHANNELS = [
    { key: '', label: '推荐' },
    { key: '电影', label: '电影' },
    { key: '漫剧', label: '漫剧' },
    { key: '真人剧', label: '真人剧' },
    { key: '经典', label: '经典' },
    { key: '知识', label: '知识' },
];

export default function DiscoverPage() {
    const [channel, setChannel] = useState('');
    const load = useMemo(
        () => (cursor: string) => Discover(channel, cursor),
        [channel],
    );
    const list = useInfiniteList<dto.Series>(load, [channel]);

    return (
        <div>
            <div className="mb-5 flex flex-wrap gap-2">
                {CHANNELS.map((c) => (
                    <button
                        key={c.key}
                        onClick={() => setChannel(c.key)}
                        className={`rounded-full px-4 py-1.5 text-sm transition ${
                            channel === c.key
                                ? 'bg-orange-500 font-medium text-white'
                                : 'bg-zinc-800 text-zinc-300 hover:bg-zinc-700'
                        }`}
                    >
                        {c.label}
                    </button>
                ))}
            </div>

            {list.items.length > 0 && <SeriesGrid items={list.items} />}
            {list.items.length === 0 && list.loading && <Loading />}
            {list.items.length === 0 && list.error && (
                <ErrorBox message={list.error} onRetry={list.reload} />
            )}
            {list.items.length > 0 && (
                <LoadMore
                    onLoad={list.loadMore}
                    hasMore={list.hasMore}
                    loading={list.loading}
                />
            )}
            {list.items.length > 0 && list.error && (
                <ErrorBox message={list.error} onRetry={list.loadMore} />
            )}
        </div>
    );
}
