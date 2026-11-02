import { useState } from 'react';
import { ClearHistory, ClearSearchHistory, ResetDevice } from '../lib/wails';
import { errorText } from '../lib/errors';

export default function SettingsPage() {
    const [message, setMessage] = useState('');

    async function run(action: () => Promise<void>, ok: string) {
        setMessage('');
        try {
            await action();
            setMessage(ok);
        } catch (e) {
            setMessage(errorText(e));
        }
    }

    return (
        <div className="max-w-2xl space-y-6">
            <h1 className="text-2xl font-bold">设置</h1>

            {message && (
                <div className="rounded-lg border border-zinc-700 bg-zinc-900 px-4 py-2 text-sm text-zinc-300">
                    {message}
                </div>
            )}

            <Row
                title="更换匿名设备"
                desc="当前设备无法播放或返回空内容时，注册一台新的匿名设备。"
                action="更换设备"
                onClick={() => run(ResetDevice, '已注册新的匿名设备')}
            />
            <Row
                title="清空搜索记录"
                desc="删除本机保存的搜索历史。"
                action="清空"
                onClick={() => run(ClearSearchHistory, '搜索记录已清空')}
            />
            <Row
                title="清空观看记录"
                desc="删除本机保存的观看进度。"
                action="清空"
                onClick={() => run(ClearHistory, '观看记录已清空')}
            />

            <div className="rounded-lg border border-zinc-800 p-4 text-sm text-zinc-400">
                <div className="font-medium text-zinc-200">关于</div>
                <p className="mt-1">
                    win-red-apple 是基于 openhongguo SDK 的 Windows 桌面短剧客户端，
                    仅供学习和研究使用，禁止商用。
                </p>
            </div>
        </div>
    );
}

function Row({
    title,
    desc,
    action,
    onClick,
}: {
    title: string;
    desc: string;
    action: string;
    onClick: () => void;
}) {
    return (
        <div className="flex items-center justify-between gap-4 rounded-lg border border-zinc-800 p-4">
            <div className="min-w-0">
                <div className="text-sm font-medium text-zinc-200">{title}</div>
                <div className="mt-0.5 text-xs text-zinc-500">{desc}</div>
            </div>
            <button
                onClick={onClick}
                className="shrink-0 rounded-lg bg-zinc-800 px-4 py-1.5 text-sm hover:bg-zinc-700"
            >
                {action}
            </button>
        </div>
    );
}
