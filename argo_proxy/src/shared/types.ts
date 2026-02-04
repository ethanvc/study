export const STORAGE_KEYS = {
    ENABLED: 'enabled',
    PROXIES: 'proxies',
    RULES: 'rules',
} as const;

export const PROXY_ID_DIRECT = 'direct';
export const RULES_LIMIT = 50;

export interface Proxy {
    id: string;
    name: string;
    type: 'http' | 'https' | 'socks';
    host: string;
    port: number;
    username?: string;
    password?: string;
}

export interface Rule {
    id: string;
    matchType: 'domain' | 'pathPrefix';
    value: string;
    proxyId: string;
    order: number;
}

export interface Config {
    enabled: boolean;
    proxies: Proxy[];
    rules: Rule[];
}
