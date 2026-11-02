import { useCallback, useEffect, useRef, useState } from 'react';
import { errorText } from '../lib/errors';

export interface Page<T> {
    items: T[];
    cursor: string;
    hasMore: boolean;
}

interface State {
    cursor: string;
    hasMore: boolean;
    loading: boolean;
    gen: number;
}

// useInfiniteList 管理游标分页：deps 变化时重置并加载第一页，
// 之后通过 loadMore 追加下一页。
export function useInfiniteList<T>(
    load: (cursor: string) => Promise<Page<T>>,
    deps: unknown[],
) {
    const loadRef = useRef(load);
    loadRef.current = load;

    const [items, setItems] = useState<T[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [hasMore, setHasMore] = useState(true);
    const state = useRef<State>({ cursor: '', hasMore: true, loading: false, gen: 0 });

    const loadMore = useCallback(async () => {
        const s = state.current;
        if (s.loading || !s.hasMore) return;
        s.loading = true;
        setLoading(true);
        setError('');
        const gen = s.gen;
        try {
            const page = await loadRef.current(s.cursor);
            if (gen !== s.gen) return;
            s.cursor = page.cursor ?? '';
            s.hasMore = page.hasMore;
            setHasMore(page.hasMore);
            setItems((prev) => [...prev, ...page.items]);
        } catch (e) {
            if (gen === s.gen) setError(errorText(e));
        } finally {
            if (gen === s.gen) setLoading(false);
            s.loading = false;
        }
    }, []);

    const reload = useCallback(() => {
        const s = state.current;
        s.gen += 1;
        s.cursor = '';
        s.hasMore = true;
        s.loading = false;
        setItems([]);
        setHasMore(true);
        setError('');
        setLoading(false);
        void loadMore();
    }, [loadMore]);

    useEffect(() => {
        reload();
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, deps);

    return { items, loading, error, hasMore, loadMore, reload };
}
