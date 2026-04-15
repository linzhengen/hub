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
- Prefer Tailwind CSS for styling.
- Keep components small, focused, and reusable.
- Ensure proper TypeScript typing for all props and state.

### State Management
- Use TanStack Query for all server state (fetching, caching, synchronization).
- Use React Context or local state for UI-only state.

## Testing & Quality

- Run `pnpm lint` (which runs `tsc --noEmit`) to check for type errors.
- Follow the project-wide TDD and strong typing practices.
- Ensure all new components are responsive and accessible.
