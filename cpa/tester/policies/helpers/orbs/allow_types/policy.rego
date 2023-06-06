package org

import data.circleci.config
import future.keywords

policy_name["orbs_allow_types"]

enable_hard["orbs_allow_types"]

orbs_allow_types = config.orbs_allow_types(["certified", "partner"])
