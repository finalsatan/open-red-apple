import { useState, type FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';

export function TopBar() {
    const [q, setQ] = useState('');
    const navigate = useNavigate();

    function submit(e: FormEvent) {
        e.preventDefault();
        const query = q.trim();
        navigate(query ? `/search?q=${encodeURIComponent(query)}` : '/search');
    }

    return (
        <header className="flex items-center gap-4 border-b border-zinc-800 px-6 py-3">
            <form onSubmit={submit} className="flex w-full max-w-xl gap-2">
                <input
                    value={q}
                    onChange={(e) => setQ(e.target.value)}
                    placeholder="搜索短剧…"
                    className="flex-1 rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm outline-none focus:border-orange-500"
                />
                <button
                    type="submit"
                    className="rounded-lg bg-orange-500 px-4 py-2 text-sm font-medium text-white hover:bg-orange-400"
                >
                    搜索
                </button>
            </form>
        </header>
    );
}
