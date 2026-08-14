// Wires Swagger UI to the specs the build copied out of server/openapi, and
// lets "Try it out" reach an API that is not this origin: the published site is
// static, so the base URL and the bearer token come from the two inputs above
// and never leave the browser.
//
// The endpoint is remembered across visits; the token deliberately is not. It
// is a live credential for someone's hub instance, so it stays in the form
// field and is gone when the tab closes.
(() => {
  const container = document.getElementById('swagger-ui');
  const baseInput = document.getElementById('api-base');
  const tokenInput = document.getElementById('api-token');
  const urls = JSON.parse(container.dataset.specs);

  const BASE_KEY = 'hub-docs.api-base';
  try {
    baseInput.value = window.localStorage.getItem(BASE_KEY) ?? '';
  } catch {
    /* storage blocked: the endpoint simply is not remembered */
  }
  baseInput.addEventListener('change', () => {
    try {
      window.localStorage.setItem(BASE_KEY, baseInput.value.trim());
    } catch {
      /* as above */
    }
  });

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
