import { useEffect, useRef } from 'react';

interface Props {
    onLoad: () => void;
    hasMore: boolean;
    loading: boolean;
}

// LoadMore 用 IntersectionObserver 在接近底部时加载下一页。
export function LoadMore({ onLoad, hasMore, loading }: Props) {
    const ref = useRef<HTMLDivElement>(null);

    useEffect(() => {
        const el = ref.current;
        if (!el || !hasMore || loading) return;
        const observer = new IntersectionObserver(
            (entries) => {
                if (entries[0]?.isIntersecting) onLoad();
            },
            { rootMargin: '400px' },
        );
        observer.observe(el);
        return () => observer.disconnect();
    }, [onLoad, hasMore, loading]);

    return (
        <div ref={ref} className="flex h-14 items-center justify-center text-sm text-zinc-500">
            {loading ? '加载中…' : hasMore ? '' : '没有更多了'}
        </div>
    );
}
