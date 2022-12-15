package org

policy_name["test"]

hard_fail["name_is_bob"]

name_is_bob = "name must be bob!" {
	input.name != "bob"
}
