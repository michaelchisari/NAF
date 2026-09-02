package main

import (
	"crypto/tls"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/valyala/fasttemplate"
)

const SERVER_PORT = ":9003"
const TPL_OPEN = "<?= "
const TPL_CLOSE = " ?>"

//go:embed static/*
var staticFiles embed.FS

//go:embed templates/pages/*
var tplPagesFiles embed.FS
var tplPages = make(map[string]*fasttemplate.Template)

//go:embed templates/components/*
var tplComponentsFiles embed.FS
var tplComponents = make(map[string]*fasttemplate.Template)

//go:embed templates/layouts/*
var tplLayoutsFiles embed.FS
var tplLayouts = make(map[string]*fasttemplate.Template)

func init() {
	// Pages
	pages, err := fs.Glob(tplPagesFiles, "templates/pages/*.html")
	if err != nil {
		log.Fatal(err)
	}

	for _, pagePath := range pages {
		pageContent, err := tplPagesFiles.ReadFile(pagePath)
		if err != nil {
			log.Fatalf("Error reading embedded files: %s: %v\n", pagePath, err)
		}
		p := path.Base(pagePath)
		i := strings.TrimSuffix(p, ".html")
		tplPages[i] = fasttemplate.New(string(pageContent), TPL_OPEN, TPL_CLOSE)
	}

	// Components
	components, err := fs.Glob(tplComponentsFiles, "templates/components/*.html")
	if err != nil {
		log.Fatal(err)
	}

	for _, componentPath := range components {
		componentContent, err := tplComponentsFiles.ReadFile(componentPath)
		if err != nil {
			log.Fatalf("Error reading embedded files: %s: %v\n", componentPath, err)
		}
		p := path.Base(componentPath)
		i := strings.TrimSuffix(p, ".html")
		tplComponents[i] = fasttemplate.New(string(componentContent), TPL_OPEN, TPL_CLOSE)
	}

	// Layouts
	layouts, err := fs.Glob(tplLayoutsFiles, "templates/layouts/*.html")
	if err != nil {
		log.Fatal(err)
	}

	for _, layoutPath := range layouts {
		layoutContent, err := tplLayoutsFiles.ReadFile(layoutPath)
		if err != nil {
			log.Fatalf("Error reading embedded files: %s: %v\n", layoutPath, err)
		}
		p := path.Base(layoutPath)
		i := strings.TrimSuffix(p, ".html")
		tplLayouts[i] = fasttemplate.New(string(layoutContent), TPL_OPEN, TPL_CLOSE)
	}
}

func main() {
	mux := http.NewServeMux()

	subFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.FS(subFS))))

	mux.HandleFunc("/", handleHubAtRoot)
	mux.HandleFunc("/elements", handleTodoPage)         // Tabs, accordions, dialogs, toasts (solved but not widely known)
	mux.HandleFunc("/navigation", handleNavigationPage) // smooth navigation across pages
	mux.HandleFunc("/infinite", handleTodoPage)         // infinite scroll
	mux.HandleFunc("/search", handleTodoPage)           // live search
	mux.HandleFunc("/tables", handleTodoPage)           // filterable/sortable tables
	mux.HandleFunc("/photos", handleTodoPage)           // photo album
	mux.HandleFunc("/wizard", handleTodoPage)           // form wizard
	mux.HandleFunc("/shopping", handleTodoPage)         // shopping site

	// Use mkcert for a trusted local cert
	// https://github.com/filosottile/mkcert
	//
	// Use http/2 by default
	const CERTFILE = "server.crt"
	const KEYFILE = "server.key"
	// Pre-validate the certificate and key pair
	_, err = tls.LoadX509KeyPair(CERTFILE, KEYFILE)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// TODO: Output instructions for generating a local cert
			log.Fatalf("Certificate or key file not found: %v", err)
		}
		log.Fatalf("Invalid TLS configuration (malformed certificate or key): %v", err)
	}

	log.Println("Server starting on", SERVER_PORT)
	err = http.ListenAndServeTLS(SERVER_PORT, CERTFILE, KEYFILE, mux)
	if err != nil {
		log.Fatal(err)
	}
}

func handleHubAtRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hub")
}

func handleNavigationPage(w http.ResponseWriter, r *http.Request) {
	headerNavComponent, ok := tplComponents["header_nav"]
	if !ok {
		slog.Error("Could not find component: header_nav")
		http.Error(w, "Could not find component: header_nav", 500)
		return
	}

	navigationPage, ok := tplPages["navigation"]
	if !ok {
		slog.Error("Could not find page: navigation")
		http.Error(w, "Could not find page: navigation", 500)
		return
	}

	baseLayout, ok := tplLayouts["base"]
	if !ok {
		slog.Error("Could not find layout: base")
		http.Error(w, "Could not find layout: base", 500)
		return
	}

	h := headerNavComponent.ExecuteString(map[string]any{})

	n := navigationPage.ExecuteString(map[string]any{
		"header_nav": h,
	})

	b := baseLayout.ExecuteString(map[string]any{
		"page": n,
	})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, b)
}

func handleTodoPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hub")
}
