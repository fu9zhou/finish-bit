# FinishBit website

Dependency-free static website with English and Simplified Chinese homepages, twelve documentation topics, and a full generated capability catalog with individual input/option pages. Content follows the repository documentation. The terminal is an illustrative workflow; the website does not execute CLI commands.

## Local preview

From the repository root, use any static HTTP server, for example:

```sh
python -m http.server 4173 --bind 127.0.0.1 --directory website
```

Open http://127.0.0.1:4173/. No package installation or frontend build is required.

## GitHub Pages

The workflow `.github/workflows/pages.yml` publishes only the website assets. It runs manually or when website files change on `main`.

1. Push the reviewed website and workflow to the repository.
2. In repository **Settings → Pages → Build and deployment → Source**, select **GitHub Actions**.
3. Run **Website / GitHub Pages** from the Actions tab, or push another website change to `main`.
4. Use the deployment URL shown by the workflow. For the current repository, the expected default is https://fu9zhou.github.io/finish-bit/.

The workflow is prepared locally; this change does not enable Pages or publish the website by itself. If Pages is not configured before the first push, enable it and rerun the workflow.

Relative asset paths work under `/finish-bit/` and custom-domain roots. Hash routes (for example `#/docs/start`) allow direct links and refreshes without server rewrite rules. `.nojekyll` also allows the assets to be hosted without Jekyll processing. No SPA 404 workaround is needed for supported hash URLs.

Reference: [GitHub's custom Pages workflow documentation](https://docs.github.com/en/pages/getting-started-with-github-pages/using-custom-workflows-with-github-pages).

## Languages

The first browser language/locale selects Chinese for `zh` or region `CN`, otherwise English. The language selector persists a manual override in localStorage. Storage restrictions do not prevent rendering. This is browser locale adaptation, not IP geolocation; it needs no location permission or external service. Both languages remain available wherever a visitor is located.

## Design and maintenance

The UI/UX Pro Max minimal / Swiss documentation direction is adapted with warm neutral surfaces, forest green, monospaced command blocks, visible keyboard focus, responsive layouts and reduced-motion support. System fonts avoid external font dependencies.

- `index.html`: static shell and metadata.
- `styles.css`: layout, responsive rules and semantic color tokens.
- `app.js`: translated content, routes, documentation search, catalog filtering and clipboard feedback.
- `catalog.js`: generated application contracts, including inputs, options and dependency requirements.

Keep CLI examples and requirements synchronized with the repository documentation. Regenerate the source catalog with `go build ./cmd/fnsh` then `node scripts/website-catalog.mjs`. Run `node scripts/website-catalog.mjs --check` to detect drift; CI enforces this check. Generation uses an isolated runtime home so local extensions never enter the website. `fnsh capabilities --json` remains authoritative for a user's installed version.

The website marks Operations introduced in v0.1.4 and explains managed package platform support. The v0.1.4 catalog has 170 Operations; future additions must clearly identify their release availability. Category filters have shareable URLs, such as `#/operations?category=pdf`. PDF, image, table, document and archive guides include installation commands and tested platform boundaries.
