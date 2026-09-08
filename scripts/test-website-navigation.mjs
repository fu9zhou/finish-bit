import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import vm from 'node:vm';

// Exercise the actual page renderers and their internal links without a browser.
// Layout and sticky behavior are checked separately in the browser.
const app = { innerHTML: '' };
const location = { hash: '#/' };
const context = vm.createContext({
  URLSearchParams, location,
  navigator: { language: 'zh-CN' },
  localStorage: { getItem: () => null },
  history: { replaceState: (_state, _title, hash) => { location.hash = hash; } },
  document: {
    documentElement: {}, body: { classList: { toggle() {} } },
    querySelector: selector => selector === '#app' ? app : null,
    addEventListener() {},
  },
  addEventListener() {},
});
for (const file of ['routes.js', 'catalog.js', 'home.js', 'app.js']) {
  vm.runInContext(readFileSync(new URL('../website/' + file, import.meta.url), 'utf8'), context, { filename: file });
}
const topicIds = vm.runInContext('topics.map(item => item[0])', context);
const operationIds = vm.runInContext('operations.map(item => item.id)', context);
const guideIds = vm.runInContext('Object.keys(capabilityGuides)', context);
const paths = [
  '#/', '#/operations',
  ...topicIds.map(id => '#/docs/' + id),
  ...guideIds.map(id => '#/operations/guides/' + id),
  ...operationIds.map(id => '#/operations/' + id),
  ...vm.runInContext('categories.slice(1).map(item => "#/operations?category=" + item[0])', context),
];
for (const language of ['zh', 'en']) {
  vm.runInContext(`lang = ${JSON.stringify(language)}`, context);
  for (const path of paths) {
    location.hash = path;
    vm.runInContext('render()', context);
    assert.equal((app.innerHTML.match(/<h1(?:\s[^>]*)?>/g) || []).length, 1, `${language} ${path}: one page heading`);
    for (const [, href] of app.innerHTML.matchAll(/href="(#[^"]+)"/g)) {
      if (href === '#main') continue;
      const target = context.finishbitRoutes.parseHash(href.replaceAll('&amp;', '&'), topicIds);
      assert.notEqual(target.name, 'not-found', `${language} ${path}: broken link ${href}`);
      if (target.name === 'operation') assert.ok(operationIds.includes(target.id), `Unknown operation ${target.id}`);
      assert.equal(target.redirect, undefined, `${path}: use the canonical URL ${href}`);
    }
  }
}
assert.equal(topicIds.length, 5);
for (const [oldPath, newPath] of [
  ['#/docs/agents', '#/docs/start'],
  ['#/docs/discovery', '#/docs/execute'],
  ...guideIds.map(id => ['#/docs/' + id, '#/operations/guides/' + id]),
]) {
  location.hash = oldPath;
  vm.runInContext('render()', context);
  assert.equal(location.hash, newPath);
}
console.log(`website navigation checks passed: ${paths.length} pages in both languages, 5 guide topics, legacy redirects`);

// Follow rendered URLs, including a fresh render as on reload or language change.
function visit(hash) {
  location.hash = hash;
  vm.runInContext('render()', context);
  return app.innerHTML;
}
function linkWithClass(html, className) {
  const match = html.match(new RegExp('class="' + className + '" href="([^\"]+)"'));
  assert.ok(match, `Missing ${className} link`);
  return match[1].replaceAll('&amp;', '&');
}
for (const language of ['zh', 'en']) {
  vm.runInContext(`lang = ${JSON.stringify(language)}`, context);
  const listing = visit('#/operations?category=pdf&q=extract');
  const detailURL = listing.match(/href="([^"]*pdf\.extract-text[^"]*)"/)[1].replaceAll('&amp;', '&');
  const detail = visit(detailURL);
  const back = linkWithClass(detail, 'back-link');
  const restored = visit(back);
  assert.match(restored, /value="extract"/);
  assert.match(restored, /data-filter="pdf" aria-pressed="true"/);
  const guideURL = linkWithClass(visit(detailURL), 'textlink');
  assert.equal(context.finishbitRoutes.parseHash(linkWithClass(visit(guideURL), 'back-link'), topicIds).query.get('q'), 'extract');

  for (const [id, category] of [['ocr.text','images'],['pdf.to-docx','pdf'],['screen.record','media']]) {
    const detail = visit('#/operations/' + id);
    const description = vm.runInContext(`operations.find(o=>o.id===${JSON.stringify(id)}).description`, context);
    assert.ok(detail.includes(description), `${id}: full behavior description`);
    if (id === 'screen.record') assert.ok(detail.includes(language === 'zh' ? '仅支持 Windows' : 'Windows only'));
    const guide = visit(linkWithClass(detail, 'textlink'));
    const returnURL = linkWithClass(guide, 'back-link');
    assert.equal(context.finishbitRoutes.parseHash(returnURL, topicIds).query.get('category'), category);
    assert.ok(visit(returnURL).includes(id), `${id}: related catalog contains the originating operation`);
  }
  const genericGuide = visit('#/operations/guides/local');
  const genericCatalog = visit(linkWithClass(genericGuide, 'back-link'));
  for (const id of ['ocr.text','pdf.to-docx','screen.record']) assert.ok(genericCatalog.includes(id));
}
console.log('website context checks passed: filters, queries, guide returns and platform limits');
