/**
 * Argo Proxy - Service Worker (MV3)
 */
import { init, getHostListByTabId } from './request_hook';
import {
    MSG_TYPE_GET_CONFIG,
    MSG_TYPE_SET_ENABLED,
    MSG_TYPE_SET_RULE_PROXY,
    MSG_TYPE_GET_HOST_LIST,
} from '../shared/messages';

init();

// ---- Types (local mirror of shared/types to avoid chunk imports) ----

type ProxyType = 'http' | 'https' | 'socks5';

interface Proxy {
    id: number;
    name: string;
    type: ProxyType;
    host: string;
    port: number;
    username?: string;
    password?: string;
}

type RuleType =
    | 'host_contains'
    | 'host_regex'
    | 'url_contains'
    | 'url_regex'
    | 'path_contains'
    | 'path_regex';

type ActionType = 'proxy' | 'block' | 'direct';

interface RuleAction {
    type: ActionType;
    proxyId?: number;
}

interface Rule {
    id: number;
    type: RuleType;
    value: string;
    enabled: boolean;
    action: RuleAction;
}

interface Profile {
    id: number;
    name: string;
    rules: Rule[];
    defaultAction: RuleAction;
}

interface StoredConfig {
    enabled: boolean;
    proxies: Proxy[];
    profiles: Profile[];
    activeProfileId: number | null;
    nextProxyId: number;
    nextProfileId: number;
    nextRuleId: number;
}

// ---- Storage ----

const CONFIG_KEY = 'config';
const RULES_LIMIT = 50;
const GATEWAY_BASE = 'https://argo-proxy-gateway.example.com/fetch';

async function getStoredConfig(): Promise<StoredConfig> {
    const out = await chrome.storage.local.get([CONFIG_KEY]);
    const raw = out[CONFIG_KEY];
    if (raw && typeof raw === 'object' && !Array.isArray(raw)) {
        const data = raw as Partial<StoredConfig>;
        return {
            enabled: data.enabled !== false,
            proxies: Array.isArray(data.proxies) ? data.proxies : [],
            profiles: Array.isArray(data.profiles) ? data.profiles : [],
            activeProfileId: data.activeProfileId ?? null,
            nextProxyId: data.nextProxyId ?? 1,
            nextProfileId: data.nextProfileId ?? 1,
            nextRuleId: data.nextRuleId ?? 1,
        };
    }
    return {
        enabled: true,
        proxies: [],
        profiles: [],
        activeProfileId: null,
        nextProxyId: 1,
        nextProfileId: 1,
        nextRuleId: 1,
    };
}

// ---- DNR Rule Building ----

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

function escapeRegex(s: string): string {
    return s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

function ruleToRegexFilter(rule: Rule): string {
    const v = rule.value;
    switch (rule.type) {
        case 'host_contains':
            return `^https?://[^/]*${escapeRegex(v)}[^/]*(:|/|$)`;
        case 'host_regex':
            return `^https?://${v}(:|/|$)`;
        case 'url_contains':
            return `.*${escapeRegex(v)}.*`;
        case 'url_regex':
            return v;
        case 'path_contains':
            return `^https?://[^/]+[^?]*${escapeRegex(v)}`;
        case 'path_regex':
            return `^https?://[^/]+${v}`;
        default:
            return `.*${escapeRegex(v)}.*`;
    }
}

function buildDnrRules(
    enabled: boolean,
    rules: Rule[],
    proxies: Proxy[]
): chrome.declarativeNetRequest.Rule[] {
    if (!enabled || !rules.length) return [];

    const activeRules = rules.filter((r) => r.enabled).slice(0, RULES_LIMIT);

    const dnrRules: chrome.declarativeNetRequest.Rule[] = [];
    let dnrId = 1;

    for (const rule of activeRules) {
        let regexFilter: string;
        try {
            regexFilter = ruleToRegexFilter(rule);
        } catch {
            continue;
        }

        if (rule.action.type === 'block') {
            dnrRules.push({
                id: dnrId++,
                priority: 1,
                action: { type: 'block' },
                condition: {
                    regexFilter,
                    resourceTypes: [...DNR_RESOURCE_TYPES],
                },
            } as chrome.declarativeNetRequest.Rule);
        } else if (rule.action.type === 'proxy') {
            const proxy = proxies.find((p) => p.id === rule.action.proxyId);
            if (!proxy) continue;
            dnrRules.push({
                id: dnrId++,
                priority: 1,
                action: {
                    type: 'redirect',
                    redirect: {
                        regexSubstitution:
                            GATEWAY_BASE +
                            '?url=\\0&proxyId=' +
                            encodeURIComponent(proxy.id.toString()),
                    },
                },
                condition: {
                    regexFilter,
                    resourceTypes: [...DNR_RESOURCE_TYPES],
                },
            } as chrome.declarativeNetRequest.Rule);
        }
        // 'direct' rules: no DNR rule needed
    }

    return dnrRules;
}

async function applyDnrRules(): Promise<void> {
    try {
        const cfg = await getStoredConfig();
        const activeProfile = cfg.profiles.find((p) => p.id === cfg.activeProfileId);
        const rules = activeProfile?.rules ?? [];

        const dynamicRules = buildDnrRules(cfg.enabled, rules, cfg.proxies);
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
        console.error('[Argo Proxy] applyDnrRules', e);
    }
}

// ---- Storage listener ----

chrome.storage.onChanged.addListener(
    (changes: { [key: string]: chrome.storage.StorageChange }, areaName: string) => {
        if (areaName !== 'local') return;
        if (Object.keys(changes).includes(CONFIG_KEY)) {
            applyDnrRules().catch((e) =>
                console.error('[Argo Proxy] storage.onChanged -> applyDnrRules', e)
            );
        }
    }
);

// ---- Message handlers ----

type IncomingMessage =
    | { type: typeof MSG_TYPE_GET_CONFIG }
    | { type: typeof MSG_TYPE_SET_ENABLED; enabled: boolean }
    | { type: typeof MSG_TYPE_SET_RULE_PROXY; ruleId: string; proxyId: string }
    | { type: typeof MSG_TYPE_GET_HOST_LIST; tabId: number };

type SendResponse = (response?: unknown) => void;

function handleGetConfig(sendResponse: SendResponse): boolean {
    getStoredConfig()
        .then(sendResponse)
        .catch((e) => {
            console.error('[Argo Proxy] getConfig', e);
            sendResponse({ enabled: true, proxies: [], profiles: [], activeProfileId: null });
        });
    return true;
}

function handleSetEnabled(enabled: boolean, sendResponse: SendResponse): boolean {
    getStoredConfig()
        .then(async (cfg) => {
            cfg.enabled = enabled;
            await chrome.storage.local.set({ [CONFIG_KEY]: cfg });
            sendResponse({ ok: true });
        })
        .catch((e) => {
            console.error('[Argo Proxy] setEnabled', e);
            sendResponse({ ok: false, error: String(e) });
        });
    return true;
}

function handleSetRuleProxy(
    ruleId: string,
    proxyId: string,
    sendResponse: SendResponse
): boolean {
    getStoredConfig()
        .then(async (cfg) => {
            let found = false;
            cfg.profiles = cfg.profiles.map((p) => ({
                ...p,
                rules: p.rules.map((r) => {
                    if (r.id.toString() === ruleId) {
                        found = true;
                        return {
                            ...r,
                            action: { type: 'proxy' as const, proxyId: parseInt(proxyId) },
                        };
                    }
                    return r;
                }),
            }));
            if (!found) {
                sendResponse({ ok: false, error: 'rule not found' });
                return;
            }
            await chrome.storage.local.set({ [CONFIG_KEY]: cfg });
            sendResponse({ ok: true });
        })
        .catch((e) => {
            console.error('[Argo Proxy] setRuleProxy', e);
            sendResponse({ ok: false, error: String(e) });
        });
    return true;
}

function handleGetHostList(tabId: number, sendResponse: SendResponse): boolean {
    try {
        const list = getHostListByTabId(tabId);
        sendResponse(list);
    } catch (e) {
        console.error('[Argo Proxy] getHostList', e);
        sendResponse([]);
    }
    return true;
}

chrome.runtime.onMessage.addListener(
    (
        msg: IncomingMessage,
        _sender: chrome.runtime.MessageSender,
        sendResponse: SendResponse
    ): boolean => {
        switch (msg.type) {
            case MSG_TYPE_GET_CONFIG:
                return handleGetConfig(sendResponse);
            case MSG_TYPE_SET_ENABLED:
                return handleSetEnabled(msg.enabled, sendResponse);
            case MSG_TYPE_SET_RULE_PROXY:
                return handleSetRuleProxy(msg.ruleId, msg.proxyId, sendResponse);
            case MSG_TYPE_GET_HOST_LIST:
                return handleGetHostList(msg.tabId, sendResponse);
            default:
                return false;
        }
    }
);

// ---- Init ----

chrome.runtime.onInstalled.addListener(() => {
    getStoredConfig()
        .then(async (cfg) => {
            // Already initialized if we have a stored config
            const raw = await chrome.storage.local.get([CONFIG_KEY]);
            if (!raw[CONFIG_KEY]) {
                await chrome.storage.local.set({ [CONFIG_KEY]: cfg });
            }
            void applyDnrRules();
        })
        .catch((e) => console.error('[Argo Proxy] onInstalled', e));
});

void applyDnrRules();
