(function (root, factory) {
  const routes = factory();
  if (typeof module === 'object' && module.exports) module.exports = routes;
  root.finishbitRoutes = routes;
})(typeof globalThis !== 'undefined' ? globalThis : this, function () {
  const guideIds = ['media', 'pdf', 'images', 'tables', 'documents', 'archives', 'local'];
  const legacyDocs = { agents: 'start', discovery: 'execute' };
  function parseHash(hash, topicIds) {
    const raw = String(hash || '#/').replace(/^#/, '');
    const question = raw.indexOf('?');
    const rawPath = question === -1 ? raw : raw.slice(0, question);
    const queryText = question === -1 ? '' : raw.slice(question + 1);
    let path = rawPath || '/';
    if (path.length > 1) path = path.replace(/\/+$/, '');
    const query = new URLSearchParams(queryText);

    if (path === '/') return { name: 'home', path, query };
    if (path === '/operations') return { name: 'catalog', path, query };

    const decode = value => {
      try { return decodeURIComponent(value); } catch { return null; }
    };

    let match = path.match(/^\/operations\/guides\/([^/]+)$/);
    if (match) {
      const topic = decode(match[1]);
      return guideIds.includes(topic) ? { name: 'guide', path, topic, query } : { name: 'not-found', path, query };
    }

    match = path.match(/^\/operations\/([^/]+)$/);
    if (match) {
      const id = decode(match[1]);
      return id === null ? { name: 'not-found', path, query } : { name: 'operation', path, id, query };
    }

    match = path.match(/^\/operation\/([^/]+)$/);
    if (match) {
      const id = decode(match[1]);
      return id === null ? { name: 'not-found', path, query } : { name: 'operation', path, id, query, legacy: true };
    }

    if (path === '/docs') return { name: 'docs', path, topic: 'start', query };
    match = path.match(/^\/docs\/([^/]+)$/);
    if (match) {
      const topic = decode(match[1]);
      if (guideIds.includes(topic)) return { name: 'guide', path, topic, query, redirect: '/operations/guides/' + topic };
      if (Object.hasOwn(legacyDocs, topic)) return { name: 'docs', path, topic: legacyDocs[topic], query, redirect: '/docs/' + legacyDocs[topic] };
      if (topicIds.includes(topic)) return { name: 'docs', path, topic, query };
    }

    return { name: 'not-found', path, query };
  }

  return { parseHash };
});
