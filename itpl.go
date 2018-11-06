package itpl // import "github.com/davidmz/itpl"

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"text/template/parse"

	"github.com/spf13/afero"
)

// Load loads template from OS filesystem and process the include actions. It returns
// processed template as a string or error if something went wrong.
func Load(fileName string) (string, error) {
	return NewLoader().Load(fileName)
}

// Loader allows to load and process templates with more options that the bare Load function.
// Loader can load files from non-standard filesystem (it uses github.com/spf13/afero as filesystem
// abstraction).
type Loader struct {
	fs           afero.Fs
	nowProcessed map[string]bool
}

// NewLoader creates a new Loader based on OS filesystem.
func NewLoader() *Loader {
	return &Loader{
		fs:           afero.NewOsFs(),
		nowProcessed: make(map[string]bool),
	}
}

// Fs allow to change filesystem of Loader. It may be useful in tests with in-memory filesystem.
func (ld *Loader) Fs(fs afero.Fs) *Loader {
	ld.fs = fs
	return ld
}

// Load loads template and process the include actions. It returns
// processed template as a string or error if something went wrong.
func (ld *Loader) Load(fileName string) (string, error) {
	fileName = filepath.Clean(fileName)
	if ld.nowProcessed[fileName] {
		return "", fmt.Errorf("circular import detected: %q is already processed", fileName)
	}
	ld.nowProcessed[fileName] = true
	defer func() { delete(ld.nowProcessed, fileName) }()

	bodyBytes, err := afero.ReadFile(ld.fs, fileName)
	if err != nil {
		return "", err
	}
	body := string(bodyBytes)

	funcs := map[string]interface{}{}

	const rootName = ""
	var tree map[string]*parse.Tree
	const maxFuncs = 100
	i := 0
	for {
		i++
		var err error
		tree, err = parse.Parse(rootName, body, "", "", builtinFuncs, funcs)
		if err == nil {
			break
		} else if found := funcErrRe.FindStringSubmatch(err.Error()); found != nil && i < maxFuncs {
			funcs[found[1]] = func() {}
		} else {
			return "", err
		}
	}

	sb := new(strings.Builder)

	for name, tr := range tree {
		if err := ld.processLists(fileName, tr.Root); err != nil {
			return "", err
		}
		if name != rootName {
			fmt.Fprintf(sb, "{{define %q}}", name)
		}
		for _, node := range tr.Root.Nodes {
			sb.WriteString(node.String())
		}
		if name != rootName {
			fmt.Fprint(sb, "{{end}}")
		}
	}

	return sb.String(), nil
}

func (ld *Loader) processLists(fileName string, lists ...*parse.ListNode) error {
	for _, list := range lists {
		if list == nil {
			continue
		}

		for idx, node := range list.Nodes {
			switch node := node.(type) {
			case *parse.ActionNode:
				args := node.Pipe.Cmds[0].Args
				if len(args) >= 2 &&
					args[0].Type() == parse.NodeIdentifier &&
					args[1].Type() == parse.NodeString {
					ident := args[0].(*parse.IdentifierNode)
					if ident.Ident == "include" {
						str := args[1].(*parse.StringNode)
						incFileName := filepath.Join(filepath.Dir(fileName), str.Text)
						incString, err := ld.Load(incFileName)
						if err != nil {
							return err
						}
						list.Nodes[idx] = newRawNode(node, incString)
						continue
					}
				}
			case *parse.IfNode:
				if err := ld.processLists(fileName, node.List, node.ElseList); err != nil {
					return err
				}
			case *parse.RangeNode:
				if err := ld.processLists(fileName, node.List, node.ElseList); err != nil {
					return err
				}
			case *parse.WithNode:
				if err := ld.processLists(fileName, node.List, node.ElseList); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

type rawNode struct {
	parse.Node
	str string
}

func newRawNode(node parse.Node, str string) parse.Node {
	return &rawNode{
		Node: node.Copy(),
		str:  str,
	}
}

func (r *rawNode) String() string { return r.str }

var builtinFuncs = map[string]interface{}{
	"and":      func() {},
	"call":     func() {},
	"html":     func() {},
	"index":    func() {},
	"js":       func() {},
	"len":      func() {},
	"not":      func() {},
	"or":       func() {},
	"print":    func() {},
	"printf":   func() {},
	"println":  func() {},
	"urlquery": func() {},

	// Comparisons
	"eq": func() {},
	"ge": func() {},
	"gt": func() {},
	"le": func() {},
	"lt": func() {},
	"ne": func() {},
}

var funcErrRe = regexp.MustCompile(`: function "(.+?)" not defined$`)
