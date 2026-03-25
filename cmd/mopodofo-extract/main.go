package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template/parse"

	"golang.org/x/tools/go/packages"
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

	for _, glob := range tmpls {
		paths, err := filepath.Glob(glob)
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

func readPackages(patterns ...string) ([]pot.Entry, error) {
	cfg := &packages.Config{Mode: packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes |
		packages.NeedTypesInfo | packages.NeedImports}
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return nil, err
	}

	var entries []pot.Entry
	for _, pkg := range pkgs {
		hasImported := false
		for _, import_ := range pkg.Imports {
			if import_.ID == "hawx.me/code/mopodofo" {
				hasImported = true
			}
		}

		if !hasImported {
			continue
		}

		// for id, obj := range pkg.TypesInfo.Uses {
		//	if obj.Pkg() != nil && obj.Pkg().Path() == "hawx.me/code/mopodofo" {
		//		if slices.Contains([]string{"Tr", "Trs", "Trf", "Trsf", "Trc", "Trcs", "Trcf", "Trcsf"}, id.Name) {
		//			fmt.Printf("%s: %q uses %s of %v\n", pkg.Fset.Position(id.Pos()), id.Name, obj.Name(), obj.Pkg())
		//		}
		//	}
		// }

		for _, file := range pkg.Syntax {
			wd, _ := os.Getwd()
			path, _ := filepath.Rel(wd, pkg.Dir+"/"+file.Name.Name+".go")
			fileEntries, err := processTree(path, pkg.Fset, file, pkg.TypesInfo)
			if err != nil {
				return nil, err
			}

			entries = append(entries, fileEntries...)
		}
	}

	return entries, nil
}

func processTree(path string, fset *token.FileSet, tree *ast.File, typeinfo *types.Info) ([]pot.Entry, error) {
	var entries []pot.Entry

	for node := range ast.Preorder(tree) {
		expr, ok := node.(*ast.CallExpr)
		if !ok {
			continue
		}

		selector, ok := expr.Fun.(*ast.SelectorExpr)
		if !ok {
			continue
		}

		ts, ok := typeinfo.Selections[selector]
		if !ok || ts.Kind() != types.MethodVal || ts.Recv().String() != "*hawx.me/code/mopodofo.Bundle" {
			continue
		}

		name := ts.Obj().Name()

		switch name {
		case "Tr":
			keyLit, ok := expr.Args[0].(*ast.BasicLit)
			if !ok {
				// do I try to follow constants, or even variables?!?!?
				continue // blow up?!
			}

			key := keyLit.Value[1 : len(keyLit.Value)-1]
			reference := fmt.Sprintf("%s:%d", path, fset.Position(expr.Pos()).Line)

			vprint("%s found %s", reference, key)
			entries = append(entries, pot.Entry{
				Reference: []string{reference},
				MsgID:     key,
			})

		case "Trs":
			keyLit, ok := expr.Args[0].(*ast.BasicLit)
			if !ok {
				// do I try to follow constants, or even variables?!?!?
				continue // blow up?!
			}

			key := keyLit.Value[1 : len(keyLit.Value)-1]
			reference := fmt.Sprintf("%s:%d", path, fset.Position(expr.Pos()).Line)

			vprint("%s found %s", reference, key)
			entries = append(entries, pot.Entry{
				Reference: []string{reference},
				MsgID:     key,
				IsPlural:  true,
			})

		case "Trf":
			keyLit, ok := expr.Args[0].(*ast.BasicLit)
			if !ok {
				// do I try to follow constants, or even variables?!?!?
				continue // blow up?!
			}

			start := 1
			if selector.Sel.Name == "Trsf" {
				start = 2
			}

			var args []string
			for i := start; i < len(expr.Args); i += 2 {
				argLit, ok := expr.Args[i].(*ast.BasicLit)
				if !ok {
					continue
				}

				args = append(args, argLit.Value[1:len(argLit.Value)-1])
			}

			key := keyLit.Value[1 : len(keyLit.Value)-1]
			reference := fmt.Sprintf("%s:%d", path, fset.Position(expr.Pos()).Line)

			vprint("%s found %s", reference, key)
			entries = append(entries, pot.Entry{
				Reference: []string{reference},
				MsgID:     key,
				ExtractedComments: []string{
					"Personalisation available: " + strings.Join(args, ", "),
				},
			})

		case "Trsf":
			keyLit, ok := expr.Args[0].(*ast.BasicLit)
			if !ok {
				// do I try to follow constants, or even variables?!?!?
				continue // blow up?!
			}

			start := 1
			if selector.Sel.Name == "Trsf" {
				start = 2
			}

			var args []string
			for i := start; i < len(expr.Args); i += 2 {
				argLit, ok := expr.Args[i].(*ast.BasicLit)
				if !ok {
					continue
				}

				args = append(args, argLit.Value[1:len(argLit.Value)-1])
			}

			key := keyLit.Value[1 : len(keyLit.Value)-1]
			reference := fmt.Sprintf("%s:%d", path, fset.Position(expr.Pos()).Line)

			vprint("%s found %s", reference, key)
			entries = append(entries, pot.Entry{
				Reference: []string{reference},
				MsgID:     key,
				IsPlural:  true,
				ExtractedComments: []string{
					"Personalisation available: " + strings.Join(args, ", "),
				},
			})

		case "Trc", "Trcs":
			ctxtLit, ok := expr.Args[0].(*ast.BasicLit)
			if !ok {
				// do I try to follow constants, or even variables?!?!?
				continue // blow up?!
			}
			keyLit, ok := expr.Args[1].(*ast.BasicLit)
			if !ok {
				// do I try to follow constants, or even variables?!?!?
				continue // blow up?!
			}

			ctxt := ctxtLit.Value[1 : len(ctxtLit.Value)-1]
			key := keyLit.Value[1 : len(keyLit.Value)-1]
			reference := fmt.Sprintf("%s:%d", path, fset.Position(expr.Pos()).Line)

			vprint("%s found [%s] %s", reference, ctxt, key)
			entries = append(entries, pot.Entry{
				Reference: []string{reference},
				MsgCtxt:   ctxt,
				MsgID:     key,
			})

		case "Trcf", "Trcsf":
			ctxtLit, ok := expr.Args[0].(*ast.BasicLit)
			if !ok {
				// do I try to follow constants, or even variables?!?!?
				continue // blow up?!
			}
			keyLit, ok := expr.Args[1].(*ast.BasicLit)
			if !ok {
				// do I try to follow constants, or even variables?!?!?
				continue // blow up?!
			}

			start := 2
			if selector.Sel.Name == "Trcsf" {
				start = 3
			}

			var args []string
			for i := start; i < len(expr.Args); i += 2 {
				argLit, ok := expr.Args[i].(*ast.BasicLit)
				if !ok {
					continue
				}

				args = append(args, argLit.Value[1:len(argLit.Value)-1])
			}

			ctxt := ctxtLit.Value[1 : len(ctxtLit.Value)-1]
			key := keyLit.Value[1 : len(keyLit.Value)-1]
			reference := fmt.Sprintf("%s:%d", path, fset.Position(expr.Pos()).Line)

			vprint("%s found [%s] %s", reference, ctxt, key)
			entries = append(entries, pot.Entry{
				Reference: []string{reference},
				MsgCtxt:   ctxt,
				MsgID:     key,
				ExtractedComments: []string{
					"Personalisation available: " + strings.Join(args, ", "),
				},
			})
		}
	}

	return entries, nil
}

func readTemplate(path string) ([]pot.Entry, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	treeSet := map[string]*parse.Tree{}

	tree := parse.New(filepath.Base(path), nil)
	tree.Mode = parse.SkipFuncCheck

	tree, err = tree.Parse(string(file), "{{", "}}", treeSet)
	if err != nil {
		return nil, err
	}

	entries := visitTemplate(path, tree.Root.Nodes)

	return entries, nil
}

func visitTemplate(path string, nodes []parse.Node) []pot.Entry {
	var entries []pot.Entry
	for _, node := range nodes {
		switch node.Type() {
		case parse.NodeAction:
			anode := node.(*parse.ActionNode)
			if entry, ok := readTemplateAction(path, anode); ok {
				entries = append(entries, entry)
			}

		case parse.NodeRange:
			rnode := node.(*parse.RangeNode)
			found := visitTemplate(path, rnode.List.Nodes)
			entries = append(entries, found...)

		default:
			// fmt.Println("skip", node)

		}
	}

	return entries
}

func readTemplateAction(path string, anode *parse.ActionNode) (pot.Entry, bool) {
	for _, cmd := range anode.Pipe.Cmds {
		switch cmd.Args[0].String() {
		case "tr":
			if len(cmd.Args) < 2 {
				continue
			}

			key, err := strconv.Unquote(cmd.Args[1].String())
			if err != nil {
				return pot.Entry{}, false
			}

			reference := fmt.Sprintf("%s:%d", path, cmd.Pos)

			vprint("%s found %s", reference, key)
			return pot.Entry{
				Reference: []string{reference},
				MsgID:     key,
			}, true

		case "trs":
			if len(cmd.Args) < 3 {
				continue
			}

			key, err := strconv.Unquote(cmd.Args[1].String())
			if err != nil {
				return pot.Entry{}, false
			}

			// skip int

			reference := fmt.Sprintf("%s:%d", path, cmd.Pos)

			vprint("%s found %s", reference, key)
			return pot.Entry{
				Reference: []string{reference},
				MsgID:     key,
				IsPlural:  true,
			}, true

		case "trf":
			if len(cmd.Args) < 4 && len(cmd.Args)%2 != 0 {
				continue
			}

			key, err := strconv.Unquote(cmd.Args[1].String())
			if err != nil {
				return pot.Entry{}, false
			}

			var args []string
			for i := 2; i < len(cmd.Args); i += 2 {
				arg, err := strconv.Unquote(cmd.Args[i].String())
				if err != nil {
					return pot.Entry{}, false
				}

				args = append(args, arg)
			}

			reference := fmt.Sprintf("%s:%d", path, cmd.Pos)

			vprint("%s found %s", reference, key)
			return pot.Entry{
				Reference: []string{reference},
				MsgID:     key,
				ExtractedComments: []string{
					"Personalisation available: " + strings.Join(args, ", "),
				},
			}, true

		case "trsf":
			if len(cmd.Args) < 5 && len(cmd.Args)%2 != 1 {
				continue
			}

			key, err := strconv.Unquote(cmd.Args[1].String())
			if err != nil {
				return pot.Entry{}, false
			}

			// skip int

			var args []string
			for i := 3; i < len(cmd.Args); i += 2 {
				arg, err := strconv.Unquote(cmd.Args[i].String())
				if err != nil {
					return pot.Entry{}, false
				}

				args = append(args, arg)
			}

			reference := fmt.Sprintf("%s:%d", path, cmd.Pos)

			vprint("%s found %s", reference, key)
			return pot.Entry{
				Reference: []string{reference},
				MsgID:     key,
				ExtractedComments: []string{
					"Personalisation available: " + strings.Join(args, ", "),
				},
			}, true

		case "trc":
			if len(cmd.Args) < 3 {
				continue
			}

			context, err := strconv.Unquote(cmd.Args[1].String())
			if err != nil {
				return pot.Entry{}, false
			}

			key, err := strconv.Unquote(cmd.Args[2].String())
			if err != nil {
				return pot.Entry{}, false
			}

			reference := fmt.Sprintf("%s:%d", path, cmd.Pos)

			vprint("%s found [%s] %s", reference, context, key)
			return pot.Entry{
				Reference: []string{reference},
				MsgCtxt:   context,
				MsgID:     key,
			}, true

		case "trcs":
			if len(cmd.Args) < 4 {
				continue
			}

			context, err := strconv.Unquote(cmd.Args[1].String())
			if err != nil {
				return pot.Entry{}, false
			}

			key, err := strconv.Unquote(cmd.Args[2].String())
			if err != nil {
				return pot.Entry{}, false
			}

			// skip int

			reference := fmt.Sprintf("%s:%d", path, cmd.Pos)

			vprint("%s found [%s] %s", reference, context, key)
			return pot.Entry{
				Reference: []string{reference},
				MsgCtxt:   context,
				MsgID:     key,
				IsPlural:  true,
			}, true

		case "trcf":
			if len(cmd.Args) < 5 && len(cmd.Args)%2 != 1 {
				continue
			}

			context, err := strconv.Unquote(cmd.Args[1].String())
			if err != nil {
				return pot.Entry{}, false
			}

			key, err := strconv.Unquote(cmd.Args[2].String())
			if err != nil {
				return pot.Entry{}, false
			}

			var args []string
			for i := 3; i < len(cmd.Args); i += 2 {
				arg, err := strconv.Unquote(cmd.Args[i].String())
				if err != nil {
					return pot.Entry{}, false
				}

				args = append(args, arg)
			}

			reference := fmt.Sprintf("%s:%d", path, cmd.Pos)

			vprint("%s found [%s] %s", reference, context, key)
			return pot.Entry{
				Reference: []string{reference},
				MsgCtxt:   context,
				MsgID:     key,
				ExtractedComments: []string{
					"Personalisation available: " + strings.Join(args, ", "),
				},
			}, true

		case "trcsf":
			if len(cmd.Args) < 6 && len(cmd.Args)%2 != 0 {
				continue
			}

			context, err := strconv.Unquote(cmd.Args[1].String())
			if err != nil {
				return pot.Entry{}, false
			}

			key, err := strconv.Unquote(cmd.Args[2].String())
			if err != nil {
				return pot.Entry{}, false
			}

			// skip int

			var args []string
			for i := 4; i < len(cmd.Args); i += 2 {
				arg, err := strconv.Unquote(cmd.Args[i].String())
				if err != nil {
					return pot.Entry{}, false
				}

				args = append(args, arg)
			}

			reference := fmt.Sprintf("%s:%d", path, cmd.Pos)

			vprint("%s found [%s] %s", reference, context, key)
			return pot.Entry{
				Reference: []string{reference},
				MsgCtxt:   context,
				MsgID:     key,
				ExtractedComments: []string{
					"Personalisation available: " + strings.Join(args, ", "),
				},
			}, true

		}
	}

	return pot.Entry{}, false
}
