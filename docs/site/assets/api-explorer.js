// Wires Swagger UI to the specs the build copied out of server/openapi, and
// lets "Try it out" reach an API that is not this origin: the published site is
// static, so the base URL and the bearer token come from the two inputs above
// and never leave the browser.
(() => {
  const container = document.getElementById('swagger-ui');
  const baseInput = document.getElementById('api-base');
  const tokenInput = document.getElementById('api-token');
  const urls = JSON.parse(container.dataset.specs);

  const KEYS = { base: 'hub-docs.api-base', token: 'hub-docs.api-token' };
  const load = (key) => {
    try {
      return window.localStorage.getItem(key) ?? '';
    } catch {
      return '';
    }
  };
  const save = (key, value) => {
    try {
      window.localStorage.setItem(key, value);
    } catch {
      /* private browsing: the value simply is not remembered */
    }
  };

  baseInput.value = load(KEYS.base);
  tokenInput.value = load(KEYS.token);
  baseInput.addEventListener('change', () => save(KEYS.base, baseInput.value.trim()));
  tokenInput.addEventListener('change', () => save(KEYS.token, tokenInput.value.trim()));

  SwaggerUIBundle({
    domNode: container,
    urls,
    'urls.primaryName': urls[0]?.name,
    deepLinking: true,
    docExpansion: 'list',
    defaultModelsExpandDepth: 0,
    tryItOutEnabled: true,
    persistAuthorization: false,
    presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
    layout: 'StandaloneLayout',
    requestInterceptor(request) {
      const base = baseInput.value.trim().replace(/\/$/, '');
      if (base) {
        // The specs carry no `servers`, so Swagger UI resolves paths against
        // this origin. Re-point the request at the API the reader is running.
        const path = new URL(request.url, window.location.href);
        request.url = base + path.pathname + path.search;
      }
      const token = tokenInput.value.trim();
      if (token) request.headers.Authorization = `Bearer ${token}`;
      return request;
    },
  });
})();
