package cpa

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
	"golang.org/x/exp/slices"

	"github.com/CircleCI-Public/circle-policy-agent/internal/helpers"
)

const policyName = "policy_name"

var policyNameExpr = regexp.MustCompile(`^\w+$`)

func parsePolicyName(m *ast.Module) (string, error) {
	if len(m.Rules) == 0 {
		return "", fmt.Errorf("must declare rule %q but module contains no rules", policyName)
	}

	head := m.Rules[0].Head
	if head == nil {
		return "", fmt.Errorf("cannot parse rule head")
	}
	if name := head.Name.String(); name != policyName {
		return "", fmt.Errorf("first rule declaration must be %q but found %q", policyName, name)
	}

	var name string
	if head.Key == nil {
		return "", fmt.Errorf("invalid %s declaration: must declare as key", policyName)
	}
	if err := json.Unmarshal([]byte(head.Key.String()), &name); err != nil {
		return "", fmt.Errorf("invalid %s: %v", policyName, err)
	}

	if len(name) == 0 {
		return "", fmt.Errorf("%s must not be empty", policyName)
	}
	if len(name) > 80 {
		return "", fmt.Errorf("%s must be maximum 80 characters but got %d", policyName, len(name))
	}
	if !policyNameExpr.MatchString(name) {
		return "", fmt.Errorf("%q must use alphanumeric and underscore characters only", policyName)
	}

	return name, nil
}

func hasImport(mods map[string]*ast.Module, importPath string) bool {
	for _, m := range mods {
		for _, i := range m.Imports {
			if i.Path.String() == importPath {
				return true
			}
		}
	}
	return false
}

// parseBundle will parse multiple rego files together into a bundle
func parseBundle(bundle map[string]string, rules ...LintRule) (*Policy, error) {
	moduleMap := make(map[string]*ast.Module, len(bundle))
	source := make(map[string]string, len(bundle))
	nameCount := make(map[string]uint32, len(bundle))

	var multiErr MultiError

	for file, rego := range bundle {
		mod, err := ast.ParseModule(file, rego)
		if err != nil {
			multiErr = append(multiErr, fmt.Errorf("failed to parse file %q: %w", file, err))
			continue
		}

		name, err := parsePolicyName(mod)
		if err != nil {
			multiErr = append(multiErr, fmt.Errorf("failed to parse file: %q: %w", file, err))
			continue
		}

		moduleMap[name] = mod
		source[name] = rego
		nameCount[name]++
	}

	if len(multiErr) > 0 {
		return nil, fmt.Errorf("failed to parse policy file(s): %w", multiErr)
	}

	for name, count := range nameCount {
		if count > 1 {
			multiErr = append(multiErr, fmt.Errorf("policy %q declared %d times", name, count))
		}
	}

	if len(multiErr) > 0 {
		return nil, fmt.Errorf("failed to parse bundle: %w", multiErr)
	}

	for policyName, mod := range moduleMap {
		for _, rule := range rules {
			if err := rule(mod); err != nil {
				multiErr = append(multiErr, fmt.Errorf("%q: %w", policyName, err))
			}
		}
	}

	if len(multiErr) > 0 {
		return nil, fmt.Errorf("failed policy linting: %w", multiErr)
	}

	rego.RegisterBuiltin2(&rego.Function{
		Name: "validate_orb_types",
		Decl: types.NewFunction(types.Args(types.A, types.A), types.A),
	},
		func(_ rego.BuiltinContext, _allowedOrbTypes, _orbsUsed *ast.Term) (*ast.Term, error) {

			// This is a dummy orb registry. Some assumptions:
			// 1. This would be replaced by a HTTP GET to a public endpoint which fetches similar data
			// 2. An orb's type (certified/partner/public) can be known just by the namespace (eg all orbs under circleci/* are 'certified').
			dummyOrbRegistry := map[string]string{
				"certified_namespace":   "certified",
				"certified_namespace_2": "certified",
				"certified_namespace_3": "certified",
				"partner_namespace":     "partner",
				"partner_namespace_2":   "partner",
				"partner_namespace_3":   "partner",
				"public_namespace":      "public",
				"public_namespace_2":    "public",
				"public_namespace_3":    "public",
			}

			var orbsUsed map[string]string
			var allowedOrbTypesSlice []string //TODO: Should ideally be a set for faster query, converted to map below
			var violatingOrbs []string
			err := ast.As(_orbsUsed.Value, &orbsUsed)
			if err != nil {
				panic(err)
			}
			err = ast.As(_allowedOrbTypes.Value, &allowedOrbTypesSlice)
			if err != nil {
				panic(err)
			}
			allowedOrbTypes := make(map[string]bool, len(allowedOrbTypesSlice))
			for _, v := range allowedOrbTypesSlice {
				allowedOrbTypes[v] = true
			}

			for orbUsed := range orbsUsed {
				parts := strings.Split(orbUsed, "/")
				orbNamespace := parts[0]
				namespaceType, exists := dummyOrbRegistry[orbNamespace]
				if !exists {
					panic("unknown orb")
				}
				if _, ok := allowedOrbTypes[namespaceType]; !ok {
					violatingOrbs = append(violatingOrbs, orbUsed)
				}
			}

			a := ast.MustInterfaceToValue(violatingOrbs)
			return ast.NewTerm(a), nil
		})

	if hasImport(moduleMap, "data.circleci.config") {
		helpers.AppendHelpers(moduleMap, helpers.Config)
	}

	// Utils check should happen after other circleci modules since
	// circleci modules may use utils themselves
	if hasImport(moduleMap, "data.circleci.utils") {
		helpers.AppendHelpers(moduleMap, helpers.Utils)
	}

	capabilities := ast.CapabilitiesForThisVersion()
	capabilities.AllowNet = []string{}

	disallowedBuiltins := []string{"http.send", "net.lookup_ip_addr"}
	for _, builtin := range disallowedBuiltins {
		i := slices.IndexFunc(capabilities.Builtins, func(elem *ast.Builtin) bool {
			return elem.Name == builtin
		})
		capabilities.Builtins = slices.Delete(capabilities.Builtins, i, i+1)
	}

	compiler := ast.
		NewCompiler().
		WithCapabilities(capabilities).
		WithStrict(true)

	if compiler.Compile(moduleMap); compiler.Failed() {
		return nil, fmt.Errorf("failed to compile policy: %w", compiler.Errors)
	}

	return &Policy{compiler, source}, nil
}

// ParseBundle will restrict package name to 'org'. This allows us to more easily extract information from the OPA output after evaluating a
// policy, because we know what the keys will be in the map that contains the results (e.g., map["org"]["enable_rule"] to find enabled rules).
//
//nolint:lll
func ParseBundle(files map[string]string) (*Policy, error) {
	return parseBundle(files, AllowedPackages("org"), DisallowMetaBranch())
}

type MultiError []error

func (err MultiError) Error() string {
	messages := make([]string, len(err))
	for i, e := range err {
		messages[i] = e.Error()
	}

	switch len(messages) {
	case 0:
		return "no errors"
	case 1:
		return messages[0]
	default:
		return fmt.Sprintf("%d error(s) occurred: %s", len(err), strings.Join(messages, "; "))
	}
}

func (err MultiError) Unwrap() []error {
	return []error(err)
}
