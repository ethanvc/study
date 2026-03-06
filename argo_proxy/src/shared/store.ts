import { create } from 'zustand';
import { Config, STORAGE_KEYS } from './types';
import type { Proxy, ProxyType, Profile, Rule, RuleType, RuleAction, ActionType } from './types';

// Re-export types for consumers
export type { Proxy, ProxyType, Profile, Rule, RuleType, RuleAction, ActionType };

interface ProxyStore {
    config: Config;

    // Load from chrome.storage
    loadConfig: () => Promise<void>;

    // Toggle the global enabled flag (used by popup)
    setEnabled: (enabled: boolean) => Promise<void>;

    // Proxy CRUD
    addProxy: (proxy: Omit<Proxy, 'id'>) => Promise<void>;
    updateProxy: (id: number, patch: Partial<Omit<Proxy, 'id'>>) => Promise<void>;
    deleteProxy: (id: number) => Promise<void>;

    // Profile CRUD
    addProfile: (profile: Omit<Profile, 'id'>) => Promise<void>;
    updateProfile: (id: number, patch: Partial<Omit<Profile, 'id'>>) => Promise<void>;
    deleteProfile: (id: number) => Promise<void>;
    setActiveProfile: (id: number) => Promise<void>;

    // Rule CRUD
    addRule: (profileId: number, rule: Omit<Rule, 'id'>) => Promise<void>;
    updateRule: (profileId: number, ruleId: number, patch: Partial<Omit<Rule, 'id'>>) => Promise<void>;
    deleteRule: (profileId: number, ruleId: number) => Promise<void>;
    toggleRule: (profileId: number, ruleId: number) => Promise<void>;
    reorderRules: (profileId: number, rules: Rule[]) => Promise<void>;

    // Default action
    updateDefaultAction: (profileId: number, action: RuleAction) => Promise<void>;
}

async function persist(config: Config): Promise<void> {
    try {
        await chrome.storage.local.set(config.toStorage());
    } catch (e) {
        console.error('[Argo Proxy] persist', e);
    }
}

async function load(): Promise<Config> {
    try {
        const out = await chrome.storage.local.get([STORAGE_KEYS.CONFIG]);
        return Config.fromStorage(out);
    } catch (e) {
        console.error('[Argo Proxy] load', e);
        return new Config();
    }
}

export const useProxyStore = create<ProxyStore>((set, get) => ({
    config: new Config(),

    loadConfig: async () => {
        const config = await load();
        set({ config });
    },

    setEnabled: async (enabled) => {
        const prev = get().config;
        const next = new Config({ ...prev, enabled });
        set({ config: next });
        await persist(next);
    },

    // --- Proxy CRUD ---

    addProxy: async (proxy) => {
        const prev = get().config;
        const id = prev.nextProxyId;
        const next = new Config({
            ...prev,
            proxies: [...prev.proxies, { ...proxy, id }],
            nextProxyId: id + 1,
        });
        set({ config: next });
        await persist(next);
    },

    updateProxy: async (id, patch) => {
        const prev = get().config;
        const next = new Config({
            ...prev,
            proxies: prev.proxies.map((p) => (p.id === id ? { ...p, ...patch } : p)),
        });
        set({ config: next });
        await persist(next);
    },

    deleteProxy: async (id) => {
        const prev = get().config;
        const next = new Config({
            ...prev,
            proxies: prev.proxies.filter((p) => p.id !== id),
        });
        set({ config: next });
        await persist(next);
    },

    // --- Profile CRUD ---

    addProfile: async (profile) => {
        const prev = get().config;
        const id = prev.nextProfileId;
        const next = new Config({
            ...prev,
            profiles: [...prev.profiles, { ...profile, id }],
            nextProfileId: id + 1,
        });
        set({ config: next });
        await persist(next);
    },

    updateProfile: async (id, patch) => {
        const prev = get().config;
        const next = new Config({
            ...prev,
            profiles: prev.profiles.map((p) => (p.id === id ? { ...p, ...patch } : p)),
        });
        set({ config: next });
        await persist(next);
    },

    deleteProfile: async (id) => {
        const prev = get().config;
        const next = new Config({
            ...prev,
            profiles: prev.profiles.filter((p) => p.id !== id),
            activeProfileId: prev.activeProfileId === id ? null : prev.activeProfileId,
        });
        set({ config: next });
        await persist(next);
    },

    setActiveProfile: async (id) => {
        const prev = get().config;
        const next = new Config({ ...prev, activeProfileId: id });
        set({ config: next });
        await persist(next);
    },

    // --- Rule CRUD ---

    addRule: async (profileId, rule) => {
        const prev = get().config;
        const id = prev.nextRuleId;
        const next = new Config({
            ...prev,
            profiles: prev.profiles.map((p) =>
                p.id === profileId
                    ? { ...p, rules: [...p.rules, { ...rule, id }] }
                    : p
            ),
            nextRuleId: id + 1,
        });
        set({ config: next });
        await persist(next);
    },

    updateRule: async (profileId, ruleId, patch) => {
        const prev = get().config;
        const next = new Config({
            ...prev,
            profiles: prev.profiles.map((p) =>
                p.id === profileId
                    ? {
                          ...p,
                          rules: p.rules.map((r) =>
                              r.id === ruleId ? { ...r, ...patch } : r
                          ),
                      }
                    : p
            ),
        });
        set({ config: next });
        await persist(next);
    },

    deleteRule: async (profileId, ruleId) => {
        const prev = get().config;
        const next = new Config({
            ...prev,
            profiles: prev.profiles.map((p) =>
                p.id === profileId
                    ? { ...p, rules: p.rules.filter((r) => r.id !== ruleId) }
                    : p
            ),
        });
        set({ config: next });
        await persist(next);
    },

    toggleRule: async (profileId, ruleId) => {
        const prev = get().config;
        const next = new Config({
            ...prev,
            profiles: prev.profiles.map((p) =>
                p.id === profileId
                    ? {
                          ...p,
                          rules: p.rules.map((r) =>
                              r.id === ruleId ? { ...r, enabled: !r.enabled } : r
                          ),
                      }
                    : p
            ),
        });
        set({ config: next });
        await persist(next);
    },

    reorderRules: async (profileId, rules) => {
        const prev = get().config;
        const next = new Config({
            ...prev,
            profiles: prev.profiles.map((p) =>
                p.id === profileId ? { ...p, rules } : p
            ),
        });
        set({ config: next });
        await persist(next);
    },

    updateDefaultAction: async (profileId, action) => {
        const prev = get().config;
        const next = new Config({
            ...prev,
            profiles: prev.profiles.map((p) =>
                p.id === profileId ? { ...p, defaultAction: action } : p
            ),
        });
        set({ config: next });
        await persist(next);
    },
}));

// Backwards-compatible alias used by popup
export const useConfigStore = useProxyStore;
