import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { resolve } from 'path';

export default defineConfig({
    plugins: [react()],
    resolve: {
        alias: {
            '@': resolve(__dirname, 'src'),
        },
    },
    build: {
        outDir: 'dist',
        emptyOutDir: true,
        rollupOptions: {
            input: {
                popup: resolve(__dirname, 'src/popup/popup.html'),
                settings: resolve(__dirname, 'src/settings/settings.html'),
                'service-worker': resolve(__dirname, 'src/background/service-worker.ts'),
            },
            output: {
                entryFileNames: (chunkInfo) => {
                    // service-worker 需要放在根目录
                    if (chunkInfo.name === 'service-worker') {
                        return '[name].js';
                    }
                    return 'assets/[name]-[hash].js';
                },
                chunkFileNames: 'assets/[name]-[hash].js',
                assetFileNames: 'assets/[name]-[hash].[ext]',
            },
        },
    },
});
