import assert from 'node:assert/strict';
import { readFileSync, readdirSync, statSync } from 'node:fs';
import { resolve } from 'node:path';

// Check the deployable directory, not just the source tree: missing assets in
// the Pages upload must fail before deployment.
const root = resolve(process.argv[2] || 'website');
let checked = 0;
for (const file of readdirSync(root)) {
  if (!/\.(html|js|css)$/.test(file)) continue;
  const source = readFileSync(resolve(root, file), 'utf8');
  for (const match of source.matchAll(/["'(](\.\/[^"'()\s?#]+\.(?:png|svg|webp|jpg|ico|css|js))(?:[?#][^"'()\s]*)?["')]/g)) {
    const path = resolve(root, match[1]);
    assert.ok(statSync(path, { throwIfNoEntry: false })?.isFile(), `${file}: missing deployed asset ${match[1]}`);
    if (path.endsWith('.png')) {
      assert.equal(readFileSync(path).subarray(0, 8).toString('hex'), '89504e470d0a1a0a', `${match[1]} is not a PNG`);
    }
    checked++;
  }
}
assert.ok(checked > 0, 'No local asset references checked');
console.log(`Verified ${checked} local asset references in ${root}`);
