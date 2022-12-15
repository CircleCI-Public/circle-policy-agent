package org

policy_name["test"]

enable_rule := ["name_is_bob", "type_is_person"]

hard_fail := ["type_is_person"]

name_is_bob = "name must be bob!" {
	input.name != "bob"
}

type_is_person = "type must be person" {
	input.type != "person"
}
