import { useState } from 'react';

interface Props {
    src?: string;
    alt?: string;
    className?: string;
}

// Cover 在图片缺失或加载失败时显示占位。
export function Cover({ src, alt, className }: Props) {
    const [failed, setFailed] = useState(false);
    if (!src || failed) {
        return (
            <div
                className={`flex items-center justify-center bg-zinc-800 text-xs text-zinc-600 ${className ?? ''}`}
            >
                暂无封面
            </div>
        );
    }
    return (
        <img
            src={src}
            alt={alt ?? ''}
            loading="lazy"
            onError={() => setFailed(true)}
            className={className}
        />
    );
}
