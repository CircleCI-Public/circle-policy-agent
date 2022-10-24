package circleci.config

import data.circleci.utils

import future.keywords

blocklist_by_project(project_id, context_list) = {reason |
	utils.to_set(project_id)[data.meta.project_id]

	some wfName, workflow in input.workflows
	some job in workflow.jobs
	some jobName, jobInfo in job

	jobContexts := utils.to_set(jobInfo.context)
	blockContexts := utils.to_set(context_list)
	
    some ctx in jobContexts & blockContexts
	
	reason := sprintf("context %q used in job %q has been banned from current project", [ctx, jobName])
}

