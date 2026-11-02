import { useEffect, useMemo, useState } from 'react';
import { useParams } from 'react-router-dom';
import { Person, PersonWorks, dto } from '../lib/wails';
import { useInfiniteList } from '../hooks/useInfiniteList';
import { SeriesGrid } from '../components/SeriesGrid';
import { LoadMore } from '../components/LoadMore';
import { Cover } from '../components/Cover';
import { ErrorBox, Loading } from '../components/StateBox';
import { errorText } from '../lib/errors';
import { formatHeat } from '../lib/format';

export default function PersonPage() {
    const { id = '' } = useParams();
    const [person, setPerson] = useState<dto.Person | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');

    useEffect(() => {
        let alive = true;
        setLoading(true);
        setError('');
        Person(id)
            .then((p) => {
                if (alive) setPerson(p);
            })
            .catch((e) => {
                if (alive) setError(errorText(e));
            })
            .finally(() => {
                if (alive) setLoading(false);
            });
        return () => {
            alive = false;
        };
    }, [id]);

    const load = useMemo(() => (cursor: string) => PersonWorks(id, cursor), [id]);
    const works = useInfiniteList<dto.Work>(load, [id]);

    const items = works.items.map((w) => ({
        id: w.seriesId,
        title: w.title,
        coverUrl: w.coverUrl,
        episodeCount: w.episodeCount,
        rightText: w.tags?.slice(0, 2).join(' · '),
    }));

    if (loading) return <Loading />;
    if (error) return <ErrorBox message={error} />;
    if (!person) return null;

    return (
        <div className="space-y-8">
            <div className="flex items-center gap-5">
                <Cover
                    src={person.avatarUrl}
                    alt={person.name}
                    className="h-24 w-24 shrink-0 rounded-full object-cover"
                />
                <div className="min-w-0">
                    <h1 className="text-2xl font-bold">{person.name}</h1>
                    <div className="mt-1 flex flex-wrap items-center gap-2 text-xs text-zinc-400">
                        {person.tags?.map((t) => (
                            <span key={t} className="rounded bg-zinc-800 px-2 py-0.5">
                                {t}
                            </span>
                        ))}
                        {person.labels?.map((t) => (
                            <span key={t}>{t}</span>
                        ))}
                    </div>
                    <div className="mt-2 text-sm text-zinc-400">
                        粉丝 {formatHeat(person.followers)} · 获赞 {formatHeat(person.likes)}
                    </div>
                    {person.description && (
                        <p className="mt-2 max-w-2xl text-sm text-zinc-400">
                            {person.description}
                        </p>
                    )}
                </div>
            </div>

            <section>
                <h2 className="mb-3 text-lg font-semibold">作品</h2>
                {items.length > 0 && <SeriesGrid items={items} />}
                {items.length === 0 && works.loading && <Loading />}
                {items.length > 0 && (
                    <LoadMore
                        onLoad={works.loadMore}
                        hasMore={works.hasMore}
                        loading={works.loading}
                    />
                )}
            </section>
        </div>
    );
}
