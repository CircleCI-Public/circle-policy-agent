package org

import data.circleci.config

policy_name["context_lists"]

enable_rule["context_allowlist"]

context_allowlist := config.context_allowlist_by_project_id(data.meta.project_list, data.meta.allowlist)

enable_rule["context_blocklist"]

context_blocklist = config.context_blocklist_by_project_id(data.meta.project_list, data.meta.blocklist)

enable_rule["context_branch_allowlist"]

context_branch_allowlist = config.context_allowlist_by_branch(data.meta.branch_list, data.meta.allowlist)