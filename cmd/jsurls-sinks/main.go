package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strings"

	sitter "github.com/odvcencio/gotreesitter"
	"github.com/odvcencio/gotreesitter/grammars"
)

func main() {

	reWhitespace := regexp.MustCompile(`\s{2,}`)
	reJSName := regexp.MustCompile(`^[a-zA-Z0-9_$.-]+$`)

	flag.Parse()
	source, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	lang := grammars.JavascriptLanguage()
	parser := sitter.NewParser(lang)

	tree, err2 := parser.Parse(source)
	if err2 != nil {
		log.Fatal(err2)
	}
	root := tree.RootNode()

	enter := func(n *sitter.Node) {
		switch n.Type(lang) {
		case "assignment_expression":
			left := n.ChildByFieldName("left", lang)
			right := n.ChildByFieldName("right", lang)
			if left == nil || right == nil {
				return
			}

			rightContent := right.Text(source)
			if !startsWithString(rightContent) {
				return
			}

			rightContent = reWhitespace.ReplaceAllString(rightContent, " ")
			rightStr := dequote(right.Text(source))

			if couldBePath(rightStr) {
				fmt.Printf("%s (assignment)\n", left.Text(source))
			}

		case "call_expression":
			callName := n.ChildByFieldName("function", lang).Text(source)
			// It's common to find things like immediately called anonymous functions
			// in JS source, and we don't care about those because we could never match
			// on them
			if !reJSName.MatchString(callName) {
				return
			}

			arguments := n.ChildByFieldName("arguments", lang)
			if arguments == nil {
				return
			}

			// we want to iterate over the arguments and find
			// any that look like a url
			c := sitter.NewTreeCursor(arguments, tree)

			// no args
			if !c.GotoFirstChild() {
				return
			}

			foundPath := false
			position := 0
			for {
				arg := c.CurrentNode()
				if arg == nil {
					break
				}

				// named args only (i.e. don't count commas etc)
				if arg.IsNamed() {

					argContent := arg.Text(source)
					if startsWithString(argContent) && couldBePath(dequote(argContent)) {
						foundPath = true
						break
					}
					position++
				}

				if !c.GotoNextSibling() {
					break
				}
			}

			if foundPath {
				fmt.Printf("%s (arg %d)\n", callName, position)
			}
		}
	}
	queryNodes(root, enter, lang, source, tree)
}

func startsWithString(in string) bool {
	if len(in) < 2 {
		return false
	}

	p := in[0:1]
	if p == `"` || p == "'" || p == "`" {
		return true
	}

	return false
}

func couldBePath(in string) bool {

	if (strings.HasPrefix(in, "http:") && len(in) > 7) ||
		(strings.HasPrefix(in, "https:") && len(in) > 8) ||
		(strings.HasPrefix(in, "/") && len(in) > 3) ||
		(strings.HasPrefix(in, "./") && len(in) > 4) {
		return true
	}

	return false
}

func queryNodes(n *sitter.Node, enter func(*sitter.Node), lang *sitter.Language, source []byte, tree *sitter.Tree) {

	query, err := sitter.NewQuery(
		"[(assignment_expression) (call_expression)] @matches",
		lang,
	)
	if err != nil {
		log.Fatal(err)
	}

	cursor := query.Exec(n, lang, source)

	for {
		match, exists := cursor.NextMatch()
		if !exists {
			break
		}

		for _, capture := range match.Captures {
			enter(capture.Node)
		}
	}
}

func walk(n *sitter.Node, enter func(*sitter.Node), tree *sitter.Tree) {

	c := sitter.NewTreeCursor(n, tree)

	// walkies
	recurse := true
	for {
		// descend into the tree
		if recurse && c.GotoFirstChild() {
			recurse = true
			enter(c.CurrentNode())
			continue
		}

		// move sideways
		if c.GotoNextSibling() {
			recurse = true
			enter(c.CurrentNode())
			continue
		}

		// climb back up the tree, but make sure we don't descend right back to where we were
		if c.GotoParent() {
			recurse = false
			continue
		}
		break
	}

}

func dequote(in string) string {
	return strings.Trim(in, "'\"`")
}
