import { cpSync, mkdirSync, existsSync } from 'fs';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const root = join(__dirname, '..');
const staticDir = join(root, 'static');
const dist = join(root, 'dist');

if (!existsSync(staticDir)) {
    console.warn('static/ not found, skip copy');
    process.exit(0);
}

// background is built from src/background/service-worker.ts; only copy icons
const dirs = ['icons'];
for (const d of dirs) {
    const src = join(staticDir, d);
    const dest = join(dist, d);
    if (existsSync(src)) {
        mkdirSync(dest, { recursive: true });
        cpSync(src, dest, { recursive: true });
        console.log('Copied', d);
    }
}
const manifestSrc = join(staticDir, 'manifest.json');
if (existsSync(manifestSrc)) {
    cpSync(manifestSrc, join(dist, 'manifest.json'));
    console.log('Copied manifest.json');
}
