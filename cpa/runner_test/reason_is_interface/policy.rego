package org

policy_name["test"]

enable_rule["name_starts_with_a_or_b"]

name_starts_with_a_or_b = reason {
	not startswith(input.name, "a")
	reason := ["input does not start with a", "input does not start with b"]
}
