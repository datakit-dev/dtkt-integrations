import fs from 'fs';
import path from 'path';
import archiver from 'archiver';

const distPath = path.resolve('dist');
const pkg = JSON.parse(fs.readFileSync(path.resolve('package.json'), 'utf-8'));
const zipPath = path.resolve(`zips/extension-v${pkg.version}.zip`);

// make sure dist exists
if (!fs.existsSync(distPath)) {
  console.error('❌ dist/ folder not found. Run `pnpm build` first.');
  process.exit(1);
}

const output = fs.createWriteStream(zipPath);
const archive = archiver('zip', { zlib: { level: 9 } });

output.on('close', () => {
  console.log(`✅ Created ${zipPath} (${archive.pointer()} bytes)`);
});

archive.on('error', (err) => {
  throw err;
});

archive.pipe(output);
archive.directory(distPath, false);
archive.finalize();
