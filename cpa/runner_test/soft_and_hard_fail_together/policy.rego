package org

import future.keywords

policy_name["soft_and_hard_fails"]

enable_rule contains rule if {some rule in ["soft_failure_rule", "hard_failure_rule"]}

hard_fail contains "hard_failure_rule"

soft_failure_rule = "soft failure!"
hard_failure_rule = "hard failure!"
