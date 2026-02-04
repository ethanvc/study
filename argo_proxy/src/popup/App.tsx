import { useEffect, useState } from 'react';
import { useConfigStore } from '../shared/store';
import { matchCurrentPage, proxyIdToName } from '../shared/match';
import { PROXY_ID_DIRECT } from '../shared/types';

export default function App() {
    const { enabled, rules, proxies, setEnabled, setRuleProxy, loadConfig } = useConfigStore();
    const [view, setView] = useState<'main' | 'host'>('main');
    const [currentPage, setCurrentPage] = useState<{ host: string; proxyName: string } | null>(null);
    const [currentPageError, setCurrentPageError] = useState(false);

    useEffect(() => {
        loadConfig();
    }, [loadConfig]);

    useEffect(() => {
        if (view !== 'host') return;
        setCurrentPage(null);
        setCurrentPageError(false);
        chrome.tabs?.query({ active: true, currentWindow: true }, (tabs) => {
            const tab = tabs?.[0];
            const url = tab?.url;
            if (!url || url.startsWith('chrome://') || url.startsWith('edge://') || url.startsWith('about:')) {
                setCurrentPageError(true);
                return;
            }
            const matched = matchCurrentPage(url, enabled ? rules : []);
            let hostLabel = url;
            try {
                const u = new URL(url);
                hostLabel = u.hostname + (u.pathname !== '/' ? u.pathname : '');
            } catch {
                // keep url
            }
            const proxyName = matched ? proxyIdToName(matched.proxyId, proxies, '直连') : '直连';
            setCurrentPage({ host: hostLabel, proxyName });
        });
    }, [view, enabled, rules, proxies]);

    const sortedRules = [...rules].sort((a, b) => (a.order ?? 0) - (b.order ?? 0));

    if (view === 'main') {
        return (
            <div className="p-3 bg-gray-100 min-w-[260px]">
                <div className="flex items-center justify-between gap-4 p-3 mb-2 bg-white rounded-lg">
                    <span className="font-medium text-sm">启用代理路由</span>
                    <button
                        type="button"
                        role="switch"
                        aria-checked={enabled}
                        onClick={() => setEnabled(!enabled)}
                        className={`relative w-10 h-6 rounded-full flex-shrink-0 transition-colors ${enabled ? 'bg-green-500' : 'bg-gray-400'
                            }`}
                    >
                        <span
                            className={`absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full transition-transform ${enabled ? 'translate-x-0' : 'translate-x-4'
                                }`}
                        />
                    </button>
                </div>
                <button
                    type="button"
                    onClick={() => setView('host')}
                    className="flex items-center justify-between w-full gap-2 px-3 py-2.5 mb-1 text-left bg-white rounded-lg hover:bg-gray-100 text-sm font-medium"
                >
                    <span>Host 列表</span>
                    <span className={`text-xs ${!enabled ? 'text-orange-600' : 'text-gray-500'}`}>
                        {enabled ? `${rules.length} 条` : '已暂停'}
                    </span>
                    <span className="text-gray-400">›</span>
                </button>
                <button
                    type="button"
                    onClick={() => chrome.runtime?.openOptionsPage?.()}
                    className="flex items-center justify-between w-full px-3 py-2.5 text-left bg-white rounded-lg hover:bg-gray-100 text-sm font-medium"
                >
                    设置
                </button>
            </div>
        );
    }

    return (
        <div className="p-3 bg-gray-100 min-w-[340px]">
            <div className="flex items-center gap-2 mb-3">
                <button
                    type="button"
                    onClick={() => setView('main')}
                    className="flex items-center justify-center w-7 h-7 bg-white rounded-md hover:bg-gray-100 text-gray-600"
                    aria-label="返回"
                >
                    ‹
                </button>
                <span className="font-semibold text-sm">Host 列表</span>
            </div>

            {currentPageError && (
                <div className="p-3 mb-3 bg-white rounded-lg text-sm">
                    <div className="text-gray-500 text-xs">当前页面</div>
                    <span className="text-gray-700">无法获取当前页面</span>
                </div>
            )}
            {currentPage && !currentPageError && (
                <div className="p-3 mb-3 bg-white rounded-lg text-sm">
                    <div className="text-gray-500 text-xs">当前页面</div>
                    <span className="font-mono text-gray-800">{currentPage.host}</span>
                    <span className="text-blue-600 ml-1">→ {currentPage.proxyName}</span>
                </div>
            )}

            {sortedRules.length === 0 ? (
                <div className="p-6 text-center bg-white rounded-lg text-gray-500 text-sm">
                    暂无规则
                    <br />
                    <button
                        type="button"
                        onClick={() => chrome.runtime?.openOptionsPage?.()}
                        className="mt-2 text-blue-600 hover:underline"
                    >
                        去设置页添加
                    </button>
                </div>
            ) : (
                <div className="bg-white rounded-lg overflow-hidden">
                    {sortedRules.map((rule) => (
                        <div
                            key={rule.id}
                            className="flex items-center gap-1.5 px-3 py-2.5 border-b border-gray-100 last:border-0 text-sm"
                        >
                            <span className="flex-1 min-w-0 truncate font-mono text-gray-800">{rule.value}</span>
                            <span className="text-gray-400 text-[10px]">→</span>
                            <select
                                className="flex-shrink-0 w-[120px] ml-auto py-1 px-2 text-xs border border-gray-300 rounded bg-white text-blue-600 cursor-pointer"
                                aria-label="选择代理"
                                value={rule.proxyId || PROXY_ID_DIRECT}
                                onChange={(e) => setRuleProxy(rule.id, e.target.value)}
                            >
                                <option value={PROXY_ID_DIRECT}>直连</option>
                                {proxies.map((p) => (
                                    <option key={p.id} value={p.id}>
                                        {p.name}
                                    </option>
                                ))}
                            </select>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}
