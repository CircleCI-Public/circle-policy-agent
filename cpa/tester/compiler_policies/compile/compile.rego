package org

import data.circleci.config

policy_name["compile_example"]

enable_rule["context_blocklist"]

context_blocklist = config.contexts_blocked_by_project_ids("project", {"c1", "c2", "c3"})
