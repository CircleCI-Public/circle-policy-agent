package org

import future.keywords
import data.circleci.config

policy_name["require_jobs"]

enable_rule["require_jobs"]

require_jobs = config.require_jobs(["job_one", "job_two"])

test_get_job_name_object = config.get_job_name({"build": {}}) == "build"
test_get_job_name_string = config.get_job_name("build") == "build"