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

// ---- PAC Script Generation ----
//
// The PAC (Proxy Auto-Config) script is a JavaScript function evaluated by
// the browser for every request. It returns a proxy server string like
// "PROXY host:port", "SOCKS5 host:port", or "DIRECT" — without modifying
// the URL in any way.

function proxyToString(proxy: Proxy): string {
    if (proxy.type === 'socks5') {
        return `SOCKS5 ${proxy.host}:${proxy.port}`;
    }
    // http and https proxies both use the HTTP CONNECT tunnel
    return `PROXY ${proxy.host}:${proxy.port}`;
}

function escapeForPacString(s: string): string {
    return s.replace(/\\/g, '\\\\').replace(/"/g, '\\"');
}

function buildPacScript(
    rules: Rule[],
    proxies: Proxy[],
    defaultAction: RuleAction
): string {
    const activeRules = rules
        .filter((r) => r.enabled && r.action.type !== 'block')
        .slice(0, RULES_LIMIT);

    const lines: string[] = ['function FindProxyForURL(url, host) {'];
    // Extract path once for path_* rules
    lines.push('    var path = url.replace(/^[a-z]+:\\/\\/[^\\/]+/, "");');

    for (const rule of activeRules) {
        const { action } = rule;
        let proxyStr: string;
        if (action.type === 'direct') {
            proxyStr = 'DIRECT';
        } else if (action.type === 'proxy') {
            const proxy = proxies.find((p) => p.id === action.proxyId);
            if (!proxy) continue;
            proxyStr = proxyToString(proxy);
        } else {
            continue;
        }

        let condition: string;
        const v = escapeForPacString(rule.value);
        try {
            switch (rule.type) {
                case 'host_contains':
                    condition = `host.indexOf("${v}") !== -1`;
                    break;
                case 'host_regex':
                    // Validate regex before embedding
                    new RegExp(rule.value);
                    condition = `/${rule.value}/.test(host)`;
                    break;
                case 'url_contains':
                    condition = `url.indexOf("${v}") !== -1`;
                    break;
                case 'url_regex':
                    new RegExp(rule.value);
                    condition = `/${rule.value}/.test(url)`;
                    break;
                case 'path_contains':
                    condition = `path.indexOf("${v}") !== -1`;
                    break;
                case 'path_regex':
                    new RegExp(rule.value);
                    condition = `/${rule.value}/.test(path)`;
                    break;
                default:
                    continue;
            }
        } catch {
            // Invalid regex — skip this rule
            continue;
        }

        lines.push(`    if (${condition}) return "${proxyStr}";`);
    }

    // Default action
    let defaultStr = 'DIRECT';
    if (defaultAction.type === 'proxy') {
        const proxy = proxies.find((p) => p.id === defaultAction.proxyId);
        if (proxy) defaultStr = proxyToString(proxy);
    }
    // For default 'block': use a loopback address that refuses connections,
    // which causes an immediate network error without any redirect.
    if (defaultAction.type === 'block') {
        defaultStr = 'PROXY 127.0.0.1:0';
    }

    lines.push(`    return "${defaultStr}";`);
    lines.push('}');
    return lines.join('\n');
}

// ---- chrome.proxy helpers (callback → Promise) ----

function setChromeProxy(config: chrome.proxy.ProxyConfig): Promise<void> {
    return new Promise((resolve, reject) => {
        chrome.proxy.settings.set({ value: config, scope: 'regular' }, () => {
            if (chrome.runtime.lastError) {
                reject(new Error(chrome.runtime.lastError.message));
            } else {
                resolve();
            }
        });
    });
}

function clearChromeProxy(): Promise<void> {
    return new Promise((resolve, reject) => {
        chrome.proxy.settings.clear({ scope: 'regular' }, () => {
            if (chrome.runtime.lastError) {
                reject(new Error(chrome.runtime.lastError.message));
            } else {
                resolve();
            }
        });
    });
}

// ---- DNR helpers (block rules only) ----

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

function buildDnrBlockRules(
    enabled: boolean,
    rules: Rule[]
): chrome.declarativeNetRequest.Rule[] {
    if (!enabled) return [];
    return rules
        .filter((r) => r.enabled && r.action.type === 'block')
        .slice(0, RULES_LIMIT)
        .map((rule, index) => ({
            id: index + 1,
            priority: 1,
            action: { type: 'block' as const },
            condition: {
                regexFilter: ruleToRegexFilter(rule),
                resourceTypes: [...DNR_RESOURCE_TYPES],
            },
        }));
}

async function applyDnrBlockRules(enabled: boolean, rules: Rule[]): Promise<void> {
    const blockRules = buildDnrBlockRules(enabled, rules);
    const existing = await chrome.declarativeNetRequest.getDynamicRules();
    await chrome.declarativeNetRequest.updateDynamicRules({
        removeRuleIds: existing.map((r) => r.id),
        addRules: blockRules,
    });
}

// ---- Main apply function ----

async function applyProxySettings(): Promise<void> {
    try {
        const cfg = await getStoredConfig();
        const activeProfile = cfg.profiles.find((p) => p.id === cfg.activeProfileId);

        if (!cfg.enabled || !activeProfile) {
            await clearChromeProxy();
        } else {
            const pacScript = buildPacScript(
                activeProfile.rules,
                cfg.proxies,
                activeProfile.defaultAction
            );
            await setChromeProxy({
                mode: 'pac_script',
                pacScript: { data: pacScript },
            });
        }

        await applyDnrBlockRules(cfg.enabled, activeProfile?.rules ?? []);
    } catch (e) {
        console.error('[Argo Proxy] applyProxySettings', e);
    }
}

// ---- Storage listener ----

chrome.storage.onChanged.addListener(
    (changes: { [key: string]: chrome.storage.StorageChange }, areaName: string) => {
        if (areaName !== 'local') return;
        if (Object.keys(changes).includes(CONFIG_KEY)) {
            applyProxySettings().catch((e) =>
                console.error('[Argo Proxy] storage.onChanged -> applyProxySettings', e)
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
            const raw = await chrome.storage.local.get([CONFIG_KEY]);
            if (!raw[CONFIG_KEY]) {
                await chrome.storage.local.set({ [CONFIG_KEY]: cfg });
            }
            void applyProxySettings();
        })
        .catch((e) => console.error('[Argo Proxy] onInstalled', e));
});

void applyProxySettings();
