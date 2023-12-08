package tester

import (
	"cmp"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/CircleCI-Public/circle-policy-agent/internal"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v2"
)

type Test struct {
	Input              any
	Meta               any
	Decision           any
	Compile            *bool
	PipelineParameters map[string]any `yaml:"pipeline_parameters"`
	Cases              map[string]*Test
}

func (t Test) NamedCases() []NamedTest {
	result := make([]NamedTest, 0, len(t.Cases))
	for name, test := range t.Cases {
		result = append(result, NamedTest{name, test})
	}
	slices.SortFunc(result, func(a, b NamedTest) int { return cmp.Compare(a.Name, b.Name) })
	return result
}

type NamedTest struct {
	Name string
	*Test
}

type ParentTestContext struct {
	Name               string
	Input              any
	Meta               any
	Decision           any
	Compile            bool
	PipelineParameters map[string]any
}

func loadTests(path string) (tests map[string]*Test, err error) {
	//nolint gosec
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, &tests)
	if err != nil {
		return
	}

	for name := range tests {
		if !strings.HasPrefix(name, "test_") {
			delete(tests, name)
		}
	}

	for _, t := range tests {
		sanitizeTest(t)
	}
	return
}

func getTestFolders(path string) (folders []string, err error) {
	if path != "./..." {
		path = strings.TrimPrefix(path, "./")
	}
	if !strings.HasSuffix(path, "/...") {
		return []string{path}, nil
	}
	err = filepath.WalkDir(path[:len(path)-4], func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if name := d.Name(); len(name) > 1 && name[0] == '.' {
			return filepath.SkipDir
		}
		folders = append(folders, path)
		return nil
	})
	return
}

func sanitizeTest(t *Test) {
	if t == nil {
		return
	}

	t.Decision = internal.ConvertYAMLMapKeyTypes(t.Decision)
	t.Input = internal.ConvertYAMLMapKeyTypes(t.Input)
	t.Meta = internal.ConvertYAMLMapKeyTypes(t.Meta)
	t.PipelineParameters = internal.ConvertYAMLMapKeyTypes(t.PipelineParameters).(map[string]any)

	for _, t := range t.Cases {
		sanitizeTest(t)
	}
}
