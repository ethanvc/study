import type { Rule } from './types';

export function matchCurrentPage(url: string, rules: Rule[]): Rule | null {
    if (!url || !rules?.length) return null;
    let host = '';
    let pathname = '/';
    try {
        const u = new URL(url);
        host = u.hostname || '';
        pathname = u.pathname || '/';
    } catch {
        return null;
    }
    const sorted = [...rules].sort((a, b) => (a.order ?? 0) - (b.order ?? 0));
    for (const rule of sorted) {
        if (rule.matchType === 'pathPrefix') {
            if (pathname.startsWith(rule.value)) return rule;
        } else {
            const domain = (rule.value || '').replace(/^\./, '');
            if (host === domain || host.endsWith('.' + domain)) return rule;
        }
    }
    return null;
}

export function proxyIdToName(proxyId: string, proxies: { id: string; name: string }[], directLabel: string): string {
    if (proxyId === 'direct' || !proxyId) return directLabel;
    const p = proxies.find((x) => x.id === proxyId);
    return p ? p.name : proxyId;
}
