package org

policy_name["reason_types"]

enable_rule["reason_type_rule"]

reason_type_rule := reason {
	data.meta.type == "string"
	reason := "string error"
} else := reason {
	data.meta.type == "array"
	reason := ["reason one", "reason two"]
} else := reason {
	data.meta.type == "map"
	reason := {"k1": "r1", "k2": "r2"}
}
