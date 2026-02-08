export interface Proxy {
    id: string;
    name: string;
    type: string;
    host: string;
    port: number;
}

export interface Rule {
    id: string;
    matchType: 'domain' | 'pathPrefix';
    value: string;
    proxyId: string;
    order: number;
}

export const STORAGE_KEYS = {
    ENABLED: 'enabled',
    PROXIES: 'proxies',
    RULES: 'rules',
} as const;

export const PROXY_ID_DIRECT = 'direct';
export const RULES_LIMIT = 50;
