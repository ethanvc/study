(function () {
    const PROXY_ID_DIRECT = (typeof window !== 'undefined' && window.PROXY_ID_DIRECT) || 'direct';

    const viewMain = document.getElementById('view-main');
    const viewSub = document.getElementById('view-sub');
    const switchEl = document.getElementById('switch');
    const btnHostList = document.getElementById('btn-host-list');
    const hostCount = document.getElementById('host-count');
    const btnBack = document.getElementById('btn-back');
    const hostListContainer = document.getElementById('host-list-container');
    const hostListEmpty = document.getElementById('host-list-empty');
    const linkToSettings = document.getElementById('link-to-settings');
    const btnSettings = document.getElementById('btn-settings');

    let config = { enabled: true, proxies: [], rules: [] };

    function logError(scope, err) {
        console.error('[Argo Proxy]', scope, err);
    }

    function sendMessage(msg) {
        return new Promise((resolve, reject) => {
            chrome.runtime.sendMessage(msg, (response) => {
                if (chrome.runtime.lastError) {
                    logError('sendMessage', chrome.runtime.lastError);
                    reject(chrome.runtime.lastError);
                } else {
                    resolve(response);
                }
            });
        });
    }

    function loadConfig() {
        return sendMessage({ type: 'getConfig' }).then((c) => {
            config = c;
            return c;
        });
    }

    function renderMain() {
        switchEl.classList.toggle('off', !config.enabled);
        switchEl.setAttribute('aria-checked', config.enabled ? 'true' : 'false');
        const n = (config.rules && config.rules.length) || 0;
        hostCount.textContent = config.enabled ? n + ' 条' : '已暂停';
        btnHostList.classList.toggle('paused', !config.enabled);
    }

    /** Match URL host/path against rules (same logic as background). Returns matched rule or null. */
    function matchCurrentPage(url, rules) {
        if (!url || !rules || rules.length === 0) return null;
        let host = '';
        let pathname = '/';
        try {
            const u = new URL(url);
            host = u.hostname || '';
            pathname = u.pathname || '/';
        } catch (e) {
            logError('matchCurrentPage URL parse', e);
            return null;
        }
        const sorted = [...rules].sort((a, b) => (a.order ?? 0) - (b.order ?? 0));
        for (const rule of sorted) {
            if (rule.matchType === 'pathPrefix') {
                if (pathname.startsWith(rule.value)) return rule;
            } else {
                const domain = (rule.value || '').replace(/^\./, '');
                if (host === domain || host.endsWith('.' + domain)) return rule;
            }
        }
        return null;
    }

    function proxyIdToName(proxyId) {
        if (proxyId === PROXY_ID_DIRECT || !proxyId) return '直连';
        const p = (config.proxies || []).find((x) => x.id === proxyId);
        return p ? p.name : proxyId;
    }

    function renderCurrentPageBlock(hostLabel, proxyName) {
        const block = document.getElementById('current-page-block');
        if (!block) return;
        block.style.display = 'block';
        block.innerHTML = '<div class="label">当前页面</div><span class="value">' + escapeHtml(hostLabel) + '</span><span class="proxy-name">→ ' + escapeHtml(proxyName) + '</span>';
    }

    function escapeHtml(s) {
        const div = document.createElement('div');
        div.textContent = s;
        return div.innerHTML;
    }

    function renderHostList() {
        const rules = config.rules || [];
        const proxies = config.proxies || [];
        const currentPageBlock = document.getElementById('current-page-block');
        if (currentPageBlock) currentPageBlock.style.display = 'none';
        hostListContainer.innerHTML = '';

        chrome.tabs.query({ active: true, currentWindow: true }, (tabs) => {
            const tab = tabs && tabs[0];
            const url = tab && tab.url;
            if (url && !url.startsWith('chrome://') && !url.startsWith('edge://') && !url.startsWith('about:')) {
                const matched = matchCurrentPage(url, config.enabled ? rules : []);
                let hostLabel = '';
                try {
                    const u = new URL(url);
                    hostLabel = u.hostname + (u.pathname !== '/' ? u.pathname : '');
                } catch (e) {
                    logError('renderHostList URL parse', e);
                    hostLabel = url;
                }
                const proxyName = matched ? proxyIdToName(matched.proxyId) : '直连';
                renderCurrentPageBlock(hostLabel, proxyName);
            } else if (currentPageBlock) {
                currentPageBlock.style.display = 'block';
                currentPageBlock.innerHTML = '<div class="label">当前页面</div><span class="value">无法获取当前页面</span>';
            }
        });

        if (rules.length === 0) {
            hostListEmpty.style.display = 'block';
            return;
        }
        hostListEmpty.style.display = 'none';
        const sorted = [...rules].sort((a, b) => (a.order ?? 0) - (b.order ?? 0));

        sorted.forEach((rule) => {
            const row = document.createElement('div');
            row.className = 'host-item';
            row.dataset.ruleId = rule.id;
            const hostLabel = rule.matchType === 'pathPrefix' ? rule.value : rule.value;
            const select = document.createElement('select');
            select.className = 'proxy-select';
            select.setAttribute('aria-label', '选择代理');
            const optDirect = document.createElement('option');
            optDirect.value = PROXY_ID_DIRECT;
            optDirect.textContent = '直连';
            select.appendChild(optDirect);
            (proxies || []).forEach((p) => {
                const opt = document.createElement('option');
                opt.value = p.id;
                opt.textContent = p.name;
                if (rule.proxyId === p.id) opt.selected = true;
                select.appendChild(opt);
            });
            if (rule.proxyId === PROXY_ID_DIRECT || !rule.proxyId) optDirect.selected = true;

            select.addEventListener('change', function () {
                const proxyId = select.value;
                sendMessage({ type: 'setRuleProxy', ruleId: rule.id, proxyId })
                    .then(() => loadConfig().then(renderMain))
                    .catch((e) => logError('setRuleProxy', e));
            });

            row.appendChild(document.createElement('span')).className = 'host';
            row.querySelector('.host').textContent = hostLabel;
            const arrow = document.createElement('span');
            arrow.className = 'arrow-s';
            arrow.textContent = '→';
            row.appendChild(arrow);
            row.appendChild(select);
            hostListContainer.appendChild(row);
        });
    }

    function setEnabled(on) {
        return sendMessage({ type: 'setEnabled', enabled: on }).then(() => {
            config.enabled = on;
            renderMain();
        });
    }

    switchEl.addEventListener('click', function () {
        setEnabled(!config.enabled).catch((e) => logError('setEnabled (switch)', e));
    });
    switchEl.addEventListener('keydown', function (e) {
        if (e.key === ' ' || e.key === 'Enter') {
            e.preventDefault();
            setEnabled(!config.enabled).catch((e) => logError('setEnabled (keyboard)', e));
        }
    });

    btnHostList.addEventListener('click', function () {
        viewMain.classList.remove('active');
        viewSub.classList.add('active');
        renderHostList();
    });

    btnBack.addEventListener('click', function () {
        viewSub.classList.remove('active');
        viewMain.classList.add('active');
    });

    linkToSettings.addEventListener('click', function () {
        chrome.runtime.openOptionsPage();
    });

    btnSettings.addEventListener('click', function () {
        chrome.runtime.openOptionsPage();
    });

    loadConfig().then(renderMain).catch((e) => logError('loadConfig (init)', e));
})();
