package org

policy_name["test"]

enable_rule["name_is_bob"]

hard_fail["name_is_bob"] { data.meta.hard }

name_is_bob = "name must be bob!" {
	input.name != "bob"
}
