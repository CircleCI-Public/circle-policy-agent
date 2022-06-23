package circleci.config

import future.keywords.in

jobs := {job | walk(input, [path, value])
	path[_] == "workflows"
	job = get_job_name(value[_].jobs[_])
}

# after the CircleCI config is converted from YAML to JSON, the job can either be a string or a map
get_job_name(job) = job_name {
	# if a value is returned here (i.e., this evaluates to true), the job is a map and the key is the job's name
	_ = job[key]
	job_name = key
} else = job

require_jobs(job_names) = { job_name: msg | job_name := job_names[_]
        not job_name in jobs
        msg := sprintf("%s job is required", [job_name])
}