package circleci.config

import data.circleci.utils

import future.keywords

context_blocklist_by_project_id(project_id, context_list) = {reason |
	utils.to_set(project_id)[data.meta.project_id]

	some wf_name, workflow in input.workflows
	some job_name, job_info in workflow.jobs[_]

	illegal_contexts := utils.to_set(job_info.context) & utils.to_set(context_list)
	count(illegal_contexts) > 0

	reason := sprintf("%s.%s: uses context value(s) in blocklist for project: %s", [wf_name, job_name, concat(", ", illegal_contexts)])
}

context_allowlist_by_project_id(project_id, context_list) = {reason |
	utils.to_set(project_id)[data.meta.project_id]

	some wf_name, workflow in input.workflows
	some job_name, job_info in workflow.jobs[_]

	illegal_contexts := utils.to_set(job_info.context) - utils.to_set(context_list)
	count(illegal_contexts) > 0

	reason := sprintf("%s.%s: uses context value(s) not in allowlist for project: %s", [wf_name, job_name, concat(", ", illegal_contexts)])
}

context_allowlist_by_branch(branch_list, context_list) = {reason |
	some wf_name, workflow in input.workflows
	some job_name, job_info in workflow.jobs[_]

	protected_contexts := utils.to_set(context_list) & utils.to_set(job_info.context)
	count(protected_contexts) > 0

	invalid_branches := utils.to_set(job_info.filters.branches.only) - utils.to_set(branch_list)
	count(invalid_branches) > 0

	reason := sprintf(
		"%s.%s: uses context value(s): %s - that cannot be used with branches: %s",
		[wf_name, job_name, concat(", ", protected_contexts), concat(", ", invalid_branches)],
	)
} | {reason |
	some wf_name, workflow in input.workflows
	some job_name, job_info in workflow.jobs[_]

	protected_contexts := utils.to_set(context_list) & utils.to_set(job_info.context)
	count(protected_contexts) > 0

	not job_info.filters.branches.only

	reason := sprintf(
		"%s.%s: uses context value(s): %s - that must be limited to branches: %s",
		[wf_name, job_name, concat(", ", protected_contexts), concat(", ", utils.to_set(branch_list))],
	)
}
