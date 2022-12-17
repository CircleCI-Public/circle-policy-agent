package circleci.config

import future.keywords

resource_class_by_project(reservered_classes_to_projects) = {reason | 
        some job_name, job_info in input.jobs
        allowed_projects := reservered_classes_to_projects[job_info.resource_class]
        not data.meta.project_id in allowed_projects
        reason := sprintf("project is not allowed to use resource_class %q declared in job %q", [job_info.resource_class, job_name])
}
