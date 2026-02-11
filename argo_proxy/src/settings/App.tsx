import { useEffect, useState } from 'react';
import { Config } from '../shared/types';
import { useConfigStore } from '../shared/store';
import type { Proxy } from '../shared/types';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from '@/components/ui/card';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import { Plus, Pencil, Trash2, Server } from 'lucide-react';
import { cn } from '@/lib/utils';

const SIDEBAR_ITEMS = [
    { id: 'proxy', label: 'Proxy', icon: Server },
] as const;

export default function App() {
    const { config, loadConfig, setConfig } = useConfigStore();
    const proxies = config.proxies;
    const [activeSection, setActiveSection] = useState<(typeof SIDEBAR_ITEMS)[number]['id']>('proxy');
    const [proxyForm, setProxyForm] = useState<Partial<Proxy> & { username?: string; password?: string } | null>(null);
    const [editingProxyName, setEditingProxyName] = useState<string | null>(null);

    useEffect(() => {
        loadConfig();
    }, [loadConfig]);

    const handleSaveProxy = async () => {
        const { name, protocol, host, port } = proxyForm ?? {};
        if (!name?.trim() || !host?.trim() || !port || port < 1 || port > 65535) {
            console.error('[Argo Proxy] saveProxy', new Error('invalid input'));
            return;
        }
        const payload = {
            name: name.trim(),
            protocol: (protocol as Proxy['protocol']) ?? 'http',
            host: host.trim(),
            port,
        };
        if (editingProxyName !== null) {
            const idx = config.proxies.findIndex((x) => x.name === editingProxyName);
            if (idx === -1) return;
            const next = [...config.proxies];
            next[idx] = { ...next[idx], ...payload };
            await setConfig(new Config({ ...config, proxies: next }));
        } else {
            await setConfig(new Config({ ...config, proxies: [...config.proxies, payload as Proxy] }));
        }
        setProxyForm(null);
        setEditingProxyName(null);
    };

    return (
        <div className="flex h-screen">
            {/* 左侧边栏 */}
            <aside className="w-52 border-r bg-muted/30 flex flex-col">
                <div className="p-4">
                    <h1 className="text-lg font-semibold">Argo Proxy</h1>
                    <p className="text-xs text-muted-foreground mt-1">设置</p>
                </div>
                <Separator />
                <nav className="flex-1 p-2 space-y-1">
                    {SIDEBAR_ITEMS.map((item) => {
                        const Icon = item.icon;
                        return (
                            <button
                                key={item.id}
                                type="button"
                                onClick={() => setActiveSection(item.id)}
                                className={cn(
                                    'w-full flex items-center gap-2 px-3 py-2 rounded-md text-sm font-medium transition-colors',
                                    activeSection === item.id
                                        ? 'bg-primary text-primary-foreground'
                                        : 'hover:bg-accent hover:text-accent-foreground'
                                )}
                            >
                                <Icon className="h-4 w-4" />
                                {item.label}
                            </button>
                        );
                    })}
                </nav>
            </aside>

            {/* 右侧内容区 */}
            <main className="flex-1 overflow-auto p-6">
                {activeSection === 'proxy' && (
                    <div className="max-w-2xl space-y-6">
                        <div>
                            <h2 className="text-xl font-semibold">代理列表</h2>
                            <p className="text-sm text-muted-foreground mt-1">
                                管理 HTTP/HTTPS/SOCKS 代理
                            </p>
                        </div>

                        <Card>
                            <CardHeader className="pb-4">
                                <CardTitle>已配置的代理</CardTitle>
                                <CardDescription>
                                    添加、编辑或删除代理服务器
                                </CardDescription>
                            </CardHeader>
                            <CardContent className="space-y-4">
                                <div className="space-y-2">
                                    {proxies.map((p) => (
                                        <div
                                            key={p.name}
                                            className="flex items-center gap-3 py-3 px-4 rounded-lg border bg-card hover:bg-accent/50 transition-colors"
                                        >
                                            <div className="flex-1 min-w-0">
                                                <span className="font-mono font-medium">{p.name}</span>
                                                <span className="text-xs text-muted-foreground ml-2">
                                                    {p.protocol} · {p.host}:{p.port}
                                                </span>
                                            </div>
                                            <div className="flex gap-2">
                                                <Button
                                                    variant="outline"
                                                    size="sm"
                                                    onClick={() => {
                                                        setEditingProxyName(p.name);
                                                        setProxyForm({ ...p });
                                                    }}
                                                >
                                                    <Pencil className="h-3.5 w-3.5" />
                                                    编辑
                                                </Button>
                                                <Button
                                                    variant="destructive"
                                                    size="sm"
                                                    onClick={async () => {
                                                        if (confirm('确定删除该代理？')) {
                                                            await setConfig(
                                                                new Config({
                                                                    ...config,
                                                                    proxies: config.proxies.filter(
                                                                        (x) => x.name !== p.name
                                                                    ),
                                                                })
                                                            );
                                                        }
                                                    }}
                                                >
                                                    <Trash2 className="h-3.5 w-3.5" />
                                                    删除
                                                </Button>
                                            </div>
                                        </div>
                                    ))}
                                    {proxies.length === 0 && (
                                        <p className="text-sm text-muted-foreground py-6 text-center">
                                            暂无代理，点击下方添加
                                        </p>
                                    )}
                                </div>

                                <Separator />

                                {proxyForm ? (
                                    <Card>
                                        <CardHeader className="pb-3">
                                            <CardTitle className="text-base">
                                                {editingProxyName ? '编辑代理' : '添加代理'}
                                            </CardTitle>
                                        </CardHeader>
                                        <CardContent className="space-y-4">
                                            <div className="space-y-2">
                                                <Label htmlFor="name">名称</Label>
                                                <Input
                                                    id="name"
                                                    placeholder="例如: 公司代理"
                                                    value={proxyForm.name ?? ''}
                                                    onChange={(e) =>
                                                        setProxyForm((f) => ({ ...f, name: e.target.value }))
                                                    }
                                                />
                                            </div>
                                            <div className="space-y-2">
                                                <Label htmlFor="protocol">协议</Label>
                                                <Select
                                                    value={proxyForm.protocol ?? 'http'}
                                                    onValueChange={(v: string) =>
                                                        setProxyForm((f) => ({
                                                            ...f,
                                                            protocol: v as Proxy['protocol'],
                                                        }))
                                                    }
                                                >
                                                    <SelectTrigger id="protocol">
                                                        <SelectValue placeholder="选择协议" />
                                                    </SelectTrigger>
                                                    <SelectContent>
                                                        <SelectItem value="http">HTTP</SelectItem>
                                                        <SelectItem value="https">HTTPS</SelectItem>
                                                        <SelectItem value="socks">SOCKS</SelectItem>
                                                    </SelectContent>
                                                </Select>
                                            </div>
                                            <div className="space-y-2">
                                                <Label htmlFor="host">主机</Label>
                                                <Input
                                                    id="host"
                                                    placeholder="127.0.0.1 或 proxy.example.com"
                                                    value={proxyForm.host ?? ''}
                                                    onChange={(e) =>
                                                        setProxyForm((f) => ({ ...f, host: e.target.value }))
                                                    }
                                                />
                                            </div>
                                            <div className="space-y-2">
                                                <Label htmlFor="port">端口</Label>
                                                <Input
                                                    id="port"
                                                    type="number"
                                                    placeholder="8080"
                                                    min={1}
                                                    max={65535}
                                                    value={proxyForm.port ?? ''}
                                                    onChange={(e) =>
                                                        setProxyForm((f) => ({
                                                            ...f,
                                                            port: parseInt(e.target.value, 10) || 0,
                                                        }))
                                                    }
                                                />
                                            </div>
                                            <div className="flex gap-2 pt-2">
                                                <Button onClick={handleSaveProxy}>保存</Button>
                                                <Button
                                                    variant="outline"
                                                    onClick={() => {
                                                        setProxyForm(null);
                                                        setEditingProxyName(null);
                                                    }}
                                                >
                                                    取消
                                                </Button>
                                            </div>
                                        </CardContent>
                                    </Card>
                                ) : (
                                    <Button
                                        variant="outline"
                                        className="w-full"
                                        onClick={() => {
                                            setEditingProxyName(null);
                                            setProxyForm({
                                                name: '',
                                                protocol: 'http',
                                                host: '',
                                                port: 0,
                                            });
                                        }}
                                    >
                                        <Plus className="h-4 w-4" />
                                        添加代理
                                    </Button>
                                )}
                            </CardContent>
                        </Card>
                    </div>
                )}
            </main>
        </div>
    );
}
