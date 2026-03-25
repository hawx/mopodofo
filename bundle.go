package mopodofo

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"hawx.me/code/mopodofo/internal/mo"
)

// Bundles holds bundles for all the available languages.
type Bundles struct {
	langs map[string]*Bundle
}

func ReadAll(dir string) (*Bundles, error) {
	glob, err := filepath.Glob(filepath.Join(dir, "*", "LC_MESSAGES", "*.mo"))
	if err != nil {
		return nil, err
	}

	bundles := &Bundles{langs: map[string]*Bundle{}}
	for _, path := range glob {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		data, err := mo.Decode(file)
		if err != nil {
			return nil, err
		}

		bundles.langs[data.Metadata.Language] = &Bundle{file: data}
	}

	return bundles, nil
}

func (b *Bundles) For(lang string) (*Bundle, bool) {
	bundle, ok := b.langs[lang]
	return bundle, ok
}

func (b *Bundles) Must(lang string) *Bundle {
	bundle, ok := b.langs[lang]
	if !ok {
		panic(fmt.Sprintf("bundle for %s not found", lang))
	}
	return bundle
}

// A Bundle holds the translations for a particular language.
type Bundle struct {
	file *mo.File
}

// Tr returns the singular translation for the given key.
func (b *Bundle) Tr(key string) string {
	return b.file.Singles[key]
}

// Trs returns the plural translation for the given key and count.
func (b *Bundle) Trs(key string, count int) string {
	return b.file.Plurals[key][b.file.Metadata.Plurals(count)]
}

// Trf returns the singular translation for the given key, formatted with args.
func (b *Bundle) Trf(key string, args ...string) string {
	tmpl := template.Must(template.New("").Parse(b.Tr(key)))

	argmap := map[string]string{}
	for i := 0; i < len(args); i += 2 {
		argmap[args[i]] = args[i+1]
	}

	var buf bytes.Buffer
	tmpl.Execute(&buf, argmap)

	return buf.String()
}

// Trsf returns the plural translation for the given key and count, formatted with args.
func (b *Bundle) Trsf(key string, count int, args ...string) string {
	tmpl := template.Must(template.New("").Parse(b.Trs(key, count)))

	argmap := map[string]string{}
	for i := 0; i < len(args); i += 2 {
		argmap[args[i]] = args[i+1]
	}

	var buf bytes.Buffer
	tmpl.Execute(&buf, argmap)

	return buf.String()
}

// Trc returns the singular translation for the given key in context.
func (b *Bundle) Trc(context, key string) string {
	return b.file.ContextSingles[context][key]
}

// Trs returns the plural translation for the given key in context and count.
func (b *Bundle) Trcs(context, key string, count int) string {
	return b.file.ContextPlurals[context][key][b.file.Metadata.Plurals(count)]
}

// Trcf returns the singular translation for the given key in context, formatted with args.
func (b *Bundle) Trcf(context, key string, args ...string) string {
	tmpl := template.Must(template.New("").Parse(b.Trc(context, key)))

	argmap := map[string]string{}
	for i := 0; i < len(args); i += 2 {
		argmap[args[i]] = args[i+1]
	}

	var buf bytes.Buffer
	tmpl.Execute(&buf, argmap)

	return buf.String()
}

// Trcsf returns the plural translation for the given key in context and count, formatted with args.
func (b *Bundle) Trcsf(context, key string, count int, args ...string) string {
	tmpl := template.Must(template.New("").Parse(b.Trcs(context, key, count)))

	argmap := map[string]string{}
	for i := 0; i < len(args); i += 2 {
		argmap[args[i]] = args[i+1]
	}

	var buf bytes.Buffer
	tmpl.Execute(&buf, argmap)

	return buf.String()
}
