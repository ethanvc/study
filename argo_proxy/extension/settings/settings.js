(function () {
    const KEYS = (typeof window !== 'undefined' && window.STORAGE_KEYS) || { ENABLED: 'enabled', PROXIES: 'proxies', RULES: 'rules' };
    const PROXY_ID_DIRECT = (typeof window !== 'undefined' && window.PROXY_ID_DIRECT) || 'direct';
    const RULES_LIMIT = (typeof window !== 'undefined' && window.RULES_LIMIT) || 50;

    function logError(scope, err) {
        console.error('[Argo Proxy]', scope, err);
    }

    function id() {
        return crypto.randomUUID ? crypto.randomUUID() : 'id-' + Date.now() + '-' + Math.random().toString(36).slice(2, 9);
    }

    function getStorage() {
        return new Promise((resolve, reject) => {
            chrome.storage.local.get([KEYS.PROXIES, KEYS.RULES], (out) => {
                if (chrome.runtime.lastError) {
                    logError('getStorage', chrome.runtime.lastError);
                    reject(chrome.runtime.lastError);
                } else {
                    resolve({
                        proxies: Array.isArray(out[KEYS.PROXIES]) ? out[KEYS.PROXIES] : [],
                        rules: Array.isArray(out[KEYS.RULES]) ? out[KEYS.RULES] : [],
                    });
                }
            });
        });
    }

    function setStorage(partial) {
        return new Promise((resolve, reject) => {
            chrome.storage.local.set(partial, () => {
                if (chrome.runtime.lastError) {
                    logError('setStorage', chrome.runtime.lastError);
                    reject(chrome.runtime.lastError);
                } else {
                    resolve();
                }
            });
        });
    }

    const proxyListEl = document.getElementById('proxy-list');
    const proxyFormCard = document.getElementById('proxy-form-card');
    const proxyName = document.getElementById('proxy-name');
    const proxyType = document.getElementById('proxy-type');
    const proxyHost = document.getElementById('proxy-host');
    const proxyPort = document.getElementById('proxy-port');
    const proxyUsername = document.getElementById('proxy-username');
    const proxyPassword = document.getElementById('proxy-password');
    const proxySave = document.getElementById('proxy-save');
    const proxyCancel = document.getElementById('proxy-cancel');
    const proxyAdd = document.getElementById('proxy-add');

    const ruleListEl = document.getElementById('rule-list');
    const ruleFormCard = document.getElementById('rule-form-card');
    const ruleMatchType = document.getElementById('rule-match-type');
    const ruleValue = document.getElementById('rule-value');
    const ruleProxySelect = document.getElementById('rule-proxy');
    const ruleOrder = document.getElementById('rule-order');
    const ruleSave = document.getElementById('rule-save');
    const ruleCancel = document.getElementById('rule-cancel');
    const ruleAdd = document.getElementById('rule-add');

    let editingProxyId = null;
    let editingRuleId = null;
    let proxies = [];
    let rules = [];

    function renderProxies() {
        proxyListEl.innerHTML = '';
        proxies.forEach((p) => {
            const row = document.createElement('div');
            row.className = 'item-row';
            row.dataset.id = p.id;
            const desc = p.type + ' · ' + (p.host || '') + (p.port ? ':' + p.port : '');
            row.innerHTML =
                '<span class="flex host">' + escapeHtml(p.name || '') + '</span>' +
                '<span style="color:#666;">' + escapeHtml(desc) + '</span>' +
                '<div class="actions">' +
                '<button type="button" class="btn btn-ghost btn-edit-proxy" style="padding:4px 10px;font-size:12px;" data-id="' + escapeAttr(p.id) + '">编辑</button>' +
                '<button type="button" class="btn btn-danger btn-delete-proxy" style="padding:4px 10px;font-size:12px;" data-id="' + escapeAttr(p.id) + '">删除</button>' +
                '</div>';
            proxyListEl.appendChild(row);
        });
        proxyListEl.querySelectorAll('.btn-edit-proxy').forEach((btn) => {
            btn.addEventListener('click', () => startEditProxy(btn.dataset.id));
        });
        proxyListEl.querySelectorAll('.btn-delete-proxy').forEach((btn) => {
            btn.addEventListener('click', () => deleteProxy(btn.dataset.id));
        });
    }

    function escapeHtml(s) {
        const div = document.createElement('div');
        div.textContent = s;
        return div.innerHTML;
    }

    function escapeAttr(s) {
        return String(s).replace(/"/g, '&quot;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    }

    function startEditProxy(proxyId) {
        const p = proxies.find((x) => x.id === proxyId);
        if (!p) return;
        editingProxyId = proxyId;
        proxyName.value = p.name || '';
        proxyType.value = p.type || 'http';
        proxyHost.value = p.host || '';
        proxyPort.value = p.port != null ? p.port : '';
        proxyUsername.value = p.username || '';
        proxyPassword.value = p.password || '';
        proxyFormCard.classList.remove('hidden');
    }

    function cancelProxyForm() {
        editingProxyId = null;
        proxyName.value = '';
        proxyType.value = 'http';
        proxyHost.value = '';
        proxyPort.value = '';
        proxyUsername.value = '';
        proxyPassword.value = '';
        proxyFormCard.classList.add('hidden');
    }

    function saveProxy() {
        const name = proxyName.value.trim();
        const host = proxyHost.value.trim();
        const port = parseInt(proxyPort.value, 10);
        if (!name || !host || !port || port < 1 || port > 65535) {
            logError('saveProxy', new Error('invalid input: name/host/port required, port 1-65535'));
            return;
        }
        const payload = {
            name,
            type: proxyType.value || 'http',
            host,
            port,
            username: proxyUsername.value.trim() || undefined,
            password: proxyPassword.value.trim() || undefined,
        };
        if (editingProxyId) {
            const idx = proxies.findIndex((p) => p.id === editingProxyId);
            if (idx !== -1) {
                proxies = [...proxies];
                proxies[idx] = { ...proxies[idx], ...payload };
            }
        } else {
            proxies = [...proxies, { id: id(), ...payload }];
        }
        setStorage({ [KEYS.PROXIES]: proxies })
            .then(() => {
                cancelProxyForm();
                renderProxies();
            })
            .catch((e) => logError('saveProxy (setStorage)', e));
    }

    function deleteProxy(proxyId) {
        if (!confirm('确定删除该代理？使用该代理的规则将改为直连。')) return;
        proxies = proxies.filter((p) => p.id !== proxyId);
        rules = rules.map((r) => (r.proxyId === proxyId ? { ...r, proxyId: PROXY_ID_DIRECT } : r));
        setStorage({ [KEYS.PROXIES]: proxies, [KEYS.RULES]: rules })
            .then(() => {
                renderProxies();
                renderRules();
            })
            .catch((e) => logError('deleteProxy (setStorage)', e));
    }

    proxyAdd.addEventListener('click', () => {
        editingProxyId = null;
        proxyName.value = '';
        proxyType.value = 'http';
        proxyHost.value = '';
        proxyPort.value = '';
        proxyUsername.value = '';
        proxyPassword.value = '';
        proxyFormCard.classList.remove('hidden');
    });
    proxyCancel.addEventListener('click', cancelProxyForm);
    proxySave.addEventListener('click', saveProxy);

    function proxyNameToDisplay(proxyId) {
        if (proxyId === PROXY_ID_DIRECT) return '直连';
        const p = proxies.find((x) => x.id === proxyId);
        return p ? p.name : proxyId;
    }

    function fillRuleProxySelect(selectedId) {
        ruleProxySelect.innerHTML = '';
        const opt0 = document.createElement('option');
        opt0.value = PROXY_ID_DIRECT;
        opt0.textContent = '直连';
        if (selectedId === PROXY_ID_DIRECT) opt0.selected = true;
        ruleProxySelect.appendChild(opt0);
        proxies.forEach((p) => {
            const opt = document.createElement('option');
            opt.value = p.id;
            opt.textContent = p.name;
            if (p.id === selectedId) opt.selected = true;
            ruleProxySelect.appendChild(opt);
        });
    }

    function renderRules() {
        ruleListEl.innerHTML = '';
        const sorted = [...rules].sort((a, b) => (a.order ?? 0) - (b.order ?? 0));
        sorted.forEach((r) => {
            const row = document.createElement('div');
            row.className = 'item-row';
            row.dataset.id = r.id;
            const valueLabel = r.matchType === 'pathPrefix' ? r.value + ' (路径前缀)' : r.value;
            row.innerHTML =
                '<span class="flex host">' + escapeHtml(valueLabel) + '</span>' +
                '<span style="color:#999;">→</span>' +
                '<span class="proxy-name">' + escapeHtml(proxyNameToDisplay(r.proxyId)) + '</span>' +
                '<div class="actions">' +
                '<button type="button" class="btn btn-ghost btn-edit-rule" style="padding:4px 10px;font-size:12px;" data-id="' + escapeAttr(r.id) + '">编辑</button>' +
                '<button type="button" class="btn btn-danger btn-delete-rule" style="padding:4px 10px;font-size:12px;" data-id="' + escapeAttr(r.id) + '">删除</button>' +
                '</div>';
            ruleListEl.appendChild(row);
        });
        ruleListEl.querySelectorAll('.btn-edit-rule').forEach((btn) => {
            btn.addEventListener('click', () => startEditRule(btn.dataset.id));
        });
        ruleListEl.querySelectorAll('.btn-delete-rule').forEach((btn) => {
            btn.addEventListener('click', () => deleteRule(btn.dataset.id));
        });
    }

    function startEditRule(ruleId) {
        const r = rules.find((x) => x.id === ruleId);
        if (!r) return;
        editingRuleId = ruleId;
        ruleMatchType.value = r.matchType || 'domain';
        ruleValue.value = r.value || '';
        ruleOrder.value = r.order != null ? r.order : 0;
        fillRuleProxySelect(r.proxyId);
        ruleFormCard.style.display = 'block';
    }

    function cancelRuleForm() {
        editingRuleId = null;
        ruleMatchType.value = 'domain';
        ruleValue.value = '';
        ruleOrder.value = '0';
        ruleFormCard.style.display = 'none';
    }

    function saveRule() {
        const value = ruleValue.value.trim();
        if (!value) {
            logError('saveRule', new Error('匹配值不能为空'));
            return;
        }
        if (rules.length >= RULES_LIMIT && !editingRuleId) {
            logError('saveRule', new Error('规则数量已达上限 ' + RULES_LIMIT + ' 条'));
            alert('规则数量已达上限 ' + RULES_LIMIT + ' 条');
            return;
        }
        const payload = {
            matchType: ruleMatchType.value || 'domain',
            value,
            proxyId: ruleProxySelect.value || PROXY_ID_DIRECT,
            order: parseInt(ruleOrder.value, 10) || 0,
        };
        if (editingRuleId) {
            const idx = rules.findIndex((r) => r.id === editingRuleId);
            if (idx !== -1) {
                rules = [...rules];
                rules[idx] = { ...rules[idx], ...payload };
            }
        } else {
            rules = [...rules, { id: id(), ...payload }];
        }
        setStorage({ [KEYS.RULES]: rules })
            .then(() => {
                cancelRuleForm();
                renderRules();
            })
            .catch((e) => logError('saveRule (setStorage)', e));
    }

    function deleteRule(ruleId) {
        rules = rules.filter((r) => r.id !== ruleId);
        setStorage({ [KEYS.RULES]: rules })
            .then(renderRules)
            .catch((e) => logError('deleteRule (setStorage)', e));
    }

    ruleAdd.addEventListener('click', () => {
        editingRuleId = null;
        ruleMatchType.value = 'domain';
        ruleValue.value = '';
        ruleOrder.value = String(rules.length);
        fillRuleProxySelect(PROXY_ID_DIRECT);
        ruleFormCard.style.display = 'block';
    });
    ruleCancel.addEventListener('click', cancelRuleForm);
    ruleSave.addEventListener('click', saveRule);

    getStorage()
        .then((data) => {
            proxies = data.proxies;
            rules = data.rules;
            renderProxies();
            renderRules();
        })
        .catch((e) => logError('getStorage (init)', e));
})();
