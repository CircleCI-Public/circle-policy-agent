package org

import future.keywords
import data.circleci.config

policy_name["require_jobs"]

enable_rule["require_jobs"]

require_jobs = config.require_jobs(["job_one", "job_two"])
