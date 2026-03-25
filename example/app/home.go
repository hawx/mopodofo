package app

import (
	"html/template"
	"log"
	"net/http"

	"hawx.me/code/mopodofo/example/notes"
)

func home(tmpl *template.Template, db *notes.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := tmpl.ExecuteTemplate(w, "layout", struct {
			PathName string
			Notes    *notes.DB
		}{
			PathName: "home",
			Notes:    db,
		}); err != nil {
			log.Println("execute home.gohtml:", err)
		}
	}
}
