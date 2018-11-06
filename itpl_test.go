package itpl_test

import (
	"testing"

	"github.com/davidmz/itpl"
	"github.com/spf13/afero"
)

type Fls map[string]string

var testData = []struct {
	Files  Fls
	Result string
}{
	{Fls{"/entry": `ABC`}, `ABC`},
	{Fls{"/entry": `{{xxx}}ABC`}, `{{xxx}}ABC`},
	{Fls{"/entry": `{{- xxx -}} ABC`}, `{{xxx}}ABC`},
	{Fls{"/entry": `{{ xxx "yyy"}} ABC`}, `{{xxx "yyy"}} ABC`},
	{Fls{"/entry": `ABC {{include "index2"}}`, "/index2": `DEF`}, `ABC DEF`},
	{Fls{"/entry": `ABC {{include "a/index2"}}`, "/a/index2": `DEF`}, `ABC DEF`},
	{Fls{"/entry": `ABC {{- include "a/index2"}}`, "/a/index2": `DEF`}, `ABCDEF`},
	{Fls{"/entry": `ABC{{include "index3"}} {{include "index2"}}`, "/index2": `DEF{{include "index3"}}`, "/index3": `!`}, `ABC! DEF!`},
	{Fls{"/entry": `{{block "A" .}}ABC{{end}}`}, `{{define "A"}}ABC{{end}}{{template "A" .}}`},
	{Fls{"/entry": `{{xxx|len}}ABC{{yyy|zzz}}`}, `{{xxx | len}}ABC{{yyy | zzz}}`},
	{Fls{"/entry": `{{if .x}}{{include "./inc"}}{{end}}`, "/inc": `Hi!`}, `{{if .x}}Hi!{{end}}`},
}

func TestTable(t *testing.T) {
	for _, entry := range testData {
		t.Run(entry.Result, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			for name, content := range entry.Files {
				afero.WriteFile(fs, name, []byte(content), 0644)
			}
			out, err := itpl.NewLoader().Fs(fs).Load("/entry")
			if err != nil {
				t.Errorf("got error %v", err)
			} else if out != entry.Result {
				t.Errorf("got %q, expected %q", out, entry.Result)
			}
		})
	}
}
