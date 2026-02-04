import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { resolve } from 'path';
import type { Plugin } from 'rollup';

const root = resolve(__dirname, 'src');
const dist = resolve(__dirname, 'dist');
const alias = { '@': root };

/** Strip ESM export so background SW can run as classic script (no "import/export" error). */
function stripBackgroundExport(): Plugin {
    return {
        name: 'strip-background-export',
        generateBundle(_, bundle) {
            const f = bundle['background/service-worker.js'];
            if (f && f.type === 'chunk' && f.code) {
                f.code = f.code.replace(/\s*export\s*\{\s*\}\s*;?\s*$/, '');
            }
        },
    };
}

export default defineConfig({
    plugins: [react(), stripBackgroundExport()],
    root,
    publicDir: false,
    build: {
        outDir: dist,
        emptyOutDir: true,
        rollupOptions: {
            input: {
                popup: resolve(__dirname, 'src/popup/popup.html'),
                settings: resolve(__dirname, 'src/settings/settings.html'),
                'background/service-worker': resolve(
                    __dirname,
                    'src/background/service-worker.ts'
                ),
            } as Record<string, string>,
            output: {
                entryFileNames: (chunkInfo) =>
                    chunkInfo.name === 'background/service-worker'
                        ? 'background/service-worker.js'
                        : '[name]/[name].js',
                chunkFileNames: 'assets/[name]-[hash].js',
                assetFileNames: 'assets/[name]-[hash][extname]',
            },
        },
    },
    resolve: { alias },
});
