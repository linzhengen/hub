# Documentation Site

The site published to GitHub Pages at <https://linzhengen.github.io/hub/>.

## Tech Stack

- **Build**: a single Node script, `build.mjs` — no framework, no static site generator
- **Markdown**: `marked` with build-time syntax highlighting from `highlight.js`
- **Diagrams**: `mermaid`, loaded only on pages that contain a diagram
- **API explorer**: `swagger-ui-dist`
- **Package management**: `pnpm`

Every asset is vendored out of `node_modules` at build time, so the published
pages load nothing from a third-party CDN.

## Directory Structure

- `build.mjs`: the whole build. Nav, languages and page list live at the top.
- `site/`: hand-written sources.
  - `*.md`, `ja/*.md`: the guides, one Markdown file per page and language.
  - `assets/`: stylesheet, favicon, the API explorer's script.
  - `partials/swagger.html`: the body of the API explorer page.
- `screenshots/`: images used by the guides and by the READMEs.
- `_site/`: build output. Git-ignored; never edit or commit it.

## Development Workflow

```bash
pnpm install
pnpm serve      # builds _site and serves it on http://localhost:4000
pnpm build      # build only
```

Adding a page means adding the Markdown file under `site/` and one entry to
`NAV` in `build.mjs`, for each language. Headings get anchor ids automatically.

## Generated Content

Two pages are not written here and must not be copied here:

- **API explorer** (`api.html`) reads the OpenAPI documents in
  `server/openapi/`, which `make gen` derives from the protos. A new service
  appears in the definition selector on its own. Only the `info` block is
  rewritten at build time, to replace the `.proto` file name that
  protoc-gen-openapiv2 uses as a title.
- **API reference** (`api-reference.html`) renders
  `.agents/skills/hub-api/references/api-reference.md`, which `hub api docs`
  generates. Fix it at the proto and re-run `make gen`.

## Publishing

`.github/workflows/pages.yml` builds the site and deploys it on every push to
`main` that touches `docs/`, `server/openapi/` or the generated API reference.
Pull requests build the site without publishing. The repository must have Pages
enabled with **Source: GitHub Actions**.
