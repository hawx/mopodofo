package mo

import (
	"fmt"
	"strconv"
	"strings"
)

// parseFormula turns a simple formula on the variable n into a function that
// takes an integer value. Operators allowed are:
//
//	==, !=, >, <, >=, <=, &&, ||, %, ? :
//
// Along with parentheses and whitespace.
func parseFormula(formula string) (func(int) int, error) {
	tokens, err := tokenise(formula)
	if err != nil {
		return nil, err
	}

	node, err := parse(&tokens, 0)
	if err != nil {
		return nil, err
	}

	return func(n int) int {
		return node.eval(n)
	}, nil
}

type Tokens []Token

func (ts *Tokens) Next() Token {
	if len(*ts) == 0 {
		return EOF
	}

	t := (*ts)[0]
	*ts = (*ts)[1:]
	return t
}

func (ts *Tokens) Peek() Token {
	if len(*ts) == 0 {
		return EOF
	}

	return (*ts)[0]
}

func (ts Tokens) String() string {
	var s strings.Builder
	for _, t := range ts {
		s.WriteString(t.String())
		if t != EOF {
			s.WriteString(" ")
		}
	}
	return s.String()
}

type Token int

const (
	// TODO: maybe verify when parsing ints that they don't stray into this range,
	// also maybe move these to the very bottom of int?

	N Token = iota - 200000 // let's assume nobody is putting -200,000 in their formulas
	Eq
	Neq
	Gt
	Lt
	Gte
	Lte
	And
	Or
	Mod
	If
	Else
	Open
	Close
	EOF
)

func (t Token) IsOp() bool {
	return t > N
}

func (t Token) IsInt() bool {
	return t > -200000
}

func (t Token) String() string {
	switch t {
	case N:
		return "n"
	case Eq:
		return "=="
	case Neq:
		return "!="
	case Gt:
		return ">"
	case Lt:
		return "<"
	case Gte:
		return ">="
	case Lte:
		return "<="
	case And:
		return "&&"
	case Or:
		return "||"
	case Mod:
		return "%"
	case If:
		return "?"
	case Else:
		return ":"
	case Open:
		return "("
	case Close:
		return ")"
	case EOF:
		return "<END>"
	default:
		return strconv.Itoa(int(t))
	}
}

func tokenise(s string) (Tokens, error) {
	var tokens []Token
	for i := 0; i < len(s); i++ {
		switch c := s[i]; c {
		case '\t', ' ':
			// Ignore whitespace

		case 'n':
			tokens = append(tokens, N)
		case '%':
			tokens = append(tokens, Mod)
		case '?':
			tokens = append(tokens, If)
		case ':':
			tokens = append(tokens, Else)
		case '(':
			tokens = append(tokens, Open)
		case ')':
			tokens = append(tokens, Close)

		case '=': // must be followed by =
			if s[i+1] == '=' {
				tokens = append(tokens, Eq)
				i++
			} else {
				return nil, fmt.Errorf("could not parse lone %c", c)
			}
		case '!':
			if s[i+1] == '=' {
				tokens = append(tokens, Neq)
				i++
			} else {
				return nil, fmt.Errorf("could not parse lone %c", c)
			}

		case '>':
			if s[i+1] == '=' {
				tokens = append(tokens, Gte)
				i++
			} else {
				tokens = append(tokens, Gt)
			}
		case '<': // may be followed by =
			if s[i+1] == '=' {
				tokens = append(tokens, Lte)
				i++
			} else {
				tokens = append(tokens, Lt)
			}

		case '&':
			if s[i+1] == c {
				tokens = append(tokens, And)
				i++
			} else {
				return nil, fmt.Errorf("could not parse lone %c", c)
			}
		case '|':
			if s[i+1] == c {
				tokens = append(tokens, Or)
				i++
			} else {
				return nil, fmt.Errorf("could not parse lone %c", c)
			}

			// maybe handle negatives, if needed???
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			start := i
			for i+1 < len(s) && s[i+1] >= '0' && s[i+1] <= '9' {
				i++
			}

			num, _ := strconv.Atoi(s[start : i+1])
			tokens = append(tokens, Token(num))
		}
	}

	return tokens, nil
}

type node interface {
	eval(x int) int
	String() string
}

type nNode struct{}

func (n nNode) String() string { return "N" }

func (n nNode) eval(x int) int { return x }

type intNode struct{ i int }

func (n intNode) String() string { return strconv.Itoa(n.i) }

func (n intNode) eval(_ int) int { return n.i }

type binNode struct {
	left, right node
	op          Token
}

func (n binNode) String() string {
	return "(" + n.op.String() + " " + n.left.String() + " " + n.right.String() + ")"
}

func (n binNode) eval(x int) int {
	l := n.left.eval(x)
	r := n.right.eval(x)

	switch n.op {
	case Eq:
		if l == r {
			return 1
		} else {
			return 0
		}
	case Neq:
		if l != r {
			return 1
		} else {
			return 0
		}
	case Gt:
		if l > r {
			return 1
		} else {
			return 0
		}
	case Lt:
		if l < r {
			return 1
		} else {
			return 0
		}
	case Gte:
		if l >= r {
			return 1
		} else {
			return 0
		}
	case Lte:
		if l <= r {
			return 1
		} else {
			return 0
		}
	case Mod:
		return l % r
	case And:
		if l > 0 && r > 0 {
			return 1
		} else {
			return 0
		}
	case Or:
		if l > 0 || r > 0 {
			return 1
		} else {
			return 0
		}
	}

	panic("unexpected binop: " + n.op.String())
}

type ternaryNode struct {
	cond, yes, no node
}

func (n ternaryNode) String() string {
	return "if " + n.cond.String() + " then " + n.yes.String() + " else " + n.no.String() + " endif"
}

func (n ternaryNode) eval(x int) int {
	if n.cond.eval(x) > 0 {
		return n.yes.eval(x)
	} else {
		return n.no.eval(x)
	}
}

func parse(tokens *Tokens, minBinding uint8) (node, error) {
	// with many thanks to https://matklad.github.io/2020/04/13/simple-but-powerful-pratt-parsing.html
	next := tokens.Next()

	var lhs node
	switch next {
	case N:
		lhs = nNode{}
	case Open:
		sub, err := parse(tokens, 0)
		if err != nil {
			return nil, err
		}
		if tokens.Next() != Close {
			return nil, fmt.Errorf("unbalanced parens")
		}
		lhs = sub
	default:
		if next.IsInt() {
			lhs = intNode{i: int(next)}
		} else {
			return nil, fmt.Errorf("bad token: %s", next)
		}
	}

	for {
		var op Token

		next = tokens.Peek()
		if next == EOF {
			break
		} else if next.IsOp() {
			op = next
		} else {
			return nil, fmt.Errorf("bad token: %s", next)
		}

		lBinding, rBinding, ok := infixBindingPower(op)
		if !ok || lBinding < minBinding {
			break
		}

		tokens.Next()

		if op == If {
			mhs, err := parse(tokens, 0)
			if err != nil {
				return nil, err
			}
			if tokens.Next() != Else {
				return nil, fmt.Errorf("ternary missing :")
			}
			rhs, err := parse(tokens, rBinding)
			if err != nil {
				return nil, err
			}

			lhs = ternaryNode{cond: lhs, yes: mhs, no: rhs}
		} else {
			rhs, err := parse(tokens, rBinding)
			if err != nil {
				return nil, err
			}

			lhs = binNode{left: lhs, op: op, right: rhs}
		}
	}

	return lhs, nil
}

func infixBindingPower(op Token) (uint8, uint8, bool) {
	switch op {
	case If:
		return 2, 1, true
	case And, Or:
		return 3, 4, true
	case Eq, Neq:
		return 5, 6, true
	case Lt, Gt, Lte, Gte:
		return 7, 8, true
	case Mod:
		return 9, 10, true
	default:
		return 0, 0, false
	}
}
