package main

import (
	"log"
	"net/http"

	"hawx.me/code/mopodofo"
	"hawx.me/code/mopodofo/example/app"
	"hawx.me/code/mopodofo/example/notes"
	"hawx.me/code/mopodofo/example/paths"
)

func main() {
	db := &notes.DB{}

	bundles, err := mopodofo.ReadAll("example/lang")
	if err != nil {
		panic(err)
	}

	en := bundles.Must("en")
	cy := bundles.Must("cy")

	mux := http.NewServeMux()
	app.Register(mux, db, en, cy)
	app.Register(mux, db, cy, en)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, paths.Home.String(en), http.StatusFound)
	})

	log.Println(
		en.Trf("runningOn", "Port", "8080"),
		"/",
		cy.Trf("runningOn", "Port", "8080"),
	)
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Println(err)
	}
}
