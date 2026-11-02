import { useCallback, useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ShortRank, dto } from '../lib/wails';
import { Cover } from '../components/Cover';
import { LoadMore } from '../components/LoadMore';
import { ErrorBox, Loading } from '../components/StateBox';
import { errorText } from '../lib/errors';

export default function RankPage() {
    const [content, setContent] = useState('');
    const [board, setBoard] = useState('');
    const [contents, setContents] = useState<dto.RankContent[]>([]);
    const [items, setItems] = useState<dto.RankItem[]>([]);
    const [cursor, setCursor] = useState('');
    const [hasMore, setHasMore] = useState(true);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const reqRef = useRef(0);
    const navigate = useNavigate();

    const fetchPage = useCallback(
        async (nextCursor: string, reset: boolean) => {
            const req = ++reqRef.current;
            setLoading(true);
            setError('');
            try {
                const page = await ShortRank(content, board, nextCursor);
                if (req !== reqRef.current) return;
                if (page.contents.length > 0) setContents(page.contents);
                setItems((prev) => (reset ? page.items : [...prev, ...page.items]));
                setCursor(page.cursor);
                setHasMore(page.hasMore);
            } catch (e) {
                if (req === reqRef.current) setError(errorText(e));
            } finally {
                if (req === reqRef.current) setLoading(false);
            }
        },
        [content, board],
    );

    useEffect(() => {
        setItems([]);
        setCursor('');
        setHasMore(true);
        void fetchPage('', true);
    }, [fetchPage]);

    const active = contents.find((c) => c.id === content) ?? contents[0];

    return (
        <div className="space-y-5">
            <div className="flex flex-wrap gap-2">
                {contents.map((c) => (
                    <button
                        key={c.id}
                        onClick={() => {
                            setContent(c.id);
                            setBoard('');
                        }}
                        className={`rounded-full px-4 py-1.5 text-sm transition ${
                            active?.id === c.id
                                ? 'bg-orange-500 font-medium text-white'
                                : 'bg-zinc-800 text-zinc-300 hover:bg-zinc-700'
                        }`}
                    >
                        {c.name}
                    </button>
                ))}
            </div>

            {active && active.boards.length > 0 && (
                <div className="flex flex-wrap gap-2">
                    {active.boards.map((b) => (
                        <button
                            key={b.id}
                            onClick={() => setBoard(b.id)}
                            className={`rounded-md px-3 py-1 text-xs transition ${
                                (board || active.boards.find((x) => x.selected)?.id) === b.id
                                    ? 'bg-zinc-700 text-white'
                                    : 'bg-zinc-900 text-zinc-400 hover:bg-zinc-800'
                            }`}
                        >
                            {b.name}
                        </button>
                    ))}
                </div>
            )}

            {items.length === 0 && loading && <Loading />}
            {items.length === 0 && error && (
                <ErrorBox message={error} onRetry={() => fetchPage('', true)} />
            )}

            <div className="space-y-1">
                {items.map((it, idx) => {
                    const isActor = !!it.userId;
                    const position = it.position || idx + 1;
                    return (
                        <button
                            key={`${it.seriesId || it.userId}-${idx}`}
                            onClick={() => {
                                if (isActor) navigate(`/person/${it.userId}`);
                                else if (it.seriesId) navigate(`/series/${it.seriesId}`);
                            }}
                            className="flex w-full items-center gap-4 rounded-lg px-3 py-2 text-left hover:bg-zinc-800"
                        >
                            <span
                                className={`w-6 text-center text-base ${
                                    position <= 3 ? 'font-bold text-orange-400' : 'text-zinc-500'
                                }`}
                            >
                                {position}
                            </span>
                            {isActor ? (
                                <Cover
                                    src={it.avatarUrl}
                                    alt={it.actor}
                                    className="h-14 w-14 shrink-0 rounded-full object-cover"
                                />
                            ) : (
                                <Cover
                                    src={it.coverUrl}
                                    alt={it.title}
                                    className="h-16 w-12 shrink-0 rounded object-cover"
                                />
                            )}
                            <div className="min-w-0 flex-1">
                                <div className="truncate text-sm font-medium">
                                    {isActor ? it.actor : it.title}
                                </div>
                                <div className="truncate text-xs text-zinc-500">
                                    {isActor ? it.details?.join(' · ') : it.score}
                                </div>
                            </div>
                            <ChangeTag change={it.change} />
                        </button>
                    );
                })}
            </div>

            {items.length > 0 && (
                <LoadMore
                    onLoad={() => fetchPage(cursor, false)}
                    hasMore={hasMore}
                    loading={loading}
                />
            )}
        </div>
    );
}

function ChangeTag({ change }: { change?: dto.RankChange | null }) {
    if (!change) return null;
    // 0 新上榜 1 上升 2 下降 3 不变
    switch (change.kind) {
        case 0:
            return <span className="text-xs text-orange-400">新</span>;
        case 1:
            return <span className="text-xs text-red-400">↑{change.places}</span>;
        case 2:
            return <span className="text-xs text-green-400">↓{change.places}</span>;
        default:
            return <span className="text-xs text-zinc-600">—</span>;
    }
}
