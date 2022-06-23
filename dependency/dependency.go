// Package dependency helps when scanning this repository for vulnerable dependencies. Often, the scans will identify
// vulnerabilities in transitive dependencies. Adding the dependency to the list in this package, and then pinning a
// more recent (non-vulnerable) version of the dependency in the go.mod file resolves this issue.
package dependency

import (
	// nolint:revive
	_ "github.com/aws/aws-sdk-go"
	_ "github.com/containerd/containerd"
	_ "github.com/emicklei/go-restful"
	_ "golang.org/x/crypto/argon2"
	_ "golang.org/x/text"
	_ "gopkg.in/yaml.v2"
	_ "k8s.io/client-go"
	_ "k8s.io/kube-openapi/pkg/validation/spec"
)
