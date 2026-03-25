package main

import (
	"html/template"
	"os"

	"hawx.me/code/mopodofo"
	"hawx.me/code/mopodofo/example/pkg"
)

type Data struct {
	Messages []Message
}

type Message struct {
	Person, Content string
}

func main() {
	bundles, err := mopodofo.ReadAll("lang")
	if err != nil {
		panic(err)
	}

	en := bundles.Must("en")
	cy := bundles.Must("cy")

	pkg.MyFunc(en)
	pkg.MyFunc(cy)

	tmpl := template.Must(template.New("template.gohtml").Funcs(map[string]any{
		"tr":    en.Tr,
		"trs":   en.Trs,
		"trf":   en.Trf,
		"trsf":  en.Trsf,
		"trc":   en.Trc,
		"trcs":  en.Trcs,
		"trcf":  en.Trcf,
		"trcsf": en.Trcsf,
	}).ParseFiles("example/template/template.gohtml"))
	if err := tmpl.Execute(os.Stdout, Data{
		Messages: []Message{
			{Person: "John", Content: "How are you"},
			{Person: "Barry", Content: "Good"},
		},
	}); err != nil {
		panic(err)
	}
}
