package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"hawx.me/code/mopodofo/internal/pot"
)

var vprint = func(string, ...any) {}

func main() {
	var (
		verbose     = flag.Bool("v", false, "Show verbose output")
		out         = flag.String("out", "lang/messages.pot", "File to write output")
		pkgs, tmpls []string
	)
	flag.Func("pkg", "Package pattern to extract from", func(s string) error {
		pkgs = append(pkgs, s)
		return nil
	})
	flag.Func("tmpl", "Template glob to extract from", func(s string) error {
		tmpls = append(tmpls, s)
		return nil
	})
	flag.Usage = func() {
		fmt.Fprint(os.Stdout, `Usage: mopodofo-extract OPTIONS...

This tool extracts translatable content from .go source files and Go
text/template files into a .pot format file.

Options:
	-pkg PATTERN    Parse .go files found by the PATTERN
	-tmpl GLOB      Parse text/template files found by GLOB
	-out PATH       Write the .pot to PATH (default: lang/messages.pot)
	-v              Print extra information when run
`)
	}
	flag.Parse()

	if *verbose {
		vprint = func(msg string, args ...any) {
			fmt.Printf(msg+"\n", args...)
		}
	}

	if err := run(*out, pkgs, tmpls); err != nil {
		fmt.Println(err)
	}
}

func run(out string, pkgs, tmpls []string) error {
	if len(pkgs) == 0 && len(tmpls) == 0 {
		flag.Usage()
		return nil
	}

	file := pot.File{}

	{
		entries, err := readPackages(pkgs...)
		if err != nil {
			return err
		}
		file.Entries = append(file.Entries, entries...)
	}

	for _, pattern := range tmpls {
		paths, err := glob(pattern)
		if err != nil {
			return err
		}

		for _, path := range paths {
			vprint("%s: reading template file", path)
			entries, err := readTemplate(path)
			if err != nil {
				return err
			}
			file.Entries = append(file.Entries, entries...)
		}
	}

	file.Entries = pot.Merge(file.Entries)

	var buf bytes.Buffer
	if err := pot.Encode(&buf, file); err != nil {
		return err
	}

	vprint("writing %s", out)
	return os.WriteFile(out, buf.Bytes(), 0600)
}

// glob is a poor extension to filepath.Glob that allows a single use of '/**/'
// to mean every directory. Best to assume it is not going to be correct for
// every situation, but will work for the usual simple cases...
func glob(pattern string) ([]string, error) {
	if pre, post, ok := strings.Cut(pattern, "/**/"); ok {
		var results []string

		filepath.WalkDir(pre, func(path string, d fs.DirEntry, err error) error {
			if depth := strings.Count(post, string(filepath.Separator)); depth > 0 {
				parts := strings.Split(path, string(filepath.Separator))
				start := len(parts) - depth - 1

				match, err := filepath.Match(post, filepath.Join(parts[start:]...))
				if err != nil {
					return err
				}

				if match {
					results = append(results, path)
				}

			} else {
				match, err := filepath.Match(post, filepath.Base(path))
				if err != nil {
					return err
				}

				if match {
					results = append(results, path)
				}
			}

			return nil
		})

		return results, nil
	}

	return filepath.Glob(pattern)
}
