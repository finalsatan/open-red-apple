// 加载 / 错误 / 空状态。
export function Loading({ text = '加载中…' }: { text?: string }) {
    return <div className="py-16 text-center text-sm text-zinc-500">{text}</div>;
}

export function ErrorBox({ message, onRetry }: { message: string; onRetry?: () => void }) {
    return (
        <div className="py-16 text-center">
            <div className="text-sm text-red-400">{message}</div>
            {onRetry && (
                <button
                    onClick={onRetry}
                    className="mt-3 rounded-md bg-zinc-700 px-4 py-1.5 text-sm hover:bg-zinc-600"
                >
                    重试
                </button>
            )}
        </div>
    );
}

export function Empty({ text = '没有内容' }: { text?: string }) {
    return <div className="py-16 text-center text-sm text-zinc-500">{text}</div>;
}
