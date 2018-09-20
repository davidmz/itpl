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
	fs afero.Fs
}

// NewLoader creates a new Loader based on OS filesystem.
func NewLoader() *Loader {
	return new(Loader).Fs(afero.NewOsFs())
}

// Fs allow to change filesystem of Loader. It may be useful in tests with in-memory filesystem.
func (ld *Loader) Fs(fs afero.Fs) *Loader {
	ld.fs = fs
	return ld
}

// funcRe matches action names (or function names). It can be greedy
// and can match any function-like names, it is ok.
var funcRe = regexp.MustCompile(`{{(?:- )?\s*(\p{L}[\p{L}\p{Nd}]*)`)

// Load loads template and process the include actions. It returns
// processed template as a string or error if something went wrong.
func (ld *Loader) Load(fileName string) (string, error) {
	fileName = filepath.Clean(fileName)

	bodyBytes, err := afero.ReadFile(ld.fs, fileName)
	if err != nil {
		return "", err
	}
	body := string(bodyBytes)

	funcMatches := funcRe.FindAllStringSubmatch(body, -1)
	funcs := map[string]interface{}{}
	for _, m := range funcMatches {
		name := m[1]
		if _, ok := funcs[name]; !ok {
			funcs[name] = func() {}
		}
	}

	const rootName = ""
	tree, err := parse.Parse(rootName, body, "", "", funcs)
	if err != nil {
		return "", err
	}

	sb := new(strings.Builder)

	str, err := ld.processTree(tree[rootName], fileName)
	if err != nil {
		return "", err
	}
	sb.WriteString(str)
	for name, tr := range tree {
		if name == rootName {
			continue
		}
		str, err := ld.processTree(tr, fileName)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(sb, "{{define %q}}%s{{end}}", name, str)
	}

	return sb.String(), nil
}

func (ld *Loader) processTree(tr *parse.Tree, fileName string) (string, error) {
	sb := new(strings.Builder)
	for _, n := range tr.Root.Nodes {
		if n.Type() == parse.NodeAction {
			a := n.(*parse.ActionNode)
			if len(a.Pipe.Cmds) == 1 {
				args := a.Pipe.Cmds[0].Args
				if len(args) >= 2 && args[0].Type() == parse.NodeIdentifier && args[1].Type() == parse.NodeString {
					ident := args[0].(*parse.IdentifierNode)
					if ident.Ident == "include" {
						str := args[1].(*parse.StringNode)
						incFileName := filepath.Join(filepath.Dir(fileName), str.Text)
						incString, err := ld.Load(incFileName)
						if err != nil {
							return "", err
						}
						sb.WriteString(incString)
						continue
					}
				}
			}
		}
		sb.WriteString(n.String())
	}
	return sb.String(), nil
}
