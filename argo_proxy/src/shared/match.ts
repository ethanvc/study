// matchCurrentPage 函数已移除，因为 Rule 类型已被删除

export function proxyIdToName(proxyId: string, proxies: { name: string }[], directLabel: string): string {
    if (proxyId === 'direct' || !proxyId) return directLabel;
    const p = proxies.find((x) => x.name === proxyId);
    return p ? p.name : proxyId;
}
