/**
 * Argo Proxy - Request Logger
 * 监听 Chrome HTTP 请求，记录请求时间、结果与 tab 关系，输出到 console。
 */

/** 单条请求记录 */
export interface RequestRecord {
    requestId: string;
    tabId: number;
    url: string;
    method: string;
    type: string;
    startTime: number;
    endTime?: number;
    statusCode?: number;
    statusLine?: string;
    fromCache?: boolean;
    error?: string;
}

/** 请求统计 */
export interface RequestStats {
    tabId: number;
    count: number;
}

const MAX_RECORDS_PER_TAB = 200;
const requestLogByTab = new Map<number, RequestRecord[]>();
const pendingRequests = new Map<string, RequestRecord>();

function log(msg: string, data?: unknown): void {
    if (data !== undefined) {
        console.log('[Argo Proxy][RequestList]', msg, data);
    } else {
        console.log('[Argo Proxy][RequestList]', msg);
    }
}

function addRecordToTab(record: RequestRecord): void {
    const { tabId } = record;
    if (tabId < 0) return;

    let list = requestLogByTab.get(tabId);
    if (!list) {
        list = [];
        requestLogByTab.set(tabId, list);
    }
    list.push(record);
    if (list.length > MAX_RECORDS_PER_TAB) {
        list.shift();
    }
}

function onBeforeRequest(details: chrome.webRequest.WebRequestBodyDetails): void {
    const record: RequestRecord = {
        requestId: details.requestId,
        tabId: details.tabId,
        url: details.url,
        method: details.method,
        type: details.type,
        startTime: details.timeStamp,
    };
    pendingRequests.set(details.requestId, record);
    log('Request started', {
        tabId: record.tabId,
        method: record.method,
        url: record.url,
    });
}

function onCompleted(details: chrome.webRequest.WebResponseCacheDetails): void {
    const record = pendingRequests.get(details.requestId);
    if (!record) return;
    pendingRequests.delete(details.requestId);

    record.endTime = details.timeStamp;
    record.statusCode = details.statusCode;
    record.statusLine = details.statusLine;
    record.fromCache = details.fromCache;

    addRecordToTab(record);

    const duration = record.endTime - record.startTime;
    log('Request completed', {
        tabId: record.tabId,
        status: record.statusCode,
        duration: `${duration.toFixed(0)}ms`,
        url: record.url,
    });
}

function onErrorOccurred(details: chrome.webRequest.WebResponseErrorDetails): void {
    const record = pendingRequests.get(details.requestId);
    if (!record) return;
    pendingRequests.delete(details.requestId);

    record.endTime = details.timeStamp;
    record.error = details.error;

    addRecordToTab(record);

    const duration = record.endTime - record.startTime;
    log('Request error', {
        tabId: record.tabId,
        error: record.error,
        duration: `${duration.toFixed(0)}ms`,
        url: record.url,
    });
}

function onTabRemoved(tabId: number): void {
    if (requestLogByTab.has(tabId)) {
        requestLogByTab.delete(tabId);
        log('Tab closed, cleared request log', { tabId });
    }
}

/** 获取某 tab 的请求列表 */
export function getRequestsForTab(tabId: number): RequestRecord[] {
    return requestLogByTab.get(tabId) ?? [];
}

/** 获取所有 tab 的请求统计 */
export function getRequestStats(): RequestStats[] {
    const stats: RequestStats[] = [];
    requestLogByTab.forEach((list, tabId) => {
        stats.push({ tabId, count: list.length });
    });
    return stats;
}

/** 初始化请求监听器 */
function initRequestLogger(): void {
    try {
        // 检查 chrome 和 webRequest API 是否可用
        if (typeof chrome === 'undefined' || !chrome.webRequest) {
            log('webRequest API not available, skipping initialization');
            return;
        }

        const filter: chrome.webRequest.RequestFilter = { urls: ['<all_urls>'] };

        chrome.webRequest.onBeforeRequest.addListener(onBeforeRequest, filter);
        chrome.webRequest.onCompleted.addListener(onCompleted, filter);
        chrome.webRequest.onErrorOccurred.addListener(onErrorOccurred, filter);

        if (chrome.tabs && chrome.tabs.onRemoved) {
            chrome.tabs.onRemoved.addListener(onTabRemoved);
        }

        log('Request logger initialized');
    } catch (e) {
        log('Failed to initialize request logger', e);
    }
}

// 自动初始化（import 时立即执行）
try {
    console.log('initRequestLogger start');
    initRequestLogger();
} catch (e) {
    console.error('[Argo Proxy][RequestList] Init error:', e);
}
