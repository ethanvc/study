import { create } from 'zustand';
import type { Proxy } from '../shared/types';
import { STORAGE_KEYS, Config } from '../shared/types';

async function getConfig(): Promise<Config> {
    const out = await chrome.storage.local.get([STORAGE_KEYS.CONFIG]);
    return Config.fromStorage(out);
}

async function saveConfig(config: Config): Promise<void> {
    await chrome.storage.local.set(config.toStorage());
}

interface SettingsStore {
    proxies: Proxy[];
    load: () => Promise<void>;
    addProxy: (p: Proxy) => Promise<void>;
    updateProxy: (name: string, p: Partial<Proxy>) => Promise<void>;
    deleteProxy: (name: string) => Promise<void>;
}

export const useSettingsStore = create<SettingsStore>((set, get) => ({
    proxies: [],

    load: async () => {
        try {
            const config = await getConfig();
            set({ proxies: config.proxies });
        } catch (e) {
            console.error('[Argo Proxy] settings load', e);
        }
    },

    addProxy: async (p) => {
        try {
            const config = await getConfig();
            config.proxies = [...config.proxies, p];
            await saveConfig(config);
            set({ proxies: config.proxies });
        } catch (e) {
            console.error('[Argo Proxy] addProxy', e);
        }
    },

    updateProxy: async (name, patch) => {
        try {
            const config = await getConfig();
            const idx = config.proxies.findIndex((x) => x.name === name);
            if (idx === -1) return;
            config.proxies = [...config.proxies];
            config.proxies[idx] = { ...config.proxies[idx], ...patch };
            await saveConfig(config);
            set({ proxies: config.proxies });
        } catch (e) {
            console.error('[Argo Proxy] updateProxy', e);
        }
    },

    deleteProxy: async (name) => {
        try {
            const config = await getConfig();
            config.proxies = config.proxies.filter((p) => p.name !== name);
            await saveConfig(config);
            set({ proxies: config.proxies });
        } catch (e) {
            console.error('[Argo Proxy] deleteProxy', e);
        }
    },
}));
