package helpers

import (
	"embed"
	"path"

	"github.com/open-policy-agent/opa/ast"
)

//go:embed rego
var regoFS embed.FS

var helpers = make(map[string]map[string]*ast.Module)

type Type string

const (
	Config Type = "config"
	Utils  Type = "utils"
)

func init() {
	entries, err := regoFS.ReadDir("rego")
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		helpers[entry.Name()] = loadRegoSubdir(path.Join("rego", entry.Name()))
	}
}

func AppendHelpers(mods map[string]*ast.Module, helperType Type) {
	for filename, helper := range helpers[string(helperType)] {
		mods[filename] = helper
	}
}

func loadRegoSubdir(root string) map[string]*ast.Module {
	entries, err := regoFS.ReadDir(root)
	if err != nil {
		panic(err)
	}
	result := make(map[string]*ast.Module)
	for _, entry := range entries {
		name := path.Join(root, entry.Name())
		data, err := regoFS.ReadFile(name)
		if err != nil {
			panic(err)
		}
		mod, err := ast.ParseModule(name, string(data))
		if err != nil {
			panic(err)
		}
		result["circleci/"+name] = mod
	}
	return result
}
