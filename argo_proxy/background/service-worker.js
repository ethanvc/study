/**
 * Argo Proxy - Service Worker (MV3)
 * Storage keys and limits (inlined from shared/storage-schema.js).
 */
const STORAGE_KEYS = { ENABLED: 'enabled', PROXIES: 'proxies', RULES: 'rules' };
const DEFAULT_ENABLED = true;
const PROXY_ID_DIRECT = 'direct';
const RULES_LIMIT = 50;

/** Placeholder: replace with real proxy gateway base URL. Redirect target will be gatewayBase + '?url=' + encodeURIComponent(originalUrl) + '&proxyId=' + proxyId */
const GATEWAY_BASE = 'https://argo-proxy-gateway.example.com/fetch';

/**
 * @returns {Promise<{enabled: boolean, proxies: Array, rules: Array}>}
 */
async function getConfig() {
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

/**
 * @param {boolean} enabled
 */
async function setEnabled(enabled) {
    await chrome.storage.local.set({ [STORAGE_KEYS.ENABLED]: enabled });
}

function escapeRegex(s) {
    return s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

function buildDnrRulesWithRegex(enabled, rules) {
    if (!enabled || !rules || rules.length === 0) {
        return [];
    }
    const sorted = [...rules]
        .filter((r) => r.proxyId && r.proxyId !== PROXY_ID_DIRECT)
        .sort((a, b) => (a.order ?? 0) - (b.order ?? 0))
        .slice(0, RULES_LIMIT);

    return sorted.map((rule, index) => {
        let regexFilter;
        if (rule.matchType === 'pathPrefix') {
            const path = escapeRegex(rule.value);
            regexFilter = '^https?:\\/\\/[^\\/]+' + path + '.*';
        } else {
            const domain = escapeRegex(rule.value.replace(/^\./, ''));
            regexFilter = '^https?:\\/\\/([^\\/]*\\.)?' + domain + '(\\/|$).*';
        }
        return {
            id: index + 1,
            priority: 1,
            action: {
                type: 'redirect',
                redirect: {
                    regexSubstitution: GATEWAY_BASE + '?url=\\0&proxyId=' + encodeURIComponent(rule.proxyId),
                },
            },
            condition: {
                regexFilter,
                resourceTypes: ['main_frame', 'sub_frame', 'xmlhttprequest', 'script', 'stylesheet', 'image', 'font', 'object', 'ping', 'csp_report', 'media', 'websocket', 'other'],
            },
        };
    });
}

async function applyDnrRules() {
    const { enabled, rules } = await getConfig();
    const dynamicRules = buildDnrRulesWithRegex(enabled, rules);
    await chrome.declarativeNetRequest.updateDynamicRules({
        removeRuleIds: (await chrome.declarativeNetRequest.getDynamicRules()).map((r) => r.id),
    });
    if (dynamicRules.length > 0) {
        await chrome.declarativeNetRequest.updateDynamicRules({ addRules: dynamicRules });
    }
}

chrome.storage.onChanged.addListener((changes, areaName) => {
    if (areaName !== 'local') return;
    const keySet = new Set(Object.keys(changes));
    if (keySet.has(STORAGE_KEYS.ENABLED) || keySet.has(STORAGE_KEYS.RULES) || keySet.has(STORAGE_KEYS.PROXIES)) {
        applyDnrRules();
    }
});

chrome.runtime.onMessage.addListener((msg, _sender, sendResponse) => {
    if (msg.type === 'getConfig') {
        getConfig().then(sendResponse);
        return true;
    }
    if (msg.type === 'setEnabled') {
        setEnabled(msg.enabled === true).then(() => sendResponse({ ok: true })).catch((e) => sendResponse({ ok: false, error: String(e) }));
        return true;
    }
    if (msg.type === 'setRuleProxy') {
        chrome.storage.local.get(STORAGE_KEYS.RULES).then((out) => {
            let rules = Array.isArray(out[STORAGE_KEYS.RULES]) ? out[STORAGE_KEYS.RULES] : [];
            const idx = rules.findIndex((r) => r.id === msg.ruleId);
            if (idx !== -1) {
                rules = [...rules];
                rules[idx] = { ...rules[idx], proxyId: msg.proxyId };
                chrome.storage.local.set({ [STORAGE_KEYS.RULES]: rules }).then(() => sendResponse({ ok: true })).catch((e) => sendResponse({ ok: false, error: String(e) }));
            } else {
                sendResponse({ ok: false, error: 'rule not found' });
            }
        });
        return true;
    }
    return false;
});

chrome.runtime.onInstalled.addListener(() => {
    chrome.storage.local.get([STORAGE_KEYS.ENABLED, STORAGE_KEYS.RULES, STORAGE_KEYS.PROXIES]).then((out) => {
        if (out[STORAGE_KEYS.ENABLED] === undefined) {
            chrome.storage.local.set({ [STORAGE_KEYS.ENABLED]: DEFAULT_ENABLED });
        }
        if (!Array.isArray(out[STORAGE_KEYS.PROXIES])) {
            chrome.storage.local.set({ [STORAGE_KEYS.PROXIES]: [] });
        }
        if (!Array.isArray(out[STORAGE_KEYS.RULES])) {
            chrome.storage.local.set({ [STORAGE_KEYS.RULES]: [] });
        }
        applyDnrRules();
    });
});

applyDnrRules();
