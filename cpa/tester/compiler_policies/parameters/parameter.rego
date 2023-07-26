package org

policy_name["parameters"]

enable_rule["assert_params"]

# for this test, the mock test compiler will inject the pipeline parameters as the _compiled_ value
# this will allow us to test the expected parameters are being passed from our test declarations
# by comparing to static values we expect in meta.
assert_params = reason {
	data.meta.expected_parameters != input._compiled_
	reason = "expected params do not match compiled params"
}
