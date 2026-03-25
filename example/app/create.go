package app

import (
	"html/template"
	"log"
	"net/http"

	"hawx.me/code/mopodofo/example/notes"
	"hawx.me/code/mopodofo/example/paths"
)

func create(tmpl *template.Template, db *notes.DB, paths paths.Paths) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			id := db.Create(r.FormValue("content"))
			log.Println("created note:", id)
			http.Redirect(w, r, paths.Home, http.StatusFound)
			return
		}

		if err := tmpl.ExecuteTemplate(w, "layout", struct {
			PathName string
		}{
			PathName: "create",
		}); err != nil {
			log.Println("execute create.gohtml:", err)
		}
	}
}
