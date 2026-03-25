package app

import (
	"html/template"
	"net/http"
	"path/filepath"

	"hawx.me/code/mopodofo"
	"hawx.me/code/mopodofo/example/notes"
	"hawx.me/code/mopodofo/example/paths"
)

func Register(mux *http.ServeMux, db *notes.DB, bundle, obundle *mopodofo.Bundle) {
	funcMap := bundle.FuncMap()
	funcMap["otrc"] = obundle.FuncMap()["trc"] // so that the page switcher can work

	tmpls := parseTemplates("example/template", funcMap)
	paths := paths.For(bundle)

	mux.HandleFunc(paths.Home,
		home(tmpls["home.gohtml"], db))
	mux.HandleFunc(paths.Create,
		create(tmpls["create.gohtml"], db, paths))
}

func parseTemplates(dir string, funcMap template.FuncMap) map[string]*template.Template {
	layout := template.Must(template.New("").
		Funcs(funcMap).
		ParseGlob(filepath.Join(dir, "layout/*.gohtml")))

	files, _ := filepath.Glob(filepath.Join(dir, "*.gohtml"))

	tmpls := map[string]*template.Template{}
	for _, file := range files {
		clone := template.Must(layout.Clone())
		tmpl := template.Must(clone.ParseFiles(file))
		tmpls[filepath.Base(file)] = tmpl
	}

	return tmpls
}
