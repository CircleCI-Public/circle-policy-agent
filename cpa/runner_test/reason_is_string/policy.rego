package org

policy_name["test"]

enable_rule["name_is_bob"]

name_is_bob = "name must be bob" {
	input.name != "bob"
}
