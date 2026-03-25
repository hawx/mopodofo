# mopodofo

> This is proof-of-concept level, probably don't actually use it yet...

An i18n tool and package for Go. This extracts keys from your code and templates
to a `.pot` file, then provides a way to use the translated `.mo` files in your
code and templates.

It uses a key based approach to specifying content to be translated. That is you
don't specify the English, then translate that, instead specify meaningful
"keys" which can then be translated.

This approach allows a non-technical person to edit the English strings too.

In templates, use the helpers:

```html
<p>{{ tr "welcomeMessage" }}</p>
```

Or in Go files, mark translatable keys:

```go
func myFunc(ok bool, w io.Writer, bundle *mopodofo.Bundle) {
  s := "helloMessage" //mopodofo
  if !ok {
    s = "goodbyeMessage" //mopodofo
  }

  io.WriteString(w, bundle.Tr(s))
}
```

To extract the strings:

```commandline
$ go install hawx.me/code/mopodofo/cmd/mopodofo-extract
$ mopodofo-extract -v -out lang/messages.pot -tmpl template/*.gohtml -pkg .
template/welcome.gohtml: found welcomeMessage
main.go: found helloMessage
main.go: found goodbyeMessage
wrote lang/messages.pot
```

Then use `msgmerge` to get `.po` files to translate and convert those to `.mo`
with `msgfmt` as per the usual gettext workflow.

Finally, read the messages and use them:

```go
func main() {
    bundles, err := mopodofo.ReadAll("./lang")
    if err != nil {
        panic(err)
    }

    locale := os.Getenv("LANGUAGE")

    myFunc(true, os.Stdout, bundles.For(locale))

    // and template example
}

```

---

TODO:

- follow constants used as arguments?
- maybe follow variables, although I think this might just lead to pain
- follow the types so they could be wrapped? `func (x *X) gettext(s string)
  string { return x.b.Tr(s) }`
- associate other data with translations, I really want to have the URL a string
  appears on. Will require parsing *http.ServeMux type registrations and working
  out templates!?!>?!
