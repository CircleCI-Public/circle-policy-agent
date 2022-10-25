package helpers

import (
	"embed"
	"path"

	"github.com/open-policy-agent/opa/ast"
)

//go:embed rego
var regoFS embed.FS

var helpers = make(map[string]map[string]*ast.Module)

type Type interface {
	String() string
	sealed()
}

type helperType string

func (t helperType) String() string { return string(t) }
func (t helperType) sealed()        {}

const (
	Config helperType = "config"
	Utils  helperType = "utils"
)

var types = []helperType{Config, Utils}

func containsType(list []helperType, value string) bool {
	for _, elem := range list {
		if elem.String() == value {
			return true
		}
	}
	return false
}

func init() {
	entries, err := regoFS.ReadDir("rego")
	if err != nil {
		panic(err)
	}
	if len(entries) != len(types) {
		panic("mismatch between helper types and rego FS")
	}
	for _, entry := range entries {
		name := entry.Name()
		if !containsType(types, name) {
			panic("invalid helper type in rego FS: " + name)
		}
		helpers[name] = loadRegoSubdir(path.Join("rego", name))
	}
}

func AppendHelpers(mods map[string]*ast.Module, helperType Type) {
	for filename, helper := range helpers[helperType.String()] {
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
