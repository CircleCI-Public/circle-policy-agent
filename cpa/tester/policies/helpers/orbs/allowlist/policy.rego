package org

import data.circleci.config
import future.keywords

policy_name["orbs_allowlist"]

enable_hard["orbs_allowlist"] {
	not data.meta.require_version
}

orbs_allowlist = config.orbs_allowlist({"partner/useful","partner2/useful"})

