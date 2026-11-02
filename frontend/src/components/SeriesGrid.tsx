import { SeriesCard, type SeriesLike } from './SeriesCard';

// SeriesGrid 是自适应列数的短剧网格。
export function SeriesGrid({ items }: { items: SeriesLike[] }) {
    return (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
            {items.map((s) => (
                <SeriesCard key={s.id} series={s} />
            ))}
        </div>
    );
}
