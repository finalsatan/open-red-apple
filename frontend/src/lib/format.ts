// 数字格式化。
export function formatHeat(n: number): string {
    if (!n) return '';
    if (n >= 100_000_000) return (n / 100_000_000).toFixed(1) + '亿';
    if (n >= 10_000) return Math.round(n / 10_000) + '万';
    return String(n);
}

// 秒数转 mm:ss。
export function formatTime(seconds: number): string {
    if (!seconds || seconds < 0) return '00:00';
    const s = Math.floor(seconds % 60);
    const m = Math.floor(seconds / 60) % 60;
    const h = Math.floor(seconds / 3600);
    const pad = (v: number) => String(v).padStart(2, '0');
    return h > 0 ? `${h}:${pad(m)}:${pad(s)}` : `${pad(m)}:${pad(s)}`;
}
