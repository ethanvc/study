import { useState } from 'react';
import {
    Plus,
    Pencil,
    Trash2,
    Filter,
    ArrowRight,
    GripVertical,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Switch } from '@/components/ui/switch';
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
    DndContext,
    closestCenter,
    KeyboardSensor,
    PointerSensor,
    useSensor,
    useSensors,
    type DragEndEvent,
} from '@dnd-kit/core';
import {
    arrayMove,
    SortableContext,
    sortableKeyboardCoordinates,
    useSortable,
    verticalListSortingStrategy,
} from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { useProxyStore } from '@/shared/store';
import type { Rule, RuleType, ActionType, RuleAction, Profile } from '@/shared/types';
import { ruleTypeOptions, actionTypeOptions } from '@/shared/types';
import { cn } from '@/lib/utils';

interface RuleEditorProps {
    profile: Profile;
}

interface SortableRuleItemProps {
    rule: Rule;
    index: number;
    profileId: number;
    onEdit: (rule: Rule) => void;
    onDelete: (id: number) => void;
    getActionLabel: (action: RuleAction) => string;
    getActionBadgeVariant: (type: ActionType) => 'default' | 'destructive' | 'secondary';
    getRuleTypeLabel: (type: RuleType) => string;
}

function SortableRuleItem({
    rule,
    index,
    profileId,
    onEdit,
    onDelete,
    getActionLabel,
    getActionBadgeVariant,
    getRuleTypeLabel,
}: SortableRuleItemProps) {
    const { toggleRule } = useProxyStore();
    const { attributes, listeners, setNodeRef, transform, transition, isDragging } =
        useSortable({ id: rule.id });

    const style = {
        transform: CSS.Transform.toString(transform),
        transition,
        zIndex: isDragging ? 1 : 0,
    };

    return (
        <div
            ref={setNodeRef}
            style={style}
            className={cn(
                'flex items-center gap-2 rounded-md border bg-card p-2',
                isDragging && 'opacity-50 shadow-lg',
                !rule.enabled && 'opacity-50'
            )}
        >
            <button
                className="cursor-grab touch-none p-1 text-muted-foreground hover:text-foreground"
                {...attributes}
                {...listeners}
            >
                <GripVertical className="h-4 w-4" />
            </button>

            <div className="flex h-6 w-6 shrink-0 items-center justify-center rounded bg-muted text-xs font-medium">
                {index + 1}
            </div>

            <Switch
                checked={rule.enabled}
                onCheckedChange={() => void toggleRule(profileId, rule.id)}
                className="shrink-0 scale-75"
            />

            <div className="flex flex-1 items-center gap-2 min-w-0">
                <Badge variant="outline" className="shrink-0 text-xs py-0 h-5">
                    {getRuleTypeLabel(rule.type)}
                </Badge>
                <code className="flex-1 truncate rounded bg-muted px-1.5 py-0.5 text-xs font-mono">
                    {rule.value}
                </code>
                <ArrowRight className="h-3 w-3 text-muted-foreground shrink-0" />
                <Badge
                    variant={getActionBadgeVariant(rule.action.type)}
                    className="shrink-0 text-xs py-0 h-5"
                >
                    {getActionLabel(rule.action)}
                </Badge>
            </div>

            <div className="flex gap-0.5 shrink-0">
                <Button
                    variant="ghost"
                    size="icon"
                    className="h-6 w-6"
                    onClick={() => onEdit(rule)}
                >
                    <Pencil className="h-3 w-3" />
                    <span className="sr-only">编辑</span>
                </Button>
                <Button
                    variant="ghost"
                    size="icon"
                    className="h-6 w-6 text-destructive hover:text-destructive"
                    onClick={() => onDelete(rule.id)}
                >
                    <Trash2 className="h-3 w-3" />
                    <span className="sr-only">删除</span>
                </Button>
            </div>
        </div>
    );
}

export function RuleEditor({ profile }: RuleEditorProps) {
    const { config, addRule, updateRule, deleteRule, reorderRules } = useProxyStore();
    const proxies = config.proxies;

    const [isRuleDialogOpen, setIsRuleDialogOpen] = useState(false);
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
    const [editingRule, setEditingRule] = useState<Rule | null>(null);
    const [deletingRuleId, setDeletingRuleId] = useState<number | null>(null);

    const [ruleFormData, setRuleFormData] = useState({
        type: 'host_contains' as RuleType,
        value: '',
        enabled: true,
        actionType: 'direct' as ActionType,
        proxyId: '',
    });

    const sensors = useSensors(
        useSensor(PointerSensor),
        useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates })
    );

    const resetRuleForm = () => {
        setRuleFormData({
            type: 'host_contains',
            value: '',
            enabled: true,
            actionType: 'direct',
            proxyId: '',
        });
        setEditingRule(null);
    };

    const handleOpenRuleDialog = (rule?: Rule) => {
        if (rule) {
            setEditingRule(rule);
            setRuleFormData({
                type: rule.type,
                value: rule.value,
                enabled: rule.enabled,
                actionType: rule.action.type,
                proxyId: rule.action.proxyId?.toString() ?? '',
            });
        } else {
            resetRuleForm();
        }
        setIsRuleDialogOpen(true);
    };

    const handleSubmitRule = async () => {
        const action: RuleAction = {
            type: ruleFormData.actionType,
            proxyId:
                ruleFormData.actionType === 'proxy' && ruleFormData.proxyId
                    ? parseInt(ruleFormData.proxyId)
                    : undefined,
        };
        const ruleData = {
            type: ruleFormData.type,
            value: ruleFormData.value,
            enabled: ruleFormData.enabled,
            action,
        };
        if (editingRule) {
            await updateRule(profile.id, editingRule.id, ruleData);
        } else {
            await addRule(profile.id, ruleData);
        }
        setIsRuleDialogOpen(false);
        resetRuleForm();
    };

    const handleDeleteRule = (id: number) => {
        setDeletingRuleId(id);
        setIsDeleteDialogOpen(true);
    };

    const confirmDeleteRule = async () => {
        if (deletingRuleId !== null) {
            await deleteRule(profile.id, deletingRuleId);
        }
        setIsDeleteDialogOpen(false);
        setDeletingRuleId(null);
    };

    const handleDragEnd = (event: DragEndEvent) => {
        const { active, over } = event;
        if (over && active.id !== over.id) {
            const oldIndex = profile.rules.findIndex((r) => r.id === active.id);
            const newIndex = profile.rules.findIndex((r) => r.id === over.id);
            const newRules = arrayMove(profile.rules, oldIndex, newIndex);
            void reorderRules(profile.id, newRules);
        }
    };

    const getActionLabel = (action: RuleAction): string => {
        if (action.type === 'proxy') {
            const proxy = proxies.find((p) => p.id === action.proxyId);
            return proxy ? `代理: ${proxy.name}` : '代理: 未知';
        }
        return action.type === 'block' ? '阻止' : '直连';
    };

    const getActionBadgeVariant = (
        type: ActionType
    ): 'default' | 'destructive' | 'secondary' => {
        if (type === 'proxy') return 'default';
        if (type === 'block') return 'destructive';
        return 'secondary';
    };

    const getRuleTypeLabel = (type: RuleType): string =>
        ruleTypeOptions.find((o) => o.value === type)?.label ?? type;

    return (
        <div className="space-y-4">
            <div className="flex items-center justify-between">
                <div>
                    <h3 className="text-base font-semibold">规则列表</h3>
                    <p className="text-xs text-muted-foreground">
                        共 {profile.rules.length} 条规则，拖拽调整顺序
                    </p>
                </div>
                <Button
                    onClick={() => handleOpenRuleDialog()}
                    size="sm"
                    className="gap-1.5"
                >
                    <Plus className="h-3.5 w-3.5" />
                    新增规则
                </Button>
            </div>

            {profile.rules.length > 0 ? (
                <DndContext
                    sensors={sensors}
                    collisionDetection={closestCenter}
                    onDragEnd={handleDragEnd}
                >
                    <SortableContext
                        items={profile.rules.map((r) => r.id)}
                        strategy={verticalListSortingStrategy}
                    >
                        <div className="space-y-1.5">
                            {profile.rules.map((rule, index) => (
                                <SortableRuleItem
                                    key={rule.id}
                                    rule={rule}
                                    index={index}
                                    profileId={profile.id}
                                    onEdit={handleOpenRuleDialog}
                                    onDelete={handleDeleteRule}
                                    getActionLabel={getActionLabel}
                                    getActionBadgeVariant={getActionBadgeVariant}
                                    getRuleTypeLabel={getRuleTypeLabel}
                                />
                            ))}
                        </div>
                    </SortableContext>
                </DndContext>
            ) : (
                <Card className="border-dashed">
                    <CardContent className="flex flex-col items-center justify-center py-8">
                        <Filter className="h-10 w-10 text-muted-foreground/50" />
                        <p className="mt-3 text-sm text-muted-foreground">暂无规则</p>
                        <Button
                            variant="outline"
                            size="sm"
                            className="mt-3 gap-1.5"
                            onClick={() => handleOpenRuleDialog()}
                        >
                            <Plus className="h-3.5 w-3.5" />
                            添加第一条规则
                        </Button>
                    </CardContent>
                </Card>
            )}

            {/* Add/Edit Rule Dialog */}
            <Dialog open={isRuleDialogOpen} onOpenChange={setIsRuleDialogOpen}>
                <DialogContent className="sm:max-w-lg">
                    <DialogHeader>
                        <DialogTitle>{editingRule ? '编辑规则' : '新增规则'}</DialogTitle>
                        <DialogDescription>
                            {editingRule ? '修改规则配置' : '添加一条新的匹配规则'}
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                        <div className="space-y-2">
                            <Label>规则类型</Label>
                            <Select
                                value={ruleFormData.type}
                                onValueChange={(v) =>
                                    setRuleFormData({ ...ruleFormData, type: v as RuleType })
                                }
                            >
                                <SelectTrigger>
                                    <SelectValue placeholder="选择规则类型" />
                                </SelectTrigger>
                                <SelectContent>
                                    {ruleTypeOptions.map((opt) => (
                                        <SelectItem key={opt.value} value={opt.value}>
                                            {opt.label}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="rule-value">规则值</Label>
                            <Input
                                id="rule-value"
                                placeholder={
                                    ruleFormData.type.includes('regex')
                                        ? '正则表达式'
                                        : '匹配字符串'
                                }
                                value={ruleFormData.value}
                                className="font-mono"
                                onChange={(e) =>
                                    setRuleFormData({ ...ruleFormData, value: e.target.value })
                                }
                            />
                            {ruleFormData.type.includes('regex') && (
                                <p className="text-xs text-muted-foreground">使用正则表达式进行匹配</p>
                            )}
                        </div>
                        <div className="space-y-2">
                            <Label>动作</Label>
                            <Select
                                value={ruleFormData.actionType}
                                onValueChange={(v) =>
                                    setRuleFormData({ ...ruleFormData, actionType: v as ActionType })
                                }
                            >
                                <SelectTrigger>
                                    <SelectValue placeholder="选择动作" />
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
                        {ruleFormData.actionType === 'proxy' && (
                            <div className="space-y-2">
                                <Label>选择代理</Label>
                                <Select
                                    value={ruleFormData.proxyId}
                                    onValueChange={(v) =>
                                        setRuleFormData({ ...ruleFormData, proxyId: v })
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
                                {proxies.length === 0 && (
                                    <p className="text-xs text-destructive">
                                        暂无可用代理，请先添加代理
                                    </p>
                                )}
                            </div>
                        )}
                        <div className="flex items-center space-x-2">
                            <Switch
                                id="rule-enabled"
                                checked={ruleFormData.enabled}
                                onCheckedChange={(checked) =>
                                    setRuleFormData({ ...ruleFormData, enabled: checked })
                                }
                            />
                            <Label htmlFor="rule-enabled">启用规则</Label>
                        </div>
                    </div>
                    <DialogFooter>
                        <Button
                            variant="outline"
                            onClick={() => setIsRuleDialogOpen(false)}
                        >
                            取消
                        </Button>
                        <Button
                            onClick={handleSubmitRule}
                            disabled={
                                !ruleFormData.value ||
                                (ruleFormData.actionType === 'proxy' && !ruleFormData.proxyId)
                            }
                        >
                            {editingRule ? '保存' : '创建'}
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
                            你确定要删除这条规则吗？此操作无法撤销。
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogCancel>取消</AlertDialogCancel>
                        <AlertDialogAction
                            onClick={confirmDeleteRule}
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
