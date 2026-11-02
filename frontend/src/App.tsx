import { useEffect, useState } from 'react';
import './App.css';
import { Discover, Search } from '../wailsjs/go/main/App';
import { dto } from '../wailsjs/go/models';

function App() {
    const [items, setItems] = useState<dto.Series[]>([]);
    const [query, setQuery] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');

    async function load(run: () => Promise<dto.SeriesPage>) {
        setLoading(true);
        setError('');
        try {
            const page = await run();
            setItems(page.items ?? []);
        } catch (e) {
            setError(String(e));
            setItems([]);
        } finally {
            setLoading(false);
        }
    }

    useEffect(() => {
        load(() => Discover(''));
    }, []);

    function onSearch(e: React.FormEvent) {
        e.preventDefault();
        const q = query.trim();
        if (!q) {
            load(() => Discover(''));
            return;
        }
        load(() => Search(q));
    }

    return (
        <div id="App">
            <header className="topbar">
                <h1>红果短剧</h1>
                <form onSubmit={onSearch}>
                    <input
                        value={query}
                        onChange={(e) => setQuery(e.target.value)}
                        placeholder="搜索短剧…"
                        autoComplete="off"
                    />
                    <button type="submit">搜索</button>
                </form>
            </header>

            {error && <div className="error">{error}</div>}
            {loading && <div className="hint">加载中…</div>}
            {!loading && !error && items.length === 0 && (
                <div className="hint">没有结果</div>
            )}

            <main className="grid">
                {items.map((s) => (
                    <article className="card" key={s.id}>
                        {s.coverUrl ? (
                            <img src={s.coverUrl} alt={s.title} loading="lazy" />
                        ) : (
                            <div className="cover-placeholder" />
                        )}
                        <div className="meta">
                            <div className="title" title={s.title}>{s.title}</div>
                            <div className="sub">{s.episodeCount} 集</div>
                        </div>
                    </article>
                ))}
            </main>
        </div>
    );
}

export default App;
