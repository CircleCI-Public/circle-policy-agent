package org

import data.circleci.utils

policy_name["utils_tests"]

test_to_array_scalar = utils.to_array(1) == [1]
test_to_array_array = utils.to_array([1]) == [1]
test_to_array_set = utils.to_array({1}) == [{1}]

test_to_set_scalar = utils.to_set(1) == {1}
test_to_set_array = utils.to_set([1,2,3]) == {1,2,3}
test_to_set_set = utils.to_set({1,2,3}) == {1,2,3}

test_get_element_name_string = utils.get_element_name("hello") == "hello"
test_get_element_name_object = utils.get_element_name({"world": {}}) == "world"
test_get_element_name_invalid_object = true if { not utils.get_element_name({"k1": "v1","k2":"v2"}) }

test_is_parameterized_expression_for_parameters_true = utils.is_parameterized_expression("test<< parameters.test >>") == true
test_is_parameterized_expression_for_pipeline_true = utils.is_parameterized_expression("<<pipeline.test>>") == true
test_is_parameterized_expression_false = utils.is_parameterized_expression("test-parameter") == false