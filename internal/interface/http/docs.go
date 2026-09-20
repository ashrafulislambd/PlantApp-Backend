package httpapi

import (
	_ "embed"
	"net/http"
)

//go:embed openapi.yaml
var openAPISpec []byte

// swaggerUIPage loads Swagger UI's static assets from a CDN and points it
// at the embedded spec below, so no swagger-ui files need to be vendored.
const swaggerUIPage = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <title>MyPlantPal API Docs</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = () => {
      window.ui = SwaggerUIBundle({
        url: '/openapi.yaml',
        dom_id: '#swagger-ui',
      });
    };
  </script>
</body>
</html>
`

func openAPISpecHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	_, _ = w.Write(openAPISpec)
}

func docsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(swaggerUIPage))
}
