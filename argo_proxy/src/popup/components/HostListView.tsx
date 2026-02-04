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

            {sortedRules.length === 0 ? (
                <div className="p-6 text-center bg-white rounded-lg text-gray-500 text-sm">
                    暂无规则
                    <br />
                    <button
                        type="button"
                        onClick={onOpenSettings}
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
                            <span className="flex-1 min-w-0 truncate font-mono text-gray-800">
                                {rule.value}
                            </span>
                            <span className="text-gray-400 text-[10px]">→</span>
                            <select
                                className="flex-shrink-0 w-[120px] ml-auto py-1 px-2 text-xs border border-gray-300 rounded bg-white text-blue-600 cursor-pointer"
                                aria-label="选择代理"
                                value={rule.proxyId || PROXY_ID_DIRECT}
                                onChange={(e) => onSetRuleProxy(rule.id, e.target.value)}
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
