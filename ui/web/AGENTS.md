## Frontend Tech Stack

- **Framework**: React 19 with Vite
- **Language**: TypeScript (Strict Mode)
- **Styling**: Tailwind CSS 4, Lucide Icons
- **UI Components**: Shadcn UI (located in `src/components/ui`), Ant Design Icons
- **State Management & Data Fetching**: TanStack Query (React Query) v5
- **Routing**: React Router v7
- **Forms**: React Hook Form with Zod validation
- **API Client**: Custom `fetchApi` wrapper in `src/lib/api-client.ts`

## Directory Structure

- `src/api/schema`: Generated TypeScript definitions from OpenAPI/Protobuf.
- `src/services`: Domain-specific API service wrappers using `fetchApi`.
- `src/hooks`: Custom React hooks, including data fetching hooks using TanStack Query.
- `src/components/ui`: Base UI components (mostly Shadcn UI).
- `src/components/common`: Shared business components.
- `src/pages`: Application pages/routes.
- `src/lib`: Utility functions and shared configurations.
- `src/providers`: React context providers (QueryClient, Auth, Theme).

## Development Workflow

### API Integration
1. Generate/Update schemas in `src/api/schema` using `pnpm gen-api`.
2. Implement or update a service in `src/services` that uses the generated types and `fetchApi`.
3. Use `@tanstack/react-query` to consume the services in components/pages (using `useQuery` or `useMutation`).
4. If needed, create custom hooks in `src/hooks` for complex data fetching or state logic.

### Component Guidelines
- Use Shadcn UI primitives from `src/components/ui` whenever possible.
- To add a new Shadcn component, run the CLI on demand — `npx shadcn@latest add <component>`
  (it reads `components.json`). The CLI is deliberately not a project dependency:
  it pulls in a large server-side tree that only added vulnerability alerts.
- Prefer Tailwind CSS for styling.
- Keep components small, focused, and reusable.
- Ensure proper TypeScript typing for all props and state.

### State Management
- Use TanStack Query for all server state (fetching, caching, synchronization).
- Use React Context or local state for UI-only state.

## Testing & Quality

- Run `pnpm lint` (which runs `tsc --noEmit`) to check for type errors.
- Run `pnpm lint:eslint` to check for ESLint issues (`eslint.config.js`).
- Run `pnpm test` (Vitest, single run) or `pnpm test:watch` while developing.
  All three are run in CI (`web-ci.yml`).
- Follow the project-wide TDD and strong typing practices.
- Ensure all new components are responsive and accessible.

### Unit Tests

- **Runner**: Vitest with the `jsdom` environment, configured in the `test`
  block of `vite.config.ts` (which imports `defineConfig` from `vitest/config`).
  `src/test/setup.ts` holds the global setup — currently a `matchMedia` stub,
  which Ant Design's responsive components need and jsdom does not provide.
- **Location**: co-located with the code under test as `*.test.ts` / `*.test.tsx`
  under `src/`. The `@/` alias resolves in tests as it does in the app.
- **Components/hooks**: use `@testing-library/react`. Query by role/label rather
  than by class name so the tests keep tracking accessibility.
- Prefer testing observable behaviour (returned values, dispatched events,
  rendered roles) over implementation details.

### E2E Tests

- **Runner**: Playwright (`playwright.config.ts`), specs in `e2e/`. Run with
  `pnpm test:e2e`, or `pnpm test:e2e:ui` for the interactive runner. CI runs
  them in the separate `Web E2E` job in `web-ci.yml`.
- **No servers required.** Keycloak's authorization/token/logout endpoints and
  the backend API are stubbed at the network boundary via Playwright route
  interception (`e2e/keycloak-stub.ts`, `e2e/api-stub.ts`). The app itself —
  keycloak-js, AuthProvider, api-client — runs unmodified. That means these
  tests cover how the app drives the OIDC flow, not Keycloak's own behaviour
  (signature validation, session management, SMTP).
- The suite runs against the **dev server, not a production build**: React only
  double-invokes effects in development, and that is the condition under which
  Keycloak initialization used to break. Do not switch it to `vite preview`.
- If the environment already ships a Chromium, point
  `PLAYWRIGHT_CHROMIUM_PATH` at it instead of downloading a matching build.

### Dialogs

- Never use the native `window.confirm` / `window.alert`. They do not follow the
  app's Ant Design theme and behave differently for focus trapping and keyboard
  navigation.
- For destructive confirmations use `useDangerConfirm` (`src/hooks/useDangerConfirm.ts`).
- For other modals, take `modal` / `message` / `notification` from
  `App.useApp()` rather than the static `Modal.*` methods: the app is wrapped in
  antd's `<App>` (`src/App.tsx`), and only the hook-based API inherits the
  `ConfigProvider` theme (light/dark).
