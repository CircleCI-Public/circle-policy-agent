# policy-agent

[![CircleCI](https://circleci.com/gh/CircleCI-Public/circle-policy-agent/tree/main.svg?style=svg&circle-token=2b2f376a5a6ed8e77fa543586116a5879285986c)](https://circleci.com/gh/CircleCI-Public/circle-policy-agent/tree/main)

The policy-agent is essentially a CircleCI-flavored wrapper library around the [Open Policy Agent](https://github.com/open-policy-agent/opa) (OPA), which will allow the users to write the policy documents in CircleCI terminology.

policy-agent is responsible, at its core, for doing the following:

- [Integrating with the Go API](https://www.openpolicyagent.org/docs/latest/integration/#integrating-with-the-go-api) to evaluate policies.
- Expose [custom built-in functions](https://www.openpolicyagent.org/docs/latest/extensions/#custom-built-in-functions-in-go) that make policies (i.e., rules) easier to write.

# tools setup

We use [`golangci-lint`](https://github.com/golangci/golangci-lint) and [`task`](https://taskfile.dev/#/) in this repository:

```
$ brew install go-task/tap/go-task
$ brew install golangci-lint
```

or

```
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.46.2
```

See common project tasks with `task -l`, including:

```
task lint
task fmt
task test
```

# run lint

```
task lint
```

# run tests

```
task test
```

## API Documentation

With [`godoc`](https://pkg.go.dev/golang.org/x/tools/cmd/godoc), you can view generated html
documentation on your machine for policy-agent. There are two `task` commands for convenience:

```
* doc: 		Run 'godoc', print docs url
* doc-open: Run 'godoc', open the docs url in your browser
```

## Helpers

CircleCI has provided helper functions to make it easier to write Rego policies. To use
this code, please include `import data.circleci.config` in your Rego file.

Currently supported helpers:

- [orbs](./internal/helpers/docs/orbs.MD#orbs)
  - [ban_orbs](./internal/helpers/docs/orbs.MD#ban_orbs)
  - [ban_orbs_version](./internal/helpers/docs/orbs.MD#ban_orbs_version)
