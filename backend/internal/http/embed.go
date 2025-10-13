package http

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"

	rez "github.com/rezible/rezible"
	"github.com/rs/zerolog/log"
)

var (
	frontendFilesDir = "frontend-dist"
	//go:embed all:frontend-dist/*
	frontendFiles embed.FS
)

func GetEmbeddedFrontendFiles() (fs.FS, error) {
	files, filesErr := fs.Sub(frontendFiles, frontendFilesDir)
	if filesErr != nil {
		return nil, fmt.Errorf("failed to open embedded frontend files: %w", filesErr)
	}
	// TODO: check frontend files dir is not empty
	return files, nil
}

func makeEmbeddedFrontendFilesServer(files fs.FS) http.Handler {
	if rez.DebugMode {
		// redirect to frontend vite dev server
		return http.RedirectHandler(rez.FrontendUrl, http.StatusFound)
	}

	fileServer := http.FileServer(http.FS(files))
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filePath := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		f, openErr := files.Open(filePath)
		if openErr == nil {
			defer func(f fs.File) {
				if closeErr := f.Close(); closeErr != nil {
					log.Error().
						Err(closeErr).
						Str("filePath", filePath).
						Msg("failed to close file")
				}
			}(f)
		}
		// redirect to index.html if no file matched
		if os.IsNotExist(openErr) {
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})

	return handlerFunc
}

var (
	docsBodyScalar = []byte(`<!doctype html>
<html lang="en">
	<head>
		<title>API Reference</title>
		<meta charset="utf-8" />
		<meta
		name="viewport"
		content="width=device-width, initial-scale=1" />
	</head>
	<body>
		<script id="api-reference" data-url="/api/v1/openapi.json"></script>
		<script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
	</body>
</html>`)

	docsBodyStoplight = []byte(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="referrer" content="same-origin" />
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no" />
    <title>API Dev Docs</title>
    <link href="https://unpkg.com/@stoplight/elements/styles.min.css" rel="stylesheet" />
    <script src="https://unpkg.com/@stoplight/elements/web-components.min.js"></script>
  </head>
  <body style="height: 100vh;">
    <elements-api
      apiDescriptionUrl="/api/v1/openapi.json"
      router="hash"
      layout="sidebar"
      tryItCredentialsPolicy="same-origin"
    />
  </body>
</html>`)
)

func serveApiDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if _, wErr := w.Write(docsBodyScalar); wErr != nil {
		log.Error().Err(wErr).Msg("failed to write embedded docs body")
	}
}
