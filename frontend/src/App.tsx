import { HashRouter, Route, Routes } from 'react-router-dom';
import { Sidebar } from './components/Sidebar';
import { TopBar } from './components/TopBar';
import DiscoverPage from './pages/Discover';
import SearchPage from './pages/Search';
import SeriesDetailPage from './pages/SeriesDetail';
import RankPage from './pages/Rank';
import NewReleasesPage from './pages/NewReleases';
import PersonPage from './pages/Person';
import FavoritesPage from './pages/Favorites';
import HistoryPage from './pages/History';
import SettingsPage from './pages/Settings';
import PlayPlaceholder from './pages/PlayPlaceholder';

export default function App() {
    return (
        <HashRouter>
            <div className="flex h-full">
                <Sidebar />
                <div className="flex min-w-0 flex-1 flex-col">
                    <TopBar />
                    <main className="flex-1 overflow-y-auto px-6 py-5">
                        <Routes>
                            <Route path="/" element={<DiscoverPage />} />
                            <Route path="/search" element={<SearchPage />} />
                            <Route path="/series/:id" element={<SeriesDetailPage />} />
                            <Route path="/play/:seriesId/:episodeId" element={<PlayPlaceholder />} />
                            <Route path="/rank" element={<RankPage />} />
                            <Route path="/new" element={<NewReleasesPage />} />
                            <Route path="/person/:id" element={<PersonPage />} />
                            <Route path="/favorites" element={<FavoritesPage />} />
                            <Route path="/history" element={<HistoryPage />} />
                            <Route path="/settings" element={<SettingsPage />} />
                        </Routes>
                    </main>
                </div>
            </div>
        </HashRouter>
    );
}
