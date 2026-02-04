import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { resolve } from 'path';

export default defineConfig({
    plugins: [react()],
    root: resolve(__dirname, 'src'),
    publicDir: false,
    build: {
        outDir: resolve(__dirname, 'dist'),
        emptyOutDir: true,
        rollupOptions: {
            input: {
                popup: resolve(__dirname, 'src/popup/popup.html'),
                settings: resolve(__dirname, 'src/settings/settings.html'),
            } as Record<string, string>,
            output: {
                entryFileNames: '[name]/[name].js',
                chunkFileNames: 'assets/[name]-[hash].js',
                assetFileNames: 'assets/[name]-[hash][extname]',
            },
        },
    },
    resolve: {
        alias: {
            '@': resolve(__dirname, 'src'),
        },
    },
});
