package org

import data.circleci.config

policy_name["context_lists"]

enable_rule["context_allowlist"]

context_allowlist := config.contexts_allowed_by_project_ids(data.meta.project_list, data.meta.allowlist)

enable_rule["context_blocklist"]

context_blocklist = config.contexts_blocked_by_project_ids(data.meta.project_list, data.meta.blocklist)

enable_rule["context_reservelist"]

context_reservelist = config.contexts_reserved_by_project_ids(data.meta.project_list, data.meta.reservelist)

enable_rule["context_branch_reservelist"]

context_branch_reservelist = config.contexts_reserved_by_branches(data.meta.branch_list, data.meta.reservelist)