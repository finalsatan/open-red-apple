import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { ListHistory, dto } from '../lib/wails';
import { Cover } from '../components/Cover';
import { Empty, ErrorBox, Loading } from '../components/StateBox';
import { errorText } from '../lib/errors';
import { formatTime } from '../lib/format';

export default function HistoryPage() {
    const [items, setItems] = useState<dto.HistoryItem[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');

    useEffect(() => {
        ListHistory()
            .then(setItems)
            .catch((e) => setError(errorText(e)))
            .finally(() => setLoading(false));
    }, []);

    if (loading) return <Loading />;
    if (error) return <ErrorBox message={error} />;
    if (items.length === 0) return <Empty text="还没有观看记录" />;

    return (
        <div className="space-y-2">
            {items.map((h) => {
                const percent =
                    h.duration > 0 ? Math.min(100, (h.position / h.duration) * 100) : 0;
                return (
                    <Link
                        key={h.seriesId}
                        to={`/series/${h.seriesId}`}
                        className="flex gap-4 rounded-lg p-2 hover:bg-zinc-800"
                    >
                        <Cover
                            src={h.coverUrl}
                            alt={h.title}
                            className="h-20 w-14 shrink-0 rounded object-cover"
                        />
                        <div className="min-w-0 flex-1 self-center">
                            <div className="truncate text-sm font-medium">{h.title}</div>
                            <div className="mt-1 text-xs text-zinc-500">
                                第 {h.episodeNumber} 集 ·{' '}
                                {h.finished
                                    ? '已看完'
                                    : `${formatTime(h.position)} / ${formatTime(h.duration)}`}
                            </div>
                            {h.duration > 0 && (
                                <div className="mt-2 h-1 overflow-hidden rounded bg-zinc-800">
                                    <div
                                        className="h-full bg-orange-500"
                                        style={{ width: `${percent}%` }}
                                    />
                                </div>
                            )}
                        </div>
                    </Link>
                );
            })}
        </div>
    );
}
