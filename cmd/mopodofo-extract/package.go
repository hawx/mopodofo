package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
	"hawx.me/code/mopodofo/internal/pot"
)

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
