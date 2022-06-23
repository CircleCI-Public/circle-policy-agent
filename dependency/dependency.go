package dependency

// For more information:
// https://circleci.atlassian.net/wiki/spaces/SD/pages/6530269417/Upgrade+Direct+and+Transitive+Dependencies
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
