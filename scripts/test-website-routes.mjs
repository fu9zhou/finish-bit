import assert from 'node:assert/strict';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const { parseHash } = require('../website/routes.js');
const topics = ['start', 'pdf', 'documents'];

assert.deepEqual(parseHash('#/', topics).name, 'home');
assert.deepEqual(parseHash('#/operations?category=pdf&q=text', topics).name, 'catalog');
assert.equal(parseHash('#/operations?category=pdf&q=text', topics).query.get('category'), 'pdf');
assert.equal(parseHash('#/operations/pdf.extract-text', topics).id, 'pdf.extract-text');
assert.equal(parseHash('#/operation/pdf.extract-text', topics).legacy, true);
assert.equal(parseHash('#/docs/pdf', topics).topic, 'pdf');
assert.equal(parseHash('#/docs', topics).topic, 'start');
assert.equal(parseHash('#/docs/missing', topics).name, 'not-found');
assert.equal(parseHash('#/anything', topics).name, 'not-found');
assert.equal(parseHash('#/operations/%E0%A4%A', topics).name, 'not-found');

console.log('website route checks passed');
