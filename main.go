package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/valyala/fasttemplate"
)

const SERVER_PORT = ":9003"
const TPL_OPEN = "<?= "
const TPL_CLOSE = " ?>"

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
	http.HandleFunc("/", handleHubAtRoot)
	http.HandleFunc("/navigation", handleNavigationPage)

	log.Println("Server starting on", SERVER_PORT)
	if err := http.ListenAndServe(SERVER_PORT, nil); err != nil {
		log.Fatal(err)
	}
}

func handleHubAtRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hub")
}

func handleNavigationPage(w http.ResponseWriter, r *http.Request) {
	headerNavComponent, ok := tplComponents["header_nav"]
	if !ok {
		fmt.Fprint(w, "Could not find component: header_nav")
		return
	}

	navigationPage, ok := tplPages["navigation"]
	if !ok {
		fmt.Fprint(w, "Could not find component: navigation")
		return
	}

	baseLayout, ok := tplLayouts["base"]
	if !ok {
		fmt.Fprint(w, "Could not find layout: base")
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
