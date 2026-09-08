import assert from 'node:assert/strict';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const { parseHash } = require('../website/routes.js');
const topics = ['start', 'install', 'troubleshooting', 'execute', 'extensions'];

assert.deepEqual(parseHash('#/', topics).name, 'home');
assert.deepEqual(parseHash('#/operations?category=pdf&q=text', topics).name, 'catalog');
assert.equal(parseHash('#/operations?category=pdf&q=text', topics).query.get('category'), 'pdf');
assert.equal(parseHash('#/operations/pdf.extract-text', topics).id, 'pdf.extract-text');
assert.equal(parseHash('#/operation/pdf.extract-text', topics).legacy, true);
assert.equal(parseHash('#/docs/pdf', topics).topic, 'pdf');
for (const topic of ['media', 'pdf', 'images', 'tables', 'documents', 'archives', 'local']) {
  const legacy = parseHash('#/docs/' + topic, topics);
  assert.equal(legacy.name, 'guide');
  assert.equal(legacy.redirect, '/operations/guides/' + topic);
  const current = parseHash('#' + legacy.redirect, topics);
  assert.equal(current.name, 'guide');
  assert.equal(current.topic, topic);
}
assert.equal(parseHash('#/docs/agents', topics).topic, 'start');
assert.equal(parseHash('#/docs/discovery', topics).topic, 'execute');
assert.equal(parseHash('#/docs/troubleshooting', topics).name, 'docs');
assert.equal(parseHash('#/operations/guides/missing', topics).name, 'not-found');
assert.equal(parseHash('#/operations/guides/%E0%A4%A', topics).name, 'not-found');
assert.equal(parseHash('#/docs/toString', topics).name, 'not-found');
assert.equal(parseHash('#/docs', topics).topic, 'start');
assert.equal(parseHash('#/docs/missing', topics).name, 'not-found');
assert.equal(parseHash('#/anything', topics).name, 'not-found');
assert.equal(parseHash('#/operations/%E0%A4%A', topics).name, 'not-found');

console.log('website route checks passed');
