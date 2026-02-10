export interface Proxy {
    name: string;
    protocol: string;
    host: string;
    port: number;
}

export const STORAGE_KEYS = {
    CONFIG: 'config',
} as const;

/**
 * 统一表示当前的配置对象（代理相关：开关 + 代理列表）。
 * 使用单个 storage key 存储整个配置对象。
 */
export class Config {
    enabled: boolean;
    proxies: Proxy[];

    constructor(params?: Partial<Config>) {
        this.enabled = params?.enabled ?? true;
        this.proxies = params?.proxies ?? [];
    }

    /** 从 chrome.storage.local.get 的结果中构造 Config */
    static fromStorage(out: { [key: string]: unknown }): Config {
        const configData = out[STORAGE_KEYS.CONFIG];
        if (configData && typeof configData === 'object' && !Array.isArray(configData)) {
            const data = configData as Partial<Config>;
            return new Config({
                enabled: data.enabled !== false,
                proxies: Array.isArray(data.proxies) ? data.proxies : [],
            });
        }
        return new Config();
    }

    /** 转回可直接传给 chrome.storage.local.set 的对象 */
    toStorage(): Record<string, unknown> {
        return {
            [STORAGE_KEYS.CONFIG]: {
                enabled: this.enabled,
                proxies: this.proxies,
            },
        };
    }
}

export const PROXY_ID_DIRECT = 'direct';
