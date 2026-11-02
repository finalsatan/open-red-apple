import { useEffect, useState } from 'react';
import { ListFavorites, dto } from '../lib/wails';
import { SeriesGrid } from '../components/SeriesGrid';
import { Empty, ErrorBox, Loading } from '../components/StateBox';
import { errorText } from '../lib/errors';

export default function FavoritesPage() {
    const [items, setItems] = useState<dto.Favorite[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');

    useEffect(() => {
        ListFavorites()
            .then(setItems)
            .catch((e) => setError(errorText(e)))
            .finally(() => setLoading(false));
    }, []);

    if (loading) return <Loading />;
    if (error) return <ErrorBox message={error} />;
    if (items.length === 0) return <Empty text="还没有收藏" />;

    return (
        <SeriesGrid
            items={items.map((f) => ({
                id: f.seriesId,
                title: f.title,
                coverUrl: f.coverUrl,
            }))}
        />
    );
}
