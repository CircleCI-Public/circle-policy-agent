package org

policy_name["enable_hard"]

# if enabled twice, the violation should only be reported once
enable_rule["double_enabled_rule"]
enable_hard["double_enabled_rule"]

enable_hard["some_hard_rule"]

double_enabled_rule = "this hard rule was enabled twice but reported once"
some_hard_rule = "this hard rule was enabled in one line - cool!"
