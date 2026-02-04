import { useEffect, useState } from 'react';
import { useConfigStore } from '../shared/store';
import { matchCurrentPage, proxyIdToName } from '../shared/match';
import EnableProxySwitch from './components/EnableProxySwitch';
import HostListView from './components/HostListView';

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
            if (
                !url ||
                url.startsWith('chrome://') ||
                url.startsWith('edge://') ||
                url.startsWith('about:')
            ) {
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
            const proxyName = matched
                ? proxyIdToName(matched.proxyId, proxies, '直连')
                : '直连';
            setCurrentPage({ host: hostLabel, proxyName });
        });
    }, [view, enabled, rules, proxies]);

    const sortedRules = [...rules].sort((a, b) => (a.order ?? 0) - (b.order ?? 0));

    const openSettings = () => chrome.runtime?.openOptionsPage?.();

    if (view === 'main') {
        return (
            <div className="p-3 bg-gray-100 min-w-[260px]">
                <EnableProxySwitch enabled={enabled} onToggle={() => setEnabled(!enabled)} />
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
                    onClick={openSettings}
                    className="flex items-center justify-between w-full px-3 py-2.5 text-left bg-white rounded-lg hover:bg-gray-100 text-sm font-medium"
                >
                    设置
                </button>
            </div>
        );
    }

    return (
        <HostListView
            currentPage={currentPage}
            currentPageError={currentPageError}
            sortedRules={sortedRules}
            proxies={proxies}
            onBack={() => setView('main')}
            onSetRuleProxy={setRuleProxy}
            onOpenSettings={openSettings}
        />
    );
}
