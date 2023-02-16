package circleci.utils

import future.keywords

# Convert value to set if it isn't one 
to_set(value) := value if {
	is_set(value)
} else := {elem | some i;  elem := value[i]} {
	is_array(value)
} else := {value}

to_array(value) := value if is_array(value) else := [value]

get_element_name(value) := value if is_string(value) else := key { 
	is_object(value)
	count(value) == 1
	some key, _ in value
}