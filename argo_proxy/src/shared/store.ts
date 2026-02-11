import { create } from 'zustand';
import { STORAGE_KEYS, Config } from './types';

interface ConfigStore {
    config: Config;
    loadConfig: () => Promise<void>;
    setConfig: (config: Config) => Promise<void>;
}

export const useConfigStore = create<ConfigStore>((set) => ({
    config: new Config(),

    loadConfig: async () => {
        try {
            // 直接从 storage 读取，避免 service worker 未启动的问题
            const out = await chrome.storage.local.get([STORAGE_KEYS.CONFIG]);
            const config = Config.fromStorage(out);
            set({ config });
        } catch (e) {
            console.error('[Argo Proxy] loadConfig', e);
        }
    },

    setConfig: async (config: Config) => {
        try {
            await chrome.storage.local.set(config.toStorage());
            set({ config });
        } catch (e) {
            console.error('[Argo Proxy] setConfig', e);
        }
    },
}));
