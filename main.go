package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/CircleCI-Public/circle-policy-agent/cpa"
)

func main() {
	b, err := cpa.ParseBundle(map[string]string{
		"example.rego": `
			package org

			policy_name["bob"]

			enable_rule["meta_out"]

			meta_out = data.meta
		`,
	})
	if err != nil {
		panic(err)
	}

	value, err := b.Eval(context.Background(), "data", map[string]string{}, cpa.Meta(struct {
		K1 string `json:",omitempty"`
	}{
		K1: "",
	}))
	if err != nil {
		panic(err)
	}

	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")

	e.Encode(value)
}
