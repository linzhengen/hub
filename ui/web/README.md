# hub web

React + Vite + TypeScript frontend for the hub project. See [AGENTS.md](AGENTS.md) for architecture and conventions.

## Prerequisites

- Node.js 26+ (see `engines` in `package.json`)
- pnpm

## Run Locally

1. Install dependencies:
   `pnpm install`
2. Run the dev server:
   `pnpm dev`

## Scripts

- `pnpm dev` — start the Vite dev server
- `pnpm build` — production build
- `pnpm preview` — preview a production build locally
- `pnpm lint` — type-check with `tsc --noEmit`
- `pnpm gen-api` — regenerate API schemas from OpenAPI/Protobuf
