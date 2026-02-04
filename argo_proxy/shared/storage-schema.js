/**
 * Storage keys and default values for Argo Proxy.
 * Load via <script src=".../shared/storage-schema.js">; then use window.STORAGE_KEYS etc.
 * Background inlines same constants (no script bundle).
 */
(function (w) {
    w.STORAGE_KEYS = { ENABLED: 'enabled', PROXIES: 'proxies', RULES: 'rules' };
    w.DEFAULT_ENABLED = true;
    w.DEFAULT_PROXIES = [];
    w.DEFAULT_RULES = [];
    w.PROXY_ID_DIRECT = 'direct';
    w.RULES_LIMIT = 50;
})(typeof window !== 'undefined' ? window : typeof globalThis !== 'undefined' ? globalThis : self);
