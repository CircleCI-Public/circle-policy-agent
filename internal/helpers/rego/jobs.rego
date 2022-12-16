package circleci.config

import future.keywords

jobs := { job_name | 
	some job in input.workflows[_].jobs
	job_name = get_job_name(job)
}

# after the CircleCI config is converted from YAML to JSON, the job can either be a string or a map
get_job_name(job) = job_name {
	is_object(job)
	count(job) == 1
	some job_name, _ in job	
} else = job { is_string(job) }

require_jobs(job_names) = { job_name: msg | job_name := job_names[_]
        not job_name in jobs
        msg := sprintf("%s job is required", [job_name])
}