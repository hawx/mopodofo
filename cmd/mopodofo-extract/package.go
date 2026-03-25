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
			fileEntries, err := processTree(wd, pkg.Fset, file, pkg.TypesInfo)
			if err != nil {
				return nil, err
			}

			entries = append(entries, fileEntries...)
		}
	}

	return entries, nil
}

func processTree(wd string, fset *token.FileSet, tree *ast.File, typeinfo *types.Info) ([]pot.Entry, error) {
	fns := map[string]argDef{
		"Tr":    {},
		"Trs":   {count: true},
		"Trf":   {extra: true},
		"Trsf":  {count: true, extra: true},
		"Trc":   {ctxt: true},
		"Trcs":  {ctxt: true, count: true},
		"Trcf":  {ctxt: true, extra: true},
		"Trcsf": {ctxt: true, count: true, extra: true},
	}

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

		def, ok := fns[ts.Obj().Name()]
		if !ok {
			continue
		}

		args := expr.Args
		entry := pot.Entry{}
		var personalisation []string

		if def.ctxt {
			lit, ok := args[0].(*ast.BasicLit)
			if !ok {
				// do I try to follow constants, or even variables?!?!?
				continue // blow up?!
			}
			args = args[1:]
			entry.MsgCtxt = lit.Value[1 : len(lit.Value)-1]
		}

		{
			lit, ok := args[0].(*ast.BasicLit)
			if !ok {
				// do I try to follow constants, or even variables?!?!?
				continue // blow up?!
			}
			args = args[1:]
			entry.MsgID = lit.Value[1 : len(lit.Value)-1]
		}

		if def.count {
			args = args[1:]
			entry.IsPlural = true
			personalisation = append(personalisation, "Count")
		}

		if def.extra {
			for i := 0; i < len(args); i += 2 {
				lit, ok := args[i].(*ast.BasicLit)
				if !ok {
					continue
				}

				personalisation = append(personalisation, lit.Value[1:len(lit.Value)-1])
			}

			entry.ExtractedComments = append(entry.ExtractedComments, "Personalisation available: "+strings.Join(personalisation, ", "))
		}

		entry.Reference = []string{referenceForExpr(wd, fset, expr)}
		if entry.MsgCtxt != "" {
			vprint("%s found [%s] %s", entry.Reference[0], entry.MsgCtxt, entry.MsgID)
		} else {
			vprint("%s found %s", entry.Reference[0], entry.MsgID)
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func referenceForExpr(wd string, fset *token.FileSet, expr *ast.CallExpr) string {
	file := fset.File(expr.Pos()).Name()
	if rel, err := filepath.Rel(wd, file); err == nil {
		file = rel
	}

	return fmt.Sprintf("%s:%d", file, fset.Position(expr.Pos()).Line)
}
