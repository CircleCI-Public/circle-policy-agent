package org

import data.circleci.config
import future.keywords

policy_name["ban_orbs"]

enable_rule["ban_orbs"] {
	not data.meta.require_version
}

ban_orbs = config.ban_orbs(["bad"])

enable_rule["ban_orbs_version"] {
	data.meta.require_version
}

ban_orbs_version = config.ban_orbs_version(["bad@1.2.3"])
