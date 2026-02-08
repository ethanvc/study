import { create } from 'zustand';
import type { Proxy, Rule } from './types';
import { STORAGE_KEYS } from './types';

interface ConfigStore {
    enabled: boolean;
    proxies: Proxy[];
    rules: Rule[];
    loadConfig: () => Promise<void>;
    setEnabled: (enabled: boolean) => Promise<void>;
    setRuleProxy: (ruleId: string, proxyId: string) => Promise<void>;
}

export const useConfigStore = create<ConfigStore>((set, get) => ({
    enabled: true,
    proxies: [],
    rules: [],

    loadConfig: async () => {
        try {
            // 直接从 storage 读取，避免 service worker 未启动的问题
            const out = await chrome.storage.local.get([
                STORAGE_KEYS.ENABLED,
                STORAGE_KEYS.PROXIES,
                STORAGE_KEYS.RULES,
            ]);
            set({
                enabled: out[STORAGE_KEYS.ENABLED] !== false,
                proxies: Array.isArray(out[STORAGE_KEYS.PROXIES]) ? out[STORAGE_KEYS.PROXIES] : [],
                rules: Array.isArray(out[STORAGE_KEYS.RULES]) ? out[STORAGE_KEYS.RULES] : [],
            });
        } catch (e) {
            console.error('[Argo Proxy] loadConfig', e);
        }
    },

    setEnabled: async (enabled: boolean) => {
        try {
            await chrome.storage.local.set({ [STORAGE_KEYS.ENABLED]: enabled });
            set({ enabled });
        } catch (e) {
            console.error('[Argo Proxy] setEnabled', e);
        }
    },

    setRuleProxy: async (ruleId: string, proxyId: string) => {
        try {
            const { rules } = get();
            const updatedRules = rules.map((r) =>
                r.id === ruleId ? { ...r, proxyId } : r
            );
            await chrome.storage.local.set({ [STORAGE_KEYS.RULES]: updatedRules });
            set({ rules: updatedRules });
        } catch (e) {
            console.error('[Argo Proxy] setRuleProxy', e);
        }
    },
}));
