const https = require('https');
const fs = require('fs');
const path = require('path');
const zlib = require('zlib');
const { execSync } = require('child_process');

const pkg = require('./package.json');
const version = pkg.version;

if (process.env.OC_USAGE_BINARY) {
  console.log(`OC_USAGE_BINARY is set, skipping download: ${process.env.OC_USAGE_BINARY}`);
  process.exit(0);
}

const platform = process.platform;
const arch = process.arch;

const osMap = { darwin: 'darwin', linux: 'linux', win32: 'windows' };
const archMap = { arm64: 'arm64', x64: 'amd64', amd64: 'amd64' };

const os = osMap[platform];
const binaryArch = archMap[arch];

if (!os || !binaryArch) {
  console.warn(`oc-usage: Unsupported platform ${platform}-${arch}, skipping binary download.`);
  process.exit(0);
}

let binName = `bin-${os}-${binaryArch}`;
if (platform === 'win32') binName += '.exe';

const binPath = path.join(__dirname, binName);

if (fs.existsSync(binPath)) {
  process.exit(0);
}

const remoteName = platform === 'win32' ? `oc-usage-${os}-${binaryArch}.exe` : `oc-usage-${os}-${binaryArch}`;
const url = `https://github.com/sigco3111/opencode-usage-cli/releases/download/v${version}/${remoteName}`;

function download(url, dest) {
  return new Promise((resolve, reject) => {
    function follow(url, redirects) {
      if (redirects > 10) return reject(new Error('Too many redirects'));
      https.get(url, { headers: { 'User-Agent': 'node' } }, (res) => {
        if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
          return follow(res.headers.location, redirects + 1);
        }
        if (res.statusCode !== 200) {
          return reject(new Error(`HTTP ${res.statusCode}`));
        }
        const stream = fs.createWriteStream(dest, { mode: 0o755 });
        res.pipe(stream);
        stream.on('finish', () => {
          stream.close();
          fs.chmodSync(dest, 0o755);
          resolve();
        });
        stream.on('error', reject);
        res.on('error', reject);
      }).on('error', reject);
    }
    follow(url, 0);
  });
}

download(url, binPath)
  .then(() => {
    console.log(`oc-usage v${version}: downloaded ${binName}`);
  })
  .catch((e) => {
    console.warn(`oc-usage: Failed to download binary: ${e.message}`);
    console.warn('You can build from source: cd opencode-usage && CGO_ENABLED=0 go build -o oc-usage .');
  });
