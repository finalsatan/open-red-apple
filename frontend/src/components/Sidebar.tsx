import { NavLink } from 'react-router-dom';

const ITEMS = [
    { to: '/', label: '发现', icon: '🏠', end: true },
    { to: '/rank', label: '榜单', icon: '🏆' },
    { to: '/new', label: '上新', icon: '🆕' },
    { to: '/favorites', label: '收藏', icon: '❤️' },
    { to: '/history', label: '记录', icon: '🕘' },
    { to: '/settings', label: '设置', icon: '⚙️' },
];

export function Sidebar() {
    return (
        <aside className="flex w-44 shrink-0 flex-col border-r border-zinc-800 bg-zinc-900/50 p-3">
            <div className="mb-5 px-2 py-2 text-lg font-bold text-orange-400">红果短剧</div>
            <nav className="flex flex-col gap-1">
                {ITEMS.map((it) => (
                    <NavLink
                        key={it.to}
                        to={it.to}
                        end={it.end}
                        className={({ isActive }) =>
                            `flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition ${
                                isActive
                                    ? 'bg-orange-500/15 text-orange-400'
                                    : 'text-zinc-300 hover:bg-zinc-800'
                            }`
                        }
                    >
                        <span className="w-5 text-center">{it.icon}</span>
                        {it.label}
                    </NavLink>
                ))}
            </nav>
            <div className="mt-auto px-2 text-xs leading-5 text-zinc-600">
                学习用途
                <br />
                禁止商用
            </div>
        </aside>
    );
}
