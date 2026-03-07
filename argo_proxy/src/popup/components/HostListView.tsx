import type { HostStatusItem } from '../App';

interface CurrentPageInfo {
    host: string;
    proxyName: string;
}

interface HostListViewProps {
    currentPage: CurrentPageInfo | null;
    currentPageError: boolean;
    hostList: HostStatusItem[];
    hostListLoading: boolean;
    onBack: () => void;
    onOpenSettings: () => void;
}

function statusLabel(status: string): string {
    switch (status) {
        case 'pending':
            return '请求中';
        case 'ok':
            return '成功';
        case 'failed':
            return '失败';
        case 'timeout':
            return '超时';
        default:
            return status;
    }
}

function statusClass(status: string): string {
    switch (status) {
        case 'ok':
            return 'text-primary';
        case 'failed':
        case 'timeout':
            return 'text-destructive';
        default:
            return 'text-muted-foreground';
    }
}

export default function HostListView({
    currentPage,
    currentPageError,
    hostList,
    hostListLoading,
    onBack,
}: HostListViewProps) {
    return (
        <div className="p-3 bg-background min-w-[340px]">
            <div className="flex items-center gap-2 mb-3">
                <button
                    type="button"
                    onClick={onBack}
                    className="flex items-center justify-center w-7 h-7 bg-card border border-border rounded-md hover:bg-secondary text-foreground"
                    aria-label="返回"
                >
                    ‹
                </button>
                <span className="font-semibold text-sm">Host 列表</span>
            </div>

            {currentPageError && (
                <div className="p-3 mb-3 bg-card border border-border rounded-lg text-sm">
                    <div className="text-muted-foreground text-xs">当前页面</div>
                    <span className="text-foreground">无法获取当前页面</span>
                </div>
            )}
            {currentPage && !currentPageError && (
                <div className="p-3 mb-3 bg-card border border-border rounded-lg text-sm">
                    <div className="text-muted-foreground text-xs">当前页面</div>
                    <span className="font-mono text-foreground">{currentPage.host}</span>
                    <span className="text-primary ml-1">→ {currentPage.proxyName}</span>
                </div>
            )}

            <div className="mt-2">
                <div className="text-muted-foreground text-xs mb-1">本页 Host 访问状态</div>
                {hostListLoading ? (
                    <div className="p-3 bg-card border border-border rounded-lg text-sm text-muted-foreground">加载中...</div>
                ) : hostList.length === 0 ? (
                    <div className="p-3 bg-card border border-border rounded-lg text-sm text-muted-foreground">暂无记录</div>
                ) : (
                    <ul className="space-y-1">
                        {hostList.map((item) => (
                            <li
                                key={item.host}
                                className="flex items-center justify-between gap-2 p-2 bg-card border border-border rounded-lg text-sm"
                            >
                                <span className="font-mono text-foreground truncate flex-1 min-w-0">
                                    {item.host}
                                </span>
                                <span className={`shrink-0 text-xs ${statusClass(item.status)}`}>
                                    {statusLabel(item.status)}
                                    {item.pendingRequestCount > 0 && ` (${item.pendingRequestCount})`}
                                </span>
                            </li>
                        ))}
                    </ul>
                )}
            </div>
        </div>
    );
}
