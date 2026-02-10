import { useEffect, useState } from 'react';
import { useConfigStore } from '../shared/store';
import { MSG_TYPE_GET_HOST_LIST } from '../shared/messages';
import EnableProxySwitch from './components/EnableProxySwitch';
import HostListView from './components/HostListView';

export interface HostStatusItem {
    host: string;
    status: string;
    lastAccessTime: number;
    pendingRequestCount: number;
}

export default function App() {
    const { enabled, proxies, setEnabled, loadConfig } = useConfigStore();
    const [view, setView] = useState<'main' | 'host'>('main');
    const [currentPage, setCurrentPage] = useState<{ host: string; proxyName: string } | null>(null);
    const [currentPageError, setCurrentPageError] = useState(false);
    const [hostList, setHostList] = useState<HostStatusItem[]>([]);
    const [hostListLoading, setHostListLoading] = useState(false);

    useEffect(() => {
        loadConfig();
    }, [loadConfig]);

    useEffect(() => {
        if (view !== 'host') return;
        setCurrentPage(null);
        setCurrentPageError(false);
        setHostList([]);
        chrome.tabs?.query({ active: true, currentWindow: true }, (tabs) => {
            const tab = tabs?.[0];
            const url = tab?.url;
            const tabId = tab?.id;
            if (
                !url ||
                url.startsWith('chrome://') ||
                url.startsWith('edge://') ||
                url.startsWith('about:')
            ) {
                setCurrentPageError(true);
            } else {
                let hostLabel = url;
                try {
                    const u = new URL(url);
                    hostLabel = u.hostname + (u.pathname !== '/' ? u.pathname : '');
                } catch {
                    // keep url
                }
                setCurrentPage({ host: hostLabel, proxyName: '直连' });
            }
            if (tabId !== undefined) {
                setHostListLoading(true);
                chrome.runtime
                    .sendMessage({ type: MSG_TYPE_GET_HOST_LIST, tabId })
                    .then((list: HostStatusItem[]) => {
                        setHostList(Array.isArray(list) ? list : []);
                    })
                    .catch(() => setHostList([]))
                    .finally(() => setHostListLoading(false));
            }
        });
    }, [view, enabled, proxies]);

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
                        {enabled ? '已启用' : '已暂停'}
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
            hostList={hostList}
            hostListLoading={hostListLoading}
            onBack={() => setView('main')}
            onOpenSettings={openSettings}
        />
    );
}
