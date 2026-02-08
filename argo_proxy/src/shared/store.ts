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
            const response = await chrome.runtime.sendMessage({ type: 'getConfig' });
            if (response) {
                set({
                    enabled: response.enabled ?? true,
                    proxies: response.proxies ?? [],
                    rules: response.rules ?? [],
                });
            }
        } catch (e) {
            console.error('[Argo Proxy] loadConfig', e);
        }
    },

    setEnabled: async (enabled: boolean) => {
        try {
            await chrome.runtime.sendMessage({ type: 'setEnabled', enabled });
            set({ enabled });
        } catch (e) {
            console.error('[Argo Proxy] setEnabled', e);
        }
    },

    setRuleProxy: async (ruleId: string, proxyId: string) => {
        try {
            await chrome.runtime.sendMessage({ type: 'setRuleProxy', ruleId, proxyId });
            const { rules } = get();
            const updatedRules = rules.map((r) =>
                r.id === ruleId ? { ...r, proxyId } : r
            );
            set({ rules: updatedRules });
        } catch (e) {
            console.error('[Argo Proxy] setRuleProxy', e);
        }
    },
}));
