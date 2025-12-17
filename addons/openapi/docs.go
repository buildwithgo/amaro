package openapi

import "fmt"

// ScalarHTML returns a simple HTML page that loads the Scalar API reference.
// url is the path to the OpenAPI JSON file (e.g. "/swagger.json").
func ScalarHTML(url string) string {
	return fmt.Sprintf(`<!doctype html>
<html>
  <head>
    <title>API Reference</title>
    <meta charset="utf-8" />
    <meta
      name="viewport"
      content="width=device-width, initial-scale=1" />
    <style>
      body {
        margin: 0;
      }
    </style>
  </head>
  <body>
    <script
      id="api-reference"
      data-url="%s"></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
  </body>
</html>`, url)
}
