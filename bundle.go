package mopodofo

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"hawx.me/code/mopodofo/internal/mo"
)

var NilBundle = &Bundle{}

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

func (b *Bundles) For(lang string) *Bundle {
	bundle, ok := b.langs[lang]
	if !ok {
		return NilBundle
	}

	return bundle
}

// A Bundle holds the translations for a particular language.
type Bundle struct {
	file  *mo.File
	funcs template.FuncMap
}

func (b *Bundle) Funcs(funcs template.FuncMap) *Bundle {
	b.funcs = funcs
	return b
}

// Tr returns the singular translation for the given key.
func (b *Bundle) Tr(key string) string {
	return b.file.Singles[key]
}

// Trs returns the plural translation for the given key and count.
func (b *Bundle) Trs(key string, count int) string {
	s := b.file.Plurals[key][b.file.Metadata.Plurals(count)]

	return b.format(s, []any{"Count", count})
}

// Trf returns the singular translation for the given key, formatted with args.
func (b *Bundle) Trf(key string, args ...any) string {
	return b.format(b.Tr(key), args)
}

// Trsf returns the plural translation for the given key and count, formatted with args.
func (b *Bundle) Trsf(key string, count int, args ...any) string {
	s := b.file.Plurals[key][b.file.Metadata.Plurals(count)]

	return b.format(s, append(args, "Count", count))
}

// Trc returns the singular translation for the given key in context.
func (b *Bundle) Trc(context, key string) string {
	return b.file.Singles[context+"\x04"+key]
}

// Trs returns the plural translation for the given key in context and count.
func (b *Bundle) Trcs(context, key string, count int) string {
	s := b.file.Plurals[context+"\x04"+key][b.file.Metadata.Plurals(count)]

	return b.format(s, []any{"Count", count})
}

// Trcf returns the singular translation for the given key in context, formatted with args.
func (b *Bundle) Trcf(context, key string, args ...any) string {
	return b.format(b.Trc(context, key), args)
}

// Trcsf returns the plural translation for the given key in context and count, formatted with args.
func (b *Bundle) Trcsf(context, key string, count int, args ...any) string {
	s := b.file.Plurals[context+"\x04"+key][b.file.Metadata.Plurals(count)]

	return b.format(s, append(args, "Count", count))
}

func (b *Bundle) format(s string, args []any) string {
	tmpl := template.Must(template.New("").Funcs(b.funcs).Parse(s))

	argmap := map[string]any{}
	for i := 0; i < len(args); i += 2 {
		argmap[args[i].(string)] = args[i+1]
	}

	var buf bytes.Buffer
	tmpl.Execute(&buf, argmap)

	return buf.String()
}

// StealContext allows a Bundle to take strings from another Bundle for a
// particular context, but changing them to another context.
//
// My particular use-case for this is to allow each bundle to contain URLs. But
// each language will need the others to allow a language switcher to go between
// the pages. I could pass both bundles in, but it seems like introducing this
// concept will be nicer.
func (b *Bundle) StealContext(other *Bundle, fromctxt, toctxt string) *Bundle {
	for single, value := range other.file.Singles {
		if ctxt, key, ok := strings.Cut(single, "\x04"); ok && ctxt == fromctxt {
			b.file.Singles[toctxt+"\x04"+key] = value
		}
	}

	for plural, value := range other.file.Plurals {
		if ctxt, key, ok := strings.Cut(plural, "\x04"); ok && ctxt == fromctxt {
			b.file.Plurals[toctxt+"\x04"+key] = value
		}
	}

	return b
}

func (b *Bundle) FuncMap() map[string]any {
	return map[string]any{
		"tr":    b.Tr,
		"trs":   b.Trs,
		"trf":   b.Trf,
		"trsf":  b.Trsf,
		"trc":   b.Trc,
		"trcs":  b.Trcs,
		"trcf":  b.Trcf,
		"trcsf": b.Trcsf,
	}
}
