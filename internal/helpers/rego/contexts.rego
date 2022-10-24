package circleci.config

# import data.circleci.utils

import future.keywords

blocklist_by_project(project_id, context_list) = {reason |
	to_set(project_id)[data.meta.project_id]

	some wfName, workflow in input.workflows
	some job in workflow.jobs
	some jobName, jobInfo in job

	jobContexts := to_set(jobInfo.context)
	blockContexts := to_set(context_list)
	
    some ctx in jobContexts & blockContexts
	
	reason := sprintf("context %q used in job %q has been banned from current project", [ctx, jobName])
}

to_set(value) = {elem |
	some i
	elem := value[i]
}

if {
	is_array(value)
}

else = {value} {
	true
}
