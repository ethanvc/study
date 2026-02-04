import { create } from 'zustand';
import type { Proxy, Rule } from '../shared/types';
import { STORAGE_KEYS, PROXY_ID_DIRECT, RULES_LIMIT } from '../shared/types';

function id(): string {
    return crypto.randomUUID?.() ?? 'id-' + Date.now() + '-' + Math.random().toString(36).slice(2, 9);
}

function getStorage(): Promise<{ proxies: Proxy[]; rules: Rule[] }> {
    return new Promise((resolve, reject) => {
        chrome.storage.local.get([STORAGE_KEYS.PROXIES, STORAGE_KEYS.RULES], (out) => {
            if (chrome.runtime.lastError) {
                console.error('[Argo Proxy] getStorage', chrome.runtime.lastError);
                reject(chrome.runtime.lastError);
            } else {
                resolve({
                    proxies: Array.isArray(out[STORAGE_KEYS.PROXIES]) ? out[STORAGE_KEYS.PROXIES] : [],
                    rules: Array.isArray(out[STORAGE_KEYS.RULES]) ? out[STORAGE_KEYS.RULES] : [],
                });
            }
        });
    });
}

function setStorage(partial: Partial<Record<string, Proxy[] | Rule[]>>): Promise<void> {
    return new Promise((resolve, reject) => {
        chrome.storage.local.set(partial, () => {
            if (chrome.runtime.lastError) {
                console.error('[Argo Proxy] setStorage', chrome.runtime.lastError);
                reject(chrome.runtime.lastError);
            } else resolve();
        });
    });
}

interface SettingsStore {
    proxies: Proxy[];
    rules: Rule[];
    load: () => Promise<void>;
    addProxy: (p: Omit<Proxy, 'id'>) => Promise<void>;
    updateProxy: (id: string, p: Partial<Proxy>) => Promise<void>;
    deleteProxy: (id: string) => Promise<void>;
    addRule: (r: Omit<Rule, 'id'>) => Promise<void>;
    updateRule: (id: string, r: Partial<Rule>) => Promise<void>;
    deleteRule: (id: string) => Promise<void>;
}

export const useSettingsStore = create<SettingsStore>((set, get) => ({
    proxies: [],
    rules: [],

    load: async () => {
        try {
            const data = await getStorage();
            set(data);
        } catch (e) {
            console.error('[Argo Proxy] settings load', e);
        }
    },

    addProxy: async (p) => {
        const { proxies } = get();
        const next: Proxy = { ...p, id: id() };
        await setStorage({ [STORAGE_KEYS.PROXIES]: [...proxies, next] });
        set({ proxies: [...proxies, next] });
    },

    updateProxy: async (proxyId, patch) => {
        const { proxies } = get();
        const idx = proxies.findIndex((x) => x.id === proxyId);
        if (idx === -1) return;
        const next = [...proxies];
        next[idx] = { ...next[idx], ...patch };
        await setStorage({ [STORAGE_KEYS.PROXIES]: next });
        set({ proxies: next });
    },

    deleteProxy: async (proxyId) => {
        const { proxies, rules } = get();
        const nextProxies = proxies.filter((p) => p.id !== proxyId);
        const nextRules = rules.map((r) => (r.proxyId === proxyId ? { ...r, proxyId: PROXY_ID_DIRECT } : r));
        await setStorage({ [STORAGE_KEYS.PROXIES]: nextProxies, [STORAGE_KEYS.RULES]: nextRules });
        set({ proxies: nextProxies, rules: nextRules });
    },

    addRule: async (r) => {
        const { rules } = get();
        if (rules.length >= RULES_LIMIT) {
            console.error('[Argo Proxy] saveRule', new Error('规则数量已达上限 ' + RULES_LIMIT + ' 条'));
            alert('规则数量已达上限 ' + RULES_LIMIT + ' 条');
            return;
        }
        const next: Rule = { ...r, id: id() };
        await setStorage({ [STORAGE_KEYS.RULES]: [...rules, next] });
        set({ rules: [...rules, next] });
    },

    updateRule: async (ruleId, patch) => {
        const { rules } = get();
        const idx = rules.findIndex((x) => x.id === ruleId);
        if (idx === -1) return;
        const next = [...rules];
        next[idx] = { ...next[idx], ...patch };
        await setStorage({ [STORAGE_KEYS.RULES]: next });
        set({ rules: next });
    },

    deleteRule: async (ruleId) => {
        const { rules } = get();
        const next = rules.filter((r) => r.id !== ruleId);
        await setStorage({ [STORAGE_KEYS.RULES]: next });
        set({ rules: next });
    },
}));
