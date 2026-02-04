import type { Proxy, Rule } from '../../shared/types';
import { PROXY_ID_DIRECT } from '../../shared/types';

interface CurrentPageInfo {
    host: string;
    proxyName: string;
}

interface HostListViewProps {
    currentPage: CurrentPageInfo | null;
    currentPageError: boolean;
    sortedRules: Rule[];
    proxies: Proxy[];
    onBack: () => void;
    onSetRuleProxy: (ruleId: string, proxyId: string) => void;
    onOpenSettings: () => void;
}

export default function HostListView({
    currentPage,
    currentPageError,
    sortedRules,
    proxies,
    onBack,
    onSetRuleProxy,
    onOpenSettings,
}: HostListViewProps) {
    return (
        <div className="p-3 bg-gray-100 min-w-[340px]">
            <div className="flex items-center gap-2 mb-3">
                <button
                    type="button"
                    onClick={onBack}
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
        </div>
    );
}
