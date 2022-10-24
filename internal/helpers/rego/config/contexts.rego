package circleci.config

import data.circleci.utils

import future.keywords

context_blocklist_by_project(project_id, context_list) = {reason |
	utils.to_set(project_id)[data.meta.project_id]

	some wfName, workflow in input.workflows
	some job in workflow.jobs
	some jobName, jobInfo in job

	jobContexts := utils.to_set(jobInfo.context)
	blockContexts := utils.to_set(context_list)
	
    some ctx in jobContexts & blockContexts
	
	reason := sprintf("context %q used in job %q has been banned from current project", [ctx, jobName])
}

context_allowlist_by_project(project_id, context_list) = { reason  | 
	utils.to_set(project_id)[data.meta.project_id]

	some wfName, workflow in input.workflows
	some job in workflow.jobs
	some jobName, jobInfo in job

	
	some ctx in utils.to_set(jobInfo.context)
	
	allowContexts := utils.to_set(context_list)
	not allowContexts[ctx]
	
	reason := sprintf("context %q used in job %q is not part of allowed list of contexts for project", [ctx, jobName])
}

