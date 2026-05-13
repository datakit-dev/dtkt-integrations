import fs from 'fs';
import path from 'path';

const packagePath = path.resolve('package.json');
const manifestPath = path.resolve('src/manifest.json');

if (!fs.existsSync(packagePath) || !fs.existsSync(manifestPath)) {
  console.error('❌ Missing package.json or manifest.json');
  process.exit(1);
}

const pkg = JSON.parse(fs.readFileSync(packagePath, 'utf-8'));
const manifest = JSON.parse(fs.readFileSync(manifestPath, 'utf-8'));

if (!pkg.version) {
  console.error('❌ No version found in package.json');
  process.exit(1);
}

if (manifest.version !== pkg.version) {
  console.log(`🔄 Syncing version: ${manifest.version} → ${pkg.version}`);
  manifest.version = pkg.version;
  fs.writeFileSync(manifestPath, JSON.stringify(manifest, null, 2));
  console.log('✅ manifest.json updated');
} else {
  console.log('✅ manifest.json already in sync');
}
