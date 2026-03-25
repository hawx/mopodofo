package mo

import (
	"testing"

	"hawx.me/code/assert"
)

func TestParseFormula(t *testing.T) {
	testcases := map[string]struct {
		formula  string
		tokens   Tokens
		node     node
		expected map[int]int
		match    func(int) int
	}{
		"no plurals": {
			formula: "0",
			tokens:  Tokens{0},
			node:    intNode{i: 0},
			expected: map[int]int{
				0: 0,
				1: 0,
				2: 0,
			},
			match: func(n int) int { return 0 },
		},
		"english style plurals": {
			formula: "n != 1",
			tokens:  Tokens{N, Neq, 1},
			node:    binNode{left: nNode{}, op: Neq, right: intNode{i: 1}},
			expected: map[int]int{
				0: 1,
				1: 0,
				2: 1,
			},
			match: func(n int) int {
				if n != 1 {
					return 1
				}
				return 0
			},
		},
		"french style plurals": {
			formula: "n>1",
			tokens:  Tokens{N, Gt, 1},
			node:    binNode{left: nNode{}, op: Gt, right: intNode{i: 1}},
			expected: map[int]int{
				0: 0,
				1: 0,
				2: 1,
			},
			match: func(n int) int {
				if n > 1 {
					return 1
				}
				return 0
			},
		},
		"arabic style plurals": {
			formula: "n==0 ? 0 : n==1 ? 1 : n==2 ? 2 : n%100>=3 && n%100<=10 ? 3 : n%100>=11 && n%100<=99 ? 4 : 5",
			tokens:  Tokens{N, Eq, 0, If, 0, Else, N, Eq, 1, If, 1, Else, N, Eq, 2, If, 2, Else, N, Mod, 100, Gte, 3, And, N, Mod, 100, Lte, 10, If, 3, Else, N, Mod, 100, Gte, 11, And, N, Mod, 100, Lte, 99, If, 4, Else, 5},
			node: ternaryNode{
				cond: binNode{left: nNode{}, op: Eq, right: intNode{i: 0}},
				yes:  intNode{i: 0},
				no: ternaryNode{
					cond: binNode{left: nNode{}, op: Eq, right: intNode{i: 1}},
					yes:  intNode{i: 1},
					no: ternaryNode{
						cond: binNode{left: nNode{}, op: Eq, right: intNode{i: 2}},
						yes:  intNode{i: 2},
						no: ternaryNode{
							cond: binNode{
								left: binNode{
									left:  binNode{left: nNode{}, op: Mod, right: intNode{i: 100}},
									op:    Gte,
									right: intNode{i: 3},
								},
								op: And,
								right: binNode{
									left:  binNode{left: nNode{}, op: Mod, right: intNode{i: 100}},
									op:    Lte,
									right: intNode{i: 10},
								},
							},
							yes: intNode{i: 3},
							no: ternaryNode{
								cond: binNode{
									left: binNode{
										left:  binNode{left: nNode{}, op: Mod, right: intNode{i: 100}},
										op:    Gte,
										right: intNode{i: 11},
									},
									op: And,
									right: binNode{
										left:  binNode{left: nNode{}, op: Mod, right: intNode{i: 100}},
										op:    Lte,
										right: intNode{i: 99},
									},
								},
								yes: intNode{i: 4},
								no:  intNode{i: 5},
							},
						},
					},
				},
			},
			expected: map[int]int{
				0:   0,
				1:   1,
				2:   2,
				103: 3,
				110: 3,
				111: 4,
				199: 4,
				200: 5,
			},
			match: func(n int) int {
				if n == 0 {
					return 0
				}
				if n == 1 {
					return 1
				}
				if n == 2 {
					return 2
				}

				if n%100 >= 3 && n%100 <= 10 {
					return 3
				}

				if n%100 >= 11 && n%100 <= 99 {
					return 4
				}

				return 5
			},
		},
		"with parens": {
			formula: "n > 100 && (n < 200 || n == 5)",
			tokens:  Tokens{N, Gt, 100, And, Open, N, Lt, 200, Or, N, Eq, 5, Close},
			node: binNode{
				left: binNode{left: nNode{}, op: Gt, right: intNode{i: 100}},
				op:   And,
				right: binNode{
					left:  binNode{left: nNode{}, op: Lt, right: intNode{i: 200}},
					op:    Or,
					right: binNode{left: nNode{}, op: Eq, right: intNode{i: 5}},
				},
			},
			match: func(n int) int {
				if n > 100 && (n < 200 || n == 5) {
					return 1
				}
				return 0
			},
		},
		"silly parens": {
			formula: "(((0)))",
			tokens:  Tokens{Open, Open, Open, 0, Close, Close, Close},
			node:    intNode{i: 0},
			expected: map[int]int{
				0: 0,
				1: 0,
				2: 0,
			},
			match: func(n int) int { return 0 },
		},
	}

	for name, tc := range testcases {
		t.Run(name+"/tokenise", func(t *testing.T) {
			tokens, err := tokenise(tc.formula)
			assert.Nil(t, err)
			assert.Equal(t, tc.tokens, tokens)
		})

		t.Run(name+"/parse", func(t *testing.T) {
			node, err := parse(&tc.tokens, 0)
			assert.Nil(t, err)
			assert.Equal(t, tc.node, node)
		})

		t.Run(name, func(t *testing.T) {
			f, err := parseFormula(tc.formula)
			assert.Nil(t, err)

			for in, out := range tc.expected {
				assert.Equal(t, out, f(in))
			}
		})

		t.Run(name+"/match", func(t *testing.T) {
			f, err := parseFormula(tc.formula)
			assert.Nil(t, err)

			for n := range 10000 {
				assert.Equal(t, tc.match(n), f(n))
			}
		})
	}
}
