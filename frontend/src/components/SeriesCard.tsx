import { Link } from 'react-router-dom';
import { Cover } from './Cover';

// SeriesLike 是卡片渲染所需的最小字段，dto.Series 天然满足。
export interface SeriesLike {
    id: string;
    title: string;
    coverUrl?: string;
    episodeCount?: number;
    rightText?: string;
}

export function SeriesCard({ series }: { series: SeriesLike }) {
    return (
        <Link to={`/series/${series.id}`} className="group block">
            <div className="relative aspect-[3/4] overflow-hidden rounded-lg bg-zinc-800">
                <Cover
                    src={series.coverUrl}
                    alt={series.title}
                    className="h-full w-full object-cover transition duration-200 group-hover:scale-105"
                />
                {!!series.episodeCount && series.episodeCount > 0 && (
                    <span className="absolute bottom-1 right-1 rounded bg-black/70 px-1.5 py-0.5 text-xs text-white">
                        {series.episodeCount}集
                    </span>
                )}
            </div>
            <div className="mt-2 truncate text-sm font-medium">{series.title}</div>
            {series.rightText && (
                <div className="truncate text-xs text-zinc-500">{series.rightText}</div>
            )}
        </Link>
    );
}
