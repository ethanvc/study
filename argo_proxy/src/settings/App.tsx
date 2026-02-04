import { useEffect, useState } from 'react';
import { useSettingsStore } from './settingsStore';
import type { Proxy, Rule } from '../shared/types';
import { PROXY_ID_DIRECT, RULES_LIMIT } from '../shared/types';

function proxyDisplayName(proxyId: string, proxies: Proxy[]): string {
    if (proxyId === PROXY_ID_DIRECT) return '直连';
    const p = proxies.find((x) => x.id === proxyId);
    return p ? p.name : proxyId;
}

export default function App() {
    const { proxies, rules, load, addProxy, updateProxy, deleteProxy, addRule, updateRule, deleteRule } =
        useSettingsStore();
    const [proxyForm, setProxyForm] = useState<Partial<Proxy> | null>(null);
    const [ruleForm, setRuleForm] = useState<Partial<Rule> & { id?: string } | null>(null);

    useEffect(() => {
        load();
    }, [load]);

    const handleSaveProxy = async () => {
        const { name, type, host, port } = proxyForm ?? {};
        if (!name?.trim() || !host?.trim() || !port || port < 1 || port > 65535) {
            console.error('[Argo Proxy] saveProxy', new Error('invalid input'));
            return;
        }
        if (proxyForm?.id) {
            await updateProxy(proxyForm.id, {
                name: name.trim(),
                type: (type as Proxy['type']) ?? 'http',
                host: host.trim(),
                port,
                username: proxyForm.username?.trim() || undefined,
                password: proxyForm.password?.trim() || undefined,
            });
        } else {
            await addProxy({
                name: name.trim(),
                type: (type as Proxy['type']) ?? 'http',
                host: host.trim(),
                port,
                username: proxyForm?.username?.trim() || undefined,
                password: proxyForm?.password?.trim() || undefined,
            });
        }
        setProxyForm(null);
    };

    const handleSaveRule = async () => {
        const { value, matchType, proxyId, order } = ruleForm ?? {};
        if (!value?.trim()) {
            console.error('[Argo Proxy] saveRule', new Error('匹配值不能为空'));
            return;
        }
        if (ruleForm?.id) {
            await updateRule(ruleForm.id, {
                value: value.trim(),
                matchType: (matchType as Rule['matchType']) ?? 'domain',
                proxyId: proxyId ?? PROXY_ID_DIRECT,
                order: order ?? 0,
            });
        } else {
            if (rules.length >= RULES_LIMIT) {
                alert('规则数量已达上限 ' + RULES_LIMIT + ' 条');
                return;
            }
            await addRule({
                value: value.trim(),
                matchType: (matchType as Rule['matchType']) ?? 'domain',
                proxyId: proxyId ?? PROXY_ID_DIRECT,
                order: order ?? 0,
            });
        }
        setRuleForm(null);
    };

    const sortedRules = [...rules].sort((a, b) => (a.order ?? 0) - (b.order ?? 0));

    return (
        <div className="p-5 max-w-[560px]">
            <h1 className="text-xl font-semibold mb-6">Argo Proxy 设置</h1>

            <h2 className="text-[15px] font-semibold text-gray-800 mt-5 mb-2">代理列表</h2>
            <div className="bg-white rounded-lg p-4 mb-4">
                {proxies.map((p) => (
                    <div
                        key={p.id}
                        className="flex items-center gap-3 py-2.5 px-3 bg-gray-50 rounded mb-2 text-sm last:mb-0"
                    >
                        <span className="flex-1 font-mono">{p.name}</span>
                        <span className="text-gray-500 text-xs">
                            {p.type} · {p.host}
                            {p.port ? ':' + p.port : ''}
                        </span>
                        <button
                            type="button"
                            onClick={() => setProxyForm({ ...p })}
                            className="px-2.5 py-1 text-xs bg-gray-200 rounded hover:bg-gray-300"
                        >
                            编辑
                        </button>
                        <button
                            type="button"
                            onClick={async () => {
                                if (confirm('确定删除该代理？使用该代理的规则将改为直连。')) await deleteProxy(p.id);
                            }}
                            className="px-2.5 py-1 text-xs bg-red-600 text-white rounded hover:bg-red-700"
                        >
                            删除
                        </button>
                    </div>
                ))}

                {proxyForm && (
                    <div className="mt-3 pt-3 border-t border-gray-200 space-y-2">
                        <input
                            className="w-full px-2.5 py-2 border rounded text-sm"
                            placeholder="名称"
                            value={proxyForm.name ?? ''}
                            onChange={(e) => setProxyForm((f) => ({ ...f, name: e.target.value }))}
                        />
                        <select
                            className="w-full px-2.5 py-2 border rounded text-sm"
                            value={proxyForm.type ?? 'http'}
                            onChange={(e) => setProxyForm((f) => ({ ...f, type: e.target.value as Proxy['type'] }))}
                        >
                            <option value="http">HTTP</option>
                            <option value="https">HTTPS</option>
                            <option value="socks">SOCKS</option>
                        </select>
                        <input
                            className="w-full px-2.5 py-2 border rounded text-sm"
                            placeholder="主机"
                            value={proxyForm.host ?? ''}
                            onChange={(e) => setProxyForm((f) => ({ ...f, host: e.target.value }))}
                        />
                        <input
                            type="number"
                            className="w-full px-2.5 py-2 border rounded text-sm"
                            placeholder="端口"
                            min={1}
                            max={65535}
                            value={proxyForm.port ?? ''}
                            onChange={(e) => setProxyForm((f) => ({ ...f, port: parseInt(e.target.value, 10) || 0 }))}
                        />
                        <input
                            className="w-full px-2.5 py-2 border rounded text-sm"
                            placeholder="用户名（可选）"
                            value={proxyForm.username ?? ''}
                            onChange={(e) => setProxyForm((f) => ({ ...f, username: e.target.value }))}
                        />
                        <input
                            type="password"
                            className="w-full px-2.5 py-2 border rounded text-sm"
                            placeholder="密码（可选）"
                            value={proxyForm.password ?? ''}
                            onChange={(e) => setProxyForm((f) => ({ ...f, password: e.target.value }))}
                        />
                        <div className="flex gap-2">
                            <button
                                type="button"
                                onClick={handleSaveProxy}
                                className="px-3 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700"
                            >
                                保存
                            </button>
                            <button
                                type="button"
                                onClick={() => setProxyForm(null)}
                                className="px-3 py-2 bg-gray-200 rounded text-sm hover:bg-gray-300"
                            >
                                取消
                            </button>
                        </div>
                    </div>
                )}
                <button
                    type="button"
                    onClick={() => setProxyForm({ name: '', type: 'http', host: '', port: 0 })}
                    className="mt-3 px-3 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700"
                >
                    + 添加代理
                </button>
            </div>

            <h2 className="text-[15px] font-semibold text-gray-800 mt-5 mb-2">路由规则</h2>
            <p className="text-gray-500 text-sm mb-2">匹配条件（域名或路径前缀）→ 选择代理，按顺序匹配。</p>
            <div className="bg-white rounded-lg p-4 mb-4">
                {sortedRules.map((r) => (
                    <div
                        key={r.id}
                        className="flex items-center gap-3 py-2.5 px-3 bg-gray-50 rounded mb-2 text-sm last:mb-0"
                    >
                        <span className="flex-1 font-mono">
                            {r.matchType === 'pathPrefix' ? r.value + ' (路径前缀)' : r.value}
                        </span>
                        <span className="text-gray-400">→</span>
                        <span className="text-blue-600">{proxyDisplayName(r.proxyId, proxies)}</span>
                        <button
                            type="button"
                            onClick={() => setRuleForm({ ...r })}
                            className="px-2.5 py-1 text-xs bg-gray-200 rounded hover:bg-gray-300"
                        >
                            编辑
                        </button>
                        <button
                            type="button"
                            onClick={() => deleteRule(r.id)}
                            className="px-2.5 py-1 text-xs bg-red-600 text-white rounded hover:bg-red-700"
                        >
                            删除
                        </button>
                    </div>
                ))}

                {ruleForm && (
                    <div className="mt-3 pt-3 border-t border-gray-200 space-y-2">
                        <select
                            className="w-full px-2.5 py-2 border rounded text-sm"
                            value={ruleForm.matchType ?? 'domain'}
                            onChange={(e) => setRuleForm((f) => ({ ...f, matchType: e.target.value as Rule['matchType'] }))}
                        >
                            <option value="domain">域名</option>
                            <option value="pathPrefix">路径前缀</option>
                        </select>
                        <input
                            className="w-full px-2.5 py-2 border rounded text-sm"
                            placeholder="如 api.example.com 或 /api/order"
                            value={ruleForm.value ?? ''}
                            onChange={(e) => setRuleForm((f) => ({ ...f, value: e.target.value }))}
                        />
                        <select
                            className="w-full px-2.5 py-2 border rounded text-sm"
                            value={ruleForm.proxyId ?? PROXY_ID_DIRECT}
                            onChange={(e) => setRuleForm((f) => ({ ...f, proxyId: e.target.value }))}
                        >
                            <option value={PROXY_ID_DIRECT}>直连</option>
                            {proxies.map((p) => (
                                <option key={p.id} value={p.id}>
                                    {p.name}
                                </option>
                            ))}
                        </select>
                        <input
                            type="number"
                            className="w-full px-2.5 py-2 border rounded text-sm"
                            placeholder="顺序"
                            min={0}
                            value={ruleForm.order ?? 0}
                            onChange={(e) => setRuleForm((f) => ({ ...f, order: parseInt(e.target.value, 10) || 0 }))}
                        />
                        <div className="flex gap-2">
                            <button
                                type="button"
                                onClick={handleSaveRule}
                                className="px-3 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700"
                            >
                                保存
                            </button>
                            <button
                                type="button"
                                onClick={() => setRuleForm(null)}
                                className="px-3 py-2 bg-gray-200 rounded text-sm hover:bg-gray-300"
                            >
                                取消
                            </button>
                        </div>
                    </div>
                )}
                <button
                    type="button"
                    onClick={() => setRuleForm({ value: '', matchType: 'domain', proxyId: PROXY_ID_DIRECT, order: rules.length })}
                    className="mt-3 px-3 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700"
                >
                    + 添加规则
                </button>
            </div>
        </div>
    );
}
