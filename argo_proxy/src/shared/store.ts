import { create } from 'zustand';
import type { Config } from './types';

interface ConfigStore extends Config {
    setConfig: (c: Partial<Config>) => void;
    setEnabled: (enabled: boolean) => void;
    setRuleProxy: (ruleId: string, proxyId: string) => void;
    loadConfig: () => Promise<void>;
}

const sendMessage = (msg: object): Promise<Config> =>
    new Promise((resolve, reject) => {
        chrome.runtime.sendMessage(msg, (response: unknown) => {
            if (chrome.runtime.lastError) {
                console.error('[Argo Proxy] sendMessage', chrome.runtime.lastError);
                reject(chrome.runtime.lastError);
            } else {
                resolve(response as Config);
            }
        });
    });

export const useConfigStore = create<ConfigStore>((set, get) => ({
    enabled: true,
    proxies: [],
    rules: [],

    setConfig: (c) => set((state) => ({ ...state, ...c })),

    setEnabled: async (enabled) => {
        try {
            await sendMessage({ type: 'setEnabled', enabled });
            set({ enabled });
        } catch (e) {
            console.error('[Argo Proxy] setEnabled', e);
        }
    },

    setRuleProxy: async (ruleId, proxyId) => {
        try {
            await sendMessage({ type: 'setRuleProxy', ruleId, proxyId });
            const { rules } = get();
            set({
                rules: rules.map((r) => (r.id === ruleId ? { ...r, proxyId } : r)),
            });
        } catch (e) {
            console.error('[Argo Proxy] setRuleProxy', e);
        }
    },

    loadConfig: async () => {
        try {
            const config = await sendMessage({ type: 'getConfig' });
            set(config);
        } catch (e) {
            console.error('[Argo Proxy] loadConfig', e);
        }
    },
}));
