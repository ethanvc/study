import { useState } from 'react';
import { Plus, Pencil, Trash2, FileText, Check } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
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
import {
    Tooltip,
    TooltipContent,
    TooltipProvider,
    TooltipTrigger,
} from '@/components/ui/tooltip';
import { useProxyStore } from '@/shared/store';
import { RuleEditor } from './RuleEditor';
import type { Profile, ActionType, RuleAction } from '@/shared/types';
import { actionTypeOptions } from '@/shared/types';
import { cn } from '@/lib/utils';

export function ProfileManager() {
    const {
        config,
        addProfile,
        updateProfile,
        deleteProfile,
        setActiveProfile,
    } = useProxyStore();
    const { profiles, proxies, activeProfileId } = config;

    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
    const [editingProfile, setEditingProfile] = useState<Profile | null>(null);
    const [deletingProfileId, setDeletingProfileId] = useState<number | null>(null);
    const [selectedProfileId, setSelectedProfileId] = useState<number | null>(
        activeProfileId
    );

    const [formData, setFormData] = useState({
        name: '',
        defaultActionType: 'direct' as ActionType,
        defaultProxyId: '',
    });

    const selectedProfile = profiles.find((p) => p.id === selectedProfileId);
    const canDeleteProfile = profiles.length > 1;

    const getActionLabel = (action: RuleAction): string => {
        if (action.type === 'proxy') {
            const proxy = proxies.find((p) => p.id === action.proxyId);
            return proxy ? `代理: ${proxy.name}` : '代理: 未知';
        }
        return action.type === 'block' ? '阻止' : '直连';
    };

    const resetForm = () => {
        setFormData({ name: '', defaultActionType: 'direct', defaultProxyId: '' });
        setEditingProfile(null);
    };

    const handleOpenDialog = (profile?: Profile) => {
        if (profile) {
            setEditingProfile(profile);
            setFormData({
                name: profile.name,
                defaultActionType: profile.defaultAction.type,
                defaultProxyId: profile.defaultAction.proxyId?.toString() ?? '',
            });
        } else {
            resetForm();
        }
        setIsDialogOpen(true);
    };

    const handleSubmit = async () => {
        if (!formData.name.trim()) return;
        const defaultAction: RuleAction = {
            type: formData.defaultActionType,
            proxyId:
                formData.defaultActionType === 'proxy' && formData.defaultProxyId
                    ? parseInt(formData.defaultProxyId)
                    : undefined,
        };
        if (editingProfile) {
            await updateProfile(editingProfile.id, {
                name: formData.name.trim(),
                defaultAction,
            });
        } else {
            await addProfile({ name: formData.name.trim(), rules: [], defaultAction });
        }
        setIsDialogOpen(false);
        resetForm();
    };

    const handleDelete = (id: number) => {
        if (!canDeleteProfile) return;
        setDeletingProfileId(id);
        setIsDeleteDialogOpen(true);
    };

    const confirmDelete = async () => {
        if (deletingProfileId !== null) {
            await deleteProfile(deletingProfileId);
            if (selectedProfileId === deletingProfileId) {
                const remaining = profiles.find((p) => p.id !== deletingProfileId);
                setSelectedProfileId(remaining?.id ?? null);
            }
        }
        setIsDeleteDialogOpen(false);
        setDeletingProfileId(null);
    };

    return (
        <TooltipProvider>
            <div className="space-y-6">
                <div className="flex items-center justify-between">
                    <div>
                        <h2 className="text-2xl font-semibold tracking-tight">Profile 配置</h2>
                        <p className="text-sm text-muted-foreground">管理规则配置文件</p>
                    </div>
                    <Button onClick={() => handleOpenDialog()} className="gap-2">
                        <Plus className="h-4 w-4" />
                        新增 Profile
                    </Button>
                </div>

                {profiles.length === 0 ? (
                    <Card className="border-dashed">
                        <CardContent className="flex flex-col items-center justify-center py-12">
                            <FileText className="h-12 w-12 text-muted-foreground/50" />
                            <p className="mt-4 text-sm text-muted-foreground">暂无 Profile</p>
                            <Button
                                variant="outline"
                                className="mt-4 gap-2"
                                onClick={() => handleOpenDialog()}
                            >
                                <Plus className="h-4 w-4" />
                                创建第一个 Profile
                            </Button>
                        </CardContent>
                    </Card>
                ) : (
                    <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-3">
                        {profiles.map((profile) => (
                            <Card
                                key={profile.id}
                                className={cn(
                                    'cursor-pointer transition-all hover:border-primary/50',
                                    selectedProfileId === profile.id &&
                                        'border-primary ring-1 ring-primary'
                                )}
                                onClick={() => setSelectedProfileId(profile.id)}
                            >
                                <CardHeader className="pb-2">
                                    <div className="flex items-start justify-between">
                                        <div className="flex items-center gap-2">
                                            <FileText className="h-4 w-4 text-muted-foreground" />
                                            <CardTitle className="text-sm font-medium">
                                                {profile.name}
                                            </CardTitle>
                                        </div>
                                        {activeProfileId === profile.id && (
                                            <Badge variant="default" className="gap-1">
                                                <Check className="h-3 w-3" />
                                                启用中
                                            </Badge>
                                        )}
                                    </div>
                                </CardHeader>
                                <CardContent className="pb-3">
                                    <div className="flex items-center justify-between text-xs text-muted-foreground">
                                        <span>{profile.rules.length} 条规则</span>
                                        <span>
                                            默认: {getActionLabel(profile.defaultAction)}
                                        </span>
                                    </div>
                                    <div
                                        className="mt-3 flex gap-2"
                                        onClick={(e) => e.stopPropagation()}
                                    >
                                        {activeProfileId !== profile.id && (
                                            <Button
                                                variant="outline"
                                                size="sm"
                                                className="flex-1 text-xs"
                                                onClick={() =>
                                                    void setActiveProfile(profile.id)
                                                }
                                            >
                                                启用
                                            </Button>
                                        )}
                                        <Button
                                            variant="ghost"
                                            size="icon"
                                            className="h-7 w-7"
                                            onClick={() => handleOpenDialog(profile)}
                                        >
                                            <Pencil className="h-3 w-3" />
                                        </Button>
                                        {canDeleteProfile ? (
                                            <Button
                                                variant="ghost"
                                                size="icon"
                                                className="h-7 w-7 text-destructive hover:text-destructive"
                                                onClick={() => handleDelete(profile.id)}
                                            >
                                                <Trash2 className="h-3 w-3" />
                                            </Button>
                                        ) : (
                                            <Tooltip>
                                                <TooltipTrigger asChild>
                                                    <span>
                                                        <Button
                                                            variant="ghost"
                                                            size="icon"
                                                            className="h-7 w-7 text-muted-foreground cursor-not-allowed"
                                                            disabled
                                                        >
                                                            <Trash2 className="h-3 w-3" />
                                                        </Button>
                                                    </span>
                                                </TooltipTrigger>
                                                <TooltipContent>
                                                    <p>至少需要保留一个 Profile</p>
                                                </TooltipContent>
                                            </Tooltip>
                                        )}
                                    </div>
                                </CardContent>
                            </Card>
                        ))}
                    </div>
                )}

                {/* Selected profile rule editor */}
                {selectedProfile && (
                    <div className="border-t pt-6">
                        <div className="mb-4 flex items-center justify-between">
                            <div className="flex items-center gap-2">
                                <h3 className="text-lg font-semibold">
                                    {selectedProfile.name}
                                </h3>
                                <Badge variant="outline">ID: {selectedProfile.id}</Badge>
                            </div>
                            <div className="flex items-center gap-2 text-sm text-muted-foreground">
                                <span>默认动作:</span>
                                <Badge variant="secondary">
                                    {getActionLabel(selectedProfile.defaultAction)}
                                </Badge>
                            </div>
                        </div>
                        <RuleEditor profile={selectedProfile} />
                    </div>
                )}

                {/* Add/Edit Profile Dialog */}
                <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
                    <DialogContent className="sm:max-w-md">
                        <DialogHeader>
                            <DialogTitle>
                                {editingProfile ? '编辑 Profile' : '新增 Profile'}
                            </DialogTitle>
                            <DialogDescription>
                                {editingProfile
                                    ? '修改 Profile 配置'
                                    : '创建一个新的规则配置文件'}
                            </DialogDescription>
                        </DialogHeader>
                        <div className="space-y-4 py-4">
                            <div className="space-y-2">
                                <Label htmlFor="profile-name">名称</Label>
                                <Input
                                    id="profile-name"
                                    placeholder="Profile 名称"
                                    value={formData.name}
                                    onChange={(e) =>
                                        setFormData({ ...formData, name: e.target.value })
                                    }
                                />
                            </div>
                            <div className="space-y-2">
                                <Label>默认动作</Label>
                                <p className="text-xs text-muted-foreground">
                                    当请求不匹配任何规则时执行的动作
                                </p>
                                <Select
                                    value={formData.defaultActionType}
                                    onValueChange={(v) =>
                                        setFormData({
                                            ...formData,
                                            defaultActionType: v as ActionType,
                                        })
                                    }
                                >
                                    <SelectTrigger>
                                        <SelectValue placeholder="选择默认动作" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {actionTypeOptions.map((opt) => (
                                            <SelectItem key={opt.value} value={opt.value}>
                                                {opt.label}
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                            </div>
                            {formData.defaultActionType === 'proxy' && (
                                <div className="space-y-2">
                                    <Label>选择代理</Label>
                                    <Select
                                        value={formData.defaultProxyId}
                                        onValueChange={(v) =>
                                            setFormData({ ...formData, defaultProxyId: v })
                                        }
                                    >
                                        <SelectTrigger>
                                            <SelectValue placeholder="选择代理服务器" />
                                        </SelectTrigger>
                                        <SelectContent>
                                            {proxies.map((proxy) => (
                                                <SelectItem
                                                    key={proxy.id}
                                                    value={proxy.id.toString()}
                                                >
                                                    {proxy.name} ({proxy.type.toUpperCase()})
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                </div>
                            )}
                        </div>
                        <DialogFooter>
                            <Button
                                variant="outline"
                                onClick={() => setIsDialogOpen(false)}
                            >
                                取消
                            </Button>
                            <Button
                                onClick={handleSubmit}
                                disabled={
                                    !formData.name ||
                                    (formData.defaultActionType === 'proxy' &&
                                        !formData.defaultProxyId)
                                }
                            >
                                {editingProfile ? '保存' : '创建'}
                            </Button>
                        </DialogFooter>
                    </DialogContent>
                </Dialog>

                {/* Delete Confirm */}
                <AlertDialog
                    open={isDeleteDialogOpen}
                    onOpenChange={setIsDeleteDialogOpen}
                >
                    <AlertDialogContent>
                        <AlertDialogHeader>
                            <AlertDialogTitle>确认删除</AlertDialogTitle>
                            <AlertDialogDescription>
                                你确定要删除这个 Profile 吗？此操作无法撤销，所有关联的规则也会被删除。
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
        </TooltipProvider>
    );
}
