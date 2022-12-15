package org

policy_name["test"]

enable_rule["name_must_be_bob"]

name_must_be_bob[name] = reason {
	name := input.names[_]
	name != "bob"
	reason := sprintf("%s is not bob", [name])
}
