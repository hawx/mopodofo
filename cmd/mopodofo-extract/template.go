package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template/parse"

	"hawx.me/code/mopodofo/internal/pot"
)

type argDef struct {
	// ctxt extracts the first argument as msgctxt and the second as msgid
	ctxt bool
	// count expects the argument after msgid to be a count
	count bool
	// extra checks for pairs of args, extracting the first of each pair to a
	// comment
	extra bool
}

func readTemplate(path string) ([]pot.Entry, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	treeSet := map[string]*parse.Tree{}

	tree := parse.New(filepath.Base(path), nil)
	tree.Mode = parse.SkipFuncCheck

	if _, err := tree.Parse(string(file), "{{", "}}", treeSet); err != nil {
		return nil, err
	}

	lineEnds := bytes.IndexByte(file, '\n') + 1
	currLine := 1

	fns := map[string]argDef{
		"tr":    {},
		"trs":   {count: true},
		"trf":   {extra: true},
		"trsf":  {count: true, extra: true},
		"trc":   {ctxt: true},
		"trcs":  {ctxt: true, count: true},
		"trcf":  {ctxt: true, extra: true},
		"trcsf": {ctxt: true, count: true, extra: true},
	}

	var entries []pot.Entry
	for _, tree := range treeSet {
		visitTemplate(tree.Root, func(node parse.Node) bool {
			switch node.Type() {
			case parse.NodeCommand:
				cnode := node.(*parse.CommandNode)
				args := cnode.Args

				inode, ok := args[0].(*parse.IdentifierNode)
				if !ok {
					return true
				}
				args = args[1:]

				def, ok := fns[inode.Ident]
				if !ok {
					return true
				}

				entry := pot.Entry{}

				if def.ctxt {
					ctxt, err := strconv.Unquote(args[0].String())
					if err != nil {
						return true
					}
					args = args[1:]
					entry.MsgCtxt = ctxt
				}

				id, err := strconv.Unquote(args[0].String())
				if err != nil {
					return true
				}
				args = args[1:]
				entry.MsgID = id

				if def.count {
					args = args[1:]
					entry.IsPlural = true
				}

				if def.extra {
					var extra []string
					for i := 0; i < len(args); i += 2 {
						arg, err := strconv.Unquote(args[i].String())
						if err != nil {
							return true
						}

						extra = append(extra, arg)
					}

					entry.ExtractedComments = append(entry.ExtractedComments, "Personalisation available: "+strings.Join(extra, ", "))
				}

				for {
					if int(cnode.Pos) < lineEnds || lineEnds > len(file) {
						break
					}

					i := bytes.IndexByte(file[lineEnds:], '\n')
					if i == -1 {
						break
					}

					lineEnds += i + 1
					currLine++
				}

				entry.Reference = []string{fmt.Sprintf("%s:%d", path, currLine)}
				vprint("%s found %s", entry.Reference[0], entry.MsgID)

				entries = append(entries, entry)
				return false

			}
			return true
		})
	}

	return entries, nil
}

func visitTemplate(node parse.Node, fn func(parse.Node) bool) {
	if node == nil || !fn(node) {
		return
	}

	switch node.Type() {
	case parse.NodeAction:
		anode := node.(*parse.ActionNode)
		visitTemplate(anode.Pipe, fn)

	case parse.NodeChain:
		cnode := node.(*parse.ChainNode)
		visitTemplate(cnode.Node, fn)

	case parse.NodeCommand:
		cnode := node.(*parse.CommandNode)
		for _, n := range cnode.Args {
			visitTemplate(n, fn)
		}

	case parse.NodeList:
		lnode := node.(*parse.ListNode)
		for _, n := range lnode.Nodes {
			visitTemplate(n, fn)
		}

	case parse.NodePipe:
		pnode := node.(*parse.PipeNode)
		for _, n := range pnode.Cmds {
			visitTemplate(n, fn)
		}
		for _, n := range pnode.Decl {
			visitTemplate(n, fn)
		}

	case parse.NodeIf:
		bnode := node.(*parse.IfNode)
		visitTemplate(bnode.Pipe, fn)
		visitTemplate(bnode.List, fn)
		if bnode.ElseList != nil {
			visitTemplate(bnode.ElseList, fn)
		}

	case parse.NodeWith:
		bnode := node.(*parse.WithNode)
		visitTemplate(bnode.Pipe, fn)
		visitTemplate(bnode.List, fn)
		if bnode.ElseList != nil {
			visitTemplate(bnode.ElseList, fn)
		}

	case parse.NodeTemplate:
		tnode := node.(*parse.TemplateNode)
		visitTemplate(tnode.Pipe, fn)

	case parse.NodeRange:
		bnode := node.(*parse.RangeNode)
		visitTemplate(bnode.Pipe, fn)
		visitTemplate(bnode.List, fn)
		if bnode.ElseList != nil {
			visitTemplate(bnode.ElseList, fn)
		}
	}
}
