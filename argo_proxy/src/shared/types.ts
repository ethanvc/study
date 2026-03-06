// Proxy types
export type ProxyType = 'http' | 'https' | 'socks5';

export interface Proxy {
    id: number;
    name: string;
    type: ProxyType;
    host: string;
    port: number;
    username?: string;
    password?: string;
}

// Rule types
export type RuleType =
    | 'host_contains'
    | 'host_regex'
    | 'url_contains'
    | 'url_regex'
    | 'path_contains'
    | 'path_regex';

export type ActionType = 'proxy' | 'block' | 'direct';

export interface RuleAction {
    type: ActionType;
    proxyId?: number;
}

export interface Rule {
    id: number;
    type: RuleType;
    value: string;
    enabled: boolean;
    action: RuleAction;
}

// Profile type
export interface Profile {
    id: number;
    name: string;
    rules: Rule[];
    defaultAction: RuleAction;
}

// Option lists for UI selects
export const proxyTypeOptions: { value: ProxyType; label: string }[] = [
    { value: 'http', label: 'HTTP' },
    { value: 'https', label: 'HTTPS' },
    { value: 'socks5', label: 'SOCKS5' },
];

export const ruleTypeOptions: { value: RuleType; label: string }[] = [
    { value: 'host_contains', label: 'Host 包含' },
    { value: 'host_regex', label: 'Host 正则' },
    { value: 'url_contains', label: 'URL 包含' },
    { value: 'url_regex', label: 'URL 正则' },
    { value: 'path_contains', label: 'Path 包含' },
    { value: 'path_regex', label: 'Path 正则' },
];

export const actionTypeOptions: { value: ActionType; label: string }[] = [
    { value: 'proxy', label: '使用代理' },
    { value: 'block', label: '阻止' },
    { value: 'direct', label: '直连' },
];

// Storage
export const STORAGE_KEYS = {
    CONFIG: 'config',
} as const;

export const PROXY_ID_DIRECT = 'direct';

// Serialized config shape stored in chrome.storage.local
export interface StoredConfig {
    enabled: boolean;
    proxies: Proxy[];
    profiles: Profile[];
    activeProfileId: number | null;
    nextProxyId: number;
    nextProfileId: number;
    nextRuleId: number;
}

/**
 * Unified config object for the extension.
 * Stored as a single key in chrome.storage.local.
 */
export class Config {
    enabled: boolean;
    proxies: Proxy[];
    profiles: Profile[];
    activeProfileId: number | null;
    nextProxyId: number;
    nextProfileId: number;
    nextRuleId: number;

    constructor(params?: Partial<StoredConfig>) {
        this.enabled = params?.enabled ?? true;
        this.proxies = params?.proxies ?? [];
        this.profiles = params?.profiles ?? [];
        this.activeProfileId = params?.activeProfileId ?? null;
        this.nextProxyId = params?.nextProxyId ?? 1;
        this.nextProfileId = params?.nextProfileId ?? 1;
        this.nextRuleId = params?.nextRuleId ?? 1;
    }

    static fromStorage(out: { [key: string]: unknown }): Config {
        const raw = out[STORAGE_KEYS.CONFIG];
        if (raw && typeof raw === 'object' && !Array.isArray(raw)) {
            const data = raw as Partial<StoredConfig>;
            return new Config({
                enabled: data.enabled !== false,
                proxies: Array.isArray(data.proxies) ? data.proxies : [],
                profiles: Array.isArray(data.profiles) ? data.profiles : [],
                activeProfileId: data.activeProfileId ?? null,
                nextProxyId: data.nextProxyId ?? 1,
                nextProfileId: data.nextProfileId ?? 1,
                nextRuleId: data.nextRuleId ?? 1,
            });
        }
        return new Config();
    }

    toStorage(): Record<string, unknown> {
        const stored: StoredConfig = {
            enabled: this.enabled,
            proxies: this.proxies,
            profiles: this.profiles,
            activeProfileId: this.activeProfileId,
            nextProxyId: this.nextProxyId,
            nextProfileId: this.nextProfileId,
            nextRuleId: this.nextRuleId,
        };
        return { [STORAGE_KEYS.CONFIG]: stored };
    }
}
