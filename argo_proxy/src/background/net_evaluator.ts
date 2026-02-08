export enum RequestStatus {
    Pending = 'pending',
    Timeout = 'timeout',
    Failed = 'failed',
    Ok = 'ok',
}

export interface RequestInfo {
    requestId: string;
    url: string;
    chromeTabId: number;
    startTime: number;
}

export interface HostInfo {
    host: string;
    lastAccessTime: number;
    requests: Map<string, RequestInfo>;
    status: RequestStatus;
}

const MAX_REQUEST_COUNT = 5;
const CLEANUP_INTERVAL_MS = 1 * 5 * 1000; // 3 minutes
const HOST_EXPIRE_MS = 2 * 60 * 1000; // 2 minutes

/**
 * 按 tab -> host -> requestId 存储请求，用于按 tab/域名维度统计或查询。
 */
export class NetEvaluator {
    /** chromeTabId -> host -> HostInfo */
    private readonly storage = new Map<number, Map<string, HostInfo>>();
    private readonly requestIndexMap = new Map<string, RequestInfo>();
    private readonly cleanupTimer: ReturnType<typeof setInterval>;

    constructor() {
        this.cleanupTimer = setInterval(() => {
            this.clearExpiredData();
        }, CLEANUP_INTERVAL_MS);
    }

    /** 停止定时清理（如需销毁实例时调用）。 */
    dispose(): void {
        clearInterval(this.cleanupTimer);
    }

    /** 清理过期数据：移除不存在的 tab 和超过2分钟未访问的 host。 */
    async clearExpiredData(): Promise<void> {
        for (const [tabId, hostMap] of this.storage) {
            const tabExists = await this.isTabExists(tabId);
            if (!tabExists) {
                this.removeTabAndCleanIndex(tabId);
                console.log(`clearExpiredData: removed closed tab ${tabId}`);
                continue;
            }

            const now = Date.now();
            for (const [host, hostInfo] of hostMap) {
                if (now - hostInfo.lastAccessTime > HOST_EXPIRE_MS) {
                    // 清理 requestIndexMap 中该 host 的请求
                    for (const requestId of hostInfo.requests.keys()) {
                        this.requestIndexMap.delete(requestId);
                    }
                    hostMap.delete(host);
                    console.log(`clearExpiredData: removed expired host ${host} from tab ${tabId}`);
                }
            }
            // 如果 tab 下已无 host，移除整个 tab
            if (hostMap.size === 0) {
                this.storage.delete(tabId);
            }
        }
        console.log(`clearExpiredData finished, remain tab size: `, this.storage.size);
    }

    private async isTabExists(tabId: number): Promise<boolean> {
        try {
            await chrome.tabs.get(tabId);
            return true;
        } catch {
            return false;
        }
    }

    private removeTabAndCleanIndex(tabId: number): void {
        const hostMap = this.storage.get(tabId);
        if (hostMap) {
            for (const hostInfo of hostMap.values()) {
                for (const requestId of hostInfo.requests.keys()) {
                    this.requestIndexMap.delete(requestId);
                }
            }
        }
        this.storage.delete(tabId);
    }

    addRequest(chromeTabId: number, url: string, requestId: string): void {
        const host = this.getHostFromUrl(url);
        if (host === '') {
            return;
        }
        let hostMap = this.storage.get(chromeTabId);
        if (!hostMap) {
            hostMap = new Map();
            this.storage.set(chromeTabId, hostMap);
        }
        let hostInfo = hostMap.get(host);
        if (!hostInfo) {
            hostInfo = { host, lastAccessTime: Date.now(), requests: new Map(), status: RequestStatus.Pending };
            hostMap.set(host, hostInfo);
        }
        if (hostInfo.requests.size > MAX_REQUEST_COUNT) {
            console.info(`request count exceed, id:${requestId}, url: ${url}, tabId: ${chromeTabId}`);
            return;
        }
        const requestInfo: RequestInfo = {
            requestId,
            url,
            chromeTabId,
            startTime: Date.now(),
        };
        hostInfo.requests.set(requestId, requestInfo);
        this.requestIndexMap.set(requestId, requestInfo);
        console.log(`add request, id:${requestId}, url: ${url}, tabId: ${chromeTabId}`);
    }

    finishRequest(requestId: string, status: RequestStatus): void {
        const requestInfo = this.requestIndexMap.get(requestId);
        if (!requestInfo) {
            return;
        }
        this.requestIndexMap.delete(requestId);
        const tabHosts = this.storage.get(requestInfo.chromeTabId);
        if (!tabHosts) {
            return;
        }
        const hostInfo = tabHosts.get(this.getHostFromUrl(requestInfo.url));
        if (!hostInfo) {
            return;
        }
        hostInfo.lastAccessTime = Date.now();
        hostInfo.status = status;
        hostInfo.requests.delete(requestId);
        console.log(`finish request, id:${requestId}, url: ${requestInfo.url}, tabId: ${requestInfo.chromeTabId}`);
    }

    private getHostFromUrl(url: string): string {
        try {
            return new URL(url).host;
        } catch {
            console.error(`parse url error: ${url}`);
            return '';
        }
    }

    /** 按 chromeTabId 获取该 tab 下所有 host 的 HostInfo（只读）。 */
    getByTabId(chromeTabId: number): Map<string, HostInfo> | undefined {
        return this.storage.get(chromeTabId);
    }

    /** 移除整个 tab 的数据（如 tab 关闭时调用）。 */
    removeTab(chromeTabId: number): void {
        this.storage.delete(chromeTabId);
    }
}
