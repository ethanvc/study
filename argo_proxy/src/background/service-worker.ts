/**
 * Argo Proxy - Service Worker (MV3)
 * Inlined constants so this entry is bundled as a single file (no shared chunk → no top-level import).
 */
// Request Logger: 导入并调用以确保不被 tree-shake
import { init, getHostListByTabId } from './request_hook';
// 强制保留模块（避免被 tree-shake）
init();
const STORAGE_KEYS = { ENABLED: 'enabled', PROXIES: 'proxies', RULES: 'rules' } as const;
const PROXY_ID_DIRECT = 'direct';
const RULES_LIMIT = 50;
const DEFAULT_ENABLED = true;

interface Config {
    enabled: boolean;
    proxies: { id: string; name: string; type: string; host: string; port: number }[];
    rules: Rule[];
}
interface Rule {
    id: string;
    matchType: 'domain' | 'pathPrefix';
    value: string;
    proxyId: string;
    order: number;
}

/** Placeholder: replace with real proxy gateway base URL. */
const GATEWAY_BASE = 'https://argo-proxy-gateway.example.com/fetch';

const DNR_RESOURCE_TYPES = [
    'main_frame',
    'sub_frame',
    'xmlhttprequest',
    'script',
    'stylesheet',
    'image',
    'font',
    'object',
    'ping',
    'csp_report',
    'media',
    'websocket',
    'other',
] as const;

function logError(scope: string, err: unknown): void {
    console.error('[Argo Proxy]', scope, err);
}

async function getConfig(): Promise<Config> {
    const out = await chrome.storage.local.get([
        STORAGE_KEYS.ENABLED,
        STORAGE_KEYS.PROXIES,
        STORAGE_KEYS.RULES,
    ]);
    return {
        enabled: out[STORAGE_KEYS.ENABLED] !== false,
        proxies: Array.isArray(out[STORAGE_KEYS.PROXIES]) ? out[STORAGE_KEYS.PROXIES] : [],
        rules: Array.isArray(out[STORAGE_KEYS.RULES]) ? out[STORAGE_KEYS.RULES] : [],
    };
}

async function setEnabled(enabled: boolean): Promise<void> {
    await chrome.storage.local.set({ [STORAGE_KEYS.ENABLED]: enabled });
}

function escapeRegex(s: string): string {
    return s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

function buildDnrRulesWithRegex(
    enabled: boolean,
    rules: Rule[]
): chrome.declarativeNetRequest.Rule[] {
    if (!enabled || !rules.length) {
        return [];
    }
    const sorted = [...rules]
        .filter((r) => r.proxyId && r.proxyId !== PROXY_ID_DIRECT)
        .sort((a, b) => (a.order ?? 0) - (b.order ?? 0))
        .slice(0, RULES_LIMIT);

    return sorted.map((rule, index) => {
        const regexFilter =
            rule.matchType === 'pathPrefix'
                ? '^https?:\\/\\/[^\\/]+' + escapeRegex(rule.value) + '.*'
                : '^https?:\\/\\/([^\\/]*\\.)?' +
                escapeRegex(rule.value.replace(/^\./, '')) +
                '(\\/|$).*';
        return {
            id: index + 1,
            priority: 1,
            action: {
                type: 'redirect',
                redirect: {
                    regexSubstitution:
                        GATEWAY_BASE +
                        '?url=\\0&proxyId=' +
                        encodeURIComponent(rule.proxyId),
                },
            },
            condition: {
                regexFilter,
                resourceTypes: [...DNR_RESOURCE_TYPES],
            },
        } as chrome.declarativeNetRequest.Rule;
    });
}

async function applyDnrRules(): Promise<void> {
    try {
        const { enabled, rules } = await getConfig();
        const dynamicRules = buildDnrRulesWithRegex(enabled, rules);
        const existing = await chrome.declarativeNetRequest.getDynamicRules();
        await chrome.declarativeNetRequest.updateDynamicRules({
            removeRuleIds: existing.map((r) => r.id),
        });
        if (dynamicRules.length > 0) {
            await chrome.declarativeNetRequest.updateDynamicRules({
                addRules: dynamicRules,
            });
        }
    } catch (e) {
        logError('applyDnrRules', e);
    }
}

chrome.storage.onChanged.addListener(
    (changes: { [key: string]: chrome.storage.StorageChange }, areaName: string) => {
        if (areaName !== 'local') return;
        const keySet = new Set(Object.keys(changes));
        if (
            keySet.has(STORAGE_KEYS.ENABLED) ||
            keySet.has(STORAGE_KEYS.RULES) ||
            keySet.has(STORAGE_KEYS.PROXIES)
        ) {
            applyDnrRules().catch((e) =>
                logError('storage.onChanged -> applyDnrRules', e)
            );
        }
    }
);

type IncomingMessage =
    | { type: 'getConfig' }
    | { type: 'setEnabled'; enabled: boolean }
    | { type: 'setRuleProxy'; ruleId: string; proxyId: string }
    | { type: 'getHostList'; tabId: number };

chrome.runtime.onMessage.addListener(
    (
        msg: IncomingMessage,
        _sender: chrome.runtime.MessageSender,
        sendResponse: (response?: unknown) => void
    ): boolean => {
        if (msg.type === 'getConfig') {
            getConfig()
                .then(sendResponse)
                .catch((e) => {
                    logError('getConfig', e);
                    sendResponse({ enabled: true, proxies: [], rules: [] });
                });
            return true;
        }
        if (msg.type === 'setEnabled') {
            setEnabled(msg.enabled === true)
                .then(() => sendResponse({ ok: true }))
                .catch((e) => {
                    logError('setEnabled', e);
                    sendResponse({ ok: false, error: String(e) });
                });
            return true;
        }
        if (msg.type === 'setRuleProxy') {
            chrome.storage.local.get(STORAGE_KEYS.RULES).then((out) => {
                let rules: Rule[] = Array.isArray(out[STORAGE_KEYS.RULES])
                    ? out[STORAGE_KEYS.RULES]
                    : [];
                const idx = rules.findIndex((r) => r.id === msg.ruleId);
                if (idx !== -1) {
                    rules = [...rules];
                    rules[idx] = { ...rules[idx], proxyId: msg.proxyId };
                    chrome.storage.local
                        .set({ [STORAGE_KEYS.RULES]: rules })
                        .then(() => sendResponse({ ok: true }))
                        .catch((e) => {
                            logError('setRuleProxy (set)', e);
                            sendResponse({ ok: false, error: String(e) });
                        });
                } else {
                    logError('setRuleProxy', new Error('rule not found: ' + msg.ruleId));
                    sendResponse({ ok: false, error: 'rule not found' });
                }
            });
            return true;
        }
        if (msg.type === 'getHostList') {
            try {
                const list = getHostListByTabId(msg.tabId);
                sendResponse(list);
            } catch (e) {
                logError('getHostList', e);
                sendResponse([]);
            }
            return true;
        }
        return false;
    }
);

chrome.runtime.onInstalled.addListener(() => {
    chrome.storage.local
        .get([STORAGE_KEYS.ENABLED, STORAGE_KEYS.RULES, STORAGE_KEYS.PROXIES])
        .then(
            (out: { [key: string]: unknown }) => {
                if (out[STORAGE_KEYS.ENABLED] === undefined) {
                    chrome.storage.local.set({ [STORAGE_KEYS.ENABLED]: DEFAULT_ENABLED });
                }
                if (!Array.isArray(out[STORAGE_KEYS.PROXIES])) {
                    chrome.storage.local.set({ [STORAGE_KEYS.PROXIES]: [] });
                }
                if (!Array.isArray(out[STORAGE_KEYS.RULES])) {
                    chrome.storage.local.set({ [STORAGE_KEYS.RULES]: [] });
                }
                void applyDnrRules();
            },
            (e) => logError('onInstalled getStorage', e)
        );
});

void applyDnrRules();
