import { useEffect, useState } from 'react';
import { Server, FileText, Settings2 } from 'lucide-react';
import { useProxyStore } from '@/shared/store';
import { ProxyList } from './components/ProxyList';
import { ProfileManager } from './components/ProfileManager';
import { cn } from '@/lib/utils';

type Section = 'proxy' | 'profile';

const NAV_ITEMS: { id: Section; label: string; icon: typeof Server }[] = [
    { id: 'proxy', label: '代理列表', icon: Server },
    { id: 'profile', label: 'Profile 配置', icon: FileText },
];

export default function App() {
    const { loadConfig } = useProxyStore();
    const [activeSection, setActiveSection] = useState<Section>('proxy');

    useEffect(() => {
        void loadConfig();
    }, [loadConfig]);

    return (
        <div className="flex h-screen overflow-hidden bg-background text-foreground">
            {/* Sidebar */}
            <aside className="w-56 shrink-0 flex flex-col border-r border-border bg-[hsl(285,2%,14%)]">
                {/* Header */}
                <div className="flex items-center gap-2.5 px-4 py-4 border-b border-border">
                    <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/20">
                        <Settings2 className="h-4 w-4 text-primary" />
                    </div>
                    <span className="font-semibold text-sm">代理管理器</span>
                </div>

                {/* Nav */}
                <nav className="flex-1 p-2 space-y-0.5">
                    {NAV_ITEMS.map((item) => {
                        const Icon = item.icon;
                        const active = activeSection === item.id;
                        return (
                            <button
                                key={item.id}
                                type="button"
                                onClick={() => setActiveSection(item.id)}
                                className={cn(
                                    'w-full flex items-center gap-2.5 px-3 py-2 rounded-md text-sm font-medium transition-colors',
                                    active
                                        ? 'bg-primary/15 text-primary'
                                        : 'text-muted-foreground hover:bg-secondary hover:text-foreground'
                                )}
                            >
                                <Icon className="h-4 w-4 shrink-0" />
                                {item.label}
                            </button>
                        );
                    })}
                </nav>
            </aside>

            {/* Main content */}
            <main className="flex-1 overflow-auto p-6">
                <div className="max-w-4xl mx-auto">
                    {activeSection === 'proxy' && <ProxyList />}
                    {activeSection === 'profile' && <ProfileManager />}
                </div>
            </main>
        </div>
    );
}
