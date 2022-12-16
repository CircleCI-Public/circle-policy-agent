package org

import data.circleci.config
import future.keywords

policy_name["orb_helper"]

enable_rule["require_security_orbs"] {
	not data.meta.require_version
}

require_security_orbs = config.require_orbs(["circleci/security"])

enable_rule["require_security_orbs_version"] {
	data.meta.require_version
}

require_security_orbs_version = config.require_orbs_version(["circleci/security@1.2.3"])
