import { createHash } from 'crypto';
import { readFileSync } from 'fs';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const root = join(__dirname, '..');

/**
 * 计算文件内容的 MD5（规范化换行与末尾空白后）。
 * @param {string} filePath - 相对项目根目录的路径
 * @returns {string} 十六进制 MD5
 */
function md5OfFile(filePath) {
    const content = readFileSync(join(root, filePath), 'utf8');
    const normalized = content.replace(/\r\n/g, '\n').trimEnd();
    return createHash('md5').update(normalized, 'utf8').digest('hex');
}

/** 需保持内容一致的文件对，路径相对项目根 */
const FILE_PAIRS = [
    ['src/shared/messages.ts', 'src/background/messages.ts'],
];

for (const [pathA, pathB] of FILE_PAIRS) {
    const hashA = md5OfFile(pathA);
    const hashB = md5OfFile(pathB);
    if (hashA !== hashB) {
        console.error('[check-messages] 以下两文件内容不一致:');
        console.error('  ', pathA, '->', hashA);
        console.error('  ', pathB, '->', hashB);
        process.exit(1);
    }
}
