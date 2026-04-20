import { spawnSync } from 'node:child_process';

const forwarded = process.argv.slice(2).filter((arg) => arg !== '--runInBand');
const result = spawnSync('npx', ['vitest', 'run', ...forwarded], {
  stdio: 'inherit',
  shell: process.platform === 'win32'
});

process.exit(result.status ?? 1);
