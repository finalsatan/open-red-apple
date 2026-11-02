import { useEffect, useMemo, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import {
    AddSearchKeyword,
    HotSearch,
    Search,
    SearchCues,
    SearchHistory,
    Suggest,
    dto,
} from '../lib/wails';
import { useInfiniteList } from '../hooks/useInfiniteList';
import { SeriesGrid } from '../components/SeriesGrid';
import { LoadMore } from '../components/LoadMore';
import { Cover } from '../components/Cover';
import { Empty, ErrorBox, Loading } from '../components/StateBox';

export default function SearchPage() {
    const [params, setParams] = useSearchParams();
    const query = params.get('q') ?? '';
    const [input, setInput] = useState(query);
    const [cues, setCues] = useState<string[]>([]);
    const [history, setHistory] = useState<string[]>([]);
    const [hot, setHot] = useState<dto.HotSearch | null>(null);
    const [suggests, setSuggests] = useState<dto.Suggestion[]>([]);

    useEffect(() => setInput(query), [query]);

    useEffect(() => {
        SearchCues().then(setCues).catch(() => {});
        SearchHistory().then(setHistory).catch(() => {});
        HotSearch().then(setHot).catch(() => {});
    }, []);

    useEffect(() => {
        const text = input.trim();
        if (!text) {
            setSuggests([]);
            return;
        }
        const timer = setTimeout(() => {
            Suggest(text)
                .then(setSuggests)
                .catch(() => setSuggests([]));
        }, 250);
        return () => clearTimeout(timer);
    }, [input]);

    const load = useMemo(
        () => (cursor: string) =>
            query
                ? Search(query, cursor, '', [])
                : Promise.resolve({ items: [], cursor: '', hasMore: false }),
        [query],
    );
    const list = useInfiniteList<dto.Series>(load, [query]);

    function run(raw: string) {
        const text = raw.trim();
        if (!text) return;
        AddSearchKeyword(text).catch(() => {});
        setHistory((prev) => [text, ...prev.filter((x) => x !== text)].slice(0, 20));
        setParams({ q: text });
    }

    return (
        <div>
            <form
                onSubmit={(e) => {
                    e.preventDefault();
                    run(input);
                }}
                className="mb-6 flex gap-2"
            >
                <input
                    value={input}
                    onChange={(e) => setInput(e.target.value)}
                    autoFocus
                    placeholder="搜索短剧、演员…"
                    className="flex-1 rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm outline-none focus:border-orange-500"
                />
                <button className="rounded-lg bg-orange-500 px-4 py-2 text-sm font-medium text-white hover:bg-orange-400">
                    搜索
                </button>
            </form>

            {!query ? (
                <div className="space-y-7">
                    {suggests.length > 0 && (
                        <Chips title="联想" items={suggests.map((s) => s.text)} onPick={run} />
                    )}
                    {cues.length > 0 && <Chips title="提示" items={cues} onPick={run} />}
                    {history.length > 0 && <Chips title="历史" items={history} onPick={run} />}

                    {hot && hot.guesses.length > 0 && (
                        <section>
                            <h2 className="mb-3 text-sm font-semibold text-zinc-400">猜你想搜</h2>
                            <div className="grid grid-cols-2 gap-3 md:grid-cols-3">
                                {hot.guesses.map((g) => (
                                    <button
                                        key={g.seriesId}
                                        onClick={() => run(g.title)}
                                        className="flex items-center gap-3 rounded-lg bg-zinc-900 p-2 text-left hover:bg-zinc-800"
                                    >
                                        <Cover
                                            src={g.coverUrl}
                                            alt={g.title}
                                            className="h-16 w-12 shrink-0 rounded object-cover"
                                        />
                                        <div className="min-w-0">
                                            <div className="truncate text-sm">{g.title}</div>
                                            <div className="text-xs text-zinc-500">{g.heat}</div>
                                        </div>
                                    </button>
                                ))}
                            </div>
                        </section>
                    )}

                    {hot?.boards.map((b, i) => (
                        <section key={i}>
                            <h2 className="mb-2 text-sm font-semibold text-zinc-400">{b.name}</h2>
                            <div className="space-y-0.5">
                                {b.items.map((it) => (
                                    <button
                                        key={it.seriesId}
                                        onClick={() => run(it.title)}
                                        className="flex w-full items-center gap-3 rounded px-2 py-1.5 text-left hover:bg-zinc-800"
                                    >
                                        <span
                                            className={`w-5 text-center text-sm ${
                                                it.position <= 3
                                                    ? 'font-bold text-orange-400'
                                                    : 'text-zinc-500'
                                            }`}
                                        >
                                            {it.position}
                                        </span>
                                        <span className="flex-1 truncate text-sm">{it.title}</span>
                                        <span className="text-xs text-zinc-500">
                                            {it.searchScore || it.heat}
                                        </span>
                                    </button>
                                ))}
                                {b.topics.map((t) => (
                                    <button
                                        key={t.id}
                                        onClick={() => run(t.query || t.title)}
                                        className="flex w-full items-center gap-3 rounded px-2 py-1.5 text-left hover:bg-zinc-800"
                                    >
                                        <span className="w-5 text-center text-sm text-zinc-500">#</span>
                                        <span className="flex-1 truncate text-sm">{t.title}</span>
                                        <span className="text-xs text-zinc-500">{t.heat}</span>
                                    </button>
                                ))}
                            </div>
                        </section>
                    ))}
                </div>
            ) : (
                <>
                    {list.items.length > 0 && <SeriesGrid items={list.items} />}
                    {list.items.length === 0 && list.loading && <Loading />}
                    {list.items.length === 0 && list.error && (
                        <ErrorBox message={list.error} onRetry={list.reload} />
                    )}
                    {list.items.length === 0 && !list.loading && !list.error && (
                        <Empty text="没有搜索结果" />
                    )}
                    {list.items.length > 0 && (
                        <LoadMore
                            onLoad={list.loadMore}
                            hasMore={list.hasMore}
                            loading={list.loading}
                        />
                    )}
                </>
            )}
        </div>
    );
}

function Chips({
    title,
    items,
    onPick,
}: {
    title: string;
    items: string[];
    onPick: (v: string) => void;
}) {
    return (
        <section>
            <h2 className="mb-2 text-sm font-semibold text-zinc-400">{title}</h2>
            <div className="flex flex-wrap gap-2">
                {items.map((it, i) => (
                    <button
                        key={i}
                        onClick={() => onPick(it)}
                        className="rounded-full bg-zinc-800 px-3 py-1 text-sm hover:bg-zinc-700"
                    >
                        {it}
                    </button>
                ))}
            </div>
        </section>
    );
}
