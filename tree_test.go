package jsluice

import (
	"strconv"
	"testing"

	sitter "github.com/odvcencio/gotreesitter"
	"github.com/odvcencio/gotreesitter/grammars"
)

func TestCollapsedString(t *testing.T) {
	cases := []struct {
		JS       []byte
		Expected string
	}{
		{[]byte(`"./login.php?redirect="+url`), "./login.php?redirect=EXPR"},
		{[]byte(`'/path/'+['one', 'two', 'three'].join('/')`), "/path/EXPR"},
		{[]byte(`someVar`), "EXPR"},
	}

	lang := grammars.JavascriptLanguage()
	parser := sitter.NewParser(lang)

	for i, c := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			tree, _ := parser.Parse(c.JS)
			root := NewNode(tree.RootNode(), c.JS, lang)

			// Example tree:
			//   program
			//     expression_statement
			//       binary_expression
			//         left: string ("./login.php?redirect=")
			//         right: identifier (url)
			//
			// We want the binary_expression to pass to CollapsedString, which is
			// the first Named Child of the first Named Child of the root node.
			actual := root.NamedChild(0).NamedChild(0).CollapsedString()

			if actual != c.Expected {
				t.Errorf("want %s for CollapsedString(%s), have: %s", c.Expected, c.JS, actual)
			}
		})
	}
}
