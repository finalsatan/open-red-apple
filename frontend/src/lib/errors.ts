// 把 Wails 抛出的错误字符串转成可读文案。
const MESSAGES: Array<[string, string]> = [
    ['RATE_LIMITED', '请求太频繁，请稍后再试'],
    ['ANONYMOUS_LIMITED', '服务端没有返回内容，请重试'],
    ['CONTENT_UNAVAILABLE', '内容已下架或不可用'],
    ['AUTHENTICATION_REQUIRED', '该内容需要登录，暂不支持'],
    ['DEVICE_REJECTED', '设备被拒绝，请到设置里更换设备'],
    ['TRANSPORT', '网络请求失败，请检查网络'],
    ['CHALLENGE_REQUIRED', '服务端要求验证，暂不支持'],
    ['END_OF_LIST', '没有更多内容了'],
];

export function errorText(e: unknown): string {
    const raw =
        typeof e === 'string' ? e : e instanceof Error ? e.message : String(e ?? '');
    for (const [code, text] of MESSAGES) {
        if (raw.includes(code)) return text;
    }
    return raw || '发生未知错误';
}
