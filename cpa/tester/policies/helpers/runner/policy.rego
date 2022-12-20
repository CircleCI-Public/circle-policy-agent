package org

import future.keywords
import data.circleci.config

policy_name["runner_helpers"]

enable_rule["manage_resource_classes"]

manage_resource_classes = config.resource_class_by_project({
    "medium": {"large", "medium"},
    "large": {"large"}
})
