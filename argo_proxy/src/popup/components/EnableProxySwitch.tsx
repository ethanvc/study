interface EnableProxySwitchProps {
    enabled: boolean;
    onToggle: () => void;
}

export default function EnableProxySwitch({ enabled, onToggle }: EnableProxySwitchProps) {
    return (
        <div className="flex items-center justify-between gap-4 p-3 mb-2 bg-white rounded-lg">
            <span className="font-medium text-sm">启用代理路由</span>
            <button
                type="button"
                role="switch"
                aria-checked={enabled}
                onClick={onToggle}
                className={`relative w-10 h-6 rounded-full flex-shrink-0 transition-colors ${enabled ? 'bg-green-500' : 'bg-gray-400'
                    }`}
            >
                <span
                    className={`absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full transition-transform ${enabled ? 'translate-x-0' : 'translate-x-4'
                        }`}
                />
            </button>
        </div>
    );
}
