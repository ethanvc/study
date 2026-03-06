import { useState } from 'react';
import { Plus, Pencil, Trash2, Server, Eye, EyeOff } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';
import { useProxyStore } from '@/shared/store';
import type { Proxy, ProxyType } from '@/shared/types';
import { proxyTypeOptions } from '@/shared/types';

export function ProxyList() {
    const { config, addProxy, updateProxy, deleteProxy } = useProxyStore();
    const proxies = config.proxies;

    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
    const [editingProxy, setEditingProxy] = useState<Proxy | null>(null);
    const [deletingProxyId, setDeletingProxyId] = useState<number | null>(null);
    const [visiblePasswords, setVisiblePasswords] = useState<Record<number, boolean>>({});
    const [formPasswordVisible, setFormPasswordVisible] = useState(false);

    const [formData, setFormData] = useState({
        name: '',
        type: 'http' as ProxyType,
        host: '',
        port: '',
        username: '',
        password: '',
    });

    const resetForm = () => {
        setFormData({ name: '', type: 'http', host: '', port: '', username: '', password: '' });
        setEditingProxy(null);
        setFormPasswordVisible(false);
    };

    const handleOpenDialog = (proxy?: Proxy) => {
        if (proxy) {
            setEditingProxy(proxy);
            setFormData({
                name: proxy.name,
                type: proxy.type,
                host: proxy.host,
                port: proxy.port.toString(),
                username: proxy.username ?? '',
                password: proxy.password ?? '',
            });
        } else {
            resetForm();
        }
        setIsDialogOpen(true);
    };

    const handleSubmit = async () => {
        const proxyData = {
            name: formData.name.trim(),
            type: formData.type,
            host: formData.host.trim(),
            port: parseInt(formData.port) || 0,
            username: formData.username || undefined,
            password: formData.password || undefined,
        };
        if (!proxyData.name || !proxyData.host || !proxyData.port) return;

        if (editingProxy) {
            await updateProxy(editingProxy.id, proxyData);
        } else {
            await addProxy(proxyData);
        }
        setIsDialogOpen(false);
        resetForm();
    };

    const handleDelete = (id: number) => {
        setDeletingProxyId(id);
        setIsDeleteDialogOpen(true);
    };

    const confirmDelete = async () => {
        if (deletingProxyId !== null) {
            await deleteProxy(deletingProxyId);
        }
        setIsDeleteDialogOpen(false);
        setDeletingProxyId(null);
    };

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-2xl font-semibold tracking-tight">代理列表</h2>
                    <p className="text-sm text-muted-foreground">管理你的代理服务器配置</p>
                </div>
                <Button onClick={() => handleOpenDialog()} className="gap-2">
                    <Plus className="h-4 w-4" />
                    新增代理
                </Button>
            </div>

            {proxies.length === 0 ? (
                <Card className="border-dashed">
                    <CardContent className="flex flex-col items-center justify-center py-12">
                        <Server className="h-12 w-12 text-muted-foreground/50" />
                        <p className="mt-4 text-sm text-muted-foreground">暂无代理配置</p>
                        <Button
                            variant="outline"
                            className="mt-4 gap-2"
                            onClick={() => handleOpenDialog()}
                        >
                            <Plus className="h-4 w-4" />
                            添加第一个代理
                        </Button>
                    </CardContent>
                </Card>
            ) : (
                <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                    {proxies.map((proxy) => (
                        <Card key={proxy.id} className="flex flex-col">
                            <CardHeader className="pb-3">
                                <div className="flex items-start justify-between">
                                    <div className="flex items-center gap-3">
                                        <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10">
                                            <Server className="h-5 w-5 text-primary" />
                                        </div>
                                        <div>
                                            <CardTitle className="text-base">{proxy.name}</CardTitle>
                                            <CardDescription className="text-xs">
                                                ID: {proxy.id}
                                            </CardDescription>
                                        </div>
                                    </div>
                                    <Badge variant="secondary" className="uppercase text-xs">
                                        {proxy.type}
                                    </Badge>
                                </div>
                            </CardHeader>
                            <CardContent className="flex flex-col flex-1">
                                <div className="space-y-2 text-sm flex-1">
                                    <div className="flex justify-between">
                                        <span className="text-muted-foreground">地址</span>
                                        <span className="font-mono">{proxy.host}</span>
                                    </div>
                                    <div className="flex justify-between">
                                        <span className="text-muted-foreground">端口</span>
                                        <span className="font-mono">{proxy.port}</span>
                                    </div>
                                    <div className="flex justify-between">
                                        <span className="text-muted-foreground">用户名</span>
                                        <span className="font-mono">{proxy.username ?? '-'}</span>
                                    </div>
                                    <div className="flex justify-between items-center">
                                        <span className="text-muted-foreground">密码</span>
                                        <div className="flex items-center gap-1">
                                            <span className="font-mono">
                                                {proxy.password
                                                    ? visiblePasswords[proxy.id]
                                                        ? proxy.password
                                                        : '••••••••'
                                                    : '-'}
                                            </span>
                                            {proxy.password && (
                                                <Button
                                                    variant="ghost"
                                                    size="icon"
                                                    className="h-6 w-6"
                                                    onClick={() =>
                                                        setVisiblePasswords((prev) => ({
                                                            ...prev,
                                                            [proxy.id]: !prev[proxy.id],
                                                        }))
                                                    }
                                                >
                                                    {visiblePasswords[proxy.id] ? (
                                                        <EyeOff className="h-3 w-3 text-muted-foreground" />
                                                    ) : (
                                                        <Eye className="h-3 w-3 text-muted-foreground" />
                                                    )}
                                                </Button>
                                            )}
                                        </div>
                                    </div>
                                </div>
                                <div className="mt-4 flex gap-2">
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        className="flex-1 gap-1"
                                        onClick={() => handleOpenDialog(proxy)}
                                    >
                                        <Pencil className="h-3 w-3" />
                                        编辑
                                    </Button>
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        className="flex-1 gap-1 text-destructive hover:text-destructive"
                                        onClick={() => handleDelete(proxy.id)}
                                    >
                                        <Trash2 className="h-3 w-3" />
                                        删除
                                    </Button>
                                </div>
                            </CardContent>
                        </Card>
                    ))}
                </div>
            )}

            {/* Add/Edit Dialog */}
            <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
                <DialogContent className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle>{editingProxy ? '编辑代理' : '新增代理'}</DialogTitle>
                        <DialogDescription>
                            {editingProxy ? '修改代理服务器配置' : '添加一个新的代理服务器'}
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                        <div className="space-y-2">
                            <Label htmlFor="proxy-name">名称</Label>
                            <Input
                                id="proxy-name"
                                placeholder="代理名称"
                                value={formData.name}
                                onChange={(e) =>
                                    setFormData({ ...formData, name: e.target.value })
                                }
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="proxy-type">类型</Label>
                            <Select
                                value={formData.type}
                                onValueChange={(v) =>
                                    setFormData({ ...formData, type: v as ProxyType })
                                }
                            >
                                <SelectTrigger id="proxy-type">
                                    <SelectValue placeholder="选择代理类型" />
                                </SelectTrigger>
                                <SelectContent>
                                    {proxyTypeOptions.map((opt) => (
                                        <SelectItem key={opt.value} value={opt.value}>
                                            {opt.label}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>
                        <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-2">
                                <Label htmlFor="proxy-host">主机地址</Label>
                                <Input
                                    id="proxy-host"
                                    placeholder="127.0.0.1"
                                    value={formData.host}
                                    onChange={(e) =>
                                        setFormData({ ...formData, host: e.target.value })
                                    }
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="proxy-port">端口</Label>
                                <Input
                                    id="proxy-port"
                                    type="number"
                                    placeholder="8080"
                                    value={formData.port}
                                    onChange={(e) =>
                                        setFormData({ ...formData, port: e.target.value })
                                    }
                                />
                            </div>
                        </div>
                        <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-2">
                                <Label htmlFor="proxy-username">用户名（可选）</Label>
                                <Input
                                    id="proxy-username"
                                    placeholder="用户名"
                                    value={formData.username}
                                    onChange={(e) =>
                                        setFormData({ ...formData, username: e.target.value })
                                    }
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="proxy-password">密码（可选）</Label>
                                <div className="relative">
                                    <Input
                                        id="proxy-password"
                                        type={formPasswordVisible ? 'text' : 'password'}
                                        placeholder="密码"
                                        value={formData.password}
                                        className="pr-10"
                                        onChange={(e) =>
                                            setFormData({ ...formData, password: e.target.value })
                                        }
                                    />
                                    <Button
                                        type="button"
                                        variant="ghost"
                                        size="icon"
                                        className="absolute right-0 top-0 h-full px-3 hover:bg-transparent"
                                        onClick={() =>
                                            setFormPasswordVisible(!formPasswordVisible)
                                        }
                                    >
                                        {formPasswordVisible ? (
                                            <EyeOff className="h-4 w-4 text-muted-foreground" />
                                        ) : (
                                            <Eye className="h-4 w-4 text-muted-foreground" />
                                        )}
                                    </Button>
                                </div>
                            </div>
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setIsDialogOpen(false)}>
                            取消
                        </Button>
                        <Button onClick={handleSubmit}>
                            {editingProxy ? '保存' : '创建'}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            {/* Delete Confirm */}
            <AlertDialog open={isDeleteDialogOpen} onOpenChange={setIsDeleteDialogOpen}>
                <AlertDialogContent>
                    <AlertDialogHeader>
                        <AlertDialogTitle>确认删除</AlertDialogTitle>
                        <AlertDialogDescription>
                            你确定要删除这个代理吗？此操作无法撤销，使用该代理的规则将失效。
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogCancel>取消</AlertDialogCancel>
                        <AlertDialogAction
                            onClick={confirmDelete}
                            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                        >
                            删除
                        </AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>
        </div>
    );
}
