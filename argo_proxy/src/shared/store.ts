import { create } from 'zustand';
import type { Proxy } from './types';
import { STORAGE_KEYS, Config } from './types';

interface ConfigStore {
    enabled: boolean;
    proxies: Proxy[];
    loadConfig: () => Promise<void>;
    setEnabled: (enabled: boolean) => Promise<void>;
}

export const useConfigStore = create<ConfigStore>((set) => ({
    enabled: true,
    proxies: [],

    loadConfig: async () => {
        try {
            // 直接从 storage 读取，避免 service worker 未启动的问题
            const out = await chrome.storage.local.get([STORAGE_KEYS.CONFIG]);
            const config = Config.fromStorage(out);
            set({
                enabled: config.enabled,
                proxies: config.proxies,
            });
        } catch (e) {
            console.error('[Argo Proxy] loadConfig', e);
        }
    },

    setEnabled: async (enabled: boolean) => {
        try {
            // 读取当前配置，更新 enabled 后一起写回
            const out = await chrome.storage.local.get([STORAGE_KEYS.CONFIG]);
            const config = Config.fromStorage(out);
            config.enabled = enabled;
            await chrome.storage.local.set(config.toStorage());
            set({ enabled });
        } catch (e) {
            console.error('[Argo Proxy] setEnabled', e);
        }
    },
}));
