# Keycloak Theme Tech Stack

- **Theme Engine**: Keycloak FreeMarker Templates (`.ftl`)
- **Parent Theme**: `base`
- **Styling**: Tailwind CSS 4 (via CDN in `template.ftl`)
- **Icons**: Emoji or Custom SVGs

## Directory Structure

- `themes/hub/login`: Custom login theme for "hub" realm.
  - `template.ftl`: Base layout including Tailwind CSS setup and dark mode logic.
  - `login.ftl`: Main login page.
  - `register.ftl`: Registration page.
  - `theme.properties`: Theme configuration, inherits from `base`.
  - `resources/img`: Static images like logos.

## Development Workflow

### Template Customization
1. Modify `.ftl` files in `themes/hub/login`.
2. Use Tailwind CSS classes for styling. Tailwind 4 is loaded via CDN, so no build step is required for CSS.
3. Handle Dark Mode using the `dark` class on the `<html>` element (logic is in `template.ftl`).

### Static Resources
1. Place images in `resources/img`.
2. Reference them in templates using `${url.resourcesPath}/img/filename`.

## Testing & Quality
- Verify changes by deploying the theme to a Keycloak instance.
- Ensure responsive design and dark mode compatibility.
- Follow the project-wide consistent naming and clean architecture principles.
